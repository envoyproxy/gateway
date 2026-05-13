
IMAGE_PREFIX ?= envoyproxy/gateway-
APP_NAME ?= sds-server
TAG ?= latest
GO_LDFLAGS ?=
PORT ?= 18001
NODE_ID ?= sds-test-node
COMMON_NAME ?= sds-test.example.com
DNS_NAMES ?= sds-test.example.com,*.example.com,localhost

.PHONY: build
build:
	go build -o $(APP_NAME) -ldflags "$(GO_LDFLAGS)" .

.PHONY: run
run: build
	./$(APP_NAME) -port $(PORT) -node $(NODE_ID) -cn $(COMMON_NAME) -dns $(DNS_NAMES)

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -f $(APP_NAME)

.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: docker-buildx
docker-buildx:
	docker buildx build . -t $(IMAGE_PREFIX)$(APP_NAME):$(TAG) --build-arg GO_LDFLAGS="$(GO_LDFLAGS)" --load

.PHONY: docker-run
docker-run:
	docker run -p $(PORT):$(PORT) $(IMAGE_PREFIX)$(APP_NAME):$(TAG) -port $(PORT) -node $(NODE_ID) -cn $(COMMON_NAME) -dns $(DNS_NAMES)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the SDS server binary"
	@echo "  run           - Build and run the SDS server"
	@echo "  test          - Run tests"
	@echo "  clean         - Remove built binaries"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  docker-buildx - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo ""
	@echo "Variables:"
	@echo "  PORT          - SDS server port (default: 18001)"
	@echo "  NODE_ID       - Envoy node ID (default: sds-test-node)"
	@echo "  COMMON_NAME   - Certificate Common Name (default: sds-test.example.com)"
	@echo "  DNS_NAMES     - Comma-separated DNS names (default: sds-test.example.com,*.example.com,localhost)"
	@echo "  TAG           - Docker image tag (default: latest)"
