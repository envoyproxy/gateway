---
title: "Example Task"
description: "This is an example task document using the new task archetype structure"
weight: 10
---

This task provides instructions for configuring an example feature in Envoy Gateway.

## Prerequisites

{{< boilerplate prerequisites >}}

## Objective

Learn how to implement a sample configuration using Envoy Gateway to demonstrate the task archetype.

## Procedure

### Step 1: Create the Configuration

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOT | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: example-task
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
EOT
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: example-task
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

Verify the configuration:

```shell
kubectl get httproute/example-task -o yaml
```

The output should include a status section indicating that the HTTPRoute has been accepted by the Gateway.

### Step 2: Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set.

```shell
echo $GATEWAY_HOST
```

Send a request to the endpoint:

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/example"
```

You should receive a response from the backend service, confirming that your route is working correctly.

### Step 3: Troubleshooting

If you encounter issues, check the following:

1. Verify the Gateway status:
   ```shell
   kubectl get gateway/eg -o yaml
   ```

2. Look for errors in the HTTPRoute status:
   ```shell
   kubectl get httproute/example-task -o jsonpath='{.status}'
   ```

3. Check the Envoy Gateway logs:
   ```shell
   kubectl logs -n envoy-gateway-system deployment/envoy-gateway
   ```

## Clean-Up

Remove the resources created in this task:

```shell
kubectl delete httproute/example-task
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project. 