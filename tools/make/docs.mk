DOCS_OUTPUT_DIR := site/public
RELEASE_VERSIONS ?= $(foreach v,$(wildcard ${ROOT_DIR}/docs/*),$(notdir ${v}))
# TODO: github.com does not allow access too often, there are a lot of 429 errors
#       find a way to remove github.com from ignore list
# TODO: example.com is not a valid domain, we should remove it from ignore list
# TODO: https://www.gnu.org/software/make became unstable, we should remove it from ignore list later
LINKINATOR_IGNORE := "github.com jwt.io githubusercontent.com example.com github.io gnu.org _print canva.com sched.co sap.com"
CLEAN_NODE_MODULES ?= true

##@ Docs

.PHONY: docs
docs: docs.clean helm-readme-gen docs-api copy-current-release-docs ## Generate Envoy Gateway Docs Sources
	@$(LOG_TARGET)
	cd $(ROOT_DIR)/site && npm install
	cd $(ROOT_DIR)/site && npm run build:production
	cp tools/hack/get-egctl.sh $(DOCS_OUTPUT_DIR)

.PHONY: copy-current-release-docs
copy-current-release-docs:  ## Copy the current release docs to the docs folder
	@$(LOG_TARGET)
	@CURRENT_RELEASE=$(shell ls $(ROOT_DIR)/site/content/en | grep -E '^v[0-9]+\.[0-9]+$$' | sort | tail -n 1); \
	echo "Copying the current release $$CURRENT_RELEASE docs to the docs folder"; \
	rm -rf $(ROOT_DIR)/site/content/en/docs; \
	mkdir -p $(ROOT_DIR)/site/content/en/docs; \
	cp -r $(ROOT_DIR)/site/content/en/$$CURRENT_RELEASE/** $(ROOT_DIR)/site/content/en/docs

.PHONY: docs-release
docs-release: docs-release-prepare docs-release-gen docs  ## Generate Envoy Gateway Release Docs

.PHONY: docs-serve
docs-serve: copy-current-release-docs ## Start Envoy Gateway Site Locally
	@$(LOG_TARGET)
	cd $(ROOT_DIR)/site && npm run serve

.PHONY: clean
clean: ## Remove all files that are created during builds.
clean: docs.clean

.PHONY: docs.clean
docs.clean:
	@$(LOG_TARGET)
	rm -rf $(DOCS_OUTPUT_DIR)
ifeq ($(CLEAN_NODE_MODULES),true)
	rm -rf site/node_modules
endif
	rm -rf site/resources
	rm -f site/package-lock.json
	rm -f site/.hugo_build.lock

.PHONY: docs-api
docs-api: docs-api-gen helm-readme-gen

.PHONY: helm-readme-gen
helm-readme-gen:
	@for chart in $(CHARTS); do \
		$(LOG_TARGET); \
		$(MAKE) $(addprefix helm-readme-gen., $$(basename $${chart})); \
	done

.PHONY: helm-readme-gen.%
helm-readme-gen.%:
	$(eval COMMAND := $(word 1,$(subst ., ,$*)))
	$(eval CHART_NAME := $(COMMAND))
	# use production ENV to generate helm api doc
	@if test -f "charts/${CHART_NAME}/values.tmpl.yaml"; then \
		ImageRepository=docker.io/envoyproxy/gateway ImageTag=latest ImagePullPolicy=IfNotPresent \
		envsubst < charts/${CHART_NAME}/values.tmpl.yaml > ./charts/${CHART_NAME}/values.yaml; \
	fi

	# generate helm readme doc
	@go tool helm-docs --template-files=tools/helm-docs/readme.${CHART_NAME}.gotmpl -g charts/${CHART_NAME} -f values.yaml -o README.md

	# change the placeholder to title before api helm docs generated: split by '-' and capitalize the first letters
	$(eval CHART_TITLE := $(shell echo "$(CHART_NAME)" | sed -E 's/\<./\U&/g; s/-/ /g' | awk '{for(i=1;i<=NF;i++){ $$i=toupper(substr($$i,1,1)) substr($$i,2) }}1'))
	sed 's/{CHART-NAME}/$(CHART_TITLE)/g' tools/helm-docs/api.gotmpl > tools/helm-docs/api.${CHART_NAME}.gotmpl
	@go tool helm-docs --template-files=tools/helm-docs/api.${CHART_NAME}.gotmpl -g charts/${CHART_NAME} -f values.yaml -o api.md
	mv charts/${CHART_NAME}/api.md site/content/en/latest/install/${CHART_NAME}-api.md
	rm tools/helm-docs/api.${CHART_NAME}.gotmpl

.PHONY: docs-api-gen
docs-api-gen:
	@$(LOG_TARGET)
	go tool crd-ref-docs \
	--source-path=api/v1alpha1 \
	--config=tools/crd-ref-docs/config.yaml \
	--templates-dir=tools/crd-ref-docs/templates \
	--output-path=site/content/en/latest/api/extension_types.md \
	--max-depth 100 \
	--renderer=markdown

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
	$(eval DOC_VERSION := $(shell cat VERSION | cut -d "." -f 1,2))
	@$(call log, "Added Release Doc: site/content/en/$(DOC_VERSION)")
	cp -r site/content/en/latest/ site/content/en/$(DOC_VERSION)/
	@echo "" >> site/hugo.toml
	@echo '[[params.versions]]' >> site/hugo.toml
	@echo '  version = "$(DOC_VERSION)"' >> site/hugo.toml
	@echo '  url = "/$(DOC_VERSION)"' >> site/hugo.toml

.PHONY: docs-check-links
docs-check-links: # Check for broken links in the docs
	@$(LOG_TARGET)
	linkinator site/public/ -r --concurrency 25 --skip $(LINKINATOR_IGNORE)

docs-markdown-lint:
	markdownlint -c .github/markdown_lint_config.json site/content/*

release-notes-docs: $(tools/release-notes-docs)
	@$(LOG_TARGET)
	@for file in $(wildcard release-notes/v*.yaml); do \
		$(tools/release-notes-docs) $$file site/content/en/news/releases/notes; \
	done
