---
title: "API Key Authentication"
---

This task provides instructions for configuring API Key Authentication. 
API Key Authentication verifies whether an incoming request includes a valid API key in the header, parameter, or cookie before routing the request to 
a backend service.

Envoy Gateway introduces a new CRD called [SecurityPolicy][SecurityPolicy] that allows the user to configure Api Key 
authentication. 
This instantiated resource can be linked to a [Gateway][Gateway], [HTTPRoute][HTTPRoute] or [GRPCRoute][GRPCRoute] resource.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

API Key must be stored in a kubernetes secret and referenced in the [SecurityPolicy][SecurityPolicy] configuration.
The secret is an Opaque secret, with each API key stored under a key corresponding to the client ID.

### Create a API Key Secret

Create an Opaque Secret containing the client ID and its corresponding API key

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: apikey-secret
stringData:
  client1: supersecret
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: apikey-secret
stringData:
  client1: supersecret
```

{{% /tab %}}
{{< /tabpane >}}

### Create a SecurityPolicy

The below example defines a SecurityPolicy that authenticates requests against the client list in the kubernetes
secret created in the previous step.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: apikey-auth-example
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  apiKeyAuth:
    credentialRefs:
    - group: ""
      kind: Secret
      name: apikey-secret
    extractFrom:
    - headers:
      - x-api-key
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
  name: apikey-auth-example
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend
  apiKeyAuth:
    credentialRefs:
    - group: ""
      kind: Secret
      name: apikey-secret
    extractFrom:
    - headers:
      - x-api-key
```

{{% /tab %}}
{{< /tabpane >}}

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/apikey-auth-example -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Send a request to the backend service without `x-api-key` header:

```shell
curl -kv -H "Host: www.example.com" "http://${GATEWAY_HOST}/" 
```

You should see `401 Unauthorized` in the response, indicating that the request is not allowed without providing valid API Key in `x-api-key` header.

```shell
* Connected to 127.0.0.1 (127.0.0.1) port 80
...
> GET / HTTP/2
> Host: www.example.com
> User-Agent: curl/8.7.1
> Accept: */*
...
< HTTP/2 401
< content-length: 58
< content-type: text/plain
< date: Sun, 19 Jan 2025 12:55:39 GMT
<

* Connection #0 to host 127.0.0.1 left intact
Client authentication failed.
```

Send a request to the backend service with `x-api-key` header:

```shell
curl -v -H "Host: www.example.com" -H 'x-api-key: supersecret' "http://${GATEWAY_HOST}/" 
```

The request should be allowed and you should see the response from the backend service.

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy and the secret

```shell
kubectl delete securitypolicy/apikey-auth-example
kubectl delete secret/apikey-secret
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[SecurityPolicy]: ../../../contributions/design/security-policy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
