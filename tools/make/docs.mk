DOCS_OUTPUT_DIR := site/public
RELEASE_VERSIONS ?= $(foreach v,$(wildcard ${ROOT_DIR}/docs/*),$(notdir ${v}))

##@ Docs

.PHONY: docs
docs: docs.clean helm-readme-gen docs-api docs-api-headings ## Generate Envoy Gateway Docs Sources
	@$(LOG_TARGET)
	cd $(ROOT_DIR)/site && npm install
	cd $(ROOT_DIR)/site && npm run build:production
	cp tools/hack/get-egctl.sh $(DOCS_OUTPUT_DIR)

.PHONY: docs-release
docs-release: docs-release-prepare release-notes-docs docs-release-gen docs  ## Generate Envoy Gateway Release Docs

.PHONY: docs-serve
docs-serve: ## Start Envoy Gateway Site Locally
	@$(LOG_TARGET)
	cd $(ROOT_DIR)/site && npm run serve

.PHONY: clean
clean: ## Remove all files that are created during builds.
clean: docs.clean

.PHONY: docs.clean
docs.clean:
	@$(LOG_TARGET)
	rm -rf $(DOCS_OUTPUT_DIR)
	rm -rf site/node_modules
	rm -rf site/resources
	rm -f site/package-lock.json
	rm -f site/.hugo_build.lock

.PHONY: docs-api
docs-api: docs-api-gen helm-readme-gen docs-api-headings

.PHONY: helm-readme-gen
helm-readme-gen: $(tools/helm-docs)
	@$(LOG_TARGET)
	$(tools/helm-docs) charts/gateway-helm/ -f values.tmpl.yaml -o api.md
	mv charts/gateway-helm/api.md site/content/en/latest/install/api.md

.PHONY: docs-api-gen
docs-api-gen: $(tools/crd-ref-docs)
	@$(LOG_TARGET)
	$(tools/crd-ref-docs) \
	--source-path=api/v1alpha1 \
	--config=tools/crd-ref-docs/config.yaml \
	--templates-dir=tools/crd-ref-docs/templates \
	--output-path=site/content/en/latest/api/extension_types.md \
	--max-depth 10 \
	--renderer=markdown

.PHONY: docs-api-headings # Required since sphinx mst does not link to h4 headings.
docs-api-headings:
	@$(LOG_TARGET)
	tools/hack/docs-headings.sh site/content/en/latest/api/extension_types.md
	tools/hack/docs-headings.sh site/content/en/latest/install/api.md

.PHONY: docs-release-prepare
docs-release-prepare:
	@$(LOG_TARGET)
	mkdir -p $(OUTPUT_DIR)
	@$(call log, "Updated Release Version: $(TAG)")
	$(eval LAST_VERSION := $(shell cat VERSION))
	echo $(TAG) > VERSION

.PHONY: docs-release-gen
docs-release-gen:
	@$(LOG_TARGET)
	@$(call log, "Added Release Doc: site/content/en/$(TAG)")
	cp -r site/content/en/latest site/content/en/$(TAG)
	@for DOC in $(shell ls site/content/en/latest/user); do \
		cp site/content/en/$(TAG)/user/$$DOC $(OUTPUT_DIR)/$$DOC ; \
		cat $(OUTPUT_DIR)/$$DOC | sed "s;v0.0.0-latest;$(TAG);g" | sed "s;latest;$(TAG);g" > $(OUTPUT_DIR)/$(TAG)-$$DOC ; \
		mv $(OUTPUT_DIR)/$(TAG)-$$DOC site/content/en/$(TAG)/user/$$DOC ; \
		$(call log, "Updated: site/content/en/$(TAG)/user/$$DOC") ; \
	done

	@echo '[[params.versions]]' >> site/hugo.toml
	@echo '  version = "$(TAG)"' >> site/hugo.toml
	@echo '  url = "/$(TAG)"' >> site/hugo.toml

.PHONY: docs-check-links
docs-check-links:
	@$(LOG_TARGET)
	# Check for broken links, right now we are focusing on the v1.0.0
	# github.com does not allow access too often, there're a lot of 429 errors
	# TODO: find a way to remove github.com from ignore list
	# TODO: example.com is not a valid domain, we should remove it from ignore list
	linkinator site/public/ -r --concurrency 25 -s "github.com example.com _print v0.6.0 v0.5.0 v0.4.0 v0.3.0 v0.2.0"

release-notes-docs: $(tools/release-notes-docs)
	@$(LOG_TARGET)
	$(tools/release-notes-docs) release-notes/$(TAG).yaml site/content/en/latest/releases/; \
