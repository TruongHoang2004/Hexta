"""Router decision value and the deterministic fallback matcher. No network access."""

from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

# Go handoff vocabulary (messaging_command_service): Python only proposes one of these categories.
HANDOFF_REASONS = ("customer_requested_staff", "insufficient_information", "ai_error", "business_review")


class RouteDecision(BaseModel):
    model_config = ConfigDict(extra="forbid")
    route: Literal["script", "answer", "generate", "clarify", "human"]
    script_id: str | None = None
    probability: float = Field(default=0, ge=0, le=1)
    router: Literal["jev", "deterministic", "provider"]
    router_model: str = ""
    reason: str = ""
    handoff_reason: Literal["", *HANDOFF_REASONS] = ""
    fallback_reason: str = ""


def exact_script_match(message: str, scripts: list[dict]) -> str | None:
    """The unique active script whose trigger equals the message (case-insensitive), else None."""
    wanted = message.strip().casefold()
    matches = [s["id"] for s in scripts if s.get("active", True) and s["trigger"].strip().casefold() == wanted]
    return matches[0] if len(matches) == 1 else None


def deterministic_decision(message: str, scripts: list[dict], fallback_reason: str) -> RouteDecision:
    """Fallback when the primary router is unavailable or uncertain: unique exact match, else human."""
    script_id = exact_script_match(message, scripts)
    if script_id:
        return RouteDecision(route="script", script_id=script_id, probability=1, router="deterministic",
                             reason="deterministic_exact_match", fallback_reason=fallback_reason)
    return RouteDecision(route="human", router="deterministic", reason="router_uncertain",
                         handoff_reason="insufficient_information", fallback_reason=fallback_reason)
