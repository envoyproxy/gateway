# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION ?= 1.28.0
# Need run cel validation across multiple versions of k8s
ENVTEST_K8S_VERSIONS ?= 1.27.1 1.28.0 1.29.0
# GATEWAY_API_VERSION refers to the version of Gateway API CRDs.
# For more details, see https://gateway-api.sigs.k8s.io/guides/getting-started/#installing-gateway-api
GATEWAY_API_VERSION ?= $(shell go list -m -f '{{.Version}}' sigs.k8s.io/gateway-api)

GATEWAY_RELEASE_URL ?= https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/experimental-install.yaml

WAIT_TIMEOUT ?= 15m

BENCHMARK_TIMEOUT ?= 60m
BENCHMARK_CPU_LIMITS ?= 1000m
BENCHMARK_MEMORY_LIMITS ?= 1024Mi
BENCHMARK_RPS ?= 10000
BENCHMARK_CONNECTIONS ?= 100
BENCHMARK_DURATION ?= 60
BENCHMARK_REPORT_DIR ?= benchmark_report

E2E_RUN_TEST ?=
E2E_CLEANUP ?= true
E2E_TEST_ARGS ?= -v -tags e2e -timeout 15m

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
	$(tools/controller-gen) crd:allowDangerousTypes=true paths="./api/..." output:crd:artifacts:config=charts/gateway-helm/crds/generated

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

.PHONY: kube-deploy
kube-deploy: manifests helm-generate.gateway-helm ## Install Envoy Gateway into the Kubernetes cluster specified in ~/.kube/config.
	@$(LOG_TARGET)
	helm install eg charts/gateway-helm --set deployment.envoyGateway.imagePullPolicy=$(IMAGE_PULL_POLICY) -n envoy-gateway-system --create-namespace --debug --timeout='$(WAIT_TIMEOUT)' --wait --wait-for-jobs

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

.PHONY: e2e
e2e: create-cluster kube-install-image kube-deploy install-ratelimit install-e2e-telemetry run-e2e delete-cluster

.PHONY: install-ratelimit
install-ratelimit:
	@$(LOG_TARGET)
	kubectl apply -f examples/redis/redis.yaml
	kubectl rollout restart deployment envoy-gateway -n envoy-gateway-system
	kubectl rollout status --watch --timeout=5m -n envoy-gateway-system deployment/envoy-gateway
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	tools/hack/deployment-exists.sh "app.kubernetes.io/name=envoy-ratelimit" "envoy-gateway-system"
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-ratelimit --for=condition=Available

.PHONY: e2e-prepare
e2e-prepare: ## Prepare the environment for running e2e tests
	@$(LOG_TARGET)
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-ratelimit --for=condition=Available
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl apply -f test/config/gatewayclass.yaml

.PHONY: run-e2e
run-e2e: e2e-prepare ## Run e2e tests
	@$(LOG_TARGET)
ifeq ($(E2E_RUN_TEST),)
	go test $(E2E_TEST_ARGS) ./test/e2e --gateway-class=envoy-gateway --debug=true --cleanup-base-resources=false
	go test $(E2E_TEST_ARGS) ./test/e2e/merge_gateways --gateway-class=merge-gateways --debug=true --cleanup-base-resources=false
	go test $(E2E_TEST_ARGS) ./test/e2e/multiple_gc --debug=true --cleanup-base-resources=true
	go test $(E2E_TEST_ARGS) ./test/e2e/upgrade --gateway-class=upgrade --debug=true --cleanup-base-resources=$(E2E_CLEANUP)
else
	go test $(E2E_TEST_ARGS) ./test/e2e --gateway-class=envoy-gateway --debug=true --cleanup-base-resources=$(E2E_CLEANUP) \
		--run-test $(E2E_RUN_TEST)
endif

.PHONY: run-benchmark
run-benchmark: install-benchmark-server ## Run benchmark tests
	@$(LOG_TARGET)
	mkdir -p $(OUTPUT_DIR)/benchmark
	kubectl wait --timeout=$(WAIT_TIMEOUT) -n benchmark-test deployment/nighthawk-test-server --for=condition=Available
	kubectl wait --timeout=$(WAIT_TIMEOUT) -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl apply -f test/benchmark/config/gatewayclass.yaml
	go test -v -tags benchmark -timeout $(BENCHMARK_TIMEOUT) ./test/benchmark --rps=$(BENCHMARK_RPS) --connections=$(BENCHMARK_CONNECTIONS) --duration=$(BENCHMARK_DURATION) --report-save-dir=$(BENCHMARK_REPORT_DIR)
	# render benchmark profiles into image
	dot -V
	@for profile in $(wildcard test/benchmark/$(BENCHMARK_REPORT_DIR)/profiles/*.pprof); do \
		$(call log, "Rendering profile image for: $${profile}"); \
		go tool pprof -png $${profile} > $${profile}.png; \
	done

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

.PHONY: install-e2e-telemetry
install-e2e-telemetry: helm-generate.gateway-addons-helm
	@$(LOG_TARGET)
	helm upgrade -i eg-addons charts/gateway-addons-helm --set grafana.enabled=false,opentelemetry-collector.enabled=true -n monitoring --create-namespace --timeout='$(WAIT_TIMEOUT)' --wait --wait-for-jobs
	# Change loki service type from ClusterIP to LoadBalancer
	kubectl patch service loki -n monitoring -p '{"spec": {"type": "LoadBalancer"}}'
	# Wait service Ready
	kubectl rollout status --watch --timeout=5m -n monitoring deployment/prometheus
	kubectl rollout status --watch --timeout=5m statefulset/loki -n monitoring
	kubectl rollout status --watch --timeout=5m statefulset/tempo -n monitoring
	# Restart otel-collector to make sure otlp exporter worked
	kubectl rollout restart -n monitoring deployment/otel-collector
	kubectl rollout status --watch --timeout=5m -n monitoring deployment/otel-collector

.PHONY: uninstall-e2e-telemetry
uninstall-e2e-telemetry:
	@$(LOG_TARGET)
	helm delete $(shell helm list -n monitoring -q) -n monitoring

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
generate-manifests: helm-generate.gateway-helm ## Generate Kubernetes release manifests.
	@$(LOG_TARGET)
	@$(call log, "Generating kubernetes manifests")
	mkdir -p $(OUTPUT_DIR)/
	helm template --set createNamespace=true eg charts/gateway-helm --include-crds --namespace envoy-gateway-system > $(OUTPUT_DIR)/install.yaml
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
