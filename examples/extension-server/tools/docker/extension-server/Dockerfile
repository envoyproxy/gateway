# Stage 1: Build the statically-linked Go binary
# Use a specific Go version for reproducible builds
FROM golang:1.24-alpine AS builder

# Set build arguments that can be passed in from the Makefile
ARG GO_LDFLAGS

# Set the working directory
WORKDIR /src

# Copy the go.mod and go.sum files to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# All remaining
COPY . .

# Build a statically-linked binary for the target platform.
# The 'distroless/static' base image requires a static binary.
# CGO_ENABLED=0 is crucial for this.
# buildx automatically provides TARGETPLATFORM (e.g., linux/amd64)
RUN CGO_ENABLED=0 go build \
    -ldflags="${GO_LDFLAGS}" \
    -o /extension-server \
    ./cmd/extension-server

# Stage 2: Create the final minimal image using the distroless base
# This is the exact base image from your original Dockerfile
FROM gcr.io/distroless/static:nonroot@sha256:e9ac71e2b8e279a8372741b7a0293afda17650d926900233ec3a7b2b7c22a246

# Copy ONLY the compiled binary from the 'builder' stage
COPY --from=builder /extension-server /usr/local/bin/extension-server

# This USER line is optional as 'nonroot' image defaults to it, but it's good to be explicit.
USER 65532:65532

# Set the entrypoint for the container to run our application
ENTRYPOINT ["/usr/local/bin/extension-server"]
