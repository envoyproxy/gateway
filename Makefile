# REGISTRY is the image registry to use for build and push image targets.
REGISTRY ?= docker.io/envoyproxy
# IMAGE is the image URL for build and push image targets.
IMAGE ?= ${REGISTRY}/gateway-dev
# REV is the short git sha of latest commit.
REV=$(shell git rev-parse --short HEAD)
# Tag is the tag to use for build and push image targets.
TAG ?= $(REV)

.PHONY: build
build:
	@CGO_ENABLED=0 go build -a -o ./bin/ github.com/envoyproxy/gateway/cmd/envoy-gateway

.PHONY: test
test:
	@go test ./...

.PHONY: docker-build
docker-build: test ## Build the envoy-gateway docker image.
	docker build -t $(IMAGE):$(TAG) -f Dockerfile . 

.PHONY: docker-push
docker-push: ## Push docker image for envoy-gateway.
	docker push $(IMAGE):$(TAG)

