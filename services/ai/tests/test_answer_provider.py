"""Answer port and embeddings over local transports: prompt layout, refusals, defaults. No paid calls."""

import asyncio
import json
from unittest.mock import patch

import httpx
import pytest

from hexta_ai.providers import ProviderError, provider_from_env
from tests.test_providers import make_provider, sdk_response

ANSWER = {"kind": "answer", "text": "Dạ áo giá 299.000đ ạ.", "cited_source_ids": ["product:p1"],
          "confidence": 0.9, "needs_human": False}
CONTEXT = {"knowledge": [{"id": "product:p1", "title": "Áo", "text": "Áo 299.000đ", "attributes": {}}],
           "conversation": [], "customer_message": "giá áo?", "constraints": []}


def test_anthropic_answer_uses_cached_stable_system_prompt_then_context():
    from hexta_ai.providers.base import ANSWER_INSTRUCTIONS
    requests = []

    def handler(request):
        requests.append(json.loads(request.content))
        return httpx.Response(200, json=sdk_response("anthropic", ANSWER))

    async def scenario():
        provider = make_provider("anthropic", handler)
        try:
            answer, usage = await provider.answer(CONTEXT)
            assert answer.cited_source_ids == ["product:p1"] and usage.input_tokens == 11
        finally:
            await provider.close()
    asyncio.run(scenario())
    body = requests[0]
    assert body["system"][0]["text"] == ANSWER_INSTRUCTIONS
    assert body["system"][0]["cache_control"] == {"type": "ephemeral"}
    content = body["messages"][0]["content"]
    text = content if isinstance(content, str) else content[0]["text"]
    assert text.index('"knowledge"') < text.index('"conversation"') < text.index('"customer_message"')
    assert body["output_config"]["format"]["type"] == "json_schema"


@pytest.mark.parametrize("name", ["gemini", "openai"])
def test_other_providers_answer_with_plain_system_prompt(name):
    requests = []

    def handler(request):
        requests.append(json.loads(request.content))
        return httpx.Response(200, json=sdk_response(name, ANSWER))

    async def scenario():
        provider = make_provider(name, handler)
        try:
            answer, _ = await provider.answer(CONTEXT)
            assert answer.kind == "answer"
        finally:
            await provider.close()
    asyncio.run(scenario())
    assert "cache_control" not in json.dumps(requests[0])


@pytest.mark.parametrize("name,value", [("anthropic", "refusal"), ("gemini", "SAFETY")])
def test_refusals_are_a_distinct_safe_category(name, value):
    def handler(request):
        body = sdk_response(name, ANSWER)
        if name == "gemini":
            body["candidates"][0]["finishReason"] = value
        else:
            body["stop_reason"] = value
        return httpx.Response(200, json=body)

    async def scenario():
        provider = make_provider(name, handler)
        try:
            with pytest.raises(ProviderError, match="LLM_REFUSAL"):
                await provider.answer(CONTEXT)
        finally:
            await provider.close()
    asyncio.run(scenario())


def test_claude_is_the_default_anthropic_model(monkeypatch):
    from hexta_ai.providers.factory import DEFAULT_MODELS
    assert DEFAULT_MODELS["anthropic"] == "claude-sonnet-5-5"
    monkeypatch.setenv("AI_PROVIDER", "anthropic")
    monkeypatch.setenv("ANTHROPIC_API_KEY", "test-only")
    monkeypatch.delenv("AI_MODEL", raising=False)
    with patch("hexta_ai.providers.langchain.LangChainProvider") as constructor:
        provider_from_env()
        assert constructor.call_args.args[1].model == "claude-sonnet-5-5"


def test_gemini_embeddings_offline_tasks_dimensions_and_errors():
    from langchain_google_genai import GoogleGenerativeAIEmbeddings

    from hexta_ai.knowledge.embeddings import DIMENSIONS, EmbeddingConfig, EmbeddingError, GeminiEmbeddings
    requests = []

    def handler(request):
        requests.append(json.loads(request.content))
        if len(requests) == 3:
            return httpx.Response(429, json={"error": {"code": 429, "message": "secret quota", "status": "RESOURCE_EXHAUSTED"}})
        count = len(requests[-1].get("requests", [])) or 1
        return httpx.Response(200, json={"embeddings": [{"values": [3.0] + [0.0] * (DIMENSIONS - 1)}] * count})

    async def scenario():
        client = GoogleGenerativeAIEmbeddings(model="gemini-embedding-001", google_api_key="test-only", vertexai=False,
                                              client_args={"transport": httpx.MockTransport(handler)})
        embeddings = GeminiEmbeddings(EmbeddingConfig(provider="gemini", api_key="test-only"), client=client)
        vectors = await embeddings.embed_documents(["a", "b"])
        assert len(vectors) == 2 and vectors[0][0] == pytest.approx(1.0)
        assert (await embeddings.embed_query("q"))[0] == pytest.approx(1.0)
        with pytest.raises(EmbeddingError, match="EMBEDDING_QUOTA"):
            await embeddings.embed_query("q")
    asyncio.run(scenario())
    assert "RETRIEVAL_DOCUMENT" in json.dumps(requests[0]) and "RETRIEVAL_QUERY" in json.dumps(requests[1])
    assert str(DIMENSIONS) in json.dumps(requests[1])


def test_embedding_factory_defaults_fake_and_requires_key(monkeypatch):
    from hexta_ai.knowledge.embeddings import HashingEmbeddings, embeddings_from_env
    monkeypatch.delenv("AI_EMBEDDING_PROVIDER", raising=False)
    assert isinstance(embeddings_from_env(), HashingEmbeddings)
    monkeypatch.setenv("AI_EMBEDDING_PROVIDER", "gemini")
    monkeypatch.setenv("GEMINI_API_KEY", "")
    monkeypatch.delenv("AI_EMBEDDING_API_KEY", raising=False)
    with pytest.raises(RuntimeError, match="require"):
        embeddings_from_env()
