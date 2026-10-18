"""Durable RPC lifecycle and leased whole-graph execution for the fake provider."""

import asyncio
import logging
from contextlib import suppress

import grpc
from psycopg import Error as DatabaseError

from hexta.ai.v1 import service_pb2 as pb
from hexta_ai.graph import PreviewRequest
from hexta_ai.rpc import AIService, run_graph
from hexta_ai.store import InputConflict, RunStore

log = logging.getLogger(__name__)


class DurableAIService(AIService):
    def __init__(self, token, store: RunStore, runner=run_graph):
        super().__init__(token, runner)
        self.store = store
        self.worker = None

    async def start(self):
        await self.store.migrate()
        self.worker = asyncio.create_task(self.work())

    async def ready(self):
        if self.worker is None or self.worker.done():
            raise RuntimeError("AI worker is unavailable")
        await self.store.ping()

    async def submit(self, request, preview, context):
        # A committed insert precedes acknowledgement; no in-memory fallback on outage.
        return await self.store.submit(request)

    async def StartRun(self, request, context):
        try:
            return await super().StartRun(request, context)
        except InputConflict:
            await context.abort(grpc.StatusCode.ALREADY_EXISTS, "Run ID has different input")
        except DatabaseError:
            await context.abort(grpc.StatusCode.UNAVAILABLE, "AI run storage unavailable")

    async def lookup(self, request, context):
        await self.authorize(context)
        await self.validate_ids(context, request.tenant_id, request.run_id)
        try:
            run = await self.store.get(request.tenant_id, request.run_id)
        except DatabaseError:
            await context.abort(grpc.StatusCode.UNAVAILABLE, "AI run storage unavailable")
        if run is None:
            await context.abort(grpc.StatusCode.NOT_FOUND, "Run not found")
        return run

    async def CancelRun(self, request, context):
        await self.authorize(context)
        await self.validate_ids(context, request.tenant_id, request.run_id)
        try:
            run = await self.store.cancel(request.tenant_id, request.run_id)
        except DatabaseError:
            await context.abort(grpc.StatusCode.UNAVAILABLE, "AI run storage unavailable")
        if run is None:
            await context.abort(grpc.StatusCode.NOT_FOUND, "Run not found")
        return run

    async def work(self):
        while True:
            try:
                row = await self.store.claim()
                if row:
                    await self.process(row)
                else:
                    await asyncio.sleep(0.2)
            except DatabaseError:
                log.warning("AI storage operation failed; retrying without logging payloads")
                await asyncio.sleep(1)

    async def process(self, row):
        request = pb.StartRunRequest.FromString(bytes(row["request"]))
        preview = PreviewRequest(message=request.message, scripts=[{
            "id": s.id, "trigger": s.trigger, "text": s.text, "active": s.active,
        } for s in request.scripts])
        task = asyncio.create_task(self.runner(preview))
        try:
            async with asyncio.timeout(60):
                while not task.done():
                    done, _ = await asyncio.wait({task}, timeout=min(1, self.store.lease_seconds / 3))
                    if not done and not await self.store.renew(row):
                        task.cancel()
                        return
                result = await task
            candidate = self.store.decode(row)
            candidate.route = result.route
            candidate.text = result.text or ""
            candidate.script_id = result.script_id or ""
            candidate.reason = result.reason
            candidate.steps.extend(result.steps)
            candidate.status = pb.RUN_STATUS_SUCCEEDED
            await self.store.finish(row, candidate)
        except asyncio.CancelledError:
            # Only graceful shutdown requeues; lost leases must retain recovery attempts.
            with suppress(DatabaseError):
                await self.store.release(row)
            raise
        except DatabaseError:
            raise
        except Exception:
            failed = self.store.decode(row)
            failed.status = pb.RUN_STATUS_FAILED
            failed.error_code = "SIMULATION_RUN_FAILED"
            await self.store.finish(row, failed)
        finally:
            task.cancel()
            with suppress(asyncio.CancelledError, Exception):
                await task

    async def close(self):
        if self.worker:
            self.worker.cancel()
            with suppress(asyncio.CancelledError):
                await self.worker
