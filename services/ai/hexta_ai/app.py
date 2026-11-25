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
from hexta_ai.knowledge.service import knowledge_from_env
from hexta_ai.providers import BaseAIProvider, ProviderError, provider_from_env
from hexta_ai.routing import router_from_env
from hexta_ai.schemas.consultant import ConsultantRequest, ConsultantResponse
from hexta_ai.graph_consultant import consultant_graph


_FROM_ENV = object()


class RuntimeProfile:
    """Identity checked by durable recovery: paid when the answer model, Jev or embeddings are real. An offline
    router (AI_FAKE_ROUTER) is simulated and keeps the profile simulated."""

    def __init__(self, provider, router, knowledge):
        self.name = provider.name
        self.model = provider.model
        self.simulated = (provider.simulated and (router is None or getattr(router, "simulated", False))
                          and (knowledge is None or knowledge.simulated))


def create_app(internal_token: str | None = None, provider: BaseAIProvider | None = None, *,
               router=_FROM_ENV, knowledge=_FROM_ENV) -> FastAPI:
    token = internal_token if internal_token is not None else os.getenv("AI_INTERNAL_TOKEN", "")
    if len(token) < 24:
        raise RuntimeError("AI_INTERNAL_TOKEN must contain at least 24 characters")
    selected = provider if provider is not None else provider_from_env()
    router = router_from_env() if router is _FROM_ENV else router
    knowledge = knowledge_from_env() if knowledge is _FROM_ENV else knowledge
    profile = RuntimeProfile(selected, router, knowledge)
    graph = build_graph(selected, router=router, knowledge=knowledge, simulated=profile.simulated)
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
                    token, os.getenv("AI_GRPC_ADDRESS", "0.0.0.0:50051"), runner=run, provider=profile,
                    knowledge=knowledge,
                )
            runtime["service"] = service
            yield
        finally:
            if server:
                await server.stop(grace=2)
            if service:
                await service.close()
            await selected.close()
            if router is not None:
                await router.close()
            if knowledge is not None:
                await knowledge.close()

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
        return {"status": "ok", "provider": selected.name, "simulated": profile.simulated}

    @app.get("/health/ready")
    async def ready() -> dict:
        service = runtime.get("service")
        if service and hasattr(service, "ready"):
            try:
                await service.ready()
            except Exception:
                raise HTTPException(status_code=503, detail="AI runtime unavailable") from None
        return {"status": "ready", "provider": selected.name, "simulated": profile.simulated}

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
