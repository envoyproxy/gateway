---
title: " Extensibility Options"
---

## Overview
Envoy Gateway can be extended to support custom logic and integrations through filters, external processors, and WebAssembly modules.

Envoy Gateway exposes extension hooks via its configuration APIs.  
Depending on the extension model, custom modules can be:

- **Loaded dynamically** inside Envoy (for WASM and Lua).  
- **Invoked remotely** via gRPC (for external processors).  
- **Integrated through translation layers** (for patch policy or extension server).

This design keeps the control plane declarative and extensible while maintaining Envoy’s runtime performance and isolation guarantees.

## Key Concepts

| Concept | Description |
|----------|--------------|
| Filters | Processing units in request/response path. |
| [External Processors](./../tasks/extensibility/ext-proc.md) | Run custom logic in out-of-process services. |
| [WebAssembly](./../tasks/extensibility/wasm.md) | Execute sandboxed plugins inside Envoy. |
| [Lua Extensions](./../tasks/extensibility/lua.md) | Lightweight scripting support within Envoy configuration for quick customizations without rebuilding the proxy. | Simple request manipulation or header injection. |
| [Extension Server](./../tasks/extensibility/extension-server.md) | Mechanisms to directly modify or transform Envoy xDS configuration during translation. Useful for advanced or unsupported features. | Injecting Envoy-native features not yet supported by Envoy Gateway APIs. |

## Use Cases
Different extension methods suit different needs:

- **For inline logic** close to the data path → use **Filters** or **WASM**.  
- **For complex or external integrations** → use **External Processing**.  
- **For rapid prototyping or scripting** → use **Lua**.  
- **For deep Envoy configuration access** → use **Patch Policy** or **Extension Server**.

## Implementation
Envoy Gateway exposes extension hooks via configuration APIs. Custom extensions can be loaded dynamically through WASM or external processors.

## Examples
- Add custom authentication or JWT validation logic.  
- Transform or sanitize headers before routing.  
- Implement dynamic routing for A/B testing.  
- Enrich telemetry data with external metadata.  
- Prototype new traffic management features via Lua.

## Related Resources
- [Extension Options](./../tasks/extensibility/_index.md)
