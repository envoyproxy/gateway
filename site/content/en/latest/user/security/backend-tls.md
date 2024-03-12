---
title: "Backend TLS: Gateway to Backend"
---

This guide demonstrates how TLS can be achieved between the Gateway and a backend. The guide uses a self-signed CA, so it should be used for
testing and demonstration purposes only.

Envoy Gateway supports the Gateway-API defined [BackendTLSPolicy][].

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

Follow the steps from the [Quickstart Guide](../../quickstart) to install Envoy Gateway and the example manifest.

## TLS Certificates

Generate the certificates and keys used by the backend to terminate TLS connections from the Gateways. 

Create a root certificate and private key to sign certificates:

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout ca.key -out ca.crt
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

Store the CA Cert in another Secret:

```shell
kubectl create configmap example-ca --from-file=ca.crt
```

## Setup TLS on the backend

Patch the existing quickstart backend to enable TLS. The patch will mount the TLS certificate secret into the backend as volume. 

```shell
kubectl patch deployment backend --type=json --patch '[
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/volumeMounts",
    "value": [
      {
        "name": "secret-volume",
        "mountPath": "/etc/secret-volume"
      }
    ]
  },
  {
    "op": "add",
    "path": "/spec/template/spec/volumes",
    "value": [
      {
        "name": "secret-volume",
        "secret": {
          "secretName": "example-cert",
          "items": [
            {
              "key": "tls.crt",
              "path": "crt"
            },
            {
              "key": "tls.key",
              "path": "key"
            }
          ]
        }
      }
    ]
  },
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/env/-",
    "value": {
      "name": "TLS_SERVER_CERT",
      "value": "/etc/secret-volume/crt"
    }
  },
  {
    "op": "add",
    "path": "/spec/template/spec/containers/0/env/-",
    "value": {
      "name": "TLS_SERVER_PRIVKEY",
      "value": "/etc/secret-volume/key"
    }
  }
]'
```

Create a service that exposes port 443 on the backend service. 

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  labels:
    app: backend
    service: backend
  name: tls-backend
  namespace: default
spec:
  selector:
    app: backend
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 8443
EOF
```

Create a [BackendTLSPolicy][] instructing Envoy Gateway to establish a TLS connection with the backend and validate the backend certificate is issued by a trusted CA and contains an appropriate DNS SAN.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: BackendTLSPolicy
metadata:
  name: enable-backend-tls
  namespace: default
spec:
  targetRef:
    group: ''
    kind: Service
    name: tls-backend
    sectionName: "443"
  tls:
    caCertRefs:
      - name: example-ca
        group: ''
        kind: ConfigMap
    hostname: www.example.com
EOF
```

Patch the HTTPRoute's backend reference, so that it refers to the new TLS-enabled service:

```shell
kubectl patch HTTPRoute backend --type=json --patch '[
    {
        "op": "replace",
        "path": "/spec/rules/0/backendRefs/0/port",
        "value": 443
    },
    {
        "op": "replace",
        "path": "/spec/rules/0/backendRefs/0/name",
        "value": "tls-backend"
    }
]'
```

Verify the HTTPRoute status:

```shell
kubectl get HTTPRoute backend -o yaml
```

## Testing

### Clusters without External LoadBalancer Support

Get the name of the Envoy service created the by the example Gateway:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

Port forward to the Envoy service:

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 80:80 &
```

Query the TLS-enabled backend through Envoy proxy:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:80:127.0.0.1" \
http://www.example.com:80/get
```

Inspect the output and see that the response contains the details of the TLS handshake between Envoy and the backend:

```shell
< HTTP/1.1 200 OK
[...]
 "tls": {
  "version": "TLSv1.2",
  "serverName": "www.example.com",
  "negotiatedProtocol": "http/1.1",
  "cipherSuite": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
 }
```

### Clusters with External LoadBalancer Support

Get the External IP of the Gateway:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Query the example app through the Gateway:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:80:${GATEWAY_HOST}" \
http://www.example.com:80/get
```

Inspect the output and see that the response contains the details of the TLS handshake between Envoy and the backend:

```shell
< HTTP/1.1 200 OK
[...]
 "tls": {
  "version": "TLSv1.2",
  "serverName": "www.example.com",
  "negotiatedProtocol": "http/1.1",
  "cipherSuite": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
 }
```

[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/