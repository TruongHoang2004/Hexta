"""Internal, simulation-only HTTP and gRPC bootstrap. No provider or Pancake credentials."""

import os
import secrets
from contextlib import asynccontextmanager
from typing import Annotated

from fastapi import Depends, FastAPI, Header, HTTPException
from langsmith import tracing_context

from hexta_ai.rpc import start_server
from hexta_ai.graph import PreviewRequest, PreviewResult, preview_graph
from hexta_ai.schemas.consultant import ConsultantRequest, ConsultantResponse
from hexta_ai.graph_consultant import consultant_graph

SUPPORTED_PROVIDERS = {"fake", "openai", "chatgpt", "google", "gemini", "ollama", "local"}


def create_app(internal_token: str | None = None) -> FastAPI:
    token = internal_token if internal_token is not None else os.getenv("AI_INTERNAL_TOKEN", "")
    if len(token) < 24:
        raise RuntimeError("AI_INTERNAL_TOKEN must contain at least 24 characters")
    provider = os.getenv("AI_PROVIDER", "fake").strip().lower()
    if provider not in SUPPORTED_PROVIDERS:
        raise RuntimeError(
            f"Unsupported AI_PROVIDER '{provider}'. Supported providers: {', '.join(sorted(SUPPORTED_PROVIDERS))}"
        )

    runtime = {}

    @asynccontextmanager
    async def lifespan(app):
        server = service = None
        if os.getenv("AI_GRPC_ENABLED", "true").lower() == "true":
            server, service = await start_server(token, os.getenv("AI_GRPC_ADDRESS", "0.0.0.0:50051"))
        runtime["service"] = service
        try:
            yield
        finally:
            if server:
                await server.stop(grace=2)
                await service.close()

    app = FastAPI(title="Hexta AI Service", version="0.2.0", docs_url=None, redoc_url=None, lifespan=lifespan)

    def authenticate(authorization: Annotated[str | None, Header()] = None) -> None:
        if not secrets.compare_digest((authorization or "").encode(), f"Bearer {token}".encode()):
            raise HTTPException(status_code=401, detail="Invalid service credentials")

    @app.get("/health/live")
    def live() -> dict:
        return {"status": "ok", "provider": provider, "simulated": provider == "fake"}

    @app.get("/health/ready")
    async def ready() -> dict:
        service = runtime.get("service")
        if service and hasattr(service, "ready"):
            try:
                await service.ready()
            except Exception:
                raise HTTPException(status_code=503, detail="AI runtime unavailable") from None
        return {"status": "ready", "provider": provider, "simulated": provider == "fake"}

    @app.post("/internal/v1/preview", response_model=PreviewResult,
              dependencies=[Depends(authenticate)])
    async def preview(request: PreviewRequest) -> PreviewResult:
        # Simulation must not inherit hosted tracing from the developer's shell.
        with tracing_context(enabled=False):
            result = await preview_graph.ainvoke({"request": request}, {"recursion_limit": 8})
        return result["result"]

    @app.post("/internal/v1/consult", response_model=ConsultantResponse,
              dependencies=[Depends(authenticate)])
    async def consult(request: ConsultantRequest) -> ConsultantResponse:
        with tracing_context(enabled=False):
            result = await consultant_graph.ainvoke({"request": request}, {"recursion_limit": 10})
        return result["final_response"]

    return app
