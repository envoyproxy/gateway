#!/usr/bin/env bash

# Copyright Envoy Gateway Authors
# SPDX-License-Identifier: Apache-2.0
# The full text of the Apache license is available in the LICENSE file at
# the root of the repo.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
RELEASE_NOTES_DIR="${ROOT_DIR}/release-notes"

# Check if release-notes directory exists
if [[ ! -d "${RELEASE_NOTES_DIR}" ]]; then
    echo "Error: release-notes directory not found at ${RELEASE_NOTES_DIR}"
    exit 1
fi

# Valid patterns:
# - current.yaml
# - vX.Y.Z.yaml (e.g., v1.2.3.yaml)
# - vX.Y.Z-rc.M.yaml (e.g., v1.0.0-rc.1.yaml)

VALID_PATTERN='^(current|v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?)\.yaml$'

# Legacy files to ignore (grandfathered in)
LEGACY_FILES=(
    "v0.2.0-rc1.yaml"
    "v0.2.0-rc2.yaml"
)

INVALID_FILES=()

# Check all .yaml files in release-notes directory
while IFS= read -r -d '' file; do
    filename="$(basename "${file}")"
    
    # Skip if it's a legacy file
    is_legacy=false
    for legacy in "${LEGACY_FILES[@]}"; do
        if [[ "${filename}" == "${legacy}" ]]; then
            is_legacy=true
            break
        fi
    done
    
    if [[ "${is_legacy}" == "true" ]]; then
        continue
    fi
    
    if [[ ! "${filename}" =~ ${VALID_PATTERN} ]]; then
        INVALID_FILES+=("${filename}")
    fi
done < <(find "${RELEASE_NOTES_DIR}" -maxdepth 1 -type f -name "*.yaml" -print0)

# Report results
if [[ ${#INVALID_FILES[@]} -gt 0 ]]; then
    echo "❌ Invalid release notes filenames found:"
    for file in "${INVALID_FILES[@]}"; do
        echo "  - ${file}"
    done
    echo ""
    echo "Valid formats:"
    echo "  - current.yaml"
    echo "  - vX.Y.Z.yaml (e.g., v1.2.3.yaml)"
    echo "  - vX.Y.Z-rc.M.yaml (e.g., v1.0.0-rc.1.yaml)"
    exit 1
else
    echo "✅ All release notes filenames are valid"
fi
