FROM golang:1.27.3@sha256:957861185dbf29d93668ace41e011d091c6e779953d47f0316ef2cc58e28fe2b AS builder

ARG GO_LDFLAGS=""

WORKDIR /workspace/server
COPY examples/static-file-server/go.mod examples/static-file-server/go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY examples/static-file-server/ ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0  \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -o /bin/static-file-server -ldflags "${GO_LDFLAGS}" .

WORKDIR /workspace/wasm
COPY examples/wasm-module-test/go.mod examples/wasm-module-test/go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY examples/wasm-module-test/*.go ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    GOOS=wasip1 GOARCH=wasm \
    go build -buildmode=c-shared -ldflags="-s -w" -o /bin/backend_callout.wasm .

# Need root user for UDS
FROM gcr.io/distroless/static-debian11@sha256:1dbe426d60caed5d19597532a2d74c8056cd7b1674042b88f7328690b5ead8ed
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/static-file-server /
COPY examples/static-file-server/files/ files/
COPY --from=builder /bin/backend_callout.wasm /files/wasm/backend_callout.wasm

ENTRYPOINT ["/static-file-server"]
