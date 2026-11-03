"""Deterministic offline routing, without network clients or paid inference."""

from hexta_ai.providers.base import BaseAIProvider, Candidate, Classification, Usage


class FakeProvider(BaseAIProvider):
    async def classify(self, context):
        matches = [s for s in context["scripts"]
                   if s["trigger"].strip().casefold() == context["message"].strip().casefold()]
        unique = len(matches) == 1
        return Classification(route="script" if unique else "human",
                              script_id=matches[0]["id"] if unique else None,
                              confidence=1 if unique else 0), Usage()

    async def generate_candidate(self, context):
        return Candidate(text="Please ask a team member for assistance.", fact_ids=[]), Usage()
