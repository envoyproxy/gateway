FROM golang:1.27.3@sha256:957861185dbf29d93668ace41e011d091c6e779953d47f0316ef2cc58e28fe2b AS builder

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
    go build -o /bin/preserve-case-backend -ldflags "${GO_LDFLAGS}" .

# Need root user for UDS
FROM gcr.io/distroless/static-debian11@sha256:1dbe426d60caed5d19597532a2d74c8056cd7b1674042b88f7328690b5ead8ed
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/preserve-case-backend /

ENTRYPOINT ["/preserve-case-backend"]
