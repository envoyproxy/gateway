---
title: "CORS"
---

This task provides instructions for configuring [Cross-Origin Resource Sharing (CORS)][cors] on Envoy Gateway.
CORS defines a way for client web applications that are loaded in one domain to interact with resources in a different
domain.

Envoy Gateway introduces a new CRD called [SecurityPolicy][] that allows the user to configure CORS.
This instantiated resource can be linked to a [Gateway][], [HTTPRoute][] or [GRPCRoute][] resource.

You can also configure CORS using the Gateway API's [HTTPCORSFilter][], which offers a simpler option but
is only supported on the [HTTPRoute][] resource.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

When configuring CORS either an origin with a precise hostname can be configured or an hostname containing a wildcard prefix,
allowing all subdomains of the specified hostname.
In addition to that the entire origin (with or without specifying a scheme) can be a wildcard to allow all origins.

### Configuring CORS with SecurityPolicy

The below example defines a SecurityPolicy that allows CORS for requests originating from `http://*.foo.com`.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: cors-example
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  cors:
    allowOrigins:
    - "http://*.foo.com"
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

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: cors-example
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  cors:
    allowOrigins:
    - "http://*.foo.com"
    allowMethods:
    - GET
    - POST
    allowHeaders:
    - "x-header-1"
    - "x-header-2"
    exposeHeaders:
    - "x-header-3"
    - "x-header-4"
```

{{% /tab %}}
{{< /tabpane >}}

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/cors-example -o yaml
```

### Configuring CORS with HTTPCORSFilter

Alternatively, you can configure CORS using the Gateway API's `HTTPCORSFilter`. This filter can be directly configured
within an `HTTPRoute` resource, which is simpler if you only need to apply CORS to a specific route.

The below example applies CORS to the `backend` HTTPRoute, allowing requests from `http://*.foo.com`.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
spec:
  hostnames:
  - www.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
  rules:
  - backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
    filters:
    - type: CORS
      cors:
        allowOrigins:
        - "http://*.foo.com"
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

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
spec:
  hostnames:
  - www.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
  rules:
  - backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
    filters:
    - type: CORS
      cors:
        allowOrigins:
        - "http://*.foo.com"
        allowMethods:
        - GET
        - POST
        allowHeaders:
        - "x-header-1"
        - "x-header-2"
        exposeHeaders:
        - "x-header-3"
        - "x-header-4"
```

{{% /tab %}}
{{< /tabpane >}}

Verify the HTTPRoute configuration:

```shell
kubectl get httproute/backend -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the
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

If you try to send a request from `http://www.foo.com:8080`, you should also see similar response because the port number
`8080` is not included in the allowed origins.

```shell
```shell
curl -H "Origin: http://www.foo.com:8080" \
  -H "Host: www.example.com" \
  -H "Access-Control-Request-Method: GET" \
  -X OPTIONS -v -s \
  http://$GATEWAY_HOST \
  1> /dev/null
```

Note:
* CORS specification requires that the browsers to send a preflight request to the server to ask if it's allowed
to access the limited resource in another domains. The browsers are supposed to follow the response from the server to
determine whether to send the actual request or not. The CORS filter only response to the preflight requests according to
its configuration. It won't deny any requests. The browsers are responsible for enforcing the CORS policy.
* The targeted HTTPRoute or the HTTPRoutes that the targeted Gateway routes to must allow the OPTIONS method for the CORS
filter to work. Otherwise, the OPTIONS request won't match the routes and the CORS filter won't be invoked.


## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy:

```shell
kubectl delete securitypolicy/cors-example
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[SecurityPolicy]: ../../../api/extension_types#securitypolicy
[cors]: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
[HTTPCORSFilter]: https://gateway-api.sigs.k8s.io/reference/spec#httpcorsfilter
