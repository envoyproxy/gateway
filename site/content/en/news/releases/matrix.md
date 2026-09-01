---
title: Compatibility Matrix
description: This section includes Compatibility Matrix of Envoy Gateway.
---

Envoy Gateway relies on the Envoy Proxy and the Gateway API, and runs within a Kubernetes cluster. Not all versions of each of these products can function together for Envoy Gateway. Supported version combinations are listed below; **bold** type indicates the versions of the Envoy Proxy and the Gateway API actually compiled into each Envoy Gateway release.

Each Envoy Gateway release supports exactly one Envoy Proxy minor version: the one it is built and tested against. This is unlike the Kubernetes column, which lists a range of tested versions. Running a different Envoy Proxy minor version is not supported.

| Envoy Gateway version | Envoy Proxy version         | Rate Limit version | Gateway API version | Kubernetes version         | End of Life |
| --------------------- | --------------------------- | ------------------ | ------------------- | -------------------------- | ----------- |
| latest                | **dev-latest**              | **master**         | **v1.6.1**          | v1.33, v1.34, v1.35, v1.36 | n/a         |
| v1.9                  | **distroless-v1.39.x**      | **17b1956c**       | **v1.6.1**          | v1.33, v1.34, v1.35, v1.36 | 2027/02/14  |
| v1.8                  | **distroless-v1.38.x**      | **fe26676d**       | **v1.5.1**          | v1.32, v1.33, v1.34, v1.35 | 2026/11/08  |
| v1.7                  | **distroless-v1.37.x**      | **3fb70258**       | **v1.4.1**          | v1.32, v1.33, v1.34, v1.35 | 2026/08/05  |
| v1.6                  | **distroless-v1.36.x**      | **3fb70258**       | **v1.4.0**          | v1.30, v1.31, v1.32, v1.33 | 2026/05/13  |
| v1.5                  | **distroless-v1.35.x**      | **a90e0e5d**       | **v1.3.0**          | v1.30, v1.31, v1.32, v1.33 | 2026/02/13  |
| v1.4                  | **distroless-v1.34.x**      | **3e085e5b**       | **v1.3.0**          | v1.30, v1.31, v1.32, v1.33 | 2025/11/13  |
| v1.3                  | **distroless-v1.33.x**      | **60d8e81b**       | **v1.2.1**          | v1.29, v1.30, v1.31, v1.32 | 2025/07/30  |
| v1.2                  | **distroless-v1.32.x**      | **28b1629a**       | **v1.2.0**          | v1.28, v1.29, v1.30, v1.31 | 2025/05/06  |
| v1.1                  | **distroless-v1.31.x**      | **91484c59**       | **v1.1.0**          | v1.27, v1.28, v1.29, v1.30 | 2025/01/22  |
| v1.0                  | **distroless-v1.29.x**      | **19f2079f**       | **v1.0.0**          | v1.26, v1.27, v1.28, v1.29 | 2024/09/13  |
| v0.6                  | **distroless-v1.28-latest** | **b9796237**       | **v1.0.0**          | v1.26, v1.27, v1.28        | 2024/05/02  |
| v0.5                  | **v1.27-latest**            | **e059638d**       | **v0.7.1**          | v1.25, v1.26, v1.27        | 2024/01/02  |
| v0.4                  | **v1.26-latest**            | **542a6047**       | **v0.6.2**          | v1.25, v1.26, v1.27        | 2023/10/24  |
| v0.3                  | **v1.25-latest**            | **f28024e3**       | **v0.6.1**          | v1.24, v1.25, v1.26        | 2023/08/09  |
| v0.2                  | **v1.23-latest**            |                    | **v0.5.1**          | v1.24                      | 2023/04/20  |

The Envoy Proxy column shows the supported minor version. The exact image that a given Envoy Gateway patch release ships may be a newer patch within that same minor version. To find the exact image for a release, check that release's notes.

{{% alert title="Overriding the Envoy Proxy image" color="warning" %}}
Setting `image` on the EnvoyProxy resource pins the data plane independently of the control plane. If the pinned image is a different Envoy Proxy minor version than the one listed above, Envoy Gateway may generate configuration that the proxy rejects.

A rejected configuration is silent while the proxies keep running, because Envoy continues to serve its last known good configuration. The failure surfaces when a proxy restarts, because a new proxy has no previous configuration to fall back on — and pod readiness can still pass while a listener is missing. To detect this, watch the xDS rejection metrics described in [Gateway Exported Metrics](/latest/tasks/observability/gateway-exported-metrics/).

If you only need to pull the image from a private registry, prefer `imageRepository` over `image`. Envoy Gateway then supplies the tag, so the data plane tracks the control plane automatically.
{{% /alert %}}
