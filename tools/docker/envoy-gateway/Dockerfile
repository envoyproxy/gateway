FROM busybox@sha256:e226d6308690dbe282443c8c7e57365c96b5228f0fe7f40731b5d84d37a06839 AS source
# Create the data directory for eg
RUN mkdir -p /var/lib/eg && chmod -R 0777 /var/lib/eg

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/base-nossl:nonroot@sha256:d90f132a368246c5dcd65a2de958a86a99891945ffbf0005996bb7597eec2a2a
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/envoy-gateway /usr/local/bin/
COPY --from=source /var/lib /var/lib

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/envoy-gateway"]
