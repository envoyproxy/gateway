# This is a wrapper to set common variables
#
# All make targets related to common variables are defined in this file.

# Set Root Directory Path
ifeq ($(origin ROOT_DIR),undefined)
ROOT_DIR := $(abspath $(shell  pwd -P))
endif

# Set Output Directory Path
ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(ROOT_DIR)/bin
$(shell mkdir -p $(OUTPUT_DIR))
endif

# Set TMP Directory Path
ifeq ($(origin TMP_DIR),undefined)
TMP_DIR := $(OUTPUT_DIR)/tmp
$(shell mkdir -p $(TMP_DIR))
endif

# Set the version number. you should not need to do this
# for the majority of scenarios.
ifeq ($(origin VERSION), undefined)
VERSION := $(shell git describe --abbrev=0 --dirty --always --tags | sed 's/-/./g')
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
