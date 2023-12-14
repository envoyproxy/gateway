---
title: "TLS Passthrough"
---

This guide will walk through the steps required to configure TLS Passthrough via Envoy Gateway. Unlike configuring
Secure Gateways, where the Gateway terminates the client TLS connection, TLS Passthrough allows the application itself
to terminate the TLS connection, while the Gateway routes the requests to the application based on SNI headers.

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

Follow the steps from the [Quickstart Guide](../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## TLS Certificates

Generate the certificates and keys used by the Service to terminate client TLS connections.
For the application, we'll deploy a sample echoserver app, with the certificates loaded in the application Pod.

__Note:__ These certificates will not be used by the Gateway, but will remain in the application scope.

Create a root certificate and private key to sign certificates:

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout example.com.key -out example.com.crt
```

Create a certificate and a private key for `passthrough.example.com`:

```shell
openssl req -out passthrough.example.com.csr -newkey rsa:2048 -nodes -keyout passthrough.example.com.key -subj "/CN=passthrough.example.com/O=some organization"
openssl x509 -req -sha256 -days 365 -CA example.com.crt -CAkey example.com.key -set_serial 0 -in passthrough.example.com.csr -out passthrough.example.com.crt
```

Store the cert/keys in A Secret:

```shell
kubectl create secret tls server-certs --key=passthrough.example.com.key --cert=passthrough.example.com.crt
```

## Deployment

Deploy TLS Passthrough application Deployment, Service and TLSRoute:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.4.0/examples/kubernetes/tls-passthrough.yaml
```

Patch the Gateway from the Quickstart guide to include a TLS listener that listens on port `6443` and is configured for
TLS mode Passthrough:

```shell
kubectl patch gateway eg --type=json --patch '[{
   "op": "add",
   "path": "/spec/listeners/-",
   "value": {
      "name": "tls",
      "protocol": "TLS",
      "hostname": "passthrough.example.com",
      "tls": {"mode": "Passthrough"}, 
      "port": 6443,
    },
}]'
```

## Testing

### Clusters without External LoadBalancer Support

Get the name of the Envoy service created the by the example Gateway:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

Port forward to the Envoy service:

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 6043:6443 &
```

Curl the example app through Envoy proxy:

```shell
curl -v --resolve "passthrough.example.com:6043:127.0.0.1" https://passthrough.example.com:6043 \
--cacert passthrough.example.com.crt
```

### Clusters with External LoadBalancer Support

You can also test the same functionality by sending traffic to the External IP of the Gateway:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Curl the example app through the Gateway, e.g. Envoy proxy:

```shell
curl -v -HHost:passthrough.example.com --resolve "passthrough.example.com:6443:${GATEWAY_HOST}" \
--cacert example.com.crt https://passthrough.example.com:6443/get
```

## Clean-Up

Follow the steps from the [Quickstart Guide](../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the Secret:

```shell
kubectl delete secret/server-certs
```

## Next Steps

Checkout the [Developer Guide](../../contributions/develop/) to get involved in the project.
