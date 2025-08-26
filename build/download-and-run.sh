#!/bin/sh

# pre-commit helper to run download latest release binary if missing before executing
# linting with it.

set -e

REPO=open-policy-agent/regal
BASE_URL=https://github.com/${REPO}

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
BIN_PATH="${SCRIPTPATH}/regal"

download()
{
    DETECTED_SYSTEM=$(uname -s)
    DETECTED_ARCHITECTURE=$(uname -m)

    REGAL_VERSION=${REGAL_VERSION:-latest}
    SYSTEM=${REGAL_SYSTEM:-${DETECTED_SYSTEM}}
    ARCHITECTURE=${REGAL_ARCHITECTURE:-${DETECTED_ARCHITECTURE}}

    echo "Downloading regal for ${SYSTEM} ${ARCHITECTURE}, ${REGAL_VERSION}…"
    BINARY_URL=${BASE_URL}/releases/${REGAL_VERSION}/download/regal_${SYSTEM}_${ARCHITECTURE}
    curl --fail -Lo "${BIN_PATH}" ${BINARY_URL}
    chmod +x "${BIN_PATH}"
}

if [ ! -x "${BIN_PATH}" ]; then download; fi
"${BIN_PATH}" $@
