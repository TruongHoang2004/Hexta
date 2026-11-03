"""Pydantic schemas for the AI Shopping Consultant and structured outputs."""

from typing import Any, Literal
from pydantic import BaseModel, ConfigDict, Field


IntentType = Literal[
    "discovery",       # Customer searching for products / browsing
    "refinement",      # Customer refining by price, size, color, attributes
    "product_detail",  # Customer asking detailed question about specific product
    "order_draft",     # Customer requesting to buy / draft an order
    "faq_policy",      # Asking about shipping, return policies, store location
    "human_handoff",   # Unclear, complex complaint, or explicitly requests human staff
]


class ExtractedFilters(BaseModel):
    model_config = ConfigDict(extra="ignore")
    query_keywords: list[str] = Field(default_factory=list, description="Keywords for search")
    category: str | None = Field(None, description="Product category if mentioned")
    min_price: float | None = Field(None, description="Minimum price filter in VND")
    max_price: float | None = Field(None, description="Maximum price filter in VND")
    attributes: dict[str, str] = Field(default_factory=dict, description="Key attributes like size, color, gender")
    in_stock_only: bool = Field(True, description="Filter only in-stock items")


class CatalogItem(BaseModel):
    model_config = ConfigDict(extra="ignore")
    product_id: str
    sku: str
    name: str
    category: str
    price: float
    original_price: float | None = None
    stock: int = 0
    attributes: dict[str, str] = Field(default_factory=dict)
    description: str = ""
    image_url: str = ""


class OrderDraftItem(BaseModel):
    model_config = ConfigDict(extra="ignore")
    product_id: str
    product_name: str
    variant: str = ""
    quantity: int = Field(default=1, ge=1)
    unit_price: float = Field(ge=0)


class OrderDraft(BaseModel):
    model_config = ConfigDict(extra="ignore")
    draft_id: str
    customer_name: str | None = None
    customer_phone: str | None = None
    shipping_address: str | None = None
    items: list[OrderDraftItem] = Field(default_factory=list)
    total_amount: float = 0.0
    payment_method: Literal["cod", "bank_transfer", "cash"] = "cod"
    notes: str | None = None


class StructuredConsultantDecision(BaseModel):
    """Schema enforced on LLM output via LangChain structured output."""
    model_config = ConfigDict(extra="ignore")
    intent: IntentType = Field(description="Detected user intent")
    thought_process: str = Field(description="Internal reasoning for recommendations or actions")
    reply_message: str = Field(description="Conversational Vietnamese reply to customer")
    recommended_product_ids: list[str] = Field(default_factory=list, description="IDs of recommended products")
    draft_order: OrderDraft | None = Field(None, description="Structured draft order if customer is ready to buy")
    quick_replies: list[str] = Field(default_factory=list, description="Suggested next prompt chips for customer")
    needs_human: bool = Field(False, description="True if handoff to human staff is necessary")


class ConsultantRequest(BaseModel):
    model_config = ConfigDict(extra="ignore", str_strip_whitespace=True)
    tenant_id: str = Field(min_length=1, max_length=100)
    session_id: str = Field(default="default_session", min_length=1, max_length=100)
    message: str = Field(min_length=1, max_length=8000)
    history: list[dict[str, str]] = Field(default_factory=list, description="List of {role, content}")
    customer_context: dict[str, Any] = Field(default_factory=dict, description="Known customer details (name, phone)")


class ConsultantResponse(BaseModel):
    model_config = ConfigDict(extra="ignore")
    tenant_id: str
    session_id: str
    message: str
    intent: IntentType
    recommended_products: list[CatalogItem] = Field(default_factory=list)
    order_draft: OrderDraft | None = None
    quick_replies: list[str] = Field(default_factory=list)
    needs_human: bool = False
    reason: str = ""
    steps: list[str] = Field(default_factory=list)
