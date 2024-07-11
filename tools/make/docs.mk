DOCS_OUTPUT_DIR := site/public
RELEASE_VERSIONS ?= $(foreach v,$(wildcard ${ROOT_DIR}/docs/*),$(notdir ${v}))
CLEAN_NODE_MODULES ?= true

##@ Docs

.PHONY: docs
docs: docs.clean helm-readme-gen docs-api ## Generate Envoy Gateway Docs Sources
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
helm-readme-gen.%: $(tools/helm-docs)
	$(eval COMMAND := $(word 1,$(subst ., ,$*)))
	$(eval CHART_NAME := $(COMMAND))
	# use production ENV to generate helm api doc
	@if test -f "charts/${CHART_NAME}/values.tmpl.yaml"; then \
		ImageRepository=docker.io/envoyproxy/gateway ImageTag=latest ImagePullPolicy=IfNotPresent \
		envsubst < charts/${CHART_NAME}/values.tmpl.yaml > ./charts/${CHART_NAME}/values.yaml; \
	fi

	# generate helm readme doc
	$(tools/helm-docs) --template-files=tools/helm-docs/readme.${CHART_NAME}.gotmpl -g charts/${CHART_NAME} -f values.yaml -o README.md

	# change the placeholder to title before api helm docs generated: split by '-' and capitalize the first letters
	$(eval CHART_TITLE := $(shell echo "$(CHART_NAME)" | sed -E 's/\<./\U&/g; s/-/ /g' | awk '{for(i=1;i<=NF;i++){ $$i=toupper(substr($$i,1,1)) substr($$i,2) }}1'))
	sed 's/{CHART-NAME}/$(CHART_TITLE)/g' tools/helm-docs/api.gotmpl > tools/helm-docs/api.${CHART_NAME}.gotmpl
	$(tools/helm-docs) --template-files=tools/helm-docs/api.${CHART_NAME}.gotmpl -g charts/${CHART_NAME} -f values.yaml -o api.md
	mv charts/${CHART_NAME}/api.md site/content/en/latest/install/${CHART_NAME}-api.md
	rm tools/helm-docs/api.${CHART_NAME}.gotmpl

	# below line copy command for sync English api doc into Chinese
	cp site/content/en/latest/install/${CHART_NAME}-api.md site/content/zh/latest/install/${CHART_NAME}-api.md

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
	# below line copy command for sync English api doc into Chinese
	cp site/content/en/latest/api/extension_types.md site/content/zh/latest/api/extension_types.md

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
	# github.com does not allow access too often, there are a lot of 429 errors
	# TODO: find a way to remove github.com from ignore list
	# TODO: example.com is not a valid domain, we should remove it from ignore list
	linkinator site/public/ -r --concurrency 25 --skip "github.com example.com github.io _print v0.6.0 v0.5.0 v0.4.0 v0.3.0 v0.2.0"

release-notes-docs: $(tools/release-notes-docs)
	@$(LOG_TARGET)
	@for file in $(wildcard release-notes/*.yaml); do \
		$(tools/release-notes-docs) $$file site/content/en/news/releases/notes; \
	done
