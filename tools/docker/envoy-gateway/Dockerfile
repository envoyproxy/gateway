FROM busybox@sha256:37f7b378a29ceb4c551b1b5582e27747b855bbfaa73fa11914fe0df028dc581f AS source
# Create the data directory for eg
RUN mkdir -p /var/lib/eg && chmod -R 0660 /var/lib/eg

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/base-nossl:nonroot@sha256:ecbab76d6a504ddf7c58a9d786e70f1f1731fa546b1ac0b20dab35c6fc2f3138
ARG TARGETPLATFORM
COPY --chown=65532:65532 $TARGETPLATFORM/envoy-gateway /usr/local/bin/
COPY --from=source /var/lib /var/lib

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/envoy-gateway"]
