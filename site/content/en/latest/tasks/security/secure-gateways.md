---
title: "Secure Gateways"
---

This task will help you get started using secure Gateways.
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

Update the Gateway from the Quickstart to include an HTTPS listener that listens on port `443` and references the
`example-cert` Secret:

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

Verify the Gateway status:

```shell
kubectl get gateway/eg -o yaml
```

## OCSP Stapling

Online Certificate Status Protocol (OCSP) stapling allows the gateway to proactively attach proof that a certificate is still valid, eliminating the need for each client to query the certificate authority (CA) during the TLS handshake. This reduces latency and prevents client browsing information from being exposed to the CA.

Envoy Gateway supports OCSP stapling by attaching an OCSP response during the TLS handshake whenever the Gateway listener’s certificate Secret includes one. Specifically, Envoy Gateway looks for the OCSP response in the Secret’s `tls.ocsp-staple` data field. If present, Envoy Gateway staples the response to the handshake automatically.

## Testing

{{< tabpane text=true >}}
{{% tab header="With External LoadBalancer Support" %}}

Get the External IP of the Gateway:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Query the example app through the Gateway:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com/get
```

{{% /tab %}}
{{% tab header="Without LoadBalancer Support" %}}

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
--cacert example.com.crt https://www.example.com:8443/get
```

{{% /tab %}}
{{< /tabpane >}}


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
kubectl patch gateway eg --type=json --patch '
  - op: add
    path: /spec/listeners/-
    value:
      name: https-foo
      protocol: HTTPS
      port: 443
      hostname: foo.example.com
      tls:
        mode: Terminate
        certificateRefs:
        - kind: Secret
          group: ""
          name: foo-cert
  '
```

Update the HTTPRoute to route traffic for hostname `foo.example.com` to the example backend service:

```shell
kubectl patch httproute backend --type=json --patch '
  - op: add
    path: /spec/hostnames/-
    value: foo.example.com
  '
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

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1beta1
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

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1beta1
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
```

{{% /tab %}}
{{< /tabpane >}}

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

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  listeners:
    - name: http
      protocol: HTTP
      port: 80
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - kind: Secret
            group: ""
            name: example-cert
            namespace: envoy-gateway-system
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  listeners:
    - name: http
      protocol: HTTP
      port: 80
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - kind: Secret
            group: ""
            name: example-cert
            namespace: envoy-gateway-system
```

{{% /tab %}}
{{< /tabpane >}}

The Gateway HTTPS listener status should now surface the `Ready: True` condition and you should once again be able to
query the HTTPS backend through the Gateway.

Lastly, test connectivity using the above [Testing section](#testing).

## Clean-Up

Follow the steps from the [Quickstart](../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the Secrets:

```shell
kubectl delete secret/example-cert
kubectl delete secret/foo-cert
```

# RSA + ECDSA Dual stack certificates

This section gives a walkthrough to generate RSA and ECDSA derived certificates and keys for the Server, which can then be configured in the Gateway listener, to terminate TLS traffic.

## Prerequisites

{{< boilerplate prerequisites >}}

Follow the steps in the [TLS Certificates](#tls-certificates) section to generate self-signed RSA derived Server certificate and private key, and configure those in the Gateway listener configuration to terminate HTTPS traffic.

## Pre-checks

{{< tabpane text=true >}}
{{% tab header="With External LoadBalancer Support" %}}

Get the External IP of the Gateway:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Query the example app through Envoy Proxy while enforcing an RSA cipher:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com/get  -Isv --ciphers ECDHE-RSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2
```

The call should succeed with the following message:
```
...
* TLSv1.2 (IN), TLS change cipher, Change cipher spec (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-CHACHA20-POLY1305
...
...
```

Since the Secret configured at this point is an RSA based Secret, if we enforce the usage of an ECDSA cipher, the call should fail:

```shell
 curl -v -HHost:www.example.com --resolve "www.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com/get  -Isv --ciphers ECDHE-ECDSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2
```

The call above fails with the following message:
```
* Added www.example.com:443:127.0.0.1 to DNS cache
* Hostname www.example.com was found in DNS cache
*   Trying 127.0.0.1:443...
* Connected to www.example.com (127.0.0.1) port 443
* ALPN: curl offers h2,http/1.1
* Cipher selection: ECDHE-ECDSA-CHACHA20-POLY1305
* TLSv1.2 (OUT), TLS handshake, Client hello (1):
*  CAfile: example.com.crt
*  CApath: none
* TLSv1.2 (IN), TLS alert, handshake failure (552):
* OpenSSL/3.0.14: error:0A000410:SSL routines::sslv3 alert handshake failure
* Closing connection
```

{{% /tab %}}
{{% tab header="Without LoadBalancer Support" %}}

Query the example app through Envoy proxy while enforcing an RSA cipher:
```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8443/get  -Isv --ciphers ECDHE-RSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2
```

The command should succeed with the following message:
```
...
* TLSv1.2 (IN), TLS change cipher, Change cipher spec (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-CHACHA20-POLY1305
...
```

Since the Secret configured at this point is an RSA based Secret, if we enforce the usage of an ECDSA cipher, the call should fail:

```shell
 curl -v -HHost:www.example.com --resolve "www.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com/get  -Isv --ciphers ECDHE-ECDSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2
```

The command above fails with the following message:
```
* Added www.example.com:8443:127.0.0.1 to DNS cache
* Hostname www.example.com was found in DNS cache
*   Trying 127.0.0.1:8443...
* Connected to www.example.com (127.0.0.1) port 8443 (#0)
* ALPN: offers h2
* ALPN: offers http/1.1
* Cipher selection: ECDHE-ECDSA-CHACHA20-POLY1305
*  CAfile: example.com.crt
*  CApath: none
* (304) (OUT), TLS handshake, Client hello (1):
* error:1404B410:SSL routines:ST_CONNECT:sslv3 alert handshake failure
* Closing connection 0
```

{{% /tab %}}
{{< /tabpane >}}

Moving forward in the doc, we will be configuring the existing Gateway listener to accept both kinds of ciphers.

## TLS Certificates

Reuse the CA certificate and key pair generated in the [Secure Gateways](#tls-certificates) task and use this CA to sign both RSA and ECDSA Server certificates.
Note the CA certificate and key names are `example.com.crt` and `example.com.key` respectively.


Create an ECDSA certificate and a private key for `www.example.com`:

```shell
openssl ecparam -noout -genkey -name prime256v1 -out www.example.com.ecdsa.key
openssl req -new -SHA384 -key www.example.com.ecdsa.key -nodes -out www.example.com.ecdsa.csr -subj "/CN=www.example.com/O=example organization"
openssl x509 -req -SHA384  -days 365 -in www.example.com.ecdsa.csr -CA example.com.crt -CAkey example.com.key -CAcreateserial -out www.example.com.ecdsa.crt
```

Store the cert/key in a Secret:

```shell
kubectl create secret tls example-cert-ecdsa --key=www.example.com.ecdsa.key --cert=www.example.com.ecdsa.crt
```

Patch the Gateway with this additional ECDSA Secret:

```shell
kubectl patch gateway eg --type=json --patch '
  - op: add
    path: /spec/listeners/1/tls/certificateRefs/-
    value:
      name: example-cert-ecdsa
  '
```

Verify the Gateway status:

```shell
kubectl get gateway/eg -o yaml
```

## Testing

{{< tabpane text=true >}}
{{% tab header="With External LoadBalancer Support" %}}

Get the External IP of the Gateway:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Query the example app through Envoy Proxy while enforcing an RSA cipher:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com/get  -Isv --ciphers ECDHE-RSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2
```

The call should succeed with the following message:
```
...
* TLSv1.2 (IN), TLS change cipher, Change cipher spec (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-CHACHA20-POLY1305
...
...
```

Additionally, querying the example app while enforcing an ECDSA cipher should also work now:

```shell
 curl -v -HHost:www.example.com --resolve "www.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com/get  -Isv --ciphers ECDHE-ECDSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2
```

The call above succeeds with the following message:
```
...
* TLSv1.2 (IN), TLS change cipher, Change cipher spec (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-ECDSA-CHACHA20-POLY1305
...
```

{{% /tab %}}
{{% tab header="Without LoadBalancer Support" %}}

Query the example app through Envoy proxy while enforcing an RSA cipher:
```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8443/get  -Isv --ciphers ECDHE-RSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2
```

The command should succeed with the following message:
```
...
* TLSv1.2 (IN), TLS change cipher, Change cipher spec (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-CHACHA20-POLY1305
...
```

Additionally, querying the example app while enforcing an ECDSA cipher should also work now:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8443/get  -Isv --ciphers ECDHE-ECDSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2
```

The command above succeeds with the following message:
```
...
* TLSv1.2 (IN), TLS change cipher, Change cipher spec (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-ECDSA-CHACHA20-POLY1305
...
```

{{% /tab %}}
{{< /tabpane >}}

# SNI based Certificate selection

This sections gives a walkthrough to generate multiple certificates corresponding to different FQDNs. The same Gateway listener can then be configured to terminate TLS traffic for multiple FQDNs based on the SNI matching.

## Prerequisites

{{< boilerplate prerequisites >}}

Follow the steps in the [TLS Certificates](#tls-certificates) section to generate self-signed RSA derived Server certificate and private key, and configure those in the Gateway listener configuration to terminate HTTPS traffic.

## Additional Configurations

Using the [TLS Certificates](#tls-certificates) section, we first generate additional Secret for another Host `www.sample.com`.

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=sample Inc./CN=sample.com' -keyout sample.com.key -out sample.com.crt

openssl req -out www.sample.com.csr -newkey rsa:2048 -nodes -keyout www.sample.com.key -subj "/CN=www.sample.com/O=sample organization"
openssl x509 -req -days 365 -CA sample.com.crt -CAkey sample.com.key -set_serial 0 -in www.sample.com.csr -out www.sample.com.crt

kubectl create secret tls sample-cert --key=www.sample.com.key --cert=www.sample.com.crt
```

Note that all occurrences of `example.com` were just replaced with `sample.com`


Next we update the `Gateway` configuration to accommodate the new Certificate which will be used to Terminate TLS traffic:

```shell
kubectl patch gateway eg --type=json --patch '
  - op: add
    path: /spec/listeners/1/tls/certificateRefs/-
    value:
      name: sample-cert
  '
```

Finally, we update the HTTPRoute to route traffic for hostname `www.sample.com` to the example backend service:

```shell
kubectl patch httproute backend --type=json --patch '
  - op: add
    path: /spec/hostnames/-
    value: www.sample.com
  '
```

## Testing

{{< tabpane text=true >}}
{{% tab header="With External LoadBalancer Support" %}}

Get the External IP of the Gateway:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Query the example app through Envoy Proxy:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com/get -I
```

Similarly, query the sample app through the same Envoy proxy:

```shell
curl -v -HHost:www.sample.com --resolve "www.sample.com:443:${GATEWAY_HOST}" \
--cacert sample.com.crt https://www.sample.com/get -I
```

{{% /tab %}}
{{% tab header="Without LoadBalancer Support" %}}

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
--cacert example.com.crt https://www.example.com:8443/get -I
```

Similarly, query the sample app through the same Envoy proxy:

```shell
curl -v -HHost:www.sample.com --resolve "www.sample.com:8443:127.0.0.1" \
--cacert sample.com.crt https://www.sample.com:8443/get -I
```

{{% /tab %}}
{{< /tabpane >}}

Since the multiple certificates are configured on the same Gateway listener, Envoy was able to provide the client with appropriate certificate based on the SNI in the client request.

## Customize Gateway TLS Parameters

In addition to enablement of TLS with Gateway-API, Envoy Gateway supports customizing TLS parameters.
To achieve this, the [ClientTrafficPolicy][] resource can be used to specify TLS parameters.
We will customize the minimum supported TLS version in this example to TLSv1.3.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enforce-tls-13
  namespace: default
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
  tls:
    minVersion: "1.3"
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
  name: enforce-tls-13
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  tls:
    minVersion: "1.3"
```

{{% /tab %}}
{{< /tabpane >}}


## Testing TLS Parameters

{{< tabpane text=true >}}
{{% tab header="With External LoadBalancer Support" %}}

Attempt to connect using an unsupported TLS version:

```shell
curl -v -HHost:www.sample.com --resolve "www.sample.com:443:${GATEWAY_HOST}" \
--cacert sample.com.crt --tlsv1.2 --tls-max 1.2 https://www.sample.com/get -I
```

You should receive the following error:
```
[...]

* ALPN: curl offers h2,http/1.1
* (304) (OUT), TLS handshake, Client hello (1):
* LibreSSL/3.3.6: error:1404B42E:SSL routines:ST_CONNECT:tlsv1 alert protocol version
* Closing connection
curl: (35) LibreSSL/3.3.6: error:1404B42E:SSL routines:ST_CONNECT:tlsv1 alert protocol version
```

The output shows that the connection fails due to an unsupported TLS protocol version used by the client. Now, connect to the Gateway without specifying a client version, and note that the connection is established with TLSv1.3.

```shell
curl -v -HHost:www.sample.com --resolve "www.sample.com:443:${GATEWAY_HOST}" \
--cacert sample.com.crt https://www.sample.com/get -I
```

The command above should succeed and output the following:
```
[...]
* SSL connection using TLSv1.3 / AEAD-CHACHA20-POLY1305-SHA256 / [blank] / UNDEF
```


{{% /tab %}}
{{% tab header="Without LoadBalancer Support" %}}

Attempt to connecting using an unsupported TLS version:

```shell
curl -v -HHost:www.sample.com --resolve "www.sample.com:8443:127.0.0.1" \
--cacert sample.com.crt --tlsv1.2 --tls-max 1.2 https://www.sample.com:8443/get -I
```

You should receive the following error:
```
[...]

* ALPN: curl offers h2,http/1.1
* (304) (OUT), TLS handshake, Client hello (1):
* LibreSSL/3.3.6: error:1404B42E:SSL routines:ST_CONNECT:tlsv1 alert protocol version
* Closing connection
curl: (35) LibreSSL/3.3.6: error:1404B42E:SSL routines:ST_CONNECT:tlsv1 alert protocol version
```

The output shows that the connection fails due to an unsupported TLS protocol version used by the client. Now, connect
to the Gateway without specifying a client version, and note that the connection is established with TLSv1.3.

```shell
curl -v -HHost:www.sample.com --resolve "www.sample.com:8443:127.0.0.1" \
--cacert sample.com.crt https://www.sample.com:8443/get -I
```

The command above should succeed and output the following:
```
[...]
* SSL connection using TLSv1.3 / AEAD-CHACHA20-POLY1305-SHA256 / [blank] / UNDEF
```

{{% /tab %}}
{{< /tabpane >}}

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[ReferenceGrant]: https://gateway-api.sigs.k8s.io/api-types/referencegrant/
[ClientTrafficPolicy]: ../../api/extension_types#clienttrafficpolicy
