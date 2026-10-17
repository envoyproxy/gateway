# Custom Envoy image with a Wasm module baked into the filesystem for e2e.
ARG ENVOY_VERSION=dev

FROM docker.io/envoyproxy/envoy:distroless-${ENVOY_VERSION}
COPY envoy_filter_http_wasm_example.wasm /var/lib/envoy/envoy_filter_http_wasm_example.wasm
