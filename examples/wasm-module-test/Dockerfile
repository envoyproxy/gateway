# Custom Envoy image with a Wasm module baked into the filesystem for e2e.
ARG ENVOY_VERSION=dev

FROM python:3.12-alpine AS modules
COPY envoy_filter_http_wasm_example.wasm /module.wasm
# A same-length header replacement preserves the Wasm binary layout.
RUN python3 -c 'from pathlib import Path; p = Path("/module.wasm"); Path("/different.wasm").write_bytes(p.read_bytes().replace(b"FOO", b"BAR"))'

FROM docker.io/envoyproxy/envoy:distroless-${ENVOY_VERSION}
COPY envoy_filter_http_wasm_example.wasm /var/lib/envoy/envoy_filter_http_wasm_example.wasm
COPY envoy_filter_http_wasm_example.wasm /var/lib/envoy/example-copy.wasm
COPY --from=modules /different.wasm /var/lib/envoy/example-different.wasm
