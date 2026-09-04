---
title: "TLS Passthrough"
---

This task will walk through the steps required to configure TLS Passthrough via Envoy Gateway. Unlike configuring
Secure Gateways, where the Gateway terminates the client TLS connection, TLS Passthrough allows the application itself
to terminate the TLS connection, while the Gateway routes the requests to the application based on SNI headers.

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

{{< boilerplate prerequisites >}}

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
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/tls-passthrough.yaml
```

Patch the Gateway from the Quickstart to include a TLS listener that listens on port `6443` and is configured for
TLS mode Passthrough:

```shell
kubectl patch gateway eg --type=json --patch '
  - op: add
    path: /spec/listeners/-
    value:
      name: tls
      protocol: TLS
      hostname: passthrough.example.com
      port: 6443
      tls:
        mode: Passthrough
   '
```

## Testing

{{< tabpane text=true >}}
{{% tab header="With External LoadBalancer Support" %}}

You can also test the same functionality by sending traffic to the External IP of the Gateway:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Curl the example app through the Gateway, e.g. Envoy proxy.

Use the **CA** certificate (`example.com.crt`) with `--cacert`. That is the
issuer of `passthrough.example.com.crt`. Prefer the CA over the leaf server
certificate: some TLS stacks accept an explicitly trusted leaf as a trust
anchor, but others report `SSL certificate problem: unable to get local issuer
certificate`. The CA path is the portable choice.

```shell
curl -v -HHost:passthrough.example.com --resolve "passthrough.example.com:6443:${GATEWAY_HOST}" \
--cacert example.com.crt https://passthrough.example.com:6443/get
```

{{% /tab %}}
{{% tab header="Without LoadBalancer Support" %}}

Get the name of the Envoy service created the by the example Gateway:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

Port forward to the Envoy service:

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 6043:6443 &
```

Curl the example app through Envoy proxy. As above, `--cacert` must be the
**CA** (`example.com.crt`), not the leaf `passthrough.example.com.crt`:

```shell
curl -v --resolve "passthrough.example.com:6043:127.0.0.1" \
  -HHost:passthrough.example.com \
  --cacert example.com.crt \
  https://passthrough.example.com:6043/get
```

{{% /tab %}}
{{< /tabpane >}}

### Troubleshooting certificate verification

If curl reports `unable to get local issuer certificate` or `unknown CA`:

1. **Verify the client trust anchor**:
   - For the sample certificates generated in this guide, confirm `--cacert` points at `example.com.crt` (the CA that signed the application certificate), not at `passthrough.example.com.crt`.
   - If you replaced the sample certificates with a publicly trusted certificate (e.g., Let's Encrypt), omit the `--cacert` flag so curl uses the system trust store.
   - If using a private or internal CA, point `--cacert` at that CA's root certificate bundle.
2. Confirm the app Secret still holds the keypair created earlier:

   ```shell
   kubectl get secret server-certs -o yaml
   ```

3. If you replaced the sample certs with an intermediate-signed certificate, the TLS Secret presented by the **application** must include the full chain the client needs (leaf plus intermediates). Kubernetes `tls.crt` may contain multiple PEM blocks concatenated leaf-first:

   ```shell
   cat passthrough.example.com.crt intermediate.crt > fullchain.crt
   kubectl create secret tls server-certs \
     --key=passthrough.example.com.key \
     --cert=fullchain.crt \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

   The sample Deployment mounts the Secret and loads the keypair at process
   start. After replacing `server-certs`, restart the app so it presents the
   new certificate:

   ```shell
   kubectl rollout restart deployment/passthrough-echoserver
   kubectl rollout status deployment/passthrough-echoserver
   ```

4. For a quick connectivity check only, `curl -k` skips verification; do not use
   that as a substitute for fixing the CA/chain in real deployments.

Because this task uses **TLS Passthrough**, Envoy does not terminate TLS and
does not use a Gateway TLS Secret for this listener. Certificate problems in
the curl client almost always come from the **backend** certificate and the
CA file passed to curl, not from Gateway TLS settings.

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the Secret:

```shell
kubectl delete secret/server-certs
```

## Next Steps

Checkout the [Developer Guide](/community/develop) to get involved in the project.
