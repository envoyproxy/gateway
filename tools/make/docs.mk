DOCS_OUTPUT_DIR := docs/html
RELEASE_VERSIONS := latest $(RELEASE_VERSION)

##@ Docs

.PHONY: docs
docs: $(tools/sphinx-build) ## Generate Envoy Gateway Docs Sources
	mkdir -p $(DOCS_OUTPUT_DIR)
	cp docs/index.html $(DOCS_OUTPUT_DIR)/index.html
	@for VERSION in $(RELEASE_VERSIONS); do \
		env BUILD_VERSION=$$VERSION $(shell go run ./cmd/envoy-gateway versions --env) $(tools/sphinx-build) -j auto -b html docs/$$VERSION $(DOCS_OUTPUT_DIR)/$$VERSION; \
	done

.PHONY: docs-release
docs-release: ## Generate Envoy Gateway Release Docs
	mkdir -p $(OUTPUT_DIR)
	@echo "\033[36m===========> Updated Release Version: $(TAG)\033[0m"
	echo $(TAG) > VERSION
	@echo "\033[36m===========> Added Release Doc: docs/$(TAG)\033[0m"
	cp -r docs/latest docs/$(TAG)
	@for DOC in $(shell ls docs/latest/user); do \
		cp docs/$(TAG)/user/$$DOC $(OUTPUT_DIR)/$$DOC ; \
		cat $(OUTPUT_DIR)/$$DOC | sed "s;latest;$(TAG);g" > $(OUTPUT_DIR)/$(TAG)-$$DOC ; \
		mv $(OUTPUT_DIR)/$(TAG)-$$DOC docs/$(TAG)/user/$$DOC ; \
		echo "\033[36m===========> Updated: docs/$(TAG)/user/$$DOC\033[0m" ; \
	done

.PHONY: docs.clean
docs.clean: ## Clean the built docs
	@$(call log, "Cleaning all built docs")
	rm -rf $(DOCS_OUTPUT_DIR)

.PHONY: clean
clean: ## Remove all files that are created during builds.
clean: docs.clean
