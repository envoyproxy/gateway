# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION ?= 1.24.1
# GATEWAY_API_VERSION refers to the version of Gateway API CRDs.
# For more details, see https://gateway-api.sigs.k8s.io/guides/getting-started/#installing-gateway-api 
GATEWAY_API_VERSION ?= $(shell go list -m -f '{{.Version}}' sigs.k8s.io/gateway-api)

GATEWAY_RELEASE_URL ?= https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/experimental-install.yaml

CONFORMANCE_UNIQUE_PORTS ?= true

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
manifests: $(tools/controller-gen) generate-gwapi-manifests ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	@$(LOG_TARGET)
	$(tools/controller-gen) rbac:roleName=envoy-gateway-role crd webhook paths="./..." output:crd:artifacts:config=charts/gateway-helm/crds/generated output:rbac:artifacts:config=charts/gateway-helm/templates/generated/rbac output:webhook:artifacts:config=charts/gateway-helm/templates/generated/webhook

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
	KUBEBUILDER_ASSETS="$(shell $(tools/setup-envtest) use $(ENVTEST_K8S_VERSION) -p path)" go test --tags=integration ./... -coverprofile cover.out

##@ Kubernetes Deployment

ifndef ignore-not-found
  ignore-not-found = true
endif

IMAGE_PULL_POLICY ?= Always

.PHONY: kube-deploy
kube-deploy: manifests ## Install Envoy Gateway into the Kubernetes cluster specified in ~/.kube/config.
	@$(LOG_TARGET)
	helm install eg charts/gateway-helm --set deployment.envoyGateway.image.repository=$(IMAGE) --set deployment.envoyGateway.image.tag=$(TAG) --set deployment.envoyGateway.imagePullPolicy=$(IMAGE_PULL_POLICY) -n envoy-gateway-system --create-namespace

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
	kubectl wait --timeout=5m -n gateway-system deployment/gateway-api-admission-server --for=condition=Available
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl wait --timeout=5m -n gateway-system job/gateway-api-admission --for=condition=Complete
	kubectl apply -f test/config/gatewayclass.yaml
	go test -v -tags conformance ./test/conformance --gateway-class=envoy-gateway --debug=true --use-unique-ports=$(CONFORMANCE_UNIQUE_PORTS)

.PHONY: delete-cluster
delete-cluster: $(tools/kind) ## Delete kind cluster.
	@$(LOG_TARGET)
	$(tools/kind) delete cluster --name envoy-gateway

.PHONY: generate-manifests
generate-manifests: ## Generate Kubernetes release manifests.
	@$(LOG_TARGET)
	@$(call log, "Generating kubernetes manifests")
	mkdir -p $(OUTPUT_DIR)/
	helm template eg charts/gateway-helm --include-crds --set deployment.envoyGateway.image.repository=$(IMAGE) --set deployment.envoyGateway.image.tag=$(TAG) --set deployment.envoyGateway.imagePullPolicy=$(IMAGE_PULL_POLICY) > $(OUTPUT_DIR)/install.yaml
	@$(call log, "Added: $(OUTPUT_DIR)/install.yaml")
	cp examples/kubernetes/quickstart.yaml $(OUTPUT_DIR)/quickstart.yaml
	@$(call log, "Added: $(OUTPUT_DIR)/quickstart.yaml")

.PHONY: generate-artifacts
generate-artifacts: generate-manifests ## Generate release artifacts.
	@$(LOG_TARGET)
	cp -r $(ROOT_DIR)/release-notes/$(TAG).yaml $(OUTPUT_DIR)/release-notes.yaml
	@$(call log, "Added: $(OUTPUT_DIR)/release-notes.yaml")
