---
title: "HTTP3"
---

This task will help you get started using HTTP3 using EG.
This task uses a self-signed CA, so it should be used for testing and demonstration purposes only.

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

{{< boilerplate prerequisites >}}

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

Update the [Gateway][] from the Quickstart to include an HTTPS listener that listens on port `443` and references the
`example-cert` [Secret][]:

```shell
kubectl patch gateway eg --type=json --patch '
  - op: add
    path: /spec/listeners/-
    value:
      name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
        - kind: Secret
          group: ""
          name: example-cert
  '
```

Apply the following [ClientTrafficPolicy][] to enable HTTP3

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-http3
spec:
  http3: {}
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-http3
spec:
  http3: {}
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
```

{{% /tab %}}
{{< /tabpane >}}

By default, Envoy Gateway advertises the Gateway listener port in the HTTP/3 `alt-svc` response header.
If an external load balancer exposes the Gateway on a different port, set `advertisedPort` to the
external port without changing the listener or Service ports:

```yaml
spec:
  http3:
    advertisedPort: 443
```

Verify the [Gateway][] status:

```shell
kubectl get gateway/eg -o yaml
```

## Testing

{{< tabpane text=true >}}
{{% tab header="With External LoadBalancer Support" %}}

Get the External IP of the [Gateway][]:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Query the example app through the [Gateway][]:

The below example uses a custom docker image with custom `curl` binary with built-in http3.

```shell
docker run --net=host --rm ghcr.io/macbre/curl-http3 curl -kv --http3 -HHost:www.example.com --resolve "www.example.com:443:${GATEWAY_HOST}" https://www.example.com/get
```

{{% /tab %}}
{{% tab header="Without LoadBalancer Support" %}}

It is not possible at the moment to port-forward UDP protocol in kubernetes service
check out https://github.com/kubernetes/kubernetes/issues/47862.
Hence we need external loadbalancer to test this feature out.

{{% /tab %}}
{{< /tabpane >}}

## HTTP/3 to the Backend

Everything above configures HTTP/3 between the client and the Gateway. HTTP/3 to the backend is
configured separately, with the `http3` field of a [BackendTrafficPolicy][].

QUIC always runs over TLS, so the backend must already be configured with TLS through a
[BackendTLSPolicy][] or a [Backend][] resource's `spec.tls`. A BackendTrafficPolicy that enables
`http3` for a plaintext backend is rejected with an `Accepted=False` condition; where the same
setting is reachable outside a BackendTrafficPolicy, HTTP/3 is left off instead. For a backend with
a self-signed certificate, set `insecureSkipVerify: true` on the [Backend][] resource: the QUIC
handshake still happens, but the certificate is not verified.

There are two modes:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: backend-http3
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  http3:
    mode: Auto
```

`Auto` is the default. Envoy uses HTTP/3 only for backends that advertise support for it through an
`alt-svc` response header. It caches that advertisement, then races a QUIC connection against a TCP
one and uses whichever is established first, so it falls back to HTTP/1.1 or HTTP/2 whenever QUIC is
unavailable. This is the safe choice for backends reached over a network where UDP may be blocked.

`Always` sends every request over HTTP/3 and never falls back to TCP. Use it only where the backend
is known to speak HTTP/3 and UDP is known to work, since there is no recovery if QUIC fails.

Note that `http3` cannot be combined with `useClientProtocol` or `proxyProtocol`: the first would
have Envoy pick the upstream protocol from the downstream request, and PROXY protocol is a TCP
preamble with no QUIC equivalent. It is also rejected for backends that declare an HTTP/2 or gRPC
`appProtocol`, since those ask for a protocol HTTP/3 cannot provide.

[Gateway]: https://gateway-api.sigs.k8s.io/reference/api-types/gateway/
[ClientTrafficPolicy]: ../../../api/extension_types#clienttrafficpolicy
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[Backend]: ../../../api/extension_types#backend
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[Secret]: https://kubernetes.io/docs/concepts/configuration/secret/
