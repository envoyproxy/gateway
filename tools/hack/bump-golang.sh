#!/usr/bin/env bash

# shellcheck disable=SC2038

TO_GOLANG_VERSION=${1:-"1.25.3"}

echo "Bumping golang version from $FROM_GOLANG_VERSION to $TO_GOLANG_VERSION"

# Fetch the digest for the new golang image so Dockerfiles stay pinned.
echo "Fetching digest for golang:$TO_GOLANG_VERSION..."
GOLANG_DIGEST=$(docker pull "golang:$TO_GOLANG_VERSION" 2>&1 | grep "Digest:" | awk '{print $2}')
if [ -z "$GOLANG_DIGEST" ]; then
    echo "Failed to fetch digest for golang:$TO_GOLANG_VERSION" >&2
    exit 1
fi
echo "Digest: $GOLANG_DIGEST"

GOLANG_IMAGE="golang:$TO_GOLANG_VERSION@$GOLANG_DIGEST"

# detect gnu-sed or sed
GNU_SED=$(sed --version >/dev/null 2>&1 && echo "yes" || echo "no")
if [ "$GNU_SED" == "yes" ]; then
    find . -type f -name "Dockerfile" | xargs sed -i'' "s|^FROM golang:.*\$|FROM $GOLANG_IMAGE AS builder|g"
    find . -type f -name "go.mod" | xargs sed -i'' "s/^go.*\$/go $TO_GOLANG_VERSION/g"
else
    find . -type f -name "Dockerfile" | xargs sed -i '' "s|^FROM golang:.*\$|FROM $GOLANG_IMAGE AS builder|g"
    find . -type f -name "go.mod" | xargs sed -i '' "s/^go.*\$/go $TO_GOLANG_VERSION/g"
fi
