"""Internal candidate preview and durable gRPC runtime. No customer channel access."""

import os
import secrets
from contextlib import asynccontextmanager
from typing import Annotated

from fastapi import Depends, FastAPI, Header, HTTPException
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from langsmith import tracing_context

from hexta_ai.rpc import start_server
from hexta_ai.graph import PreviewRequest, PreviewResult, build_graph
from hexta_ai.providers import BaseAIProvider, ProviderError, provider_from_env
from hexta_ai.schemas.consultant import ConsultantRequest, ConsultantResponse
from hexta_ai.graph_consultant import consultant_graph


def create_app(internal_token: str | None = None, provider: BaseAIProvider | None = None) -> FastAPI:
    token = internal_token if internal_token is not None else os.getenv("AI_INTERNAL_TOKEN", "")
    if len(token) < 24:
        raise RuntimeError("AI_INTERNAL_TOKEN must contain at least 24 characters")
    selected = provider if provider is not None else provider_from_env()
    graph = build_graph(selected)
    runtime = {}

    async def run(request: PreviewRequest) -> PreviewResult:
        # Never inherit hosted tracing settings for prompts, PII or credentials.
        with tracing_context(enabled=False):
            result = await graph.ainvoke({"request": request}, {"recursion_limit": 12})
        return result["result"]

    @asynccontextmanager
    async def lifespan(app):
        server = service = None
        try:
            if os.getenv("AI_GRPC_ENABLED", "true").lower() == "true":
                server, service = await start_server(
                    token, os.getenv("AI_GRPC_ADDRESS", "0.0.0.0:50051"), runner=run, provider=selected,
                )
            runtime["service"] = service
            yield
        finally:
            if server:
                await server.stop(grace=2)
            if service:
                await service.close()
            await selected.close()

    app = FastAPI(lifespan=lifespan, title="Hexta AI Service", version="0.2.0",
                  docs_url=None, redoc_url=None)

    @app.exception_handler(RequestValidationError)
    async def invalid_input(request, error):
        return JSONResponse(status_code=422, content={"detail": "Invalid preview input"})

    def authenticate(authorization: Annotated[str | None, Header()] = None) -> None:
        if not secrets.compare_digest((authorization or "").encode(), f"Bearer {token}".encode()):
            raise HTTPException(status_code=401, detail="Invalid service credentials")

    @app.get("/health/live")
    def live() -> dict:
        return {"status": "ok", "provider": selected.name, "simulated": selected.simulated}

    @app.get("/health/ready")
    async def ready() -> dict:
        service = runtime.get("service")
        if service and hasattr(service, "ready"):
            try:
                await service.ready()
            except Exception:
                raise HTTPException(status_code=503, detail="AI runtime unavailable") from None
        return {"status": "ready", "provider": selected.name, "simulated": selected.simulated}

    @app.post("/internal/v1/preview", response_model=PreviewResult,
              dependencies=[Depends(authenticate)])
    async def preview(request: PreviewRequest) -> PreviewResult:
        try:
            return await run(request)
        except ProviderError as error:
            status = 429 if error.code == "QUOTA_EXCEEDED" else 422 if error.code == "INPUT_BUDGET_EXCEEDED" else 503
            raise HTTPException(status_code=status, detail=error.code) from None
        except Exception:
            raise HTTPException(status_code=503, detail="LLM_PROVIDER_ERROR") from None

    @app.post("/internal/v1/consult", response_model=ConsultantResponse,
              dependencies=[Depends(authenticate)])
    async def consult(request: ConsultantRequest) -> ConsultantResponse:
        with tracing_context(enabled=False):
            result = await consultant_graph.ainvoke({"request": request}, {"recursion_limit": 10})
        return result["final_response"]

    return app
