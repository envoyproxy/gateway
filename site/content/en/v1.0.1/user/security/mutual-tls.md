---
title: "Mutual TLS: External Clients to the Gateway"
---

This guide demonstrates how mutual TLS can be achieved between external clients and the Gateway. The guide uses a self-signed CA, so it should be used for
testing and demonstration purposes only.

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

Follow the steps from the [Quickstart Guide](../../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

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
kubectl create secret tls example-cert --key=www.example.com.key --cert=www.example.com.crt --certificate-authority=example.com.crt
```

Store the CA Cert in another Secret:

```shell
kubectl create secret generic example-ca-cert --from-file=ca.crt=example.com.crt
```

Create a certificate and a private key for the client `client.example.com`:

```shell
openssl req -out client.example.com.csr -newkey rsa:2048 -nodes -keyout client.example.com.key -subj "/CN=client.example.com/O=example organization"
openssl x509 -req -days 365 -CA example.com.crt -CAkey example.com.key -set_serial 0 -in client.example.com.csr -out client.example.com.crt
```

Update the Gateway from the Quickstart guide to include an HTTPS listener that listens on port `443` and references the
`example-cert` Secret:

```shell
kubectl patch gateway eg --type=json --patch '[{
   "op": "add",
   "path": "/spec/listeners/-",
   "value": {
      "name": "https",
      "protocol": "HTTPS",
      "port": 443,
      "tls": {
        "mode": "Terminate",
        "certificateRefs": [{
          "kind": "Secret",
          "group": "",
          "name": "example-cert",
        }],
      },
    },
}]'
```

Verify the Gateway status:

```shell
kubectl get gateway/eg -o yaml
```

Create a [ClientTrafficPolicy][] to enforce client validation using the CA Certificate as a trusted anchor.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-mtls
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
  tls:
    clientValidation:
      caCertificateRefs:
      - kind: "Secret"
        group: ""
        name: "example-ca-cert"
EOF
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
--cert client.example.com.crt --key client.example.com.key \
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
--cert client.example.com.crt --key client.example.com.key \
--cacert example.com.crt https://www.example.com/get
```

Dont specify the client key and certificate in the above command, and ensure that the connection fails

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com/get
```

[ClientTrafficPolicy]: ../../../api/extension_types#clienttrafficpolicy
