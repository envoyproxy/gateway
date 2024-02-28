# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot@sha256:49af06135e8bbe8ddc46c1d28b0bd00961aae9c9ed090bbc0237f58e1462dd4b
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/envoy-gateway /usr/local/bin/

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/envoy-gateway"]
