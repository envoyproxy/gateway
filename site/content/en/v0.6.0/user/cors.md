---
title: "CORS"
---

This guide provides instructions for configuring [Cross-Origin Resource Sharing (CORS)][cors] on Envoy Gateway.
CORS defines a way for client web applications that are loaded in one domain to interact with resources in a different
domain.

Envoy Gateway introduces a new CRD called [SecurityPolicy][SecurityPolicy] that allows the user to configure CORS.
This instantiated resource can be linked to a [Gateway][Gateway], [HTTPRoute][HTTPRoute] or [GRPCRoute][GRPCRoute] resource.

## Prerequisites

Follow the steps from the [Quickstart](../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Configuration

The below example defines a SecurityPolicy that allows CORS requests from `www.foo.com`.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: cors-example
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  cors:
    allowOrigins:
    - type: Exact
      value: "www.foo.com"
    allowMethods:
    - GET
    - POST
    allowHeaders:
    - "x-header-1"
    - "x-header-2"
    exposeHeaders:
    - "x-header-3"
    - "x-header-4"
EOF
```

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/cors-example -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../quickstart) guide is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Verify that the CORS headers are present in the response of the OPTIONS request from `http://www.foo.com`:

```shell
curl -H "Origin: http://www.foo.com" \
  -H "Host: www.example.com" \
  -H "Access-Control-Request-Method: GET" \
  -X OPTIONS -v -s \
  http://$GATEWAY_HOST \
  1> /dev/null
```

You should see the below response, indicating that the request from `http://www.foo.com` is allowed:

```shell
< access-control-allow-origin: http://www.foo.com
< access-control-allow-methods: GET, POST
< access-control-allow-headers: x-header-1, x-header-2
< access-control-max-age: 86400
< access-control-expose-headers: x-header-3, x-header-4
```

If you try to send a request from `http://www.bar.com`, you should see the below response:

```shell
curl -H "Origin: http://www.bar.com" \
  -H "Host: www.example.com" \
  -H "Access-Control-Request-Method: GET" \
  -X OPTIONS -v -s \
  http://$GATEWAY_HOST \
  1> /dev/null
```

You won't see any CORS headers in the response, indicating that the request from `http://www.bar.com` was not allowed.

Note: CORS specification requires that the browsers to send a preflight request to the server to ask if it's allowed
to access the limited resource in another domains. The browsers are supposed to follow the response from the server to
determine whether to send the actual request or not. The CORS filter only response to the preflight requests according to
its configuration. It won't deny any requests. The browsers are responsible for enforcing the CORS policy.


## Clean-Up

Follow the steps from the [Quickstart](../quickstart) guide to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy:

```shell
kubectl delete securitypolicy/cors-example
```

## Next Steps

Checkout the [Developer Guide](../../contributions/develop/) to get involved in the project.

[SecurityPolicy]: ../../design/security-policy/
[cors]: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
