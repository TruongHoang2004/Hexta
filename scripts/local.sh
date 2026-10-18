#!/bin/sh
# Reproducible local stack; never sources repository environment files.
set -eu
cd "$(dirname "$0")/.."
env_file=infrastructure/.env.local
if [ ! -f "$env_file" ]; then
    umask 077
    temp_file=$(mktemp infrastructure/.env.local.XXXXXX)
    trap 'rm -f "$temp_file"' EXIT HUP INT TERM
    docker run --rm python:3.14-slim python -c 'import secrets; print("\n".join(k+"="+secrets.token_urlsafe(32) for k in ["DEMO_AI_INTERNAL_TOKEN","LOCAL_JWT_ACCESS_SECRET","LOCAL_JWT_REFRESH_SECRET"]))' > "$temp_file"
    mv "$temp_file" "$env_file"
    trap - EXIT HUP INT TERM
fi
if [ "$#" -eq 0 ]; then
    set -- up --build -d --wait
fi
exec docker compose --env-file "$env_file" -f infrastructure/compose.demo.yml "$@"
