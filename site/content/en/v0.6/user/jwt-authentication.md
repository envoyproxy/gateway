---
title: "JWT Authentication"
---

This guide provides instructions for configuring [JSON Web Token (JWT)][jwt] authentication. JWT authentication checks
if an incoming request has a valid JWT before routing the request to a backend service. Currently, Envoy Gateway only
supports validating a JWT from an HTTP header, e.g. `Authorization: Bearer <token>`.

Envoy Gateway introduces a new CRD called [SecurityPolicy][SecurityPolicy] that allows the user to configure JWT authentication. 
This instantiated resource can be linked to a [Gateway][Gateway], [HTTPRoute][HTTPRoute] or [GRPCRoute][GRPCRoute] resource.

## Prerequisites

Follow the steps from the [Quickstart](../quickstart) guide to install Envoy Gateway and the example manifest.
For GRPC - follow the steps from the [GRPC Routing](../grpc-routing/) example.
Before proceeding, you should be able to query the example backend using HTTP or GRPC.

## Configuration

Allow requests with a valid JWT by creating an [SecurityPolicy][SecurityPolicy] and attaching it to the example
HTTPRoute or GRPCRoute.

### HTTPRoute

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.6.0/examples/kubernetes/jwt/jwt.yaml
```

Two HTTPRoute has been created, one for `/foo` and another for `/bar`. A SecurityPolicy has been created and targeted
HTTPRoute foo to authenticate requests for `/foo`. The HTTPRoute bar is not targeted by the SecurityPolicy and will allow   
unauthenticated requests to `/bar`.

Verify the HTTPRoute configuration and status:

```shell
kubectl get httproute/foo -o yaml
kubectl get httproute/bar -o yaml
```

The SecurityPolicy is configured for JWT authentication and uses a single [JSON Web Key Set (JWKS)][jwks]
provider for authenticating the JWT.

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/jwt-example -o yaml
```

### GRPCRoute

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.6.0/examples/kubernetes/jwt/grpc-jwt.yaml
```

A SecurityPolicy has been created and targeted GRPCRoute yages to authenticate all requests for `yages` service..

Verify the GRPCRoute configuration and status:

```shell
kubectl get grpcroute/yages -o yaml
```

The SecurityPolicy is configured for JWT authentication and uses a single [JSON Web Key Set (JWKS)][jwks]
provider for authenticating the JWT.

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/jwt-example -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../quickstart) guide is set. If not, follow the
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
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/test.jwt -s) && echo "$TOKEN" | cut -d '.' -f2 - | base64 --decode -
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
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/test.jwt -s) && echo "$TOKEN" | cut -d '.' -f2 - | base64 --decode -
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

Follow the steps from the [Quickstart](../quickstart) guide to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy:

```shell
kubectl delete securitypolicy/jwt-example
```

## Next Steps

Checkout the [Developer Guide](../../contributions/develop/) to get involved in the project.

[SecurityPolicy]: ../../design/security-policy
[jwt]: https://tools.ietf.org/html/rfc7519
[jwks]: https://tools.ietf.org/html/rfc7517
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
