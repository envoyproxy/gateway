---
title: "Backend "
---

## Overview

This design document introduces the `Backend` API allowing system administrators to represent backends without the use 
of a K8s `Service` resource. 

Common use cases for non-Service backends in the K8s and Envoy ecosystem include:
- Cluster-external endpoints, which are currently second-class citizens in Gateway-API 
  (supported using [Services and FQDN endpoints][]).
- Host-local endpoints, such as sidecars or daemons that listen on unix domain sockets, 
  that cannot be represented by a K8s service at all.

Several projects currently support backends that are not registered in the infrastructure-specific service registry. 
- K8s Ingress: [Resource Backends][]
- Istio: [Service Entry][]
- Gloo Edge: [Upstream][]
- Consul: [External Services][]

## Goals
* Add an API definition to hold settings for configuring Unix Domain Socket, FQDN and IP.
* Determine which resources may reference the new backend resource.
* Determine which existing Gateway-API and Envoy Gateway policies may attach to the new backend resource. 

## Non Goals
* Support specific backend types, such as S3 Bucket, Redis, AMQP, InfluxDB, etc.  

## Implementation

The `Backend` resource is an implementation-specific Gateway-API [BackendObjectReference Extension][]. 

### Example
Here is an example highlighting how a user can configure this API for the External Processing extension.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: backend-svc
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: backend
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: backend-uds
spec:
  addresses:
    - unixDomainSocketAddress:
        path: /var/run/backend.sock
      applicationProtocol: HTTP2
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: backend-ips
spec:
  addresses:
    - socketAddress:
        address: 10.244.0.28
        port: 3000
    - socketAddress:
        address: 10.244.0.29
        port: 3000        
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - backendRefs:
        - group: "gateway.envoyproxy.io"
          kind: Backend
          name: backend-uds
          port: 3000
          weight: 1
        - group: ""
          kind: Service
          name: backend
          port: 3000
          weight: 1          
      matches:
        - path:
            type: PathPrefix
            value: /

```

## Design Decisions
* All existing and future `BackendObjectReference` in Envoy Gateway MUST support the `Backend` kind. 
* Gateway-API and Envoy Gateway policies that attach to Services ([BackendTLSPolicy][https://gateway-api.sigs.k8s.io/geps/gep-1897/], [BackendLBPolicy][https://gateway-api.sigs.k8s.io/geps/gep-1619/]) MUST support `Backend` attachment. 
* The `Backend` API SHOULD support other Gateway-API backend features, such as [Backend Protocol Selection][].
* This API resource MUST be part of same namespace as the targetRef resource. The `Backend` API MUST be subject to the same cross-namespace reference restriction as referenced `Service` resources.   
* To limit the overall maintenance effort, the `Backend` API SHOULD support multiple generic address types 
  (UDS, FQDN, IP). The `Backend` API MUST NOT support vendor-specific backend types.
* Both `Backend` and `Service` resources may appear in the same `BackendRefs` list.
* The Optional `Port` field is not evaluated when referencing a `Backend`.  
* Referenced `Backend` resources MUST be translated to envoy endpoints, similar to the current `Service` translation.
* Certain combinations of `Backend` and `Service` are incompatible. For example, a Unix Domain Socket and a FQDN service
  require different cluster service discovery types (Static/EDS and Strict-DNS accordingly).
* If a Backend that is references by a route cannot be translated, the `Route` resource will have an `Accepted=False` 
  condition with a `RouteReasonUnsupportedValue` reason. 
  
## Alternatives
* The project can indefinitely wait for these configuration parameters to be part of the [Gateway API][].
* Users can leverage the existing [Envoy Patch Policy][] or [Envoy Extension Manager][] to inject custom envoy clusters
  and route configuration. However, these features require a high level of envoy expertise, investment and maintenance. 

[BackendObjectReference Extension]: https://gateway-api.sigs.k8s.io/guides/migrating-from-ingress/?h=extensi#approach-to-extensibility
[Resource Backends]: https://kubernetes.io/docs/concepts/services-networking/ingress/#resource-backend
[Services and FQDN endpoints]: https://gateway.envoyproxy.io/latest/user/traffic/routing-outside-kubernetes/
[Service Entry]: https://istio.io/latest/docs/reference/config/networking/service-entry/
[Upstream]: https://docs.solo.io/gloo-edge/1.7.23/reference/api/github.com/solo-io/gloo/projects/gloo/api/v1/upstream.proto.sk/
[External Services]: https://developer.hashicorp.com/consul/tutorials/developer-mesh/terminating-gateways-connect-external-services
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/geps/gep-1897/
[BackendLBPolicy]: https://gateway-api.sigs.k8s.io/geps/gep-1619/
[Gateway API]: https://gateway-api.sigs.k8s.io/
[Envoy Gateway]: ../../api/extension_types#envoygateway
[Envoy Patch Policy]: ../../api/extension_types#envoypatchpolicy
[Envoy Extension Manager]: ./extending-envoy-gateway
