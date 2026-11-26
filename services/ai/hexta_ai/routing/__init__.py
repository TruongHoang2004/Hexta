"""Message routing: Jev primary router, offline keyword router and deterministic fallback."""

import logging
import os

from hexta_ai.routing.decision import HANDOFF_REASONS, RouteDecision, deterministic_decision, exact_script_match
from hexta_ai.routing.jev import JevConfig, JevRouter, JevUnavailable, jev_config_from_env, jev_router_from_env
from hexta_ai.routing.keyword import KeywordRouter

FAKE_ROUTERS = {"keyword": KeywordRouter}

log = logging.getLogger(__name__)


def router_from_env(*, ai_enabled: bool = True, paid_allowed: bool | None = None):
    """Primary router: Jev when enabled with a key and a real answer model is configured (paid_allowed, default
    ai_enabled); else AI_FAKE_ROUTER (offline, simulated, the explicit demo flag); else None, which keeps the
    existing provider classifier. With AI disabled there is no router at all (exact scripts or staff). Invalid Jev
    settings are logged and keep Jev off instead of failing startup."""
    try:
        jev = jev_router_from_env(allowed=ai_enabled if paid_allowed is None else paid_allowed)
    except RuntimeError:
        log.error("Jev router off: invalid AI_JEV_* settings (misconfiguration)")
        jev = None
    if jev is not None:
        return jev
    if not ai_enabled:
        return None
    name = os.getenv("AI_FAKE_ROUTER", "").strip().lower()
    if not name:
        return None
    if name not in FAKE_ROUTERS:
        raise RuntimeError("Unsupported AI_FAKE_ROUTER")
    return FAKE_ROUTERS[name]()


__all__ = ["HANDOFF_REASONS", "JevConfig", "JevRouter", "JevUnavailable", "KeywordRouter", "RouteDecision",
           "deterministic_decision", "exact_script_match", "jev_config_from_env", "jev_router_from_env",
           "router_from_env"]
