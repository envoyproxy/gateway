FROM busybox@sha256:db142d433cdde11f10ae479dbf92f3b13d693fd1c91053da9979728cceb1dc68 AS source
# Create the data directory for eg
RUN mkdir -p /var/lib/eg

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot@sha256:6cd937e9155bdfd805d1b94e037f9d6a899603306030936a3b11680af0c2ed58
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/envoy-gateway /usr/local/bin/
COPY --from=source --chown=65532:65532 /var/lib /var/lib

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/envoy-gateway"]
