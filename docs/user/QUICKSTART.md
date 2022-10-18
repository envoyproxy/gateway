# Quickstart

This guide will help you get started with Envoy Gateway in a few simple steps.

## Prerequisites

A Kubernetes cluster.

__Note:__ Envoy Gateway is tested against Kubernetes v1.24.0.

## Installation

Install the Gateway API CRDs:

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0-rc2/gatewayapi-crds.yaml
```

Deploy Envoy Gateway:

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0-rc2/install.yaml
```

Deploy the example app:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.2.0-rc2/examples/kubernetes/httpbin.yaml
```

The Gateway API resources must be created in the following order:

1. First, create a GatewayClass:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.2.0-rc2/examples/kubernetes/gatewayclass.yaml
   ```

2. Create a Gateway:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.2.0-rc2/examples/kubernetes/gateway.yaml
   ```

   After applying the Gateway manifest above, a corresponding Envoy proxy Deployment, Service, and ServiceAccount will be created.

3. Create an HTTPRoute to route traffic through Envoy proxy to the example app:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.2.0-rc2/examples/kubernetes/httproute.yaml
   ```

### Testing the configuration

Port forward to the Envoy service that was created:

```shell
kubectl -n envoy-gateway-system port-forward service/envoy-default-eg 8888:8080 &
```

Curl the example app through Envoy proxy:

```shell
curl --verbose --header "Host: www.example.com" http://localhost:8888/get
```

You can replace `get` with any of the supported [httpbin methods][httpbin_methods].

### For clusters with External Loadbalancer support

You can also test the same functionality by sending traffic to the External IP. To get the external IP of the
Envoy service, run:

```shell
export GATEWAY_HOST=$(kubectl get svc/envoy-default-eg -n envoy-gateway-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

In certain environments, the load balancer may be exposed using a hostname, instead of an IP address. If so, replace
`ip` in the above command with `hostname`.

Curl the example app through Envoy proxy:

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST:8080/get
```

You can replace `get` with any of the supported [httpbin methods][httpbin_methods].

## Clean-Up

Use the steps in this section to uninstall everything from the quickstart guide.

1. Delete the HTTPRoute:

   ```shell
   kubectl delete httproute/httpbin
   ```

2. Delete the Gateway:

   ```shell
   kubectl delete gateway/eg
   ```

3. Delete the GatewayClass:

   ```shell
   kubectl delete gc/eg
   ```

4. Uninstall the example app:

   ```shell
   kubectl delete -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.2.0-rc2/examples/kubernetes/httpbin.yaml
   ```

5. Uninstall Envoy Gateway:

   ```shell
   kubectl delete -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0-rc2/install.yaml
   ```

6. Uninstall Gateway API CRDs:

   ```shell
   kubectl delete -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0-rc2/gatewayapi-crds.yaml
   ```

## Next Steps

Checkout the [Developer Guide](../dev/README.md) to get involved in the project.

[httpbin_methods]: https://httpbin.org/#/HTTP_Methods
