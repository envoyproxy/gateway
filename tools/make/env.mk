# This is a wrapper to hold common environment variables used in other make wrappers
#
# This file does not contain any specific make targets.

# Docker global variables

# REGISTRY is the image registry to use for build and push image targets.
REGISTRY ?= docker.io

# Docker Envoy Gateway variables

# REPOSITORY is the image repository
# Use envoyproxy/gateway-dev when developing
# Use envoyproxy/gateway when releasing an image.
REPOSITORY ?= envoyproxy/gateway
# REPOSITORY ?= envoyproxy/gateway-dev

# TAG is the image tag, defaults to current revision
TAG ?= $(REV)

# Docker Envoy Ratelimit variables

# RATELIMIT_REPOSITORY is the ratelimit repository
RATELIMIT_REPOSITORY ?= envoyproxy/ratelimit
# RATELIMIT_TAG is the ratelimit image tag
RATELIMIT_TAG ?= master

# Fuzzing variables

# FUZZ_TIME is the time to run the fuzzer for
FUZZ_TIME ?= 5s
