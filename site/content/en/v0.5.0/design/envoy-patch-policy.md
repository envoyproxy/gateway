---
title: "EnvoyPatchPolicy"
---

## Overview

This design introduces the `EnvoyPatchPolicy` API allowing users to modify the generated Envoy xDS Configuration
that Envoy Gateway generates before sending it to Envoy Proxy.

Envoy Gateway allows users to configure networking and security intent using the
upstream [Gateway API][] as well as implementation specific [Extension APIs][] defined in this project
to provide a more batteries included experience for application developers.
* These APIs are an abstracted version of the underlying Envoy xDS API to provide a better user experience for the application developer, exposing and setting only a subset of the fields for a specific feature, sometimes in a opinionated way (e.g [RateLimit][])
* These APIs do not expose all the features capabilities that Envoy has either because these features are desired but the API
is not defined yet or the project cannot support such an extensive list of features.
To alleviate this problem, and provide an interim solution for a small section of advanced users who are well versed in
Envoy xDS API and its capabilities, this API is being introduced.

## Goals
* Add an API allowing users to modify the generated xDS Configuration

## Non Goals
* Support multiple patch mechanisims

## Implementation
`EnvoyPatchPolicy` is a [Direct Policy Attachment][] type API that can be used to extend [Gateway API][]
Modifications to the generated xDS configuration can be provided as a JSON Patch which is defined in
[RFC 6902][]. This patching mechanism has been adopted in [Kubernetes][] as well as [Kustomize][] to update
resource objects.

### Example
Here is an example highlighting how a user can configure global ratelimiting using an external rate limit service using this API.

```
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1beta1
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
apiVersion: gateway.networking.k8s.io/v1beta1
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
kind: EnvoyPatchPolicy
metadata:
  name: ratelimit-patch-policy
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
  type: JSONPatch
  jsonPatches:
    - type: "type.googleapis.com/envoy.config.listener.v3.Listener"
      # The listener name is of the form <GatewayNamespace>/<GatewayName>/<GatewayListenerName>
      name: default/eg/http
      operation:
        op: add
        path: "/default_filter_chain/filters/0/typed_config/http_filters/0"
        value:
          name: "envoy.filters.http.ratelimit"
          typed_config:
            "@type": "type.googleapis.com/envoy.extensions.filters.http.ratelimit.v3.RateLimit"
            domain: "eag-ratelimit"
            failure_mode_deny: true
            timeout: 1s
            rate_limit_service:
              grpc_service:
                envoy_grpc:
                  cluster_name: rate-limit-cluster
              transport_api_version: V3
    - type: "type.googleapis.com/envoy.config.route.v3.RouteConfiguration"
      # The route name is of the form <GatewayNamespace>/<GatewayName>/<GatewayListenerName>
      name: default/eg/http
      operation:
        op: add
        path: "/virtual_hosts/0/rate_limits"
        value:
          - actions:
              - remote_address: {}
    - type: "type.googleapis.com/envoy.config.cluster.v3.Cluster"
      name: rate-limit-cluster
      operation:
        op: add
        path: ""
        value:
          name: rate-limit-cluster
          type: STRICT_DNS
          connect_timeout: 10s
          lb_policy: ROUND_ROBIN
          http2_protocol_options: {}
          load_assignment:
            cluster_name: rate-limit-cluster
            endpoints:
              - lb_endpoints:
                  - endpoint:
                      address:
                        socket_address:
                          address: ratelimit.svc.cluster.local
                          port_value: 8081
```


## Verification
* Offline - Leverage [egctl x translate][] to ensure that the `EnvoyPatchPolicy` can be successfully applied and the desired
output xDS is created.
* Runtime - Use the `Status` field within `EnvoyPatchPolicy` to highlight whether the patch was applied successfully or not.

## State of the World
* Istio - Supports the [EnvoyFilter][] API which allows users to customize the output xDS using patches and proto based merge
semantics.

## Design Decisions
* This API will only support a single `targetRef` and can bind to only a `Gateway` resource. This simplifies reasoning of how
patches will work.
* This API will always be an experimental API and cannot be graduated into a stable API because Envoy Gateway cannot garuntee
  * that the naming scheme for the generated resources names will not change across releases
  * that the underlying Envoy Proxy API will not change across releases
* This API needs to be explicitly enabled using the [EnvoyGateway][] API

## Open Questions
* Should the value only support JSON or YAML as well (which is a JSON superset) ?

## Alternatives
* Users can customize the Envoy [Bootstrap configuration using EnvoyProxy API][] and provide static xDS configuration.
* Users can extend functionality by [Extending the Control Plane][] and adding gRPC hooks to modify the generated xDS configuration.



[Direct Policy Attachment]: https://gateway-api.sigs.k8s.io/references/policy-attachment/#direct-policy-attachment 
[RFC 6902]: https://datatracker.ietf.org/doc/html/rfc6902
[Gateway API]: https://gateway-api.sigs.k8s.io/
[Kubernetes]: https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/
[Kustomize]: https://github.com/kubernetes-sigs/kustomize/blob/master/examples/jsonpatch.md
[Extension APIs]: https://gateway.envoyproxy.io/latest/api/extension_types.html
[RateLimit]: https://gateway.envoyproxy.io/latest/user/rate-limit.html
[EnvoyGateway]: https://gateway.envoyproxy.io/latest/api/config_types.html#envoygateway
[Extending the Control Plane]: https://gateway.envoyproxy.io/latest/design/extending-envoy-gateway.html
[EnvoyFilter]: https://istio.io/latest/docs/reference/config/networking/envoy-filter
[egctl x translate]: https://gateway.envoyproxy.io/latest/user/egctl.html#egctl-experimental-translate
[Bootstrap configuration using EnvoyProxy API]: https://gateway.envoyproxy.io/latest/user/customize-envoyproxy.html#customize-envoyproxy-bootstrap-config
