# This is a wrapper to hold common environment variables used in other make wrappers
#
# This file does not contain any specific make targets.


# Docker variables

# REGISTRY is the image registry to use for build and push image targets.
REGISTRY ?= docker.io/envoyproxy
# IMAGE_NAME is the name of EG image
# Use gateway-dev in default when developing
# Use gateway when releasing an image.
IMAGE_NAME ?= gateway-dev
# IMAGE is the image URL for build and push image targets.
IMAGE ?= ${REGISTRY}/${IMAGE_NAME}
# Tag is the tag to use for build and push image targets.
TAG ?= $(REV)
# ENVOY_PROXY_IMAGE is the default Envoy proxy image rendered by the gateway-helm chart.
# It is sourced from DefaultEnvoyProxyImage so the chart and the controller never drift apart.
ENVOY_PROXY_IMAGE ?= $(shell grep -m1 'DefaultEnvoyProxyImage *=' api/v1alpha1/shared_types.go | cut -d '"' -f2)

# Fuzzing variables

# FUZZ_TIME is the time to run the fuzzer for
FUZZ_TIME ?= 5s