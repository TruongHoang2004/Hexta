"""Standardized multi-model adapter using LangChain with strict structured output enforcement."""

import json
import os
import re
from typing import Any, Type, TypeVar
from pydantic import BaseModel
from langchain_core.language_models.chat_models import BaseChatModel
from langchain_core.messages import AIMessage, BaseMessage
from langchain_core.outputs import ChatGeneration, ChatResult

from hexta_ai.schemas.consultant import (
    StructuredConsultantDecision,
    OrderDraft,
    OrderDraftItem,
)

T = TypeVar("T", bound=BaseModel)


class FakeStructuredChatModel(BaseChatModel):
    """Deterministic simulation model for offline tests, CI, and local sandbox verification."""
    
    def _generate(
        self,
        messages: list[BaseMessage],
        stop: list[str] | None = None,
        run_manager: Any = None,
        **kwargs: Any,
    ) -> ChatResult:
        # Extract catalog context from messages if present
        catalog_context = ""
        for m in messages:
            if hasattr(m, "content"):
                c = str(m.content)
                if "DANH MỤC SẢN PHẨM" in c:
                    catalog_context = c
                    break

        last_msg = messages[-1].content if messages else ""
        text = str(last_msg).lower()

        # Parse available products from catalog_context
        # Format in prompt: "- [Mã: <id> | SKU: <sku>] <name> | Giá: <price>đ ..."
        available_products: list[dict[str, Any]] = []
        for match in re.finditer(
            r"- \[Mã:\s*([^ |]+)\s*\|\s*SKU:\s*([^ \]]+)\]\s*([^|]+)\|\s*Giá:\s*([\d\.,]+)đ",
            catalog_context,
        ):
            pid = match.group(1).strip()
            sku = match.group(2).strip()
            pname = match.group(3).strip()
            price_str = match.group(4).replace(".", "").replace(",", "")
            try:
                price = float(price_str)
            except ValueError:
                price = 0.0
            available_products.append({"id": pid, "sku": sku, "name": pname, "price": price})

        # Price constraint parsing from query
        max_p = None
        if "dưới 300k" in text or "dưới 300.000" in text:
            max_p = 300000.0
        elif "max_price=1000" in text or "dưới 1000" in text:
            max_p = 1000.0

        if max_p is not None:
            available_products = [p for p in available_products if p["price"] <= max_p]

        has_catalog_context = "DANH MỤC SẢN PHẨM" in catalog_context
        if not has_catalog_context:
            from hexta_ai.catalog.service import DEFAULT_CATALOG_ITEMS
            available_products = [
                {"id": it.product_id, "sku": it.variant_id or it.sku or it.product_id, "name": it.name, "price": it.price}
                for it in DEFAULT_CATALOG_ITEMS
            ]
            catalog_empty = False
        else:
            catalog_empty = (
                "Hiện tại cửa hàng không có sản phẩm nào" in catalog_context
                or "Hiện không tìm thấy" in catalog_context
                or not available_products
            )

        # Match specific variant in user text if multiple items available
        matching_products = []
        for p in available_products:
            p_text = f"{p['name']} {p['sku']}".lower()
            if "size l" in text and ("size l" in p_text or "size_l" in p_text or "polo-l" in p_text):
                matching_products.append(p)
            elif "size s" in text and ("size s" in p_text or "size_s" in p_text or "polo-s" in p_text):
                matching_products.append(p)
            elif "size m" in text and ("size m" in p_text or "size_m" in p_text or "polo-m" in p_text):
                matching_products.append(p)

        target_pool = matching_products if matching_products else available_products

        # Rule-based decision for fake simulation
        if any(w in text for w in ["người", "nhân viên", "tổng đài", "gặp tư vấn viên"]):
            decision = StructuredConsultantDecision(
                intent="human_handoff",
                thought_process="Customer requested live human staff support.",
                reply_message="Dạ em đang kết nối bạn với chuyên viên tư vấn của shop, bạn vui lòng đợi trong giây lát ạ!",
                needs_human=True,
                quick_replies=["Quay lại chat tự động"],
            )
        elif catalog_empty:
            decision = StructuredConsultantDecision(
                intent="discovery",
                thought_process="No matching products available in catalog context.",
                reply_message="Dạ hiện tại shop chưa có sản phẩm nào phù hợp với yêu cầu của bạn ạ.",
                recommended_product_ids=[],
                draft_order=None,
                quick_replies=["Xem sản phẩm khác"],
            )
        elif any(w in text for w in ["chốt đơn", "mua", "lấy cái này", "đặt hàng"]):
            chosen = target_pool[0]
            chosen_ref = chosen["sku"] if (chosen["sku"] and chosen["sku"] != chosen["id"]) else chosen["id"]
            decision = StructuredConsultantDecision(
                intent="order_draft",
                thought_process="Customer showed explicit purchase intent. Formulating draft order.",
                reply_message=f"Dạ em đã tạo đơn hàng nháp cho bạn với sản phẩm {chosen['name']}. Bạn vui lòng kiểm tra lại thông tin đơn hàng bên dưới nhé!",
                recommended_product_ids=[chosen_ref],
                draft_order=OrderDraft(
                    draft_id="draft_sim_001",
                    customer_name="Nguyễn Văn A",
                    customer_phone="0912345678",
                    shipping_address="Hà Nội",
                    items=[
                        OrderDraftItem(
                            product_id=chosen["id"],
                            variant_id=chosen["sku"] if (chosen["sku"] and chosen["sku"] != chosen["id"]) else None,
                            product_name=chosen["name"],
                            variant="Mặc định",
                            quantity=1,
                            unit_price=chosen["price"],
                        )
                    ],
                    total_amount=chosen["price"],
                    payment_method="cod",
                ),
                quick_replies=["Xác nhận đơn", "Sửa địa chỉ", "Hủy"],
            )
        elif any(w in text for w in ["đổi trả", "bảo hành", "vận chuyển", "freeship"]):
            decision = StructuredConsultantDecision(
                intent="faq_policy",
                thought_process="Customer asked store policy FAQ.",
                reply_message="Shop hỗ trợ đổi trả miễn phí trong vòng 7 ngày và miễn phí vận chuyển cho đơn hàng từ 500.000đ bạn nhé!",
                quick_replies=["Xem sản phẩm hot", "Tư vấn chọn size"],
            )
        else:
            rec_ids = [
                p["sku"] if (p["sku"] and p["sku"] != p["id"]) else p["id"]
                for p in target_pool[:2]
            ]
            decision = StructuredConsultantDecision(
                intent="discovery",
                thought_process="Customer browsing or asking for recommendations.",
                reply_message="Dạ chào bạn, shop xin gợi ý các mẫu sản phẩm phù hợp với yêu cầu của bạn ạ!",
                recommended_product_ids=rec_ids,
                quick_replies=["Xem thêm mẫu khác", "Lọc theo giá", "Chọn size"],
            )
        
        json_output = decision.model_dump_json()
        message = AIMessage(content=json_output)
        return ChatResult(generations=[ChatGeneration(message=message)])

    @property
    def _llm_type(self) -> str:
        return "fake_structured_model"

    def with_structured_output(self, schema: Type[T], **kwargs: Any):
        """Emulates LangChain's structured output mechanism."""
        class _BoundStructuredRunnable:
            def __init__(self, model: "FakeStructuredChatModel", schema_cls: Type[T]):
                self.model = model
                self.schema_cls = schema_cls

            async def ainvoke(self, input_data: Any, config: Any = None) -> T:
                if isinstance(input_data, list):
                    msgs = input_data
                elif isinstance(input_data, dict) and "messages" in input_data:
                    msgs = input_data["messages"]
                else:
                    msgs = [AIMessage(content=str(input_data))]
                res = self.model._generate(msgs)
                raw_json = res.generations[0].message.content
                return self.schema_cls.model_validate_json(raw_json)

            def invoke(self, input_data: Any, config: Any = None) -> T:
                if isinstance(input_data, list):
                    msgs = input_data
                elif isinstance(input_data, dict) and "messages" in input_data:
                    msgs = input_data["messages"]
                else:
                    msgs = [AIMessage(content=str(input_data))]
                res = self.model._generate(msgs)
                raw_json = res.generations[0].message.content
                return self.schema_cls.model_validate_json(raw_json)

        return _BoundStructuredRunnable(self, schema)


class ModelAdapter:
    """Factory and manager for multi-provider LLMs with structured output governance."""

    @staticmethod
    def get_model(
        provider: str | None = None,
        model_name: str | None = None,
        temperature: float = 0.2,
        **kwargs: Any,
    ) -> BaseChatModel:
        prov = (provider or os.getenv("AI_PROVIDER", "fake")).strip().lower()

        # Configurable limits with safe defaults: 0 retries on 429/failures and 30s timeout
        timeout = float(kwargs.get("timeout") or kwargs.get("request_timeout") or os.getenv("AI_TIMEOUT_SECONDS", "30.0"))
        max_retries = int(kwargs.get("max_retries") if kwargs.get("max_retries") is not None else os.getenv("AI_MAX_RETRIES", "0"))
        max_tokens = int(kwargs.get("max_tokens") or os.getenv("AI_MAX_TOKENS", "2048"))

        if prov == "fake":
            return FakeStructuredChatModel()

        if prov in {"openai", "chatgpt"}:
            from langchain_openai import ChatOpenAI
            api_key = kwargs.get("api_key") or os.getenv("OPENAI_API_KEY")
            base_url = kwargs.get("base_url") or os.getenv("OPENAI_API_BASE") or os.getenv("AI_OPENAI_BASE_URL")
            selected_model = model_name or os.getenv("AI_MODEL_NAME", "gpt-4o-mini")
            clean_kwargs = {
                k: v for k, v in kwargs.items()
                if k not in {"api_key", "base_url", "timeout", "request_timeout", "max_retries", "max_tokens"}
            }
            return ChatOpenAI(
                model=selected_model,
                api_key=api_key,
                base_url=base_url,
                temperature=temperature,
                timeout=timeout,
                max_retries=max_retries,
                max_tokens=max_tokens,
                **clean_kwargs,
            )

        if prov in {"google", "gemini"}:
            from langchain_google_genai import ChatGoogleGenerativeAI
            api_key = kwargs.get("api_key") or os.getenv("GOOGLE_API_KEY") or os.getenv("AI_API_KEY")
            selected_model = model_name or os.getenv("AI_MODEL_NAME", "gemini-2.0-flash")
            clean_kwargs = {
                k: v for k, v in kwargs.items()
                if k not in {"api_key", "timeout", "request_timeout", "max_retries", "max_tokens"}
            }
            return ChatGoogleGenerativeAI(
                model=selected_model,
                google_api_key=api_key,
                temperature=temperature,
                timeout=timeout,
                max_retries=max_retries,
                max_output_tokens=max_tokens,
                **clean_kwargs,
            )

        if prov in {"ollama", "local"}:
            from langchain_openai import ChatOpenAI
            base_url = kwargs.get("base_url") or os.getenv("OLLAMA_BASE_URL", "http://localhost:11434/v1")
            selected_model = model_name or os.getenv("AI_MODEL_NAME", "llama3.2")
            clean_kwargs = {
                k: v for k, v in kwargs.items()
                if k not in {"api_key", "base_url", "timeout", "request_timeout", "max_retries", "max_tokens"}
            }
            return ChatOpenAI(
                model=selected_model,
                api_key="ollama",
                base_url=base_url,
                temperature=temperature,
                timeout=timeout,
                max_retries=max_retries,
                max_tokens=max_tokens,
                **clean_kwargs,
            )

        raise ValueError(
            f"Unsupported AI_PROVIDER '{prov}'. Supported providers: 'fake', 'openai', 'google', 'ollama'."
        )

    @staticmethod
    def create_structured_consultant_runnable(
        provider: str | None = None,
        model_name: str | None = None,
        schema: Type[T] = StructuredConsultantDecision,
        **kwargs: Any,
    ):
        """Returns a LangChain runnable guaranteed to produce the requested typed schema."""
        model = ModelAdapter.get_model(provider=provider, model_name=model_name, **kwargs)
        return model.with_structured_output(schema)
