#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/../..

venvPath="${SCRIPT_ROOT}/tools/bin/release-notes-docs.d/venv"

source ${venvPath}/bin/activate
python ${SCRIPT_ROOT}/tools/src/release-notes-docs/yml2md.py "$@"
deactivate
