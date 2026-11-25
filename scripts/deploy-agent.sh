#!/bin/bash
# Pull-based deploy agent for the single-host production stack (run by the hexta-deploy systemd timer).
#
# Desired state lives in git: infrastructure/prod/release.env on origin/main names an image tag sha-<commit>.
# When it differs from the deployed tag, the agent checks out <commit> (so configuration matches the images),
# verifies all images exist, backs up the database, pulls and starts the stack, and checks health through Caddy.
# On failure it restores the previous commit and tag and does not retry the same tag until release.env changes.
#
# Everything runs inside main() so a checkout that replaces this file cannot affect the running copy.
set -euo pipefail

SERVICES="api connector ai web migrator"
# Compose volume of the production database and the glibc marker written by README.deploy.md section 10a.
PG_VOLUME="hexta-prod_postgres_data"
PG_MARKER="HEXTA_GLIBC_COLLATION_OK"

main() {
    cd "$(dirname "$0")/.."
    local root prod state
    root=$(pwd)
    prod="$root/infrastructure/prod"
    state="$prod/state"
    mkdir -p "$state" "$prod/backups"

    exec 9>"$state/agent.lock"
    if ! flock -n 9; then
        log "another deploy is running"
        return 0
    fi

    local registry desired current failed
    registry=$(env_value HEXTA_IMAGE_REGISTRY)
    registry=${registry:-ghcr.io/truonghoang2004/hexta}

    git fetch --quiet origin main
    desired=$(git show origin/main:infrastructure/prod/release.env | sed -n 's/^HEXTA_IMAGE_TAG=//p' | tail -1)
    current=$(cat "$state/deployed" 2>/dev/null || true)
    failed=$(cat "$state/failed" 2>/dev/null || true)

    if [ -z "$desired" ] || [ "$desired" = "$current" ]; then
        return 0
    fi
    if [ "$desired" = "$failed" ]; then
        return 0
    fi
    case "$desired" in
        sha-[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]) ;;
        *) fail_release "$desired" "release tag is not sha-<7 hex>"; return 1 ;;
    esac

    local commit
    commit=$(git rev-parse --verify --quiet "${desired#sha-}^{commit}") || {
        fail_release "$desired" "commit not found on origin"
        return 1
    }
    for svc in $SERVICES; do
        if ! docker manifest inspect "$registry/$svc:$desired" >/dev/null 2>&1; then
            # CI may still be publishing; retry on the next run without marking the release failed.
            log "image $svc:$desired not available yet"
            return 0
        fi
    done

    if musl_downgrade "$commit"; then
        fail_release "$desired" "its Alpine (musl) Postgres image must not start on the glibc data volume; roll back with README.deploy.md section 10a"
        return 1
    fi
    if ! backup "$desired"; then
        fail_release "$desired" "database backup failed; nothing was changed"
        return 1
    fi
    log "deploying $desired (from ${current:-nothing})"
    notify "Hexta: deploying $desired (from ${current:-nothing})"

    local previous_commit
    previous_commit=$(git rev-parse HEAD)
    git checkout --quiet --detach "$commit"
    if deploy "$desired"; then
        echo "$desired" >"$state/deployed"
        rm -f "$state/failed"
        log "deployed $desired"
        notify "Hexta: $desired is live"
        return 0
    fi

    echo "$desired" >"$state/failed"
    if musl_downgrade "$previous_commit"; then
        # The failed release already started Postgres on glibc; the previous (musl) image would corrupt indexes.
        log "deploy of $desired failed; automatic rollback to $current refused (musl Postgres on a glibc volume)"
        notify "Hexta: deploy of $desired FAILED; rollback to $current refused (Postgres collation), see README.deploy.md 10a"
        return 1
    fi
    git checkout --quiet --detach "$previous_commit"
    if [ -n "$current" ] && deploy "$current"; then
        log "deploy of $desired failed; rolled back to $current"
        notify "Hexta: deploy of $desired FAILED; rolled back to $current"
    else
        log "deploy of $desired failed; rollback unavailable or failed"
        notify "Hexta: deploy of $desired FAILED and rollback did not complete; check the server"
    fi
    return 1
}

# deploy TAG: pull and start the stack for TAG, then check the public health endpoint through Caddy.
deploy() {
    local tag=$1 domain
    domain=$(env_value DOMAIN)
    HEXTA_IMAGE_TAG=$tag ./scripts/prod.sh pull --quiet || return 1
    HEXTA_IMAGE_TAG=$tag ./scripts/prod.sh up -d --wait --remove-orphans || return 1
    for _ in 1 2 3 4 5 6; do
        if curl -fsS -o /dev/null --max-time 5 --resolve "$domain:443:127.0.0.1" "https://$domain/api/v1/health"; then
            return 0
        fi
        sleep 5
    done
    return 1
}

# musl_downgrade COMMIT: true when COMMIT predates the Compose collation guard (its Postgres is the Alpine/musl image)
# while the data volume already carries the glibc marker. Starting it would corrupt text indexes; the only way back
# is the volume copy or dump restore in README.deploy.md section 10a. Compose in such an old commit cannot check this.
musl_downgrade() {
    if git show "$1:infrastructure/compose.prod.yml" 2>/dev/null | grep -q '^  postgres-collation-guard:'; then
        return 1
    fi
    docker volume inspect "$PG_VOLUME" >/dev/null 2>&1 || return 1
    docker run --rm --network none --user root -v "$PG_VOLUME:/data:ro" --entrypoint test \
        pgvector/pgvector:0.8.1-pg17-bookworm -f "/data/$PG_MARKER"
}

backup() {
    local file
    file="infrastructure/prod/backups/pre-$1-$(date +%Y%m%d-%H%M%S).sql.gz"
    # First deploy: no database yet, nothing to back up.
    if ! ./scripts/prod.sh ps --status running --services 2>/dev/null | grep -qx postgres; then
        return 0
    fi
    ./scripts/prod.sh exec -T postgres pg_dumpall -U postgres | gzip >"$file" || { rm -f "$file"; return 1; }
    find infrastructure/prod/backups -name 'pre-*.sql.gz' -mtime +14 -delete
    log "backup written: $file"
}

fail_release() {
    echo "$1" >"infrastructure/prod/state/failed"
    log "release $1 rejected: $2"
    notify "Hexta: release $1 rejected: $2"
}

# env_value KEY: one value from infrastructure/prod/.env without sourcing the file.
env_value() {
    sed -n "s/^$1=//p" infrastructure/prod/.env 2>/dev/null | tail -1
}

log() {
    echo "$(date -u +%FT%TZ) deploy-agent: $*"
}

# Optional Telegram notification; DEPLOY_NOTIFY_TELEGRAM_* live only in prod/.env and are never printed.
notify() {
    local token chat
    token=$(env_value DEPLOY_NOTIFY_TELEGRAM_TOKEN)
    chat=$(env_value DEPLOY_NOTIFY_TELEGRAM_CHAT_ID)
    if [ -n "$token" ] && [ -n "$chat" ]; then
        curl -fsS -o /dev/null --max-time 10 "https://api.telegram.org/bot$token/sendMessage" \
            --data-urlencode "chat_id=$chat" --data-urlencode "text=$*" >/dev/null 2>&1 || true
    fi
}

main "$@"
exit $?
