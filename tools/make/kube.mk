# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
# To know the available versions check:
# - https://github.com/kubernetes-sigs/controller-tools/blob/main/envtest-releases.yaml
ENVTEST_K8S_VERSION ?= 1.29.5
# Need run cel validation across multiple versions of k8s
# TODO: zhaohuabing update kubebuilder assets to 1.33.0 when available
ENVTEST_K8S_VERSIONS ?= 1.29.5 1.30.3 1.31.0 1.32.0

# GATEWAY_API_VERSION refers to the version of Gateway API CRDs.
# For more details, see https://gateway-api.sigs.k8s.io/guides/getting-started/#installing-gateway-api
GATEWAY_API_VERSION ?= v1.4.0-rc.1

GATEWAY_API_RELEASE_URL ?= https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}
EXPERIMENTAL_GATEWAY_API_RELEASE_URL ?= ${GATEWAY_API_RELEASE_URL}/experimental-install.yaml
STANDARD_GATEWAY_API_RELEASE_URL ?= ${GATEWAY_API_RELEASE_URL}/standard-install.yaml

WAIT_TIMEOUT ?= 15m

IP_FAMILY ?= ipv4
BENCHMARK_TIMEOUT ?= 60m
BENCHMARK_CPU_LIMITS ?= 1000m
BENCHMARK_MEMORY_LIMITS ?= 1024Mi
BENCHMARK_RPS ?= 10000
BENCHMARK_CONNECTIONS ?= 100
BENCHMARK_DURATION ?= 60
BENCHMARK_REPORT_DIR ?= benchmark_report

# Benchmark toggles
# Disable PNG rendering by default to speed up CI
BENCHMARK_RENDER_PNG ?= false

CONFORMANCE_RUN_TEST ?=

E2E_RUN_TEST ?=
E2E_CLEANUP ?= true
E2E_TIMEOUT ?= 20m
# E2E_REDIRECT allow you specified a redirect when run e2e test locally, e.g. `>> test_output.out 2>&1`
E2E_REDIRECT ?=
E2E_TEST_ARGS ?= -v -tags e2e -timeout $(E2E_TIMEOUT)

DOCKER_MAC_NET_CONNECT ?= true
HOMEBREW_GOPROXY ?=

KUBE_DEPLOY_PROFILE ?= default
KUBE_DEPLOY_HELM_VALUES_FILE = $(ROOT_DIR)/test/config/helm/$(KUBE_DEPLOY_PROFILE).yaml
KUBE_DEPLOY_EG_CONFIG_FILE = $(ROOT_DIR)/test/config/envoy-gateaway-config/$(KUBE_DEPLOY_PROFILE).yaml

# Set Kubernetes Resources Directory Path
ifeq ($(origin KUBE_PROVIDER_DIR),undefined)
KUBE_PROVIDER_DIR := $(ROOT_DIR)/internal/provider/kubernetes/config
endif

# Set Infra Resources Directory Path
ifeq ($(origin KUBE_INFRA_DIR),undefined)
KUBE_INFRA_DIR := $(ROOT_DIR)/internal/infrastructure/kubernetes/config
endif

ifeq ($(IP_FAMILY),ipv4)
ENVOY_PROXY_IP_FAMILY := IPv4
else ifeq ($(IP_FAMILY),ipv6)
ENVOY_PROXY_IP_FAMILY := IPv6
else ifeq ($(IP_FAMILY),dual)
ENVOY_PROXY_IP_FAMILY := DualStack
endif

##@ Kubernetes Development

GNU_SED := $(shell sed --version >/dev/null 2>&1 && echo "yes" || echo "no")

YEAR := $(shell date +%Y)
CONTROLLERGEN_OBJECT_FLAGS :=  object:headerFile="$(ROOT_DIR)/tools/boilerplate/boilerplate.generatego.txt",year=$(YEAR)

.PHONY: prepare-ip-family
prepare-ip-family:
ifeq ($(GNU_SED),yes)
	@find ./test -type f -name "*.yaml" | xargs sed -i'' 's/ipFamily: IPv4/ipFamily: $(ENVOY_PROXY_IP_FAMILY)/g'
else
	@find ./test -type f -name "*.yaml" | xargs sed -i '' 's/ipFamily: IPv4/ipFamily: $(ENVOY_PROXY_IP_FAMILY)/g'
endif

.PHONY: manifests
manifests: generate-gwapi-manifests ## Generate WebhookConfiguration and CustomResourceDefinition objects.
	@$(LOG_TARGET)
	$(GO_TOOL) controller-gen crd:allowDangerousTypes=true paths="./api/..." output:crd:artifacts:config=charts/gateway-helm/crds/generated
	@mkdir -p charts/gateway-helm/templates/generated
	$(GO_TOOL) controller-gen crd:allowDangerousTypes=true paths="./api/..." output:crd:artifacts:config=charts/gateway-crds-helm/templates/generated
	@for file in charts/gateway-crds-helm/templates/generated/*.yaml; do \
		sed -i.bak '1s/^/{{- if .Values.crds.envoyGateway.enabled }}\n/' $$file && \
		echo '{{- end }}' >> $$file && \
		rm -f $$file.bak; \
	done

.PHONY: generate-gwapi-manifests
generate-gwapi-manifests: ## Generate Gateway API manifests and make it consistent with the go mod version.
	@$(LOG_TARGET)
	@echo "Generating Gateway API CRDs"
	@mkdir -p $(OUTPUT_DIR)/
	@curl -sLo $(OUTPUT_DIR)/experimental-gatewayapi-crds.yaml ${EXPERIMENTAL_GATEWAY_API_RELEASE_URL}
	@curl -sLo $(OUTPUT_DIR)/standard-gatewayapi-crds.yaml ${STANDARD_GATEWAY_API_RELEASE_URL}
	cp $(OUTPUT_DIR)/experimental-gatewayapi-crds.yaml charts/gateway-helm/crds/gatewayapi-crds.yaml
	@sed -i.bak '1s/^/{{- if and .Values.crds.gatewayAPI.enabled (eq .Values.crds.gatewayAPI.channel "standard") }}\n/' $(OUTPUT_DIR)/standard-gatewayapi-crds.yaml && \
	echo '{{- end }}' >> $(OUTPUT_DIR)/standard-gatewayapi-crds.yaml && \
	sed -i.bak '1s/^/{{- if and .Values.crds.gatewayAPI.enabled (or (eq .Values.crds.gatewayAPI.channel "experimental") (eq .Values.crds.gatewayAPI.channel "")) }}\n/' $(OUTPUT_DIR)/experimental-gatewayapi-crds.yaml && \
	echo '{{- end }}' >> $(OUTPUT_DIR)/experimental-gatewayapi-crds.yaml && \
	rm -f $(OUTPUT_DIR)/standard-gatewayapi-crds.yaml.bak && \
	rm -f $(OUTPUT_DIR)/experimental-gatewayapi-crds.yaml.bak
	@mv $(OUTPUT_DIR)/experimental-gatewayapi-crds.yaml charts/gateway-crds-helm/templates/experimental-gatewayapi-crds.yaml
	@mv $(OUTPUT_DIR)/standard-gatewayapi-crds.yaml charts/gateway-crds-helm/templates/standard-gatewayapi-crds.yaml

.PHONY: kube-generate
kube-generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
# Note that the paths can't just be "./..." with the header file, or the tool will panic on run. Sorry.
	@$(LOG_TARGET)
	$(GO_TOOL) controller-gen $(CONTROLLERGEN_OBJECT_FLAGS) paths="{$(ROOT_DIR)/api/...,$(ROOT_DIR)/internal/ir/...,$(ROOT_DIR)/internal/gatewayapi/...}"

.PHONY: kube-test
kube-test: manifests generate ## Run Kubernetes provider tests.
	@$(LOG_TARGET)
	KUBEBUILDER_ASSETS="$(shell $(GO_TOOL) setup-envtest use $(ENVTEST_K8S_VERSION) -p path)" go test --tags=integration,celvalidation ./... -coverprofile cover.out

##@ Kubernetes Deployment

ifndef ignore-not-found
  ignore-not-found = true
endif

.PHONY: kube-deploy
kube-deploy: manifests helm-generate.gateway-helm ## Install Envoy Gateway into the Kubernetes cluster specified in ~/.kube/config.
	@$(LOG_TARGET)
	helm install eg charts/gateway-helm \
		--set deployment.envoyGateway.imagePullPolicy=$(IMAGE_PULL_POLICY) \
		-n envoy-gateway-system --create-namespace \
		--debug --timeout='$(WAIT_TIMEOUT)' \
		--wait --wait-for-jobs \
		-f $(KUBE_DEPLOY_HELM_VALUES_FILE)

.PHONY: kube-deploy-for-benchmark-test
kube-deploy-for-benchmark-test: manifests helm-generate ## Install Envoy Gateway and prometheus-server for benchmark test purpose only.
	@$(LOG_TARGET)
	# Install Envoy Gateway
	helm install eg charts/gateway-helm --set deployment.envoyGateway.imagePullPolicy=$(IMAGE_PULL_POLICY) \
		--set deployment.envoyGateway.resources.limits.cpu=$(BENCHMARK_CPU_LIMITS) \
		--set deployment.envoyGateway.resources.limits.memory=$(BENCHMARK_MEMORY_LIMITS) \
		--set config.envoyGateway.admin.enablePprof=true \
		-n envoy-gateway-system --create-namespace --debug --timeout='$(WAIT_TIMEOUT)' --wait --wait-for-jobs
	# Install Prometheus-server only
	helm install eg-addons charts/gateway-addons-helm --set loki.enabled=false \
		--set tempo.enabled=false \
		--set grafana.enabled=false \
		--set fluent-bit.enabled=false \
 		--set opentelemetry-collector.enabled=false \
 		--set prometheus.enabled=true \
 		-n monitoring --create-namespace --debug --timeout='$(WAIT_TIMEOUT)' --wait --wait-for-jobs

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

.PHONY: benchmark
benchmark: create-cluster kube-install-image kube-deploy-for-benchmark-test run-benchmark delete-cluster ## Create a kind cluster, deploy EG into it, run Envoy Gateway benchmark test, and clean up.

.PHONY: resilience
resilience: create-cluster kube-install-image kube-install-examples-image kube-deploy install-eg-addons enable-simple-extension-server run-resilience delete-cluster ## Create a kind cluster, deploy EG into it, run Envoy Gateway resilience test, and clean up.

.PHONY: e2e
e2e: create-cluster kube-install-image kube-deploy \
	install-ratelimit install-eg-addons kube-install-examples-image \
	e2e-prepare setup-mac-net-connect run-e2e delete-cluster

.PHONY: install-ratelimit
install-ratelimit:
	@$(LOG_TARGET)
	kubectl apply -f examples/redis/redis.yaml
	tools/hack/deployment-exists.sh "app.kubernetes.io/name=envoy-ratelimit" "envoy-gateway-system"
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-ratelimit --for=condition=Available

.PHONY: enable-simple-extension-server
enable-simple-extension-server:
	@$(LOG_TARGET)
	kubectl apply -f examples/simple-extension-server/simple-extension-server.yaml
	tools/hack/deployment-exists.sh "app.kubernetes.io/name=gateway-simple-extension-server" "envoy-gateway-system"
	kubectl rollout status --watch --timeout=5m -n envoy-gateway-system deployment/envoy-gateway
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available

.PHONY: e2e-prepare
e2e-prepare: prepare-ip-family ## Prepare the environment for running e2e tests
	@$(LOG_TARGET)
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-ratelimit --for=condition=Available
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl apply -f $(KUBE_DEPLOY_EG_CONFIG_FILE)
	kubectl apply -f test/config/gatewayclass.yaml

.PHONY: setup-mac-net-connect
setup-mac-net-connect:
	@$(LOG_TARGET)
	DOCKER_MAC_NET_CONNECT=$(DOCKER_MAC_NET_CONNECT) HOMEBREW_GOPROXY=$(HOMEBREW_GOPROXY) tools/hack/manage-mac-net-connect.sh setup

.PHONY: run-e2e
run-e2e: ## Run e2e tests
	@$(LOG_TARGET)
ifeq ($(E2E_RUN_TEST),)
	go test $(E2E_TEST_ARGS) ./test/e2e --gateway-class=envoy-gateway --debug=true --cleanup-base-resources=false $(E2E_REDIRECT)
	go test $(E2E_TEST_ARGS) ./test/e2e/merge_gateways --gateway-class=merge-gateways --debug=true --cleanup-base-resources=false
	go test $(E2E_TEST_ARGS) ./test/e2e/multiple_gc --debug=true --cleanup-base-resources=true
	LAST_VERSION_TAG=$(shell cat VERSION) go test $(E2E_TEST_ARGS) ./test/e2e/upgrade --gateway-class=upgrade --debug=true --cleanup-base-resources=$(E2E_CLEANUP)
else
	go test $(E2E_TEST_ARGS) ./test/e2e --gateway-class=envoy-gateway --debug=true --cleanup-base-resources=$(E2E_CLEANUP) \
		--run-test $(E2E_RUN_TEST) $(E2E_REDIRECT)
endif

run-e2e-upgrade:
	go test $(E2E_TEST_ARGS) ./test/e2e/upgrade --gateway-class=upgrade --debug=true --cleanup-base-resources=$(E2E_CLEANUP)

.PHONY: run-resilience
run-resilience: ## Run resilience tests
	@$(LOG_TARGET)
	go test -v -tags resilience ./test/resilience --gateway-class=envoy-gateway

.PHONY: run-benchmark
run-benchmark: install-benchmark-server prepare-ip-family ## Run benchmark tests
	@$(LOG_TARGET)
	mkdir -p $(OUTPUT_DIR)/benchmark
	kubectl wait --timeout=$(WAIT_TIMEOUT) -n benchmark-test deployment/nighthawk-test-server --for=condition=Available
	kubectl wait --timeout=$(WAIT_TIMEOUT) -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl apply -f test/benchmark/config/gatewayclass.yaml
	go test -v -tags benchmark -timeout $(BENCHMARK_TIMEOUT) ./test/benchmark --rps=$(BENCHMARK_RPS) --connections=$(BENCHMARK_CONNECTIONS) --duration=$(BENCHMARK_DURATION) --report-save-dir=$(BENCHMARK_REPORT_DIR)
	# render benchmark profiles into image
	@if [ "$(BENCHMARK_RENDER_PNG)" != "false" ]; then dot -V; fi
	@if [ "$(BENCHMARK_RENDER_PNG)" != "false" ]; then find test/benchmark/$(BENCHMARK_REPORT_DIR)/profiles -name "*.pprof" -type f -exec sh -c 'go tool pprof -png "$$1" > "$$${1%.pprof}.png"' _ {} \; ; fi

.PHONY: install-benchmark-server
install-benchmark-server: ## Install nighthawk server for benchmark test
	@$(LOG_TARGET)
	kubectl create namespace benchmark-test
	kubectl -n benchmark-test create configmap test-server-config --from-file=test/benchmark/config/nighthawk-test-server-config.yaml -o yaml
	kubectl apply -f test/benchmark/config/nighthawk-test-server.yaml

.PHONY: uninstall-benchmark-server
uninstall-benchmark-server: ## Uninstall nighthawk server for benchmark test
	@$(LOG_TARGET)
	kubectl delete job -n benchmark-test -l benchmark-test/client=true
	kubectl delete -f test/benchmark/config/nighthawk-test-server.yaml
	kubectl delete configmap test-server-config -n benchmark-test
	kubectl delete namespace benchmark-test

.PHONY: install-eg-addons
install-eg-addons: helm-generate.gateway-addons-helm
	@$(LOG_TARGET)
	helm upgrade -i eg-addons charts/gateway-addons-helm -f test/helm/gateway-addons-helm/e2e.in.yaml -n monitoring --create-namespace --timeout='$(WAIT_TIMEOUT)' --wait --wait-for-jobs
	# Change loki service type from ClusterIP to LoadBalancer
	kubectl patch service loki -n monitoring -p '{"spec": {"type": "LoadBalancer"}}'
	# Wait service Ready
	kubectl rollout status --watch --timeout=5m -n monitoring deployment/prometheus
	kubectl rollout status --watch --timeout=5m statefulset/loki -n monitoring
	kubectl rollout status --watch --timeout=5m statefulset/tempo -n monitoring
	# Restart otel-collector to make sure otlp exporter worked
	kubectl rollout restart -n monitoring deployment/otel-collector
	kubectl rollout status --watch --timeout=5m -n monitoring deployment/otel-collector

.PHONY: uninstall-eg-addons
uninstall-eg-addons:
	@$(LOG_TARGET)
	helm delete $(shell helm list -n monitoring -q) -n monitoring

.PHONY: create-cluster
create-cluster: ## Create a kind cluster suitable for running Gateway API conformance.
	@$(LOG_TARGET)
	tools/hack/create-cluster.sh

.PHONY: kube-install-image
kube-install-image: image.build ## Install the EG image to a kind cluster using the provided $IMAGE and $TAG.
	@$(LOG_TARGET)
	tools/hack/kind-load-image.sh $(IMAGE) $(TAG)

.PHONY: run-conformance
run-conformance: prepare-ip-family ## Run Gateway API conformance.
	@$(LOG_TARGET)
	kubectl wait --timeout=$(WAIT_TIMEOUT) -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl apply -f test/config/gatewayclass.yaml
ifeq ($(CONFORMANCE_RUN_TEST),)
	go test -v -tags conformance ./test/conformance --gateway-class=envoy-gateway --debug=true
else
	go test -v -tags conformance ./test/conformance --gateway-class=envoy-gateway --debug=true --run-test $(CONFORMANCE_RUN_TEST)
endif

CONFORMANCE_REPORT_PATH ?=

.PHONY: run-experimental-conformance
run-experimental-conformance: prepare-ip-family ## Run Experimental Gateway API conformance.
	@$(LOG_TARGET)
	kubectl wait --timeout=$(WAIT_TIMEOUT) -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl apply -f test/config/gatewayclass.yaml
	go test -v -tags experimental ./test/conformance -run TestExperimentalConformance --gateway-class=envoy-gateway --debug=true \
		--organization=envoyproxy --project=envoy-gateway --url=https://github.com/envoyproxy/gateway --version=latest \
		--report-output="$(CONFORMANCE_REPORT_PATH)" --contact=https://github.com/envoyproxy/gateway/blob/main/GOVERNANCE.md \
		--mode="$(KUBE_DEPLOY_PROFILE)" --version=$(TAG)

.PHONY: delete-cluster
delete-cluster: ## Delete kind cluster.
	@$(LOG_TARGET)
	$(GO_TOOL) kind delete cluster --name envoy-gateway

.PHONY: generate-manifests
generate-manifests: helm-generate.gateway-helm ## Generate Kubernetes release manifests.
	@$(LOG_TARGET)
	@$(call log, "Generating kubernetes manifests")
	mkdir -p $(OUTPUT_DIR)/
	helm template --set createNamespace=true eg charts/gateway-helm --include-crds --namespace envoy-gateway-system > $(OUTPUT_DIR)/install.yaml
	@$(call log, "Added: $(OUTPUT_DIR)/install.yaml")
	cp examples/kubernetes/quickstart.yaml $(OUTPUT_DIR)/quickstart.yaml
	@$(call log, "Added: $(OUTPUT_DIR)/quickstart.yaml")
	cat charts/gateway-helm/crds/generated/* >> $(OUTPUT_DIR)/envoy-gateway-crds.yaml
	@$(call log, "Added: $(OUTPUT_DIR)/envoy-gateway-crds.yaml")

.PHONY: generate-artifacts
generate-artifacts: generate-manifests ## Generate release artifacts.
	@$(LOG_TARGET)
	cp -r $(ROOT_DIR)/release-notes/$(TAG).yaml $(OUTPUT_DIR)/release-notes.yaml
	@$(call log, "Added: $(OUTPUT_DIR)/release-notes.yaml")
