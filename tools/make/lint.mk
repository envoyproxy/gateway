# This is a wrapper to do lint checks
#
# All make targets related to lint are defined in this file.

##@ Lint

.PHONY: lint
lint: ## Run all linter of code sources, including golint, yamllint, whitenoise lint and codespell.
lint: lint.golint lint.yamllint lint.codespell lint.whitenoise

.PHONY: lint.golint
lint.golint:
	@echo Running Go linter ...
	golangci-lint run --build-tags=e2e --config=tools/linter/golangci-lint/.golangci.yml

.PHONY: lint.yamllint
lint.yamllint:
	@echo Running YAML linter ...
	yamllint --config-file=tools/linter/yamllint/.yamllint changelogs/

.PHONY: lint.codespell
lint.codespell: CODESPELL_SKIP := $(shell cat tools/linter/codespell/.codespell.skip | tr \\n ',')
lint.codespell:
	@echo Running Codespell linter ...
	codespell --skip $(CODESPELL_SKIP) --ignore-words tools/linter/codespell/.codespell.ignorewords --check-filenames --check-hidden -q2

.PHONY: lint.whitenoise
lint.whitenoise:
	@echo Running WhiteNoise linter ...
	tools/linter/lint-whitenoise
