# ENVOY_VERSION must match the SDK version in go.mod to ensure ABI compatibility.
# Update both together when changing the target Envoy version.
ARG ENVOY_VERSION=dev

FROM golang:1.27.2@sha256:5bc7f572bbaa98885a3a1fd9c0aa76b59e3e14e8628bfc316bbfd0c701e4818c AS builder

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
