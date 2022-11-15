# This is a wrapper to build and push docker image
#
# All make targets related to docker image are defined in this file.

# Docker variables
# REGISTRY is the image registry to use for build and push image targets.
REGISTRY ?= docker.io/envoyproxy
# IMAGE_NAME is the name of EG image
# Use gateway-dev in default when developing
# Use gateway when releasing an image.
IMAGE_NAME ?= gateway-dev
# IMAGE is the image URL for build and push image targets.
IMAGE ?= ${REGISTRY}/${IMAGE_NAME}
# Tag is the tag to use for build and push image targets.
TAG ?= $(REV)

DOCKER := DOCKER_BUILDKIT=1 docker
DOCKER_SUPPORTED_API_VERSION ?= 1.32

# Determine image files by looking into build/docker/*/Dockerfile
IMAGES_DIR ?= $(wildcard ${ROOT_DIR}tools/docker/*)
# Determine images names by stripping out the dir names
IMAGES ?= envoy-gateway
IMAGE_PLATFORMS ?= linux_amd64 linux_arm64

BUILDX_CONTEXT = gateway-build-tools-builder
# Convert to linux/arm64,linux/amd64
$(eval BUILDX_PLATFORMS = $(shell echo "${IMAGE_PLATFORMS}" | sed "s# #,#;s#_#/#g"))
EMULATE_PLATFORMS = amd64 arm64
EMULATE_TARGETS = $(addprefix image.multiarch.emulate.,$(EMULATE_PLATFORMS))

ifeq (${IMAGES},)
  $(error Could not determine IMAGES, set ROOT_DIR or run in source dir)
endif

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
image.build: $(addprefix image.build.$(IMAGE_PLAT)., $(IMAGES))

.PHONY: image.build.%
image.build.%: image.verify
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval IMAGES := $(COMMAND))
	$(eval IMAGE_PLAT := $(subst _,/,$(PLATFORM)))
	@$(call log, "Building image $(IMAGES) in tag $(TAG) for $(IMAGE_PLAT)")
	$(eval BUILD_SUFFIX := --pull -t $(IMAGE):$(TAG) -f $(ROOT_DIR)/tools/docker/$(IMAGES)/Dockerfile bin)
	@$(call log, "Creating image tag $(REGISTRY)/$(IMAGES):$(TAG) for $(ARCH)")
	$(DOCKER) build --platform $(IMAGE_PLAT) $(BUILD_SUFFIX)

.PHONY: image.push
image.push: $(addprefix image.push.$(IMAGE_PLAT)., $(IMAGES))

.PHONY: image.push.%
image.push.%: image.build.%
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval IMAGES := $(COMMAND))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	$(eval IMAGE_PLAT := $(subst _,/,$(PLATFORM)))
	@$(call log, "Pushing image $(IMAGES) $(TAG) to $(REGISTRY)")
	@$(call log, "Pushing docker image tag $(IMAGE):$(TAG) for $(ARCH)")
	$(DOCKER) push $(IMAGE):$(TAG)

.PHONY: image.multiarch.verify
image.multiarch.verify:
	$(eval PASS := $(shell docker buildx --help | grep "docker buildx" ))
	@if [ -z "$(PASS)" ]; then \
		echo "Cannot find docker buildx, please install first"; \
		exit 1;\
	fi

.PHONY: image.multiarch.emulate $(EMULATE_TARGETS)
image.multiarch.emulate: $(EMULATE_TARGETS)
$(EMULATE_TARGETS): image.multiarch.emulate.%:
# Install QEMU emulator, the same emulator as the host will report an error but can safe ignore
	docker run --rm --privileged tonistiigi/binfmt --install linux/$*

.PHONY: image.multiarch.setup
image.multiarch.setup: image.verify image.multiarch.verify image.multiarch.emulate
	docker buildx rm $(BUILDX_CONTEXT) || :
	docker buildx create --use --name $(BUILDX_CONTEXT) --platform "${BUILDX_PLATFORMS}"

.PHONY: image.build.multiarch
image.build.multiarch:
	docker buildx build bin -f "$(ROOT_DIR)/tools/docker/$(IMAGES)/Dockerfile" -t "${IMAGE}:${TAG}" --platform "${BUILDX_PLATFORMS}"

.PHONY: image.push.multiarch
image.push.multiarch:
	docker buildx build bin -f "$(ROOT_DIR)/tools/docker/$(IMAGES)/Dockerfile" -t "${IMAGE}:${TAG}" --platform "${BUILDX_PLATFORMS}" --push

##@ Image

.PHONY: image
image: ## Build docker images for host platform. See Option PLATFORM and BINS.
image: go.build image.build

.PHONY: image-multiarch
image-multiarch: ## Build docker images for multiple platforms. See Option PLATFORMS and IMAGES.
image-multiarch: image.multiarch.setup go.build.multiarch image.build.multiarch

.PHONY: push
push: ## Push docker images to registry.
push: image.push

.PHONY: push-multiarch
push-multiarch: ## Push docker images for multiple platforms to registry.
push-multiarch: image.multiarch.setup go.build.multiarch image.push.multiarch

