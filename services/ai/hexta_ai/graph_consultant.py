"""Production-grade Shopping Consultant State Machine powered by LangGraph & LangChain."""

import json
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
    retrieved_items: list[CatalogItem]
    decision: StructuredConsultantDecision
    final_response: ConsultantResponse
    steps: list[str]


SYSTEM_CONSULTANT_PROMPT = """Bạn là Hexta AI - Trợ lý tư vấn bán hàng thông minh, tận tâm và chuyên nghiệp của cửa hàng.
Quy tắc cốt lõi:
1. Giao tiếp bằng tiếng Việt tự nhiên, lịch sự, thân thiện và súc tích.
2. Dựa HOÀN TOÀN vào danh mục sản phẩm được cung cấp bên dưới để tư vấn. TUYỆT ĐỐI KHÔNG tự bịa ra sản phẩm, giá bán hoặc chính sách khuyến mãi không có trong dữ liệu.
3. Khi khách hàng có ý định chốt đơn hoặc yêu cầu mua hàng, hãy tạo cấu trúc `draft_order` với đúng thông tin sản phẩm và giá trong kho.
4. Nếu khách hàng cần hỗ trợ đặc biệt, khiếu nại gay gắt hoặc yêu cầu gặp người thật, hãy đánh dấu `needs_human: true`.
5. Đưa ra các câu gợi ý ngắn gọn tiếp theo (`quick_replies`) để khách dễ dàng tiếp tục trò chuyện.

DANH MỤC SẢN PHẨM HIỆN CÓ TRONG KHO:
{catalog_context}
"""


def create_consultant_graph(catalog_service: CatalogService | None = None, provider: str | None = None):
    service = catalog_service or CatalogService()

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
        query = state["sanitized_message"]
        
        # Simple extraction heuristics for filter bounds
        min_p = None
        max_p = None
        if "dưới 300k" in query.lower() or "dưới 300.000" in query.lower():
            max_p = 300000.0
        elif "dưới 500k" in query.lower() or "dưới 500.000" in query.lower():
            max_p = 500000.0
        elif "trên 1 triệu" in query.lower() or "trên 1tr" in query.lower():
            min_p = 1000000.0

        filters = ExtractedFilters(min_price=min_p, max_price=max_p)
        items = service.search(query=query, filters=filters, limit=4)
        return {
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
                f"- [Mã: {it.product_id} | SKU: {it.sku}] {it.name} | Giá: {it.price:,.0f}đ ({status}) | Thuộc tính: {json.dumps(it.attributes, ensure_ascii=False)} | Mô tả: {it.description}"
            )
        catalog_context = "\n".join(catalog_lines) if catalog_lines else "Hiện không tìm thấy sản phẩm nào khớp chính xác."

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
            # Fallback for unexpected parse error
            decision = StructuredConsultantDecision(
                intent="discovery",
                thought_process="Model invoke fallback",
                reply_message="Dạ chào bạn, bạn cần shop tư vấn mẫu áo, quần hay giày thể thao nào ạ?",
                recommended_product_ids=[it.product_id for it in items[:2]],
                quick_replies=["Xem áo polo", "Xem quần jean", "Xem giày chạy bộ"],
            )

        return {"decision": decision, "steps": steps}

    async def guardrail_output_node(state: ConsultantState) -> dict:
        steps = [*state["steps"], "guardrail_output"]
        req = state["request"]
        decision = state["decision"]
        items = list(state.get("retrieved_items", []))

        # Ensure all items referenced by decision are hydrated from catalog service
        items_dict = {it.product_id: it for it in items}
        if decision.draft_order:
            for d_item in decision.draft_order.items:
                if d_item.product_id not in items_dict:
                    found = service.get_by_id(d_item.product_id)
                    if found:
                        items.append(found)
                        items_dict[found.product_id] = found

        for pid in decision.recommended_product_ids:
            if pid not in items_dict:
                found = service.get_by_id(pid)
                if found:
                    items.append(found)
                    items_dict[found.product_id] = found

        # Validate against real products and prices
        validated_decision, audit_notes = OutputGuardrail.validate_and_correct(
            decision=decision,
            hydrated_items=items,
        )

        catalog_map = {it.product_id: it for it in items}
        recommended_items = [catalog_map[pid] for pid in validated_decision.recommended_product_ids if pid in catalog_map]

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
