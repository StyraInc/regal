#!/bin/sh

set -e
set -u
set -o pipefail

TEMP="$(mktemp -d)"
trap "rm -rf '$TEMP'" EXIT

cd "$(dirname "$0")"
dest="$(pwd)"

cd "$TEMP"

set -x

git clone https://github.com/StyraInc/enterprise-opa.git > "$TEMP/clone.log" 2>&1

cp -R "$TEMP/enterprise-opa/capabilities"/*.json "$dest"


