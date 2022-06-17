# ROOT Options:
ROOT_PACKAGE=github.com/envoyproxy/gateway

# Set Root Directory Path
ifeq ($(origin ROOT_DIR),undefined)
ROOT_DIR := $(abspath $(shell  pwd -P))
endif

# Set Output Directory Path
ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(ROOT_DIR)/bin
$(shell mkdir -p $(OUTPUT_DIR))
endif

# Set the version number. you should not need to do this
# for the majority of scenarios.
ifeq ($(origin VERSION), undefined)
VERSION := $(shell git describe --abbrev=0 --dirty --always --tags | sed 's/-/./g')
endif

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

# Docker variables
# REGISTRY is the image registry to use for build and push image targets.
REGISTRY ?= docker.io/envoyproxy
# IMAGE is the image URL for build and push image targets.
IMAGE ?= ${REGISTRY}/gateway-dev
# REV is the short git sha of latest commit.
REV=$(shell git rev-parse --short HEAD)
# Tag is the tag to use for build and push image targets.
TAG ?= $(REV)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help
	@echo Envoy Gateway is an open source project for managing Envoy Proxy as a standalone or Kubernetes-based application gateway
	@echo
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

# Build the envoy-gateway binaries in target platforms
.PHONY: build.%
build.%:
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	@echo "===========> Building binary $(COMMAND) $(VERSION) for $(OS) $(ARCH)"
	@mkdir -p $(OUTPUT_DIR)/$(OS)/$(ARCH)
	@CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUTPUT_DIR)/$(OS)/$(ARCH)/$(COMMAND) -ldflags "$(GO_LDFLAGS)" $(ROOT_PACKAGE)/cmd/$(COMMAND)

# Build the envoy-gateway binaries in hosted platforms
.PHONY: build
build:  $(addprefix build., $(addprefix $(PLATFORM)., $(BINS)))

# Build the envoy-gateway binaries in multi platforms
.PHONY: build.multiarch
build.multiarch:  $(foreach p,$(PLATFORMS),$(addprefix build., $(addprefix $(p)., $(BINS))))

.PHONY: clean
clean: ## Clean the building output files
	@echo "===========> Cleaning all build output"
	@rm -rf $(OUTPUT_DIR)

.PHONY: test
test: ## Run go unit tests
	@go test ./...

.PHONY: docker-build
docker-build: build ## Build the envoy-gateway docker image.
	$(eval IMAGE_NAME := $(COMMAND))
	$(eval IMAGE_PLAT := $(subst _,/,$(PLATFORM)))
	@echo "===========> Building docker image $(IMAGE_NAME) $(VERSION) for $(IMAGE_PLAT)"
	@DOCKER_BUILDKIT=1 docker build -t $(IMAGE):$(TAG) -f Dockerfile bin

.PHONY: docker-push
docker-push: ## Push the docker image for envoy-gateway.
	@docker push $(IMAGE):$(TAG)

.PHONY: lint
lint: ## Run lint checks
lint: lint-golint lint-yamllint lint-codespell

.PHONY: lint-golint
lint-golint:
	@echo Running Go linter ...
	@golangci-lint run --build-tags=e2e --config=tools/golangci-lint/.golangci.yml

.PHONY: lint-yamllint
lint-yamllint:
	@echo Running YAML linter ...
	## TODO(lianghao208): add other directories later
	@yamllint --config-file=tools/yamllint/.yamllint changelogs/

.PHONY: lint-codespell
lint-codespell: CODESPELL_SKIP := $(shell cat tools/codespell/.codespell.skip | tr \\n ',')
lint-codespell:
	@codespell --skip $(CODESPELL_SKIP) --ignore-words tools/codespell/.codespell.ignorewords --check-filenames --check-hidden -q2

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=envoy-gateway-role crd webhook paths="./..." output:crd:artifacts:config=pkg/provider/kubernetes/config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object paths="./..."

.PHONY: kube-test
kube-test: manifests generate lint envtest ## Run Kubernetes provider tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: kube-install
kube-install: manifests kustomize ## Install CRDs into the Kubernetes cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build pkg/provider/kubernetes/config/crd | kubectl apply -f -

.PHONY: kube-uninstall
kube-uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build pkg/provider/kubernetes/config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.8.0

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
