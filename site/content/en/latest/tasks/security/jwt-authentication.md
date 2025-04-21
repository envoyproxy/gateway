---
title: "JWT Authentication"
---

This task provides instructions for configuring [JSON Web Token (JWT)][jwt] authentication. JWT authentication checks
if an incoming request has a valid JWT before routing the request to a backend service. Currently, Envoy Gateway only
supports validating a JWT from an HTTP header, e.g. `Authorization: Bearer <token>`.

Envoy Gateway introduces a new CRD called [SecurityPolicy][SecurityPolicy] that allows the user to configure JWT authentication.
This instantiated resource can be linked to a [Gateway][Gateway], [HTTPRoute][HTTPRoute] or [GRPCRoute][GRPCRoute] resource.

## Prerequisites

{{< boilerplate prerequisites >}}

For GRPC - follow the steps from the [GRPC Routing](../traffic/grpc-routing) example.

## Configuration

Allow requests with a valid JWT by creating an [SecurityPolicy][SecurityPolicy] and attaching it to the example
HTTPRoute or GRPCRoute.

### HTTPRoute

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/jwt/jwt.yaml
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
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/jwt/grpc-jwt.yaml
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

{{< boilerplate testing-the-configuration >}}

### HTTPRoute

Verify that requests to `/foo` are denied without a JWT:

```shell
curl -sS -o /dev/null -H "Host: www.example.com" -w "%{http_code}\n" http://$GATEWAY_HOST/foo
```

A `401` HTTP response code should be returned.

Get the JWT used for testing request authentication:

```shell
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/test.jwt -s) && echo "$TOKEN" | cut -d '.' -f2 - | base64 --decode
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
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/test.jwt -s) && echo "$TOKEN" | cut -d '.' -f2 - | base64 --decode
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

## Connect to a remote JWKS with Self-Signed Certificate

To connect to a remote JWKS with a self-signed certificate, you need to configure it using the [Backend] resource within the [SecurityPolicy]. Additionally, use the [BackendTLSPolicy] to specify the CA certificate required to authenticate the JWKS host.

The following example demonstrates how to configure the remote JWKS with a self-signed certificate.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: jwt-example
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: foo
  jwt:
    providers:
    - name: example
      remoteJWKS:
        backendRefs:
          - group: gateway.envoyproxy.io
            kind: Backend
            name: remote-jwks
            port: 443
        backendSettings:
          retry:
            numRetries: 3
            perRetry:
              backOff:
                baseInterval: 1s
                maxInterval: 5s
            retryOn:
              triggers: ["5xx", "gateway-error", "reset"]
        uri: https://foo.bar.com/jwks.json
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: remote-jwks
spec:
  endpoints:
  - fqdn:
      hostname: foo.bar.com
      port: 443
---
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: BackendTLSPolicy
metadata:
  name: remote-jwks-btls
spec:
  targetRefs:
  - group: gateway.envoyproxy.io
    kind: Backend
    name: remote-jwks
    sectionName: "443"
  validation:
    caCertificateRefs:
    - name: remote-jwks-server-ca
      group: ""
      kind: ConfigMap
    hostname: foo.bar.com
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: jwt-example
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: foo
  jwt:
    providers:
    - name: example
      remoteJWKS:
        backendRefs:
          - group: gateway.envoyproxy.io
            kind: Backend
            name: remote-jwks
            port: 443
        backendSettings:
          retry:
            numRetries: 3
            perRetry:
              backOff:
                baseInterval: 1s
                maxInterval: 5s
            retryOn:
              triggers: ["5xx", "gateway-error", "reset"]
        uri: https://foo.bar.com/jwks.json
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: remote-jwks
spec:
  endpoints:
  - fqdn:
      hostname: foo.bar.com
      port: 443
---
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: BackendTLSPolicy
metadata:
  name: remote-jwks-btls
spec:
  targetRefs:
  - group: gateway.envoyproxy.io
    kind: Backend
    name: remote-jwks
    sectionName: "443"
  validation:
    caCertificateRefs:
    - name: remote-jwks-server-ca
      group: ""
      kind: ConfigMap
    hostname: foo.bar.com
```

{{% /tab %}}
{{< /tabpane >}}

As shown in the example above, the [SecurityPolicy] resource is configured with a remote JWKS within its JWT settings. The `backendRefs` field references the [Backend] resource that defines the JWKS host. The [BackendTLSPolicy] resource specifies the CA certificate required to authenticate the JWKS host.

Additional connection settings for the remote JWKS host can be configured in the [backendSettings]. Currently, only the retry policy is supported.

For more information about [Backend] and [BackendTLSPolicy], refer to the [Backend Routing][backend-routing] and [Backend TLS: Gateway to Backend][backend-tls] tasks.

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy:

```shell
kubectl delete securitypolicy/jwt-example
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[SecurityPolicy]: ../../../contributions/design/security-policy
[jwt]: https://tools.ietf.org/html/rfc7519
[jwks]: https://tools.ietf.org/html/rfc7517
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
[Backend]: ../../../api/extension_types#backend
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[backend-routing]: ../traffic/backend
[backend-tls]: ../backend-tls
[BackendSettings]: ../../../api/extension_types/#clustersettings
