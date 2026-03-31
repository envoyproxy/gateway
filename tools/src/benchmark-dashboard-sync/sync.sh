#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
TOOLS_DIR=$(cd -- "${SCRIPT_DIR}/../.." && pwd)
REPO_ROOT=$(cd -- "${TOOLS_DIR}/.." && pwd)

cd "${TOOLS_DIR}"
go run ./src/benchmark-dashboard-sync "$@"

cd "${REPO_ROOT}/site/tools/benchmark-dashboard"
npm run build
