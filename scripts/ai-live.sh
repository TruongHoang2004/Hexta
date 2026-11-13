#!/bin/sh
# Start real AI on an already provisioned local stack; never applies migrations.
set -eu
cd "$(dirname "$0")/.."
if [ ! -f infrastructure/.env.local ] || [ ! -f services/ai/.env ]; then
  echo "Private local stack and AI environment files are required." >&2
  exit 1
fi
if [ "$#" -eq 0 ]; then
  set -- up --build -d --no-deps ai api
fi
exec docker compose --env-file infrastructure/.env.local --env-file services/ai/.env \
  -f infrastructure/compose.demo.yml -f infrastructure/compose.ai-live.yml "$@"
