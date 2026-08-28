FROM golang:1.27.0@sha256:640a234f4bea3e399c056b7b8f9c667c4939befae8db2f14e9785e16eccd4205 AS builder

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
    go build -o /bin/envoy-ext-auth -ldflags "${GO_LDFLAGS}" .

# Make our production image
FROM gcr.io/distroless/static-debian11:nonroot
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/envoy-ext-auth /

USER nonroot:nonroot
ENTRYPOINT ["/envoy-ext-auth"]
