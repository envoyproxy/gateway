# This is a wrapper to build and push docker image
#
# All make targets related to docker image are defined in this file.

# Docker variables
# REGISTRY is the image registry to use for build and push image targets.
REGISTRY ?= docker.io/envoyproxy
# IMAGE is the image URL for build and push image targets.
IMAGE ?= ${REGISTRY}/gateway-dev
# Tag is the tag to use for build and push image targets.
TAG ?= $(REV)

DOCKER := docker
DOCKER_SUPPORTED_API_VERSION ?= 1.32

# Determine image files by looking into build/docker/*/Dockerfile
IMAGES_DIR ?= $(wildcard ${ROOT_DIR}tools/docker/*)
# Determine images names by stripping out the dir names
IMAGES ?= envoy-gateway
IMAGE_PLATFORMS ?= linux_amd64 linux_arm64 

ifeq (${IMAGES},)
  $(error Could not determine IMAGES, set ROOT_DIR or run in source dir)
endif

##@ Image

.PHONY: image
image: ## Build docker images for host platform. See Option PLATFORM and BINS.
	@$(MAKE) image.build

.PHONY: push
push: ## Push docker images to registry.
	@$(MAKE) image.push

.PHONY: image.multiarch
image.multiarch: ## Build docker images for multiple platforms. See Option PLATFORMS and IMAGES.
	@$(MAKE) image.build.multiarch

.PHONY: push.multiarch
push.multiarch: ## Push docker images for multiple platforms to registry.
	@$(MAKE) image.push.multiarch

.PHONY: image.verify
image.verify:
	$(eval API_VERSION := $(shell $(DOCKER) version | grep -E 'API version: {1,6}[0-9]' | head -n1 | awk '{print $$3} END { if (NR==0) print 0}' ))
	$(eval PASS := $(shell echo "$(API_VERSION) > $(DOCKER_SUPPORTED_API_VERSION)" | bc))
	@if [ $(PASS) -ne 1 ]; then \
		$(DOCKER) -v ;\
		echo "Unsupported docker version. Docker API version should be greater than $(DOCKER_SUPPORTED_API_VERSION)"; \
		exit 1; \
	fi

.PHONY: image.build
image.build: image.verify
	@$(MAKE) $(addprefix image.build., $(addprefix $(IMAGE_PLAT)., $(IMAGES)))

.PHONY: image.build.multiarch
image.build.multiarch: image.verify
	# TODO: use buildx

.PHONY: image.build.%
image.build.%: go.build.%
	$(eval IMAGES := $(COMMAND))
	$(eval IMAGE_PLAT := $(subst _,/,$(PLATFORM)))
	@echo "===========> Building image $(IMAGES) in tag $(TAG) for $(IMAGE_PLAT)"
	$(eval BUILD_SUFFIX := --pull -t $(IMAGE):$(TAG) -f $(ROOT_DIR)/tools/docker/$(IMAGES)/Dockerfile bin)
	@echo "===========> Creating image tag $(REGISTRY)/$(IMAGES):$(TAG) for $(ARCH)"; \
	$(DOCKER) build --platform $(IMAGE_PLAT) $(BUILD_SUFFIX)

.PHONY: image.push
image.push: image.verify $(addprefix image.push., $(addprefix $(IMAGE_PLAT)., $(IMAGES)))

.PHONY: image.push.multiarch
image.push.multiarch: image.verify  $(foreach p,$(IMAGE_PLATFORMS),$(addprefix image.push., $(addprefix $(p)., $(IMAGES)))) 

.PHONY: image.push.%
image.push.%:
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval IMAGES := $(COMMAND))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	$(eval IMAGE_PLAT := $(subst _,/,$(PLATFORM)))
	@echo "===========> Pushing image $(IMAGES) $(TAG) to $(REGISTRY)"
	@echo "===========> Pushing docker image tag $(IMAGE):$(TAG) for $(ARCH)"; \
	$(DOCKER) push $(IMAGE):$(TAG); \
