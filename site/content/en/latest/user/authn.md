---
title: "Request Authentication"
---

This guide provides instructions for configuring [JSON Web Token (JWT)][jwt] authentication. JWT authentication checks
if an incoming request has a valid JWT before routing the request to a backend service. Currently, Envoy Gateway only
supports validating a JWT from an HTTP header, e.g. `Authorization: Bearer <token>`.

## Installation

Follow the steps from the [Quickstart](quickstart.md) guide to install Envoy Gateway and the example manifest.
For GRPC - follow the steps from the [GRPC Routing](grpc-routing.md) example.
Before proceeding, you should be able to query the example backend using HTTP or GRPC.

## Configuration

Allow requests with a valid JWT by creating an [AuthenticationFilter][] and referencing it from the example HTTPRoute or GRPCRoute.

### HTTPRoute

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/authn/jwt.yaml
```

The HTTPRoute is now updated to authenticate requests for `/foo` and allow unauthenticated requests to `/bar`. The
`/foo` route rule references an AuthenticationFilter that provides the JWT authentication configuration.

Verify the HTTPRoute configuration and status:

```shell
kubectl get httproute/backend -o yaml
```

The AuthenticationFilter is configured for JWT authentication and uses a single [JSON Web Key Set (JWKS)][jwks]
provider for authenticating the JWT.

Verify the AuthenticationFilter configuration:

```shell
kubectl get authenticationfilter/jwt-example -o yaml
```

### GRPCRoute

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/authn/grtpc-jwt.yaml
```

The GRPCRoute is now updated to authenticate all requests to `yages` service, by referencing an AuthenticationFilter that provides the JWT authentication configuration.

Verify the GRPCRoute configuration and status:

```shell
kubectl get grpcroute/yages -o yaml
```

The AuthenticationFilter is configured for JWT authentication and uses a single [JSON Web Key Set (JWKS)][jwks]
provider for authenticating the JWT.

Verify the AuthenticationFilter configuration:

```shell
kubectl get authenticationfilter/jwt-example -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](quickstart.md) guide is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

### HTTPRoute

Verify that requests to `/foo` are denied without a JWT:

```shell
curl -sS -o /dev/null -H "Host: www.example.com" -w "%{http_code}\n" http://$GATEWAY_HOST/foo
```

A `401` HTTP response code should be returned.

Get the JWT used for testing request authentication:

```shell
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/authn/test.jwt -s) && echo "$TOKEN" | cut -d '.' -f2 - | base64 --decode -
```

__Note:__ The above command decodes and returns the token's payload. You can replace `f2` with `f1` to view the token's
header.

Verify that a request to `/foo` with a valid JWT is allowed:

```shell
curl -sS -o /dev/null -H "Host: www.example.com" -H "Authorization: Bearer $TOKEN" -w "%{http_code}\n" http://$GATEWAY_HOST/foo
```

A `200` HTTP response code should be returned.

Verify that requests to `/bar` are allowed __without__ a JWT:

```shell
curl -sS -o /dev/null -H "Host: www.example.com" -w "%{http_code}\n" http://$GATEWAY_HOST/bar
```

### GRPCRoute

Verify that requests to `yages`service are denied without a JWT:

```shell
grpcurl -plaintext -authority=grpc-example.com ${GATEWAY_HOST}:80 yages.Echo/Ping
```

You should see the below response

```shell
Error invoking method "yages.Echo/Ping": rpc error: code = Unauthenticated desc = failed to query for service descriptor "yages.Echo": Jwt is missing
```

Get the JWT used for testing request authentication:

```shell
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/authn/test.jwt -s) && echo "$TOKEN" | cut -d '.' -f2 - | base64 --decode -
```

__Note:__ The above command decodes and returns the token's payload. You can replace `f2` with `f1` to view the token's
header.

Verify that a request to `yages` service with a valid JWT is allowed:

```shell
grpcurl -plaintext -H "authorization: Bearer $TOKEN" -authority=grpc-example.com ${GATEWAY_HOST}:80 yages.Echo/Ping
```

You should see the below response

```shell
{
  "text": "pong"
}
```

## Clean-Up

Follow the steps from the [Quickstart](quickstart.md) guide to uninstall Envoy Gateway and the example manifest.

Delete the AuthenticationFilter:

```shell
kubectl delete authenticationfilter/jwt-example
```

## Next Steps

Checkout the [Developer Guide](../../contributions/develop/) to get involved in the project.

[jwt]: https://tools.ietf.org/html/rfc7519
[AuthenticationFilter]: https://gateway.envoyproxy.io/latest/api/extension_types.html#authenticationfilter
[jwks]: https://tools.ietf.org/html/rfc7517
