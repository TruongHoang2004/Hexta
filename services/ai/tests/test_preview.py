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
    assert result["steps"] == ["prepare_context", "route", "render_script", "validate_output"]
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


def test_real_provider_cannot_be_enabled_accidentally(monkeypatch):
    monkeypatch.setenv("AI_PROVIDER", "openai")
    with pytest.raises(RuntimeError, match="fake only"):
        create_app(TOKEN)


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

    state = preview_graph.invoke({"request": PreviewRequest(message="  unknown  ")})
    assert state["steps"] == state["result"].steps
    assert state["steps"][-1] == "validate_output"
