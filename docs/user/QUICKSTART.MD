# Introduction
This guide will help you get started with Envoy Gateway in a few simple steps.

## Prerequisites
A Kubernetes cluster with support for services of type `LoadBalancer` or a [Kind][kind] cluster.

__Note:__ Envoy Gateway is tested against Kubernetes v1.24.1.

## Installation
Install the Gateway API CRDs:
```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0-rc1/gatewayapi-crds.yaml
```

Run Envoy Gateway:
```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0-rc1/install.yaml
```

Run the example app:
```shell
kubectl apply -f https://raw.githubusercontent.com//envoyproxy/gateway/tree/v0.2.0-rc1/examples/kubernetes/httpbin.yaml
```

Due to [Issue 301][issue_301], Gateway API resources must be created in the following order. First, create the
GatewayClass:
```shell
kubectl apply -f https://raw.githubusercontent.com//envoyproxy/gateway/tree/v0.2.0-rc1/examples/kubernetes/gatewayclass.yaml
```

Create the Gateway:
```shell
kubectl apply -f https://raw.githubusercontent.com//envoyproxy/gateway/tree/v0.2.0-rc1/examples/kubernetes/gateway.yaml
```

Create the HTTPRoute to route traffic through Envoy proxy to the example app:
```shell
kubectl apply -f https://raw.githubusercontent.com//envoyproxy/gateway/tree/v0.2.0-rc1/examples/kubernetes/httproute.yaml
```

### For Kind Clusters
In a separate terminal, port forward the Envoy service:
```shell
kubectl -n envoy-gateway-system port-forward service/envoy-eg 8888:8080
```

Curl the example app through Envoy proxy:
```shell
curl --verbose --header "Host: www.example.com" http://localhost:8888/
```

### For Kubernetes Clusters
Get the external IP of the Envoy service:
```shell
export GATEWAY_HOST=$(kubectl svc/envoy-eg -n envoy-gateway-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

In certain environments, the load balancer may be exposed using a hostname, instead of an IP address. If so, replace
`ip` in the above command with `hostname`.

Curl the example app through Envoy proxy:
```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST:8888/
```

[issue_301]: https://github.com/envoyproxy/gateway/issues/301
[kind]: https://kind.sigs.k8s.io/
