---
title: "Backend Mutual TLS: Gateway to Backend"
---

This task demonstrates how mTLS can be achieved between the Gateway and a backend.
This task uses a self-signed CA, so it should be used for testing and demonstration purposes only.

Envoy Gateway supports the Gateway-API defined [BackendTLSPolicy][] to establish TLS. For mTLS, the Gateway must authenticate by presenting a client certificate to the backend.

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

Follow the steps from the [Backend TLS][] to install Envoy Gateway and configure TLS to the backend server. 

## TLS Certificates

Generate the certificates and keys used by the Gateway for authentication against the backend. 

Create a root certificate and private key to sign certificates:

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout clientca.key -out clientca.crt
```

Create a certificate and a private key for `www.example.com`:

```shell
openssl req -new -newkey rsa:2048 -nodes -keyout client.key -out client.csr -subj "/CN=example-client/O=example organization"
openssl x509 -req -days 365 -CA clientca.crt -CAkey clientca.key -set_serial 0 -in client.csr -out client.crt
```

Store the cert/key in a Secret:

```shell
kubectl -n envoy-gateway-system create secret tls example-client-cert --key=client.key --cert=client.crt
```

Store the CA Cert in another Secret:

```shell
kubectl create configmap example-client-ca --from-file=clientca.crt
```

## Enforce Client Certificate Authentication on the backend

Patch the existing quickstart backend to enforce Client Certificate Authentication. The patch will mount the server certificate and key required for TLS, and the CA certificate into the backend as volumes. 

```shell
kubectl patch deployment backend --type=json --patch '
  - op: add
    path: /spec/template/spec/containers/0/volumeMounts
    value:
    - name: client-certs-volume
      mountPath: /etc/client-certs
    - name: secret-volume
      mountPath: /etc/secret-volume      
  - op: add
    path: /spec/template/spec/volumes
    value:
    - name: client-certs-volume
      configMap:
        name: example-client-ca
        items:
        - key: clientca.crt
          path: crt
    - name: secret-volume
      secret:
        secretName: example-cert
        items:
        - key: tls.crt
          path: crt
        - key: tls.key
          path: key          
  - op: add
    path: /spec/template/spec/containers/0/env/-
    value:
      name: TLS_CLIENT_CACERTS
      value: /etc/client-certs/crt
  '
```

## Configure Envoy Proxy to use a client certificate 

In addition to enablement of backend TLS with the Gateway-API BackendTLSPolicy, Envoy Gateway supports customizing TLS parameters such as TLS Client Certificate.
To achieve this, the [EnvoyProxy][] resource can be used to specify a TLS Client Certificate.

First, you need to add ParametersRef in GatewayClass, and refer to EnvoyProxy Config:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: custom-proxy-config
    namespace: envoy-gateway-system
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: custom-proxy-config
    namespace: envoy-gateway-system
```

{{% /tab %}}
{{< /tabpane >}}

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  backendTLS:
    clientCertificateRef: 
      kind: Secret
      name: example-client-cert
      namespace: envoy-gateway-system
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  backendTLS:
    clientCertificateRef:
      kind: Secret
      name: example-client-cert
      namespace: envoy-gateway-system
```

{{% /tab %}}
{{< /tabpane >}}

## Testing mTLS

Query the TLS-enabled backend through Envoy proxy:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:80:127.0.0.1" \
http://www.example.com:80/get
```

Inspect the output and see that the response contains the details of the TLS handshake between Envoy and the backend. 
The response now contains a "peerCertificates" attribute that reflects the client certificate used by the Gateway to establish mTLS with the backend. 

```shell
< HTTP/1.1 200 OK
[...]
 "tls": {
  "version": "TLSv1.2",
  "serverName": "www.example.com",
  "negotiatedProtocol": "http/1.1",
  "cipherSuite": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
  "peerCertificates": ["-----BEGIN CERTIFICATE-----\n[...]-----END CERTIFICATE-----\n"]
 }
```

[Backend TLS]: ./backend-tls
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[EnvoyProxy]: ../../api/extension_types#envoyproxy