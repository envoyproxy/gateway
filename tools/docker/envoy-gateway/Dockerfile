FROM busybox@sha256:37f7b378a29ceb4c551b1b5582e27747b855bbfaa73fa11914fe0df028dc581f AS source
# Create the data directory for eg
RUN mkdir -p /var/lib/eg

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/base-nossl:nonroot@sha256:d1fc914c43cea489c26c896721344a49a1761b9bb678bcba1758772d22913302
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/envoy-gateway /usr/local/bin/
COPY --from=source /var/lib /var/lib

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/envoy-gateway"]
