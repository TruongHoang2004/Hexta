"""Durable execution with fake replay and conservative paid-call recovery."""

import asyncio
import logging
from contextlib import suppress

import grpc
from psycopg import Error as DatabaseError

from hexta.ai.v1 import service_pb2 as pb
from hexta_ai.providers.base import ProviderError
from hexta_ai.run_mapper import completed_run, preview_request
from hexta_ai.rpc import AIService, run_graph
from hexta_ai.store import InputConflict, RunStore

log = logging.getLogger(__name__)


class DurableAIService(AIService):
    def __init__(self, token, store: RunStore, runner=run_graph, provider=None):
        super().__init__(token, runner, provider=provider)
        self.store = store
        self.worker = None

    async def start(self):
        await self.store.ping()  # Atlas must apply schema changes before service startup.
        self.worker = asyncio.create_task(self.work())

    async def ready(self):
        if self.worker is None or self.worker.done():
            raise RuntimeError("AI worker is unavailable")
        await self.store.ping()

    async def submit(self, request, preview, context):
        # A committed insert precedes acknowledgement; no in-memory fallback on outage.
        return await self.store.submit(request, simulated=self.provider.simulated if self.provider else True,
                                       provider=self.provider.name if self.provider else "fake",
                                       model=self.provider.model if self.provider else "deterministic")

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
        task = None
        try:
            saved = pb.Run.FromString(bytes(row["result"]))
            # A stored candidate always wins over replay, even for a legacy partial row.
            if saved.status == pb.RUN_STATUS_SUCCEEDED:
                await self.store.finish(row, saved)
                return
            if row.get("error_code") == "LLM_CALL_STARTED":
                raise ProviderError("LLM_RECOVERY_UNCERTAIN")
            if self.provider:
                if (saved.simulated != self.provider.simulated
                        or saved.provider not in ("", self.provider.name)
                        or saved.model not in ("", self.provider.model)):
                    raise ProviderError("PROVIDER_CONFIG_CHANGED")
                if not self.provider.simulated and not await self.store.begin_paid_execution(row):
                    raise ProviderError("LLM_RECOVERY_UNCERTAIN")
            request = pb.StartRunRequest.FromString(bytes(row["request"]))
            preview = preview_request(request)
            task = asyncio.create_task(self.runner(preview))
            async with asyncio.timeout(65):
                while not task.done():
                    done, _ = await asyncio.wait({task}, timeout=min(1, self.store.lease_seconds / 3))
                    if not done and not await self.store.renew(row):
                        task.cancel()
                        return
                result = await task
            candidate = completed_run(self.store.decode(row), result)
            await self.store.finish(row, candidate)
        except asyncio.CancelledError:
            # Only graceful shutdown requeues; lost leases must retain recovery attempts.
            with suppress(DatabaseError):
                await self.store.release(row)
            raise
        except DatabaseError:
            raise
        except Exception as error:
            failed = self.store.decode(row)
            failed.status = pb.RUN_STATUS_FAILED
            failed.error_code = (error.code if isinstance(error, ProviderError) else
                                 "LLM_PROVIDER_ERROR" if self.provider and not self.provider.simulated else
                                 "SIMULATION_RUN_FAILED")
            await self.store.finish(row, failed)
        finally:
            if task:
                task.cancel()
                with suppress(asyncio.CancelledError, Exception):
                    await task

    async def close(self):
        if self.worker:
            self.worker.cancel()
            with suppress(asyncio.CancelledError):
                await self.worker
