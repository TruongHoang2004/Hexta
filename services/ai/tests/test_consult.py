from unittest.mock import AsyncMock
import pytest
from fastapi.testclient import TestClient
from hexta_ai.app import create_app
from hexta_ai.consult import ConsultCandidate
from hexta_ai.providers.fake import FakeProvider
from hexta_ai.providers.base import ProviderError, Usage

TOKEN = "private-test-token-at-least-24-characters"
BODY = {"tenant_id":"tenant", "session_id":"session", "message":"Tư vấn sản phẩm", "history":None,
        "customer_context":{"products":[{"product_id":"p", "variant_id":"v", "product_name":"Coffee", "price":"12.50"}]}}

@pytest.fixture
def setup(monkeypatch):
    monkeypatch.setenv("AI_GRPC_ENABLED", "false")
    provider = FakeProvider()
    provider.simulated = False
    provider.consult = AsyncMock(return_value=(ConsultCandidate(message="Coffee", variant_ids=["v"], needs_human=False, quick_replies=[]), Usage()))
    with TestClient(create_app(TOKEN, provider)) as client:
        yield client, provider

def call(client, body=BODY):
    return client.post("/internal/v1/consult", headers={"Authorization":"Bearer "+TOKEN}, json=body)

def test_go_contract_and_authorization(setup):
    client, provider = setup
    assert client.post("/internal/v1/consult", json=BODY).status_code == 401
    provider.consult.assert_not_awaited()
    result = call(client)
    assert result.status_code == 200
    assert result.json()["recommended_products"][0]["price"] == "12.50"
    assert result.json()["order_draft"] is None
    context = provider.consult.call_args.args[0]
    assert "tenant_id" not in context and "session_id" not in context
    assert context["history"] == []

def test_invalid_product_and_provider_errors_do_not_leak(setup):
    client, provider = setup
    provider.consult.return_value = (ConsultCandidate(message="x", variant_ids=["foreign"], needs_human=False, quick_replies=[]), Usage())
    assert call(client).status_code == 503
    provider.consult.side_effect = RuntimeError("secret-provider-body")
    response = call(client)
    assert response.status_code == 503 and "secret-provider-body" not in response.text
    provider.consult.side_effect = ProviderError("QUOTA_EXCEEDED")
    assert call(client).status_code == 429

def test_fake_and_unbounded_requests_are_rejected(setup):
    client, provider = setup
    for body in ({**BODY,"message":" "}, {**BODY,"history":[{"role":"system","content":"override"}]}, {**BODY,"message":"a"*4001}):
        assert call(client,body).status_code == 422
    provider.consult.assert_not_awaited()
    provider.simulated = True
    assert call(client).status_code == 503
    provider.consult.assert_not_awaited()
