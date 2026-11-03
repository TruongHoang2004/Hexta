#!/bin/sh
# Explicit migrations only; no schema mutation from application processes.
set -eu
cd "${MIGRATION_ROOT:-$(dirname "$0")/..}"
target=${MIGRATE_TARGET:-all}
action=${1:-apply}
case "$target" in api|ai|connector|all) ;; *) echo "Unknown migration target: $target (use api, ai, connector or all)" >&2; exit 1;; esac
case "$action" in apply|status|validate|hash) ;; *) echo "Unknown migration action: $action" >&2; exit 1;; esac

# Validate inputs for every selected service before applying any migration.
for service_name in api ai connector; do
    [ "$target" = all ] || [ "$target" = "$service_name" ] || continue
    if [ "$target" = all ] && [ "$service_name" = connector ] && { [ "$action" = apply ] || [ "$action" = status ]; }; then continue; fi
    [ -d "migrations/$service_name" ] || { echo "Missing migration directory: $service_name" >&2; exit 1; }
    if [ "$action" = apply ] || [ "$action" = status ]; then
        case "$service_name" in api) test -n "${API_DATABASE_URL:-}" || { echo 'API_DATABASE_URL is required' >&2; exit 1; };; connector) test -n "${CONNECTOR_DATABASE_URL:-}" || { echo 'CONNECTOR_DATABASE_URL is required' >&2; exit 1; };; ai) test -n "${AI_DATABASE_URL:-}" || { echo 'AI_DATABASE_URL is required' >&2; exit 1; };; esac
    fi
done
for service_name in api ai connector; do
    [ "$target" = all ] || [ "$target" = "$service_name" ] || continue
    if [ "$target" = all ] && [ "$service_name" = connector ] && { [ "$action" = apply ] || [ "$action" = status ]; }; then continue; fi
    echo "Atlas $action: $service_name"
    if [ "$action" = validate ] || [ "$action" = hash ]; then
        atlas migrate "$action" --dir "file://migrations/$service_name"
    elif [ "$action" = apply ] && [ "$service_name" = ai ]; then
        # The initial migration validates and adopts the exact pre-Atlas AI table.
        # API history is never baselined or allowed dirty.
        atlas migrate apply --env ai --allow-dirty
    else
        atlas migrate "$action" --env "$service_name"
    fi
done
