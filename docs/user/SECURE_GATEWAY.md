# Secure Gateways
This guide will help you get started using secure Gateways. The guide uses a self-signed CA, so it should be used for
testing and demonstration purposes only.

## Prerequisites
- A Kubernetes cluster with `kubectl` context configured for the cluster.
- OpenSSL to generate TLS assets.

__Note:__ Envoy Gateway is tested against Kubernetes v1.24.

## Installation
Follow the steps from the [Quickstart Guide](QUICKSTART.md) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to curl the example backend using HTTP.

## TLS Certificates

Generate the certificates and keys used by the Gateway to terminate client TLS connections.

For macOS users, verify curl is compiled with the LibreSSL library:
```shell
curl --version | grep LibreSSL
curl 7.54.0 (x86_64-apple-darwin17.0) libcurl/7.54.0 LibreSSL/2.0.20 zlib/1.2.11 nghttp2/1.24.0
```

Create a root certificate and private key to sign certificates:
```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout example.com.key -out example.com.crt
```

Create a certificate and a private key for `www.example.com`:
```shell
openssl req -out www.example.com.csr -newkey rsa:2048 -nodes -keyout www.example.com.key -subj "/CN=www.example.com/O=httpbin organization"
openssl x509 -req -days 365 -CA example.com.crt -CAkey example.com.key -set_serial 0 -in www.example.com.csr -out www.example.com.crt
```

Store the cert/key in a Secret:
```shell
kubectl create secret tls example-cert --key=www.example.com.key --cert=www.example.com.crt
```

## Testing

### Clusters without External Loadbalancer Support

Port forward to the Envoy service:
```shell
kubectl -n envoy-gateway-system port-forward service/envoy-default-eg 8043:8443 &
```

Curl the example app through Envoy proxy:
```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8043:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8043/get
```
You can replace `get` with any of the supported [httpbin methods][httpbin_methods].

### Clusters with External Loadbalancer Support

Get the External IP of the Gateway:
```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Curl the example app through the Gateway:
```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com:8443/get
```
You can replace `get` with any of the supported [httpbin methods][httpbin_methods].

## Multiple HTTPS Listeners
Due to [Issue 520][], multiple HTTP listeners must use different port numbers. For example:
```shell
cat <<EOF | kubectl apply -f -
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
    - name: https-2
      protocol: HTTPS
      port: 8444
      tls:
        mode: Terminate
        certificateRefs:
          - kind: Secret
            group: ""
            name: example-cert-2
 EOF
```
Store the previously created cert/key in Secret "example-cert-2":
```shell
kubectl create secret tls example-cert-2 --key=www.example.com.key --cert=www.example.com.crt
```

Follow the steps in the [Testing section](#testing) to test connectivity to the backend app through both Gateway
listeners.

## Cross Namespace Certificate References
A Gateway can be configured to reference a certificate in a different namespace. This is allowed by a [ReferenceGrant][]
created in the target namespace. Without the ReferenceGrant, the cross-namespace reference would be invalid.

To demonstrate cross namespace certificate references, create a ReferenceGrant that allows Gateways from the "default"
namespace to reference Secrets in the "envoy-gateway-system" namespace:
```shell
cat <<EOF | kubectl apply -f -
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

Recreate the example Secret in the "envoy-gateway-system" namespace:
```shell
kubectl create secret tls example-cert -n envoy-gateway-system --key=www.example.com.key --cert=www.example.com.crt
```

Update the Gateway HTTPS listener with `namespace: envoy-gateway-system`, for example:
```shell
cat <<EOF | kubectl apply -f -
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

Lastly, test connectivity using the above [Testing section](#testing).

## Clean-Up
Follow the steps from the [Quickstart Guide](QUICKSTART.md) to uninstall Envoy Gateway and the example manifest.

Delete the Secrets:
```shell
kubectl delete secret/example-cert
kubectl delete secret/example-cert-2
```

## Next Steps
Checkout the [Developer Guide](../../DEVELOPER.md) to get involved in the project.

[kind]: https://kind.sigs.k8s.io/
[httpbin_methods]: https://httpbin.org/#/HTTP_Methods
[Issue 520]: https://github.com/envoyproxy/gateway/issues/520
[ReferenceGrant]: https://gateway-api.sigs.k8s.io/api-types/referencegrant/
