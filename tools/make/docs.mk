DOCS_OUTPUT_DIR := docs/html
RELEASE_VERSIONS ?= $(foreach v,$(wildcard ${ROOT_DIR}/docs/*),$(notdir ${v}))

##@ Docs

.PHONY: docs
docs: docs.clean $(tools/sphinx-build) docs-api helm-readme-gen ## Generate Envoy Gateway Docs Sources
	@$(LOG_TARGET)
	mkdir -p $(DOCS_OUTPUT_DIR)
	cp docs/index.html $(DOCS_OUTPUT_DIR)/index.html
	cp tools/hack/get-egctl.sh $(DOCS_OUTPUT_DIR)/get-egctl.sh
	@for VERSION in $(RELEASE_VERSIONS); do \
		env BUILD_VERSION=$$VERSION \
		ENVOY_PROXY_VERSION=$(shell go run ./cmd/envoy-gateway versions -o json | jq -r ".envoyProxyVersion") \
		GATEWAYAPI_VERSION=$(shell go run ./cmd/envoy-gateway versions -o json | jq -r ".gatewayAPIVersion") \
		$(tools/sphinx-build) -j auto -b html docs/$$VERSION $(DOCS_OUTPUT_DIR)/$$VERSION; \
	done

.PHONY: docs-release
docs-release: docs-release-prepare docs-release-gen docs  ## Generate Envoy Gateway Release Docs

.PHONY: docs-serve
docs-serve: ## Start Envoy Gateway Site Locally
	@$(LOG_TARGET)
	python3 -m http.server -d $(DOCS_OUTPUT_DIR)

.PHONY: clean
clean: ## Remove all files that are created during builds.
clean: docs.clean

.PHONY: docs.clean
docs.clean:
	@$(LOG_TARGET)
	rm -rf $(DOCS_OUTPUT_DIR)

.PHONY: docs-api
docs-api: docs-api-gen docs-api-headings

.PHONY: helm-readme-gen
helm-readme-gen: $(tools/helm-docs)
	@$(LOG_TARGET)
	$(tools/helm-docs) charts/gateway-helm/ -f values.tmpl.yaml -o api.md
	mv charts/gateway-helm/api.md docs/latest/helm/api.md

.PHONY: docs-api-gen
docs-api-gen: $(tools/crd-ref-docs)
	$(tools/crd-ref-docs) \
	--source-path=api/v1alpha1 \
	--config=tools/crd-ref-docs/config.yaml \
	--output-path=docs/latest/api/extension_types.md \
	--max-depth 10 \
	--renderer=markdown

.PHONY: docs-api-headings # Required since sphinx mst does not link to h4 headings.
docs-api-headings:
	@$(LOG_TARGET)
	tools/hack/docs-headings.sh

.PHONY: docs-release-prepare
docs-release-prepare:
	@$(LOG_TARGET)
	mkdir -p $(OUTPUT_DIR)
	@$(call log, "Updated Release Version: $(TAG)")
	$(eval LAST_VERSION := $(shell cat VERSION))
	cat docs/index.html | sed "s;$(LAST_VERSION);$(TAG);g" > $(OUTPUT_DIR)/index.html
	mv $(OUTPUT_DIR)/index.html docs/index.html
	echo $(TAG) > VERSION

.PHONY: docs-release-gen
docs-release-gen:
	@$(LOG_TARGET)
	@$(call log, "Added Release Doc: docs/$(TAG)")
	cp -r docs/latest docs/$(TAG)
	@for DOC in $(shell ls docs/latest/user); do \
		cp docs/$(TAG)/user/$$DOC $(OUTPUT_DIR)/$$DOC ; \
		cat $(OUTPUT_DIR)/$$DOC | sed "s;v0.0.0-latest;$(TAG);g" | sed "s;latest;$(TAG);g" > $(OUTPUT_DIR)/$(TAG)-$$DOC ; \
		mv $(OUTPUT_DIR)/$(TAG)-$$DOC docs/$(TAG)/user/$$DOC ; \
		$(call log, "Updated: docs/$(TAG)/user/$$DOC") ; \
	done
