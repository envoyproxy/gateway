# Use distroless as minimal base image to package the extension service
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
ARG TARGETPLATFORM
ARG BUILD_DIR=.

WORKDIR /
COPY $BUILD_DIR/$TARGETPLATFORM/extension-service /usr/local/bin/

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/extension-service"]
