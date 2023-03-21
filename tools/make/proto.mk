##@ Protobufs

.PHONY: protos
protos: $(tools/buf) ## Compile all protobufs
	$(tools/buf) generate

.PHONY: buf-mod-update
buf-mod-update: $(tools/buf) ## Update buf.lock for protobuf dependency updates
	$(tools/buf) mod update
