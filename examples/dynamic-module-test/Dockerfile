FROM golang:1.26.1@sha256:595c7847cff97c9a9e76f015083c481d26078f961c9c8dca3923132f51fe12f1 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 go build -buildmode=c-shared -o /build/libheader_mutation.so .

# Envoy image tag and digest are updated by tools/hack/bump-envoy-dynamic-modules.sh during releases.
FROM docker.io/envoyproxy/envoy:distroless-dev@sha256:1679d1bb44c7f90aca4a0f5e33f7c7c5723e96a90b38f9ad5a5b158ed4c95a40
COPY --from=builder /build/libheader_mutation.so /usr/local/lib/libheader_mutation.so
