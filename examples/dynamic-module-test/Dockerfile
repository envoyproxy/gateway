# ENVOY_VERSION must match the SDK version in go.mod to ensure ABI compatibility.
# Update both together when changing the target Envoy version.
ARG ENVOY_VERSION=dev

FROM golang:1.27.4@sha256:08bcb99a38f85a265f9a93d1d3323e2440eec333380f3ea26ee0dec9264bb41c AS builder

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
