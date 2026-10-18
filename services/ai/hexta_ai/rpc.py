"""Shared RPC validation and offline memory adapter; runtime selects durable storage."""

import asyncio
import hashlib
import os
import re
import secrets
from collections.abc import Awaitable, Callable

import grpc
from langsmith import tracing_context
from pydantic import ValidationError

from hexta.ai.v1 import service_pb2 as pb
from hexta.ai.v1 import service_pb2_grpc as rpc
from hexta_ai.graph import PreviewRequest, PreviewResult, preview_graph


async def run_graph(request: PreviewRequest) -> PreviewResult:
    with tracing_context(enabled=False):
        state = await preview_graph.ainvoke({"request": request}, {"recursion_limit": 8})
    return state["result"]


class AIService(rpc.AIServiceServicer):
    def __init__(self, token: str, runner: Callable[[PreviewRequest], Awaitable[PreviewResult]] = run_graph,
                 capacity: int = 1000):
        self.token = token
        self.runner = runner
        self.capacity = capacity
        self.runs: dict[tuple[str, str], pb.Run] = {}
        self.digests: dict[tuple[str, str], bytes] = {}
        self.tasks: dict[tuple[str, str], asyncio.Task] = {}

    async def authorize(self, context):
        credentials = [value for key, value in context.invocation_metadata() if key == "authorization"]
        if len(credentials) != 1 or not secrets.compare_digest(
            credentials[0].encode(), f"Bearer {self.token}".encode()
        ):
            await context.abort(grpc.StatusCode.UNAUTHENTICATED, "Invalid service credentials")

    async def validate_ids(self, context, *values):
        if any(not re.fullmatch(r"[A-Za-z0-9_-]{1,100}", value) for value in values):
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "Invalid identifier")

    async def StartRun(self, request, context):
        await self.authorize(context)
        await self.validate_ids(context, request.tenant_id, request.run_id, request.turn_id)
        if request.context_revision < 0 or request.control_generation < 0:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "Revisions must be nonnegative")
        try:
            preview = PreviewRequest(message=request.message, scripts=[{
                "id": script.id, "trigger": script.trigger, "text": script.text, "active": script.active,
            } for script in request.scripts])
        except ValidationError:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, "Invalid preview input")
        return await self.submit(request, preview, context)

    async def submit(self, request, preview, context):
        key = (request.tenant_id, request.run_id)
        digest = hashlib.sha256(request.SerializeToString(deterministic=True)).digest()
        if key in self.runs:
            if self.digests[key] != digest:
                await context.abort(grpc.StatusCode.ALREADY_EXISTS, "Run ID has different input")
            return self.runs[key]
        if len(self.runs) >= self.capacity:
            await context.abort(grpc.StatusCode.RESOURCE_EXHAUSTED, "Simulation run capacity reached")
        run = pb.Run(tenant_id=request.tenant_id, run_id=request.run_id, turn_id=request.turn_id,
                     context_revision=request.context_revision, control_generation=request.control_generation,
                     status=pb.RUN_STATUS_QUEUED, simulated=True)
        self.runs[key] = run
        self.digests[key] = digest
        self.tasks[key] = asyncio.create_task(self.execute(key, preview))
        self.tasks[key].add_done_callback(lambda _: self.tasks.pop(key, None))
        return run

    async def execute(self, key, request):
        run = self.runs[key]
        try:
            run.status = pb.RUN_STATUS_RUNNING
            result = await self.runner(request)
            if run.status == pb.RUN_STATUS_CANCELLED:
                return
            # Publish only after every candidate field has serialized successfully.
            completed = pb.Run()
            completed.CopyFrom(run)
            completed.route = result.route
            completed.text = result.text or ""
            completed.script_id = result.script_id or ""
            completed.reason = result.reason
            completed.steps.extend(result.steps)
            completed.status = pb.RUN_STATUS_SUCCEEDED
            self.runs[key] = completed
        except asyncio.CancelledError:
            run.status = pb.RUN_STATUS_CANCELLED
        except Exception:
            if run.status == pb.RUN_STATUS_CANCELLED:
                return
            # Do not expose prompts, secrets, or provider exception bodies over RPC.
            run.status = pb.RUN_STATUS_FAILED
            run.error_code = "SIMULATION_RUN_FAILED"

    async def lookup(self, request, context):
        await self.authorize(context)
        await self.validate_ids(context, request.tenant_id, request.run_id)
        run = self.runs.get((request.tenant_id, request.run_id))
        if run is None:
            await context.abort(grpc.StatusCode.NOT_FOUND, "Run not found")
        return run

    async def GetRun(self, request, context):
        return await self.lookup(request, context)

    async def CancelRun(self, request, context):
        run = await self.lookup(request, context)
        if run.status in (pb.RUN_STATUS_QUEUED, pb.RUN_STATUS_RUNNING):
            run.status = pb.RUN_STATUS_CANCELLED
            task = self.tasks.get((request.tenant_id, request.run_id))
            if task:
                task.cancel()
        return run

    async def close(self):
        tasks = list(self.tasks.values())
        for task in tasks:
            task.cancel()
        await asyncio.gather(*tasks, return_exceptions=True)
        self.tasks.clear()


async def start_server(token: str, address: str, service: AIService | None = None):
    if service is None:
        dsn = os.getenv("AI_DATABASE_URL")
        if not dsn:
            raise RuntimeError("AI_DATABASE_URL is required for the gRPC runtime")
        from hexta_ai.durable import DurableAIService
        from hexta_ai.store import RunStore
        service = DurableAIService(token, RunStore(dsn))
        await service.start()
    server = grpc.aio.server(options=[("grpc.max_receive_message_length", 1024 * 1024)])
    rpc.add_AIServiceServicer_to_server(service, server)
    if not server.add_insecure_port(address):
        raise RuntimeError("Cannot bind AI gRPC listener")
    await server.start()
    return server, service
