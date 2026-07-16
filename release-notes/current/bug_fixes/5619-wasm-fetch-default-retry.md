Fixed Wasm extensions remaining permanently failed after transient errors fetching the Wasm module.
Envoy's built-in behavior only retried the fetch once after ~1 second and never re-attempted it,
leaving the filter failed until the next configuration update. Envoy Gateway now configures the
fetch with up to 10 retries using jittered exponential backoff (1s base interval, 30s max interval).
