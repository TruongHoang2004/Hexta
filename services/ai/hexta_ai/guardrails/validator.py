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


def parse_vietnamese_price_filter(val_str: str, unit: str | None) -> float | None:
    val_str = val_str.strip()
    if not val_str:
        return None

    has_dot = "." in val_str
    has_comma = "," in val_str

    if has_dot and has_comma:
        last_dot = val_str.rfind(".")
        last_comma = val_str.rfind(",")
        if last_dot > last_comma:
            clean_str = val_str[:last_dot].replace(",", "") + "." + val_str[last_dot + 1:]
        else:
            clean_str = val_str[:last_comma].replace(".", "") + "." + val_str[last_comma + 1:]
    elif has_dot or has_comma:
        sep = "." if has_dot else ","
        parts = val_str.split(sep)
        if len(parts) > 2:
            clean_str = "".join(parts)
        else:
            decimal_part = parts[1]
            if unit:
                unit_lower = unit.lower()
                if unit_lower in {"tr", "triệu"} or len(decimal_part) != 3:
                    clean_str = f"{parts[0]}.{parts[1]}"
                else:
                    clean_str = "".join(parts)
            else:
                if len(decimal_part) == 3:
                    clean_str = "".join(parts)
                else:
                    clean_str = f"{parts[0]}.{parts[1]}"
    else:
        clean_str = val_str

    try:
        val = float(clean_str)
    except ValueError:
        return None

    if unit:
        unit_lower = unit.lower()
        if unit_lower in {"tr", "triệu"}:
            val *= 1_000_000.0
        elif unit_lower in {"k", "nghìn", "ngàn"}:
            val *= 1_000.0

    return val


def extract_mentioned_prices(text: str) -> list[float]:
    """Extracts monetary values mentioned in text."""
    prices: list[float] = []
    pattern = r"(?:\b|\bgiá\s*|\bchỉ\s*|\bcòn\s*|\btổng\s*)(\d+(?:[.,]\d+)*)\s*(k|nghìn|ngàn|tr|triệu|đ|vnd|đồng)\b"
    for m in re.finditer(pattern, text, re.IGNORECASE):
        v = parse_vietnamese_price_filter(m.group(1), m.group(2))
        if v is not None and v > 0:
            prices.append(v)
    return prices


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

        # 2. Verify and Correct Draft Order Pricing, SKUs, Stock, Hard Filters, and Total Amount
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

                # Ensure stock availability
                if real_item.stock <= 0:
                    audit_notes.append(f"Sản phẩm {real_item.name} đã hết hàng trong kho, không thể lên đơn")
                    items_removed = True
                    continue

                # Enforce customer price and budget filters on draft items
                if filters:
                    if filters.min_price is not None and real_item.price < filters.min_price:
                        audit_notes.append(
                            f"Loại bỏ item {real_item.name} khỏi đơn nháp vì giá {real_item.price:,.0f}đ < min_price {filters.min_price:,.0f}đ"
                        )
                        items_removed = True
                        continue
                    if filters.max_price is not None and real_item.price > filters.max_price:
                        audit_notes.append(
                            f"Loại bỏ item {real_item.name} khỏi đơn nháp vì giá {real_item.price:,.0f}đ > max_price {filters.max_price:,.0f}đ"
                        )
                        items_removed = True
                        continue

                # Enforce real database price (prevent hallucinated discounts)
                if abs(draft_item.unit_price - real_item.price) > 0.01:
                    price_adjusted = True
                    audit_notes.append(
                        f"Tự động sửa giá item {real_item.name} từ {draft_item.unit_price:,.0f}đ thành giá thực tế {real_item.price:,.0f}đ"
                    )
                    draft_item.unit_price = real_item.price

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
                audit_notes.append("Hủy tạo đơn nháp do không có sản phẩm hợp lệ thỏa mãn điều kiện ngân sách hoặc tồn kho")
                if filters and (filters.max_price is not None or filters.min_price is not None):
                    decision.reply_message = (
                        "Dạ rất tiếc, sản phẩm bạn yêu cầu có mức giá vượt quá ngân sách bạn mong muốn "
                        "nên shop chưa thể tạo đơn hàng nháp ạ. Bạn vui lòng tham khảo các mẫu có mức giá phù hợp hơn nhé!"
                    )
                else:
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

        # 3. Grounding and Fact-checking for reply_message (Prices and Policies)
        if decision.intent == "faq_policy":
            audit_notes.append("Chuẩn hóa thông tin chính sách theo quy định cửa hàng")
            decision.reply_message = (
                "Dạ chính sách của shop: Miễn phí vận chuyển cho đơn hàng từ 500.000đ (phí ship 30.000đ cho đơn dưới 500.000đ) "
                "và hỗ trợ đổi trả trong vòng 7 ngày cho sản phẩm còn nguyên tem mác bạn nhé!"
            )
        else:
            allowed_prices: set[float] = set()
            for it in hydrated_items:
                allowed_prices.add(round(it.price, 2))
            if decision.draft_order:
                allowed_prices.add(round(decision.draft_order.total_amount, 2))
                for it in decision.draft_order.items:
                    allowed_prices.add(round(it.unit_price, 2))
            if filters:
                if filters.min_price is not None:
                    allowed_prices.add(round(filters.min_price, 2))
                if filters.max_price is not None:
                    allowed_prices.add(round(filters.max_price, 2))
            allowed_prices.add(500000.0)
            allowed_prices.add(30000.0)

            mentioned_prices = extract_mentioned_prices(decision.reply_message)
            has_hallucinated_price = False
            for p in mentioned_prices:
                if not any(abs(p - ap) < 1.0 for ap in allowed_prices):
                    has_hallucinated_price = True
                    audit_notes.append(f"Phát hiện giá bịa {p:,.0f}đ trong lời thoại")
                    break

            has_unauthorized_policy = False
            if re.search(r"(?:miễn phí vận chuyển|freeship|free ship|miễn phí giao hàng)", decision.reply_message, re.IGNORECASE):
                qualified = False
                if decision.draft_order and decision.draft_order.total_amount >= 500000.0:
                    qualified = True
                elif re.search(r"500\s*(?:k|nghìn|ngàn|000|\.000|,000)", decision.reply_message, re.IGNORECASE):
                    qualified = True
                if not qualified:
                    has_unauthorized_policy = True
                    audit_notes.append("Phát hiện cam kết miễn phí vận chuyển không đủ điều kiện trong lời thoại")

            if (has_hallucinated_price or has_unauthorized_policy) and not (price_adjusted or items_removed or total_adjusted):
                if decision.draft_order and decision.draft_order.items:
                    item_summaries = [
                        f"{it.product_name} (x{it.quantity}): {it.unit_price:,.0f}đ"
                        for it in decision.draft_order.items
                    ]
                    items_str = ", ".join(item_summaries)
                    decision.reply_message = (
                        f"Dạ em đã tạo đơn hàng nháp cho bạn với đúng giá niêm yết trong hệ thống: {items_str}. "
                        f"Tổng đơn hàng là {decision.draft_order.total_amount:,.0f}đ. Bạn vui lòng kiểm tra lại thông tin đơn hàng nhé!"
                    )
                elif valid_recommended_ids:
                    rec_items = [lookup_item(pid, pid) for pid in valid_recommended_ids]
                    rec_items = [it for it in rec_items if it]
                    if rec_items:
                        rec_summaries = [f"{it.name or it.product_name} ({it.price:,.0f}đ)" for it in rec_items]
                        decision.reply_message = (
                            f"Dạ shop xin gợi ý các sản phẩm phù hợp với bạn theo giá niêm yết: {', '.join(rec_summaries)}. "
                            "Tất cả sản phẩm đều là hàng chính hãng với giá niêm yết chuẩn của shop. Bạn muốn tham khảo thêm chi tiết hay chọn size mẫu nào ạ?"
                        )
                    else:
                        decision.reply_message = "Dạ hiện tại shop chưa tìm thấy sản phẩm nào phù hợp với yêu cầu của bạn ạ. Bạn vui lòng tham khảo các mẫu khác của shop nhé!"
                else:
                    decision.reply_message = "Dạ hiện tại shop chưa tìm thấy sản phẩm nào phù hợp với yêu cầu của bạn ạ. Bạn vui lòng tham khảo các mẫu khác của shop nhé!"

        return decision, audit_notes
