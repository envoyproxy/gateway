# This is a wrapper to manage helm chart
#
# All make targets related to helmß are defined in this file.

OCI_REGISTRY ?= oci://geocomply.jfrog.io/apps-idp-gateway-docker-local/envoyproxy/gateway
CHART_NAME ?= gateway-helm
CHART_VERSION ?= ${RELEASE_VERSION}

##@ Helm
helm-package:
helm-package: ## Package envoy gateway helm chart.
	@$(LOG_TARGET)
	helm package charts/${CHART_NAME} --app-version ${TAG} --version ${CHART_VERSION} --destination ${OUTPUT_DIR}/charts/

helm-push:
helm-push: ## Push envoy gateway helm chart to OCI registry.
	@$(LOG_TARGET)
	helm push ${OUTPUT_DIR}/charts/${CHART_NAME}-${CHART_VERSION}.tgz ${OCI_REGISTRY}

helm-install:
helm-install: ## Install envoy gateway helm chart from OCI registry.
	@$(LOG_TARGET)
	helm install eg ${OCI_REGISTRY}/${CHART_NAME} --version ${CHART_VERSION} -n envoy-gateway-system --create-namespace
