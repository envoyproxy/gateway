---
title: "Backend"
---

## Overview

This design document introduces the `Backend` API allowing system administrators to represent backends without the use 
of a K8s `Service` resource. 

Common use cases for non-Service backends in the K8s and Envoy ecosystem include:
- Cluster-external endpoints, which are currently second-class citizens in Gateway-API 
  (supported using [Services and FQDN endpoints][]).
- Host-local endpoints, such as sidecars or daemons that listen on [unix domain sockets][] or envoy [internal listeners][], 
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
Here is an example highlighting how a user can configure a route that forwards traffic to both a K8s Service and a Backend 
that has both unix domain socket and ip endpoints. A [BackendTLSPolicy][] is attached to the backend resource, enabling TLS.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: backend
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
  name: backend-mixed-ip-uds
spec:
  appProtocols: 
    - gateway.envoyproxy.io/h2c
  endpoints:
    - unix:
        path: /var/run/backend.sock   
    - ip:
        address: 10.244.0.28
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
        - group: gateway.envoyproxy.io
          kind: Backend
          name: backend-mixed-ip-uds
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
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: BackendTLSPolicy
metadata:
  name: policy-btls
spec:
  targetRef:
    group: gateway.envoyproxy.io
    kind: Backend
    name: backend-mixed-ip-uds
  tls:
    caCertRefs:
      - name: backend-tls-checks-certificate
        group: ''
        kind: ConfigMap
    hostname: example.com
```

## Design Decisions
* All instances of `BackendObjectReference` in Envoy Gateway MAY support referencing the `Backend` kind.
* For security reasons, Envoy Gateway MUST reject references to a `Backend` in xRoute resources. For example, UDS and 
  localhost references will not be supported for xRoutes.  
* All attributes of the Envoy Gateway extended `BackendRef` resource MUST be implemented for the `Backend` resource.  
* A `Backend` resource referenced by `BackendObjectReference` will be translated to Envoy Gateway's IR DestinationSetting.
  As such, all `BackendAdresses` are treated as equivalent endpoints with identical weights, TLS settings, etc.  
* Gateway-API and Envoy Gateway policies that attach to Services ([BackendTLSPolicy][], [BackendLBPolicy][]) 
  MUST support attachment to the `Backend` resource in Envoy Gateway. 
* Policy attachment to a named section of the `Backend` resource is not supported at this time. Currently, 
  `BackendObjectReference` can only select ports, and not generic section names. Hence, a named section of `Backend` 
  cannot be referenced by routes, and so attachment of policies to named sections will create translation ambiguity. 
  Users that wish to attach policies to some of the `BackendAddresses` in a `Backend` resource can use multiple `Backend` 
  resources and pluralized `BackendRefs` instead. 
* The `Backend` API SHOULD support other Gateway-API backend features, such as [Backend Protocol Selection][]. 
  Translation of explicit upstream application protocol setting SHOULD be consistent with the existing implementation for
  `Service` resources. 
* The `Backend` upstream transport protocol (TCP, UDP) is inferred from the xRoute kind: TCP is inferred for all routes 
  except for `UDPRoute` which is resolved to UDP.    
* This API resource MUST be part of same namespace as the targetRef resource. The `Backend` API MUST be subject to 
  the same cross-namespace reference restriction as referenced `Service` resources.    
* The `Backend` resource translation MUST NOT modify Infrastructure. Any change to infrastructure that is required to 
  achieve connectivity to a backend (mounting a socket, adding a sidecar container, modifying a network policy, ...) 
  MUST be implemented with an appropriate infrastructure patch in the [EnvoyProxy][] API. 
* To limit the overall maintenance effort related to supporting of non-Service backends, the `Backend` API SHOULD 
  support multiple generic address types (UDS, FQDN, IPv4, IPv6), and MUST NOT support vendor-specific backend types.
* Both `Backend` and `Service` resources may appear in the same `BackendRefs` list.
* The Optional `Port` field SHOULD NOT be evaluated when referencing a `Backend`.  
* Referenced `Backend` resources MUST be translated to envoy endpoints, similar to the current `Service` translation.
* Certain combinations of `Backend` and `Service` are incompatible. For example, a Unix Domain Socket and a FQDN service
  require different cluster service discovery types (Static/EDS and Strict-DNS accordingly).
* If a Backend that is referenced by a route cannot be translated, the `Route` resource will have an `Accepted=False` 
  condition with a `UnsupportedValue` reason. 
* This API needs to be explicitly enabled using the [EnvoyGateway][] API   
  
## Alternatives
* The project can indefinitely wait for these configuration parameters to be part of the [Gateway API][].
* Users can leverage the existing [Envoy Patch Policy][] or [Envoy Extension Manager][] to inject custom envoy clusters
  and route configuration. However, these features require a high level of envoy expertise, investment and maintenance. 

[BackendObjectReference Extension]: https://gateway-api.sigs.k8s.io/guides/migrating-from-ingress/?h=extensi#approach-to-extensibility
[internal listeners]: https://www.envoyproxy.io/docs/envoy/latest/configuration/other_features/internal_listener
[unix domain sockets]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#envoy-v3-api-msg-config-core-v3-pipe
[Resource Backends]: https://kubernetes.io/docs/concepts/services-networking/ingress/#resource-backend
[Services and FQDN endpoints]: ./../../latest/tasks/traffic/routing-outside-kubernetes.md
[Service Entry]: https://istio.io/latest/docs/reference/config/networking/service-entry/
[Upstream]: https://docs.solo.io/gloo-edge/1.7.23/reference/api/github.com/solo-io/gloo/projects/gloo/api/v1/upstream.proto.sk/
[External Services]: https://developer.hashicorp.com/consul/tutorials/developer-mesh/terminating-gateways-connect-external-services
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/geps/gep-1897/
[BackendLBPolicy]: https://gateway-api.sigs.k8s.io/geps/gep-1619/
[Backend Protocol Selection]: https://gateway-api.sigs.k8s.io/geps/gep-1911/
[EnvoyProxy]:../../latest/api/extension_types#envoyproxy
[EnvoyGateway]: ../../latest/api/extension_types#envoygateway
[Gateway API]: https://gateway-api.sigs.k8s.io/
[Envoy Patch Policy]: ../../latest/api/extension_types#envoypatchpolicy
[Envoy Extension Manager]: ./extending-envoy-gateway
