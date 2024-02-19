# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot@sha256:6a3500b086c2856fbc189f5d11351bdbcf7c4dc5673c2b6070aac9d607da90d7
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/envoy-gateway /usr/local/bin/

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/envoy-gateway"]
