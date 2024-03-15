# All make targets are implemented in make-targets/*.mk
# Run `make help` to see more information about available targets

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Alias `make` to `make build`
.PHONY: all
all: build

include make-targets/_core.mk
include make-targets/help.mk
include make-targets/lint.mk
include make-targets/generate.mk
include make-targets/build.mk
