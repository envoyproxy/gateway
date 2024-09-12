---
title: "TLS Termination for TCP"
---

This guide will walk through the steps required to configure TLS Terminate mode for TCP traffic via Envoy Gateway. The guide uses a self-signed CA, so it should be used for testing and demonstration purposes only.

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

Follow the steps from the [Quickstart Guide](../quickstart) to install Envoy Gateway.

## TLS Certificates
Generate the certificates and keys used by the Gateway to terminate client TLS connections.

Create a root certificate and private key to sign certificates:

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout example.com.key -out example.com.crt
```

Create a certificate and a private key for `www.example.com`:

```shell
openssl req -out www.example.com.csr -newkey rsa:2048 -nodes -keyout www.example.com.key -subj "/CN=www.example.com/O=example organization"
openssl x509 -req -days 365 -CA example.com.crt -CAkey example.com.key -set_serial 0 -in www.example.com.csr -out www.example.com.crt
```

Store the cert/key in a Secret:

```shell
kubectl create secret tls example-cert --key=www.example.com.key --cert=www.example.com.crt
```

Install the TLS Termination for TCP example resources:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.5.0/examples/kubernetes/tls-termination.yaml
```

Verify the Gateway status:

```shell
kubectl get gateway/eg -o yaml
```

## Testing

### Clusters without External LoadBalancer Support

Get the name of the Envoy service created the by the example Gateway:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

Port forward to the Envoy service:

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8443:443 &
```

Query the example app through Envoy proxy:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8443/get
```

### Clusters with External LoadBalancer Support

Get the External IP of the Gateway:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Query the example app through the Gateway:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com/get
```
