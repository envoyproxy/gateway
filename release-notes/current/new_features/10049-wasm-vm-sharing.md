Added the per-entry `shareVM` option to EnvoyExtensionPolicy to allow compatible Wasm extensions in the same policy namespace to reuse a VM on each Envoy worker. Sharing is disabled by default.
