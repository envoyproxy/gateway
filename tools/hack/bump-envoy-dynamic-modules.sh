#!/usr/bin/env bash

# shellcheck disable=SC2038

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Read envoy version from DefaultEnvoyProxyImage in source code.
# On main this is "distroless-dev" (no version), on release branches it's "distroless-vX.Y.Z".
ENVOY_VERSION=$(grep 'DefaultEnvoyProxyImage' "${ROOT_DIR}/api/v1alpha1/shared_types.go" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' || true)
if [ -z "${ENVOY_VERSION}" ]; then
    echo "No envoy release version found in DefaultEnvoyProxyImage (dev image?) — skipping." >&2
    exit 0
fi

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
DYNAMIC_MODULE_DIRS=$(find "${ROOT_DIR}/examples" -name "go.mod" \
    -exec grep -l "envoy/source/extensions/dynamic_modules" {} \; \
    | xargs -I{} dirname {})

for dir in $DYNAMIC_MODULE_DIRS; do
    echo "Updating ${dir}..."

    # Update Dockerfile envoy image FROM line
    if [ -f "${dir}/Dockerfile" ]; then
        ENVOY_IMAGE="docker.io/envoyproxy/envoy:distroless-${ENVOY_VERSION}"
        ENVOY_DIGEST=$(docker buildx imagetools inspect "${ENVOY_IMAGE}" 2>/dev/null | grep -m1 'Digest:' | awk '{print $2}')
        if [ -z "${ENVOY_DIGEST}" ]; then
            echo "Error: Could not resolve digest for ${ENVOY_IMAGE}" >&2
            exit 1
        fi

        NEW_FROM="FROM ${ENVOY_IMAGE}@${ENVOY_DIGEST}"

        if [ "$GNU_SED" = "yes" ]; then
            sed -i'' "s|^FROM docker.io/envoyproxy/envoy:distroless.*|${NEW_FROM}|" "${dir}/Dockerfile"
        else
            sed -i '' "s|^FROM docker.io/envoyproxy/envoy:distroless.*|${NEW_FROM}|" "${dir}/Dockerfile"
        fi
    fi

    # Update Go SDK dependency
    pushd "${dir}" > /dev/null
    go get "github.com/envoyproxy/envoy/source/extensions/dynamic_modules@${COMMIT_SHA}"
    go mod tidy
    popd > /dev/null
done

echo "Done. Dynamic module dependencies updated to ${ENVOY_VERSION}."
