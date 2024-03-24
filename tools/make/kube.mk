# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION ?= 1.27.1
# GATEWAY_API_VERSION refers to the version of Gateway API CRDs.
# For more details, see https://gateway-api.sigs.k8s.io/guides/getting-started/#installing-gateway-api 
GATEWAY_API_VERSION ?= $(shell go list -m -f '{{.Version}}' sigs.k8s.io/gateway-api)

GATEWAY_RELEASE_URL ?= https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/experimental-install.yaml

WAIT_TIMEOUT ?= 15m

FLUENT_BIT_CHART_VERSION ?= 0.30.4
OTEL_COLLECTOR_CHART_VERSION ?= 0.73.1
TEMPO_CHART_VERSION ?= 1.3.1
E2E_RUN_TEST ?=
E2E_RUN_EG_UPGRADE_TESTS ?= false
E2E_CLEANUP ?= true

# Set Kubernetes Resources Directory Path
ifeq ($(origin KUBE_PROVIDER_DIR),undefined)
KUBE_PROVIDER_DIR := $(ROOT_DIR)/internal/provider/kubernetes/config
endif

# Set Infra Resources Directory Path
ifeq ($(origin KUBE_INFRA_DIR),undefined)
KUBE_INFRA_DIR := $(ROOT_DIR)/internal/infrastructure/kubernetes/config
endif

##@ Kubernetes Development

YEAR := $(shell date +%Y)
CONTROLLERGEN_OBJECT_FLAGS :=  object:headerFile="$(ROOT_DIR)/tools/boilerplate/boilerplate.generatego.txt",year=$(YEAR)

.PHONY: manifests
manifests: $(tools/controller-gen) generate-gwapi-manifests ## Generate WebhookConfiguration and CustomResourceDefinition objects.

	@$(LOG_TARGET)
	$(tools/controller-gen) crd:allowDangerousTypes=true paths="./..." output:crd:artifacts:config=charts/gateway-helm/crds/generated
.PHONY: generate-gwapi-manifests
generate-gwapi-manifests:
generate-gwapi-manifests: ## Generate GWAPI manifests and make it consistent with the go mod version.
	@$(LOG_TARGET)
	@mkdir -p $(OUTPUT_DIR)/
	curl -sLo $(OUTPUT_DIR)/gatewayapi-crds.yaml ${GATEWAY_RELEASE_URL}
	mv $(OUTPUT_DIR)/gatewayapi-crds.yaml charts/gateway-helm/crds/gatewayapi-crds.yaml

.PHONY: kube-generate
kube-generate: $(tools/controller-gen) ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
# Note that the paths can't just be "./..." with the header file, or the tool will panic on run. Sorry.
	@$(LOG_TARGET)
	$(tools/controller-gen) $(CONTROLLERGEN_OBJECT_FLAGS) paths="{$(ROOT_DIR)/api/...,$(ROOT_DIR)/internal/ir/...,$(ROOT_DIR)/internal/gatewayapi/...}"

.PHONY: kube-test
kube-test: manifests generate $(tools/setup-envtest) ## Run Kubernetes provider tests.
	@$(LOG_TARGET)
	KUBEBUILDER_ASSETS="$(shell $(tools/setup-envtest) use $(ENVTEST_K8S_VERSION) -p path)" go test --tags=integration,celvalidation ./... -coverprofile cover.out

##@ Kubernetes Deployment

ifndef ignore-not-found
  ignore-not-found = true
endif

IMAGE_PULL_POLICY ?= Always

.PHONY: kube-deploy
kube-deploy: manifests helm-generate ## Install Envoy Gateway into the Kubernetes cluster specified in ~/.kube/config.
	@$(LOG_TARGET)
	helm install eg charts/gateway-helm --set deployment.envoyGateway.imagePullPolicy=$(IMAGE_PULL_POLICY) -n envoy-gateway-system --create-namespace --debug --timeout='$(WAIT_TIMEOUT)' --wait --wait-for-jobs

.PHONY: kube-undeploy
kube-undeploy: manifests ## Uninstall the Envoy Gateway into the Kubernetes cluster specified in ~/.kube/config.
	@$(LOG_TARGET)
	helm uninstall eg -n envoy-gateway-system

.PHONY: kube-demo-prepare
kube-demo-prepare:
	@$(LOG_TARGET)
	kubectl apply -f examples/kubernetes/quickstart.yaml -n default
	kubectl wait --timeout=5m -n default gateway eg --for=condition=Programmed

.PHONY: kube-demo
kube-demo: kube-demo-prepare ## Deploy a demo backend service, gatewayclass, gateway and httproute resource and test the configuration.
	@$(LOG_TARGET)
	$(eval ENVOY_SERVICE := $(shell kubectl get service -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}'))
	@echo -e "\nPort forward to the Envoy service using the command below"
	@echo -e "kubectl -n envoy-gateway-system port-forward service/$(ENVOY_SERVICE) 8888:80 &"
	@echo -e "\nCurl the app through Envoy proxy using the command below"
	@echo -e "curl --verbose --header \"Host: www.example.com\" http://localhost:8888/get\n"

.PHONY: kube-demo-undeploy
kube-demo-undeploy: ## Uninstall the Kubernetes resources installed from the `make kube-demo` command.
	@$(LOG_TARGET)
	kubectl delete -f examples/kubernetes/quickstart.yaml --ignore-not-found=$(ignore-not-found) -n default

# Uncomment when https://github.com/envoyproxy/gateway/issues/256 is fixed.
#.PHONY: run-kube-local
#run-kube-local: build kube-install ## Run Envoy Gateway locally.
#	tools/hack/run-kube-local.sh

.PHONY: conformance
conformance: create-cluster kube-install-image kube-deploy run-conformance delete-cluster ## Create a kind cluster, deploy EG into it, run Gateway API conformance, and clean up.

.PHONY: experimental-conformance ## Create a kind cluster, deploy EG into it, run Gateway API experimental conformance, and clean up.
experimental-conformance: create-cluster kube-install-image kube-deploy run-experimental-conformance delete-cluster ## Create a kind cluster, deploy EG into it, run Gateway API conformance, and clean up.

.PHONY: e2e
e2e: create-cluster kube-install-image kube-deploy install-ratelimit run-e2e delete-cluster

.PHONY: install-ratelimit
install-ratelimit:
	@$(LOG_TARGET)
	kubectl apply -f examples/redis/redis.yaml
	kubectl rollout restart deployment envoy-gateway -n envoy-gateway-system
	kubectl rollout status --watch --timeout=5m -n envoy-gateway-system deployment/envoy-gateway
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-ratelimit --for=condition=Available

.PHONY: run-e2e
run-e2e: install-e2e-telemetry
	@$(LOG_TARGET)
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-ratelimit --for=condition=Available
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl apply -f test/config/gatewayclass.yaml
ifeq ($(E2E_RUN_TEST),)
	go test -v -tags e2e ./test/e2e --gateway-class=envoy-gateway --debug=true --cleanup-base-resources=false
	go test -v -tags e2e ./test/e2e/upgrade --gateway-class=upgrade --debug=true --cleanup-base-resources=$(E2E_CLEANUP)
else
ifeq ($(E2E_RUN_EG_UPGRADE_TESTS),false)
	go test -v -tags e2e ./test/e2e --gateway-class=envoy-gateway --debug=true --cleanup-base-resources=$(E2E_CLEANUP) \
		--run-test $(E2E_RUN_TEST)
else
	go test -v -tags e2e ./test/e2e/upgrade --gateway-class=upgrade --debug=true --cleanup-base-resources=$(E2E_CLEANUP) \
		--run-test $(E2E_RUN_TEST)
endif
endif
.PHONY: install-e2e-telemetry
install-e2e-telemetry: prepare-helm-repo install-fluent-bit install-loki install-tempo install-otel-collector install-prometheus
	@$(LOG_TARGET)
	kubectl rollout status daemonset fluent-bit -n monitoring --timeout 5m
	kubectl rollout status statefulset loki -n monitoring --timeout 5m
	kubectl rollout status statefulset tempo -n monitoring --timeout 5m
	kubectl rollout status deployment otel-collector -n monitoring --timeout 5m
	kubectl rollout status deployment prometheus -n monitoring --timeout 5m

.PHONY: uninstall-e2e-telemetry
uninstall-e2e-telemetry:
	@$(LOG_TARGET)
	kubectl delete -f examples/loki/loki.yaml -n monitoring --ignore-not-found
	helm delete $(shell helm list -n monitoring -q) -n monitoring

.PHONY: prepare-helm-repo
prepare-helm-repo:
	@$(LOG_TARGET)
	helm repo add fluent https://fluent.github.io/helm-charts
	helm repo add grafana https://grafana.github.io/helm-charts
	helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo update

.PHONY: install-fluent-bit
install-fluent-bit:
	@$(LOG_TARGET)
	helm upgrade --install fluent-bit fluent/fluent-bit -f examples/fluent-bit/helm-values.yaml -n monitoring --create-namespace --version $(FLUENT_BIT_CHART_VERSION)

.PHONY: install-loki
install-loki:
	@$(LOG_TARGET)
	kubectl apply -f examples/loki/loki.yaml -n monitoring

.PHONY: install-tempo
install-tempo:
	@$(LOG_TARGET)
	helm upgrade --install tempo grafana/tempo -f examples/tempo/helm-values.yaml -n monitoring --create-namespace --version $(TEMPO_CHART_VERSION)

.PHONY: install-prometheus
install-prometheus:
	@$(LOG_TARGET)
	helm upgrade --install prometheus prometheus-community/prometheus -f examples/prometheus/helm-values.yaml -n monitoring --create-namespace

.PHONY: install-otel-collector
install-otel-collector:
	@$(LOG_TARGET)
	helm upgrade --install otel-collector open-telemetry/opentelemetry-collector -f examples/otel-collector/helm-values.yaml -n monitoring --create-namespace --version $(OTEL_COLLECTOR_CHART_VERSION)

.PHONY: create-cluster
create-cluster: $(tools/kind) ## Create a kind cluster suitable for running Gateway API conformance.
	@$(LOG_TARGET)
	tools/hack/create-cluster.sh

.PHONY: kube-install-image
kube-install-image: image.build $(tools/kind) ## Install the EG image to a kind cluster using the provided $IMAGE and $TAG.
	@$(LOG_TARGET)
	tools/hack/kind-load-image.sh $(IMAGE) $(TAG)

.PHONY: run-conformance
run-conformance: ## Run Gateway API conformance.
	@$(LOG_TARGET)
	kubectl wait --timeout=$(WAIT_TIMEOUT) -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl apply -f test/config/gatewayclass.yaml
	go test -v -tags conformance ./test/conformance --gateway-class=envoy-gateway --debug=true

CONFORMANCE_REPORT_PATH ?= 

.PHONY: run-experimental-conformance
run-experimental-conformance: ## Run Experimental Gateway API conformance.
	@$(LOG_TARGET)
	kubectl wait --timeout=$(WAIT_TIMEOUT) -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl apply -f test/config/gatewayclass.yaml
	go test -v -tags experimental ./test/conformance -run TestExperimentalConformance --gateway-class=envoy-gateway --debug=true --organization=envoyproxy --project=envoy-gateway --url=https://github.com/envoyproxy/gateway --version=latest --report-output="$(CONFORMANCE_REPORT_PATH)" --contact=https://github.com/envoyproxy/gateway/blob/main/GOVERNANCE.md

.PHONY: delete-cluster
delete-cluster: $(tools/kind) ## Delete kind cluster.
	@$(LOG_TARGET)
	$(tools/kind) delete cluster --name envoy-gateway

.PHONY: generate-manifests
generate-manifests: helm-generate ## Generate Kubernetes release manifests.
	@$(LOG_TARGET)
	@$(call log, "Generating kubernetes manifests")
	mkdir -p $(OUTPUT_DIR)/
	helm template --set createNamespace=true eg charts/gateway-helm --include-crds --set deployment.envoyGateway.imagePullPolicy=$(IMAGE_PULL_POLICY) --namespace envoy-gateway-system > $(OUTPUT_DIR)/install.yaml
	@$(call log, "Added: $(OUTPUT_DIR)/install.yaml")
	cp examples/kubernetes/quickstart.yaml $(OUTPUT_DIR)/quickstart.yaml
	@$(call log, "Added: $(OUTPUT_DIR)/quickstart.yaml")

.PHONY: generate-artifacts
generate-artifacts: generate-manifests generate-egctl-releases ## Generate release artifacts.
	@$(LOG_TARGET)
	cp -r $(ROOT_DIR)/release-notes/$(TAG).yaml $(OUTPUT_DIR)/release-notes.yaml
	@$(call log, "Added: $(OUTPUT_DIR)/release-notes.yaml")

.PHONY: generate-egctl-releases
generate-egctl-releases: ## Generate egctl releases
	@$(LOG_TARGET)
	mkdir -p $(OUTPUT_DIR)/
	curl -sSL https://github.com/envoyproxy/gateway/releases/download/latest/egctl_latest_darwin_amd64.tar.gz -o $(OUTPUT_DIR)/egctl_$(TAG)_darwin_amd64.tar.gz
	curl -sSL https://github.com/envoyproxy/gateway/releases/download/latest/egctl_latest_darwin_arm64.tar.gz -o $(OUTPUT_DIR)/egctl_$(TAG)_darwin_arm64.tar.gz
	curl -sSL https://github.com/envoyproxy/gateway/releases/download/latest/egctl_latest_linux_amd64.tar.gz -o $(OUTPUT_DIR)/egctl_$(TAG)_linux_amd64.tar.gz
	curl -sSL https://github.com/envoyproxy/gateway/releases/download/latest/egctl_latest_linux_arm64.tar.gz -o $(OUTPUT_DIR)/egctl_$(TAG)_linux_arm64.tar.gz
