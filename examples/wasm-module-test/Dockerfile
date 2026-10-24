# Custom Envoy image with a Wasm module baked into the filesystem for e2e.
ARG ENVOY_VERSION=dev

FROM alpine:3.20 AS modules
COPY envoy_filter_http_wasm_example.wasm /module.wasm
# A same-length header replacement preserves the Wasm binary layout.
RUN sed 's/FOO/BAR/g' /module.wasm > /different.wasm

FROM docker.io/envoyproxy/envoy:distroless-${ENVOY_VERSION}
COPY envoy_filter_http_wasm_example.wasm /var/lib/envoy/envoy_filter_http_wasm_example.wasm
COPY envoy_filter_http_wasm_example.wasm /var/lib/envoy/example-copy.wasm
COPY --from=modules /different.wasm /var/lib/envoy/example-different.wasm
