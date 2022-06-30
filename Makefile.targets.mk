# ====================================================================================================
# ROOT Options:
# ====================================================================================================

ROOT_PACKAGE=github.com/envoyproxy/gateway

# ====================================================================================================
# Includes:
# ====================================================================================================
include tools/make/common.mk
include tools/make/golang.mk
include tools/make/image.mk
include tools/make/lint.mk

# ====================================================================================================
# Targets:
# ====================================================================================================

# ====================================================================================================
## build: Build envoy-gateway for host platform. See Option PLATFORM and BINS.
# ====================================================================================================
.PHONY: build
build:
	@$(MAKE) go.build

# ====================================================================================================
## image: Build docker images for host platform. See Option PLATFORM and BINS.
# ====================================================================================================
.PHONY: image
image:
	@$(MAKE) image.build

# ====================================================================================================
## push: Push docker images to registry.
# ====================================================================================================
.PHONY: push
push:
	@$(MAKE) image.push

# ====================================================================================================
## build.multiarch: Build envoy-gateway for multiple platforms. See Option PLATFORMS and IMAGES.
# ====================================================================================================
.PHONY: build.multiarch
build.multiarch:
	@$(MAKE) go.build.multiarch

# ====================================================================================================
## image.multiarch: Build docker images for multiple platforms. See Option PLATFORMS and IMAGES.
# ====================================================================================================
.PHONY: image.multiarch
image.multiarch:
	@$(MAKE) image.build.multiarch

# ====================================================================================================
## push.multiarch: Push docker images for multiple platforms to registry.
# ====================================================================================================
.PHONY: push.multiarch
push.multiarch:
	@$(MAKE) image.push.multiarch

# ====================================================================================================
## lint: Run all linter of code sources, including golint, yamllint and codespell.
# ====================================================================================================
.PHONY: lint
lint: 
	@$(MAKE) lint.golint lint.yamllint lint.codespell

# ====================================================================================================
## test: Run all Go test of code sources.
# ====================================================================================================
.PHONY: test
test: 
	@$(MAKE) go.test.unit

# ====================================================================================================
## format: Format codes style with mod tidy, gofmt and goimports.
# ====================================================================================================
.PHONY: format
format:
	@$(MAKE) go.format

# ====================================================================================================
## clean: Remove all files that are created by building.
# ====================================================================================================
.PHONY: clean
clean:
	@$(MAKE) go.clean

# ====================================================================================================
# Usage
# ====================================================================================================

define USAGE_OPTIONS

Options:
  BINS         The binaries to build. Default is all of cmd.
               This option is available when using: make build/build.multiarch
               Example: make build BINS="envoy-gateway"
  IMAGES       Backend images to make. Default is all of cmds.
               This option is available when using: make image/image.multiarch/push/push.multiarch
               Example: make image.multiarch IMAGES="envoy-gateway"
  PLATFORM     The specified platform to build.
               This option is available when using: make build/image
               Example: make build BINS="envoy-gateway" PLATFORM="linux_amd64"
               Supported Platforms: linux_amd64 linux_arm64 darwin_amd64 darwin_arm64
  PLATFORMS    The multiple platforms to build.
               This option is available when using: make build.multiarch
               Example: make image.multiarch IMAGES="envoy-gateway" PLATFORMS="linux_amd64 linux_arm64"
               Default is "linux_amd64 linux_arm64 darwin_amd64 darwin_arm64".
endef
export USAGE_OPTIONS

# ====================================================================================================
# Help
# ====================================================================================================

## help: Show this help info.
.PHONY: help
help: Makefile
	@echo "Envoy Gateway is an open source project for managing Envoy Proxy as a standalone or Kubernetes-based application gateway\n"
	@echo "Usage: make <Targets> <Options> ...\n\nTargets:"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo "$$USAGE_OPTIONS"
