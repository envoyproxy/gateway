## Secure Gateways
This guide will help you get started with using secure Gateways. The guide uses a self-signed CA, so it should be used
for testing and demonstration purposes only.

### Prerequisites
- A Kubernetes cluster with `kubectl` context configured for the cluster.
- OpenSSL to generate TLS assets.

__Note:__ Envoy Gateway is tested against Kubernetes v1.24.0.

### Installation
Follow the steps from the [Quickstart Guide](QUICKSTART.md) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to curl the example backend using HTTP.

### TLS Certificates

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

Store the cert/keys in A Secret:
```shell
kubectl create secret tls example-cert --key=www.example.com.key --cert=www.example.com.crt
```

### Testing
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

### For clusters with External Loadbalancer support
You can also test the same functionality by sending traffic to the External IP of the Gateway:
```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Curl the example app through the Gateway, e.g. Envoy proxy:
```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com:8443/get
```
You can replace `get` with any of the supported [httpbin methods][httpbin_methods].

## Clean-Up
Follow the steps from the [Quickstart Guide](QUICKSTART.md) to uninstall Envoy Gateway and the example manifest.

Delete the Secret:
```shell
kubectl delete secret/example-cert
```

## Next Steps
Checkout the [Developer Guide](../../DEVELOPER.md) to get involved in the project.

[kind]: https://kind.sigs.k8s.io/
[httpbin_methods]: https://httpbin.org/#/HTTP_Methods
