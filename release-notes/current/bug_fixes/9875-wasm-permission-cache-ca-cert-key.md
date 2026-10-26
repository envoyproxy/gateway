Fixed the Wasm image permission cache key omitting the CA certificate, which allowed a permission check result to be reused across different TLS trust configurations.
