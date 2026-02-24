EXAMPLE_APPS := simple-extension-server extension-server envoy-ext-auth grpc-ext-proc preserve-case-backend static-file-server dynamic-module-test
EXAMPLE_IMAGE_PREFIX ?= envoyproxy/gateway-
EXAMPLE_TAG ?= latest

# Extract envoy proxy version from DefaultEnvoyProxyImage (e.g., "distroless-v1.37.0" -> "v1.37.0").
# Empty on main branch where the image is "distroless-dev".
ENVOY_PROXY_VERSION := $(shell grep 'DefaultEnvoyProxyImage' api/v1alpha1/shared_types.go | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+')

kube-generate-examples:
	@$(LOG_TARGET)
	@pushd $(ROOT_DIR)/examples/extension-server; \
		make generate; \
		popd

.PHONY: kube-build-examples-image
kube-build-examples-image:
	@$(LOG_TARGET)
	@for app in $(EXAMPLE_APPS); do \
		pushd $(ROOT_DIR)/examples/$$app; \
		if [ -n "$(ENVOY_PROXY_VERSION)" ]; then \
			make docker-buildx ENVOY_VERSION=$(ENVOY_PROXY_VERSION); \
		else \
			make docker-buildx; \
		fi; \
		popd; \
	done

.PHONY: kube-install-examples-image
kube-install-examples-image: kube-build-examples-image
	@$(LOG_TARGET)
	@for app in $(EXAMPLE_APPS); do \
		tools/hack/kind-load-image.sh $(EXAMPLE_IMAGE_PREFIX)$$app $(EXAMPLE_TAG); \
	done

.PHONY: go.mod.tidy.examples
go.mod.tidy.examples:
	@$(LOG_TARGET)
	@for app in $(EXAMPLE_APPS); do \
		pushd $(ROOT_DIR)/examples/$$app; \
		go mod tidy -compat=$(GO_VERSION); \
		popd; \
	done

.PHONY: update-dynamic-module-deps
update-dynamic-module-deps: ## Update dynamic module SDK and envoy version in examples
	@$(LOG_TARGET)
	@tools/hack/bump-envoy-dynamic-modules.sh $(ENVOY_VERSION)
