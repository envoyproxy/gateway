# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION ?= 1.24.1
# GATEWAY_API_VERSION refers to the version of Gateway API CRDs.
# For more details, see https://gateway-api.sigs.k8s.io/guides/getting-started/#installing-gateway-api 
GATEWAY_API_VERSION ?= $(shell go list -m -f '{{.Version}}' sigs.k8s.io/gateway-api)

# Set Kubernetes Resources Directory Path
ifeq ($(origin KUBE_PROVIDER_DIR),undefined)
KUBE_PROVIDER_DIR := $(ROOT_DIR)/internal/provider/kubernetes/config
endif

# Set Infra Resources Directory Path
ifeq ($(origin KUBE_INFRA_DIR),undefined)
KUBE_INFRA_DIR := $(ROOT_DIR)/internal/infrastructure/kubernetes/config
endif

##@ Kubernetes Development

.PHONY: manifests
manifests: $(tools/controller-gen) ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(tools/controller-gen) rbac:roleName=envoy-gateway-role crd webhook paths="./..." output:crd:artifacts:config=internal/provider/kubernetes/config/crd/bases output:rbac:artifacts:config=internal/provider/kubernetes/config/rbac

.PHONY: generate
generate: $(tools/controller-gen) ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(tools/controller-gen) object paths="./..."

.PHONY: kube-test
kube-test: manifests generate $(tools/setup-envtest) ## Run Kubernetes provider tests.
	KUBEBUILDER_ASSETS="$(shell $(tools/setup-envtest) use $(ENVTEST_K8S_VERSION) -p path)" go test --tags=integration ./... -coverprofile cover.out

##@ Kubernetes Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: kube-install
kube-install: manifests $(tools/kustomize) ## Install Envoy Gateway CRDs into the Kubernetes cluster specified in ~/.kube/config.
	mkdir -pv $(OUTPUT_DIR)/manifests/provider
	cp -r $(KUBE_PROVIDER_DIR) $(OUTPUT_DIR)/manifests/provider
	mkdir -pv $(OUTPUT_DIR)/manifests/infra
	cp -r $(KUBE_INFRA_DIR) $(OUTPUT_DIR)/manifests/infra
	$(tools/kustomize) build $(OUTPUT_DIR)/manifests/provider/config/crd | kubectl apply -f -
	kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/experimental-install.yaml

.PHONY: kube-uninstall
kube-uninstall: manifests $(tools/kustomize) ## Uninstall Envoy Gateway CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/experimental-install.yaml

.PHONY: kube-deploy
kube-deploy: kube-install ## Install Envoy Gateway controller into the Kubernetes cluster specified in ~/.kube/config.
	cd $(OUTPUT_DIR)/manifests/provider/config/envoy-gateway && $(ROOT_DIR)/$(tools/kustomize) edit set image envoyproxy/gateway-dev=$(IMAGE):$(TAG)
	$(tools/kustomize) build $(OUTPUT_DIR)/manifests/provider/config/default | kubectl apply -f -
	$(tools/kustomize) build $(OUTPUT_DIR)/manifests/infra/config/rbac | kubectl apply -f -

.PHONY: kube-undeploy
kube-undeploy: kube-uninstall ## Uninstall the Envoy Gateway controller into the Kubernetes cluster specified in ~/.kube/config.
	$(tools/kustomize) build $(OUTPUT_DIR)/manifests/provider/config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f - 
	rm -rf $(OUTPUT_DIR)/manifests/provider
	rm -rf $(OUTPUT_DIR)/manifests/infra

.PHONY: run-kube-local
run-kube-local: build kube-install ## Run Envoy Gateway locally.
	tools/hack/run-kube-local.sh

.PHONY: conformance 
conformance: create-cluster kube-install-image kube-deploy run-conformance delete-cluster ## Create a kind cluster, deploy EG into it, run Gateway API conformance, and clean up.

.PHONY: create-cluster
create-cluster: $(tools/kind) ## Create a kind cluster suitable for running Gateway API conformance.
	tools/hack/create-cluster.sh

.PHONY: kube-install-image
kube-install-image: image.build $(tools/kind) ## Install the EG image to a kind cluster using the provided $IMAGE and $TAG.
	tools/hack/kind-load-image.sh $(IMAGE) $(TAG)

.PHONY: run-conformance
run-conformance: ## Run Gateway API conformance.
	kubectl wait --timeout=5m -n gateway-system deployment/gateway-api-admission-server --for=condition=Available
	kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
	kubectl apply -f internal/provider/kubernetes/config/samples/gatewayclass.yaml
	go test -v -tags conformance ./test/conformance --gateway-class=envoy-gateway --debug=true

.PHONY: delete-cluster
delete-cluster: $(tools/kind) ## Delete kind cluster.
	$(tools/kind) delete cluster --name envoy-gateway

.PHONY: release-manifests
release-manifests: $(tools/kustomize) ## Generate Kubernetes release manifests.
	tools/hack/release-manifests.sh $(GATEWAY_API_VERSION) $(TAG)
