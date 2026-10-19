"""Docker-only diagnostics. Creates a disposable local account/tenant, never sends messages."""

import asyncio
import json
import os
import sys
import urllib.request
from uuid import uuid4

import grpc
from hexta.ai.v1 import service_pb2 as pb
from hexta.ai.v1 import service_pb2_grpc as rpc


def http(url, body=None, token=None):
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    req = urllib.request.Request(url, data=json.dumps(body).encode() if body is not None else None,
                                 headers=headers)
    with urllib.request.urlopen(req, timeout=10) as response:
        return json.load(response)


def application_check():
    with urllib.request.urlopen("http://web:3000/login", timeout=10) as response:
        assert response.status == 200
    health = http("http://api:8080/api/v1/health")
    assert health["data"]["status"] == "up"
    unique = uuid4().hex
    credentials = {"email": f"smoke-{unique}@example.test", "password": uuid4().hex}
    http("http://api:8080/api/v1/auth/register", credentials)
    login = http("http://api:8080/api/v1/auth/login", credentials)
    token = login["data"]["access_token"]
    created = http("http://api:8080/api/v1/tenants", {"name": "Local smoke", "slug": "smoke-" + unique}, token)
    tenants = http("http://api:8080/api/v1/tenants", token=token)
    assert created["data"]["id"] in [tenant["id"] for tenant in tenants["data"]]
    print("PASS: web login page, API DB/cache health, register/login, authenticated tenant create/list")


async def run_check(phase, run_id):
    auth = (("authorization", "Bearer " + os.environ["AI_INTERNAL_TOKEN"]),)
    request = pb.StartRunRequest(tenant_id="local-smoke", run_id=run_id, turn_id="turn1", message="hello",
                                context_revision=1, control_generation=1)
    async with grpc.aio.insecure_channel("127.0.0.1:50051") as channel:
        client = rpc.AIServiceStub(channel)
        if phase == "submit":
            await client.StartRun(request, metadata=auth, timeout=5)
        # Check phase must find existing state before retrying, proving restart durability.
        async with asyncio.timeout(10):
            while True:
                result = await client.GetRun(pb.RunRequest(tenant_id=request.tenant_id, run_id=run_id),
                                             metadata=auth, timeout=5)
                if result.status == pb.RUN_STATUS_SUCCEEDED:
                    break
                assert result.status in (pb.RUN_STATUS_QUEUED, pb.RUN_STATUS_RUNNING)
                await asyncio.sleep(0.05)
        duplicate = await client.StartRun(request, metadata=auth, timeout=5)
        assert duplicate.status == pb.RUN_STATUS_SUCCEEDED
        assert result.simulated and result.route == "human"
    print("PASS: persisted run, idempotent retry and simulated result")


if __name__ == "__main__":
    if len(sys.argv) == 1:
        application_check()
    else:
        asyncio.run(run_check(sys.argv[1], sys.argv[2]))
