# This is a wrapper to manage helm chart
#
# All make targets related to helms are defined in this file.

include tools/make/env.mk

IMAGE_PULL_POLICY ?= IfNotPresent
OCI_REGISTRY ?= oci://docker.io/envoyproxy
CHART_NAME ?= gateway-helm
ADDONS_CHART_NAME ?= gateway-addons-helm
CHART_VERSION ?= ${RELEASE_VERSION}

##@ Helm
helm-package: ## Package envoy gateway and add-ons helm chart.
helm-package: helm-generate
	@$(LOG_TARGET)
	helm package charts/${CHART_NAME} --app-version ${TAG} --version ${CHART_VERSION} --destination ${OUTPUT_DIR}/charts/
	helm package charts/${ADDONS_CHART_NAME} --app-version ${TAG} --version ${CHART_VERSION} --destination ${OUTPUT_DIR}/charts/

helm-push:
helm-push: ## Push envoy gateway helm chart to OCI registry.
	@$(LOG_TARGET)
	helm push ${OUTPUT_DIR}/charts/${CHART_NAME}-${CHART_VERSION}.tgz ${OCI_REGISTRY}
	helm push ${OUTPUT_DIR}/charts/${ADDONS_CHART_NAME}-${CHART_VERSION}.tgz ${OCI_REGISTRY}

helm-install:
helm-install: helm-generate ## Install envoy gateway helm chart from OCI registry.
	@$(LOG_TARGET)
	helm install eg ${OCI_REGISTRY}/${CHART_NAME} --version ${CHART_VERSION} -n envoy-gateway-system --create-namespace

helm-install-addons:
helm-install-addons: helm-generate ## Install envoy gateway addons helm chart from OCI registry.
	@$(LOG_TARGET)
	helm install eg ${OCI_REGISTRY}/${ADDONS_CHART_NAME} --version ${CHART_VERSION} -n monitoring --create-namespace

.PHONY: helm-generate
helm-generate:
	GatewayImage=${IMAGE}:${TAG} GatewayImagePullPolicy=${IMAGE_PULL_POLICY} envsubst < charts/${CHART_NAME}/values.tmpl.yaml > ./charts/${CHART_NAME}/values.yaml
	helm lint charts/${CHART_NAME}
	helm lint charts/${ADDONS_CHART_NAME}
	helm dependency update charts/${ADDONS_CHART_NAME} # Update dependencies for add-ons chart.

HELM_VALUES := $(wildcard test/helm/*.in.yaml)

helm-template: ## Template envoy gateway helm chart.z
	@$(LOG_TARGET)
	@for file in $(HELM_VALUES); do \
  		filename=$$(basename $${file}); \
  		output="$${filename%.in.*}.out.yaml"; \
		helm template eg charts/${CHART_NAME} -f $${file} > test/helm/$$output --namespace=envoy-gateway-system; \
	done
