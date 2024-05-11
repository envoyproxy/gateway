# This is a wrapper to manage helm chart
#
# All make targets related to helmß are defined in this file.

include tools/make/env.mk

IMAGE_PULL_POLICY ?= IfNotPresent
OCI_REGISTRY ?= oci://docker.io/envoyproxy
CHART_NAME ?= gateway-helm
CHART_VERSION ?= ${RELEASE_VERSION}

##@ Helm
helm-package:
helm-package: ## Package envoy gateway helm chart.
helm-package: helm-generate
	@$(LOG_TARGET)
	helm package charts/${CHART_NAME} --app-version ${TAG} --version ${CHART_VERSION} --destination ${OUTPUT_DIR}/charts/

helm-push:
helm-push: ## Push envoy gateway helm chart to OCI registry.
	@$(LOG_TARGET)
	helm push ${OUTPUT_DIR}/charts/${CHART_NAME}-${CHART_VERSION}.tgz ${OCI_REGISTRY}

helm-install:
helm-install: helm-generate ## Install envoy gateway helm chart from OCI registry.
	@$(LOG_TARGET)
	helm install eg ${OCI_REGISTRY}/${CHART_NAME} --version ${CHART_VERSION} -n envoy-gateway-system --create-namespace

.PHONY: helm-generate
helm-generate:
	ImageRepository=${IMAGE} ImageTag=${TAG} ImagePullPolicy=${IMAGE_PULL_POLICY} envsubst < charts/gateway-helm/values.tmpl.yaml > ./charts/gateway-helm/values.yaml
	helm lint charts/gateway-helm

helm-template: ## Template envoy gateway helm chart.
	@$(LOG_TARGET)
	helm template eg charts/gateway-helm --set deployment.envoyGateway.image.tag=latest --set config.envoyGateway.provider.kubernetes.shutdownManager.image="docker.io/envoyproxy/gateway-dev:latest" > ./test/helm/default.yaml --namespace=envoy-gateway-system
