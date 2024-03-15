# This file contains a set common variables
#
# All make targets related to common variables are defined in this file.

# =============================
# Root Settings:
# =============================

# Set project root directory path
ifeq ($(origin ROOT_DIR),undefined)
ROOT_DIR := $(abspath $(shell  pwd -P))
endif

# Directory for the built binaries
DIST_DIR ?= $(ROOT_DIR)/dist

# Commit SHA
TAG := $(shell git rev-parse --short HEAD)
