"""Input and Output Guardrails for preventing hallucinations and enforcing business rules."""

import re
from typing import Any
from hexta_ai.schemas.consultant import (
    CatalogItem,
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
    ) -> tuple[StructuredConsultantDecision, list[str]]:
        """Ensures that recommendations and draft orders contain only valid, real prices and products."""
        audit_notes: list[str] = []
        catalog_map = {item.product_id: item for item in hydrated_items}

        # 1. Verify Recommended Products
        valid_recommended_ids: list[str] = []
        for pid in decision.recommended_product_ids:
            if pid in catalog_map:
                valid_recommended_ids.append(pid)
            else:
                audit_notes.append(f"Loại bỏ sản phẩm ảo {pid} không tồn tại trong danh mục thực tế")
        decision.recommended_product_ids = valid_recommended_ids

        # 2. Verify and Correct Draft Order Pricing and SKUs
        if decision.draft_order:
            draft: OrderDraft = decision.draft_order
            corrected_items = []
            calculated_total = 0.0

            for draft_item in draft.items:
                real_item = catalog_map.get(draft_item.product_id)
                if not real_item:
                    audit_notes.append(f"Loại bỏ item {draft_item.product_id} khỏi đơn nháp vì không có trong kho")
                    continue

                # Enforce real database price (prevent hallucinated discounts)
                if abs(draft_item.unit_price - real_item.price) > 0.01:
                    audit_notes.append(
                        f"Tự động sửa giá item {real_item.name} từ {draft_item.unit_price:,.0f}đ thành giá thực tế {real_item.price:,.0f}đ"
                    )
                    draft_item.unit_price = real_item.price

                # Ensure stock availability
                if real_item.stock <= 0:
                    audit_notes.append(f"Sản phẩm {real_item.name} đã hết hàng trong kho, không thể lên đơn")
                    continue

                draft_item.product_name = real_item.name
                item_subtotal = draft_item.unit_price * draft_item.quantity
                calculated_total += item_subtotal
                corrected_items.append(draft_item)

            draft.items = corrected_items
            draft.total_amount = calculated_total

            if not corrected_items:
                decision.draft_order = None
                audit_notes.append("Hủy tạo đơn nháp do không có sản phẩm hợp lệ còn hàng")

        return decision, audit_notes
