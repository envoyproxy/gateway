# Copyright 2022 Envoyproxy Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Go Environment

# Currently we resolve it using which. But more sophisticated approach is to use infer GOROOT.
go     := $(shell which go)
goarch := $(shell $(go) env GOARCH)
goexe  := $(shell $(go) env GOEXE)
goos   := $(shell $(go) env GOOS)

# Tools

BUF = $(shell pwd)/bin/buf
BUF_VERSION = 1.4.0
buf: ## Download buf.
	$(call go-install-tool,$(BUF),github.com/bufbuild/buf/cmd/buf@v$(BUF_VERSION))

PROTOC                    = $(shell pwd)/bin/protoc
PROTOC_REL                ="https://github.com/protocolbuffers/protobuf/releases"
PROTOC_VERSION            = 3.20.1
PROTOC_ARCHIVE_NAME       = $(if $(findstring $(goos),darwin),osx-x86_64.zip,linux-x86_64.zip)
PROTOC_ZIP                = protoc-$(PROTOC_VERSION)-$(PROTOC_ARCHIVE_NAME) 
PROTOC_UNZIP_DIR          = bin/.protoc-unzip
protoc: ## Download protoc.
	@curl -LO $(PROTOC_REL)/download/v$(PROTOC_VERSION)/$(PROTOC_ZIP)
	@unzip -qq $(PROTOC_ZIP) -d $(PROTOC_UNZIP_DIR)
	@cp $(PROTOC_UNZIP_DIR)/bin/protoc $(PROTOC)
	@rm -f $(PROTOC_ZIP)
	@rm -rf $(PROTOC_UNZIP_DIR)

PROTOC-GEN-GO  = $(shell pwd)/bin/protoc-gen-go
PROTOC-GEN-GO_VERSION = 1.28.0 
protoc-gen-go: ## Download protoc-gen-go.
	$(call go-install-tool,$(PROTOC-GEN-GO),google.golang.org/protobuf/cmd/protoc-gen-go@v$(PROTOC-GEN-GO_VERSION))

PROTOC-GEN-VALIDATE = $(shell pwd)/bin/protoc-gen-validate
PROTOC-GEN-VALIDATE_VERSION = 0.6.7 
protoc-gen-validate: ## Download protoc-gen-validate.
	$(call go-install-tool,$(PROTOC-GEN-VALIDATE),github.com/envoyproxy/protoc-gen-validate@v$(PROTOC-GEN-VALIDATE_VERSION))

.PHONY: tools
tools: ## Install all the tools needed for development using go install.
	@$(MAKE) buf protoc protoc-gen-go protoc-gen-validate

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# Development

.PHONY: gen
gen: tools ## Generate go files from proto files
	@$(MAKE) gen/proto

gen/proto:
	@$(BUF) generate

gen/mod:
	@$(BUF) mod update api

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
