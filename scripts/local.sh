#!/bin/sh
# Reproducible local stack; never sources repository environment files.
set -eu
cd "$(dirname "$0")/.."
env_file=infrastructure/.env.local
# Every local secret the stack needs. Missing keys are appended, so older env files keep their values.
# LOCAL_CONNECTOR_CREDENTIAL_KEY is a 32-byte AES key as 64 hex characters; the rest are URL-safe tokens.
keys="DEMO_AI_INTERNAL_TOKEN LOCAL_JWT_ACCESS_SECRET LOCAL_JWT_REFRESH_SECRET LOCAL_CONNECTOR_MANAGEMENT_TOKEN LOCAL_CONNECTOR_CREDENTIAL_KEY LOCAL_META_APP_SECRET LOCAL_META_VERIFY_TOKEN"
missing=""
for key in $keys; do
    if [ ! -f "$env_file" ] || ! grep -q "^$key=" "$env_file"; then
        missing="$missing $key"
    fi
done
if [ -n "$missing" ]; then
    umask 077
    temp_file=$(mktemp infrastructure/.env.local.XXXXXX)
    trap 'rm -f "$temp_file"' EXIT HUP INT TERM
    if [ -f "$env_file" ]; then
        cat "$env_file" > "$temp_file"
        # Ensure the appended keys start on a new line.
        if [ -s "$temp_file" ] && [ "$(tail -c 1 "$temp_file")" != "" ]; then echo >> "$temp_file"; fi
    fi
    docker run --rm python:3.14-slim python -c 'import secrets, sys; print("\n".join(k+"="+(secrets.token_hex(32) if k=="LOCAL_CONNECTOR_CREDENTIAL_KEY" else secrets.token_urlsafe(32)) for k in sys.argv[1:]))' $missing >> "$temp_file"
    mv "$temp_file" "$env_file"
    trap - EXIT HUP INT TERM
fi
# The read-only connector reads its secrets from files; mirror them from the env file into a private,
# git-ignored directory. Files are world-readable so the container's non-root user can read the bind mount;
# the 0700 directory keeps them private on the host.
secrets_dir=infrastructure/.env.local.secrets
mkdir -p "$secrets_dir"
chmod 700 "$secrets_dir"
for pair in connector_management_token:LOCAL_CONNECTOR_MANAGEMENT_TOKEN connector_credential_key:LOCAL_CONNECTOR_CREDENTIAL_KEY meta_app_secret:LOCAL_META_APP_SECRET meta_verify_token:LOCAL_META_VERIFY_TOKEN; do
    name=${pair%%:*}
    key=${pair#*:}
    (umask 022; grep "^$key=" "$env_file" | head -n 1 | cut -d= -f2- | tr -d '\n' > "$secrets_dir/$name")
done
if [ "$#" -eq 0 ]; then
    set -- up --build -d --wait
fi
exec docker compose --env-file "$env_file" -f infrastructure/compose.demo.yml "$@"
