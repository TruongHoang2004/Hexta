"""Opt-in end-to-end over gRPC and an Atlas-migrated pgvector test DB: ingest, Jev answer route, paid fence."""

import asyncio
import os
from uuid import uuid4

import grpc
import pytest

from hexta.ai.v1 import service_pb2 as pb
from hexta.ai.v1 import service_pb2_grpc as rpc
from hexta_ai.app import RuntimeProfile
from hexta_ai.durable import DurableAIService
from hexta_ai.graph import build_graph
from hexta_ai.knowledge.embeddings import HashingEmbeddings
from hexta_ai.knowledge.service import KnowledgeService, RagConfig
from hexta_ai.knowledge.store import KnowledgeStore
from hexta_ai.providers.fake import FakeProvider
from hexta_ai.store import RunStore
from tests.fakes import fake_router, jev_response

DSN = os.getenv("AI_TEST_DATABASE_URL")
pytestmark = pytest.mark.skipif(not DSN, reason="AI_TEST_DATABASE_URL is required")
TOKEN = "integration-knowledge-token-24chars"
AUTH = (("authorization", f"Bearer {TOKEN}"),)


def test_ingest_then_answer_through_durable_run():
    async def scenario():
        tenant, other = "e2e-" + uuid4().hex, "e2e-" + uuid4().hex
        provider = FakeProvider()
        router, jev = fake_router(jev_response("answer", 0.9))
        knowledge = KnowledgeService(KnowledgeStore(DSN), HashingEmbeddings(), RagConfig(min_similarity=0.05))
        graph = build_graph(provider, router=router, knowledge=knowledge, simulated=False)

        async def runner(request):
            return (await graph.ainvoke({"request": request}))["result"]

        store = RunStore(DSN)
        service = DurableAIService(TOKEN, store, runner, RuntimeProfile(provider, router, knowledge), knowledge)
        await service.start()
        server = grpc.aio.server()
        rpc.add_AIServiceServicer_to_server(service, server)
        port = server.add_insecure_port("127.0.0.1:0")
        await server.start()
        try:
            async with grpc.aio.insecure_channel(f"127.0.0.1:{port}") as channel:
                stub = rpc.AIServiceStub(channel)
                product = pb.KnowledgeDocument(source_type="product", source_id="polo", source_revision=1,
                                               title="Áo Polo Dry-Fit", body="Áo Polo Dry-Fit màu đen, giá 299.000đ.",
                                               attributes={"stock": "15"})
                first = await stub.UpsertKnowledge(pb.UpsertKnowledgeRequest(tenant_id=tenant, documents=[product]),
                                                   metadata=AUTH)
                again = await stub.UpsertKnowledge(pb.UpsertKnowledgeRequest(tenant_id=tenant, documents=[product]),
                                                   metadata=AUTH)
                assert first.results[0].outcome == "created" and again.results[0].outcome == "unchanged"

                async def ask(tenant_id, run_id, **changes):
                    await stub.StartRun(pb.StartRunRequest(**({"tenant_id": tenant_id, "run_id": run_id, "turn_id": "t",
                                                               "message": "áo polo đen giá bao nhiêu",
                                                               "allow_knowledge_answer": True} | changes)), metadata=AUTH)
                    async with asyncio.timeout(10):
                        while True:
                            run = await stub.GetRun(pb.RunRequest(tenant_id=tenant_id, run_id=run_id), metadata=AUTH)
                            if run.status not in (pb.RUN_STATUS_QUEUED, pb.RUN_STATUS_RUNNING):
                                return run
                            await asyncio.sleep(0.05)

                run = await ask(tenant, "r1")
                assert run.status == pb.RUN_STATUS_SUCCEEDED and run.route == "generate"
                assert run.candidate_kind == "generated" and run.requires_review and not run.simulated
                assert list(run.cited_source_ids) == ["product:polo"] and "299.000đ" in run.text
                assert run.router == "jev" and run.router_model == "jev-1.14.0"
                # Another tenant with the same question has no knowledge and is handed off.
                isolated = await ask(other, "r1")
                assert isolated.route == "human" and isolated.reason == "knowledge_insufficient"
                # An old caller (no opt-in) never receives a generated answer.
                legacy = await ask(tenant, "r2", allow_knowledge_answer=False)
                assert legacy.route == "human" and legacy.reason == "knowledge_answer_not_enabled"
                assert len(jev.calls) == 3
                deleted = await stub.DeleteKnowledge(pb.DeleteKnowledgeRequest(
                    tenant_id=tenant, source_type="product", source_id="polo", source_revision=2), metadata=AUTH)
                assert deleted.outcome == "deleted"
                gone = await ask(tenant, "r3")
                assert gone.route == "human" and gone.reason == "knowledge_insufficient"
        finally:
            await server.stop(0)
            await service.close()
            async with KnowledgeStore(DSN) as cleanup:
                async with cleanup.connect() as db:
                    for value in (tenant, other):
                        await db.execute("DELETE FROM ai_runtime.knowledge_sources WHERE tenant_id=%s", (value,))
            async with RunStore(DSN) as cleanup:
                async with cleanup.connect() as db:
                    for value in (tenant, other):
                        await db.execute("DELETE FROM ai_runtime.runs WHERE tenant_id=%s", (value,))
            await knowledge.close()
    asyncio.run(scenario())
