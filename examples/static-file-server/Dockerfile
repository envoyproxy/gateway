FROM golang:1.23.6 AS builder

ARG GO_LDFLAGS=""

WORKDIR /workspace
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . ./
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0  \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -o /bin/static-file-server -ldflags "${GO_LDFLAGS}" .

# Need root user for UDS
FROM gcr.io/distroless/static-debian11@sha256:59b2e1b85788cd1a8375e4e5e6142227d19875e8a3a96169d94d72adbcf7f273
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/static-file-server /
COPY files/ files/

ENTRYPOINT ["/static-file-server"]
