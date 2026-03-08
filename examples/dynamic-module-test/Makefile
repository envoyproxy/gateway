
IMAGE_PREFIX ?= envoyproxy/gateway-
APP_NAME ?= dynamic-module-test
TAG ?= latest
ENVOY_VERSION ?= v1.37.0

.PHONY: docker-buildx
docker-buildx:
	docker buildx build . --build-arg ENVOY_VERSION=$(ENVOY_VERSION) -t $(IMAGE_PREFIX)$(APP_NAME):$(TAG) --load
