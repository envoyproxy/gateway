
IMAGE_PREFIX ?= envoyproxy/gateway-
APP_NAME ?= wasm-module-test
TAG ?= latest
ENVOY_VERSION ?= dev

.PHONY: docker-buildx
docker-buildx:
	docker buildx build . --build-arg ENVOY_VERSION=$(ENVOY_VERSION) -t $(IMAGE_PREFIX)$(APP_NAME):$(TAG) --load

.PHONY: backend-callout
backend-callout:
	mkdir -p ../../bin/wasm-module-test
	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags="-s -w" \
		-o ../../bin/wasm-module-test/backend_callout.wasm .
