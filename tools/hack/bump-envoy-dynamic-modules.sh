#!/usr/bin/env bash

# shellcheck disable=SC2038

set -euo pipefail

ENVOY_VERSION="${1:?Usage: $0 <envoy-version, e.g. v1.37.0>}"

# Extract major.minor from version (e.g., v1.37.0 -> 1.37)
MAJOR_MINOR=$(echo "${ENVOY_VERSION}" | sed 's/^v//' | cut -d. -f1,2)

# The dynamic modules Go SDK is a sub-module within the envoy monorepo.
# Go cannot resolve sub-modules using a commit that has a root-level semver tag
# (e.g., the exact v1.37.0 tag commit), so we use the HEAD of the release branch
# (e.g., release/v1.37) which is ABI-compatible and resolvable by Go tooling.
COMMIT_SHA=$(git ls-remote https://github.com/envoyproxy/envoy \
    "refs/heads/release/v${MAJOR_MINOR}" | awk '{print $1}')
if [ -z "$COMMIT_SHA" ]; then
    echo "Error: Could not find envoy release branch release/v${MAJOR_MINOR}" >&2
    exit 1
fi

echo "Resolved ${ENVOY_VERSION} (release/v${MAJOR_MINOR}) to commit ${COMMIT_SHA}"

# Detect GNU sed vs BSD sed (same pattern as bump-golang.sh)
GNU_SED=$(sed --version >/dev/null 2>&1 && echo "yes" || echo "no")

# Find all dynamic module example directories
DYNAMIC_MODULE_DIRS=$(find examples -name "go.mod" \
    -exec grep -l "envoy/source/extensions/dynamic_modules" {} \; \
    | xargs -I{} dirname {})

for dir in $DYNAMIC_MODULE_DIRS; do
    echo "Updating ${dir}..."

    # Update Dockerfile ARG ENVOY_VERSION
    if [ -f "${dir}/Dockerfile" ]; then
        if [ "$GNU_SED" = "yes" ]; then
            sed -i'' "s/^ARG ENVOY_VERSION=.*/ARG ENVOY_VERSION=${ENVOY_VERSION}/" "${dir}/Dockerfile"
        else
            sed -i '' "s/^ARG ENVOY_VERSION=.*/ARG ENVOY_VERSION=${ENVOY_VERSION}/" "${dir}/Dockerfile"
        fi
    fi

    # Update Makefile ENVOY_VERSION
    if [ -f "${dir}/Makefile" ]; then
        if [ "$GNU_SED" = "yes" ]; then
            sed -i'' "s/^ENVOY_VERSION ?=.*/ENVOY_VERSION ?= ${ENVOY_VERSION}/" "${dir}/Makefile"
        else
            sed -i '' "s/^ENVOY_VERSION ?=.*/ENVOY_VERSION ?= ${ENVOY_VERSION}/" "${dir}/Makefile"
        fi
    fi

    # Update Go SDK dependency
    pushd "${dir}" > /dev/null
    go get "github.com/envoyproxy/envoy/source/extensions/dynamic_modules@${COMMIT_SHA}"
    go mod tidy
    popd > /dev/null
done

echo "Done. Dynamic module dependencies updated to ${ENVOY_VERSION}."
