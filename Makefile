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

# Tools
ADDLICENSE = $(shell pwd)/bin/addlicense
ADDLICENSE_VERSION = 1.0.0
addlicense: ## Download addlicense.
	$(call go-install-tool,$(ADDLICENSE),github.com/google/addlicense@v$(ADDLICENSE_VERSION))

.PHONY: tools
tools: ## Install all the tools needed for development using go install.
	@$(MAKE) addlicense 

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
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

license_ignore = -ignore changelogs 
license_files  = .github Makefile cmd  
.PHONY: license
license: 
	@$(ADDLICENSE) -c "Envoyproxy Authors" $(license_ignore) $(license_files)

.PHONY: check
check: tools ## Verify contents of last commit
	@$(MAKE) license

.PHONY: build
build:
	@go build -o ./bin/ github.com/envoyproxy/gateway/cmd/envoy-gateway

.PHONY: test
test:
	@go test ./...

