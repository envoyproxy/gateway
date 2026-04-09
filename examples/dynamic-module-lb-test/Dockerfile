# ENVOY_VERSION must match the target Envoy version for ABI compatibility.
ARG ENVOY_VERSION=dev

FROM golang:1.26.1 AS builder

WORKDIR /build
COPY go.mod ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 go build -buildmode=c-shared -o /build/libfirst-host-lb.so .

ARG ENVOY_VERSION
FROM docker.io/envoyproxy/envoy:distroless-${ENVOY_VERSION}
COPY --from=builder /build/libfirst-host-lb.so /usr/local/lib/libfirst-host-lb.so
