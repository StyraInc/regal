#!/usr/bin/env bash

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

result=$(opa eval --format pretty --data "$SCRIPT_DIR/builtin_metadata.rego" 'data.build.metadata.builtin_metadata')

echo "${result}" > "$1"
