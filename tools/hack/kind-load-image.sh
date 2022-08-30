#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly KIND=${KIND:-tools/bin/kind}
readonly CLUSTER_NAME=${CLUSTER_NAME:-"envoy-gateway"}

# Docker variables
readonly IMAGE="$1"
readonly TAG="$2"

# Wrap sed to deal with GNU and BSD sed flags.
run::sed() {
    local -r vers="$(sed --version < /dev/null 2>&1 | grep -q GNU && echo gnu || echo bsd)"
    case "$vers" in
        gnu) sed -i "$@" ;;
        *) sed -i '' "$@" ;;
    esac
}

kind::cluster::exists() {
    ${KIND} get clusters | grep -q "$1"
}

kind::cluster::load() {
    ${KIND} load docker-image \
        --name "${CLUSTER_NAME}" \
        "$@"
}

if ! kind::cluster::exists "$CLUSTER_NAME" ; then
    echo "cluster $CLUSTER_NAME does not exist"
    exit 2
fi

# Update the image pull policy in the Envoy Gateway deployment manifest so
# the image is served by the kind cluster.
echo "setting \"imagePullPolicy: IfNotPresent\" for Envoy Gateway deployment"
run::sed \
  "-es|imagePullPolicy: Always|imagePullPolicy: IfNotPresent|" \
  "./internal/provider/kubernetes/config/envoy-gateway/deploy_and_ns.yaml"

# Update the image in the Envoy Gateway deployment manifest.
echo "setting \"image: ${IMAGE}:${TAG}\" for Envoy Gateway deployment"
run::sed \
  "-es|image: envoyproxy/gateway-dev:.*$|image: ${IMAGE}:${TAG}|" \
  "./internal/provider/kubernetes/config/envoy-gateway/deploy_and_ns.yaml"

# Push the Envoy Gateway image to the kind cluster.
echo "Loading image ${IMAGE}:${TAG} to kind cluster ${CLUSTER_NAME}..."
kind::cluster::load "${IMAGE}:${TAG}"
