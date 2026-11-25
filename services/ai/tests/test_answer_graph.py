"""Jev-routed graph: fallbacks, knowledge answers, structured answer validation and handoffs."""

import asyncio
import json

import pytest

from hexta_ai.graph import PreviewRequest, build_graph
from hexta_ai.providers.base import BaseAIProvider, Candidate, Classification, GroundedAnswer, ProviderError, Usage
from hexta_ai.providers.fake import FakeProvider
from hexta_ai.routing.jev import JevUnavailable
from tests.fakes import MemoryKnowledge, fake_router, jev_response

PRODUCT = {"id": "product:polo", "title": "Áo Polo Dry-Fit",
           "text": "Áo Polo Dry-Fit, chất liệu co giãn. Hotline 0912 345 678.",
           "attributes": {"price": "299.000đ", "stock": "15"}}


class AnswerProvider(BaseAIProvider):
    name = "anthropic"
    model = "claude-sonnet-5-5"
    simulated = False

    def __init__(self, answer=None, error=None):
        self.result = answer
        self.error = error
        self.contexts = []
        self.classified = 0

    async def classify(self, context):
        self.classified += 1
        return Classification(route="human", script_id=None, confidence=1), Usage()

    async def generate_candidate(self, context):
        return Candidate(text="unused", fact_ids=[]), Usage()

    async def answer(self, context):
        self.contexts.append(context)
        if self.error:
            raise self.error
        return self.result, Usage(input_tokens=100, output_tokens=20)


def good(**changes):
    return GroundedAnswer(**({"kind": "answer", "text": "Dạ áo Polo Dry-Fit giá 299.000đ, shop còn 15 cái ạ.",
                              "cited_source_ids": ["product:polo"], "confidence": 0.9, "needs_human": False} | changes))


def request(**changes):
    return PreviewRequest(**({"message": "áo polo giá bao nhiêu, sdt mình 0987 654 321", "tenant_id": "tenant1",
                              "allow_knowledge_answer": True,
                              "scripts": [{"id": "v1", "trigger": "phí ship", "text": "Phí ship 30.000đ.", "active": True}],
                              "history": [{"role": "user", "text": "mail mình a.b@gmail.com"}]} | changes))


def run(provider, value=None, *, jev=None, knowledge=None):
    router, client = fake_router(jev or jev_response("answer", 0.9))
    graph = build_graph(provider, router=router, knowledge=knowledge if knowledge is not None else MemoryKnowledge([PRODUCT]),
                        simulated=False)
    state = asyncio.run(graph.ainvoke({"request": value or request()}))
    assert state["steps"] == state["result"].steps
    return state["result"], client


def test_grounded_answer_is_generated_cited_and_requires_review():
    provider = AnswerProvider(good())
    knowledge = MemoryKnowledge([PRODUCT])
    result, _ = run(provider, knowledge=knowledge)
    assert result.route == "generate" and result.candidate_kind == "generated" and result.requires_review
    assert result.cited_source_ids == ["product:polo"] and result.confidence == 0.9
    assert result.router == "jev" and result.router_model == "jev-1.14.0" and result.route_probability == 0.9
    assert result.steps == ["validate_input", "prepare_context", "classify_intent", "retrieve_knowledge",
                            "answer_candidate", "validate_output"]
    assert result.usage.input_tokens == 100 and provider.classified == 0
    # Claude sees knowledge first, then conversation and message, without customer phone/email.
    context = provider.contexts[0]
    assert list(context) == ["knowledge", "conversation", "customer_message", "constraints"]
    sent = json.dumps(context, ensure_ascii=False)
    assert "0987" not in sent and "a.b@gmail.com" not in sent and "299.000đ" in sent
    assert knowledge.queries[0][0] == "tenant1" and "0987" not in knowledge.queries[0][1]
    assert knowledge.queries[0][2] == {"v1"}


@pytest.mark.parametrize("answer,reason,handoff", [
    (good(needs_human=True), "model_requested_human", "insufficient_information"),
    (good(cited_source_ids=[]), "ungrounded_answer", "insufficient_information"),
    (good(cited_source_ids=["product:other"]), "ungrounded_answer", "insufficient_information"),
    (good(cited_source_ids=["product:polo", "product:polo"]), "ungrounded_answer", "insufficient_information"),
    (good(confidence=0.5), "low_answer_confidence", "insufficient_information"),
    (good(text="Dạ áo giá 250.000đ ạ."), "ungrounded_number", "insufficient_information"),
    (good(text="Dạ còn 99 cái ạ."), "ungrounded_number", "insufficient_information"),
    (good(text="x" * 2001), "invalid_answer_output", "ai_error"),
    (good(text=""), "invalid_answer_output", "ai_error"),
    (good(text="Giá {price} ạ"), "unresolved_script_variables", "ai_error"),
])
def test_invalid_answers_become_handoffs(answer, reason, handoff):
    result, _ = run(AnswerProvider(answer))
    assert result.route == "human" and result.reason == reason and result.handoff_reason == handoff
    assert result.text is None and result.candidate_kind == "handoff" and not result.requires_review
    assert result.needs_human and result.cited_source_ids == []


@pytest.mark.parametrize("code", ["LLM_REFUSAL", "LLM_TIMEOUT", "QUOTA_EXCEEDED", "LLM_INVALID_OUTPUT",
                                  "INPUT_BUDGET_EXCEEDED", "LLM_PROVIDER_ERROR"])
def test_provider_refusals_and_errors_are_handoff_candidates(code):
    result, _ = run(AnswerProvider(error=ProviderError(code)))
    assert result.route == "human" and result.handoff_reason == "ai_error"
    assert result.reason == f"answer_{code.lower()}"


def test_numbers_in_the_question_and_tenant_hotline_are_allowed():
    text = "Dạ size 42 áo Polo giá 299.000đ, gọi hotline 0912 345 678 ạ."
    result, _ = run(AnswerProvider(good(text=text)), request(message="size 42 giá bao nhiêu"))
    assert result.route == "generate" and result.text == text


def test_clarify_from_model_requires_review():
    result, _ = run(AnswerProvider(good(kind="clarify", text="Bạn muốn hỏi size nào ạ?", cited_source_ids=[])))
    assert result.route == "clarify" and result.candidate_kind == "clarify" and result.requires_review


def test_missing_or_failing_knowledge_never_calls_the_model():
    provider = AnswerProvider(good())
    result, _ = run(provider, knowledge=MemoryKnowledge([]))
    assert (result.route, result.reason, result.handoff_reason) == ("human", "knowledge_insufficient", "insufficient_information")
    result, _ = run(provider, knowledge=MemoryKnowledge([PRODUCT], tenant="other"))
    assert result.reason == "knowledge_insufficient"
    result, _ = run(provider, knowledge=MemoryKnowledge(error=RuntimeError("db secret")))
    assert (result.reason, result.handoff_reason) == ("knowledge_retrieval_failed", "ai_error")
    assert provider.contexts == []


def test_caller_facts_ground_answers_without_a_store():
    provider = AnswerProvider(good(text="Dạ shop mở cửa 8h đến 22h ạ.", cited_source_ids=["fact:hours"]))
    value = request(facts=[{"id": "hours", "text": "Mở cửa 8h đến 22h."}])
    router, _ = fake_router(jev_response("answer"))
    graph = build_graph(provider, router=router, knowledge=None, simulated=False)
    result = asyncio.run(graph.ainvoke({"request": value}))["result"]
    # Without a knowledge service Jev's answer route is not enabled at all.
    assert result.route == "human" and result.reason == "knowledge_unavailable"
    result, _ = run(provider, value, knowledge=MemoryKnowledge([]))
    assert result.route == "generate" and result.cited_source_ids == ["fact:hours"]


def test_answer_route_requires_caller_opt_in():
    provider = AnswerProvider(good())
    result, _ = run(provider, request(allow_knowledge_answer=False))
    assert (result.route, result.reason, result.handoff_reason) == ("human", "knowledge_answer_not_enabled", "business_review")
    assert provider.contexts == []


def test_jev_script_route_renders_active_script():
    result, _ = run(AnswerProvider(), jev=jev_response("script", script="s1"))
    assert result.route == "script" and result.script_id == "v1" and result.candidate_kind == "script"
    assert not result.requires_review and result.reason == "jev_script_match"


@pytest.mark.parametrize("error", ["jev_timeout", "jev_rate_limited", "jev_auth", "jev_invalid_output"])
def test_jev_failure_falls_back_to_deterministic_then_human_without_paid_calls(error):
    provider = AnswerProvider(good())
    router, _ = fake_router(error=JevUnavailable(error))
    graph = build_graph(provider, router=router, knowledge=MemoryKnowledge([PRODUCT]), simulated=False)
    result = asyncio.run(graph.ainvoke({"request": request()}))["result"]
    assert (result.route, result.router, result.router_fallback_reason) == ("human", "deterministic", error)
    assert result.reason == "router_uncertain" and result.handoff_reason == "insufficient_information"
    assert provider.classified == 0 and provider.contexts == []
    exact = asyncio.run(graph.ainvoke({"request": request(message="phí ship")}))["result"]
    assert exact.route == "script" and exact.script_id == "v1" and exact.router == "deterministic"


def test_low_probability_falls_back():
    result, _ = run(AnswerProvider(good()), jev=jev_response("answer", 0.3))
    assert result.route == "human" and result.router_fallback_reason == "jev_low_probability"


def test_jev_human_and_customer_request_routes():
    result, _ = run(AnswerProvider(), jev=jev_response("answer", 0.9, wants_human=0.9))
    assert result.handoff_reason == "customer_requested_staff" and result.needs_human
    result, _ = run(AnswerProvider(), jev=jev_response("clarify", 0.9))
    assert result.route == "clarify" and not result.requires_review and result.text


def test_jev_disabled_keeps_existing_provider_path():
    provider = FakeProvider()
    result = asyncio.run(build_graph(provider).ainvoke({"request": request(message="phí ship")}))["result"]
    assert result.route == "script" and result.router == "deterministic" and result.router_fallback_reason == "jev_disabled"
    result = asyncio.run(build_graph(provider).ainvoke({"request": request()}))["result"]
    assert result.route == "human" and result.reason == "no_unique_active_script_fake_model_only"


def test_any_real_component_makes_runs_paid_for_recovery(monkeypatch):
    from fastapi.testclient import TestClient

    from hexta_ai.app import RuntimeProfile, create_app
    fake = FakeProvider()
    router, client = fake_router()
    assert RuntimeProfile(fake, None, None).simulated
    assert RuntimeProfile(fake, None, MemoryKnowledge()).simulated
    assert not RuntimeProfile(fake, router, None).simulated
    real_embeddings = MemoryKnowledge()
    real_embeddings.simulated = False
    assert not RuntimeProfile(fake, None, real_embeddings).simulated
    monkeypatch.setenv("AI_GRPC_ENABLED", "false")
    with TestClient(create_app("mock-service-token-at-least-24-chars", fake, router=router, knowledge=None)) as app:
        assert app.get("/health/live").json()["simulated"] is False
    assert client.closed


def test_fake_provider_answers_offline_from_knowledge():
    result, _ = run(FakeProvider())
    assert result.route == "generate" and result.cited_source_ids == ["product:polo"]
    # The fake answer quotes the record, so the tenant hotline stays and nothing is invented.
    assert result.text.startswith("Áo Polo Dry-Fit")
