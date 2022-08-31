#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly KUSTOMIZE=${KUSTOMIZE:-tools/bin/kustomize}
readonly GATEWAY_API_VERSION="$1"
readonly TAG="$2"

mkdir -p release/

# Wrap sed to deal with GNU and BSD sed flags.
run::sed() {
    local -r vers="$(sed --version < /dev/null 2>&1 | grep -q GNU && echo gnu || echo bsd)"
    case "$vers" in
        gnu) sed -i "$@" ;;
        *) sed -i '' "$@" ;;
    esac
}

# Download the supported Gateway API CRDs that will be supported by the release.
curl -sLo release/gatewayapi-crds.yaml https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/experimental-install.yaml

echo "Added:" release/gatewayapi-crds.yaml

# Generate the envoy gateway installation manifest supported by the release.
${KUSTOMIZE} build internal/provider/kubernetes/config/default > release/install.yaml

echo "Generated:" release/install.yaml

# Update the image in the Envoy Gateway deployment manifest.
run::sed \
  "-es|image: envoyproxy/gateway-dev:.*$|image: envoyproxy/gateway:${TAG}|" \
  "release/install.yaml"

echo "Updated the envoy gateway image:" release/install.yaml
