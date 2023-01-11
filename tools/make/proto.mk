##@ Protobufs

.PHONY: protos
protos: $(tools/buf) ## Compile all protobufs
	$(tools/buf) generate
