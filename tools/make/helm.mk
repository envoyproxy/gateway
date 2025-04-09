# This is a wrapper to manage helm chart
#
# All make targets related to helms are defined in this file.

include tools/make/env.mk

CHARTS := $(wildcard charts/*)

IMAGE_PULL_POLICY ?= IfNotPresent
OCI_REGISTRY ?= oci://docker.io/envoyproxy
CHART_NAME ?= gateway-helm
CHART_VERSION ?= ${RELEASE_VERSION}

##@ Helm
.PHONY: helm-package
helm-package: ## Package envoy gateway relevant helm charts.
helm-package:
	@for chart in $(CHARTS); do \
		$(LOG_TARGET); \
		$(MAKE) $(addprefix helm-package., $$(basename $${chart})); \
	done

.PHONY: helm-package.%
helm-package.%: helm-generate.%
	$(eval COMMAND := $(word 1,$(subst ., ,$*)))
	$(eval CHART_NAME := $(COMMAND))
	helm package charts/${CHART_NAME} --app-version ${TAG} --version ${CHART_VERSION} --destination ${OUTPUT_DIR}/charts/

.PHONY: helm-push
helm-push: ## Push envoy gateway relevant helm charts to OCI registry.
helm-push:
	@for chart in $(CHARTS); do \
		$(LOG_TARGET); \
		$(MAKE) $(addprefix helm-push., $$(basename $${chart})); \
	done

.PHONY: helm-push.%
helm-push.%: helm-package.%
	$(eval COMMAND := $(word 1,$(subst ., ,$*)))
	$(eval CHART_NAME := $(COMMAND))
	helm push ${OUTPUT_DIR}/charts/${CHART_NAME}-${CHART_VERSION}.tgz ${OCI_REGISTRY}

.PHONY: helm-generate
helm-generate:
	@for chart in $(CHARTS); do \
		$(LOG_TARGET); \
		$(MAKE) $(addprefix helm-generate., $$(basename $${chart})); \
	done

.PHONY: helm-generate.%
helm-generate.%:
	$(eval COMMAND := $(word 1,$(subst ., ,$*)))
	$(eval CHART_NAME := $(COMMAND))
	@if test -f "charts/${CHART_NAME}/values.tmpl.yaml"; then \
			GatewayImage="${REGISTRY}/${REPOSITORY}:${TAG}"\
			GatewayImagePullPolicy=${IMAGE_PULL_POLICY}\
			RatelimitImage="${REGISTRY}/${RATELIMIT_REPOSITORY}:${RATELIMIT_TAG}"\
			envsubst < charts/${CHART_NAME}/values.tmpl.yaml > ./charts/${CHART_NAME}/values.yaml; \
		fi
	helm dependency update charts/${CHART_NAME}
	helm lint charts/${CHART_NAME}

  # The jb does not support self-assigned jsonnetfile, so entering working dir before executing jb.
	@if [ ${CHART_NAME} == "gateway-addons-helm" ]; then \
			$(call log, "Run jsonnet generate for dashboards in chart: ${CHART_NAME}!"); \
			workDir="charts/${CHART_NAME}/dashboards"; \
			cd $$workDir && go tool jb install && cd ../../..; \
			for file in $$(find $${workDir} -maxdepth 1 -name '*.libsonnet'); do \
				name=$$(basename $$file .libsonnet); \
				go tool jsonnet -J $${workDir}/vendor $${workDir}/$${name}.libsonnet > $${workDir}/$${name}.gen.json; \
		done \
	fi

	$(call log, "Run helm template for chart: ${CHART_NAME}!");
	@for file in $(wildcard test/helm/${CHART_NAME}/*.in.yaml); do \
			filename=$$(basename $${file}); \
			output="$${filename%.in.*}.out.yaml"; \
			if [ ${CHART_NAME} == "gateway-addons-helm" ]; then \
				helm template ${CHART_NAME} charts/${CHART_NAME} -f $${file} > test/helm/${CHART_NAME}/$$output --namespace=monitoring; \
			else \
				helm template ${CHART_NAME} charts/${CHART_NAME} -f $${file} > test/helm/${CHART_NAME}/$$output --namespace=envoy-gateway-system; \
			fi; \
	done
