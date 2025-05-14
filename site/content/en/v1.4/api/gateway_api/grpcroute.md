+++
title = "GRPCRoute"
+++


    The `GRPCRoute` resource is GA and has been part of the Standard Channel since
    `v1.1.0`. For more information on release channels, refer to our [versioning
    guide](../concepts/versioning).

[GRPCRoute][grpcroute] is a Gateway API type for specifying routing behavior
of gRPC requests from a Gateway listener to an API object, i.e. Service.

## Background

While it is possible to route gRPC with `HTTPRoutes` or via custom, out-of-tree
CRDs, in the long run, this leads to a fragmented ecosystem.

gRPC is a [popular RPC framework adopted widely across the industry](https://grpc.io/about/#whos-using-grpc-and-why).
The protocol is used pervasively within the Kubernetes project itself as the basis for
many interfaces, including:

- [the CSI](https://github.com/container-storage-interface/spec/blob/5b0d4540158a260cb3347ef1c87ede8600afb9bf/spec.md),
- [the CRI](https://github.com/kubernetes/cri-api/blob/49fe8b135f4556ea603b1b49470f8365b62f808e/README.md),
- [the device plugin framework](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/)

Given gRPC's importance in the application-layer networking space and to
the Kubernetes project in particular, the determination was made not to allow
the ecosystem to fragment unnecessarily.

### Encapsulated Network Protocols

In general, when it is possible to route an encapsulated protocol at a lower
level, it is acceptable to introduce a route resource at the higher layer when
the following criteria are met:

- Users of the encapsulated protocol would miss out on significant conventional features from their ecosystem if forced to route at a lower layer.
- Users of the encapsulated protocol would experience a degraded user experience if forced to route at a lower layer.
- The encapsulated protocol has a significant user base, particularly in the Kubernetes community.

gRPC meets all of these criteria, so the decision was made to include `GRPCRoute`in Gateway API.

### Cross Serving

Implementations that support GRPCRoute must enforce uniqueness of
hostnames between `GRPCRoute`s and `HTTPRoute`s. If a route (A) of type `HTTPRoute` or
`GRPCRoute` is attached to a Listener and that listener already has another Route (B) of
the other type attached and the intersection of the hostnames of A and B is
non-empty, then the implementation must reject Route A. That is, the
implementation must raise an 'Accepted' condition with a status of 'False' in
the corresponding RouteParentStatus.

In general, it is recommended that separate hostnames be used for gRPC and
non-gRPC HTTP traffic. This aligns with standard practice in the gRPC community.
If however, it is a necessity to serve HTTP and gRPC on the same hostname with
the only differentiator being URI, the user should use `HTTPRoute` resources for
both gRPC and HTTP. This will come at the cost of the improved UX of the
`GRPCRoute` resource.

## Spec

The specification of a GRPCRoute consists of:

- [ParentRefs][parentRef]- Define which Gateways this Route wants to be attached
  to.
- [Hostnames][hostname] (optional)- Define a list of hostnames to use for
  matching the Host header of gRPC requests.
- [Rules][grpcrouterule]- Define a list of rules to perform actions against
  matching gRPC requests. Each rule consists of [matches][matches],
  [filters][filters] (optional), [backendRefs][backendRef] (optional), and
  [name][name] (optional) fields.

<!--- Editable SVG available at site-src/images/grpcroute-basic-example.svg -->
The following illustrates a GRPCRoute that sends all traffic to one Service:
![grpcroute-basic-example](https://gateway-api.sigs.k8s.io/images/grpcroute-basic-example.png)

### Attaching to Gateways

Each Route includes a way to reference the parent resources it wants to attach
to. In most cases, that's going to be Gateways, but there is some flexibility
here for implementations to support other types of parent resources.

The following example shows how a Route would attach to the `acme-lb` Gateway:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: grpcroute-example
spec:
  parentRefs:
  - name: acme-lb
```

Note that the target Gateway needs to allow GRPCRoutes from the route's
namespace to be attached for the attachment to be successful.

### Hostnames

Hostnames define a list of hostnames to match against the Host header of the
gRPC request. When a match occurs, the GRPCRoute is selected to perform request
routing based on rules and filters (optional). A hostname is the fully qualified
domain name of a network host, as defined by [RFC 3986][rfc-3986]. Note the
following deviations from the “host” part of the URI as defined in the RFC:

- IPs are not allowed.
- The : delimiter is not respected because ports are not allowed.

Incoming requests are matched against hostnames before the GRPCRoute rules are
evaluated. If no hostname is specified, traffic is routed based on GRPCRoute
rules and filters (optional).

The following example defines hostname "my.example.com":

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: grpcroute-example
spec:
  hostnames:
  - my.example.com
```

### Rules

Rules define semantics for matching an gRPC requests based on conditions,
optionally executing additional processing steps, and optionally forwarding
the request to an API object.

#### Matches

Matches define conditions used for matching an gRPC requests. Each match is
independent, i.e. this rule will be matched if any single match is satisfied.

Take the following matches configuration as an example:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
...
matches:
  - method:
      service: com.example.User
      method: Login
    headers:
    - name: version
      value: "2"
  - method:
      service: com.example.v2.User
      method: Login
```

For a request to match against this rule, it must satisfy EITHER of the
following conditions:

 - The `com.example.User.Login` method **AND** contains the header "version: 2"
 - The `com.example.v2.User.Login` method.

If no matches are specified, the default is to match every gRPC request.

#### Filters (optional)

Filters define processing steps that must be completed during the request or
response lifecycle. Filters act as an extension point to express additional
processing that may be performed in Gateway implementations. Some examples
include request or response modification, implementing authentication
strategies, rate-limiting, and traffic shaping.

The following example adds header "my-header: foo" to gRPC requests with Host
header "my.filter.com". Note that GRPCRoute uses HTTPRoute filters for features
with functionality identical to HTTPRoute, such as this.

```yaml
#$ Used in:
#$ - site-src/api-types/grpcroute.md
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: grpc-filter-1
spec:
  hostnames:
    - my.filter.com
  rules:
    - filters:
        - type: RequestHeaderModifier
          requestHeaderModifier:
            add:
              - name: my-header
                value: foo
      backendRefs:
        - name: my-filter-svc1
          weight: 1
          port: 50051

```

API conformance is defined based on the filter type. The effects of ordering
multiple behaviors are currently unspecified. This may change in the future
based on feedback during the alpha stage.

Conformance levels are defined by the filter type:

 - All "core" filters MUST be supported by implementations supporting GRPCRoute.
 - Implementers are encouraged to support "extended" filters.
 - "Implementation-specific" filters have no API guarantees across implementations.

Specifying a core filter multiple times has unspecified or custom conformance.

If an implementation cannot support a combinations of filters, they must clearly
document that limitation. In cases where incompatible or unsupported
filters are specified and cause the `Accepted` condition to be set to status
`False`, implementations may use the `IncompatibleFilters` reason to specify
this configuration error.

#### BackendRefs (optional)

BackendRefs defines the API objects to which matching requests should be sent. If
unspecified, the rule performs no forwarding. If unspecified and no filters
are specified that would result in a response being sent, an `UNIMPLEMENTED` error code
is returned.

The following example forwards gRPC requests for the method `User.Login` to service
"my-service1" on port `50051` and gRPC requests for the method `Things.DoThing` with
header `magic: foo` to service "my-service2" on port `50051`:

```yaml
#$ Used in:
#$ - site-src/api-types/grpcroute.md
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: example
spec:
  controllerName: acme.io/gateway-controller
  parametersRef:
    name: example
    group: acme.io
    kind: Parameters
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: my-gateway
spec:
  gatewayClassName: example
  listeners:  # Use GatewayClass defaults for listener definition.
  - name: https
    protocol: HTTPS
    port: 50051
    tls:
      certificateRefs:
      - kind: Secret
        group: ""
        name: example-com-cert
---
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: grpc-app-1
spec:
  parentRefs:
  - name: my-gateway
  hostnames:
  - "example.com"
  rules:
  - matches:
    - method:
        service: com.example.User
        method: Login
    backendRefs:
    - name: my-service1
      port: 50051
  - matches:
    - headers:
      - type: Exact
        name: magic
        value: foo
      method:
        service: com.example.Things
        method: DoThing
    backendRefs:
    - name: my-service2
      port: 50051

```

The following example uses the `weight` field to forward 90% of gRPC requests to
`foo.example.com` to the "foo-v1" Service and the other 10% to the "foo-v2"
Service:

```yaml
#$ Used in:
#$ - site-src/api-types/grpcroute.md
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: foo-route
  labels:
    gateway: prod-web-gw
spec:
  hostnames:
  - foo.example.com
  rules:
  - backendRefs:
    - name: foo-v1
      port: 50051
      weight: 90
    - name: foo-v2
      port: 50051
      weight: 10

```

Reference the [backendRef][backendRef] API documentation for additional details
on `weight` and other fields.

#### Name (optional)


    This concept has been part of the Experimental Channel since `v1.2.0`.
    For more information on release channels, refer to our
    [versioning guide](../concepts/versioning).

GRPCRoute Rules include an optional `name` field. The applications for the name of a route rule are implementation-specific. It can be used to reference individual route rules by name from other resources, such as in the `sectionName` field of metaresources ([GEP-2648](../geps/gep-2648/index.md#apply-policies-to-sections-of-a-resource)), in the status stanzas of resources related to the route object, to identify internal configuration objects generated by the implementation from GRPCRoute Rule, etc.

If specified, the value of the name field must comply with the [`SectionName`](https://github.com/kubernetes-sigs/gateway-api/blob/v1.0.0/apis/v1/shared_types.go#L607-L624) type.

## Status

Status defines the observed state of the GRPCRoute.

### RouteStatus

RouteStatus defines the observed state that is required across all route types.

#### Parents

Parents define a list of the Gateways (or other parent resources) that are
associated with the GRPCRoute, and the status of the GRPCRoute with respect to
each of these Gateways. When a GRPCRoute adds a reference to a Gateway in
parentRefs, the controller that manages the Gateway should add an entry to this
list when the controller first sees the route and should update the entry as
appropriate when the route is modified.

## Examples

The following example indicates GRPCRoute "grpc-example" has been accepted by
Gateway "gw-example" in namespace "gw-example-ns":

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: grpc-example
...
status:
  parents:
  - parentRefs:
      name: gw-example
      namespace: gw-example-ns
    conditions:
    - type: Accepted
      status: "True"
```

## Merging
Multiple GRPCRoutes can be attached to a single Gateway resource. Importantly,
only one Route rule may match each request. For more information on how conflict
resolution applies to merging, refer to the [API specification][grpcrouterule].

[grpcroute]: ../reference/spec#gateway.networking.k8s.io/v1.GRPCRoute
[grpcrouterule]: ../reference/spec#gateway.networking.k8s.io/v1.GRPCRouteRule
[hostname]: ../reference/spec#gateway.networking.k8s.io/v1.Hostname
[rfc-3986]: https://tools.ietf.org/html/rfc3986
[matches]: ../reference/spec#gateway.networking.k8s.io/v1.GRPCRouteMatch
[filters]: ../reference/spec#gateway.networking.k8s.io/v1.GRPCRouteFilter
[backendRef]: ../reference/spec#gateway.networking.k8s.io/v1.GRPCBackendRef
[parentRef]: ../reference/spec#gateway.networking.k8s.io/v1.ParentRef
[name]: ../reference/spec#gateway.networking.k8s.io/v1.SectionName
