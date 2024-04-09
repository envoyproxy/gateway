---
title: "Basic Authentication"
---

This guide provides instructions for configuring [HTTP Basic authentication][http Basic authentication]. 
HTTP Basic authentication checks if an incoming request has a valid username and password before routing the request to 
a backend service.

Envoy Gateway introduces a new CRD called [SecurityPolicy][SecurityPolicy] that allows the user to configure HTTP Basic 
authentication. 
This instantiated resource can be linked to a [Gateway][Gateway], [HTTPRoute][HTTPRoute] or [GRPCRoute][GRPCRoute] resource.

## Prerequisites

Follow the steps from the [Quickstart](../../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Configuration

Envoy Gateway uses [.htpasswd][.htpasswd] format to store the username-password pairs for authentication.
The file must be stored in a kubernetes secret and referenced in the [SecurityPolicy][SecurityPolicy] configuration. 
The secret is an Opaque secret, and the username-password pairs must be stored in the key ".htpasswd".

### Create a root certificate

Create a root certificate and private key to sign certificates:

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout example.com.key -out example.com.crt
```

### Create a certificate secret

Create a certificate and a private key for `www.example.com`:

```shell
openssl req -out www.example.com.csr -newkey rsa:2048 -nodes -keyout www.example.com.key -subj "/CN=www.example.com/O=example organization"
openssl x509 -req -days 365 -CA example.com.crt -CAkey example.com.key -set_serial 0 -in www.example.com.csr -out www.example.com.crt
```

### Create certificate 

```shell
kubectl create secret tls example-cert --key=www.example.com.key --cert=www.example.com.crt
```

### Enable HTTPS
Update the Gateway from the Quickstart guide to include an HTTPS listener that listens on port `443` and references the
`example-cert` Secret:

```shell
kubectl patch gateway eg --type=json --patch '[{
   "op": "add",
   "path": "/spec/listeners/-",
   "value": {
      "name": "https",
      "protocol": "HTTPS",
      "port": 443,
      "tls": {
        "mode": "Terminate",
        "certificateRefs": [{
          "kind": "Secret",
          "group": "",
          "name": "example-cert",
        }],
      },
    },
}]'
```

### Create a .htpasswd file
First, create a [.htpasswd][.htpasswd] file with the username and password you want to use for authentication. 

Note: Please always use HTTPS with Basic Authentication. This prevents credentials from being transmitted in plain text.

The input password won't be saved, instead, a hash will be generated and saved in the output file. When a request
tries to access protected resources, the password in the "Authorization" HTTP header will be hashed and compared with the 
saved hash.

Note: only SHA hash algorithm is supported for now.

```shell
htpasswd -cbs .htpasswd foo bar
```

You can also add more users to the file:

```shell
htpasswd -bs .htpasswd foo1 bar1
```

### Create a basic-auth secret


Next, create a kubernetes secret with the generated .htpasswd file in the previous step.

```shell
kubectl create secret generic basic-auth --from-file=.htpasswd
```

### Create a SecurityPolicy

The below example defines a SecurityPolicy that authenticates requests against the user list in the kubernetes
secret generated in the previous step.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: basic-auth-example
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  basicAuth:
    users:
      name: "basic-auth"
EOF
```

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/basic-auth-example -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) guide is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Send a request to the backend service without `Authentication` header:

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/"
```

You should see `401 Unauthorized` in the response, indicating that the request is not allowed without authentication.

```shell
* Connected to 127.0.0.1 (127.0.0.1) port 443
...
* Server certificate:
*  subject: CN=www.example.com; O=example organization
*  issuer: O=example Inc.; CN=example.com
> GET / HTTP/2
> Host: www.example.com
> User-Agent: curl/8.6.0
> Accept: */*
...
< HTTP/2 401
< content-length: 58
< content-type: text/plain
< date: Wed, 06 Mar 2024 15:59:36 GMT
<

* Connection #0 to host 127.0.0.1 left intact
User authentication failed. Missing username and password.
```

Send a request to the backend service with `Authentication` header:

```shell
curl -kv -H "Host: www.example.com" -u 'foo:bar' "https://${GATEWAY_HOST}/" 
```

The request should be allowed and you should see the response from the backend service.

```shell

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) guide to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy and the secret

```shell
kubectl delete securitypolicy/basic-auth-example
kubectl delete secret/basic-auth
kubectl delete secret/example-cert
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop/) to get involved in the project.

[SecurityPolicy]: ../../contributions/design/security-policy/
[http Basic authentication]: https://tools.ietf.org/html/rfc2617
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
[.htpasswd]: https://httpd.apache.org/docs/current/programs/htpasswd.html
