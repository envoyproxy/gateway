---
title: "Secure Gateways"
---

This guide will help you get started using secure Gateways. The guide uses a self-signed CA, so it should be used for
testing and demonstration purposes only.

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

Follow the steps from the [Quickstart Guide](../quickstart) to install Envoy Gateway and the example manifest.
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
kubectl create secret tls example-cert --key=www.example.com.key --cert=www.example.com.crt
```

Update the Gateway from the Quickstart guide to include an HTTPS listener that listens on port `8443` and references the
`example-cert` Secret:

```shell
kubectl patch gateway eg --type=json --patch '[{
   "op": "add",
   "path": "/spec/listeners/-",
   "value": {
      "name": "https",
      "protocol": "HTTPS",
      "port": 8443,
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

## Testing

### Clusters without External LoadBalancer Support

Get the name of the Envoy service created the by the example Gateway:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

Port forward to the Envoy service:

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8043:8443 &
```

Query the example app through Envoy proxy:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8043:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8043/get
```

### Clusters with External LoadBalancer Support

Get the External IP of the Gateway:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Query the example app through the Gateway:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com:8443/get
```

## Multiple HTTPS Listeners

Create a TLS cert/key for the additional HTTPS listener:

```shell
openssl req -out foo.example.com.csr -newkey rsa:2048 -nodes -keyout foo.example.com.key -subj "/CN=foo.example.com/O=example organization"
openssl x509 -req -days 365 -CA example.com.crt -CAkey example.com.key -set_serial 0 -in foo.example.com.csr -out foo.example.com.crt
```

Store the cert/key in a Secret:

```shell
kubectl create secret tls foo-cert --key=foo.example.com.key --cert=foo.example.com.crt
```

Create another HTTPS listener on the example Gateway:

```shell
kubectl patch gateway eg --type=json --patch '[{
   "op": "add",
   "path": "/spec/listeners/-",
   "value": {
      "name": "https-foo",
      "protocol": "HTTPS",
      "port": 8443,
      "hostname": "foo.example.com",
      "tls": {
        "mode": "Terminate",
        "certificateRefs": [{
          "kind": "Secret",
          "group": "",
          "name": "foo-cert",
        }],
      },
    },
}]'
```

Update the HTTPRoute to route traffic for hostname `foo.example.com` to the example backend service:

```shell
kubectl patch httproute backend --type=json --patch '[{
   "op": "add",
   "path": "/spec/hostnames/-",
   "value": "foo.example.com",
}]'
```

Verify the Gateway status:

```shell
kubectl get gateway/eg -o yaml
```

Follow the steps in the [Testing section](#testing) to test connectivity to the backend app through both Gateway
listeners. Replace `www.example.com` with `foo.example.com` to test the new HTTPS listener.

## Cross Namespace Certificate References

A Gateway can be configured to reference a certificate in a different namespace. This is allowed by a [ReferenceGrant][]
created in the target namespace. Without the ReferenceGrant, a cross-namespace reference is invalid.

Before proceeding, ensure you can query the HTTPS backend service from the [Testing section](#testing).

To demonstrate cross namespace certificate references, create a ReferenceGrant that allows Gateways from the "default"
namespace to reference Secrets in the "envoy-gateway-system" namespace:

```console
$ cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: ReferenceGrant
metadata:
  name: example
  namespace: envoy-gateway-system
spec:
  from:
  - group: gateway.networking.k8s.io
    kind: Gateway
    namespace: default
  to:
  - group: ""
    kind: Secret
EOF
```

Delete the previously created Secret:

```shell
kubectl delete secret/example-cert
```

The Gateway HTTPS listener should now surface the `Ready: False` status condition and the example HTTPS backend should
no longer be reachable through the Gateway.

```shell
kubectl get gateway/eg -o yaml
```

Recreate the example Secret in the `envoy-gateway-system` namespace:

```shell
kubectl create secret tls example-cert -n envoy-gateway-system --key=www.example.com.key --cert=www.example.com.crt
```

Update the Gateway HTTPS listener with `namespace: envoy-gateway-system`, for example:

```console
$ cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  listeners:
    - name: http
      protocol: HTTP
      port: 8080
    - name: https
      protocol: HTTPS
      port: 8443
      tls:
        mode: Terminate
        certificateRefs:
          - kind: Secret
            group: ""
            name: example-cert
            namespace: envoy-gateway-system
EOF
```

The Gateway HTTPS listener status should now surface the `Ready: True` condition and you should once again be able to
query the HTTPS backend through the Gateway.

Lastly, test connectivity using the above [Testing section](#testing).

## Clean-Up

Follow the steps from the [Quickstart Guide](../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the Secrets:

```shell
kubectl delete secret/example-cert
kubectl delete secret/foo-cert
```

## Next Steps

Checkout the [Developer Guide](../../contributions/develop/) to get involved in the project.

[ReferenceGrant]: https://gateway-api.sigs.k8s.io/api-types/referencegrant/
