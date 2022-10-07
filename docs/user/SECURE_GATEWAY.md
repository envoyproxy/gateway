## Introduction
This guide will help you get started with using secure Gateways. This document uses a self-signed CA, so it should be
used for testing and demonstration purposes only.

## Prerequisites
- A Kubernetes cluster with `kubectl` context configured for the cluster.
- OpenSSL to generate TLS assets.

__Note:__ Envoy Gateway is tested against Kubernetes v1.24.0.

## Installation
Install the Gateway API CRDs:
```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0/gatewayapi-crds.yaml
```

Run Envoy Gateway:
```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0/install.yaml
```

Run the example app:
```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.2.0/examples/kubernetes/httpbin.yaml
```

### TLS Assets

Generate TLS assets used by the Gateway.

For macOS users, verify curl is compiled with the LibreSSL library:
```shell
curl --version | grep LibreSSL
curl 7.54.0 (x86_64-apple-darwin17.0) libcurl/7.54.0 LibreSSL/2.0.20 zlib/1.2.11 nghttp2/1.24.0
```

Create a root certificate and private key to sign certificates:
```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout example.com.key -out example.com.crt
```

Create a certificate and a private key for `www.example.com`:
```shell
openssl req -out www.example.com.csr -newkey rsa:2048 -nodes -keyout www.example.com.key -subj "/CN=www.example.com/O=httpbin organization"
openssl x509 -req -days 365 -CA example.com.crt -CAkey example.com.key -set_serial 0 -in www.example.com.csr -out www.example.com.crt
```

Store the cert/keys in A Secret:
```shell
kubectl create secret tls example-cert --key=www.example.com.key --cert=www.example.com.crt
```

The Gateway API resources must be created in the following order. First, create the GatewayClass:
```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.2.0/examples/kubernetes/gatewayclass.yaml
```

Create the Gateway:
```shell
kubectl apply -f https://raw.githubusercontent.com//envoyproxy/gateway/v0.2.0/examples/kubernetes/gateway.yaml
```

Create the HTTPRoute to route traffic through Envoy proxy to the example app:
```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.2.0/examples/kubernetes/httproute.yaml
```

### Testing the configuration
Port forward to the Envoy service:
```shell
kubectl -n envoy-gateway-system port-forward service/envoy-default-eg 8043:8443 &
```

Curl the example app through Envoy proxy:
```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8043:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8043/get
```
You can replace `get` with any of the supported [httpbin methods][httpbin_methods].

### For clusters with External Loadbalancer support
You can also test the same functionality by sending traffic to the External IP of the Gateway:
```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Curl the example app through the Gateway, e.g. Envoy proxy:
```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com:8443/get
```
You can replace `get` with any of the supported [httpbin methods][httpbin_methods].

## Clean-Up
Use the steps in this section to uninstall everything from the quickstart guide.

Delete the HTTPRoute:
```shell
kubectl delete httproute/httpbin
```

Delete the Gateway:
```shell
kubectl delete gateway/eg
```

Delete the Secret:
```shell
kubectl delete secret/example-cert
```

Delete the GatewayClass:
```shell
kubectl delete gc/eg
```

Uninstall the example app:
```shell
kubectl delete -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.2.0-rc2/examples/kubernetes/httpbin.yaml
```

Uninstall Envoy Gateway:
```shell
kubectl delete -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0-rc2/install.yaml
```

Uninstall Gateway API CRDs:
```shell
kubectl delete -f https://github.com/envoyproxy/gateway/releases/download/v0.2.0-rc2/gatewayapi-crds.yaml
```

## Next Steps
Checkout the [Developer Guide](../../DEVELOPER.md) to get involved in the project.

[kind]: https://kind.sigs.k8s.io/
[httpbin_methods]: https://httpbin.org/#/HTTP_Methods
