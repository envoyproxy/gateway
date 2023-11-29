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

Follow the steps from the [Quickstart](../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Configuration

Envoy Gateway uses [.htpasswd][.htpasswd] format to store the username-password pairs for authentication.
The file must be stored in a kubernetes secret and referenced in the [SecurityPolicy][SecurityPolicy] configuration. 
The secret is an Opaque secret, and the username-password pairs must be stored in the key ".htpasswd".

### Create a .htpasswd file
First, create a [.htpasswd][.htpasswd] file with the username and password you want to use for authentication. 

The input password won't be saved, instead, a hash will be generated and saved in the output file. When a request
tries to access protected resources, the password in the "Authorization" HTTP header will be hashed and compared with the 
saved hash.

Note: only SHA hash algorithm is supported for now.

```shell
$ htpasswd -cbs .htpasswd foo bar
Adding password for user foo
```

You can also add more users to the file:

```shell
$ htpasswd -bs .htpasswd foo1 bar1
```

### Create a kubernetes secret

Next, create a kubernetes secret with the generated .htpasswd file in the previous step.

```shell
$ kubectl create secret generic basic-auth --from-file=.htpasswd
secret "basic-auth" created
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

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../quickstart) guide is set. If not, follow the
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
...
< HTTP/1.1 401 Unauthorized
< content-length: 58
< content-type: text/plain
< date: Tue, 28 Nov 2023 12:43:32 GMT
< server: envoy
< 
* Connection #0 to host 127.0.0.1 left intact
User authentication failed. Missing username and password.
```

Send a request to the backend service with `Authentication` header:

```shell
curl -v -H "Host: www.example.com" -u 'foo:bar' "http://${GATEWAY_HOST}/"
```

The request should be allowed and you should see the response from the backend service.

```shell

## Clean-Up

Follow the steps from the [Quickstart](../quickstart) guide to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy and the secret

```shell
kubectl delete securitypolicy/basic-auth-example
kubectl delete secret/basic-auth
```

## Next Steps

Checkout the [Developer Guide](../../contributions/develop/) to get involved in the project.

[SecurityPolicy]: ../../design/security-policy/
[http Basic authentication]: https://tools.ietf.org/html/rfc2617
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
[.htpasswd]: https://httpd.apache.org/docs/current/programs/htpasswd.html
