
IMAGE_PREFIX ?= envoyproxy/gateway-
APP_NAME ?= grpc-ext-auth
TAG ?= latest

.PHONY: docker-buildx
docker-buildx:
	docker buildx build . -t $(IMAGE_PREFIX)$(APP_NAME):$(TAG) --build-arg GO_LDFLAGS="$(GO_LDFLAGS)" --load
