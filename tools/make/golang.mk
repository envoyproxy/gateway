# This is a wrapper to build golang binaries
#
# All make targets related to golang are defined in this file.

GOPATH := $(shell go env GOPATH)
ifeq ($(origin GOBIN), undefined)
	GOBIN := $(GOPATH)/bin
endif

GO_VERSION = $(shell grep -oE "^go [[:digit:]]*\.[[:digit:]]*" go.mod | cut -d' ' -f2)

# Build the target binary in target platform.
# The pattern of build.% is `build.{Platform}.{Command}`.
# If we want to build envoy-gateway in linux amd64 platform, 
# just execute make go.build.linux_amd64.envoy-gateway.
.PHONY: go.build.%
go.build.%:
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	@$(call log, "Building binary $(COMMAND) with commit $(REV) for $(OS) $(ARCH)")
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUTPUT_DIR)/$(OS)/$(ARCH)/$(COMMAND) $(ROOT_PACKAGE)/cmd/$(COMMAND)

# Build the envoy-gateway binaries in the hosted platforms.
.PHONY: go.build
go.build: $(addprefix go.build., $(addprefix $(PLATFORM)., $(BINS)))

# Build the envoy-gateway binaries in multi platforms
# It will build the linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 binaries out.
.PHONY: go.build.multiarch
go.build.multiarch: $(foreach p,$(PLATFORMS),$(addprefix go.build., $(addprefix $(p)., $(BINS))))

.PHONY: go.test.unit
go.test.unit: ## Run go unit tests
	go test ./...

.PHONY: go.test.coverage
go.test.coverage: $(tools/setup-envtest) ## Run go unit and integration tests in GitHub Actions
	KUBEBUILDER_ASSETS="$(shell $(tools/setup-envtest) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... --tags=integration -race -coverprofile=coverage.xml -covermode=atomic

.PHONY: go.clean
go.clean: ## Clean the building output files
	@$(call log, "Cleaning all build output")
	rm -rf $(OUTPUT_DIR)

.PHONY: go.tidy
go.tidy:
	@$(LOG_TARGET)
	@go mod tidy -compat=$(GO_VERSION)
	@if test -n "$$(git status -s -- go.mod go.sum)"; then \
		git diff --exit-code go.mod; \
		git diff --exit-code go.sum; \
   		echo '\nError: ensure all changes have been committed!'; \
	else \
		echo 'Go module looks clean!'; \
   	fi

##@ Golang

.PHONY: build
build: ## Build envoy-gateway for host platform. See Option PLATFORM and BINS.
build: go.build

.PHONY: build-multiarch
build-multiarch: ## Build envoy-gateway for multiple platforms. See Option PLATFORMS and IMAGES.
build-multiarch: go.build.multiarch

.PHONY: test
test: ## Run all Go test of code sources.
test: go.test.unit

.PHONY: format
format: ## Update dependences with mod tidy.
format: go.tidy

.PHONY: clean
clean: ## Remove all files that are created during builds.
clean: go.clean
