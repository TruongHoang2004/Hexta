import asyncio

import grpc
import pytest

from hexta.ai.v1 import service_pb2 as pb
from hexta.ai.v1 import service_pb2_grpc as rpc
from hexta_ai.rpc import AIService

TOKEN = "test-grpc-service-token-at-least-24"
AUTH = (("authorization", f"Bearer {TOKEN}"),)


def request(**changes):
    fields = dict(tenant_id="tenant1", run_id="run1", turn_id="turn1", context_revision=2,
                  control_generation=3, message="price", scripts=[pb.Script(
                      id="s1", trigger="price", text="Please select a product.", active=True)])
    return pb.StartRunRequest(**(fields | changes))


async def harness(check, runner=None, capacity=1000):
    service = AIService(TOKEN, capacity=capacity) if runner is None else AIService(TOKEN, runner, capacity)
    server = grpc.aio.server()
    rpc.add_AIServiceServicer_to_server(service, server)
    port = server.add_insecure_port("127.0.0.1:0")
    await server.start()
    try:
        async with grpc.aio.insecure_channel(f"127.0.0.1:{port}") as channel:
            await check(rpc.AIServiceStub(channel), service)
    finally:
        await server.stop(0)
        await service.close()


async def expect_error(call, code):
    with pytest.raises(grpc.aio.AioRpcError) as error:
        await call
    assert error.value.code() == code


def test_auth_validation_and_tenant_lookup():
    async def check(stub, _):
        await expect_error(stub.StartRun(request()), grpc.StatusCode.UNAUTHENTICATED)
        await expect_error(stub.StartRun(request(message=""), metadata=AUTH), grpc.StatusCode.INVALID_ARGUMENT)
        await expect_error(stub.StartRun(request(context_revision=-1), metadata=AUTH), grpc.StatusCode.INVALID_ARGUMENT)
        await stub.StartRun(request(), metadata=AUTH)
        await expect_error(stub.GetRun(pb.RunRequest(tenant_id="tenant2", run_id="run1"), metadata=AUTH),
                           grpc.StatusCode.NOT_FOUND)
        await expect_error(stub.CancelRun(pb.RunRequest(tenant_id="tenant1", run_id="run1")),
                           grpc.StatusCode.UNAUTHENTICATED)
    asyncio.run(harness(check))


def test_idempotency_and_candidate_result():
    async def check(stub, service):
        first = await stub.StartRun(request(), metadata=AUTH)
        await stub.StartRun(request(), metadata=AUTH)
        await expect_error(stub.StartRun(request(message="changed"), metadata=AUTH), grpc.StatusCode.ALREADY_EXISTS)
        if service.tasks:
            await asyncio.gather(*list(service.tasks.values()))
        result = await stub.GetRun(pb.RunRequest(tenant_id="tenant1", run_id=first.run_id), metadata=AUTH)
        assert len(service.runs) == 1
        assert result.status == pb.RUN_STATUS_SUCCEEDED
        assert result.simulated and result.route == "script"
        assert result.text == "Please select a product."
        assert result.context_revision == 2 and result.control_generation == 3
        assert (await stub.CancelRun(pb.RunRequest(tenant_id="tenant1", run_id="run1"), metadata=AUTH)).status == pb.RUN_STATUS_SUCCEEDED
    asyncio.run(harness(check))


def test_cancellation_and_capacity():
    started = None

    async def slow(_):
        started.set()
        await asyncio.Event().wait()

    async def check(stub, service):
        nonlocal started
        started = asyncio.Event()
        await stub.StartRun(request(), metadata=AUTH)
        await started.wait()
        await expect_error(stub.StartRun(request(run_id="run2"), metadata=AUTH), grpc.StatusCode.RESOURCE_EXHAUSTED)
        result = await stub.CancelRun(pb.RunRequest(tenant_id="tenant1", run_id="run1"), metadata=AUTH)
        assert result.status == pb.RUN_STATUS_CANCELLED and not result.text
        if service.tasks:
            await asyncio.gather(*list(service.tasks.values()))
        assert service.runs[("tenant1", "run1")].status == pb.RUN_STATUS_CANCELLED
    asyncio.run(harness(check, slow, capacity=1))


def test_failure_is_sanitized():
    async def broken(_):
        raise RuntimeError("secret-provider-key")

    async def check(stub, service):
        await stub.StartRun(request(), metadata=AUTH)
        if service.tasks:
            await asyncio.gather(*list(service.tasks.values()))
        result = await stub.GetRun(pb.RunRequest(tenant_id="tenant1", run_id="run1"), metadata=AUTH)
        assert result.status == pb.RUN_STATUS_FAILED
        assert result.error_code == "SIMULATION_RUN_FAILED"
        assert "secret" not in str(result)
    asyncio.run(harness(check, broken))


def test_cancelled_candidate_cannot_be_restored_by_runner():
    async def stubborn(_):
        try:
            await asyncio.Event().wait()
        except asyncio.CancelledError:
            from hexta_ai.graph import PreviewResult
            return PreviewResult(route="script", text="late answer", reason="late", steps=[])

    async def check(stub, service):
        await stub.StartRun(request(), metadata=AUTH)
        await stub.CancelRun(pb.RunRequest(tenant_id="tenant1", run_id="run1"), metadata=AUTH)
        if service.tasks:
            await asyncio.gather(*list(service.tasks.values()))
        result = await stub.GetRun(pb.RunRequest(tenant_id="tenant1", run_id="run1"), metadata=AUTH)
        assert result.status == pb.RUN_STATUS_CANCELLED
        assert not result.text
    asyncio.run(harness(check, stubborn))


@pytest.mark.parametrize("field", ["message", "id", "trigger", "text"])
def test_whitespace_only_rpc_input_is_rejected(field):
    async def check(stub, service):
        value = request()
        if field == "message":
            value.message = " \t\n\u2003"
        else:
            setattr(value.scripts[0], field, " \t\n\u2003")
        await expect_error(stub.StartRun(value, metadata=AUTH), grpc.StatusCode.INVALID_ARGUMENT)
        assert not service.runs
    asyncio.run(harness(check))


def test_serialization_failure_does_not_publish_partial_candidate():
    from hexta_ai.graph import PreviewResult

    async def malformed(_):
        # Simulate a faulty adapter bypassing Pydantic: the final field fails protobuf serialization.
        return PreviewResult.model_construct(route="script", text="partial answer", script_id="s1",
                                             reason="test", steps=[object()])

    async def check(stub, service):
        await stub.StartRun(request(), metadata=AUTH)
        if service.tasks:
            await asyncio.gather(*list(service.tasks.values()))
        result = await stub.GetRun(pb.RunRequest(tenant_id="tenant1", run_id="run1"), metadata=AUTH)
        assert result.status == pb.RUN_STATUS_FAILED
        assert result.error_code == "SIMULATION_RUN_FAILED"
        assert not any((result.route, result.text, result.script_id, result.reason, result.steps))
        assert result.context_revision == 2 and result.control_generation == 3
    asyncio.run(harness(check, malformed))
