# RSA + ECDSA Dual stack certificates

This guide gives a walkthrough to generate RSA and ECDSA derived certificates and keys for the Server, which can then be configured in the Gateway listener, to terminate TLS traffic.

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

Follow the steps from the [Quickstart Guide](quickstart.md) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

This guide builds over the [Secure Gateways](secure-gateways.md) guide which uses a self-signed RSA derived Server certificate and private key.


## Pre-checks

While testing in Cluster without External LoadBalancer Support, we can query the example app through Envoy proxy while enforcing an RSA cipher.

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8443/get  -Isv --ciphers ECDHE-RSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2
```

At the same time, since the Secret configured at this point is an RSA based Secret, if we enforce the usage of an ECDSA cipher, the call should fail as follows

```shell
$ curl -v -HHost:www.example.com --resolve "www.example.com:8443:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8443/get  -Isv --ciphers ECDHE-ECDSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2

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

Moving forward in the doc, we will be configuring the existing Gateway listener to accept both kinds of ciphers.

## TLS Certificates

We will reuse the the CA certificate and key pair generated in the [Secure Gateways](secure-gateways.md) guide and use this CA itself to sign both RSA and ECDSA certificates of the Server. 
Note the CA certificate and key names are `example.com.crt` and `example.com.key` respectively.


Create an ECDSA certificate and a private key for `www.example.com`:

```shell
openssl ecparam -noout -genkey -name prime256v1 -out www.example.com.ecdsa.key
openssl req -new -SHA384 -key www.example.com.ecdsa.key -nodes -out www.example.com.ecdsa.csr -subj "/CN=www.example.com/O=example organization"
openssl x509 -req -SHA384  -days 365 -in www.example.com.ecdsa.csr -CA example.com.crt -CAkey example.com.key -CAcreateserial -out www.example.com.ecdsa.crt 
```

Store the cert/key in a Secret:

```shell
kubectl create secret tls example-cert --key=www.example.com.key --cert=www.example.com.crt

kubectl create secret tls example-cert-ecdsa --key=www.example.com.ecdsa.key --cert=www.example.com.ecdsa.crt
```

Patch the Gateway with this additional ECDSA Secret:

```shell
kubectl patch gateway eg --type=json --patch '[{
   "op": "add",
   "path": "/spec/listeners/1/tls/certificateRefs/-",
   "value": {
      "name": "example-cert-ecdsa",
    },
}]'
```

Verify the Gateway status:

```shell
kubectl get gateway/eg -o yaml
```

## Testing

Again, while testing in Cluster without External LoadBalancer Support, we can query the example app through Envoy proxy while enforcing an RSA cipher.


Query the example app through Envoy proxy while enforcing an RSA cipher, which should work as it did before:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8443/get  -Isv --ciphers ECDHE-RSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2
```

Additionally, querying the example app while enforcing an ECDSA cipher should also work now:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8443/get  -Isv --ciphers ECDHE-ECDSA-CHACHA20-POLY1305 --tlsv1.2 --tls-max 1.2
```
