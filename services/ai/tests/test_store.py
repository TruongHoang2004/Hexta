"""Opt-in integration tests; use only the explicit local/test database URL."""

import asyncio
import os
from uuid import uuid4

import pytest

from hexta.ai.v1 import service_pb2 as pb
from hexta_ai.durable import DurableAIService
from hexta_ai.store import InputConflict, RunStore

DSN = os.getenv("AI_TEST_DATABASE_URL")
pytestmark = pytest.mark.skipif(not DSN, reason="AI_TEST_DATABASE_URL is required")


def test_durable_submission_recovery_fencing_and_cancellation():
    async def scenario():
        store = RunStore(DSN)
        tenant = "test-" + uuid4().hex
        def request(run_id="run1", message="hello"):
            return pb.StartRunRequest(tenant_id=tenant, run_id=run_id, turn_id="turn1", message=message)
        async def expire(row, attempts=None):
            async with await store.connect() as db:
                await db.execute("""UPDATE ai_runtime.runs SET lease_until=clock_timestamp()-interval '1 second'
                    WHERE tenant_id=%s AND run_id=%s""", (tenant, row["run_id"]))
                if attempts:
                    await db.execute("UPDATE ai_runtime.runs SET attempts=%s WHERE tenant_id=%s AND run_id=%s",
                                     (attempts, tenant, row["run_id"]))
        async with await store.connect() as db:
            await db.execute("CREATE SCHEMA IF NOT EXISTS ai_runtime")
        await store.migrate()
        try:
            # Concurrent acceptance creates exactly one durable identity.
            results = await asyncio.gather(*[store.submit(request()) for _ in range(4)])
            assert all(r.run_id == "run1" for r in results)
            with pytest.raises(InputConflict):
                await store.submit(request(message="changed"))
            assert await store.get("other-tenant", "run1") is None
            # This suite is run against an otherwise idle test queue.
            claimed = await asyncio.gather(store.claim(), RunStore(DSN).claim())
            rows = [r for r in claimed if r and r["tenant_id"] == tenant]
            assert len(rows) == 1
            first = rows[0]
            await expire(first)
            recovered = await RunStore(DSN).claim()
            assert recovered["lease_token"] != first["lease_token"]
            result = store.decode(recovered)
            result.status = pb.RUN_STATUS_SUCCEEDED
            result.text = "candidate"
            assert not await store.finish(first, result)
            assert await store.finish(recovered, result)
            assert (await RunStore(DSN).get(tenant, "run1")).text == "candidate"
            assert (await store.cancel(tenant, "run1")).status == pb.RUN_STATUS_SUCCEEDED
            assert (await store.submit(request())).status == pb.RUN_STATUS_SUCCEEDED

            await store.submit(request("cancel"))
            row = await store.claim()
            assert (await store.cancel(tenant, "cancel")).status == pb.RUN_STATUS_CANCELLED
            assert not await store.renew(row)
            assert not await store.finish(row, result)
            assert (await RunStore(DSN).get(tenant, "cancel")).status == pb.RUN_STATUS_CANCELLED

            await store.submit(request("exhausted"))
            row = await store.claim()
            await expire(row, attempts=3)
            await store.claim()
            exhausted = await store.get(tenant, "exhausted")
            assert exhausted.status == pb.RUN_STATUS_FAILED
            assert exhausted.error_code == "RECOVERY_ATTEMPTS_EXHAUSTED"

            # A replacement worker executes a request accepted by another instance.
            await store.submit(request("replacement"))
            service = DurableAIService("integration-test-service-token", RunStore(DSN))
            await service.start()
            try:
                async with asyncio.timeout(5):
                    while (await store.get(tenant, "replacement")).status != pb.RUN_STATUS_SUCCEEDED:
                        await asyncio.sleep(0.02)
            finally:
                await service.close()
        finally:
            async with await store.connect() as db:
                await db.execute("DELETE FROM ai_runtime.runs WHERE tenant_id=%s", (tenant,))
    asyncio.run(scenario())
