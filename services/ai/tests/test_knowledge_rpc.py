"""Knowledge RPC validation/error mapping and wire compatibility for old and new callers."""

import asyncio

import grpc
import pytest
from psycopg import OperationalError

from hexta.ai.v1 import service_pb2 as pb
from hexta.ai.v1 import service_pb2_grpc as rpc
from hexta_ai.graph import PreviewResult
from hexta_ai.knowledge.embeddings import EmbeddingError
from hexta_ai.rpc import AIService
from hexta_ai.run_mapper import completed_run, preview_request

TOKEN = "test-grpc-service-token-at-least-24"
AUTH = (("authorization", f"Bearer {TOKEN}"),)


class RecordingKnowledge:
    simulated = True

    def __init__(self, error=None):
        self.error = error
        self.upserts = []
        self.deletes = []

    async def upsert(self, tenant, document):
        if self.error:
            raise self.error
        self.upserts.append((tenant, document))
        return "created", 1

    async def delete(self, tenant, source_type, source_id, revision):
        if self.error:
            raise self.error
        self.deletes.append((tenant, source_type, source_id, revision))
        return "deleted"


def document(**changes):
    fields = dict(source_type="product", source_id="polo", source_revision=3, title="Áo Polo",
                  body="Áo polo co giãn.", attributes={"price": "299.000đ", "stock": "15"})
    return pb.KnowledgeDocument(**(fields | changes))


async def harness(check, knowledge):
    service = AIService(TOKEN, knowledge=knowledge)
    server = grpc.aio.server()
    rpc.add_AIServiceServicer_to_server(service, server)
    port = server.add_insecure_port("127.0.0.1:0")
    await server.start()
    try:
        async with grpc.aio.insecure_channel(f"127.0.0.1:{port}") as channel:
            await check(rpc.AIServiceStub(channel))
    finally:
        await server.stop(0)
        await service.close()


async def expect(call, code):
    with pytest.raises(grpc.aio.AioRpcError) as error:
        await call
    assert error.value.code() == code
    return error.value.details()


def test_upsert_and_delete_validate_and_pass_documents():
    knowledge = RecordingKnowledge()

    async def check(stub):
        upsert = pb.UpsertKnowledgeRequest(tenant_id="tenant1", documents=[document()])
        await expect(stub.UpsertKnowledge(upsert), grpc.StatusCode.UNAUTHENTICATED)
        response = await stub.UpsertKnowledge(upsert, metadata=AUTH)
        assert [(r.source_id, r.outcome, r.chunk_count) for r in response.results] == [("polo", "created", 1)]
        for bad in (document(source_type="order"), document(source_id="bad id"), document(source_revision=-1),
                    document(body=""), document(body="x" * 20001), document(attributes={"Bad-Key": "1"})):
            await expect(stub.UpsertKnowledge(pb.UpsertKnowledgeRequest(tenant_id="tenant1", documents=[bad]),
                                              metadata=AUTH), grpc.StatusCode.INVALID_ARGUMENT)
        await expect(stub.UpsertKnowledge(pb.UpsertKnowledgeRequest(tenant_id="tenant1", documents=[document(), document()]),
                                          metadata=AUTH), grpc.StatusCode.INVALID_ARGUMENT)
        await expect(stub.UpsertKnowledge(pb.UpsertKnowledgeRequest(tenant_id="tenant1"), metadata=AUTH),
                     grpc.StatusCode.INVALID_ARGUMENT)
        await expect(stub.UpsertKnowledge(pb.UpsertKnowledgeRequest(tenant_id="../x", documents=[document()]),
                                          metadata=AUTH), grpc.StatusCode.INVALID_ARGUMENT)
        deleted = await stub.DeleteKnowledge(pb.DeleteKnowledgeRequest(
            tenant_id="tenant1", source_type="script", source_id="s1", source_revision=4), metadata=AUTH)
        assert deleted.outcome == "deleted"
        await expect(stub.DeleteKnowledge(pb.DeleteKnowledgeRequest(tenant_id="tenant1", source_type="x", source_id="s1"),
                                          metadata=AUTH), grpc.StatusCode.INVALID_ARGUMENT)
    asyncio.run(harness(check, knowledge))
    tenant, stored = knowledge.upserts[0]
    assert tenant == "tenant1" and stored.attributes == {"price": "299.000đ", "stock": "15"}
    assert knowledge.deletes == [("tenant1", "script", "s1", 4)]


@pytest.mark.parametrize("error,code", [
    (EmbeddingError("EMBEDDING_QUOTA"), grpc.StatusCode.RESOURCE_EXHAUSTED),
    (EmbeddingError("EMBEDDING_TIMEOUT"), grpc.StatusCode.UNAVAILABLE),
    (OperationalError("connection to secret-host failed"), grpc.StatusCode.UNAVAILABLE),
])
def test_storage_and_embedding_failures_are_retryable_and_sanitized(error, code):
    async def check(stub):
        details = await expect(stub.UpsertKnowledge(pb.UpsertKnowledgeRequest(tenant_id="t1", documents=[document()]),
                                                    metadata=AUTH), code)
        assert "secret-host" not in details
        await expect(stub.DeleteKnowledge(pb.DeleteKnowledgeRequest(tenant_id="t1", source_type="product", source_id="p"),
                                          metadata=AUTH), code)
    asyncio.run(harness(check, RecordingKnowledge(error)))


def test_missing_knowledge_store_is_failed_precondition():
    async def check(stub):
        await expect(stub.UpsertKnowledge(pb.UpsertKnowledgeRequest(tenant_id="t1", documents=[document()]),
                                          metadata=AUTH), grpc.StatusCode.FAILED_PRECONDITION)
    asyncio.run(harness(check, None))


def test_old_callers_keep_old_semantics_and_wire_format():
    old = pb.StartRunRequest(tenant_id="t1", run_id="r1", turn_id="u1", message="price", allow_generation=True)
    # An old client never sets field 12, so the bytes (and the idempotency digest) are unchanged.
    assert pb.StartRunRequest.FromString(old.SerializeToString()).allow_knowledge_answer is False
    assert b"\x60" not in old.SerializeToString()
    preview = preview_request(old)
    assert preview.allow_knowledge_answer is False and preview.tenant_id == "t1"
    run = completed_run(pb.Run(tenant_id="t1", run_id="r1"), PreviewResult(route="script", text="x", script_id="s1",
                                                                            reason="r", steps=[]))
    # Fields an old Go client does not know are skipped by its decoder; known fields keep their meaning.
    assert (run.route, run.script_id, run.status) == ("script", "s1", pb.RUN_STATUS_SUCCEEDED)
    assert run.candidate_kind == "" and not run.requires_review
    new = completed_run(pb.Run(), PreviewResult(route="generate", text="answer", reason="knowledge_answer", steps=[],
                                                candidate_kind="generated", requires_review=True,
                                                cited_source_ids=["product:p1"], router="jev",
                                                router_model="jev-1.14.0", route_probability=0.8))
    decoded = pb.Run.FromString(new.SerializeToString())
    assert decoded.requires_review and list(decoded.cited_source_ids) == ["product:p1"]
    assert decoded.route == "generate" and decoded.router_model == "jev-1.14.0"
