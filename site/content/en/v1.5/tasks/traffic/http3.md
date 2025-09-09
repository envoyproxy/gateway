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

[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[ClientTrafficPolicy]: ../../../api/extension_types#clienttrafficpolicy
[Secret]: https://kubernetes.io/docs/concepts/configuration/secret/
