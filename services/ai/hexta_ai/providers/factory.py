"""Explicit provider selection; real providers fail closed without credentials."""

import os

from pydantic import BaseModel, Field, SecretStr, ValidationError

from hexta_ai.providers.base import BaseAIProvider
from hexta_ai.providers.fake import FakeProvider


class ProviderConfig(BaseModel):
    api_key: SecretStr = Field(default=SecretStr(""), repr=False)
    model: str = Field(default="gemini-2.5-flash", pattern=r"^[A-Za-z0-9._:/-]{1,150}$")
    timeout: float = Field(default=20, ge=1, le=30)
    max_output_tokens: int = Field(default=2048, ge=64, le=8192)
    max_input_tokens: int = Field(default=16000, ge=256, le=64000)
    temperature: float = Field(default=0.7, ge=0, le=1)
    confidence_threshold: float = Field(default=0.8, ge=0.5, le=1)


DEFAULT_MODELS = {"fake": "deterministic", "gemini": "gemini-2.5-flash",
                  "openai": "gpt-4o-mini", "anthropic": "claude-haiku-4-5"}
KEYS = {"gemini": "GEMINI_API_KEY", "openai": "OPENAI_API_KEY", "anthropic": "ANTHROPIC_API_KEY"}


def get_provider(provider_name: str, config: ProviderConfig) -> BaseAIProvider:
    if provider_name not in DEFAULT_MODELS:
        raise RuntimeError("Unsupported AI_PROVIDER")
    if provider_name == "fake":
        return FakeProvider()
    if not config.api_key.get_secret_value().strip():
        raise RuntimeError("Selected AI provider requires a server-side API key")
    from hexta_ai.providers.langchain import LangChainProvider
    return LangChainProvider(provider_name, config)


def provider_from_env() -> BaseAIProvider:
    name = os.getenv("AI_PROVIDER", "fake").strip().lower()
    if name not in DEFAULT_MODELS:
        raise RuntimeError("Unsupported AI_PROVIDER")
    try:
        config = ProviderConfig(
            api_key=os.getenv(KEYS.get(name, "AI_API_KEY")) or os.getenv("AI_API_KEY", ""),
            model=os.getenv("AI_MODEL") or DEFAULT_MODELS[name],
            timeout=os.getenv("AI_TIMEOUT", "20"),
            max_output_tokens=os.getenv("AI_MAX_TOKENS", "2048"),
            max_input_tokens=os.getenv("AI_MAX_INPUT_TOKENS", "16000"),
            temperature=os.getenv("AI_TEMPERATURE", "0.7"),
            confidence_threshold=os.getenv("AI_CONFIDENCE_THRESHOLD", "0.8"),
        )
    except ValidationError:
        raise RuntimeError("Invalid AI provider configuration") from None
    return get_provider(name, config)
