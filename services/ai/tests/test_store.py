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
        await store.ping()  # The explicit Atlas test job provisions the schema.
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


def test_paid_marker_is_committed_fenced_and_retained_on_release():
    from unittest.mock import AsyncMock
    from hexta_ai.providers.fake import FakeProvider

    class PaidProvider(FakeProvider):
        simulated = False
        name = "gemini"
        model = "mock-model"

    async def scenario():
        store = RunStore(DSN)
        tenant = "paid-" + uuid4().hex
        request = pb.StartRunRequest(tenant_id=tenant, run_id="paid", turn_id="turn", message="hello")
        try:
            await store.submit(request, simulated=False, provider="gemini", model="mock-model")
            row = await store.claim()
            assert row["tenant_id"] == tenant
            assert await store.begin_paid_execution(row)
            assert not await store.begin_paid_execution(row)
            # Another connection sees the marker before a remote call could start.
            assert (await RunStore(DSN).get(tenant, "paid")).error_code == "LLM_CALL_STARTED"
            await store.release(row)
            recovered = await store.claim()
            assert recovered["error_code"] == "LLM_CALL_STARTED"
            assert not await store.begin_paid_execution(row)  # Old lease cannot start work.
            runner = AsyncMock()
            service = DurableAIService("test-paid-service-token", store, runner, PaidProvider())
            await service.process(recovered)
            runner.assert_not_awaited()
            saved = await store.get(tenant, "paid")
            assert saved.status == pb.RUN_STATUS_FAILED and saved.error_code == "LLM_RECOVERY_UNCERTAIN"
            assert (await store.submit(request)).status == pb.RUN_STATUS_FAILED
        finally:
            async with await store.connect() as db:
                await db.execute("DELETE FROM ai_runtime.runs WHERE tenant_id=%s", (tenant,))
    asyncio.run(scenario())


def test_durable_grpc_persists_provider_metadata_and_digest_covers_context():
    import grpc
    from hexta.ai.v1 import service_pb2_grpc as rpc
    from hexta_ai.graph import PreviewResult
    from hexta_ai.providers.base import Usage
    from hexta_ai.providers.fake import FakeProvider
    from hexta_ai.rpc import start_server

    class PaidProvider(FakeProvider):
        simulated = False
        name = "gemini"
        model = "mock-model"

    async def scenario():
        store = RunStore(DSN)
        tenant = "grpc-paid-" + uuid4().hex
        token = "test-durable-grpc-provider-token"
        auth = (("authorization", f"Bearer {token}"),)
        calls = 0
        async def runner(value):
            nonlocal calls
            calls += 1
            assert value.context_revision == 3 and value.facts[0].text == "Authorized fact."
            return PreviewResult(route="generate", text="Authorized fact.", reason="grounded_domain_candidate",
                                 simulated=False, provider="gemini", model="mock-model", confidence=0.95,
                                 usage=Usage(input_tokens=21, output_tokens=9), steps=["validate_output"])
        service = DurableAIService(token, store, runner, PaidProvider())
        await service.start()
        server = grpc.aio.server()
        rpc.add_AIServiceServicer_to_server(service, server)
        port = server.add_insecure_port("127.0.0.1:0")
        await server.start()
        try:
            async with grpc.aio.insecure_channel(f"127.0.0.1:{port}") as channel:
                stub = rpc.AIServiceStub(channel)
                value = pb.StartRunRequest(tenant_id=tenant, run_id="run", turn_id="turn", message="question",
                    context_revision=3, allow_generation=True, facts=[pb.DomainFact(id="f1", text="Authorized fact.")])
                await stub.StartRun(value, metadata=auth)
                async with asyncio.timeout(5):
                    while True:
                        saved = await stub.GetRun(pb.RunRequest(tenant_id=tenant, run_id="run"), metadata=auth)
                        if saved.status == pb.RUN_STATUS_SUCCEEDED:
                            break
                        await asyncio.sleep(0.02)
                assert saved.provider == "gemini" and saved.input_tokens == 21 and not saved.simulated
                assert saved.text == "Authorized fact." and saved.context_revision == 3
                await stub.StartRun(value, metadata=auth)
                assert calls == 1
                value.facts[0].text = "Changed fact."
                with pytest.raises(grpc.aio.AioRpcError) as error:
                    await stub.StartRun(value, metadata=auth)
                assert error.value.code() == grpc.StatusCode.ALREADY_EXISTS
        finally:
            await server.stop(0)
            await service.close()
            async with await store.connect() as db:
                await db.execute("DELETE FROM ai_runtime.runs WHERE tenant_id=%s", (tenant,))
    asyncio.run(scenario())
