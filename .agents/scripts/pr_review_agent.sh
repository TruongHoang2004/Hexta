#!/usr/bin/env bash
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export GH_REPO="${GH_REPO:-TruongHoang2004/Hexta}"
if [ -n "${HTTPS_PROXY:-}" ] || [ -n "${https_proxy:-}" ]; then
  export NO_PROXY="*"
  export no_proxy="*"
  unset HTTPS_PROXY HTTP_PROXY https_proxy http_proxy || true
fi
exec python3 "$DIR/pr_review_agent.py" "$@"
