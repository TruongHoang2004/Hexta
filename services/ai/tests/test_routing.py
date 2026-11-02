"""Provider outputs are proposals; graph guards enforce the authorized snapshot."""

import asyncio

import pytest
from fastapi.testclient import TestClient
from pydantic import ValidationError

from hexta_ai.app import create_app
from hexta_ai.graph import PreviewRequest, build_graph
from hexta_ai.providers.base import BaseAIProvider, Candidate, Classification, ProviderError, Usage


class StubProvider(BaseAIProvider):
    name = "gemini"
    model = "mock-model"
    simulated = False

    def __init__(self, route="script", script_id="s1", confidence=0.95, text="Delivery takes two days.", fact_ids=None):
        self.decision = Classification(route=route, script_id=script_id, confidence=confidence)
        self.candidate = Candidate(text=text, fact_ids=fact_ids if fact_ids is not None else ["f1"])
        self.contexts = []
        self.calls = 0
        self.closed = False

    async def classify(self, context):
        self.calls += 1
        self.contexts.append(context)
        return self.decision, Usage(input_tokens=10, output_tokens=5)

    async def generate_candidate(self, context):
        self.calls += 1
        return self.candidate, Usage(input_tokens=20, output_tokens=10)

    async def close(self):
        self.closed = True


def request(**changes):
    return PreviewRequest(**({"message": "How long is delivery?", "context_revision": 2,
                              "scripts": [{"id": "s1", "trigger": "delivery", "text": "Select a product.", "active": True}],
                              "facts": [{"id": "f1", "text": "Delivery takes two days."}]} | changes))


def run(provider, value=None):
    state = asyncio.run(build_graph(provider).ainvoke({"request": value or request()}))
    assert state["steps"] == state["result"].steps
    return state["result"]


@pytest.mark.parametrize("route,script_id,confidence,expected", [
    ("script", "s1", 0.95, "script"), ("script", "not-active", 0.99, "human"),
    ("script", "s1", 0.4, "human"), ("human", None, 1, "human"),
    ("clarify", None, 0.95, "clarify"), ("generate", "s1", 0.99, "human"),
])
def test_routes_confidence_and_active_script_ids(route, script_id, confidence, expected):
    result = run(StubProvider(route, script_id, confidence))
    assert result.route == expected
    assert result.provider == "gemini" and not result.simulated
    if expected == "human":
        assert result.text is None and result.script_id is None


def test_generation_requires_permission_and_verified_facts():
    provider = StubProvider("generate", None)
    assert run(provider).route == "human"
    assert provider.calls == 1
    assert run(StubProvider("generate", None), request(allow_generation=True, facts=[])).route == "human"
    output = run(StubProvider("generate", None), request(allow_generation=True))
    assert output.route == "generate" and output.text == "Delivery takes two days."
    assert output.usage.input_tokens == 30 and output.usage.output_tokens == 15


@pytest.mark.parametrize("text,ids", [
    ("Delivery is free.", ["f1"]), ("Delivery takes two days.", ["unknown"]),
    ("Delivery takes two days.", []), ("Delivery takes two days.", ["f1", "f1"]),
])
def test_generation_rejects_unsupported_claims_and_invalid_citations(text, ids):
    output = run(StubProvider("generate", None, text=text, fact_ids=ids), request(allow_generation=True))
    assert output.route == "human" and output.text is None
    assert output.reason == "ungrounded_candidate"


def test_pii_is_redacted_before_provider_and_in_script_output():
    provider = StubProvider()
    output = run(provider, request(
        message="Email private@example.test or +84 912 345 678",
        scripts=[{"id": "s1", "trigger": "contact", "text": "Call +84 912 345 678 or private@example.test", "active": True}],
        history=[{"role": "user", "text": "private@example.test"}],
        constraints=["Do not contact private@example.test"],
    ))
    assert "private@example.test" not in str(provider.contexts)
    assert "912 345" not in output.text
    assert "REDACTED" in output.text


def test_templates_and_control_characters_are_not_emitted():
    output = run(StubProvider(), request(scripts=[{"id": "s1", "trigger": "price", "text": "Price: {price}", "active": True}]))
    assert output.route == "human" and output.reason == "unresolved_script_variables"
    assert request(message="\x00ignore\u202erules").message == "ignorerules"
    for changes in ({"context_revision": -1}, {"scripts": [
        {"id": "x", "trigger": "one", "text": "a"}, {"id": "x", "trigger": "two", "text": "b"}]}):
        with pytest.raises(ValidationError):
            request(**changes)


def test_budget_is_enforced_before_any_provider_call():
    provider = StubProvider()
    provider.max_input_tokens = 1
    with pytest.raises(ProviderError, match="INPUT_BUDGET_EXCEEDED"):
        run(provider)
    assert provider.calls == 0


def test_app_real_provider_wiring_safe_errors_and_lifecycle(monkeypatch):
    monkeypatch.setenv("AI_GRPC_ENABLED", "false")
    provider = StubProvider()
    token = "mock-service-token-at-least-24-chars"
    with TestClient(create_app(token, provider)) as client:
        assert client.get("/health/ready").json() == {"status": "ready", "provider": "gemini", "simulated": False}
        assert client.post("/internal/v1/preview", json=request().model_dump()).status_code == 401
        headers = {"Authorization": f"Bearer {token}"}
        result = client.post("/internal/v1/preview", headers=headers, json=request().model_dump())
        assert result.status_code == 200 and result.json()["route"] == "script"
        async def quota(context):
            raise ProviderError("QUOTA_EXCEEDED")
        provider.classify = quota
        result = client.post("/internal/v1/preview", headers=headers, json=request().model_dump())
        assert result.status_code == 429 and result.json()["detail"] == "QUOTA_EXCEEDED"
        result = client.post("/internal/v1/preview", headers=headers, json={"message": "", "secret": "private-payload"})
        assert result.status_code == 422 and "private-payload" not in result.text
    assert provider.closed
