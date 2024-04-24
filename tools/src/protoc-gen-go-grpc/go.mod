module github.com/envoyproxy/gateway/tools/src/protoc-gen-go-grpc

go 1.22.2

require google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.3.0

require google.golang.org/protobuf v1.28.1 // indirect

// Resolve GHSA-8r3f-844c-mc37.
// This is a temporary fix until the next release of google.golang.org/grpc/cmd/protoc-gen-go-grpc.
// See https://github.com/grpc/grpc-go/issues/7092.
replace google.golang.org/protobuf => google.golang.org/protobuf v1.33.0
