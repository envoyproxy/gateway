FROM busybox@sha256:82742949a3709938cbeb9cec79f5eaf3e48b255389f2dcedf2de29ef96fd841c as source
# Create the data directory for eg
RUN mkdir -p /var/lib/eg

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot@sha256:42d15c647a762d3ce3a67eab394220f5268915d6ddba9006871e16e4698c3a24
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/envoy-gateway /usr/local/bin/
COPY --from=source --chown=65532:65532 /var/lib /var/lib

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/envoy-gateway"]
