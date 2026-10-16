---
title: "ListenerSets"
---

[ListenerSet][] is a Gateway API resource that defines listeners in a separate object and attaches them to a Gateway. Instead of declaring every port, hostname, and TLS certificate on the Gateway itself, teams can manage their own listeners in their own namespaces, while all listeners are served by the same Envoy Proxy fleet.

ListenerSets are useful when:

* You need more than the 64 listeners allowed on a single Gateway.
* A platform team owns the Gateway, and application teams need to self-service their own ports, hostnames, and
  TLS certificates without touching the Gateway.
* Listeners and their TLS secrets must live in a different namespace than the Gateway.

Envoy Gateway supports attaching HTTPRoute (HTTP/HTTPS), GRPCRoute, TLSRoute, TCPRoute, and UDPRoute to ListenerSet
listeners. Envoy Gateway policies, including [ClientTrafficPolicy][], [SecurityPolicy][], [BackendTrafficPolicy][], and
[EnvoyExtensionPolicy][], can target a ListenerSet, or a single listener in it via `sectionName`.

The ListenerSet CRD ships with the Gateway API CRDs installed by Envoy Gateway. No Envoy Gateway configuration is
required: the resource is reconciled whenever the CRD is present in the cluster.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

A Gateway must explicitly opt in to ListenerSet attachment through `spec.allowedListeners`. By default, no
ListenerSets are allowed. Update the Gateway from the Quickstart to accept ListenerSets from the same namespace:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  allowedListeners:
    namespaces:
      from: Same
  listeners:
  - name: http
    port: 80
    protocol: HTTP
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  allowedListeners:
    namespaces:
      from: Same
  listeners:
  - name: http
    port: 80
    protocol: HTTP
```

{{% /tab %}}
{{< /tabpane >}}

`allowedListeners.namespaces.from` also supports `All` and `Selector` to accept ListenerSets from any namespace or
from namespaces matching a label selector, which is the typical setup when application teams manage ListenerSets in
their own namespaces.

Create a ListenerSet that adds an extra HTTP listener on port 8080, and an HTTPRoute that attaches to it by
referencing the ListenerSet as its parent:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: ListenerSet
metadata:
  name: app-team-listeners
spec:
  parentRef:
    name: eg
  listeners:
  - name: http-8080
    port: 8080
    protocol: HTTP
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend-on-8080
spec:
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: ListenerSet
    name: app-team-listeners
    sectionName: http-8080
  rules:
  - backendRefs:
    - name: backend
      port: 3000
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: ListenerSet
metadata:
  name: app-team-listeners
spec:
  parentRef:
    name: eg
  listeners:
  - name: http-8080
    port: 8080
    protocol: HTTP
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend-on-8080
spec:
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: ListenerSet
    name: app-team-listeners
    sectionName: http-8080
  rules:
  - backendRefs:
    - name: backend
      port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

Verify that the ListenerSet is accepted and programmed:

```shell
kubectl get listenersets
```

The Gateway status reports the number of attached ListenerSets in `status.attachedListenerSets`.

Test the new listener:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
curl -H "Host: www.example.com" http://${GATEWAY_HOST}:8080/
```

HTTPS listeners work the same way: a ListenerSet listener can terminate TLS with `certificateRefs` pointing at
Secrets in the ListenerSet's namespace, so application teams manage their own certificates.

### Listener conflicts

Every listener needs a unique combination of port, protocol, and hostname across the Gateway and all of its
ListenerSets. When two listeners clash, the Gateway's own listeners win, and among ListenerSet listeners the
first in processing order wins.

Listeners that share a port also need compatible protocols, even if their hostnames differ. HTTPS and TLS can
share a port, and UDP can share a port with anything. If the protocols clash, the Gateway's own protocol wins.
If only ListenerSets use the port, the first listener sets the protocol, unless a single ListenerSet mixes
incompatible protocols on one port, in which case all of its listeners on that port are rejected.

Rejected listeners report `Conflicted: True` in the status, and ListenerSet listeners also get `Accepted: False`
and `Programmed: False`.

## Comparison with Merged Gateways

Before ListenerSet existed in the Gateway API, Envoy Gateway offered the custom [Merged Gateways deployment
mode][] (`mergeGateways` in the EnvoyProxy resource), which collapses the listeners of all Gateways under a
GatewayClass onto a single Envoy Proxy fleet. Both features let multiple teams share one data plane, but they work
differently:

|                       | ListenerSet                                                             | Merged Gateways (`mergeGateways`)                                        |
| --------------------- | ----------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| API                   | Gateway API-native resource, portable across implementations            | Envoy Gateway-specific EnvoyProxy setting                                 |
| Unit of sharing       | One Gateway shares its infrastructure with many ListenerSets            | All Gateways under one GatewayClass share one Envoy fleet                 |
| Opt-in                | Per Gateway, via `spec.allowedListeners`                                 | Per GatewayClass, via `mergeGateways: true` in EnvoyProxy                 |
| Delegation model      | Gateway owner controls which namespaces may attach listeners             | Any Gateway under the GatewayClass is merged, no per-Gateway consent      |
| Listener limit        | 64 per resource, scale by adding ListenerSets to one Gateway            | 64 per resource, scale by adding Gateways under the GatewayClass          |
| Status reporting      | Per-ListenerSet status, plus `attachedListenerSets` on the Gateway       | Status aggregated on each merged Gateway                                  |
| Policy attachment     | Policies can target a ListenerSet or its individual listeners            | Policies target each Gateway separately                                   |

For new deployments, prefer ListenerSet: it is the upstream Gateway API design for this use case, and it gives the Gateway owner explicit control over who may attach listeners.
Moving to ListenerSet requires converting Gateways into ListenerSets and rewriting the `parentRefs` of every route.
See [issue 1230][] for the original motivation behind `mergeGateways`.

### When Merged Gateways may still fit

ListenerSet does not yet cover every aspect of the merged Gateways model, so `mergeGateways` can still be the
better choice in some setups:

- **Tenants keep authoring plain Gateways.** With `mergeGateways`, each team owns a standard Gateway resource,which the wider ecosystem, for example cert-manager and external-dns, already understands. Tooling support for ListenerSet is still catching up.
- **No coupling to a named parent.** A ListenerSet must reference its parent Gateway by name through `spec.parentRef`, so tenants need to know, and hard-code, the shared Gateway's name and namespace. With `mergeGateways`, tenants only reference the GatewayClass.
- **Per-tenant Gateway semantics.** Each merged Gateway keeps its own status, addresses, and policy attachment
  points, so existing per-Gateway workflows keep working unchanged.

[ListenerSet]: https://gateway-api.sigs.k8s.io/reference/api-types/listenerset/
[ClientTrafficPolicy]: ../../../concepts/gateway_api_extensions/client-traffic-policy
[SecurityPolicy]: ../../../concepts/gateway_api_extensions/security-policy
[BackendTrafficPolicy]: ../../../concepts/gateway_api_extensions/backend-traffic-policy
[EnvoyExtensionPolicy]: ../../../api/extension_types#envoyextensionpolicy
[Merged Gateways deployment mode]: ../../operations/deployment-mode#merged-gateways-onto-a-single-envoyproxy-fleet
[issue 1230]: https://github.com/envoyproxy/gateway/issues/1230
