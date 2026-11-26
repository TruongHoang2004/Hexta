"""Offline deterministic keyword router for demos and tests (AI_FAKE_ROUTER=keyword). No network, no paid calls.

It has the same interface as JevRouter (async route() returning a RouteDecision or None, close()) so the graph
treats it as a primary router. Matching is diacritics-insensitive Vietnamese: customers often type without
accents. Like every router output, its decision is untrusted; Go keeps all business authority.
"""

import re

from hexta_ai.knowledge.embeddings import fold
from hexta_ai.routing.decision import RouteDecision

MODEL = "keyword-v1"

# Folded phrases (see knowledge.embeddings.fold). Matching is on whole words, so "gia" never matches "giao".
HUMAN = ("nhan vien", "gap nguoi", "nguoi that", "tu van vien", "khieu nai", "hoan tien", "tra lai tien",
         "lua dao", "boi thuong", "chua nhan duoc", "giao sai", "giao nham", "hang loi", "huy don", "don hang bi",
         "chu shop", "quan ly", "goi lai cho", "goi dien cho", "staff", "human", "complain", "refund")
ANSWER = ("gia", "bao nhieu", "bn", "con hang", "het hang", "size", "mau", "ship", "giao hang", "van chuyen",
          "phi", "doi tra", "doi hang", "tra hang", "bao hanh", "chinh sach", "thanh toan", "chat lieu",
          "khuyen mai", "giam gia", "sale", "freeship", "dia chi", "cua hang", "mo cua", "dong cua", "price",
          "stock")
VAGUE = ("alo", "hi", "hello", "chao", "chao shop", "shop oi", "ad oi", "tu van", "tu van giup", "cho hoi",
         "hoi chut", "cai nay", "cai do", "mau nay", "?")
CLOSING = ("cam on", "thanks", "thank you", "ok", "oke", "vang", "da")
# Messages with at most this many words and no topic keyword ask for clarification.
VAGUE_MAX_WORDS = 2


def _words(text: str) -> str:
    return " ".join(re.findall(r"\w+|\?", fold(text)))


def _has(text: str, phrases) -> bool:
    padded = f" {text} "
    return any(f" {phrase} " in padded for phrase in phrases)


class KeywordRouter:
    name = "keyword"
    model = MODEL
    simulated = True
    # Fallback reason recorded when no rule applies (the graph then uses the deterministic matcher).
    uncertain_reason = "keyword_no_match"

    async def route(self, message: str, scripts: list[dict], history: list[dict]) -> RouteDecision | None:
        text = _words(message)
        decision = {"router": "deterministic", "router_model": MODEL}
        script_id = self._script(text, scripts)
        if script_id:
            return RouteDecision(route="script", script_id=script_id, probability=1, reason="keyword_script_match",
                                 **decision)
        if _has(text, HUMAN):
            return RouteDecision(route="human", probability=0.9, reason="keyword_customer_requested_human",
                                 handoff_reason="customer_requested_staff", **decision)
        if _has(text, ANSWER):
            return RouteDecision(route="answer", probability=0.9, reason="keyword_knowledge_question", **decision)
        if text in CLOSING:
            return None
        if not text or text in VAGUE or len(text.split()) <= VAGUE_MAX_WORDS:
            return RouteDecision(route="clarify", probability=0.9, reason="keyword_needs_clarification", **decision)
        return None

    @staticmethod
    def _script(text: str, scripts: list[dict]) -> str | None:
        """The unique active script with a trigger phrase equal to the message (Go joins phrases with commas)."""
        matches = {s["id"] for s in scripts if s.get("active", True)
                   and any(_words(p) == text for p in s["trigger"].split(",") if _words(p))}
        return matches.pop() if len(matches) == 1 else None

    async def close(self):
        pass
