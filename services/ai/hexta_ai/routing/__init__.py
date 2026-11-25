"""Message routing: Jev primary router, offline keyword router and deterministic fallback."""

import os

from hexta_ai.routing.decision import HANDOFF_REASONS, RouteDecision, deterministic_decision, exact_script_match
from hexta_ai.routing.jev import JevConfig, JevRouter, JevUnavailable, jev_config_from_env, jev_router_from_env
from hexta_ai.routing.keyword import KeywordRouter

FAKE_ROUTERS = {"keyword": KeywordRouter}


def router_from_env():
    """Primary router: Jev when enabled with a key; else AI_FAKE_ROUTER (offline, simulated); else None, which keeps
    the existing provider classifier."""
    jev = jev_router_from_env()
    if jev is not None:
        return jev
    name = os.getenv("AI_FAKE_ROUTER", "").strip().lower()
    if not name:
        return None
    if name not in FAKE_ROUTERS:
        raise RuntimeError("Unsupported AI_FAKE_ROUTER")
    return FAKE_ROUTERS[name]()


__all__ = ["HANDOFF_REASONS", "JevConfig", "JevRouter", "JevUnavailable", "KeywordRouter", "RouteDecision",
           "deterministic_decision", "exact_script_match", "jev_config_from_env", "jev_router_from_env",
           "router_from_env"]
