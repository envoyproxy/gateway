# This is a wrapper to build golang binaries
#
# All make targets related to golang are defined in this file.

VERSION_PACKAGE := github.com/envoyproxy/gateway/internal/cmd/version

GO_LDFLAGS += -X $(VERSION_PACKAGE).envoyGatewayVersion=$(shell cat VERSION) \
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

.PHONY: go.test.fuzz
go.test.fuzz: ## Run all fuzzers in the test/fuzz folder one by one
	@$(LOG_TARGET)
	@for dir in $$(go list -f '{{.Dir}}' ./test/fuzz/...); do \
		for test in $$(go test -list=Fuzz.* $$dir | grep ^Fuzz); do \
			echo "go test -fuzz=$$test -fuzztime=$(FUZZ_TIME)"; \
			(cd $$dir && go test -fuzz=$$test -fuzztime=$(FUZZ_TIME)) || exit 1; \
		done; \
	done

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
	go test -timeout 60s github.com/envoyproxy/gateway/internal/gatewayapi/resource --override-testdata=true

.PHONY: go.test.coverage
go.test.coverage: go.test.cel ## Run go unit and integration tests in GitHub Actions
	@$(LOG_TARGET)
	KUBEBUILDER_ASSETS="$(shell go tool setup-envtest use $(ENVTEST_K8S_VERSION) -p path)" \
		go test ./... --tags=integration -race -coverprofile=coverage.xml -covermode=atomic

.PHONY: go.test.cel
go.test.cel: manifests # Run the CEL validation tests
	@$(LOG_TARGET)
	@for ver in $(ENVTEST_K8S_VERSIONS); do \
  		echo "Run CEL Validation on k8s $$ver"; \
        go clean -testcache; \
        KUBEBUILDER_ASSETS="$$(go tool setup-envtest use $$ver -p path)" \
         go test ./test/cel-validation --tags celvalidation -race; \
    done

.PHONY: go.clean
go.clean: ## Clean the building output files
	@$(LOG_TARGET)
	rm -rf $(OUTPUT_DIR)

.PHONY: go.mod.tidy
go.mod.tidy: ## Update and check dependences with go mod tidy.
	@$(LOG_TARGET)
	go mod tidy -compat=$(GO_VERSION)

.PHONY: go.mod.lint
lint: go.mod.lint
go.mod.lint: go.mod.tidy go.mod.tidy.examples ## Check if go.mod is clean
	@$(LOG_TARGET)
	@if test -n "$$(git status -s -- go.mod go.sum)"; then \
		git diff --exit-code go.mod; \
		git diff --exit-code go.sum; \
		$(call errorlog, "Error: ensure all changes have been committed!"); \
		exit 1; \
	else \
		$(call log, "Go module looks clean!"); \
   	fi

.PHONY: go.lint.fmt
go.lint.fmt:
	@$(LOG_TARGET)
	@go tool golangci-lint fmt --build-tags=$(LINT_BUILD_TAGS) --config=tools/linter/golangci-lint/.golangci.yml

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
