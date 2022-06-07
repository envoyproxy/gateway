# Golang variables
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Docker variables
# REGISTRY is the image registry to use for build and push image targets.
REGISTRY ?= docker.io/envoyproxy
# IMAGE is the image URL for build and push image targets.
IMAGE ?= ${REGISTRY}/gateway-dev
# REV is the short git sha of latest commit.
REV=$(shell git rev-parse --short HEAD)
# Tag is the tag to use for build and push image targets.
TAG ?= $(REV)

.PHONY: build
build:  ## Build the envoy-gateway binary
	@CGO_ENABLED=0 go build -a -o ./bin/${GOOS}/${GOARCH}/ github.com/envoyproxy/gateway/cmd/envoy-gateway

build-linux-amd64:
	@GOOS=linux GOARCH=amd64 $(MAKE) build

build-linux-arm64:
	@GOOS=linux GOARCH=arm64 $(MAKE) build

build-all: build-linux-amd64 build-linux-arm64

.PHONY: test
test:
	@go test ./...

.PHONY: docker-build
docker-build: build-all ## Build the envoy-gateway docker image.
	@DOCKER_BUILDKIT=1 docker build -t $(IMAGE):$(TAG) -f Dockerfile bin 

.PHONY: docker-push
docker-push: ## Push the docker image for envoy-gateway.
	docker push $(IMAGE):$(TAG)

