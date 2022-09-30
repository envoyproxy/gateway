#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly KUSTOMIZE=${KUSTOMIZE:-tools/bin/kustomize}
readonly GATEWAY_API_VERSION="$1"
readonly TAG="$2"

mkdir -p release-artifacts/

# Wrap sed to deal with GNU and BSD sed flags.
run::sed() {
  if sed --version </dev/null 2>&1 | grep -q GNU; then
    # GNU sed
    sed -i "$@"
  else
    # assume BSD sed
    sed -i '' "$@"
  fi
}

# Download the supported Gateway API CRDs that will be supported by the release.
curl -sLo release-artifacts/gatewayapi-crds.yaml https://github.com/kubernetes-sigs/gateway-api/releases/download/"${GATEWAY_API_VERSION}"/experimental-install.yaml

echo "Added:" release-artifacts/gatewayapi-crds.yaml

# Generate the envoy gateway installation manifest supported by the release.
${KUSTOMIZE} build internal/provider/kubernetes/config/default > release-artifacts/envoy-gateway.yaml
${KUSTOMIZE} build internal/infrastructure/kubernetes/config/rbac > release-artifacts/infra-manager-rbac.yaml
${KUSTOMIZE} build > release-artifacts/install.yaml

echo "Generated:" release-artifacts/install.yaml

# Update the image in the Envoy Gateway deployment manifest.
[[ -n "${TAG}" ]] && run::sed \
  "-es|image: envoyproxy/gateway-dev:.*$|image: envoyproxy/gateway:${TAG}|" \
  "release-artifacts/install.yaml"

echo "Updated the envoy gateway image:" release-artifacts/install.yaml
