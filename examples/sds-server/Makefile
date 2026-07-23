
IMAGE_PREFIX ?= envoyproxy/gateway-
APP_NAME ?= sds-server
TAG ?= latest
GO_LDFLAGS ?=
SOCKET ?= /tmp/sds.sock

.PHONY: build
build:
	go build -o $(APP_NAME) -ldflags "$(GO_LDFLAGS)" .

.PHONY: run
run: build
	./$(APP_NAME) -socket $(SOCKET)

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
	docker run -v /tmp:/tmp $(IMAGE_PREFIX)$(APP_NAME):$(TAG) -socket /tmp/sds.sock

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
	@echo "  SOCKET        - Unix domain socket path (default: /tmp/sds.sock)"
	@echo "  TAG           - Docker image tag (default: latest)"
