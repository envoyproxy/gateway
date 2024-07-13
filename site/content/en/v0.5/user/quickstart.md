---
title: "Quickstart"
weight: 1
---

This guide will help you get started with Envoy Gateway in a few simple steps.

## Prerequisites

A Kubernetes cluster.

__Note:__ Refer to the [Compatibility Matrix](/blog/2022/10/01/versions/) for supported Kubernetes versions.

## Installation

Install the Gateway API CRDs and Envoy Gateway:

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.5.0 -n envoy-gateway-system --create-namespace
```

Wait for Envoy Gateway to become available:

```shell
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

Install the GatewayClass, Gateway, HTTPRoute and example app:

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v0.5.0/quickstart.yaml -n default
```

**Note**: [`quickstart.yaml`] defines that Envoy Gateway will listen for
traffic on port 80 on its globally-routable IP address, to make it easy to use
browsers to test Envoy Gateway. When Envoy Gateway sees that its Listener is
using a privileged port (<1024), it will map this internally to an
unprivileged port, so that Envoy Gateway doesn't need additional privileges.
It's important to be aware of this mapping, since you may need to take it into
consideration when debugging.

[`quickstart.yaml`]: https://github.com/envoyproxy/gateway/releases/download/v0.5.0/quickstart.yaml

## Testing the Configuration

Get the name of the Envoy service created the by the example Gateway:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

Port forward to the Envoy service:

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8888:80 &
```

Curl the example app through Envoy proxy:

```shell
curl --verbose --header "Host: www.example.com" http://localhost:8888/get
```

### External LoadBalancer Support

You can also test the same functionality by sending traffic to the External IP. To get the external IP of the
Envoy service, run:

```shell
export GATEWAY_HOST=$(kubectl get svc/${ENVOY_SERVICE} -n envoy-gateway-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

In certain environments, the load balancer may be exposed using a hostname, instead of an IP address. If so, replace
`ip` in the above command with `hostname`.

Curl the example app through Envoy proxy:

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/get
```

## Clean-Up

Use the steps in this section to uninstall everything from the quickstart guide.

Delete the GatewayClass, Gateway, HTTPRoute and Example App:

```shell
kubectl delete -f https://github.com/envoyproxy/gateway/releases/download/v0.5.0/quickstart.yaml --ignore-not-found=true
```

Delete the Gateway API CRDs and Envoy Gateway:

```shell
helm uninstall eg -n envoy-gateway-system
```

## Next Steps

Checkout the [Developer Guide](../../contributions/develop/) to get involved in the project.
