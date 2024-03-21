---
title: "ClientTrafficPolicy "
---

## Overview

This design document introduces the `ClientTrafficPolicy` API allowing system administrators to configure
the behavior for how the Envoy Proxy server behaves with downstream clients.

## Goals

* Add an API definition to hold settings for configuring behavior of the connection between the downstream
client and Envoy Proxy listener.

## Non Goals

* Define the API configuration fields in this API.

## Implementation

`ClientTrafficPolicy` is a [Direct Policy Attachment][] type API that can be used to extend [Gateway API][]
to define configuration that affect the connection between the downstream client and Envoy Proxy listener.

### Example

Here is an example highlighting how a user can configure this API.

```
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
kind: ClientTrafficPolicy
metadata:
  name: enable-proxy-protocol-policy
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
  enableProxyProtocol: true
```

## Features / API Fields

Here is a list of features that can be included in this API

* Downstream ProxyProtocol
* Downstream Keep Alives
* IP Blocking
* Downstream HTTP3

## Design Decisions

* This API will only support a single `targetRef` and can bind to only a `Gateway` resource.
* This API resource MUST be part of same namespace as the `Gateway` resource
* There can be only be ONE policy resource attached to a specific `Listener` (section)  within a `Gateway`
* If the policy targets a resource but cannot attach to it, this information should be reflected
in the Policy Status field using the `Conflicted=True` condition.
* If multiple polices target the same resource, the oldest resource (based on creation timestamp) will
attach to the Gateway Listeners, the others will not.
* If Policy A has a `targetRef` that includes a `sectionName` i.e.
it targets a specific Listener within a `Gateway` and Policy B has a `targetRef` that targets the same
entire Gateway then
  * Policy A will be applied/attached to the specific Listener defined in the `targetRef.SectionName`
  * Policy B will be applied to the remaining Listeners within the Gateway. Policy B will have an additional
  status condition `Overridden=True`.

## Alternatives

* The project can indefintely wait for these configuration parameters to be part of the [Gateway API].

[Direct Policy Attachment]: https://gateway-api.sigs.k8s.io/references/policy-attachment/#direct-policy-attachment
[Gateway API]: https://gateway-api.sigs.k8s.io/
