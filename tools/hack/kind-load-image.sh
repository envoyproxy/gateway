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
  if sed --version </dev/null 2>&1 | grep -q GNU; then
    # GNU sed
    sed -i "$@"
  else
    # assume BSD sed
    sed -i '' "$@"
  fi
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

files=(internal/provider/kubernetes/config/envoy-gateway/deploy_and_ns.yaml internal/provider/kubernetes/config/envoy-gateway/certgen-job.yaml)

# Update the image pull policy in the Envoy Gateway deployment manifest so
# the image is served by the kind cluster.
for file in "${files[@]}" ; do
  echo "setting \"imagePullPolicy: IfNotPresent\" for $file"
  run::sed \
    "-es|imagePullPolicy: Always|imagePullPolicy: IfNotPresent|" \
    "$file"
done

# Update the image in the Envoy Gateway deployment manifest.
for file in "${files[@]}" ; do
  echo "setting \"image: ${IMAGE}:${TAG}\" for $file"
  run::sed \
    "-es|image: envoyproxy/gateway-dev:.*$|image: ${IMAGE}:${TAG}|" \
    "$file"
done

# Push the Envoy Gateway image to the kind cluster.
echo "Loading image ${IMAGE}:${TAG} to kind cluster ${CLUSTER_NAME}..."
kind::cluster::load "${IMAGE}:${TAG}"
