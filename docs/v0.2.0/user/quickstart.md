# Quickstart

This guide will help you get started with Envoy Gateway in a few simple steps.

## Prerequisites

A Kubernetes cluster.

__Note:__ Envoy Gateway is tested against Kubernetes v1.24.0.

## Installation

Install the Gateway API CRDs and Envoy Gateway:

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0/install.yaml
```

Install the GatewayClass, Gateway, HTTPRoute and example app:

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0/quickstart.yaml
```

## Testing the Configuration

Get the name of the Envoy service created the by the example Gateway:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

Port forward to the Envoy service:

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8888:8080 &
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
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST:8080/get
```

## Clean-Up

Use the steps in this section to uninstall everything from the quickstart guide.

Delete the GatewayClass, Gateway, HTTPRoute and Example App:

```shell
kubectl delete -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0/quickstart.yaml --ignore-not-found=true
```

Delete the Gateway API CRDs and Envoy Gateway:

```shell
kubectl delete -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0/install.yaml --ignore-not-found=true
```

## Next Steps

Checkout the [Developer Guide](../dev/README.md) to get involved in the project.
