---
title: " Extensibility Options"
---

## Overview
Envoy Gateway can be extended to support custom logic and integrations through filters, external processors, and WebAssembly modules.

## Key Concepts
| Concept | Description |
|----------|--------------|
| Filters | Processing units in request/response path. |
| External Processors | Run custom logic in out-of-process services. |
| WebAssembly | Execute sandboxed plugins inside Envoy. |

## Use Cases
- Add custom authentication logic.  
- Transform headers or payloads.  
- Implement A/B testing filters.  

## Implementation
Envoy Gateway exposes extension hooks via configuration APIs. Custom extensions can be loaded dynamically through WASM or external processors.

## Examples
- Add a custom WASM filter.  
- Modify headers before routing.  
- Implement request sanitization.

## Related Resources
- [WASM Extension Reference](https://www.envoyproxy.io/docs/envoy/latest/wasm)
