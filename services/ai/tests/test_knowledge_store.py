"""Opt-in pgvector tests on an explicit Atlas-migrated test database (AI_TEST_DATABASE_URL)."""

import asyncio
import os
from uuid import uuid4

import pytest

from hexta_ai.knowledge.embeddings import HashingEmbeddings
from hexta_ai.knowledge.models import KnowledgeInput
from hexta_ai.knowledge.service import KnowledgeService, RagConfig
from hexta_ai.knowledge.store import KnowledgeStore

DSN = os.getenv("AI_TEST_DATABASE_URL")
pytestmark = pytest.mark.skipif(not DSN, reason="AI_TEST_DATABASE_URL is required")


class CountingEmbeddings(HashingEmbeddings):
    def __init__(self):
        self.documents = 0

    async def embed_documents(self, texts):
        self.documents += 1
        return await super().embed_documents(texts)


def doc(source_id="polo", revision=1, body="Áo Polo Dry-Fit co giãn, thấm hút mồ hôi.", source_type="product", **changes):
    return KnowledgeInput(**({"source_type": source_type, "source_id": source_id, "source_revision": revision,
                              "title": "Áo Polo Dry-Fit", "body": body, "attributes": {"price": "299.000đ"}} | changes))


def scenario(body):
    async def run():
        embeddings = CountingEmbeddings()
        service = KnowledgeService(KnowledgeStore(DSN), embeddings, RagConfig(min_similarity=0.05))
        try:
            await service.ping()  # The explicit Atlas test job provisions the schema.
            await body(service, embeddings, "kt-" + uuid4().hex, "kt-" + uuid4().hex)
        finally:
            await service.close()
    asyncio.run(run())


def test_ingest_is_idempotent_revision_fenced_and_reembeds_only_changed_text():
    async def body(service, embeddings, tenant, _):
        assert await service.upsert(tenant, doc()) == ("created", 1)
        assert await service.upsert(tenant, doc()) == ("unchanged", 1)
        assert embeddings.documents == 1
        # A newer revision with only new attributes (price/stock) needs no embedding call.
        assert (await service.upsert(tenant, doc(revision=2, attributes={"price": "259.000đ"})))[0] == "attributes_updated"
        assert (await service.upsert(tenant, doc(revision=3, attributes={"price": "259.000đ"})))[0] == "unchanged"
        assert embeddings.documents == 1
        # Older or same-revision changed content is stale and never overwrites.
        assert (await service.upsert(tenant, doc(revision=2, body="Hàng cũ.")))[0] == "stale"
        assert (await service.upsert(tenant, doc(revision=3, body="Hàng cũ.")))[0] == "stale"
        long_body = " ".join(f"Câu mô tả số {i} về áo polo." for i in range(120))
        outcome, chunks = await service.upsert(tenant, doc(revision=4, body=long_body))
        assert outcome == "updated" and chunks > 1 and embeddings.documents == 2
        row = await service.store.source(tenant, "product", "polo")
        assert row["source_revision"] == 4 and row["chunk_count"] == chunks
        assert row["attributes"] == {"price": "299.000đ"}
    scenario(body)


def test_delete_tombstones_fence_out_of_order_upserts():
    async def body(service, _, tenant, __):
        await service.upsert(tenant, doc(revision=5))
        assert await service.delete(tenant, "product", "polo", 4) == "stale"
        assert await service.delete(tenant, "product", "polo", 6) == "deleted"
        assert await service.delete(tenant, "product", "polo", 6) == "not_found"
        assert (await service.upsert(tenant, doc(revision=6)))[0] == "stale"
        assert await service.retrieve(tenant, "áo polo", active_script_ids=set()) == []
        assert (await service.upsert(tenant, doc(revision=7)))[0] == "created"
        # Deleting a never-seen source leaves a tombstone so an older in-flight upsert cannot resurrect it.
        assert await service.delete(tenant, "document", "faq", 3) == "not_found"
        assert (await service.upsert(tenant, doc("faq", 2, source_type="document")))[0] == "stale"
    scenario(body)


def test_retrieval_is_tenant_isolated_and_script_filtered():
    async def body(service, _, tenant, other):
        await service.upsert(tenant, doc(body="Áo Polo Dry-Fit màu đen, size M L XL."))
        await service.upsert(other, doc(body="Áo Polo Dry-Fit màu đen, size M L XL.", attributes={"price": "1đ"}))
        await service.upsert(tenant, doc("ship", 1, source_type="script", title="Phí ship",
                                         body="Phí ship toàn quốc 30.000đ.", attributes={"version_id": "ver-2"}))
        items = await service.retrieve(tenant, "áo polo màu đen size L", active_script_ids=set())
        assert [i.id for i in items] == ["product:polo"]
        assert items[0].attributes == {"price": "299.000đ"}
        assert all(i.attributes.get("price") != "1đ" for i in items)
        assert await service.retrieve("kt-" + uuid4().hex, "áo polo", active_script_ids=set()) == []
        scripts = await service.retrieve(tenant, "phí ship toàn quốc", active_script_ids={"ver-1"})
        assert "script:ship" not in [i.id for i in scripts]
        scripts = await service.retrieve(tenant, "phí ship toàn quốc", active_script_ids={"ver-2"})
        assert scripts[0].id == "script:ship"
    scenario(body)


def test_similarity_floor_and_context_budget():
    async def body(service, _, tenant, __):
        await service.upsert(tenant, doc(body="Áo polo thể thao co giãn bốn chiều, thấm hút mồ hôi tốt. " * 4))
        strict = KnowledgeService(service.store, service.embeddings, RagConfig(min_similarity=0.99))
        assert await strict.retrieve(tenant, "giờ mở cửa cuối tuần", active_script_ids=set()) == []
        tiny = KnowledgeService(service.store, service.embeddings, RagConfig(min_similarity=0.05, max_context_tokens=200))
        assert await tiny.retrieve(tenant, "áo polo", active_script_ids=set()) == []
    scenario(body)
