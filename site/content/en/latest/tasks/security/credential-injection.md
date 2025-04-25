---
title: "Credential Injection"
---

This task shows how to use [HTTPRouteFilter][HTTPRouteFilter] to inject credentials into requests. It can be used to add
credentials such as Basic Authentication, JWTs, or API keys before the requests are sent to the backend service.

This is useful in scenarios where the backend service requires authentication or other credentials that are not provided
by the client. For example, you can use this feature to inject an access token into an API call to AWS.

Credentials can be injected at the HTTPRoute level or at the BackendRef level, allowing for fine-grained control over
which requests receive injected credentials.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

To inject credentials into requests, you need to create a [HTTPRouteFilter][HTTPRouteFilter] resource that defines the
credentials to be injected. The filter can be applied to an [HTTPRoute][HTTPRoute] or a [BackendRef][BackendRef] resource.

### Create a Secret with the Credential that you want to inject into requests

Create a secret with the credential you want to inject into requests. The secret should contain the credential in a field
named `credential`.

You can use any type of credential, such as a JWT token, Basic Authentication credentials, an API key,
or a custom credential. In this example, we will use a JWT token as the credential to be injected into requests.

```shell
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/test.jwt -s)
kubectl create secret generic jwt-credential --from-literal=credential=" Bearer $TOKEN"
```

### Create a HTTPRouteFilter

Create a `HTTPRouteFilter` resource that injects the JWT token into requests.

By default, the credentials will be injected into the `Authorization` header. You can also inject credentials into other
headers by specifying the `header` field in the `credentialInjection` section of the `HTTPRouteFilter`.

If the `Authorization` header or the specified header already exists in the request, the credentials won't be injected
unless you set `overwrite` to `true`.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: HTTPRouteFilter
metadata:
  name: credential-injection
spec:
  credentialInjection:
    overwrite: true
    credential:
      valueRef:
        name: jwt-credential
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: HTTPRouteFilter
metadata:
  name: credential-injection
spec:
  credentialInjection:
    overwrite: true
    credential:
      valueRef:
        name: jwt-credential
```

{{% /tab %}}
{{< /tabpane >}}

### Injecting Credentials into HTTPRoute

To inject the credentials at the [HTTPRoute][HTTPRoute] level, you need to reference the `HTTPRouteFilter` in the `filters`
section of the `HTTPRoute` resource. The credentials will be injected into all requests that match the route.

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
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: HTTPRouteFilter
        name: jwt-credential
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
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: HTTPRouteFilter
        name: jwt-credential
```

{{% /tab %}}
{{< /tabpane >}}

### Injecting Credentials into BackendRef
To inject the credentials at the [BackendRef][BackendRef] level, you need to reference the `HTTPRouteFilter` in the
`filters` section of the `BackendRef` resource. The credentials will be injected into all requests that are sent to that
backend.

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
      - type: ExtensionRef
        extensionRef:
          group: gateway.envoyproxy.io
          kind: HTTPRouteFilter
          name: credential-injection
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
      - type: ExtensionRef
        extensionRef:
          group: gateway.envoyproxy.io
          kind: HTTPRouteFilter
          name: credential-injection
```

{{% /tab %}}
{{< /tabpane >}}

## Testing

Send a request to the backend service without `Authentication` header:

```shell
curl -kv -H "Host: www.example.com" "http://${GATEWAY_HOST}/"
```

You should see the Authorization header with the JWT token in the echoed request headers in the response, indicating
that the token was successfully injected into the request.

```shell

"Authorization": [
   "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.NHVaYe26MbtOYhSKkoKYdFVomg4i8ZJd8_-RU8VNbftc4TSMb4bXP3l3YlNWACwyXPGffz5aXHc6lty1Y2t4SWRqGteragsVdZufDn5BlnJl9pdR_kdVFUsra2rWKEofkZeIC4yWytE58sMIihvo9H1ScmmVwBcQP6XETqYd0aSHp1gOa9RdUPDvoXQ5oqygTqVtxaDr6wUFKrKItgBMzWIdNZ6y7O9E0DhEPTbE9rfBo6KTFsHAZnMg4k68CDp2woYIaXbmYTWcvbzIuHO7_37GT79XdIwkm95QJ7hYC9RiwrV7mesbY4PAahERJawntho0my942XheVLmGwLMBkQ"
  ],

```

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy and the secret

```shell
kubectl delete httproutefilter/credential-injection
kubectl delete secret/jwt-credential
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[BackendRef]: https://gateway-api.sigs.k8s.io/reference/spec/#httpbackendref
[HTTPRouteFilter]: ../../../api/extension_types#httproutefilter
