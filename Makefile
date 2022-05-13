.PHONY: build
build:
	@go build github.com/envoyproxy/gateway/cmd/envoy-gateway

.PHONY: test
test:
	@go test ./...

