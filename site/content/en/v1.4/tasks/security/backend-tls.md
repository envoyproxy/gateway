---
title: "Backend TLS: Gateway to Backend"
---

This task demonstrates how TLS can be achieved between the Gateway and a backend.
This task uses a self-signed CA, so it should be used for testing and demonstration purposes only.

Envoy Gateway supports the Gateway-API defined [BackendTLSPolicy][].

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

{{< boilerplate prerequisites >}}

## TLS Certificates

Generate the certificates and keys used by the backend to terminate TLS connections from the Gateways.

Create a root certificate and private key to sign certificates:

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout ca.key -out ca.crt
```

Create a certificate and a private key for `www.example.com`.

First, create an openssl configuration file:

```shell
cat > openssl.conf  <<EOF
[req]
req_extensions = v3_req
prompt = no

[v3_req]
keyUsage = keyEncipherment, digitalSignature
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = www.example.com
EOF
```

Then create a certificate using this openssl configuration file:

```shell
openssl req -out www.example.com.csr -newkey rsa:2048 -nodes -keyout www.example.com.key -subj "/CN=www.example.com/O=example organization"
openssl x509 -req -days 365 -CA ca.crt -CAkey ca.key -set_serial 0 -in www.example.com.csr -out www.example.com.crt -extfile openssl.conf -extensions v3_req
```

Note that the certificate must contain a DNS SAN for the relevant domain.

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
kubectl patch deployment backend --type=json --patch '
  - op: add
    path: /spec/template/spec/containers/0/volumeMounts
    value:
    - name: secret-volume
      mountPath: /etc/secret-volume
  - op: add
    path: /spec/template/spec/volumes
    value:
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
      name: TLS_SERVER_CERT
      value: /etc/secret-volume/crt
  - op: add
    path: /spec/template/spec/containers/0/env/-
    value:
      name: TLS_SERVER_PRIVKEY
      value: /etc/secret-volume/key
  '
```

Create a service that exposes port 443 on the backend service.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

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

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
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
```

{{% /tab %}}
{{< /tabpane >}}

Create a [BackendTLSPolicy][] instructing Envoy Gateway to establish a TLS connection with the backend and validate the backend certificate is issued by a trusted CA and contains an appropriate DNS SAN.

Note: SectionName is an optional field that specifies the name of the port in the target backend. This example uses a Kubernetes Service as the backend target, so the sectionName is set to `https` to match the port name in the Service.
If the target is a [Backend] resource, the `sectionName` field should be set to the port number of the backend.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: BackendTLSPolicy
metadata:
  name: enable-backend-tls
  namespace: default
spec:
  targetRefs:
  - group: ''
    kind: Service
    name: tls-backend
    sectionName: https
  validation:
    caCertificateRefs:
    - name: example-ca
      group: ''
      kind: ConfigMap
    hostname: www.example.com
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: BackendTLSPolicy
metadata:
  name: enable-backend-tls
  namespace: default
spec:
  targetRefs:
  - group: ''
    kind: Service
    name: tls-backend
    sectionName: https
  validation:
    caCertificateRefs:
    - name: example-ca
      group: ''
      kind: ConfigMap
    hostname: www.example.com
```

{{% /tab %}}
{{< /tabpane >}}

Patch the HTTPRoute's backend reference, so that it refers to the new TLS-enabled service:

```shell
kubectl patch HTTPRoute backend --type=json --patch '
  - op: replace
    path: /spec/rules/0/backendRefs/0/port
    value: 443
  - op: replace
    path: /spec/rules/0/backendRefs/0/name
    value: tls-backend
  '
```

Verify the HTTPRoute status:

```shell
kubectl get HTTPRoute backend -o yaml
```

## Testing backend TLS

{{< tabpane text=true >}}
{{% tab header="With External LoadBalancer Support" %}}

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

{{% /tab %}}
{{% tab header="Without LoadBalancer Support" %}}

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

{{% /tab %}}
{{< /tabpane >}}

## Customize backend TLS Parameters

In addition to enablement of backend TLS with the Gateway-API BackendTLSPolicy, Envoy Gateway supports customizing  TLS parameters.
To achieve this, the [EnvoyProxy][] resource can be used to specify TLS parameters. We will customize the TLS version in this example.

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

You can customize the EnvoyProxy Backend TLS Parameters via EnvoyProxy Config like:

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
    minVersion: "1.3"
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
    minVersion: "1.3"
```

{{% /tab %}}
{{< /tabpane >}}

## Testing TLS Parameters

Query the TLS-enabled backend through Envoy proxy:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:80:127.0.0.1" \
http://www.example.com:80/get
```

Inspect the output and see that the response contains the details of the TLS handshake between Envoy and the backend.
The TLS version is now TLS1.3, as configured in the EnvoyProxy resource. The TLS cipher is also changed, since TLS1.3 supports different ciphers from TLS1.2.

```shell
< HTTP/1.1 200 OK
[...]
 "tls": {
  "version": "TLSv1.3",
  "serverName": "www.example.com",
  "negotiatedProtocol": "http/1.1",
  "cipherSuite": "TLS_AES_128_GCM_SHA256"
 }
```

[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[EnvoyProxy]: ../../api/extension_types#envoyproxy
[Backend]: ../../api/extension_types#backend
