# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION ?= 1.24.1
# GATEWAY_API_VERSION refers to the version of Gateway API CRDs.
# For more details, see https://gateway-api.sigs.k8s.io/guides/getting-started/#installing-gateway-api 
GATEWAY_API_VERSION ?= $(shell go list -m -f '{{.Version}}' sigs.k8s.io/gateway-api)

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
	$(tools/kustomize) build internal/provider/kubernetes/config/crd | kubectl apply -f -
	kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/experimental-install.yaml

.PHONY: kube-uninstall
kube-uninstall: manifests $(tools/kustomize) ## Uninstall Envoy Gateway CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(tools/kustomize) build internal/provider/kubernetes/config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -
	kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/experimental-install.yaml

.PHONY: kube-deploy
kube-deploy: kube-install ## Install Envoy Gateway controller into the Kubernetes cluster specified in ~/.kube/config.
	$(tools/kustomize) build internal/provider/kubernetes/config/default | kubectl apply -f -

.PHONY: kube-undeploy
kube-undeploy: kube-uninstall ## Uninstall the Envoy Gateway controller into the Kubernetes cluster specified in ~/.kube/config.
	$(tools/kustomize) build internal/provider/kubernetes/config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f - 

.PHONY: run-kube-local ## Run EG locally.
run-kube-local: kube-install
	hack/run-kube-local.sh

.PHONY: gateway-conformance 
gateway-conformance: create-conformance-cluster kube-deploy run-gateway-conformance delete-conformance-cluster ## Create a kind cluster, deploy EG into it, run Gateway API conformance, and clean up.

.PHONY: create-conformance-cluster
create-conformance-cluster: ## Create a kind cluster suitable for running Gateway API conformance.
	hack/create-conformance-cluster.sh

.PHONY: run-gateway-conformance
run-gateway-conformance: ## Run Gateway API conformance.
	kubectl wait --timeout=5m -n gateway-system deployment/gateway-api-admission-server --for=condition=Available
	kubectl apply -f internal/provider/kubernetes/config/samples/gatewayclass.yaml
	go test -tags conformance ./test/conformance --gateway-class=envoy-gateway

.PHONY: delete-conformance-cluster
delete-conformance-cluster: ## Delete conformance kind cluster.
	kind delete cluster --name envoy-gateway-conformance
