---
title: "BackendTrafficPolicy"
---

## Overview

This design document introduces the `BackendTrafficPolicy` API allowing users to configure
the behavior for how the Envoy Proxy server communicates with upstream backend services/endpoints.

## Goals

- Add an API definition to hold settings for configuring behavior of the connection between the backend services
and Envoy Proxy listener.

## Non Goals

- Define the API configuration fields in this API.

## Implementation

`BackendTrafficPolicy` is an implied hierarchy type API that can be used to extend [Gateway API][].
It can target either a `Gateway`, or an xRoute (`HTTPRoute`/`GRPCRoute`/etc.). When targeting a `Gateway`,
it will apply the configured settings within ght `BackendTrafficPolicy` to all children xRoute resources of that `Gateway`.
If a `BackendTrafficPolicy` targets an xRoute and a different `BackendTrafficPolicy` targets the `Gateway` that route belongs to,
then the configuration from the policy that is targeting the xRoute resource will win in a conflict.

### Example

Here is an example highlighting how a user can configure this API.

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
    - name: http
      protocol: HTTP
      port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: ipv4-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.foo.example.com"
  rules:
    - backendRefs:
        - group: ""
          kind: Service
          name: ipv4-service
          port: 3000
          weight: 1
      matches:
        - path:
            type: PathPrefix
            value: /
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: ipv6-route
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.bar.example.com"
  rules:
    - backendRefs:
        - group: ""
          kind: Service
          name: ipv6-service
          port: 3000
          weight: 1
      matches:
        - path:
            type: PathPrefix
            value: /
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: default-ipv-policy
  namespace: default
spec:
  protocols:
    enableIPv6: false
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: ipv6-support-policy
  namespace: default
spec:
  protocols:
    enableIPv6: true
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: ipv6-route
    namespace: default
```

## Features / API Fields

Here is a list of some features that can be included in this API. Note that this list is not exhaustive.

- Protocol configuration
- Circuit breaking
- Retries
- Keep alive probes
- Health checking
- Load balancing
- Rate limit

## Design Decisions

- This API will only support a single `targetRef` and can bind to only a `Gateway` or xRoute (`HTTPRoute`/`GRPCRoute`/etc.) resource.
- This API resource MUST be part of same namespace as the resource it targets.
- There can be only be ONE policy resource attached to a specific `Listener` (section)  within a `Gateway`
- If the policy targets a resource but cannot attach to it, this information should be reflected
in the Policy Status field using the `Conflicted=True` condition.
- If multiple polices target the same resource, the oldest resource (based on creation timestamp) will
attach to the Gateway Listeners, the others will not.
- If Policy A has a `targetRef` that includes a `sectionName` i.e.
it targets a specific Listener within a `Gateway` and Policy B has a `targetRef` that targets the same
entire Gateway then
  - Policy A will be applied/attached to the specific Listener defined in the `targetRef.SectionName`
  - Policy B will be applied to the remaining Listeners within the Gateway. Policy B will have an additional
  status condition `Overridden=True`.

## Alternatives

- The project can indefintely wait for these configuration parameters to be part of the [Gateway API][].

[Gateway API]: https://gateway-api.sigs.k8s.io/
