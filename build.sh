#!/usr/bin/env bash

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

build/fetch_data.sh "$SCRIPT_DIR/data/builtins/data.json"
