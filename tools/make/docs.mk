DOCS_OUTPUT_DIR := site/public
RELEASE_VERSIONS ?= $(foreach v,$(wildcard ${ROOT_DIR}/docs/*),$(notdir ${v}))
# TODO: github.com does not allow access too often, there are a lot of 429 errors
#       find a way to remove github.com from ignore list
# TODO: example.com is not a valid domain, we should remove it from ignore list
# TODO: https://www.gnu.org/software/make became unstable, we should remove it from ignore list later
LINKINATOR_IGNORE := "opentelemetry.io github.com jwt.io githubusercontent.com example.com github.io gnu.org _print canva.com sched.co sap.com httpbin.org nemlig.com verve.com developer.hashicorp.com"
CLEAN_NODE_MODULES ?= true

##@ Docs

.PHONY: docs-gen
docs-gen: docs.clean helm-readme-gen docs-api copy-current-release-docs docs-sync-owners ## Generate Envoy Gateway Docs Sources
	@$(LOG_TARGET)
	cd $(ROOT_DIR)/site && npm install
	cd $(ROOT_DIR)/site && npm run build:production
	cp tools/hack/get-egctl.sh $(DOCS_OUTPUT_DIR)

.PHONY: docs
docs: docs-gen docs-check ## Generate docs and verify no changes are needed

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

# Define base URLs for downloading documentation, examples, and images from the gateway-api repository.
DOC_SRC_URL=https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/main/site-src/api-types
YAML_SRC_BASE_URL=https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/main/examples
IMAGE_SRC_BASE_URL=https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/main/site-src/images

# Define destination directories for images and documentation within the envoy gateway repository.
IMAGE_DEST_DIR=$(ROOT_DIR)/site/static/img
DOC_DEST_DIR=$(ROOT_DIR)/site/content/en/latest/api/gateway_api

# List of documentation files to synchronize.
SYNC_FILES := gateway.md gatewayclass.md httproute.md grpcroute.md backendtlspolicy.md referencegrant.md

# Main target to synchronize all gateway-api documentation components.
sync-gwapi-docs: gwapi-doc-download gwapi-doc-transform gwapi-doc-download-includes gwapi-doc-replace-includes gwapi-doc-clean-includes gwapi-doc-remove-special-lines gwapi-doc-update-relative-links

# Download the documentation files from the gateway-api repository to the local destination directory.
gwapi-doc-download:
# 	@$(LOG_TARGET)
# 	@mkdir -p $(DOC_DEST_DIR)
# 	@$(foreach file, $(SYNC_FILES), curl -s -o $(DOC_DEST_DIR)/$(file) $(DOC_SRC_URL)/$(file);)

# Transform the first line of each markdown file to a header format suitable for Hugo.
gwapi-doc-transform:
# 	@$(LOG_TARGET)
# 	@$(foreach file, $(SYNC_FILES), sed -i '1s/^# \(.*\)/+++\ntitle = "\1"\n+++/' $(DOC_DEST_DIR)/$(file);)

# Download included YAML files referenced within the documentation.
gwapi-doc-download-includes:
	@$(LOG_TARGET)
	@$(foreach file, $(SYNC_FILES), \
		grep -o "{% include '.*' %}" $(DOC_DEST_DIR)/$(file) | sed "s/{% include '\(.*\)' %}/\1/" | \
		while read yaml_path; do \
			yaml_file=$$(basename $$yaml_path); \
			curl -s -o $(DOC_DEST_DIR)/$$yaml_file $(YAML_SRC_BASE_URL)/$$yaml_path; \
		done;)

# Replace include statements with the actual content of the YAML files.
gwapi-doc-replace-includes:
	@$(LOG_TARGET)
	@$(foreach file, $(SYNC_FILES), \
		perl -0777 -i -pe ' \
			while (/{% include '\''(.*?)'\'' %}/) { \
				$$yaml_path = $$1; \
				$$yaml_file = `basename $$yaml_path`; \
				$$yaml_content = `cat $(DOC_DEST_DIR)/$$yaml_file`; \
				s/{% include '\''$$yaml_path'\'' %}/$$yaml_content/; \
			} \
		' $(DOC_DEST_DIR)/$(file);)

# Clean up by removing downloaded YAML files after processing.
gwapi-doc-clean-includes:
	@$(LOG_TARGET)
	@find $(DOC_DEST_DIR) -name '*.yaml' -exec rm {} +

# Remove special lines that start with '!!!' or `???` from the documentation.
gwapi-doc-remove-special-lines:
	@$(LOG_TARGET)
# 	@$(foreach file, $(SYNC_FILES), \
# 		sed -i '/^[\?!]\{3\}/d' $(DOC_DEST_DIR)/$(file);)

# Update relative links
gwapi-doc-update-relative-links:
# 	@$(foreach file, $(SYNC_FILES), \
# 		sed -i -e 's/\(\.*\]\)(\(\.\.\/[^:]*\))/\1(https:\/\/gateway-api.sigs.k8s.io\2)/g' -e 's/\(\[.*\]: \)\(\/[^:]*\)/\1https:\/\/gateway-api.sigs.k8s.io\2/g' -e 's/\(\[.*\]: \)\(\.\.\/[^:]*\)/\1https:\/\/gateway-api.sigs.k8s.io\2/g' $(DOC_DEST_DIR)/$(file);)
# 	@$(foreach file, $(SYNC_FILES), \
# 		sed -i -e 's/https:\/\/gateway-api.sigs.k8s.io\.\./https:\/\/gateway-api.sigs.k8s.io/g' $(DOC_DEST_DIR)/$(file);)

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
	$(GO_TOOL) helm-docs --template-files=tools/helm-docs/readme.${CHART_NAME}.gotmpl -g charts/${CHART_NAME} -f values.yaml -o README.md

	# change the placeholder to title before api helm docs generated: split by '-' and capitalize the first letters
	$(eval CHART_TITLE := $(shell echo "$(CHART_NAME)" | sed -E 's/\<./\U&/g; s/-/ /g' | awk '{for(i=1;i<=NF;i++){ $$i=toupper(substr($$i,1,1)) substr($$i,2) }}1'))
	sed 's/{CHART-NAME}/$(CHART_TITLE)/g' tools/helm-docs/api.gotmpl > tools/helm-docs/api.${CHART_NAME}.gotmpl
	$(GO_TOOL) helm-docs --template-files=tools/helm-docs/api.${CHART_NAME}.gotmpl -g charts/${CHART_NAME} -f values.yaml -o api.md
	mv charts/${CHART_NAME}/api.md site/content/en/latest/install/${CHART_NAME}-api.md
	rm tools/helm-docs/api.${CHART_NAME}.gotmpl

.PHONY: docs-api-gen
docs-api-gen:
	@$(LOG_TARGET)
	$(GO_TOOL) crd-ref-docs \
	--source-path=api/v1alpha1 \
	--config=tools/crd-ref-docs/config.yaml \
	--templates-dir=tools/crd-ref-docs/templates \
	--output-path=site/content/en/latest/api/extension_types.md \
	--max-depth 100 \
	--renderer=markdown

.PHONY: docs-release-prepare
docs-release-prepare: sync-gwapi-docs
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

.PHONY: docs-sync-owners
docs-sync-owners: $(tools/sync-docs-codeowners) # Sync maintainers and emeritus-maintainers from OWNERS to CODEOWNERS.md
	@$(LOG_TARGET)
	$(tools/sync-docs-codeowners)

.PHONY: docs-check-links
docs-check-links: # Check for broken links in the docs
	@$(LOG_TARGET)
	linkinator site/public/ -r --concurrency 25 --retry-errors --retry --retry-errors-jitter --retry-errors-count 5 --skip $(LINKINATOR_IGNORE) --verbosity error

docs-markdown-lint:
	markdownlint -c .github/markdown_lint_config.json site/content/*

.PHONY: docs-check
docs-check: ## Verify no doc changes are needed
	@$(LOG_TARGET)
	@if [ ! -z "`git status --porcelain`" ]; then \
		$(call errorlog, ERROR: Some files need to be updated, please run 'make docs' to include any changed files to your PR); \
		git diff --exit-code; \
	fi

release-notes-docs: $(tools/release-notes-docs)
	@$(LOG_TARGET)
	@for file in $(wildcard release-notes/$(shell cat VERSION).yaml); do \
		$(tools/release-notes-docs) $$file site/content/en/news/releases/notes; \
	done
