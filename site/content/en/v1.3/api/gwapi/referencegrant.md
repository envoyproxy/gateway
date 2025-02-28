+++
title = "ReferenceGrant"
+++


    The `ReferenceGrant` resource is Beta and part of the 
    Standard Channel since `v0.6.0`. For more information on release
    channels, refer to our [versioning guide](https://gateway-api.sigs.k8s.io/concepts/versioning).

    This resource was originally named "ReferencePolicy". It was renamed
    to "ReferenceGrant" to avoid any confusion with policy attachment.

A ReferenceGrant can be used to enable cross namespace references within
Gateway API. In particular, Routes may forward traffic to backends in other
namespaces, or Gateways may refer to Secrets in another namespace.

![Reference Grant](https://gateway-api.sigs.k8s.io/images/referencegrant-simple.svg)
<!-- Source: https://docs.google.com/presentation/d/11HEYCgFi-aya7FS91JvAfllHiIlvfgcp7qpi_Azjk4E/edit#slide=id.g13c18e3a7ab_0_171 -->

In the past, we've seen that forwarding traffic across namespace boundaries is a
desired feature, but without a safeguard like ReferenceGrant,
[vulnerabilities](https://github.com/kubernetes/kubernetes/issues/103675) can
emerge.

If an object is referred to from outside its namespace, the object's owner must
create a ReferenceGrant resource to explicitly allow that reference. Without a
ReferenceGrant, a cross namespace reference is invalid.

## Structure
Fundamentally a ReferenceGrant is made up of two lists, a list of resources
references may come from, and a list of resources that may be referenced.

The `from` list allows you to specify the group, kind, and namespace of
resources that may reference items described in the `to` list.

The `to` list allows you to specify the group and kind of resources that may be
referenced by items described in the `from` list. The namespace is not necessary
in the `to` list because a ReferenceGrant can only be used to allow references
to resources in the same namespace as the ReferenceGrant.

## Example
The following example shows how a HTTPRoute in namespace `foo` can reference a
Service in namespace `bar`. In this example a ReferenceGrant in the `bar`
namespace explicitly allows references to Services from HTTPRoutes in the `foo`
namespace.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: foo
  namespace: foo
spec:
  rules:
  - matches:
    - path: /bar
    backendRefs:
      - name: bar
        namespace: bar
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: ReferenceGrant
metadata:
  name: bar
  namespace: bar
spec:
  from:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    namespace: foo
  to:
  - group: ""
    kind: Service
```

## API design decisions
While the API is simplistic in nature, it comes with a few notable decisions:

1. Each ReferenceGrant only supports a single From and To section. Additional
   trust relationships must be modeled with additional ReferenceGrant
   resources.
1. Resource names are intentionally excluded from the "From" section of
   ReferenceGrant because they rarely provide any meaningful protection. A user
   that is able to write to resources of a certain kind within a namespace can
   always rename resources or change the structure of the resources to match a
   given grant.
1. A single Namespace is allowed per "From" struct. Although a selector would be
   more powerful, it encourages unnecessarily insecure configuration.
1. The effect of these resources is purely additive, they stack on top of each
   other. This makes it impossible for them to conflict with each other.

Please see the [API
Specification](https://gateway-api.sigs.k8s.io/reference/spec#gateway.networking.k8s.io/v1alpha2.ReferenceGrant)
for more details on how specific ReferenceGrant fields are interpreted.

## Implementation Guidelines
This API relies on runtime verification. Implementations MUST watch for changes
to these resources and recalculate the validity of cross-namespace references
after each change or deletion.

When communicating the status of a cross-namespace reference, implementations
MUST NOT expose information about the existence of a resource in another
namespace unless a ReferenceGrant exists allowing the reference to occur. This
means that if a cross-namespace reference is made without a ReferenceGrant to a
resource that doesn't exist, any status conditions or warning messages need to
focus on the fact that a ReferenceGrant does not exist to allow this reference.
No hints should be provided about whether or not the referenced resource exists.

## Exceptions
Cross namespace Route -> Gateway binding follows a slightly different pattern
where the handshake mechanism is built into the Gateway resource. For more
information on that approach, refer to the relevant [Security Model
documentation](https://gateway-api.sigs.k8s.io/concepts/security-model). Although conceptually similar to
ReferenceGrant, this configuration is built directly into Gateway Listeners,
and allows for fine-grained per Listener configuration that would not be
possible with ReferenceGrant.

There are some situations where it MAY be acceptable to ignore ReferenceGrant
in favor of some other security mechanism. This MAY only be done if other
mechanisms like NetworkPolicy can effectively limit cross-namespace references
by the implementation.

An implementation choosing to make this exception MUST clearly document that
ReferenceGrant is not honored by their implementations and detail which
alternative safeguards are available. Note that this is unlikely to apply to
ingress implementations of the API and will not apply to all mesh
implementations.

For an example of the risks involved in cross-namespace references, refer to
[CVE-2021-25740](https://github.com/kubernetes/kubernetes/issues/103675).
Implementations of this API need to be very careful to avoid confused deputy
attacks. ReferenceGrant provides a safeguard for that. Exceptions MUST only be
made by implementations that are absolutely certain that other equally effective
safeguards are in place.

## Conformance Level
ReferenceGrant support is a "CORE" conformance level requirement for
cross-namespace references that originate from the following objects:

- Gateway
- GRPCRoute
- HTTPRoute
- TLSRoute
- TCPRoute
- UDPRoute

That is, all implementations MUST use this flow for any cross namespace
references in the Gateway and any of the core xRoute types, except as noted
in the Exceptions section above.

Other "ImplementationSpecific" objects and references MUST also use this flow
for cross-namespace references, except as noted in the Exceptions section above.

## Potential Future API Group Change

ReferenceGrant is starting to gain interest outside of Gateway API and SIG
Network use cases. It is possible that this resource may move to a more neutral
home. Users of the ReferenceGrant API may be required to transition to a
different API Group (instead of `gateway.networking.k8s.io`) at some point in
the future.
