import pytest
from httpx import ASGITransport, AsyncClient
from hexta_ai.app import create_app

TOKEN = "0123456789abcdef0123456789abcdef"


@pytest.fixture
def app():
    return create_app(internal_token=TOKEN)


@pytest.mark.asyncio
async def test_consult_endpoint_unauthorized(app):
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        res = await client.post("/internal/v1/consult", json={"tenant_id": "t1", "message": "hello"})
        assert res.status_code == 401


@pytest.mark.asyncio
async def test_consult_endpoint_success(app):
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        res = await client.post(
            "/internal/v1/consult",
            headers={"Authorization": f"Bearer {TOKEN}"},
            json={
                "tenant_id": "t1",
                "session_id": "s_1",
                "message": "Tư vấn cho tôi áo polo màu đen",
            },
        )
        assert res.status_code == 200
        data = res.json()
        assert data["tenant_id"] == "t1"
        assert data["session_id"] == "s_1"
        assert "message" in data
        assert len(data["recommended_products"]) > 0
        assert "guardrail_output" in data["steps"]
