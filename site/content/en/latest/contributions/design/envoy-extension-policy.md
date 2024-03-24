---
title: "EnvoyExtensionPolicy "
---

## Overview

This design document introduces the `EnvoyExtensionPolicy` API allowing system administrators to configure traffic
processing extensibility policies, based on existing Network and HTTP Envoy proxy [extension points][].

Envoy Gateway already provides two methods of control plane extensibility that can be used to achieve this functionality:
* [Envoy Patch Policy][] can be used to patch Listener filters and HTTP Connection Manager filters. 
* [Envoy Extension Manager][] can be used to programmatically mutate Listener filters and HTTP Connection Manager filters.

These approaches require a high level of Envoy and Envoy Gateway expertise and may create a significant operational 
burden for users (see [Alternatives][] for more details). For this reason, this document proposes to support Envoy 
data plane extensibility options as first class citizens of Envoy Gateway. 

## Goals
* Add an API definition to hold settings for configuring extensibility rules on the traffic entering the gateway.

## Non Goals
* Define the API configuration fields in this API.
* Define the API for the following extension options:
    * Native Envoy extensions: custom C++ extensions that must be compiled into the Envoy binary.
    * Non-filter extensions: services, matchers, tracers, private key providers, resource monitors, etc.

## Implementation
`EnvoyExtensionPolicy` is a [Policy Attachment][] type API that can be used to extend [Gateway API][]
to define traffic extension rules.

`BackendTrafficPolicy` is enhanced to allow users to provide per-route config for Extensions.

### Example
Here is an example highlighting how a user can configure this API for the External Processing extension.

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
kind: EnvoyExtensionPolicy
metadata:
  name: ext-proc-policy
  namespace: default
spec:
  priority: 10
  extProc:
  - service:
      backendRef:
        group: ""
        kind: Service
        name: myExtProc
        port: 3000
    processingMode:
      request:
        headers: SEND
        body: BUFFERED
      response:
        headers: SKIP
        body: STREAMED
    attributes:
      request:
      - xds.route_metadata
      - connection.requested_server_name
      response:
      - request.path
    messageTimeout: 5s
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
```

## Features / API Fields
Here is a list of features that can be included in this API
* Network Filters:
    * WASM
    * Golang
* HTTP Filters:
    * External Processing
    * Lua
    * WASM
    * Golang

## Design Decisions
* This API will only support a single `targetRef` and can bind to a `Gateway` resource or a `HTTPRoute` or `GRPCRoute` or `TCPRoute`.
* Extensions that support both Network and HTTP filter variants (e.g. WASM, Golang) will be translated to the appropriate filter type according to the sort of route that they attach to.
* Extensions that only support HTTP extensibility (Ext-Proc, LUA) can only be attached to HTTP/GRPC Routes.  
* A user-defined extension that is added to the request processing flow can have a significant impact on security,
  resilience and performance of the proxy. Gateway Operators can restrict access to the extensibility policy using K8s RBAC. 
* Extensibility will be disabled by default, and can be enabled using [Envoy Gateway][] API. 
* Users may need to customize the order of extension and built-in filters. This will be addressed in a separate issue.  
* Gateway operators may need to include multiple extensions (e.g. WASM modules developed by different teams and distributed separately). 
  This API will support attachment of multiple policies. Extension will execute in an order defined by the priority field.
* This API resource MUST be part of same namespace as the targetRef resource
* If the policy targets a resource but cannot attach to it, this information should be reflected
  in the Policy Status field using the `Conflicted=True` condition.
* If Policy A has a `targetRef` that includes a `sectionName` i.e.
  it targets a specific Listener within a `Gateway` and Policy B has a `targetRef` that targets the same
  entire Gateway then
    * Policy A will be applied/attached to the specific Listener defined in the `targetRef.SectionName`
    * Policy B will be applied to the remaining Listeners within the Gateway. Policy B will have an additional
      status condition `Overridden=True`.
* A Policy targeting the most specific scope wins over a policy targeting a lesser specific scope.
  i.e. A Policy targeting a `Listener` overrides a Policy targeting the `Gateway` the listener/section is a part of.


## Alternatives
* The project can indefinitely wait for these configuration parameters to be part of the [Gateway API][].
* The project can implement support for HTTP traffic extensions using vendor-specific [Gateway API Route Filters][]
  instead of policies. However, this option will is less convenient for definition of gateway-level extensions.
* Users can leverage the existing [Envoy Patch Policy][] to inject extension filters. However, Envoy Gateway strives 
  to provide a simple abstraction for common use cases and easy operations. Envoy patches require a high level of 
  end-user Envoy expertise, and knowledge of how Envoy Gateway generates XDS. Such patches may be too difficult 
  and fragile for some users to maintain. 
* Users can leverage the existing [Envoy Extension Manager][] to inject extension filters. However, this requires a
  significant investment by users to build and operate an extension manager alongside Envoy Gateway.
  
[extension points]: https://www.envoyproxy.io/docs/envoy/latest/extending/extending
[Policy Attachment]: https://gateway-api.sigs.k8s.io/references/policy-attachment
[Gateway API]: https://gateway-api.sigs.k8s.io/
[Gateway API Route Filters]: https://gateway-api.sigs.k8s.io/api-types/httproute/#filters-optional
[Envoy Gateway]: ../../api/extension_types#envoygateway
[Envoy Patch Policy]: ../../api/extension_types#envoypatchpolicy
[Envoy Extension Manager]: ./extending-envoy-gateway
[Alternatives]: #Alternatives
