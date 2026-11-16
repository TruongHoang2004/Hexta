#!/bin/sh
# Production Compose wrapper for the single-host deployment (see infrastructure/README.deploy.md).
# Loads only infrastructure/prod/.env; never prints resolved configuration.
set -eu
cd "$(dirname "$0")/.."
env_file=infrastructure/prod/.env
if [ ! -f "$env_file" ]; then
    echo "Missing $env_file; copy infrastructure/prod/.env.example and fill it in." >&2
    exit 1
fi
for s in connector_management_token connector_credential_key meta_app_secret meta_verify_token; do
    [ -s "infrastructure/prod/secrets/$s" ] || { echo "Missing infrastructure/prod/secrets/$s" >&2; exit 1; }
done
exec docker compose --env-file "$env_file" -f infrastructure/compose.prod.yml "$@"
