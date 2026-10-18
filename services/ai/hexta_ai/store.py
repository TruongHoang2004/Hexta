"""PostgreSQL authority for durable run state. No graph work holds DB locks."""

import hashlib
from pathlib import Path
from uuid import uuid4

from psycopg import AsyncConnection
from psycopg.rows import dict_row

from hexta.ai.v1 import service_pb2 as pb


class InputConflict(Exception):
    pass


class RunStore:
    def __init__(self, dsn: str, lease_seconds: float = 60):
        self.dsn = dsn
        self.lease_seconds = lease_seconds

    async def connect(self):
        return await AsyncConnection.connect(self.dsn, row_factory=dict_row, connect_timeout=5)

    async def migrate(self):
        async with await self.connect() as db:
            # Separate schema/owner: never run this migration over Go business tables.
            await db.execute((Path(__file__).parent / "migrations/001_runs.sql").read_text())

    async def ping(self):
        async with await self.connect() as db:
            await db.execute("SELECT 1 FROM ai_runtime.runs LIMIT 1")

    @staticmethod
    def decode(row):
        if row is None:
            return None
        run = pb.Run.FromString(bytes(row["result"]))
        run.status = row["status"]
        if row.get("error_code"):
            run.error_code = row["error_code"]
        return run

    async def submit(self, request):
        digest = hashlib.sha256(request.SerializeToString(deterministic=True)).digest()
        run = pb.Run(tenant_id=request.tenant_id, run_id=request.run_id, turn_id=request.turn_id,
                     context_revision=request.context_revision, control_generation=request.control_generation,
                     status=pb.RUN_STATUS_QUEUED, simulated=True)
        async with await self.connect() as db:
            await db.execute("""INSERT INTO ai_runtime.runs(tenant_id,run_id,request,digest,result,status)
                VALUES (%s,%s,%s,%s,%s,1) ON CONFLICT (tenant_id,run_id) DO NOTHING""",
                (request.tenant_id, request.run_id, request.SerializeToString(), digest, run.SerializeToString()))
            row = await (await db.execute("SELECT * FROM ai_runtime.runs WHERE tenant_id=%s AND run_id=%s",
                                         (request.tenant_id, request.run_id))).fetchone()
            if bytes(row["digest"]) != digest:
                raise InputConflict()
            return self.decode(row)

    async def get(self, tenant, run_id):
        async with await self.connect() as db:
            row = await (await db.execute("SELECT * FROM ai_runtime.runs WHERE tenant_id=%s AND run_id=%s",
                                         (tenant, run_id))).fetchone()
            return self.decode(row)

    async def cancel(self, tenant, run_id):
        async with await self.connect() as db:
            await db.execute("""UPDATE ai_runtime.runs SET status=5,lease_token=NULL,lease_until=NULL,
                updated_at=clock_timestamp() WHERE tenant_id=%s AND run_id=%s AND status IN (1,2)""",
                (tenant, run_id))
            row = await (await db.execute("SELECT * FROM ai_runtime.runs WHERE tenant_id=%s AND run_id=%s",
                                         (tenant, run_id))).fetchone()
            return self.decode(row)

    async def claim(self):
        token = uuid4()
        async with await self.connect() as db:
            await db.execute("""UPDATE ai_runtime.runs SET status=4,error_code='RECOVERY_ATTEMPTS_EXHAUSTED',
                lease_token=NULL,lease_until=NULL,updated_at=clock_timestamp()
                WHERE status=2 AND lease_until < clock_timestamp() AND attempts>=3""")
            row = await (await db.execute("""WITH candidate AS (
                SELECT tenant_id,run_id FROM ai_runtime.runs
                WHERE status=1 OR (status=2 AND lease_until < clock_timestamp() AND attempts<3)
                ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1)
                UPDATE ai_runtime.runs r SET status=2,attempts=attempts+1,lease_token=%s,
                lease_until=clock_timestamp()+(%s * interval '1 second'),updated_at=clock_timestamp()
                FROM candidate c WHERE r.tenant_id=c.tenant_id AND r.run_id=c.run_id RETURNING r.*""",
                (token, self.lease_seconds))).fetchone()
            return row

    async def renew(self, row):
        async with await self.connect() as db:
            cursor = await db.execute("""UPDATE ai_runtime.runs SET
                lease_until=clock_timestamp()+(%s * interval '1 second'),updated_at=clock_timestamp()
                WHERE tenant_id=%s AND run_id=%s AND lease_token=%s AND status=2
                AND lease_until>clock_timestamp()""",
                (self.lease_seconds, row["tenant_id"], row["run_id"], row["lease_token"]))
            return cursor.rowcount == 1

    async def finish(self, row, result):
        async with await self.connect() as db:
            cursor = await db.execute("""UPDATE ai_runtime.runs SET result=%s,status=%s,error_code=%s,
                lease_token=NULL,lease_until=NULL,updated_at=clock_timestamp()
                WHERE tenant_id=%s AND run_id=%s AND lease_token=%s AND status=2
                AND lease_until>clock_timestamp()""",
                (result.SerializeToString(), result.status, result.error_code,
                 row["tenant_id"], row["run_id"], row["lease_token"]))
            return cursor.rowcount == 1

    async def release(self, row):
        async with await self.connect() as db:
            await db.execute("""UPDATE ai_runtime.runs SET status=1,attempts=GREATEST(attempts-1,0),lease_token=NULL,lease_until=NULL,
                updated_at=clock_timestamp() WHERE tenant_id=%s AND run_id=%s AND lease_token=%s AND status=2""",
                (row["tenant_id"], row["run_id"], row["lease_token"]))
