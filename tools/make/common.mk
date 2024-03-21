# This is a wrapper to set common variables
#
# All make targets related to common variables are defined in this file.

# ====================================================================================================
# Configure Make itself:
# ====================================================================================================

# Turn off .INTERMEDIATE file removal by marking all files as
# .SECONDARY.  .INTERMEDIATE file removal is a space-saving hack from
# a time when drives were small; on modern computers with plenty of
# storage, it causes nothing but headaches.
#
# https://news.ycombinator.com/item?id=16486331
.SECONDARY:

SHELL:=/bin/bash

# ====================================================================================================
# ROOT Options:
# ====================================================================================================

ROOT_PACKAGE=github.com/envoyproxy/gateway

RELEASE_VERSION=$(shell cat VERSION)

# Set Root Directory Path
ifeq ($(origin ROOT_DIR),undefined)
ROOT_DIR := $(abspath $(shell  pwd -P))
endif

# Set Output Directory Path
ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(ROOT_DIR)/bin
endif

# REV is the short git sha of latest commit.
REV=$(shell git rev-parse --short HEAD)

# Supported Platforms for building multiarch binaries.
PLATFORMS ?= darwin_amd64 darwin_arm64 linux_amd64 linux_arm64 

# Set a specific PLATFORM
ifeq ($(origin PLATFORM), undefined)
	ifeq ($(origin GOOS), undefined)
		GOOS := $(shell go env GOOS)
	endif
	ifeq ($(origin GOARCH), undefined)
		GOARCH := $(shell go env GOARCH)
	endif
	PLATFORM := $(GOOS)_$(GOARCH)
	# Use linux as the default OS when building images
	IMAGE_PLAT := linux_$(GOARCH)
else
	GOOS := $(word 1, $(subst _, ,$(PLATFORM)))
	GOARCH := $(word 2, $(subst _, ,$(PLATFORM)))
	IMAGE_PLAT := $(PLATFORM)
endif

# List commands in cmd directory for building targets
COMMANDS ?= $(wildcard ${ROOT_DIR}/cmd/*)
BINS ?= $(foreach cmd,${COMMANDS},$(notdir ${cmd}))

ifeq (${COMMANDS},)
  $(error Could not determine COMMANDS, set ROOT_DIR or run in source dir)
endif
ifeq (${BINS},)
  $(error Could not determine BINS, set ROOT_DIR or run in source dir)
endif

# ====================================================================================================
# Includes:
# ====================================================================================================
include tools/make/tools.mk
include tools/make/golang.mk
include tools/make/image.mk
include tools/make/lint.mk
include tools/make/kube.mk
include tools/make/docs.mk
include tools/make/helm.mk
include tools/make/proto.mk

# Log the running target
LOG_TARGET = echo -e "\033[0;32m===========> Running $@ ... \033[0m"
# Log debugging info
define log
echo -e "\033[36m===========>$1\033[0m"
endef

define errorlog
echo -e "\033[0;31m===========>$1\033[0m"
endef

define USAGE_OPTIONS

Options:
  \033[36mBINS\033[0m       
		 The binaries to build. Default is all of cmd.
		 This option is available when using: make build|build-multiarch
		 Example: \033[36mmake build BINS="envoy-gateway"\033[0m
  \033[36mIMAGES\033[0m     
		 Backend images to make. Default is all of cmds.
		 This option is available when using: make image|image-multiarch|push|push-multiarch
		 Example: \033[36mmake image-multiarch IMAGES="envoy-gateway"\033[0m
  \033[36mPLATFORM\033[0m   
		 The specified platform to build.
		 This option is available when using: make build|image
		 Example: \033[36mmake build BINS="envoy-gateway" PLATFORM="linux_amd64""\033[0m
		 Supported Platforms: linux_amd64 linux_arm64 darwin_amd64 darwin_arm64
  \033[36mPLATFORMS\033[0m  
		 The multiple platforms to build.
		 This option is available when using: make build-multiarch
		 Example: \033[36mmake build-multiarch BINS="envoy-gateway" PLATFORMS="linux_amd64 linux_arm64"\033[0m
		 Default is "linux_amd64 linux_arm64 darwin_amd64 darwin_arm64".
endef
export USAGE_OPTIONS

##@ Common

.PHONY: generate
generate: ## Generate go code from templates and tags
generate: kube-generate helm-generate helm-template go.generate docs-api

## help: Show this help info.
.PHONY: help
help:
	@echo -e "Envoy Gateway is an open source project for managing Envoy Proxy as a standalone or Kubernetes-based application gateway\n"
	@echo -e "Usage:\n  make \033[36m<Target>\033[0m \033[36m<Option>\033[0m\n\nTargets:"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo -e "$$USAGE_OPTIONS"
