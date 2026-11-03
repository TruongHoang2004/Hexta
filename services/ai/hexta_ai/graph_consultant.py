"""Production-grade Shopping Consultant State Machine powered by LangGraph & LangChain."""

import json
import re
from typing import Any, TypedDict
from langgraph.graph import END, START, StateGraph
from langchain_core.messages import AIMessage, HumanMessage, SystemMessage

from hexta_ai.catalog.service import CatalogService
from hexta_ai.guardrails.validator import InputGuardrail, OutputGuardrail
from hexta_ai.models.adapter import ModelAdapter
from hexta_ai.schemas.consultant import (
    CatalogItem,
    ConsultantRequest,
    ConsultantResponse,
    ExtractedFilters,
    StructuredConsultantDecision,
)


class ConsultantState(TypedDict, total=False):
    request: ConsultantRequest
    sanitized_message: str
    is_safe: bool
    rejection_reason: str
    filters: ExtractedFilters
    catalog_service: CatalogService
    retrieved_items: list[CatalogItem]
    decision: StructuredConsultantDecision
    final_response: ConsultantResponse
    steps: list[str]


SYSTEM_CONSULTANT_PROMPT = """Bạn là Hexta AI - Trợ lý tư vấn bán hàng thông minh, tận tâm và chuyên nghiệp của cửa hàng.
Quy tắc cốt lõi:
1. Giao tiếp bằng tiếng Việt tự nhiên, lịch sự, thân thiện và súc tích.
2. Dựa HOÀN TOÀN vào danh mục sản phẩm được cung cấp bên dưới để tư vấn. TUYỆT ĐỐI KHÔNG tự bịa ra sản phẩm, giá bán hoặc chính sách khuyến mãi không có trong dữ liệu.
3. Nếu danh mục sản phẩm hiện không có sản phẩm nào hoặc trống, hãy lịch sự thông báo cho khách rằng cửa hàng hiện chưa có sản phẩm phù hợp. TUYỆT ĐỐI không tự tạo ra sản phẩm ảo.
4. Khi khách hàng có ý định chốt đơn hoặc yêu cầu mua hàng, hãy tạo cấu trúc `draft_order` với đúng thông tin sản phẩm và giá trong kho.
5. Nếu khách hàng cần hỗ trợ đặc biệt, khiếu nại gay gắt hoặc yêu cầu gặp người thật, hãy đánh dấu `needs_human: true`.
6. Đưa ra các câu gợi ý ngắn gọn tiếp theo (`quick_replies`) để khách dễ dàng tiếp tục trò chuyện.

DANH MỤC SẢN PHẨM HIỆN CÓ TRONG KHO:
{catalog_context}
"""


def extract_price_filters(query: str, context: dict[str, Any] | None = None) -> ExtractedFilters:
    q = query.lower()
    min_p: float | None = None
    max_p: float | None = None

    if context:
        if "min_price" in context:
            try:
                min_p = float(context["min_price"])
            except (ValueError, TypeError):
                pass
        if "max_price" in context:
            try:
                max_p = float(context["max_price"])
            except (ValueError, TypeError):
                pass

    if max_p is None:
        m_max = re.search(r"(?:dưới|<|nhỏ hơn|max_price\s*[:=])\s*([\d\.,]+)\s*(k|nghìn|ngàn|tr|triệu)?", q)
        if m_max:
            val_s = m_max.group(1).replace(".", "").replace(",", "")
            unit = m_max.group(2)
            try:
                v = float(val_s)
                if unit in {"tr", "triệu"}:
                    max_p = v * 1000000.0
                elif unit in {"k", "nghìn", "ngàn"}:
                    max_p = v * 1000.0
                else:
                    max_p = v
            except ValueError:
                pass

    if min_p is None:
        m_min = re.search(r"(?:trên|>|lớn hơn|từ|min_price\s*[:=])\s*([\d\.,]+)\s*(k|nghìn|ngàn|tr|triệu)?", q)
        if m_min:
            val_s = m_min.group(1).replace(".", "").replace(",", "")
            unit = m_min.group(2)
            try:
                v = float(val_s)
                if unit in {"tr", "triệu"}:
                    min_p = v * 1000000.0
                elif unit in {"k", "nghìn", "ngàn"}:
                    min_p = v * 1000.0
                else:
                    min_p = v
            except ValueError:
                pass

    return ExtractedFilters(min_price=min_p, max_price=max_p)


def create_consultant_graph(catalog_service: CatalogService | None = None, provider: str | None = None):
    async def guardrail_input_node(state: ConsultantState) -> dict:
        req = state["request"]
        is_safe, sanitized, reason = InputGuardrail.check_and_sanitize(req.message)
        steps = [*state.get("steps", []), "guardrail_input"]
        return {
            "sanitized_message": sanitized,
            "is_safe": is_safe,
            "rejection_reason": reason,
            "steps": steps,
        }

    def check_safety_edge(state: ConsultantState) -> str:
        return "retrieve_catalog" if state["is_safe"] else "unsafe_refusal"

    async def unsafe_refusal_node(state: ConsultantState) -> dict:
        steps = [*state["steps"], "unsafe_refusal"]
        req = state["request"]
        resp = ConsultantResponse(
            tenant_id=req.tenant_id,
            session_id=req.session_id,
            message="Xin lỗi, yêu cầu của bạn chứa nội dung không hợp lệ hoặc không an toàn. Vui lòng hỏi thông tin về sản phẩm của cửa hàng ạ!",
            intent="human_handoff",
            recommended_products=[],
            quick_replies=["Xem sản phẩm hot", "Tư vấn chọn size"],
            needs_human=True,
            reason=state.get("rejection_reason", "Unsafe input detected"),
            steps=steps,
        )
        return {"final_response": resp, "steps": steps}

    async def retrieve_catalog_node(state: ConsultantState) -> dict:
        steps = [*state["steps"], "retrieve_catalog"]
        req = state["request"]
        query = state["sanitized_message"]

        # Dynamically instantiate CatalogService from tenant products in customer_context if provided
        active_catalog = catalog_service or CatalogService.from_customer_context(req.customer_context)
        filters = extract_price_filters(query, req.customer_context)
        items = active_catalog.search(query=query, filters=filters, limit=4)

        return {
            "catalog_service": active_catalog,
            "filters": filters,
            "retrieved_items": items,
            "steps": steps,
        }

    async def generate_consultation_node(state: ConsultantState) -> dict:
        steps = [*state["steps"], "generate_consultation"]
        req = state["request"]
        items = state.get("retrieved_items", [])

        # Format catalog context for prompt
        catalog_lines = []
        for it in items:
            status = f"Còn {it.stock} sp" if it.stock > 0 else "HẾT HÀNG"
            catalog_lines.append(
                f"- [Mã: {it.product_id} | SKU: {it.sku or it.variant_id or it.product_id}] {it.name or it.product_name} | Giá: {it.price:,.0f}đ ({status}) | Thuộc tính: {json.dumps(it.attributes, ensure_ascii=False)} | Mô tả: {it.description}"
            )
        catalog_context = "\n".join(catalog_lines) if catalog_lines else "Hiện tại cửa hàng không có sản phẩm nào trong kho phù hợp với yêu cầu."

        system_msg = SystemMessage(content=SYSTEM_CONSULTANT_PROMPT.format(catalog_context=catalog_context))

        # Format conversation history
        messages: list[Any] = [system_msg]
        for h in req.history[-6:]:
            if h.get("role") == "user":
                messages.append(HumanMessage(content=h.get("content", "")))
            elif h.get("role") == "assistant":
                messages.append(AIMessage(content=h.get("content", "")))

        messages.append(HumanMessage(content=state["sanitized_message"]))

        # Invoke model with structured output
        runnable = ModelAdapter.create_structured_consultant_runnable(provider=provider)
        try:
            decision: StructuredConsultantDecision = await runnable.ainvoke(messages)
        except Exception:
            # Fallback for unexpected error
            if items:
                decision = StructuredConsultantDecision(
                    intent="discovery",
                    thought_process="Model invoke fallback with catalog items",
                    reply_message="Dạ chào bạn, bạn cần shop tư vấn sản phẩm nào ạ?",
                    recommended_product_ids=[it.product_id for it in items[:2]],
                    quick_replies=["Xem chi tiết", "Chọn size"],
                )
            else:
                decision = StructuredConsultantDecision(
                    intent="discovery",
                    thought_process="Model invoke fallback with empty catalog",
                    reply_message="Dạ hiện tại shop chưa có sản phẩm nào phù hợp với yêu cầu của bạn ạ.",
                    recommended_product_ids=[],
                    draft_order=None,
                    quick_replies=["Xem sản phẩm khác"],
                )

        return {"decision": decision, "steps": steps}

    async def guardrail_output_node(state: ConsultantState) -> dict:
        steps = [*state["steps"], "guardrail_output"]
        req = state["request"]
        decision = state["decision"]
        filters = state.get("filters")
        active_catalog = state.get("catalog_service") or catalog_service or CatalogService()
        items = list(state.get("retrieved_items", []))

        # Ensure all items referenced by decision are hydrated from catalog service
        items_dict: dict[str, CatalogItem] = {}
        for it in items:
            items_dict[it.product_id] = it
            if it.variant_id:
                items_dict[it.variant_id] = it
            if it.sku:
                items_dict[it.sku] = it

        if decision.draft_order:
            for d_item in decision.draft_order.items:
                ref_id = d_item.variant_id or d_item.product_id
                if ref_id not in items_dict:
                    found = active_catalog.get_by_id(ref_id)
                    if found:
                        items.append(found)
                        items_dict[found.product_id] = found
                        if found.variant_id:
                            items_dict[found.variant_id] = found

        for pid in decision.recommended_product_ids:
            if pid not in items_dict:
                found = active_catalog.get_by_id(pid)
                if found:
                    # ONLY hydrate if it strictly conforms to hard filters
                    if filters:
                        if filters.min_price is not None and found.price < filters.min_price:
                            continue
                        if filters.max_price is not None and found.price > filters.max_price:
                            continue
                        if filters.in_stock_only and found.stock <= 0:
                            continue
                    items.append(found)
                    items_dict[found.product_id] = found
                    if found.variant_id:
                        items_dict[found.variant_id] = found

        # Validate against real products, prices, and price/stock filters
        validated_decision, audit_notes = OutputGuardrail.validate_and_correct(
            decision=decision,
            hydrated_items=items,
            filters=filters,
        )

        catalog_map = {it.product_id: it for it in items}
        for it in items:
            if it.variant_id:
                catalog_map[it.variant_id] = it

        recommended_items: list[CatalogItem] = []
        for pid in validated_decision.recommended_product_ids:
            if pid in catalog_map:
                it = catalog_map[pid]
                if filters:
                    if filters.min_price is not None and it.price < filters.min_price:
                        continue
                    if filters.max_price is not None and it.price > filters.max_price:
                        continue
                    if filters.in_stock_only and it.stock <= 0:
                        continue
                recommended_items.append(it)

        resp = ConsultantResponse(
            tenant_id=req.tenant_id,
            session_id=req.session_id,
            message=validated_decision.reply_message,
            intent=validated_decision.intent,
            recommended_products=recommended_items,
            order_draft=validated_decision.draft_order,
            quick_replies=validated_decision.quick_replies,
            needs_human=validated_decision.needs_human,
            reason="; ".join(audit_notes) if audit_notes else validated_decision.thought_process,
            steps=steps,
        )
        return {"final_response": resp, "steps": steps}

    builder = StateGraph(ConsultantState)
    builder.add_node("guardrail_input", guardrail_input_node)
    builder.add_node("unsafe_refusal", unsafe_refusal_node)
    builder.add_node("retrieve_catalog", retrieve_catalog_node)
    builder.add_node("generate_consultation", generate_consultation_node)
    builder.add_node("guardrail_output", guardrail_output_node)

    builder.add_edge(START, "guardrail_input")
    builder.add_conditional_edges("guardrail_input", check_safety_edge)
    builder.add_edge("unsafe_refusal", END)
    builder.add_edge("retrieve_catalog", "generate_consultation")
    builder.add_edge("generate_consultation", "guardrail_output")
    builder.add_edge("guardrail_output", END)

    return builder.compile()


consultant_graph = create_consultant_graph()
