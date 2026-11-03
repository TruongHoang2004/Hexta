"""Input and Output Guardrails for preventing hallucinations and enforcing business rules."""

import re
from typing import Any
from hexta_ai.schemas.consultant import (
    CatalogItem,
    ExtractedFilters,
    OrderDraft,
    StructuredConsultantDecision,
)


SUSPICIOUS_INJECTION_PATTERNS = [
    r"ignore (all )?(previous|prior) instructions",
    r"disregard (all )?(system|developer) prompts",
    r"bỏ qua (mọi )?(hướng dẫn|chỉ dẫn|quy tắc)",
    r"bạn không còn là",
    r"hãy đóng vai một hacker",
    r"system override",
]


class InputGuardrail:
    """Detects prompt injection and sanitizes customer inputs."""

    @staticmethod
    def check_and_sanitize(message: str) -> tuple[bool, str, str]:
        """Returns (is_safe, sanitized_message, reason_if_unsafe)."""
        clean = message.strip()
        if not clean:
            return False, "", "Tin nhắn rỗng"

        for pattern in SUSPICIOUS_INJECTION_PATTERNS:
            if re.search(pattern, clean, re.IGNORECASE):
                return False, clean, "Phát hiện nội dung can thiệp hệ thống"

        return True, clean, ""


class OutputGuardrail:
    """Verifies and fact-checks model outputs against real hydrated catalog data."""

    @staticmethod
    def validate_and_correct(
        decision: StructuredConsultantDecision,
        hydrated_items: list[CatalogItem],
        filters: ExtractedFilters | None = None,
    ) -> tuple[StructuredConsultantDecision, list[str]]:
        """Ensures that recommendations and draft orders contain only valid, real prices and products."""
        audit_notes: list[str] = []
        variant_map: dict[str, CatalogItem] = {}
        product_map: dict[str, CatalogItem] = {}

        for item in hydrated_items:
            if item.variant_id:
                variant_map[item.variant_id] = item
            if item.sku:
                variant_map[item.sku] = item
            if item.product_id not in product_map:
                product_map[item.product_id] = item

        def lookup_item(vid: str | None, pid: str | None) -> CatalogItem | None:
            # 1. Prioritize variant_id lookup first
            if vid and vid in variant_map:
                return variant_map[vid]
            if vid and vid in product_map:
                return product_map[vid]
            # 2. Lookup by pid (check variant_map then product_map)
            if pid and pid in variant_map:
                return variant_map[pid]
            if pid and pid in product_map:
                return product_map[pid]
            return None

        # 1. Verify Recommended Products against Catalog and Hard Filters
        valid_recommended_ids: list[str] = []
        for pid in decision.recommended_product_ids:
            real_item = lookup_item(pid, pid)
            if not real_item:
                audit_notes.append(f"Loại bỏ sản phẩm ảo {pid} không tồn tại trong danh mục thực tế")
                continue

            # Re-apply hard filters to recommendations
            if filters:
                if filters.in_stock_only and real_item.stock <= 0:
                    audit_notes.append(f"Loại bỏ sản phẩm {real_item.name} vì đã hết hàng")
                    continue
                if filters.min_price is not None and real_item.price < filters.min_price:
                    audit_notes.append(
                        f"Loại bỏ sản phẩm {real_item.name} vì giá {real_item.price:,.0f}đ < min_price {filters.min_price:,.0f}đ"
                    )
                    continue
                if filters.max_price is not None and real_item.price > filters.max_price:
                    audit_notes.append(
                        f"Loại bỏ sản phẩm {real_item.name} vì giá {real_item.price:,.0f}đ > max_price {filters.max_price:,.0f}đ"
                    )
                    continue

            # Keep variant identifier if available to preserve variant identity across services
            target_id = real_item.variant_id or real_item.product_id
            valid_recommended_ids.append(target_id)

        decision.recommended_product_ids = valid_recommended_ids

        # 2. Verify and Correct Draft Order Pricing, SKUs and Total Amount
        price_adjusted = False
        items_removed = False
        total_adjusted = False
        if decision.draft_order:
            draft: OrderDraft = decision.draft_order
            original_total = draft.total_amount
            corrected_items = []
            calculated_total = 0.0

            for draft_item in draft.items:
                # Prioritize variant_id lookup over product_id
                real_item = lookup_item(draft_item.variant_id, draft_item.product_id)
                if not real_item:
                    audit_notes.append(f"Loại bỏ item {draft_item.product_id} khỏi đơn nháp vì không có trong kho")
                    items_removed = True
                    continue

                # Enforce real database price (prevent hallucinated discounts)
                if abs(draft_item.unit_price - real_item.price) > 0.01:
                    price_adjusted = True
                    audit_notes.append(
                        f"Tự động sửa giá item {real_item.name} từ {draft_item.unit_price:,.0f}đ thành giá thực tế {real_item.price:,.0f}đ"
                    )
                    draft_item.unit_price = real_item.price

                # Ensure stock availability
                if real_item.stock <= 0:
                    audit_notes.append(f"Sản phẩm {real_item.name} đã hết hàng trong kho, không thể lên đơn")
                    items_removed = True
                    continue

                draft_item.product_name = real_item.name or real_item.product_name
                if real_item.variant_id:
                    draft_item.variant_id = real_item.variant_id
                if real_item.product_id:
                    draft_item.product_id = real_item.product_id
                item_subtotal = draft_item.unit_price * draft_item.quantity
                calculated_total += item_subtotal
                corrected_items.append(draft_item)

            draft.items = corrected_items
            if abs(calculated_total - original_total) > 0.01:
                total_adjusted = True
                audit_notes.append(
                    f"Tự động sửa tổng tiền đơn nháp từ {original_total:,.0f}đ thành {calculated_total:,.0f}đ"
                )
            draft.total_amount = calculated_total

            if not corrected_items:
                decision.draft_order = None
                audit_notes.append("Hủy tạo đơn nháp do không có sản phẩm hợp lệ còn hàng")
                decision.reply_message = (
                    "Dạ rất tiếc, sản phẩm bạn yêu cầu hiện không có trong danh mục hoặc đã hết hàng "
                    "nên shop chưa thể tạo đơn hàng nháp ạ. Bạn vui lòng tham khảo các mẫu khác nhé!"
                )
            elif price_adjusted or items_removed or total_adjusted:
                # Reconstruct reply_message to eliminate hallucinated prices or totals from the text
                item_summaries = [
                    f"{it.product_name} (x{it.quantity}): {it.unit_price:,.0f}đ"
                    for it in draft.items
                ]
                items_str = ", ".join(item_summaries)
                decision.reply_message = (
                    f"Dạ em đã tạo đơn hàng nháp cho bạn với đúng giá niêm yết trong hệ thống: {items_str}. "
                    f"Tổng đơn hàng là {draft.total_amount:,.0f}đ. Bạn vui lòng kiểm tra lại thông tin đơn hàng nhé!"
                )

        return decision, audit_notes
