
EXAMPLE_APPS := grpc-ext-auth envoy-als grpc-ext-proc http-ext-auth preserve-case-backend
EXAMPLE_IMAGE_PREFIX ?= envoyproxy/gateway-
EXAMPLE_TAG ?= latest

.PHONY: kube-build-examples-image
kube-build-examples-image:
	@$(LOG_TARGET)
	@for app in $(EXAMPLE_APPS); do \
		pushd $(ROOT_DIR)/examples/$$app; \
		make docker-buildx; \
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