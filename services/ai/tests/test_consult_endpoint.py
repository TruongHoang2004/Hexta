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


@pytest.mark.asyncio
async def test_consult_endpoint_variants_same_product_id(app):
    """Verify Issue 1 via HTTP: size L variant preserves variant_id='size_l' and price 200,000đ."""
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        res = await client.post(
            "/internal/v1/consult",
            headers={"Authorization": f"Bearer {TOKEN}"},
            json={
                "tenant_id": "t_variants",
                "session_id": "s_variants",
                "message": "Tôi muốn mua áo size L",
                "customer_context": {
                    "products": [
                        {
                            "product_id": "shirt_01",
                            "variant_id": "size_s",
                            "product_name": "Áo Sơ Mi Oxford Size S",
                            "price": 100000.0,
                            "stock": 10,
                        },
                        {
                            "product_id": "shirt_01",
                            "variant_id": "size_l",
                            "product_name": "Áo Sơ Mi Oxford Size L",
                            "price": 200000.0,
                            "stock": 10,
                        },
                    ]
                },
            },
        )
        assert res.status_code == 200
        data = res.json()
        # Recommendations must contain size L, not overwritten by size S
        rec = data["recommended_products"]
        assert len(rec) >= 1
        # The recommended item must preserve variant_id 'size_l' and price 200,000đ
        matched = [p for p in rec if p.get("variant_id") == "size_l"]
        assert len(matched) == 1
        assert matched[0]["price"] == 200000.0
        assert "Size L" in matched[0]["name"]


@pytest.mark.asyncio
async def test_consult_endpoint_decimal_budget_filtering(app):
    """Verify Issue 3 via HTTP: 'dưới 1,5 triệu' filters out 2,000,000đ item and keeps 1,200,000đ item."""
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        res = await client.post(
            "/internal/v1/consult",
            headers={"Authorization": f"Bearer {TOKEN}"},
            json={
                "tenant_id": "t_budget",
                "session_id": "s_budget",
                "message": "Tìm áo dưới 1,5 triệu",
                "customer_context": {
                    "products": [
                        {
                            "product_id": "jacket_cheap",
                            "variant_id": "var_cheap",
                            "product_name": "Áo Khoác Gió",
                            "price": 1200000.0,
                            "stock": 5,
                        },
                        {
                            "product_id": "jacket_expensive",
                            "variant_id": "var_expensive",
                            "product_name": "Áo Khoác Da Bò",
                            "price": 2000000.0,
                            "stock": 5,
                        },
                    ]
                },
            },
        )
        assert res.status_code == 200
        data = res.json()
        pids = [p["product_id"] for p in data["recommended_products"]]
        assert "jacket_cheap" in pids
        assert "jacket_expensive" not in pids

