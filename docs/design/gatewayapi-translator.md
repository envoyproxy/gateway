# Gateway API Translator Design

## Assumptions
- initially target core conformance features only, to be followed by extended conformance features

## Inputs and Outputs

The main inputs to the Gateway API translator are:
- the GatewayClass to process
- Gateways, HTTPRoutes, Services, Secrets

The outputs of the Gateway API translator are:
- IR
- status updates for GatewayClass, Gateways, HTTPRoutes

## Listener Compatibility
Since Envoy Gateway handles all Gateways for a given GatewayClass, we need to determine the compatibility of _all_ Listeners across _all_ of those Gateways.

The rules are:
- for a given port number, every Listener using that port number must have a compatible protocol (either all HTTP, or all HTTPS/TLS).
- for a given port number, every Listener using that port number must have a distinct hostname (at most one Listener per port can have no hostname).

Listeners sharing a port that are not mutually compatible will be marked as "Conflicted: true" with an appropriate reason.

### Listener Compatibility Examples

#### Example 1: Gateways with compatible Listeners (same port & protocol, different hostnames)

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
      hostname: *.envoygateway.io
---
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: gateway-2
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
```

####  Example 2: Gateways with compatible Listeners (same port & protocol, one hostname specified, one not)

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
      hostname: *.envoygateway.io
---
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: gateway-2
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
```

#### Example 3: Gateways with incompatible Listeners (same port, protocol and hostname)

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
---
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: gateway-2
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
```

#### Example 4: Gateways with incompatible Listeners (neither specify a hostname)

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
---
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: gateway-2
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
```

## Computing Status

Gateway API specifies a rich set of status fields & conditions for each resource.
To be conformant, Envoy Gateway needs to compute the appropriate status fields and conditions as it's processing resources.

Status needs to be computed and set for:
- the GatewayClass (gatewayclass.status.conditions)
- each Listener for each Gateway (gateway.status.listeners)
- each Gateway, based on its Listeners' statuses (gateway.status.conditions)
- each ParentRef for each Route (route.status.parents)

The Gateway API translator will take the approach of populating status on the resources themselves as they're being processed, and then passing those statuses off to another component to persist the updates to the Kubernetes API or other backend.

## Outline

The following roughly outlines the translation process.
Each step may produce (1) IR; and (2) status updates on Gateway API resources.

```
1. Process Gateway Listeners
    - validate unique hostnames/ports/protcols
    - validate/compute supported kinds
    - validate allowed namespaces (validate selector if specified)
    - validate TLS details if specified, resolve secret ref

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
```

## Context Structs

To help store, access and manipulate information as it's processed during the translation process, a set of context structs will be used.
These structs will wrap a given Gateway API type, and add additional fields and methods to support processing.
For example, below is a partial sketch of the ListenerContext struct:

```go
type listenerContext struct {
    // The Listener.
    listener *gatewayapi_v1beta1.Listener
    
    // The Gateway this Listener belongs to.
    gateway *gatewayapi_v1beta1.Gateway
    
    // The TLS Secret for this Listener, if applicable.
    tlsSecret *corev1.Secret
}

// Sets a Listener condition on the Listener's Gateway's .status.listeners.
func (lctx *ListenerContext) SetCondition(type string, status bool, reason string, message string) {
    ...
}

// Returns whether or not the Listener allows a given Route kind.
func (lctx *ListenerContext) AllowsKind(kind gatewayapi_v1beta1.Kind) bool {
    ...
}
```

The exact specs of these structs will be worked out at implementation time.
