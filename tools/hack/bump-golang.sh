#!/usr/bin/env bash

# shellcheck disable=SC2038

FROM_GOLANG_VERSION=${1:-"1.23.6"}
TO_GOLANG_VERSION=${2:-"1.24.0"}

echo "Bumping golang version from $FROM_GOLANG_VERSION to $TO_GOLANG_VERSION"

# detect gnu-sed or sed
GNU_SED=$(sed --version >/dev/null 2>&1 && echo "yes" || echo "no")
if [ "$GNU_SED" == "yes" ]; then
    find . -type f -name "Dockerfile" | xargs sed -i'' "s/FROM golang:$FROM_GOLANG_VERSION AS builder/FROM golang:$TO_GOLANG_VERSION AS builder/g"
    find . -type f -name "go.mod" | xargs sed -i'' "s/go $FROM_GOLANG_VERSION/go $TO_GOLANG_VERSION/g"
else
    find . -type f -name "Dockerfile" | xargs sed -i '' "s/FROM golang:$FROM_GOLANG_VERSION AS builder/FROM golang:$TO_GOLANG_VERSION AS builder/g"
    find . -type f -name "go.mod" | xargs sed -i '' "s/go $FROM_GOLANG_VERSION/go $TO_GOLANG_VERSION/g"
fi
