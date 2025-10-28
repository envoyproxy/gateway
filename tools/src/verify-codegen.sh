#!/usr/bin/env bash

# Copyright Envoy Gateway Authors
# SPDX-License-Identifier: Apache-2.0
# The full text of the Apache license is available in the LICENSE file at
# the root of the repo.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
cd "${SCRIPT_ROOT}"

echo "Verifying generated code is up to date..."

# Save current state
TEMP_DIR=$(mktemp -d)
trap "rm -rf ${TEMP_DIR}" EXIT

if [ -d "${SCRIPT_ROOT}/pkg/client" ]; then
  cp -r "${SCRIPT_ROOT}/pkg/client" "${TEMP_DIR}/"
fi

# Regenerate
"${SCRIPT_ROOT}/tools/src/update-codegen.sh"

# Compare
if [ -d "${TEMP_DIR}/client" ]; then
  if ! diff -Naupr "${TEMP_DIR}/client" "${SCRIPT_ROOT}/pkg/client" > /dev/null 2>&1; then
    echo "ERROR: Generated client code is out of date."
    echo "Please run 'make kube-generate' and commit the changes."
    echo ""
    echo "Differences found:"
    diff -Naupr "${TEMP_DIR}/client" "${SCRIPT_ROOT}/pkg/client" || true
    exit 1
  fi
fi

echo "Generated code is up to date!"

