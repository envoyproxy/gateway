---
title: "Dynamic Module"
---

## Overview

The Dynamic Module extension allows you to load shared libraries at runtime to extend Envoy's functionality. Dynamic modules are shared libraries that implement the [Envoy ABI](https://github.com/envoyproxy/envoy/blob/main/source/extensions/dynamic_modules/abi.h) written in a pure C header file.

## Example

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: dynamicmodule-example
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  dynamicModule:
  - name: my-dynamic-module
    module: my_module
    config:
      key: value
    # Prevent the module from being unloaded
    doNotClose: true
```

## API Reference

### DynamicModule

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `name` | string | A unique name for this Dynamic Module extension. It is used to identify the Dynamic Module extension if multiple extensions are loaded. It's also used for logging/debugging. If not specified, EG will generate a unique name for the Dynamic Module extension. | No |
| `module` | string | The name of the dynamic module to load. The module name is used to search for the shared library file in the search path. The search path is configured by the environment variable ENVOY_DYNAMIC_MODULES_SEARCH_PATH. The actual search path is ${ENVOY_DYNAMIC_MODULES_SEARCH_PATH}/lib${name}.so. | Yes |
| `config` | object | The configuration for the Dynamic Module extension. This configuration will be passed to the Dynamic Module extension. | No |
| `doNotClose` | boolean | Set to true to prevent the module from being unloaded with dlclose. This is useful for modules that have global state that should not be unloaded. A module is closed when no more references to it exist in the process. For example, no HTTP filters are using the module (e.g. after configuration update). | No |

## Notes

- Dynamic modules are only supported in Envoy v1.34 and later.
- Dynamic modules are loaded at runtime and must be compatible with the Envoy binary that loads them.
- Dynamic modules run in the same process as Envoy and have the same privilege level, so they should be fully trusted.
- The dynamic module must be built with the SDK of the same version as the Envoy binary that loads it.
- For more information, see the [Envoy documentation on dynamic modules](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/dynamic_modules).
