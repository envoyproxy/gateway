# ENVOY_VERSION must match the SDK version in go.mod to ensure ABI compatibility.
# Update both together when changing the target Envoy version.
ARG ENVOY_VERSION=dev

FROM golang:1.26.5@sha256:63f132d58c1f589f0dcda584933a9bb44bfda1150f1506377f5a902f34d86033 AS builder

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
