"""Catalog search and real-time stock/price hydration service."""

import re
from typing import Any
from hexta_ai.schemas.consultant import CatalogItem, ExtractedFilters


DEFAULT_CATALOG_ITEMS: list[CatalogItem] = [
    CatalogItem(
        product_id="prod_01",
        sku="POLO-BLK-L",
        name="Áo Polo Nam Thể Thao Co Giãn Dry-Fit",
        category="Áo nam",
        price=299000.0,
        original_price=350000.0,
        stock=15,
        attributes={"size": "L", "color": "Đen", "material": "Dry-Fit Poly", "gender": "Nam"},
        description="Áo polo phong cách thể thao, chất vải thấm hút mồ hôi, co giãn 4 chiều, phù hợp mặc hằng ngày hoặc đi chơi thể thao.",
        image_url="https://images.unsplash.com/photo-1581655353564-df123a1eb820?w=400",
    ),
    CatalogItem(
        product_id="prod_02",
        sku="POLO-WHT-M",
        name="Áo Polo Nam Trắng Cổ Điển Regular Fit",
        category="Áo nam",
        price=320000.0,
        original_price=390000.0,
        stock=8,
        attributes={"size": "M", "color": "Trắng", "material": "Cotton pique", "gender": "Nam"},
        description="Chất cotton pique dày dặn, đứng form chuẩn quý ông, dễ phối cùng quần jean hoặc quần tây.",
        image_url="https://images.unsplash.com/photo-1625910513413-7a96435940c3?w=400",
    ),
    CatalogItem(
        product_id="prod_03",
        sku="JEAN-SLIM-31",
        name="Quần Jeans Nam Slim Fit Co Giãn Màu Xanh Đậm",
        category="Quần nam",
        price=450000.0,
        original_price=550000.0,
        stock=20,
        attributes={"size": "31", "color": "Xanh đậm", "fit": "Slim Fit", "gender": "Nam"},
        description="Quần jeans form ôm vừa vặn, vải denim co giãn nhẹ tôn dáng và thoải mái vận động cả ngày.",
        image_url="https://images.unsplash.com/photo-1542272604-780c96856592?w=400",
    ),
    CatalogItem(
        product_id="prod_04",
        sku="SHOE-RUN-42",
        name="Giày Chạy Bộ Thể Thao Đệm Khí Ultra Cushion",
        category="Giày dép",
        price=1250000.0,
        original_price=1500000.0,
        stock=5,
        attributes={"size": "42", "color": "Đen/Cam", "type": "Chạy bộ", "gender": "Unisex"},
        description="Đế giày đệm bọt EVA trợ lực tối ưu, giảm chấn khi chạy đường dài, vải mesh thoáng khí 100%.",
        image_url="https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=400",
    ),
    CatalogItem(
        product_id="prod_05",
        sku="BAG-LEATHER-BRN",
        name="Túi Đeo Chéo Da Bò Vintage Đựng iPad",
        category="Phụ kiện",
        price=650000.0,
        original_price=790000.0,
        stock=0,  # Out of stock to test stock checking
        attributes={"color": "Nâu da bò", "material": "Da bò thật", "type": "Túi đeo chéo"},
        description="Túi đeo chéo thời trang đựng vừa iPad 11 inch và các vật dụng cá nhân, khóa đồng chống gỉ sét.",
        image_url="https://images.unsplash.com/photo-1548036328-c9fa89d128fa?w=400",
    ),
]


class CatalogService:
    """Manages catalog retrieval, keyword filtering, and real-time hydration."""

    def __init__(self, initial_items: list[CatalogItem] | None = None):
        self._items: dict[str, CatalogItem] = {
            item.product_id: item for item in (initial_items or DEFAULT_CATALOG_ITEMS)
        }

    def get_by_id(self, product_id: str) -> CatalogItem | None:
        return self._items.get(product_id)

    def get_by_sku(self, sku: str) -> CatalogItem | None:
        norm_sku = sku.strip().upper()
        for item in self._items.values():
            if item.sku.upper() == norm_sku:
                return item
        return None

    def search(self, query: str, filters: ExtractedFilters | None = None, limit: int = 4) -> list[CatalogItem]:
        """Hybrid-style search: keyword matching + attribute + numerical price range filtering."""
        tokens = [t.lower() for t in re.findall(r"\w+", query) if len(t) > 1]
        results: list[tuple[float, CatalogItem]] = []

        for item in self._items.values():
            # Hard filter: in_stock_only
            if filters and filters.in_stock_only and item.stock <= 0:
                continue

            # Hard filter: price bounds
            if filters:
                if filters.min_price is not None and item.price < filters.min_price:
                    continue
                if filters.max_price is not None and item.price > filters.max_price:
                    continue

            # Text relevance scoring
            text_corpus = f"{item.name} {item.category} {item.description} {' '.join(item.attributes.values())}".lower()
            score = 0.0

            for token in tokens:
                if token in item.name.lower():
                    score += 3.0
                elif token in item.category.lower():
                    score += 2.0
                elif token in text_corpus:
                    score += 1.0

            # Attribute match boost
            if filters and filters.attributes:
                for attr_key, attr_val in filters.attributes.items():
                    if attr_val.lower() in text_corpus:
                        score += 2.5

            if score > 0 or not tokens:
                results.append((score, item))

        # Fallback to top in-stock items if query had no direct keyword matches
        if not results:
            for item in self._items.values():
                if not filters or not filters.in_stock_only or item.stock > 0:
                    results.append((0.1, item))
                    if len(results) >= limit:
                        break

        results.sort(key=lambda x: x[0], reverse=True)
        return [item for _, item in results[:limit]]

    def hydrate_items(self, product_ids: list[str]) -> list[CatalogItem]:
        """Simulates live stock & price hydration from transactional store."""
        hydrated: list[CatalogItem] = []
        for pid in product_ids:
            item = self.get_by_id(pid)
            if item:
                hydrated.append(item.model_copy())
        return hydrated
