FROM golang:1.26.8@sha256:9d2f36f06329b2a141b9db99ffa32765cf695ee57b813ca29e245e8670bcbfff AS builder

ARG GO_LDFLAGS=""

WORKDIR /workspace
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -o /bin/remote-infra -ldflags "${GO_LDFLAGS}" .

# Alpine rather than distroless because the sidecar's exec probes need a
# `test` binary on PATH (busybox provides it).
FROM alpine:3.20
COPY --from=builder /bin/remote-infra /

ENTRYPOINT ["/remote-infra"]
