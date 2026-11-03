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
        last_msg = messages[-1].content if messages else ""
        text = str(last_msg).lower()
        
        # Rule-based decision for fake simulation
        if any(w in text for w in ["chốt đơn", "mua", "lấy cái này", "đặt hàng"]):
            decision = StructuredConsultantDecision(
                intent="order_draft",
                thought_process="Customer showed explicit purchase intent. Formulating draft order.",
                reply_message="Dạ em đã tạo đơn hàng nháp cho bạn. Bạn vui lòng kiểm tra lại thông tin đơn hàng bên dưới nhé!",
                recommended_product_ids=["prod_01"],
                draft_order=OrderDraft(
                    draft_id="draft_sim_001",
                    customer_name="Nguyễn Văn A",
                    customer_phone="0912345678",
                    shipping_address="Hà Nội",
                    items=[
                        OrderDraftItem(
                            product_id="prod_01",
                            product_name="Áo Polo Thể Thao Nam Co Giãn",
                            variant="Size L / Đen",
                            quantity=1,
                            unit_price=299000.0,
                        )
                    ],
                    total_amount=299000.0,
                    payment_method="cod",
                ),
                quick_replies=["Xác nhận đơn", "Sửa địa chỉ", "Hủy"],
            )
        elif any(w in text for w in ["người", "nhân viên", "tổng đài", "gặp tư vấn viên"]):
            decision = StructuredConsultantDecision(
                intent="human_handoff",
                thought_process="Customer requested live human staff support.",
                reply_message="Dạ em đang kết nối bạn với chuyên viên tư vấn của shop, bạn vui lòng đợi trong giây lát ạ!",
                needs_human=True,
                quick_replies=["Quay lại chat tự động"],
            )
        elif any(w in text for w in ["đổi trả", "bảo hành", "vận chuyển", "freeship"]):
            decision = StructuredConsultantDecision(
                intent="faq_policy",
                thought_process="Customer asked store policy FAQ.",
                reply_message="Shop hỗ trợ đổi trả miễn phí trong vòng 7 ngày và miễn phí vận chuyển cho đơn hàng từ 500.000đ bạn nhé!",
                quick_replies=["Xem sản phẩm hot", "Tư vấn chọn size"],
            )
        else:
            decision = StructuredConsultantDecision(
                intent="discovery",
                thought_process="Customer browsing or asking for recommendations.",
                reply_message="Dạ chào bạn, shop xin gợi ý các mẫu áo quần đang được ưa chuộng nhất tuần này, phù hợp với yêu cầu của bạn ạ!",
                recommended_product_ids=["prod_01", "prod_02"],
                quick_replies=["Xem thêm mẫu khác", "Lọc sản phẩm dưới 300k", "Chọn size L"],
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

        if prov == "fake":
            return FakeStructuredChatModel()

        if prov in {"openai", "chatgpt"}:
            from langchain_openai import ChatOpenAI
            api_key = kwargs.get("api_key") or os.getenv("OPENAI_API_KEY")
            base_url = kwargs.get("base_url") or os.getenv("OPENAI_API_BASE") or os.getenv("AI_OPENAI_BASE_URL")
            selected_model = model_name or os.getenv("AI_MODEL_NAME", "gpt-4o-mini")
            return ChatOpenAI(
                model=selected_model,
                api_key=api_key,
                base_url=base_url,
                temperature=temperature,
                **kwargs,
            )

        if prov in {"google", "gemini"}:
            from langchain_google_genai import ChatGoogleGenerativeAI
            api_key = kwargs.get("api_key") or os.getenv("GOOGLE_API_KEY") or os.getenv("AI_API_KEY")
            selected_model = model_name or os.getenv("AI_MODEL_NAME", "gemini-2.0-flash")
            return ChatGoogleGenerativeAI(
                model=selected_model,
                google_api_key=api_key,
                temperature=temperature,
                **kwargs,
            )

        if prov in {"ollama", "local"}:
            from langchain_openai import ChatOpenAI
            base_url = kwargs.get("base_url") or os.getenv("OLLAMA_BASE_URL", "http://localhost:11434/v1")
            selected_model = model_name or os.getenv("AI_MODEL_NAME", "llama3.2")
            return ChatOpenAI(
                model=selected_model,
                api_key="ollama",
                base_url=base_url,
                temperature=temperature,
                **kwargs,
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
