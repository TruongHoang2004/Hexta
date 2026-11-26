"""Explicit provider selection.

`get_provider` is strict and raises without credentials. `provider_from_env` is the runtime entry and fails safe:
a missing, partial or invalid configuration selects the disabled provider (exact active scripts, everything else to
staff) and logs why, so a deploy without AI settings starts instead of crashing and rolling back.
"""

import logging
import os

from pydantic import BaseModel, Field, SecretStr, ValidationError

from hexta_ai.providers.base import BaseAIProvider
from hexta_ai.providers.fake import DisabledProvider, FakeProvider

log = logging.getLogger(__name__)


class ProviderConfig(BaseModel):
    api_key: SecretStr = Field(default=SecretStr(""), repr=False)
    model: str = Field(default="gemini-2.5-flash", pattern=r"^[A-Za-z0-9._:/-]{1,150}$")
    timeout: float = Field(default=20, ge=1, le=30)
    max_output_tokens: int = Field(default=2048, ge=64, le=8192)
    max_input_tokens: int = Field(default=16000, ge=256, le=64000)
    temperature: float = Field(default=0.7, ge=0, le=1)
    confidence_threshold: float = Field(default=0.8, ge=0.5, le=1)


# Recommended models (documentation and examples). Real providers need AI_MODEL set explicitly.
DEFAULT_MODELS = {"fake": "deterministic", "gemini": "gemini-2.5-flash",
                  "openai": "gpt-4o-mini", "anthropic": "claude-sonnet-5-5"}
KEYS = {"gemini": "GEMINI_API_KEY", "openai": "OPENAI_API_KEY", "anthropic": "ANTHROPIC_API_KEY"}
# A model left over from another provider (for example AI_MODEL=gpt-4o-mini with anthropic) is a partial edit.
MODEL_PREFIXES = {"gemini": ("gemini-",), "anthropic": ("claude-",),
                  "openai": ("gpt-", "chatgpt-", "o1", "o3", "o4")}
PROVIDERS = ("disabled", *DEFAULT_MODELS)


class AIConfigError(Exception):
    """Why AI stays disabled. Messages name settings only, never their values."""


def get_provider(provider_name: str, config: ProviderConfig) -> BaseAIProvider:
    if provider_name not in DEFAULT_MODELS:
        raise RuntimeError("Unsupported AI_PROVIDER")
    if provider_name == "fake":
        return FakeProvider()
    if not config.api_key.get_secret_value().strip():
        raise RuntimeError("Selected AI provider requires a server-side API key")
    from hexta_ai.providers.langchain import LangChainProvider
    return LangChainProvider(provider_name, config)


def configured_provider() -> BaseAIProvider:
    """Strict environment selection; raises AIConfigError for a missing, partial or invalid configuration."""
    name = os.getenv("AI_PROVIDER", "fake").strip().lower()
    if name == "disabled":
        return DisabledProvider()
    if name not in DEFAULT_MODELS:
        raise AIConfigError("AI_PROVIDER must be one of " + ", ".join(PROVIDERS))
    if name == "fake":
        return FakeProvider()
    model = os.getenv("AI_MODEL", "").strip()
    if not model:
        raise AIConfigError(f"AI_PROVIDER={name} needs AI_MODEL")
    if not model.startswith(MODEL_PREFIXES[name]):
        raise AIConfigError(f"AI_MODEL is not a {name} model (edit AI_PROVIDER and AI_MODEL together)")
    key = os.getenv(KEYS[name]) or os.getenv("AI_API_KEY", "")
    if not key.strip():
        raise AIConfigError(f"AI_PROVIDER={name} needs {KEYS[name]}")
    try:
        config = ProviderConfig(
            api_key=key, model=model,
            timeout=os.getenv("AI_TIMEOUT", "20"),
            max_output_tokens=os.getenv("AI_MAX_TOKENS", "2048"),
            max_input_tokens=os.getenv("AI_MAX_INPUT_TOKENS", "16000"),
            temperature=os.getenv("AI_TEMPERATURE", "0.7"),
            confidence_threshold=os.getenv("AI_CONFIDENCE_THRESHOLD", "0.8"),
        )
    except ValidationError:
        raise AIConfigError("invalid AI_MODEL, AI_TIMEOUT, AI_MAX_TOKENS, AI_MAX_INPUT_TOKENS, AI_TEMPERATURE or "
                            "AI_CONFIDENCE_THRESHOLD") from None
    return get_provider(name, config)


def is_real(provider: BaseAIProvider) -> bool:
    """A configured paid model. Only then may Jev routing or paid embeddings run: the disabled provider and the
    offline fake provider (demo/tests) never enable a paid call, whatever keys are present."""
    return provider.name != "disabled" and not provider.simulated


def provider_from_env() -> BaseAIProvider:
    """Fail-safe runtime selection: any configuration error disables AI with one startup error line."""
    try:
        return configured_provider()
    except AIConfigError as error:
        log.error("AI disabled: %s. Bot replies use exact active scripts only; every other message is handed "
                  "to staff. Set AI_PROVIDER, AI_MODEL and the provider key together to enable AI.", error)
        return DisabledProvider(reason="misconfigured")
