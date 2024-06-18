# This is a wrapper to manage helm chart
#
# All make targets related to helms are defined in this file.

include tools/make/env.mk

CHARTS := $(wildcard charts/*)

IMAGE_PULL_POLICY ?= IfNotPresent
OCI_REGISTRY ?= oci://docker.io/envoyproxy
CHART_NAME ?= gateway-helm
CHART_VERSION ?= ${RELEASE_VERSION}
RELEASE_NAMESPACE ?= envoy-gateway-system

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

.PHONY: helm-install
helm-install: ## Install envoy gateway relevant helm charts from OCI registry.
helm-install:
	@for chart in $(CHARTS); do \
		$(LOG_TARGET); \
		$(MAKE) $(addprefix helm-install., $$(basename $${chart})); \
	done

.PHONY: helm-install.%
helm-install.%: helm-generate.%
	$(eval COMMAND := $(word 1,$(subst ., ,$*)))
	$(eval CHART_NAME := $(COMMAND))
	helm install eg ${OCI_REGISTRY}/${CHART_NAME} --version ${CHART_VERSION} -n ${RELEASE_NAMESPACE} --create-namespace

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
  		GatewayImage=${IMAGE}:${TAG} GatewayImagePullPolicy=${IMAGE_PULL_POLICY} \
  		envsubst < charts/${CHART_NAME}/values.tmpl.yaml > ./charts/${CHART_NAME}/values.yaml; \
  	fi
	helm dependency update charts/${CHART_NAME} # Update dependencies for add-ons chart.
	helm lint charts/${CHART_NAME}

HELM_VALUES := $(wildcard test/helm/*.in.yaml)

helm-template: ## Template envoy gateway helm chart.z
	@$(LOG_TARGET)
	@for file in $(HELM_VALUES); do \
  		filename=$$(basename $${file}); \
  		output="$${filename%.in.*}.out.yaml"; \
		helm template eg charts/${CHART_NAME} -f $${file} > test/helm/$$output --namespace=${RELEASE_NAMESPACE}; \
	done
