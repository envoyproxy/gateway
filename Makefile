# This is a wrapper around `make` so it can run 
# directly on the host or inside a container
#
# All make targets must be defined in Makefile.targets.mk 

# Set MAKE_IN_DOCKER=1 as an envioronment variable to run `make` inside
# a Docker container with preinstalled tools.

DOCKER_BUILDER_IMAGE ?= envoyproxy/gateway-dev-builder
DOCKER_BUILDER_TAG ?= latest
DOCKER_BUILD_CMD ?= DOCKER_BUILDKIT=1 docker build
DOCKER_RUN_CMD ?= docker run \
		  --rm \
		  -it \
		  -v /var/run/docker.sock:/var/run/docker.sock \
		  -v ${PWD}:/workspace 

%:
ifeq ($(MAKE_IN_DOCKER), 1)
# Build builder image
	@$(DOCKER_BUILD_CMD) -t $(DOCKER_BUILDER_IMAGE):$(DOCKER_BUILDER_TAG) - < Dockerfile.builder
# Run make inside the builder image
# Run with MAKE_IN_DOCKER=0 to eliminate an infinite loop
	@$(DOCKER_RUN_CMD) $(DOCKER_BUILDER_IMAGE):$(DOCKER_BUILDER_TAG) MAKE_IN_DOCKER=0 $@
else
# Run make locally	
	@$(MAKE) -f Makefile.targets.mk $@
endif
