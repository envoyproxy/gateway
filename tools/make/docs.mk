DOCS_DIR := docs
DOCS_OUTPUT_DIR := docs/html

.PHONY: docs
docs: $(tools/sphinx-build)
	tools/bin/sphinx-build -b html $(DOCS_DIR) $(DOCS_OUTPUT_DIR)

.PHONY: docs.clean
docs.clean: ## Clean the built docs
	@$(call log, "Cleaning all built docs")
	rm -rf $(DOCS_OUTPUT_DIR)

.PHONY: clean
clean: ## Remove all files that are created during builds.
clean: docs.clean
