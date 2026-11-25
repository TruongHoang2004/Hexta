"""Deterministic offline routing and answers, without network clients or paid inference."""

from hexta_ai.providers.base import BaseAIProvider, Candidate, Classification, GroundedAnswer, Usage
from hexta_ai.routing.decision import exact_script_match


class FakeProvider(BaseAIProvider):
    async def classify(self, context):
        script_id = exact_script_match(context["message"], context["scripts"])
        return Classification(route="script" if script_id else "human", script_id=script_id,
                              confidence=1 if script_id else 0), Usage()

    async def generate_candidate(self, context):
        return Candidate(text="Please ask a team member for assistance.", fact_ids=[]), Usage()

    async def answer(self, context):
        # Extractive stand-in for a model: quote the best retrieved record, citing it.
        knowledge = context.get("knowledge") or []
        if not knowledge:
            return GroundedAnswer(kind="answer", text="", cited_source_ids=[], confidence=0, needs_human=True), Usage()
        best = knowledge[0]
        return GroundedAnswer(kind="answer", text=best["text"][:2000], cited_source_ids=[best["id"]],
                              confidence=0.9, needs_human=False), Usage()
