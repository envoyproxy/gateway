# Installation

## Prerequisites

A Kubernetes cluster.

__Note:__ Refer to the [Compatibility Matrix](../intro/compatibility.rst) for supported Kubernetes versions.

## Installation

Install the Gateway API CRDs and Envoy Gateway:

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace
```

Wait for Envoy Gateway to become available:

```shell
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

Install the GatewayClass, Gateway, HTTPRoute and example app:

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/latest/quickstart.yaml -n default
```

**Note**: [`quickstart.yaml`] defines that Envoy Gateway will listen for
traffic on port 80 on its globally-routable IP address, to make it easy to use
browsers to test Envoy Gateway. When Envoy Gateway sees that its Listener is
using a privileged port (<1024), it will map this internally to an
unprivileged port, so that Envoy Gateway doesn't need additional privileges.
It's important to be aware of this mapping, since you may need to take it into
consideration when debugging.

[`quickstart.yaml`]: https://github.com/envoyproxy/gateway/releases/download/latest/quickstart.yaml


## Open Ports


| Envoy Gateway           |  Port       |
|:----------------------: | :---------: |
| Xds EnvoyProxy Server   | 18000       |
| Xds RateLimit Server    | 18001       |
| Admin Server            | 19000       |


| Envoy Proxy          | Port    |
|:-----------------:   | :-----: |
| Admin Server         | 19000   |
| Heath Check Listener | 19001   |