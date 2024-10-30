
IMAGE_PREFIX ?= envoyproxy/gateway-
APP_NAME ?= preserve-case-backend
TAG ?= latest

.PHONY: docker-builx
docker-buildx:
	docker buildx build . -t $(IMAGE_PREFIX)$(APP_NAME):$(TAG) --build-arg GO_LDFLAGS="$(GO_LDFLAGS)" --load
