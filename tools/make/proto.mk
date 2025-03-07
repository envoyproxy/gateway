##@ Protobufs

.PHONY: protos
protos:
	@go tool buf generate

.PHONY: buf-dep-update
buf-dep-update: ## Update buf.lock for protobuf dependency updates
	@go tool buf dep update
