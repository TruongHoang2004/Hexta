import asyncio

import pytest
from fastapi.testclient import TestClient

from hexta_ai.app import create_app

TOKEN = "local-test-service-token-not-a-provider-key"


@pytest.fixture
def client(monkeypatch):
    monkeypatch.setenv("AI_PROVIDER", "fake")
    monkeypatch.setenv("AI_GRPC_ENABLED", "false")
    with TestClient(create_app(TOKEN)) as test_client:
        yield test_client


def call(client, scripts):
    return client.post("/internal/v1/preview", headers={"Authorization": f"Bearer {TOKEN}"},
                       json={"message": "price", "scripts": scripts})


def test_health_and_authorization(client):
    assert client.get("/health/ready").json()["simulated"] is True
    assert client.post("/internal/v1/preview", json={"message": "price"}).status_code == 401


def test_only_unique_active_script_is_rendered(client):
    script = {"id": "s1", "trigger": "price", "text": "Please select a product.", "active": True}
    result = call(client, [script]).json()
    assert result["route"] == "script"
    assert result["script_id"] == "s1"
    assert result["steps"] == ["validate_input", "prepare_context", "classify_intent", "render_script", "validate_output"]
    assert call(client, [{**script, "active": False}]).json()["route"] == "human"
    assert call(client, [script, {**script, "id": "s2"}]).json()["route"] == "human"


def test_unknown_and_missing_variables_do_not_fabricate_answers(client):
    assert call(client, []).json()["text"] is None
    result = call(client, [{"id": "s1", "trigger": "price", "text": "Price: {price}", "active": True}])
    assert result.json()["route"] == "human"
    assert result.json()["text"] is None


def test_reject_unknown_input_and_empty_message(client):
    headers = {"Authorization": f"Bearer {TOKEN}"}
    assert client.post("/internal/v1/preview", headers=headers,
                       json={"message": "price", "send": True}).status_code == 422
    assert client.post("/internal/v1/preview", headers=headers,
                       json={"message": ""}).status_code == 422


@pytest.mark.parametrize("env", [
    {"AI_PROVIDER": "unsupported_provider"},
    {"AI_PROVIDER": "openai", "AI_MODEL": "gpt-4o-mini", "OPENAI_API_KEY": "", "AI_API_KEY": ""},
    {"AI_PROVIDER": "anthropic", "AI_MODEL": "", "ANTHROPIC_API_KEY": "test-only"},
])
def test_missing_or_partial_ai_configuration_starts_disabled(monkeypatch, caplog, env):
    # Fail-safe (owner decision 2026-11-26): no crash and no deploy rollback; AI is disabled with a logged reason.
    for name, value in env.items():
        monkeypatch.setenv(name, value)
    monkeypatch.setenv("AI_GRPC_ENABLED", "false")
    with caplog.at_level("WARNING"), TestClient(create_app(TOKEN)) as test_client:
        assert test_client.get("/health/live").json()["provider"] == "disabled"
        script = {"id": "s1", "trigger": "price", "text": "Please select a product.", "active": True}
        assert call(test_client, [script]).json()["route"] == "script"
        result = call(test_client, []).json()
        assert result["route"] == "human" and result["reason"] == "ai_disabled" and result["text"] is None
    messages = " ".join(r.getMessage() for r in caplog.records)
    assert "AI disabled:" in messages and "AI runtime DISABLED" in messages
    assert "test-only" not in messages


def test_disabled_consult_never_creates_a_model(monkeypatch):
    monkeypatch.setenv("AI_PROVIDER", "disabled")
    monkeypatch.setenv("AI_GRPC_ENABLED", "false")
    from unittest.mock import patch
    with patch("hexta_ai.models.adapter.FakeStructuredChatModel") as fake_model, \
            TestClient(create_app(TOKEN)) as test_client:
        response = test_client.post("/internal/v1/consult", headers={"Authorization": f"Bearer {TOKEN}"},
                                    json={"tenant_id": "t1", "session_id": "s1", "message": "áo polo"})
        assert response.status_code == 200 and response.json()["message"]
        fake_model.assert_not_called()  # static fallback reply, no model of any kind


def test_weak_internal_token_is_rejected(monkeypatch):
    monkeypatch.setenv("AI_PROVIDER", "fake")
    monkeypatch.setenv("AI_GRPC_ENABLED", "false")
    with pytest.raises(RuntimeError, match="at least 24"):
        create_app("short")


@pytest.mark.parametrize("field", ["message", "id", "trigger", "text"])
def test_whitespace_only_inputs_are_rejected(client, field):
    script = {"id": "s1", "trigger": "price", "text": "Select a product.", "active": True}
    body = {"message": "price", "scripts": [script]}
    if field == "message":
        body[field] = " \t\n\u2003"
    else:
        script[field] = " \t\n\u2003"
    response = client.post("/internal/v1/preview", headers={"Authorization": f"Bearer {TOKEN}"}, json=body)
    assert response.status_code == 422


def test_graph_steps_stay_consistent_after_validation():
    from hexta_ai.graph import PreviewRequest, preview_graph

    state = asyncio.run(preview_graph.ainvoke({"request": PreviewRequest(message="  unknown  ")}))
    assert state["steps"] == state["result"].steps
    assert state["steps"][-1] == "validate_output"


PAID_KEYS = {"AI_JEV_ENABLED": "true", "TYPESAFE_API_KEY": "test-only-typesafe", "AI_EMBEDDING_PROVIDER": "gemini",
             "GEMINI_API_KEY": "test-only-gemini", "AI_DATABASE_URL": "postgresql://u:p@127.0.0.1:1/none"}


@pytest.mark.parametrize("env,provider,router,embeddings,simulated,state", [
    # Fake (offline) provider: Jev and Gemini keys never turn on a paid call; runs stay simulated.
    ({"AI_PROVIDER": "fake", **PAID_KEYS}, "fake", "none", "hashing-768-v2", True, "SIMULATED"),
    # Explicit offline demo flags: keyword router plus hashing embeddings, still simulated, still no Jev.
    ({"AI_PROVIDER": "fake", "AI_FAKE_ROUTER": "keyword", **PAID_KEYS}, "fake", "keyword", "hashing-768-v2",
     True, "SIMULATED"),
    # Disabled AI with every paid key present: no router, offline embeddings.
    ({"AI_PROVIDER": "disabled", **PAID_KEYS}, "disabled", "none", "hashing-768-v2", True, "DISABLED"),
    # A leftover model of another family keeps AI (and so Jev and Gemini) off.
    ({"AI_PROVIDER": "openai", "AI_MODEL": "claude-sonnet-5-5", "OPENAI_API_KEY": "test-only-openai", **PAID_KEYS},
     "disabled", "none", "hashing-768-v2", True, "DISABLED"),
    ({"AI_PROVIDER": "anthropic", "AI_MODEL": "gpt-4o-mini", "ANTHROPIC_API_KEY": "test-only-anthropic",
      **PAID_KEYS}, "disabled", "none", "hashing-768-v2", True, "DISABLED"),
    # The full enable set: real provider, Jev and Gemini embeddings (clients are constructed, never called).
    ({"AI_PROVIDER": "anthropic", "AI_MODEL": "claude-sonnet-5-5", "ANTHROPIC_API_KEY": "test-only-anthropic",
      **PAID_KEYS}, "anthropic", "jev", "gemini-embedding-001", False, "enabled"),
])
def test_runtime_configuration_matrix(monkeypatch, caplog, env, provider, router, embeddings, simulated, state):
    from unittest.mock import AsyncMock, patch
    monkeypatch.delenv("AI_FAKE_ROUTER", raising=False)
    for name, value in env.items():
        monkeypatch.setenv(name, value)
    monkeypatch.setenv("AI_GRPC_ENABLED", "false")
    with patch("langchain_google_genai.GoogleGenerativeAIEmbeddings"), \
            patch("hexta_ai.knowledge.embeddings.GeminiEmbeddings.close", AsyncMock()), \
            caplog.at_level("WARNING"), TestClient(create_app(TOKEN)) as test_client:
        health = test_client.get("/health/live").json()
        assert health["provider"] == provider and health["simulated"] is simulated
        if provider == "fake":
            # Fake answers are reachable only through the demo router and are always labelled simulated.
            script = {"id": "s1", "trigger": "price", "text": "Please select a product.", "active": True}
            assert call(test_client, [script]).json()["simulated"] is True
    startup = [r.getMessage() for r in caplog.records if r.getMessage().startswith("AI runtime")]
    assert len(startup) == 1 and state in startup[0]
    assert f"router={router} " in startup[0] and startup[0].endswith(f"embeddings={embeddings}")
    assert not any("test-only" in r.getMessage() for r in caplog.records)


@pytest.mark.parametrize("model,accepted", [
    ("gpt-4o-mini", True), ("o3-mini", True), ("chatgpt-4o-latest", True),
    ("claude-sonnet-5-5", False), ("gemini-2.5-flash", False),
])
def test_openai_requires_an_openai_model(monkeypatch, model, accepted):
    from hexta_ai.providers.factory import AIConfigError, configured_provider
    monkeypatch.setenv("AI_PROVIDER", "openai")
    monkeypatch.setenv("AI_MODEL", model)
    monkeypatch.setenv("OPENAI_API_KEY", "test-only-openai")
    if accepted:
        assert configured_provider().name == "openai"
    else:
        with pytest.raises(AIConfigError, match="not a openai model"):
            configured_provider()


@pytest.mark.parametrize("setting,value", [
    ("AI_RAG_TOP_K", "abc"), ("AI_RAG_TOP_K", "0"), ("AI_RAG_MIN_SIMILARITY", "2"),
    ("AI_RAG_MAX_CONTEXT_TOKENS", "secret-invalid"),
])
def test_invalid_rag_settings_disable_retrieval_without_failing_startup(monkeypatch, caplog, setting, value):
    monkeypatch.setenv("AI_PROVIDER", "fake")
    monkeypatch.setenv("AI_GRPC_ENABLED", "false")
    monkeypatch.setenv("AI_DATABASE_URL", "postgresql://u:p@127.0.0.1:1/none")
    monkeypatch.setenv(setting, value)
    with caplog.at_level("WARNING"), TestClient(create_app(TOKEN)) as test_client:
        assert test_client.get("/health/ready").json()["status"] == "ready"
    errors = [r for r in caplog.records if r.levelname == "ERROR"]
    assert any("Knowledge retrieval disabled" in r.getMessage() and setting in r.getMessage() for r in errors)
    assert any("embeddings=none" in r.getMessage() for r in caplog.records)
    assert not any("secret-invalid" in r.getMessage() for r in caplog.records)
