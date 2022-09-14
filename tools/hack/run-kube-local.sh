#!/usr/bin/env bash

set -euo pipefail

GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)

# Check that kubectl is installed.
if ! [ "$(command -v kubectl)" ] ; then
    echo "kubectl not installed"
    exit 1
fi

# Run the envoy gateway binary
./bin/"${GOOS}"/"${GOARCH}"/envoy-gateway server "$@"
