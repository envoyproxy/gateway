---
title: "Gateway API Translator Design"
weight: 4
---

The Gateway API translates external resources, e.g. GatewayClass, from the configured Provider to the Intermediate
Representation (IR).

## Assumptions

Initially target core conformance features only, to be followed by extended conformance features.

## Inputs and Outputs

The main inputs to the Gateway API translator are:

- GatewayClass, Gateway, HTTPRoute, TLSRoute, Service, ReferenceGrant, Namespace, and Secret resources.

__Note:__ ReferenceGrant is not fully implemented as of v0.2.

The outputs of the Gateway API translator are:

- Xds and Infra Internal Representations (IRs).
- Status updates for GatewayClass, Gateways, HTTPRoutes

## Listener Compatibility

Envoy Gateway follows Gateway API listener compatibility spec:
> Each listener in a Gateway must have a unique combination of Hostname, Port, and Protocol. An implementation MAY group
> Listeners by Port and then collapse each group of Listeners into a single Listener if the implementation determines
> that the Listeners in the group are “compatible”.

__Note:__ Envoy Gateway does not collapse listeners across multiple Gateways.

### Listener Compatibility Examples

#### Example 1: Gateway with compatible Listeners (same port & protocol, different hostnames)

```yaml
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: gateway-1
  namespace: envoy-gateway
spec:
  gatewayClassName: envoy-gateway
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
      hostname: "*.envoygateway.io"
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
      hostname: whales.envoygateway.io
```

#### Example 2: Gateway with compatible Listeners (same port & protocol, one hostname specified, one not)

```yaml
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: gateway-1
  namespace: envoy-gateway
spec:
  gatewayClassName: envoy-gateway
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
      hostname: "*.envoygateway.io"
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
```

#### Example 3: Gateway with incompatible Listeners (same port, protocol and hostname)

```yaml
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: gateway-1
  namespace: envoy-gateway
spec:
  gatewayClassName: envoy-gateway
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
      hostname: whales.envoygateway.io
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
      hostname: whales.envoygateway.io
```

#### Example 4: Gateway with incompatible Listeners (neither specify a hostname)

```yaml
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: gateway-1
  namespace: envoy-gateway
spec:
  gatewayClassName: envoy-gateway
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
```

## Computing Status

Gateway API specifies a rich set of status fields & conditions for each resource. To achieve conformance, Envoy Gateway
must compute the appropriate status fields and conditions for managed resources.

Status is computed and set for:

- The managed GatewayClass (`gatewayclass.status.conditions`).
- Each managed Gateway, based on its Listeners' status (`gateway.status.conditions`). For the Kubernetes provider, the
  Envoy Deployment and Service status are also included to calculate Gateway status.
- Listeners for each Gateway (`gateway.status.listeners`).
- The ParentRef for each Route (`route.status.parents`).

The Gateway API translator is responsible for calculating status conditions while translating Gateway API resources to
the IR and publishing status over the [message bus][]. The Status Manager subscribes to these status messages and
updates the resource status using the configured provider. For example, the Status Manager uses a Kubernetes client to
update resource status on the Kubernetes API server.

## Outline

The following roughly outlines the translation process. Each step may produce (1) IR; and (2) status updates on Gateway
API resources.

1. Process Gateway Listeners
    - Validate unique hostnames, ports, and protocols.
    - Validate and compute supported kinds.
    - Validate allowed namespaces (validate selector if specified).
    - Validate TLS fields if specified, including resolving referenced Secrets.

2. Process HTTPRoutes
    - foreach route rule:
        - compute matches
            - [core] path exact, path prefix
            - [core] header exact
            - [extended] query param exact
            - [extended] HTTP method
        - compute filters
            - [core] request header modifier (set/add/remove)
            - [core] request redirect (hostname, statuscode)
            - [extended] request mirror
        - compute backends
            - [core] Kubernetes services
    - foreach route parent ref:
        - get matching listeners (check Gateway, section name, listener validation status, listener allowed routes, hostname intersection)
        - foreach matching listener:
            - foreach hostname intersection with route:
                - add each computed route rule to host

## Context Structs

To help store, access and manipulate information as it's processed during the translation process, a set of context
structs are used. These structs wrap a given Gateway API type, and add additional fields and methods to support
processing.

`GatewayContext` wraps a Gateway and provides helper methods for setting conditions, accessing Listeners, etc.

```go
type GatewayContext struct {
	// The managed Gateway
	*v1beta1.Gateway

	// A list of Gateway ListenerContexts.
	listeners []*ListenerContext
}
```

`ListenerContext` wraps a Listener and provides helper methods for setting conditions and other status information on
the associated Gateway.

```go
type ListenerContext struct {
    // The Gateway listener.
	*v1beta1.Listener

	// The Gateway this Listener belongs to.
	gateway           *v1beta1.Gateway

	// An index used for managing this listener in the list of Gateway listeners.
	listenerStatusIdx int

	// Only Routes in namespaces selected by the selector may be attached
	// to the Gateway this listener belongs to.
	namespaceSelector labels.Selector

	// The TLS Secret for this Listener, if applicable.
	tlsSecret         *v1.Secret
}
```

`RouteContext` represents a generic Route object (HTTPRoute, TLSRoute, etc.) that can reference Gateway objects.

```go
type RouteContext interface {
	client.Object

	// GetRouteType returns the Kind of the Route object, HTTPRoute,
	// TLSRoute, TCPRoute, UDPRoute etc.
	GetRouteType() string

	// GetHostnames returns the hosts targeted by the Route object.
	GetHostnames() []string

	// GetParentReferences returns the ParentReference of the Route object.
	GetParentReferences() []v1beta1.ParentReference

	// GetRouteParentContext returns RouteParentContext by using the Route
	// objects' ParentReference.
	GetRouteParentContext(forParentRef v1beta1.ParentReference) *RouteParentContext
}
```

[message bus]: watching.md
