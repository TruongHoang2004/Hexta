"""Runtime startup checks readiness and never provisions schema."""

import asyncio

import psycopg
import pytest
from psycopg_pool import PoolClosed, PoolTimeout

from hexta_ai import rpc
from hexta_ai.durable import DurableAIService
from hexta_ai.store import RunStore

UNREACHABLE_DSN = "postgresql://startup:unused@127.0.0.1:1/none"


def test_start_only_checks_storage_before_worker():
    class ReadyStore:
        def __init__(self):
            self.reads = 0
            self.closed = False

        async def ping(self):
            self.reads += 1

        async def close(self):
            self.closed = True

        async def claim(self):
            await asyncio.Event().wait()

    async def scenario():
        store = ReadyStore()
        service = DurableAIService("test-startup-service-token", store)
        await service.start()
        try:
            assert store.reads == 1
            assert service.worker is not None
            assert not store.closed
        finally:
            await service.close()
        # The worker is stopped before its storage pool is released.
        assert store.closed and service.worker.done()

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


def test_failed_startup_releases_storage_pool(monkeypatch):
    stores = []

    class MissingSchemaStore:
        def __init__(self, dsn):
            self.closed = False
            stores.append(self)

        async def ping(self):
            raise RuntimeError("schema not provisioned")

        async def close(self):
            self.closed = True

    monkeypatch.setenv("AI_DATABASE_URL", UNREACHABLE_DSN)
    monkeypatch.setattr("hexta_ai.store.RunStore", MissingSchemaStore)

    async def scenario():
        with pytest.raises(RuntimeError, match="schema not provisioned"):
            await rpc.start_server("test-startup-service-token", "127.0.0.1:0")

    asyncio.run(scenario())
    assert len(stores) == 1 and stores[0].closed


def test_run_store_pool_is_bounded_and_lazy():
    store = RunStore(UNREACHABLE_DSN)
    assert (store.pool.min_size, store.pool.max_size) == (2, 10)
    assert store.pool.timeout == 5
    assert store.pool.closed  # Construction never opens network connections.


def test_closed_run_store_rejects_queries_without_reopening():
    async def scenario():
        store = RunStore(UNREACHABLE_DSN)
        await store.close()
        with pytest.raises(PoolClosed):
            await store.get("tenant", "run")
        assert store.pool.closed

    asyncio.run(scenario())


def test_pool_failures_use_existing_storage_error_handling():
    # DurableAIService maps psycopg.Error to UNAVAILABLE and the worker retries on it.
    assert issubclass(PoolTimeout, psycopg.Error)
    assert issubclass(PoolClosed, psycopg.Error)
