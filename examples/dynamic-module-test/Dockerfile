# ENVOY_VERSION must match the SDK version in go.mod to ensure ABI compatibility.
# Update both together when changing the target Envoy version.
ARG ENVOY_VERSION=v1.39.1

FROM golang:1.26.8@sha256:9d2f36f06329b2a141b9db99ffa32765cf695ee57b813ca29e245e8670bcbfff AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 go build -buildmode=c-shared -o /build/libheader_mutation.so .

ARG ENVOY_VERSION
FROM docker.io/envoyproxy/envoy:distroless-${ENVOY_VERSION}
COPY --from=builder /build/libheader_mutation.so /usr/local/lib/libheader_mutation.so
