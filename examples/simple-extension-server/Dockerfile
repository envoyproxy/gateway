FROM golang:1.25.5 AS builder

ARG GO_LDFLAGS=""
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy gateway module files for the replace directive (../../ from examples/simple-extension-server)
COPY go.mod go.sum ./
COPY envoygateway ./envoygateway

# Copy the simple-extension-server app
COPY examples/simple-extension-server ./examples/simple-extension-server

WORKDIR /workspace/examples/simple-extension-server
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0  \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -o /bin/simple-extension-server -ldflags "${GO_LDFLAGS}" .

# Need root user for UDS
FROM gcr.io/distroless/static-debian11
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/simple-extension-server /

ENTRYPOINT ["/simple-extension-server"]
