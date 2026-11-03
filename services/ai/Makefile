.PHONY: dev test smoke

dev:
	uv run uvicorn hexta_ai.app:create_app --factory --host 127.0.0.1 --port 58000

test:
	uv run pytest

smoke:
	uv run python smoke.py

