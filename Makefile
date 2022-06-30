# All make targets should be defined in Makefile.targets.mk 
# All make targets should be implemented in tools/make/*.mk
# ====================================================================================================
# Supported Targets: (Run `make help` to see more information)
# ====================================================================================================
## build: Build envoy-gateway for host platform. See Option PLATFORM and BINS.
## image: Build docker images for host platform. See Option PLATFORM and BINS.
## push: Push docker images to registry.
## build.multiarch: Build envoy-gateway for multiple platforms. See Option PLATFORMS and IMAGES.
## image.multiarch: Build docker images for multiple platforms. See Option PLATFORMS and IMAGES.
## push.multiarch: Push docker images for multiple platforms to registry.
## lint: Run all linter of code sources, including golint, yamllint and codespell.
## test: Run all Go test of code sources.
## format: Format codes style with mod tidy, gofmt and goimports.
## clean: Remove all files that are created by building.
# ====================================================================================================

# This is a wrapper around `make` so it can run
# directly on the host or inside a container
#
# Set MAKE_IN_DOCKER=1 as an environment variable to run `make` inside
# a Docker container with preinstalled tools.

DOCKER_BUILDER_IMAGE ?= envoyproxy/gateway-dev-builder
DOCKER_BUILDER_TAG ?= latest
DOCKER_BUILD_CMD ?= DOCKER_BUILDKIT=1 docker build
DOCKER_RUN_CMD ?= docker run \
		  --rm \
		  -t \
		  -v /var/run/docker.sock:/var/run/docker.sock \
		  -v ${PWD}:/workspace

%:
ifeq ($(MAKE_IN_DOCKER), 1)
# Build builder image
	@$(DOCKER_BUILD_CMD) -t $(DOCKER_BUILDER_IMAGE):$(DOCKER_BUILDER_TAG) - < tools/docker/builder/Dockerfile
# Run make inside the builder image
# Run with MAKE_IN_DOCKER=0 to eliminate an infinite loop
	@$(DOCKER_RUN_CMD) $(DOCKER_BUILDER_IMAGE):$(DOCKER_BUILDER_TAG) MAKE_IN_DOCKER=0 $@
else
# Run make locally
	@$(MAKE) -f Makefile.targets.mk $@
endif
