"""A lease recovery or graceful shutdown must never repeat uncertain paid work."""

import asyncio
from unittest.mock import AsyncMock

from hexta.ai.v1 import service_pb2 as pb
from hexta_ai.durable import DurableAIService
from hexta_ai.graph import PreviewResult
from hexta_ai.providers.base import Usage
from hexta_ai.providers.fake import FakeProvider


class PaidProvider(FakeProvider):
    name = "gemini"
    model = "mock-model"
    simulated = False


class Store:
    lease_seconds = 60
    decode = staticmethod(lambda row: pb.Run.FromString(row["result"]))

    def __init__(self, marker="", saved=None):
        self.row = {"request": pb.StartRunRequest(tenant_id="t", run_id="r", turn_id="turn", message="hello").SerializeToString(),
                    "result": (saved or pb.Run(tenant_id="t", run_id="r", simulated=False, provider="gemini", model="mock-model")).SerializeToString(),
                    "error_code": marker}
        self.marked = bool(marker)
        self.finished = None
        self.released = False

    async def begin_paid_execution(self, row):
        if self.marked:
            return False
        self.marked = True
        self.row["error_code"] = "LLM_CALL_STARTED"
        return True

    async def finish(self, row, result):
        self.finished = result
        return True

    async def release(self, row):
        self.released = True

    async def renew(self, row):
        return True


def result():
    return PreviewResult(simulated=False, provider="gemini", model="mock-model", confidence=0.9,
                         route="clarify", text="Which product?", reason="needs_clarification", steps=[],
                         usage=Usage(input_tokens=12, output_tokens=6))


def test_live_run_marks_before_inference_and_preserves_result_metadata():
    async def scenario():
        store = Store()
        async def runner(request):
            assert store.marked
            return result()
        await DurableAIService("test-token", store, runner, PaidProvider()).process(dict(store.row))
        assert store.finished.status == pb.RUN_STATUS_SUCCEEDED
        assert store.finished.input_tokens == 12 and store.finished.provider == "gemini"
        assert not store.finished.simulated and not store.finished.error_code
    asyncio.run(scenario())


def test_uncertain_recovery_never_invokes_paid_provider():
    async def scenario():
        store = Store("LLM_CALL_STARTED")
        runner = AsyncMock()
        await DurableAIService("test-token", store, runner, PaidProvider()).process(dict(store.row))
        runner.assert_not_awaited()
        assert store.finished.error_code == "LLM_RECOVERY_UNCERTAIN"
        assert store.finished.status == pb.RUN_STATUS_FAILED
    asyncio.run(scenario())


def test_already_stored_output_is_reused_without_paid_call():
    async def scenario():
        saved = pb.Run(tenant_id="t", run_id="r", status=pb.RUN_STATUS_SUCCEEDED, text="saved", provider="gemini")
        store = Store("LLM_CALL_STARTED", saved)
        runner = AsyncMock()
        await DurableAIService("test-token", store, runner, PaidProvider()).process(dict(store.row))
        runner.assert_not_awaited()
        assert store.finished.text == "saved" and store.finished.status == pb.RUN_STATUS_SUCCEEDED
    asyncio.run(scenario())


def test_provider_change_does_not_convert_fake_queue_to_paid_calls():
    async def scenario():
        store = Store(saved=pb.Run(simulated=True, provider="fake", model="deterministic"))
        runner = AsyncMock()
        await DurableAIService("test-token", store, runner, PaidProvider()).process(dict(store.row))
        runner.assert_not_awaited()
        assert not store.marked and store.finished.error_code == "PROVIDER_CONFIG_CHANGED"
    asyncio.run(scenario())


def test_failed_marker_commit_and_lost_lease_cannot_start_inference():
    async def scenario():
        store = Store()
        store.begin_paid_execution = AsyncMock(return_value=False)
        runner = AsyncMock()
        await DurableAIService("test-token", store, runner, PaidProvider()).process(dict(store.row))
        runner.assert_not_awaited()
        assert store.finished.error_code == "LLM_RECOVERY_UNCERTAIN"
    asyncio.run(scenario())


def test_shutdown_release_keeps_paid_marker_and_recovery_fails_without_replay():
    async def scenario():
        store = Store()
        started = asyncio.Event()
        async def slow(request):
            started.set()
            await asyncio.Event().wait()
        service = DurableAIService("test-token", store, slow, PaidProvider())
        task = asyncio.create_task(service.process(dict(store.row)))
        await started.wait()
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass
        assert store.released and store.row["error_code"] == "LLM_CALL_STARTED"
        runner = AsyncMock()
        await DurableAIService("test-token", store, runner, PaidProvider()).process(dict(store.row))
        runner.assert_not_awaited()
        assert store.finished.error_code == "LLM_RECOVERY_UNCERTAIN"
    asyncio.run(scenario())
