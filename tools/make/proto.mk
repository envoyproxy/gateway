##@ Protobufs

.PHONY: protos
protos: $(tools/protoc-gen-go) $(tools/protoc-gen-go-grpc) ## Compile all protobufs
	@go tool buf generate

.PHONY: buf-mod-update
buf-mod-update: ## Update buf.lock for protobuf dependency updates
	@go tool buf mod update
