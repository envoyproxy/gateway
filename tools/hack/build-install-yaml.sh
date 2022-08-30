#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly KUSTOMIZE=${KUSTOMIZE:-tools/bin/kustomize}
readonly GATEWAY_API_VERSION="$1"

mkdir -p release/

# Download the supported Gateway API CRDs that will be supported by the release.
curl -sLo release/gatewayapi-crds.yaml https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/experimental-install.yaml

echo "Added:" release/gatewayapi-crds.yaml

# Generate the envoy gateway installation manifest supported by the release.
${KUSTOMIZE} build internal/provider/kubernetes/config/default > release/install.yaml

echo "Generated:" release/install.yaml
