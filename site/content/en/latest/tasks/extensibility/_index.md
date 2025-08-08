---
title: "Extensibility"
weight: 4
description: This section includes Extensibility tasks.
---

Envoy Gateway provides several ways to extend its functionality beyond the built-in features.

## Extension Options

**Need access to Envoy Proxy features not available through the API ?**
- [Envoy Patch Policy](envoy-patch-policy) - Directly modify Envoy xDS configuration
- [Extension Server](extension-server) - Build external services to transform xDS configuration

**Want to add custom processing logic?**
- [WASM Extensions](wasm) - Run WebAssembly modules for high-performance custom logic
- [External Processing](ext-proc) - Call external gRPC services during request processing
- [Lua Extensions](lua) - Write lightweight scripting extensions
