"""Offline demo path: keyword router, fake grounded answers, folded hashing embeddings and the clarify text."""

import asyncio

import pytest

from hexta_ai.graph import DEFAULT_CLARIFY_TEXT, PreviewRequest, build_graph, clarify_text_from_env
from hexta_ai.knowledge.embeddings import HashingEmbeddings
from hexta_ai.knowledge.models import KnowledgeInput, document_chunks
from hexta_ai.providers.fake import FakeProvider, compose_answer, excerpt
from hexta_ai.routing import JevRouter, KeywordRouter, router_from_env
from tests.fakes import MemoryKnowledge

SCRIPTS = [{"id": "v1", "trigger": "phí ship, ship bao nhiêu", "text": "Phí ship 30.000đ.", "active": True},
           {"id": "v2", "trigger": "giờ mở cửa", "text": "Shop mở cửa 8h-22h.", "active": True}]
PRODUCT = {"id": "product:polo", "title": "Áo Polo Dry-Fit",
           "text": "Áo Polo Dry-Fit\nChất liệu co giãn, thoáng mát. Hotline 0912 345 678.",
           "attributes": {"price": "299000", "stock": "15", "sku": "POLO-M"}}


def route(message, scripts=SCRIPTS):
    return asyncio.run(KeywordRouter().route(message, scripts, []))


@pytest.mark.parametrize("message,expected,reason", [
    ("Ship bao nhieu", "script", "keyword_script_match"),  # one phrase of a comma-joined trigger, no diacritics
    ("GIỜ MỞ CỬA", "script", "keyword_script_match"),
    ("cho mình gặp nhân viên", "human", "keyword_customer_requested_human"),
    ("toi muon hoan tien don hang", "human", "keyword_customer_requested_human"),
    ("áo thun basic giá bao nhiêu", "answer", "keyword_knowledge_question"),
    ("ao thun con hang khong", "answer", "keyword_knowledge_question"),
    ("chính sách đổi trả thế nào?", "answer", "keyword_knowledge_question"),
    ("shop ơi", "clarify", "keyword_needs_clarification"),
    ("alo", "clarify", "keyword_needs_clarification"),
    ("?", "clarify", "keyword_needs_clarification"),
])
def test_keyword_router_branches(message, expected, reason):
    decision = route(message)
    assert decision.route == expected and decision.reason == reason
    assert decision.router == "deterministic" and decision.router_model == "keyword-v1"
    if expected == "script":
        assert decision.script_id in {"v1", "v2"} and decision.probability == 1
    if expected == "human":
        assert decision.handoff_reason == "customer_requested_staff"


@pytest.mark.parametrize("message", ["cảm ơn", "ok", "mình muốn giao lưu với mọi người ở đây"])
def test_keyword_router_is_uncertain_without_a_rule(message):
    # "giao" is not the word "gia": whole-word matching only.
    assert route(message) is None


def test_keyword_router_needs_a_unique_active_script():
    duplicate = [*SCRIPTS, {"id": "v3", "trigger": "phí ship", "text": "x", "active": True}]
    assert route("phí ship", duplicate).route == "answer"  # ambiguous script: falls through to the topic rule
    inactive = [{**SCRIPTS[1], "active": False}]
    assert route("giờ mở cửa", inactive).route == "answer"  # an inactive script is never chosen


def test_router_from_env(monkeypatch):
    monkeypatch.delenv("AI_FAKE_ROUTER", raising=False)
    assert router_from_env() is None
    monkeypatch.setenv("AI_FAKE_ROUTER", "keyword")
    assert isinstance(router_from_env(), KeywordRouter)
    monkeypatch.setenv("AI_FAKE_ROUTER", "magic")
    with pytest.raises(RuntimeError):
        router_from_env()
    # An enabled Jev with a key always wins over the offline router.
    monkeypatch.setenv("AI_FAKE_ROUTER", "keyword")
    monkeypatch.setenv("AI_JEV_ENABLED", "true")
    monkeypatch.setenv("TYPESAFE_API_KEY", "test-only-typesafe-key")
    router = router_from_env()
    assert isinstance(router, JevRouter)
    asyncio.run(router.close())


def test_keyword_router_keeps_runs_simulated(monkeypatch):
    from fastapi.testclient import TestClient

    from hexta_ai.app import RuntimeProfile, create_app
    assert RuntimeProfile(FakeProvider(), KeywordRouter(), MemoryKnowledge()).simulated
    monkeypatch.setenv("AI_GRPC_ENABLED", "false")
    monkeypatch.setenv("AI_FAKE_ROUTER", "keyword")
    with TestClient(create_app("mock-service-token-at-least-24-chars", FakeProvider(), knowledge=None)) as app:
        assert app.get("/health/live").json()["simulated"] is True


def run(message, *, allow=True, knowledge=None):
    graph = build_graph(FakeProvider(), router=KeywordRouter(),
                        knowledge=knowledge if knowledge is not None else MemoryKnowledge([PRODUCT]))
    request = PreviewRequest(message=message, tenant_id="tenant1", allow_knowledge_answer=allow, scripts=SCRIPTS)
    return asyncio.run(graph.ainvoke({"request": request}))["result"]


def test_offline_demo_answer_is_grounded_cited_and_pending_review():
    result = run("áo polo giá bao nhiêu")
    assert result.route == "generate" and result.candidate_kind == "generated" and result.requires_review
    assert result.cited_source_ids == ["product:polo"] and result.reason == "knowledge_answer"
    assert result.router == "deterministic" and result.router_model == "keyword-v1" and result.simulated
    assert "299000đ" in result.text and "15" in result.text and "0912 345 678" in result.text
    assert "{" not in result.text and result.text.startswith("Dạ, về Áo Polo Dry-Fit ạ: Chất liệu co giãn")


def test_offline_demo_other_routes():
    clarify = run("shop ơi")
    assert clarify.route == "clarify" and clarify.text == DEFAULT_CLARIFY_TEXT and not clarify.requires_review
    unknown = run("mình muốn giao lưu với mọi người ở đây")
    assert unknown.route == "human" and unknown.router_fallback_reason == "keyword_no_match"
    assert run("áo polo giá bao nhiêu", allow=False).handoff_reason == "business_review"
    assert run("áo polo giá bao nhiêu", knowledge=MemoryKnowledge([], tenant="other")).route == "human"
    assert run("phí ship").route == "script"


def test_fake_answer_never_cuts_numbers_or_keeps_braces():
    long = "Mẫu " + "rất đẹp " * 60 + "mã 123 456 789 hết."
    quoted = excerpt("", long)
    assert "123" not in quoted and quoted.endswith("…")
    text = compose_answer({"id": "document:p", "title": "Chính sách {x}", "text": "Đổi trả trong 7 ngày.",
                           "attributes": {"stock": "0"}})
    assert "{" not in text and "hết hàng" in text and "7 ngày" in text


async def _similarity(query, document):
    embeddings = HashingEmbeddings()
    vector = (await embeddings.embed_documents(document_chunks(document)))[0]
    return sum(a * b for a, b in zip(await embeddings.embed_query(query), vector))


def test_hashing_embeddings_fold_diacritics_for_demo_retrieval():
    product = KnowledgeInput(source_type="product", source_id="p1", source_revision=1, title="Áo thun basic",
                             body="Danh mục: Áo\nMô tả: Áo thun cotton, form rộng.\nBiến thể: Trắng size M")
    policy = KnowledgeInput(source_type="document", source_id="d1", source_revision=1, title="Chính sách đổi trả",
                            body="Khách được đổi trả trong vòng 7 ngày kể từ khi nhận hàng, còn nguyên tem mác.")
    assert HashingEmbeddings.model == "hashing-768-v2"
    accented = asyncio.run(_similarity("áo thun basic giá bao nhiêu", product))
    plain = asyncio.run(_similarity("ao thun basic gia bao nhieu", product))
    assert accented == pytest.approx(plain) and accented > 0.3
    assert asyncio.run(_similarity("doi tra may ngay", policy)) > 0.3
    assert asyncio.run(_similarity("hôm nay trời đẹp quá", product)) < 0.15


def test_clarify_text_default_override_and_validation(monkeypatch):
    monkeypatch.delenv("AI_CLARIFY_TEXT", raising=False)
    assert clarify_text_from_env() == DEFAULT_CLARIFY_TEXT and "anh/chị" in DEFAULT_CLARIFY_TEXT
    monkeypatch.setenv("AI_CLARIFY_TEXT", "  Bạn hỏi sản phẩm nào ạ?  ")
    assert clarify_text_from_env() == "Bạn hỏi sản phẩm nào ạ?"
    graph = build_graph(FakeProvider(), router=KeywordRouter())
    result = asyncio.run(graph.ainvoke({"request": PreviewRequest(message="alo")}))["result"]
    assert result.text == "Bạn hỏi sản phẩm nào ạ?"
    for invalid in ("Xin chào {name}", "x" * 501):
        monkeypatch.setenv("AI_CLARIFY_TEXT", invalid)
        with pytest.raises(RuntimeError):
            clarify_text_from_env()
        with pytest.raises(RuntimeError):
            build_graph(FakeProvider())
    with pytest.raises(RuntimeError):
        build_graph(FakeProvider(), clarify_text="{x}")
