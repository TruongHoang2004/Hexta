FROM python:3.14-slim AS runtime

WORKDIR /app
COPY pyproject.toml uv.lock ./
RUN pip install --no-cache-dir uv==0.11.11 \
    && uv sync --frozen --no-dev --no-install-project
COPY hexta_ai ./hexta_ai
COPY generated ./generated
RUN useradd --uid 10001 --create-home app
USER app
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 LANGSMITH_TRACING=false PYTHONPATH=/app/generated
EXPOSE 8000 50051
CMD ["/app/.venv/bin/uvicorn", "hexta_ai.app:create_app", "--factory", "--host", "0.0.0.0", "--port", "8000"]


FROM runtime AS test
USER root
COPY --from=arigaio/atlas:latest /atlas /usr/local/bin/atlas
RUN uv sync --frozen --no-install-project
COPY tests ./tests
USER app
CMD ["/app/.venv/bin/python", "-m", "pytest", "-q", "-p", "no:cacheprovider"]

FROM runtime AS production
