
IMAGE_PREFIX ?= envoyproxy/gateway-
APP_NAME ?= dynamic-module-test
TAG ?= latest

.PHONY: docker-buildx
docker-buildx:
	docker buildx build . -t $(IMAGE_PREFIX)$(APP_NAME):$(TAG) --load
