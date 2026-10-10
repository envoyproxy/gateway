FROM golang:1.26.9@sha256:f1f0bcc2c524a3ced375fcb4d1ecb7aa371aa7070e112599aaca45cc02d0101b AS builder

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
    go build -o /bin/backend-utilization -ldflags "${GO_LDFLAGS}" .

FROM gcr.io/distroless/static-debian11
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/backend-utilization /

ENTRYPOINT ["/backend-utilization"]
