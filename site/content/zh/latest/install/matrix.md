---
title: 兼容性表格
description: 本节包含关于 Envoy Gateway 的兼容性表格。
---

Envoy Gateway 依赖于 Envoy Proxy 和 Gateway API，并在 Kubernetes 集群中运行。
这些产品的所有版本并非都可以与 Envoy Gateway 一起运行。下面列出了支持的版本组合；
**粗体**类型表示实际编译到每个 Envoy Gateway 版本中的 Envoy Proxy 和 Gateway API 的版本。

| Envoy Gateway 版本    | Envoy Proxy 版本            | Rate Limit 版本     | Gateway API 版本    | Kubernetes 版本             |
|-----------------------|-----------------------------|--------------------|---------------------|----------------------------|
| v1.0.0                | **distroless-v1.29.2**      | **19f2079f**       | **v1.0.0**          | v1.26, v1.27, v1.28, v1.29 |
| v0.6.0                | **distroless-v1.28-latest** | **b9796237**       | **v1.0.0**          | v1.26, v1.27, v1.28        |
| v0.5.0                | **v1.27-latest**            | **e059638d**       | **v0.7.1**          | v1.25, v1.26, v1.27        |
| v0.4.0                | **v1.26-latest**            | **542a6047**       | **v0.6.2**          | v1.25, v1.26, v1.27        |
| v0.3.0                | **v1.25-latest**            | **f28024e3**       | **v0.6.1**          | v1.24, v1.25, v1.26        |
| v0.2.0                | **v1.23-latest**            |                    | **v0.5.1**          | v1.24                      |
| latest                | **dev-latest**              | **master**         | **v1.0.0**          | v1.29, v1.30, v1.31, v1.32 |
