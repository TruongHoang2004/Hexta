import os
import pytest
from hexta_ai.models.adapter import ModelAdapter, FakeStructuredChatModel
from hexta_ai.schemas.consultant import StructuredConsultantDecision


def test_fake_model_creation():
    model = ModelAdapter.get_model(provider="fake")
    assert isinstance(model, FakeStructuredChatModel)


def test_model_adapter_unsupported_provider():
    with pytest.raises(ValueError, match="Unsupported AI_PROVIDER"):
        ModelAdapter.get_model(provider="unsupported_xyz")


@pytest.mark.asyncio
async def test_fake_model_structured_output():
    runnable = ModelAdapter.create_structured_consultant_runnable(provider="fake")
    decision: StructuredConsultantDecision = await runnable.ainvoke("Tôi muốn mua cái áo polo này, chốt đơn giúp tôi")
    
    assert decision.intent == "order_draft"
    assert decision.draft_order is not None
    assert decision.draft_order.total_amount > 0
    assert len(decision.draft_order.items) == 1
    assert "chốt đơn" in decision.thought_process.lower() or "draft" in decision.thought_process.lower()


@pytest.mark.asyncio
async def test_fake_model_human_handoff():
    runnable = ModelAdapter.create_structured_consultant_runnable(provider="fake")
    decision: StructuredConsultantDecision = await runnable.ainvoke("Tôi muốn gặp nhân viên tư vấn người thật")
    
    assert decision.intent == "human_handoff"
    assert decision.needs_human is True


def test_openai_model_limits_and_retry_defaults(monkeypatch):
    """Verify Issue 5: OpenAI adapter defaults to 0 retries and bounded timeout/max_tokens."""
    monkeypatch.setenv("OPENAI_API_KEY", "sk-test-mock-key")
    model = ModelAdapter.get_model(provider="openai")
    
    assert model.max_retries == 0
    assert model.request_timeout == 30.0
    assert model.max_tokens == 2048


def test_openai_model_limits_custom_kwargs(monkeypatch):
    """Verify Issue 5: OpenAI adapter accepts customized limits."""
    monkeypatch.setenv("OPENAI_API_KEY", "sk-test-mock-key")
    model = ModelAdapter.get_model(provider="openai", max_retries=1, timeout=15.0, max_tokens=1024)
    
    assert model.max_retries == 1
    assert model.request_timeout == 15.0
    assert model.max_tokens == 1024
