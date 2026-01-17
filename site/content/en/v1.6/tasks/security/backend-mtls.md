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

Generate the client certificate and key used by the Gateway for authentication against the backend.

Create a root certificate and private key to sign the client certificate:

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout clientca.key -out clientca.crt
```

Create the client certificate and a private key for `www.example.com`:

```shell
openssl req -new -newkey rsa:2048 -nodes -keyout client.key -out client.csr -subj "/CN=example-client/O=example organization"
openssl x509 -req -days 365 -CA clientca.crt -CAkey clientca.key -set_serial 0 -in client.csr -out client.crt
```

Store the the client cert/key in a Secret:

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

Envoy Gateway supports two ways to configure client certificates for backend mTLS.
* Configure the [EnvoyProxy][] resource to specify a client certificate globally.
* Configure a [Backend][] resource to specify a client certificate per backend.

### Configure EnvoyProxy

The [EnvoyProxy][] resource can be used to specify a client certificate globally for a GatewayClass or Gateway.

The following example shows how to configure a GatewayClass to reference an EnvoyProxy resource with a client certificate.

First, add a parametersRef field in the GatewayClass to reference the EnvoyProxy configuration:

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

### Configure a Backend Resource

When multiple backends require distinct client certificates, configure each one using a dedicated [Backend][] resource that includes its own client certificate reference.
TLS settings defined in the Backend resource take precedence over the defaults set in EnvoyProxy.backendTLS.

Before creating Backend resources, make sure the Backend API is enabled in the Envoy Gateway configuration (see [Backend Routing](../traffic/backend#enable-backend) for full details). If it is not already enabled, update the `envoy-gateway-config` ConfigMap as shown below:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-gateway-config
  namespace: envoy-gateway-system
data:
  envoy-gateway.yaml: |
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyGateway
    provider:
      type: Kubernetes
    gateway:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
    extensionApis:
      enableBackend: true
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-gateway-config
  namespace: envoy-gateway-system
data:
  envoy-gateway.yaml: |
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyGateway
    provider:
      type: Kubernetes
    gateway:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
    extensionApis:
      enableBackend: true
```

{{% /tab %}}
{{< /tabpane >}}

{{< boilerplate rollout-envoy-gateway >}}

Roll out the updated Envoy Gateway deployment:

```shell
kubectl -n envoy-gateway-system rollout restart deployment envoy-gateway
```

Create the client certificate secret in the backend namespace (reusing the same certificate and key generated earlier):

```shell
kubectl -n default create secret tls example-client-cert --key=client.key --cert=client.crt
```

Define the backend and include both the trust anchor and the client credentials:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: tls-backend-client-cert
  namespace: default
spec:
  endpoints:
  - fqdn:
      hostname: tls-backend.default.svc.cluster.local
      port: 443
  tls:
    clientCertificateRef:
      kind: Secret
      name: example-client-cert
    caCertificateRefs:
    - group: ""
      kind: ConfigMap
      name: example-ca
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: tls-backend-client-cert
  namespace: default
spec:
  endpoints:
  - fqdn:
      hostname: tls-backend.default.svc.cluster.local
      port: 443
  tls:
    clientCertificateRef:
      kind: Secret
      name: example-client-cert
    caCertificateRefs:
    - group: ""
      kind: ConfigMap
      name: example-ca
```

{{% /tab %}}
{{< /tabpane >}}

Update the HTTPRoute to reference the backend instead of the Kubernetes Service:

```shell
kubectl patch HTTPRoute backend --type=json --patch '
  - op: replace
    path: /spec/rules/0/backendRefs/0/group
    value: gateway.envoyproxy.io
  - op: replace
    path: /spec/rules/0/backendRefs/0/kind
    value: Backend
  - op: replace
    path: /spec/rules/0/backendRefs/0/name
    value: tls-backend-client-cert
  - op: remove
    path: /spec/rules/0/backendRefs/0/port
  '
```

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
[Backend]: ../../api/extension_types#backend
