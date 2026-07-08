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

# Valid patterns for versioned release notes files:
# - vX.Y.Z.yaml (e.g., v1.2.3.yaml)
# - vX.Y.Z-rc.M.yaml (e.g., v1.0.0-rc.1.yaml)
#
# In-development release notes now live as per-change fragment files under
# release-notes/current/<section>/<pr-number>-<slug>.md (see release-notes/current/README.md).

VALID_PATTERN='^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?\.yaml$'

# Canonical section directories under release-notes/current/.
VALID_SECTIONS=(
    "breaking_changes"
    "security_updates"
    "new_features"
    "bug_fixes"
    "performance_improvements"
    "deprecations"
    "other_changes"
)

# Fragment filename convention: <pr-number>-<slug>.md
FRAGMENT_PATTERN='^[0-9]+-[a-z0-9-]+\.md$'

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

# Validate the in-development fragment directory: release-notes/current/
CURRENT_DIR="${RELEASE_NOTES_DIR}/current"
if [[ -d "${CURRENT_DIR}" ]]; then
    while IFS= read -r -d '' entry; do
        rel="${entry#"${CURRENT_DIR}/"}"
        base="$(basename "${entry}")"

        # Allow top-level helper files.
        if [[ "${rel}" == "README.md" ]]; then
            continue
        fi

        # Every fragment must live directly inside a known section directory.
        section="${rel%%/*}"
        is_valid_section=false
        for valid in "${VALID_SECTIONS[@]}"; do
            if [[ "${section}" == "${valid}" ]]; then
                is_valid_section=true
                break
            fi
        done
        if [[ "${is_valid_section}" == "false" ]]; then
            INVALID_FILES+=("current/${rel} (not in a known section directory)")
            continue
        fi

        # Allow the per-section placeholder.
        if [[ "${base}" == ".gitkeep" ]]; then
            continue
        fi

        # Reject nested paths (fragments must be flat inside a section dir).
        if [[ "${rel}" != "${section}/${base}" ]]; then
            INVALID_FILES+=("current/${rel} (fragments must not be nested)")
            continue
        fi

        if [[ ! "${base}" =~ ${FRAGMENT_PATTERN} ]]; then
            INVALID_FILES+=("current/${rel} (must match <pr-number>-<slug>.md)")
        fi
    done < <(find "${CURRENT_DIR}" -mindepth 1 -type f -print0)
fi

# Report results
if [[ ${#INVALID_FILES[@]} -gt 0 ]]; then
    echo "❌ Invalid release notes filenames found:"
    for file in "${INVALID_FILES[@]}"; do
        echo "  - ${file}"
    done
    echo ""
    echo "Valid formats:"
    echo "  - Released versions: vX.Y.Z.yaml / vX.Y.Z-rc.M.yaml (e.g., v1.2.3.yaml, v1.0.0-rc.1.yaml)"
    echo "  - Unreleased fragments: current/<section>/<pr-number>-<slug>.md"
    echo "    sections: breaking_changes, security_updates, new_features, bug_fixes,"
    echo "              performance_improvements, deprecations, other_changes"
    exit 1
else
    echo "✅ All release notes filenames are valid"
fi
