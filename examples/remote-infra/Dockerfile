FROM golang:1.26.7@sha256:45a5f7a810238aabcbad211d70b9ae082022d96f7c7259e94041ad1b933575ac AS builder

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
