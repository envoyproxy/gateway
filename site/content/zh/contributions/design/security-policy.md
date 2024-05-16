---
title: "SecurityPolicy"
---

## 概述 {#overview}

本设计文档引入了 `SecurityPolicy` API，允许系统管理员为进入网关的流量配置身份验证和鉴权策略。

## 目标 {#goals}

* 添加 API 定义以保存用于配置进入网关的流量的身份验证和鉴权规则的设置。

## 非目标 {#non-goals}

* 定义该 API 中的 API 配置字段。

## 实现 {#implementation}

`SecurityPolicy` 是一个[策略附件][]类型的 API，可用于扩展 [Gateway API][] 来定义身份验证和鉴权规则。

### 示例 {#example}

以下示例重点介绍了用户如何配置此 API。

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
  namespace: default
spec:
  gatewayClassName: eg
  listeners:
    - name: https
      protocol: HTTPS
      port: 443
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - backendRefs:
        - group: ""
          kind: Service
          name: backend
          port: 3000
          weight: 1
      matches:
        - path:
            type: PathPrefix
            value: /
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: jwt-authn-policy
  namespace: default
spec:
  jwt:
    providers:
    - name: example
      remoteJWKS:
        uri: https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/jwks.json
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
```

## 功能及 API 字段 {#features-api-fields}

以下是此 API 中包含的功能列表：
* JWT 基础鉴权
* OIDC 鉴权
* 外部认证
* Basic Auth
* API Key Auth
* CORS（跨域）

## 设计决策 {#design-decisions}

* 此 API 仅支持单个 `targetRef`，并且可以绑定到 `Gateway` 资源或 `HTTPRoute` 或 `GRPCRoute`。
* 此 API 资源**必须**与 targetRef 资源属于同一命名空间
* 只能有**一个**策略资源附加到特定的 targetRef，例如 `Gateway` 内的 `Listener`（部分）
* 如果策略针对某个资源但无法附加到该资源，则应使用 `Conflicted=True` 条件将该信息反映在“策略状态”字段中。
* 如果多个策略针对同一资源，则最旧的资源（基于创建时间戳）将附加到网关侦听器，其他资源则不会。
* 如果策略 A 具有包含 `sectionName` 的 `targetRef`，即它以 `Gateway` 内的特定侦听器为目标，
  并且策略 B 具有以同一整个 Gateway 为目标的 `targetRef`，则
  * 策略 A 将应用/附加到 `targetRef.SectionName` 中定义的特定监听器
  * 策略 B 将应用于 Gateway 内的其余侦听器。策略 B 将具有附加状态条件 `Overridden=True`。
* 针对拥有具体范围的策略胜过针对缺少具体范围的策略。即，针对 xRoute（`HTTPRoute` 或 `GRPCRoute`）的策略会覆盖针对侦听器的策略，
  该侦听器是该路由的 parentRef，而侦听器又会覆盖针对侦听器/部分所属 Gateway 的策略。

## 替代方案 {#alternatives}

* 项目可以无限期地等待这些配置参数成为 [Gateway API][] 的一部分。

[策略附件]: https://gateway-api.sigs.k8s.io/references/policy-attachment 
[Gateway API]: https://gateway-api.sigs.k8s.io/
