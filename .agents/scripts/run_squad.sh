#!/usr/bin/env bash
# run_squad.sh - Executable wrapper for run_squad.py

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$REPO_ROOT"
exec python3 "$SCRIPT_DIR/run_squad.py" "$@"
