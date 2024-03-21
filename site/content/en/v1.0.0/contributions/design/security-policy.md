---
title: "SecurityPolicy "
---

## Overview

This design document introduces the `SecurityPolicy` API allowing system administrators to configure
authentication and authorization policies to the traffic entering the gateway.

## Goals
* Add an API definition to hold settings for configuring authentication and authorization rules
on the traffic entering the gateway.

## Non Goals
* Define the API configuration fields in this API.

## Implementation
`SecurityPolicy` is a [Policy Attachment][] type API that can be used to extend [Gateway API][]
to define authentication and authorization rules.

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

## Features / API Fields
Here is a list of features that can be included in this API
* JWT based authentication
* OIDC Authentication
* External Authorization
* Basic Auth
* API Key Auth
* CORS

## Design Decisions
* This API will only support a single `targetRef` and can bind to a `Gateway` resource or a `HTTPRoute` or `GRPCRoute`.
* This API resource MUST be part of same namespace as the targetRef resource
* There can be only be ONE policy resource attached to a specific targetRef e.g. a `Listener` (section)  within a `Gateway`
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
* A Policy targeting the most specific scope wins over a policy targeting a lesser specific scope.
  i.e. A Policy targeting a xRoute (`HTTPRoute` or `GRPCRoute`) overrides a Policy targeting a Listener that is
this route's parentRef which in turn overrides a Policy targeting the Gateway the listener/section is a part of. 

## Alternatives
* The project can indefinitely wait for these configuration parameters to be part of the [Gateway API][].

[Policy Attachment]: https://gateway-api.sigs.k8s.io/references/policy-attachment 
[Gateway API]: https://gateway-api.sigs.k8s.io/
