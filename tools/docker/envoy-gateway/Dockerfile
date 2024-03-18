# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot@sha256:55c636171053dbc8ae07a280023bd787d2921f10e569f3e319f1539076dbba11
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/envoy-gateway /usr/local/bin/

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/envoy-gateway"]
