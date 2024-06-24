##@ Generate

.PHONY: generate
generate: ## Generates all of the CRDs
	rm -rf $(ROOT_DIR)/crds/generated
	go generate $(ROOT_DIR)/api/...
