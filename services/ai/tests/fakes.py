"""Offline doubles: a scripted Jev client and an in-memory knowledge service. No network or DB."""

from types import SimpleNamespace

from hexta_ai.knowledge.models import KnowledgeItem
from hexta_ai.routing.jev import JevConfig, JevRouter

JEV = JevConfig(enabled=True, api_key="test-only-typesafe-key")


def jev_response(intent="answer", probability=0.9, wants_human=0.1, script=None, script_probability=0.9,
                 script_fits=0.9, others=None):
    probabilities = {intent: probability, **(others or {})}
    choices = {"intent": SimpleNamespace(choice=intent, confidence=probability, probabilities=probabilities)}
    nouls = {"wants_human": SimpleNamespace(noul=wants_human)}
    if script is not None:
        choices["script"] = SimpleNamespace(choice=script, confidence=script_probability,
                                            probabilities={script: script_probability})
        nouls["script_fits"] = SimpleNamespace(noul=script_fits)
    return SimpleNamespace(choices=choices, nouls=nouls)


class FakeJevClient:
    def __init__(self, response=None, error=None):
        self.response = response if response is not None else jev_response()
        self.error = error
        self.calls = []
        self.closed = False

    async def system_one(self, state, questions, *, model=None):
        self.calls.append({"state": state, "questions": questions, "model": model})
        if self.error is not None:
            raise self.error
        return self.response

    async def aclose(self):
        self.closed = True


def fake_router(response=None, error=None, **config):
    client = FakeJevClient(response, error)
    return JevRouter(JEV.model_copy(update=config), client), client


class MemoryKnowledge:
    """Retrieval double returning fixed items for one tenant; records queries."""

    simulated = True

    def __init__(self, items=None, tenant="tenant1", error=None):
        self.items = items or []
        self.tenant = tenant
        self.error = error
        self.queries = []

    async def retrieve(self, tenant, query, *, active_script_ids):
        self.queries.append((tenant, query, set(active_script_ids)))
        if self.error:
            raise self.error
        return [KnowledgeItem(**i) for i in self.items] if tenant == self.tenant else []

    async def ping(self):
        pass

    async def close(self):
        pass
