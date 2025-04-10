---
title: "Dummy"
---

This task provides instructions for configuring dummy.

Dummy allows you to [briefly explain what this feature does and its benefits].

## Prerequisites

{{< boilerplate prerequisites >}}

## Installation

Install the necessary resources for dummy:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/[example-file].yaml
```

Alternatively, you can create the resources directly:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name:  
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /example
    backendRefs:
    - name: backend
      port: 3000
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name:  
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /example
    backendRefs:
    - name: backend
      port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

## Verification

Verify the configuration:

```shell
kubectl get httproute/  -o yaml
```

The output should include a status section indicating that the HTTPRoute has been accepted by the Gateway.

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Send a request to the endpoint:

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/example"
```

You should see a successful response, confirming that your configuration is working correctly.

Expected output:
```
< HTTP/1.1 200 OK
< content-type: application/json
...
```

## Troubleshooting

If you encounter issues, check the following:

1. Verify the Gateway status:
   ```shell
   kubectl get gateway/eg -o yaml
   ```

2. Look for errors in the resource status:
   ```shell
   kubectl get httproute/  -o jsonpath='{.status}'
   ```

3. Check the Envoy Gateway logs:
   ```shell
   kubectl logs -n envoy-gateway-system deployment/envoy-gateway
   ```

## Clean-Up

Remove the resources created in this task:

```shell
kubectl delete httproute/ 
# Add other resources that need to be cleaned up
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.
