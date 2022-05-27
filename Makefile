.PHONY: build
build:
	@go build -o ./bin/ github.com/envoyproxy/gateway/cmd/envoy-gateway

.PHONY: test
test:
	@go test ./...

