##@ Protobufs

.PHONY: protos
protos:
	$(GO_TOOL) buf generate

.PHONY: buf-dep-update
buf-dep-update: ## Update buf.lock for protobuf dependency updates
	$(GO_TOOL) buf dep update
