# This is a wrapper to build and push golang binaries
#
# All make targets related to golang are defined in this file.

GOPATH := $(shell go env GOPATH)
ifeq ($(origin GOBIN), undefined)
	GOBIN := $(GOPATH)/bin
endif

# Build the target binary in target platform.
# The pattern of build.% is `build.{Platform}.{Command}`.
# If we want to build envoy-gateway in linux amd64 platform, 
# just execute make build.linux_amd64.envoy-gateway.
.PHONY: go.build.%
go.build.%:
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	@echo "===========> Building binary $(COMMAND) with commit $(REV) in version $(VERSION) for $(OS) $(ARCH)"
	@CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUTPUT_DIR)/$(OS)/$(ARCH)/$(COMMAND) $(ROOT_PACKAGE)/cmd/$(COMMAND)

# Build the envoy-gateway binaries in the hosted platforms.
.PHONY: go.build
go.build:  $(addprefix go.build., $(addprefix $(PLATFORM)., $(BINS)))

# Build the envoy-gateway binaries in multi platforms
# It will build the linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 binaries out.
.PHONY: go.build.multiarch
go.build.multiarch:  $(foreach p,$(PLATFORMS),$(addprefix go.build., $(addprefix $(p)., $(BINS))))

.PHONY: go.test.unit
go.test.unit: ## Run go unit tests
	@go test ./...

.PHONY: go.test.coverage
go.test.coverage: ## Run go unit tests in GitHub Actions
	@go test ./... -race -coverprofile=coverage.xml -covermode=atomic

.PHONY: go.clean
go.clean: ## Clean the building output files
	@echo "===========> Cleaning all build output"
	@rm -rf $(OUTPUT_DIR)

.PHONY: go.format.verify
go.format.verify:
ifeq ($(shell which goimports), )
	@echo "===========> Installing missing goimports"
	@go get -v golang.org/x/tools/cmd/goimports
	@go install golang.org/x/tools/cmd/goimports
endif

.PHONY: go.format
go.format:  go.format.verify
	@echo "===========> Running go codes format"
	@gofmt -s -w .
	@goimports -w -local $(ROOT_PACKAGE) .
	@go mod tidy
