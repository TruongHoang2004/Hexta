"""LangChain integrations exercised with local HTTP transports; no paid calls."""

import asyncio
import json
from types import SimpleNamespace
from unittest.mock import AsyncMock, Mock, patch

import httpx
import httpx2
import pytest
from langchain.chat_models import init_chat_model

from hexta_ai.providers import ProviderConfig, ProviderError, get_provider, provider_from_env
from hexta_ai.providers.langchain import LangChainProvider

DECISION = {"route": "script", "script_id": "s1", "confidence": 0.95}
CONFIG = ProviderConfig(api_key="test-only-not-a-real-key", model="test-model", timeout=1)
CONTEXT = {"message": "price", "scripts": [{"id": "s1", "trigger": "price", "text": "Select a product."}]}


def sdk_response(name, value):
    text = json.dumps(value)
    if name == "openai":
        return {"id": "resp-test", "object": "response", "created_at": 0, "status": "completed",
                "model": "test-model", "output": [{"id": "msg-test", "type": "message", "role": "assistant",
                "status": "completed", "content": [{"type": "output_text", "text": text, "annotations": []}]}],
                "usage": {"input_tokens": 11, "output_tokens": 7, "total_tokens": 18,
                          "input_tokens_details": {"cached_tokens": 0}, "output_tokens_details": {"reasoning_tokens": 0}}}
    if name == "anthropic":
        return {"id": "msg-test", "type": "message", "role": "assistant", "model": "test-model",
                "content": [{"type": "text", "text": text}], "stop_reason": "end_turn", "stop_sequence": None,
                "usage": {"input_tokens": 11, "output_tokens": 7}}
    return {"candidates": [{"content": {"role": "model", "parts": [{"text": text}]}, "finishReason": "STOP"}],
            "usageMetadata": {"promptTokenCount": 11, "candidatesTokenCount": 7, "totalTokenCount": 18}}


def make_provider(name, handler):
    options = {"api_key": "test-only", "max_tokens": 2048, "timeout": 1,
               "max_retries": 1 if name == "gemini" else 0}
    if name == "anthropic":
        def anthropic_handler(request):
            response = handler(request)
            return httpx2.Response(response.status_code, json=json.loads(response.content))
        client = httpx2.AsyncClient(transport=httpx2.MockTransport(anthropic_handler))
        model = init_chat_model("test-model", model_provider="anthropic", **options)
        with patch("langchain_anthropic.chat_models._get_default_async_httpx_client", return_value=client):
            # Materialize only the cached test client while transport is replaced.
            model._async_client
    elif name == "openai":
        options.update(use_responses_api=True, store=False,
                       http_async_client=httpx.AsyncClient(transport=httpx.MockTransport(handler)))
        model = init_chat_model("test-model", model_provider="openai", **options)
    else:
        options.update(vertexai=False, client_args={"transport": httpx.MockTransport(handler)})
        model = init_chat_model("test-model", model_provider="google_genai", **options)
    return LangChainProvider(name, CONFIG, chat_model=model)


@pytest.mark.parametrize("name", ["gemini", "openai", "anthropic"])
def test_langchain_structured_requests_usage_generation_and_buffered_stream(name):
    requests = []
    outputs = [DECISION, {"text": "Fact text.", "fact_ids": ["f1"]}, {"text": "Fact text.", "fact_ids": ["f1"]}]

    def handler(request):
        requests.append(json.loads(request.content))
        return httpx.Response(200, json=sdk_response(name, outputs[len(requests) - 1]))

    async def scenario():
        provider = make_provider(name, handler)
        try:
            decision, usage = await provider.classify(CONTEXT)
            assert decision.script_id == "s1" and decision.confidence == 0.95
            assert usage.input_tokens == 11 and usage.output_tokens == 7 and not usage.estimated
            candidate, _ = await provider.generate_candidate(CONTEXT)
            assert candidate.fact_ids == ["f1"]
            chunks = [c async for c in provider.stream_candidate(CONTEXT)]
            assert chunks == [candidate]
            assert await provider.estimate_tokens("hello") > 0
        finally:
            await provider.close()
    asyncio.run(scenario())

    assert len(requests) == 3
    body = requests[0]
    if name == "gemini":
        assert body["generationConfig"]["maxOutputTokens"] == 2048
        assert body["generationConfig"]["responseMimeType"] == "application/json"
    elif name == "openai":
        assert body["max_output_tokens"] == 2048 and body["store"] is False
        assert body["text"]["format"]["type"] == "json_schema"
    else:
        assert body["max_tokens"] == 2048
        assert body["output_config"]["format"]["type"] == "json_schema"


@pytest.mark.parametrize("name", ["gemini", "openai", "anthropic"])
def test_truncation_never_publishes_a_partial_candidate(name):
    def handler(request):
        body = sdk_response(name, DECISION)
        if name == "gemini":
            body["candidates"][0]["finishReason"] = "MAX_TOKENS"
        elif name == "openai":
            body["status"] = "incomplete"
        else:
            body["stop_reason"] = "max_tokens"
        return httpx.Response(200, json=body)
    async def scenario():
        provider = make_provider(name, handler)
        try:
            with pytest.raises(ProviderError):
                await provider.classify(CONTEXT)
        finally:
            await provider.close()
    asyncio.run(scenario())


@pytest.mark.parametrize("name", ["gemini", "openai", "anthropic"])
@pytest.mark.parametrize("status,code", [(429, "QUOTA_EXCEEDED"), (500, "LLM_PROVIDER_ERROR")])
def test_errors_are_sanitized_and_not_retried(name, status, code, caplog):
    calls = []
    def handler(request):
        calls.append(request)
        return httpx.Response(status, json={"error": {"message": "secret-api-key and private prompt", "type": "api_error", "code": status}})
    async def scenario():
        provider = make_provider(name, handler)
        try:
            with pytest.raises(ProviderError) as error:
                await provider.classify(CONTEXT)
            assert str(error.value) == code
        finally:
            await provider.close()
    asyncio.run(scenario())
    assert len(calls) == 1
    assert "secret-api-key" not in caplog.text and "private prompt" not in caplog.text


@pytest.mark.parametrize("name", ["gemini", "openai", "anthropic"])
def test_malformed_structured_outputs_fail_closed(name):
    def handler(request):
        body = sdk_response(name, {**DECISION, "route": "send_customer_message"})
        return httpx.Response(200, json=body)
    async def scenario():
        provider = make_provider(name, handler)
        try:
            with pytest.raises(ProviderError):
                await provider.classify(CONTEXT)
        finally:
            await provider.close()
    asyncio.run(scenario())


def test_timeout_and_budget_do_not_retry():
    async def scenario():
        chain = SimpleNamespace(ainvoke=AsyncMock(side_effect=TimeoutError("secret")))
        model = SimpleNamespace(with_structured_output=Mock(return_value=chain))
        provider = LangChainProvider("openai", CONFIG, model)
        with pytest.raises(ProviderError, match="LLM_TIMEOUT"):
            await provider.classify(CONTEXT)
        assert chain.ainvoke.await_count == 1
        provider.max_input_tokens = 1
        with pytest.raises(ProviderError, match="INPUT_BUDGET_EXCEEDED"):
            await provider.classify(CONTEXT)
        assert chain.ainvoke.await_count == 1
    asyncio.run(scenario())


def test_factory_key_precedence_missing_key_and_safe_configuration(monkeypatch):
    assert get_provider("gemini", ProviderConfig()).simulated
    monkeypatch.setenv("AI_PROVIDER", "gemini")
    monkeypatch.setenv("AI_API_KEY", "generic-test")
    monkeypatch.setenv("GEMINI_API_KEY", "specific-test")
    with patch("hexta_ai.providers.langchain.LangChainProvider") as constructor:
        provider_from_env()
        assert constructor.call_args.args[1].api_key.get_secret_value() == "specific-test"
        assert "specific-test" not in repr(constructor.call_args.args[1])
    monkeypatch.setenv("AI_TIMEOUT", "secret-invalid-value")
    with pytest.raises(RuntimeError, match="^Invalid AI provider configuration$"):
        provider_from_env()
    monkeypatch.setenv("AI_PROVIDER", "unknown")
    with pytest.raises(RuntimeError, match="^Unsupported AI_PROVIDER$"):
        provider_from_env()


@pytest.mark.parametrize("name", ["gemini", "openai", "anthropic"])
def test_factory_builds_langchain_models_with_bounded_calls_and_no_retry(name):
    from langchain_core.language_models.chat_models import BaseChatModel

    async def scenario():
        provider = get_provider(name, CONFIG)
        try:
            model = provider.chat_model
            assert isinstance(model, BaseChatModel)
            # Google's integration counts total attempts; the other two count retries.
            assert model.max_retries == (1 if name == "gemini" else 0)
            if name == "gemini":
                assert model.max_output_tokens == CONFIG.max_output_tokens
                assert model.timeout == CONFIG.timeout and model.vertexai is False
            elif name == "openai":
                assert model.max_tokens == CONFIG.max_output_tokens
                assert model.request_timeout == CONFIG.timeout and model.store is False
            else:
                assert model.max_tokens == CONFIG.max_output_tokens
                assert model.default_request_timeout == CONFIG.timeout
        finally:
            await provider.close()
    asyncio.run(scenario())
