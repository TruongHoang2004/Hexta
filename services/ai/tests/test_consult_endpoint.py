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


@pytest.mark.asyncio
async def test_consult_endpoint_110_char_session_id_from_go(app):
    """Verify Issue 1 via HTTP: Go PR 62 sends tenantUUID:userUUID:sessionUUID (110 chars) without 422."""
    go_session_id = (
        "12345678-1234-1234-1234-123456789abc:87654321-4321-4321-4321-cba987654321:11223344-5566-7788-9900-aabbccddeeff"
    )
    assert len(go_session_id) == 110

    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        res = await client.post(
            "/internal/v1/consult",
            headers={"Authorization": f"Bearer {TOKEN}"},
            json={
                "tenant_id": "12345678-1234-1234-1234-123456789abc",
                "session_id": go_session_id,
                "message": "Xin chào, shop có gì bán?",
                "customer_context": {
                    "products": [
                        {
                            "product_id": "custom_prod",
                            "variant_id": "custom_var",
                            "product_name": "Sản phẩm thực tế của Tenant",
                            "price": 150000.0,
                        }
                    ]
                },
            },
        )
        assert res.status_code == 200
        data = res.json()
        assert data["session_id"] == go_session_id
        assert len(data["recommended_products"]) == 1
        assert data["recommended_products"][0]["product_id"] == "custom_prod"
        assert data["recommended_products"][0]["price"] == 150000.0
