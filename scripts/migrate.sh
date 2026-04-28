#!/bin/sh

set -eu

TARGET="${MIGRATE_TARGET:-all}"

run_migration() {
	service_name="$1"
	case "$service_name" in
		user)
			migration_dir="migrations/user"
			database_url="postgres://postgres:postgres@postgres:5432/user?sslmode=disable"
			;;
		api)
			migration_dir="migrations/api"
			database_url="postgres://postgres:postgres@postgres:5432/api?sslmode=disable"
			;;
		catalog)
			migration_dir="migrations/catalog"
			database_url="postgres://postgres:postgres@postgres:5432/catalog?sslmode=disable"
			;;
		*)
			echo "Unknown migration target: $service_name" >&2
			exit 1
			;;
	esac

	echo "Applying migrations for $service_name"
	atlas migrate apply --dir "file://$migration_dir" --url "$database_url"
}

if [ "$TARGET" = "all" ]; then
	for service_name in user api catalog; do
		run_migration "$service_name"
	done
else
	run_migration "$TARGET"
fi
