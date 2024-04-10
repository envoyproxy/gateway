---
title: "Request Authentication"
---

This guide provides instructions for configuring [JSON Web Token (JWT)][jwt] authentication. JWT authentication checks
if an incoming request has a valid JWT before routing the request to a backend service. Currently, Envoy Gateway only
supports validating a JWT from an HTTP header, e.g. `Authorization: Bearer <token>`.

## Installation

Follow the steps from the [Quickstart](../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Configuration

Allow requests with a valid JWT by creating an [AuthenticationFilter][] and referencing it from the example HTTPRoute.

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.5.0/examples/kubernetes/authn/jwt.yaml
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

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../quickstart) guide is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

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

## Clean-Up

Follow the steps from the [Quickstart](../quickstart) guide to uninstall Envoy Gateway and the example manifest.

Delete the AuthenticationFilter:

```shell
kubectl delete authenticationfilter/jwt-example
```

## Next Steps

Checkout the [Developer Guide](../../contributions/develop/) to get involved in the project.

[jwt]: https://tools.ietf.org/html/rfc7519
[AuthenticationFilter]: https://gateway.envoyproxy.io/v0.5.0/api/extension_types.html#authenticationfilter
[jwks]: https://tools.ietf.org/html/rfc7517
