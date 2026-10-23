#!/bin/sh
# Check tracked Go sources without rewriting the working tree.
set -eu
cd "$(dirname "$0")/.."

format_output=$(mktemp)
trap 'rm -f "$format_output"' EXIT HUP INT TERM
git ls-files -z -- ':(glob)services/**/*.go' ':(glob)packages/**/*.go' ':(glob)test/**/*.go' |
    xargs -0 gofmt -l > "$format_output"

if [ -s "$format_output" ]; then
    echo 'Go files need gofmt:' >&2
    cat "$format_output" >&2
    exit 1
fi
