DOCS_OUTPUT_DIR := site/public
RELEASE_VERSIONS ?= $(foreach v,$(wildcard ${ROOT_DIR}/docs/*),$(notdir ${v}))
# TODO: github.com does not allow access too often, there are a lot of 429 errors
#       find a way to remove github.com from ignore list
# TODO: example.com is not a valid domain, we should remove it from ignore list
# TODO: https://www.gnu.org/software/make became unstable, we should remove it from ignore list later
# NOTE: envoyproxy.io/slack redirects to a Slack shared invite, and Slack answers 403 to
#       automated clients, so the redirect target can't be verified by the link checker.
LINKINATOR_IGNORE := "opentelemetry.io \
	blog.envoyproxy.io \
	envoyproxy.io/slack \
	ntia.gov \
	nvd.nist.gov \
	github.com \
	jwt.io \
	githubusercontent.com \
	example.com \
	foo.bar.com \
	github.io \
	gnu.org \
	_print \
	canva.com \
	communityinviter.com \
	sched.co \
	sap.com \
	httpbin.org \
	nemlig.com \
	verve.com \
	goteleport.com \
	developer.hashicorp.com \
	www.signal-ai.com \
	v0.1 \
	v0.2 \
	v0.3 \
	v0.4 \
	v0.5 \
	v0.6 \
	v1.0 \
	v1.1 \
	v1.2 \
	v1.3 \
	v1.4"
CLEAN_NODE_MODULES ?= true

##@ Docs

.PHONY: docs-gen
docs-gen: docs.clean helm-readme-gen docs-api copy-current-release-docs docs-sync-owners ## Generate Envoy Gateway Docs Sources
	@$(LOG_TARGET)
	cd $(ROOT_DIR)/site && npm ci
	cd $(ROOT_DIR)/site && npm run build:production
	cp tools/hack/get-egctl.sh $(DOCS_OUTPUT_DIR)

.PHONY: docs
docs: docs-gen docs-check ## Generate docs and verify no changes are needed

.PHONY: sync-benchmark-dashboard
sync-benchmark-dashboard: ## Sync release benchmark dashboard data and rebuild tracked static assets. Requires VERSION=vX.Y.Z.
	@test -n "$(VERSION)" || (echo "VERSION is required, e.g. make sync-benchmark-dashboard VERSION=v1.7.1" && exit 1)
	@./tools/src/benchmark-dashboard-sync/sync.sh --version "$(VERSION)" --force

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
	rm -f site/.hugo_build.lock

.PHONY: docs-api
docs-api: docs-api-gen helm-readme-gen

# Gateway API documentation sync.
# Since Gateway API v1.6 the upstream docs are a Hugo site, so the sources live
# under site/content/en/reference/api-types (BackendTLSPolicy under policy/)
# instead of the old MkDocs site-src/api-types tree. The rewriting rules live in
# tools/hack/gwapi-doc-sync.sh.
GWAPI_DOC_SRC_PATH ?= site/content/en/reference/api-types

# List of documentation files to synchronize, relative to GWAPI_DOC_SRC_PATH.
GWAPI_SYNC_FILES ?= gateway.md gatewayclass.md httproute.md grpcroute.md referencegrant.md policy/backendtlspolicy.md

# Destination directory for the synchronized documentation.
DOC_DEST_DIR=$(ROOT_DIR)/site/content/en/latest/api/gateway_api

.PHONY: sync-gwapi-docs
sync-gwapi-docs: ## Sync the gateway-api api-types docs for the bundled Gateway API version.
	@$(LOG_TARGET)
	@GATEWAY_API_VERSION=$(GATEWAY_API_VERSION) \
		GATEWAY_API_MINOR_VERSION=$(GATEWAY_API_MINOR_VERSION) \
		DOC_SRC_PATH=$(GWAPI_DOC_SRC_PATH) \
		DOC_DEST_DIR=$(DOC_DEST_DIR) \
		SYNC_FILES="$(GWAPI_SYNC_FILES)" \
		./tools/hack/gwapi-doc-sync.sh

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
docs-check-links: $(tools/linkinator) # Check for broken links in the docs
	@$(LOG_TARGET)
	$(tools/linkinator) site/public/ -r --concurrency 25 --retry-errors --retry --retry-errors-jitter --retry-errors-count 5 --skip $(LINKINATOR_IGNORE) --verbosity error

docs-markdown-lint: $(tools/markdownlint)
	$(tools/markdownlint) -c .github/markdown_lint_config.json site/content/*

.PHONY: docs-check
docs-check: ## Verify no doc changes are needed
	@$(LOG_TARGET)
	@if [ ! -z "`git status --porcelain`" ]; then \
		$(call errorlog, ERROR: Some files need to be updated, please run 'make docs' to include any changed files to your PR); \
		git diff --exit-code; \
	fi

release-notes-docs: $(tools/release-notes-docs) # Read version from Environment variable, if not set, read from VERSION file
	@$(LOG_TARGET)
	$(eval RELEASE_NOTE_VERSION := $(if $(RELEASE_NOTE_VERSION),$(RELEASE_NOTE_VERSION),$(shell cat VERSION)))
	@echo "Generating release notes for version $(RELEASE_NOTE_VERSION)"
	@for file in $(wildcard release-notes/$(RELEASE_NOTE_VERSION).yaml); do \
		$(tools/release-notes-docs) $$file site/content/en/news/releases/notes; \
	done

.PHONY: release-notes-gen
release-notes-gen: # Compile release-notes/current/ fragments into release-notes/$(RELEASE_NOTE_VERSION).yaml and clear the fragments
	@$(LOG_TARGET)
	$(eval RELEASE_NOTE_VERSION := $(if $(RELEASE_NOTE_VERSION),$(RELEASE_NOTE_VERSION),$(shell cat VERSION)))
	@echo "Compiling release-notes/current/ fragments into release-notes/$(RELEASE_NOTE_VERSION).yaml"
	@test -n "$(RELEASE_NOTE_DATE)" || (echo "ERROR: RELEASE_NOTE_DATE is required (e.g. \"June 23, 2026\")"; exit 1)
	python3 tools/src/release-notes-docs/compile.py $(RELEASE_NOTE_VERSION) "$(RELEASE_NOTE_DATE)"
