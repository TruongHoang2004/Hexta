"""Runtime startup checks readiness and never provisions schema."""

import asyncio

import pytest

from hexta_ai.durable import DurableAIService


def test_start_only_checks_storage_before_worker():
    class ReadyStore:
        def __init__(self):
            self.reads = 0

        async def ping(self):
            self.reads += 1

        async def claim(self):
            await asyncio.Event().wait()

    async def scenario():
        store = ReadyStore()
        service = DurableAIService("test-startup-service-token", store)
        await service.start()
        try:
            assert store.reads == 1
            assert service.worker is not None
        finally:
            await service.close()

    asyncio.run(scenario())


def test_missing_schema_prevents_worker_start():
    class MissingSchemaStore:
        async def ping(self):
            raise RuntimeError("schema not provisioned")

    async def scenario():
        service = DurableAIService("test-startup-service-token", MissingSchemaStore())
        with pytest.raises(RuntimeError, match="schema not provisioned"):
            await service.start()
        assert service.worker is None

    asyncio.run(scenario())
