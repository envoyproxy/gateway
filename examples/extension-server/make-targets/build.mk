##@ Build

# REGISTRY is the image registry to use when building/pushing docker image targets.
# TODO: Set this to your dockerhub or other registry
REGISTRY ?= exampleorg/registry

# Multi-architecture builds configurations
ARCHS ?= linux/amd64,linux/arm64

# Docker images to build
DOCKER_DIRS := $(wildcard ${ROOT_DIR}/docker/*)
IMAGES := $(foreach dir,${DOCKER_DIRS},$(notdir ${dir}))

# TAG is the tag to use when building/pushing image targets.
# Use the HEAD commit SHA as the default.
TAG ?= $(REV)

.PHONY: clean
clean: ## Clean up build artifacts
	rm -rf $(DIST)


.PHONY: build
build: ## Build binaries and then the docker images
build: binaries images

# =============================
# Requirements Check
# =============================

.PHONY: _check-go
_check-go:
	@if ! command -v go >/dev/null; then \
		echo "Go is not installed. Please install Go to proceed."; \
		exit 1; \
	fi
	@echo "Go is installed."


.PHONY: _check-go-version
_check-go-version: _check-go
	$(eval MOD_GO_VERSION := $(shell go list -m -f '{{.GoVersion}}'))
	$(eval SYS_GO_VERSION := $(shell go version | awk '{print $$3}' | sed 's/go//'))
	@echo "Required Go version from go.mod: $(MOD_GO_VERSION)"
	@echo "System Go version: $(SYS_GO_VERSION)"
	@# Normalize version strings for comparison
	$(eval MOD_GO_VERSION_NORM := $(shell echo $(MOD_GO_VERSION) | awk -F. '{printf("%d%03d%03d", $$1, $$2, $$3)}'))
	$(eval SYS_GO_VERSION_NORM := $(shell echo $(SYS_GO_VERSION) | awk -F. '{printf("%d%03d%03d", $$1, $$2, $$3)}'))
	@if [ $(SYS_GO_VERSION_NORM) -lt $(MOD_GO_VERSION_NORM) ]; then \
		echo "ERROR: System Go version ($(SYS_GO_VERSION)) is less than required by go.mod ($(MOD_GO_VERSION))."; false; \
	else \
		echo "System Go version is sufficient."; \
	fi

# =============================
# Go Binary Builds
# =============================

.PHONY: binaries
binaries: ## Build all of the binaries using goreleaser. The binaries are specified in .goreleaser.yaml
binaries: _check-go-version
	GORELEASER_CURRENT_TAG=$$(git describe --tags --exact --match 'v[0-9]*.[0-9]*.[0-9]*' --abbrev=0 --candidates=1000) \
	go run github.com/goreleaser/goreleaser@v1.18.2 build --clean --snapshot

# =============================
# Docker Builds:
# =============================


.PHONY: _docker-buildx
_docker-buildx:
	@if ! docker buildx version >/dev/null 2>&1; then \
		echo "Docker buildx not available. Please install Docker buildx to proceed."; \
		exit 1; \
	fi
	@if ! docker buildx ls | grep -q "multi-arch-builder"; then \
		echo "Creating a new buildx builder instance for multi-arch builds."; \
		docker buildx create --name multi-arch-builder --use; \
	else \
		docker buildx use multi-arch-builder; \
	fi
	@echo "Buildx is set up for multi-arch builds."


.PHONY: images
images: ## Build and push multiarch docker images
images: _docker-buildx $(addprefix _image., $(IMAGES)) _post-build-summary

# This is an intermediate target and not meant for public use.
# The pattern of _docker-buildx.% is _docker-buildx.{image} where image is the Docker image name
_image.%:
	$(eval IMAGE := $*)
	docker buildx build --platform $(ARCHS) \
		--build-arg BUILD_DIR=dist \
		--tag $(REGISTRY)/$(IMAGE):$(TAG) \
		-f ./docker/$(IMAGE)/Dockerfile \
		--push \
		.

.PHONY: _post-build-summary
_post-build-summary:
	@echo "Build successful. Images pushed:"
	$(foreach img,$(IMAGES),@echo "- $(REGISTRY)/$(IMAGE_PREFIX)$(img):$(TAG)")
