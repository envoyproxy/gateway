FROM busybox@sha256:f64ff79725d0070955b368a4ef8dc729bd8f3d8667823904adcb299fe58fc3da AS source
# Create the data directory for eg
RUN mkdir -p /var/lib/eg

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/base-nossl:nonroot@sha256:ecbab76d6a504ddf7c58a9d786e70f1f1731fa546b1ac0b20dab35c6fc2f3138
ARG TARGETPLATFORM
COPY --chown=65532:65532 $TARGETPLATFORM/envoy-gateway /usr/local/bin/
COPY --from=source --chown=65532:65532 /var/lib /var/lib

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/envoy-gateway"]
