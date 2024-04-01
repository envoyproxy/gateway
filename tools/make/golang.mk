# This is a wrapper to build golang binaries
#
# All make targets related to golang are defined in this file.

VERSION_PACKAGE := github.com/envoyproxy/gateway/internal/cmd/version

GO_LDFLAGS += -X $(VERSION_PACKAGE).envoyGatewayVersion=$(shell cat VERSION) \
	-X $(VERSION_PACKAGE).shutdownManagerVersion=$(TAG) \
	-X $(VERSION_PACKAGE).gitCommitID=$(GIT_COMMIT)

GIT_COMMIT:=$(shell git rev-parse HEAD)

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
	@$(LOG_TARGET)
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	@$(call log, "Building binary $(COMMAND) with commit $(REV) for $(OS) $(ARCH)")
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUTPUT_DIR)/$(OS)/$(ARCH)/$(COMMAND) -ldflags "$(GO_LDFLAGS)" $(ROOT_PACKAGE)/cmd/$(COMMAND)

# Build the envoy-gateway binaries in the hosted platforms.
.PHONY: go.build
go.build: $(addprefix go.build., $(addprefix $(PLATFORM)., $(BINS)))

# Build the envoy-gateway binaries in multi platforms
# It will build the linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 binaries out.
.PHONY: go.build.multiarch
go.build.multiarch: $(foreach p,$(PLATFORMS),$(addprefix go.build., $(addprefix $(p)., $(BINS))))


.PHONY: go.test.unit
go.test.unit: ## Run go unit tests
	go test -race ./...

.PHONY: go.testdata.complete
go.testdata.complete: ## Override test ouputdata
	@$(LOG_TARGET)
	go test -timeout 30s github.com/envoyproxy/gateway/internal/xds/translator --override-testdata=true
	go test -timeout 30s github.com/envoyproxy/gateway/internal/cmd/egctl --override-testdata=true
	go test -timeout 30s github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit --override-testdata=true
	go test -timeout 30s github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy --override-testdata=true
	go test -timeout 30s github.com/envoyproxy/gateway/internal/xds/bootstrap --override-testdata=true
	go test -timeout 60s github.com/envoyproxy/gateway/internal/gatewayapi --override-testdata=true

.PHONY: go.test.coverage
go.test.coverage: $(tools/setup-envtest) ## Run go unit and integration tests in GitHub Actions
	@$(LOG_TARGET)
	KUBEBUILDER_ASSETS="$(shell $(tools/setup-envtest) use $(ENVTEST_K8S_VERSION) -p path)" \
		go test ./... --tags=integration,celvalidation -race -coverprofile=coverage.xml -covermode=atomic

.PHONY: go.test.cel
go.test.cel: manifests $(tools/setup-envtest) # Run the CEL validation tests
	@$(LOG_TARGET)
	go clean -testcache # Ensure we're not using cached test results
	KUBEBUILDER_ASSETS="$(shell $(tools/setup-envtest) use $(ENVTEST_K8S_VERSION) -p path)" \
		go test ./test/cel-validation --tags=celvalidation -race

.PHONY: go.clean
go.clean: ## Clean the building output files
	@$(LOG_TARGET)
	rm -rf $(OUTPUT_DIR)

.PHONY: go.mod.lint
lint: go.mod.lint
go.mod.lint:
	@$(LOG_TARGET)
	@go mod tidy -compat=$(GO_VERSION)
	@if test -n "$$(git status -s -- go.mod go.sum)"; then \
		git diff --exit-code go.mod; \
		git diff --exit-code go.sum; \
		$(call errorlog, "Error: ensure all changes have been committed!"); \
		exit 1; \
	else \
		$(call log, "Go module looks clean!"); \
   	fi

.PHONY: go.generate
go.generate: ## Generate code from templates
	@$(LOG_TARGET)
	go generate ./...

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
format: ## Update and check dependences with go mod tidy.
format: go.mod.lint

.PHONY: clean
clean: ## Remove all files that are created during builds.
clean: go.clean

.PHONY: testdata
testdata: ## Override the testdata with new configurations.
testdata: go.testdata.complete