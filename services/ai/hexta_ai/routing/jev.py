"""Jev (TypeSafe System One) message router. Output is untrusted; Go keeps all business authority.

One System One call per routed message. Only typed-redacted customer text, a few redacted recent
messages and redacted script trigger/answer excerpts are sent; facts, knowledge, constraints and
identifiers are not. Script IDs are replaced by local option labels.
"""

import asyncio
import logging
import math
import os
from collections.abc import Mapping
from typing import Any, Protocol

from pydantic import BaseModel, Field, SecretStr, ValidationError, model_validator

from hexta_ai.redaction import redact_pii
from hexta_ai.routing.decision import RouteDecision

log = logging.getLogger(__name__)

INTENTS = ("script", "answer", "clarify", "human")
INTENT_CRITERIA = {
    "script": "One of the shop's prepared replies directly and completely answers the customer's message.",
    "answer": ("A question about the shop's products, prices, stock, sizes, colors, shipping or store policies "
               "that the shop's own product and policy information could answer."),
    "clarify": "The message is too vague or incomplete to know what the customer wants.",
    "human": ("Needs a staff member: asks for a person, complains, reports a problem with an order, payment, "
              "delivery or refund, negotiates a price, or raises anything sensitive."),
}
INTENT_INSTRUCTIONS = ("Decide how an online shop's chat assistant should handle the customer's latest message. "
                       "Messages are usually Vietnamese, sometimes without diacritics or with abbreviations.")
HUMAN_INSTRUCTIONS = ("The customer explicitly asks to talk to a human staff member, complains, or reports a "
                      "problem with an existing order, payment, delivery or refund.")
SCRIPT_INSTRUCTIONS = "Which prepared reply best answers the customer's latest message?"
SCRIPT_FITS_INSTRUCTIONS = "One of these prepared replies fully answers the customer's latest message: "


class JevConfig(BaseModel):
    enabled: bool = False
    api_key: SecretStr = Field(default=SecretStr(""), repr=False)
    model: str = Field(default="jev-1.14.0", pattern=r"^[A-Za-z0-9._-]{1,64}$")
    base_url: str = Field(default="https://api.typesafe.ai", pattern=r"^https://[A-Za-z0-9.-]+(:\d+)?$")
    timeout: float = Field(default=5, ge=1, le=20)
    total_timeout: float = Field(default=12, ge=1, le=30)
    max_retries: int = Field(default=2, ge=0, le=5)
    min_probability: float = Field(default=0.6, ge=0, le=1)
    script_min_probability: float = Field(default=0.6, ge=0, le=1)
    noul_yes: float = Field(default=0.7, ge=0, le=1)
    noul_no: float = Field(default=0.3, ge=0, le=1)
    history_messages: int = Field(default=2, ge=0, le=6)
    script_excerpt_chars: int = Field(default=300, ge=0, le=1000)

    @model_validator(mode="after")
    def ordered_thresholds(self):
        if self.noul_no >= self.noul_yes:
            raise ValueError("AI_JEV_NOUL_NO must be below AI_JEV_NOUL_YES")
        if self.model in ("jev-latest", "latest"):
            raise ValueError("Pin an explicit Jev model version")
        return self

    @property
    def active(self) -> bool:
        return self.enabled and bool(self.api_key.get_secret_value().strip())


ENV = {"enabled": "AI_JEV_ENABLED", "api_key": "TYPESAFE_API_KEY", "model": "AI_JEV_MODEL",
       "base_url": "AI_JEV_BASE_URL", "timeout": "AI_JEV_TIMEOUT", "total_timeout": "AI_JEV_TOTAL_TIMEOUT",
       "max_retries": "AI_JEV_MAX_RETRIES", "min_probability": "AI_JEV_MIN_PROBABILITY",
       "script_min_probability": "AI_JEV_SCRIPT_MIN_PROBABILITY", "noul_yes": "AI_JEV_NOUL_YES",
       "noul_no": "AI_JEV_NOUL_NO", "history_messages": "AI_JEV_HISTORY_MESSAGES"}


def jev_config_from_env() -> JevConfig:
    values = {field: os.environ[name].strip() for field, name in ENV.items() if os.environ.get(name, "").strip()}
    if "enabled" in values:
        values["enabled"] = values["enabled"].lower() in ("1", "true", "yes")
    try:
        return JevConfig(**values)
    except ValidationError:
        raise RuntimeError("Invalid Jev router configuration") from None


class JevUnavailable(Exception):
    """Safe failure category; never carries response bodies or credentials."""

    CODES = {"jev_bad_request", "jev_auth", "jev_rate_limited", "jev_unavailable", "jev_timeout",
             "jev_invalid_output", "jev_error"}

    def __init__(self, code: str = "jev_error"):
        self.code = code if code in self.CODES else "jev_error"
        super().__init__(self.code)


class SystemOneClient(Protocol):
    async def system_one(self, state: Any, questions: Mapping[str, Any], *, model: str | None = None) -> Any: ...

    async def aclose(self) -> None: ...


class _SdkClient:
    """Adapter over the official async SDK; created lazily so the SDK loads only when Jev is enabled."""

    def __init__(self, config: JevConfig, transport=None):
        from typesafe_sdk import AsyncTypeSafeClient, RetryPolicy

        # The SDK logs request/response bodies at DEBUG; never let TYPESAFE_LOG_LEVEL enable that here.
        logging.getLogger("typesafe_sdk").setLevel(logging.WARNING)
        # Retry only rate limiting/overload: a timed-out or dropped request may already have been charged.
        retry = RetryPolicy(max_retries=config.max_retries, http_statuses={429, 529}, api_connection_error=False,
                            api_timeout_error=False, timeout=config.total_timeout)
        self.client = AsyncTypeSafeClient(api_key=config.api_key.get_secret_value(), model=config.model,
                                          base_url=config.base_url, timeout=config.timeout, retry=retry,
                                          transport=transport)

    async def system_one(self, state, questions, *, model=None):
        from typesafe_sdk import Choice, Noul

        built = {name: (Choice(**q) if q["type"] == "choice" else Noul(**q)) for name, q in questions.items()}
        return await self.client.system_one(state=state, questions=built, model=model)

    async def aclose(self):
        await self.client.aclose()


def _map_sdk_error(error: BaseException) -> JevUnavailable:
    try:
        from typesafe_sdk import TypeSafeAPIError, TypeSafeAPIResponseValidationError, TypeSafeAPITimeoutError
    except ImportError:  # pragma: no cover - dependency is locked
        return JevUnavailable()
    if isinstance(error, (TimeoutError, TypeSafeAPITimeoutError)):
        return JevUnavailable("jev_timeout")
    if isinstance(error, TypeSafeAPIResponseValidationError):
        return JevUnavailable("jev_invalid_output")
    if isinstance(error, TypeSafeAPIError):
        status = error.status
        if status in (400, 413, 415, 422):
            return JevUnavailable("jev_bad_request")
        if status in (401, 403):
            return JevUnavailable("jev_auth")
        if status == 429:
            return JevUnavailable("jev_rate_limited")
        if status >= 500:
            return JevUnavailable("jev_unavailable")
    return JevUnavailable()


def _probability(value) -> float:
    if isinstance(value, bool) or not isinstance(value, (int, float)) or not math.isfinite(value):
        raise JevUnavailable("jev_invalid_output")
    if not -1e-6 <= value <= 1 + 1e-6:
        raise JevUnavailable("jev_invalid_output")
    return min(1.0, max(0.0, float(value)))


def _choice(response, name, labels) -> tuple[str, float]:
    answer = getattr(response, "choices", {}).get(name)
    if answer is None or answer.choice not in labels:
        raise JevUnavailable("jev_invalid_output")
    return answer.choice, _probability(answer.probabilities.get(answer.choice))


def _noul(response, name) -> float:
    answer = getattr(response, "nouls", {}).get(name)
    if answer is None:
        raise JevUnavailable("jev_invalid_output")
    return _probability(answer.noul)


class JevRouter:
    name = "jev"
    simulated = False
    # Fallback reason recorded when Jev answers below the configured thresholds.
    uncertain_reason = "jev_low_probability"

    def __init__(self, config: JevConfig, client: SystemOneClient | None = None):
        if not config.active:
            raise RuntimeError("Jev router requires AI_JEV_ENABLED=true and TYPESAFE_API_KEY")
        self.config = config
        self.model = config.model
        self.client = client if client is not None else _SdkClient(config)

    def request(self, message: str, scripts: list[dict], history: list[dict]):
        """Build the redacted System One state and questions plus the local script label mapping."""
        recent = history[-self.config.history_messages:] if self.config.history_messages else []
        state = {"customer_message": redact_pii(message),
                 "recent_messages": [{"from": "customer" if h["role"] == "user" else "shop",
                                      "text": redact_pii(h["text"])} for h in recent]}
        labels = {f"s{i + 1}": s["id"] for i, s in enumerate(scripts)}
        options = {}
        for label, script in zip(labels, scripts):
            excerpt = script["text"][:self.config.script_excerpt_chars]
            options[label] = redact_pii(f"Customer phrases: {script['trigger']}. Prepared reply: {excerpt}")
        intents = {k: v for k, v in INTENT_CRITERIA.items() if k != "script" or options}
        questions = {
            "intent": {"type": "choice", "instructions": INTENT_INSTRUCTIONS, "criteria": intents},
            "wants_human": {"type": "noul", "instructions": HUMAN_INSTRUCTIONS},
        }
        if options:
            questions["script"] = {"type": "choice", "instructions": SCRIPT_INSTRUCTIONS, "criteria": options}
            listing = "; ".join(f"{label}: {text}" for label, text in options.items())
            questions["script_fits"] = {"type": "noul", "instructions": SCRIPT_FITS_INSTRUCTIONS + listing}
        return state, questions, labels

    async def route(self, message: str, scripts: list[dict], history: list[dict]) -> RouteDecision | None:
        """A confident decision, None when uncertain (caller falls back), JevUnavailable on failure."""
        state, questions, labels = self.request(message, scripts, history)
        try:
            async with asyncio.timeout(self.config.total_timeout + 1):
                response = await self.client.system_one(state, questions, model=self.config.model)
        except JevUnavailable:
            raise
        except asyncio.CancelledError:
            raise
        except Exception as error:
            raise _map_sdk_error(error) from None
        return self.decide(response, labels)

    def decide(self, response, labels: dict[str, str]) -> RouteDecision | None:
        c = self.config
        allowed = tuple(k for k in INTENTS if k != "script" or labels)
        intent, probability = _choice(response, "intent", allowed)
        wants_human = _noul(response, "wants_human")
        decision = {"router": "jev", "router_model": self.model}
        if wants_human >= c.noul_yes:
            return RouteDecision(route="human", probability=wants_human, reason="jev_customer_requested_human",
                                 handoff_reason="customer_requested_staff", **decision)
        if probability < c.min_probability:
            return None
        if intent == "script":
            label, script_probability = _choice(response, "script", tuple(labels))
            if script_probability < c.script_min_probability or _noul(response, "script_fits") < c.noul_yes:
                return None
            return RouteDecision(route="script", script_id=labels[label], probability=min(probability, script_probability),
                                 reason="jev_script_match", **decision)
        if intent == "answer":
            if wants_human > c.noul_no:
                return None
            return RouteDecision(route="answer", probability=probability, reason="jev_knowledge_question", **decision)
        if intent == "clarify":
            return RouteDecision(route="clarify", probability=probability, reason="jev_needs_clarification", **decision)
        return RouteDecision(route="human", probability=probability, reason="jev_needs_staff",
                             handoff_reason="business_review", **decision)

    async def close(self):
        closer = getattr(self.client, "aclose", None)
        if closer:
            await closer()


def jev_router_from_env(*, allowed: bool = True) -> JevRouter | None:
    """Jev only when enabled with a key and while AI is enabled (allowed). AI_JEV_ENABLED=true without
    TYPESAFE_API_KEY is a misconfiguration: it is logged as an error and Jev stays off (fail-safe, no crash)."""
    config = jev_config_from_env()
    if not config.enabled:
        return None
    if not config.active:
        log.error("Jev router off: AI_JEV_ENABLED=true but TYPESAFE_API_KEY is empty (misconfiguration)")
        return None
    if not allowed:
        log.warning("Jev router off: AI is disabled, so no paid routing call is made")
        return None
    return JevRouter(config)
