---
title: "IP Allowlist/Denylist"
---

This task provides instructions for configuring IP allowlist/denylist on Envoy Gateway. IP allowlist/denylist
checks if an incoming request is from an allowed IP address before routing the request to a backend service.

Envoy Gateway introduces a new CRD called [SecurityPolicy][SecurityPolicy] that allows the user to configure IP allowlist/denylist.
This instantiated resource can be linked to a [Gateway][Gateway], [HTTPRoute][HTTPRoute] or [GRPCRoute][GRPCRoute] resource.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

### Create a SecurityPolicy

The below SecurityPolicy restricts access to the backend service by allowing requests only from the IP addresses `10.0.1.0/24`. 

In this example, the default action is set to `Deny`, which means that only requests from the specified IP addresses with `Allow`
action are allowed, and all other requests are denied. You can also change the default action to `Allow` to allow all requests 
except those from the specified IP addresses with `Deny` action.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: authorization-client-ip
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  authorization:
    defaultAction: Deny
    rules:
    - action: Allow
      principal:
        clientCIDRs:
        - 10.0.1.0/24
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
  name: authorization-client-ip
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  authorization:
    defaultAction: Deny
    rules:
    - action: Allow
      principal:
        clientCIDRs:
        - 10.0.1.0/24
```

{{% /tab %}}
{{< /tabpane >}}

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/authorization-client-ip -o yaml
```

### Original Source IP

It's important to note that the IP address used for allowlist/denylist is the original source IP address of the request.
You can use a [ClientTrafficPolicy] to configure how Envoy Gateway should determine the original source IP address.

For example, the below ClientTrafficPolicy configures Envoy Gateway to use the `X-Forwarded-For` header to determine the original source IP address.
The `numTrustedHops` field specifies the number of trusted hops in the `X-Forwarded-For` header. In this example, the `numTrustedHops` is set to `1`,
which means that the first rightmost IP address in the `X-Forwarded-For` header is used as the original source IP address.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-client-ip-detection
spec:
  clientIPDetection:
    xForwardedFor:
      numTrustedHops: 1
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-client-ip-detection
spec:
  clientIPDetection:
    xForwardedFor:
      numTrustedHops: 1
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
```

{{% /tab %}}
{{< /tabpane >}}


## Testing

{{< boilerplate testing-the-configuration >}}

Send a request to the backend service without the `X-Forwarded-For` header:

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/"
```

You should see `403 Forbidden` in the response, indicating that the request is not allowed.

```shell
* Connected to 172.18.255.200 (172.18.255.200) port 80
> GET /get HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.8.0-DEV
> Accept: */*
> 
* Request completely sent off
< HTTP/1.1 403 Forbidden
< content-length: 19
< content-type: text/plain
< date: Mon, 08 Jul 2024 04:23:31 GMT
< 
* Connection #0 to host 172.18.255.200 left intact
RBAC: access denied
```

Send a request to the backend service with the `X-Forwarded-For` header:

```shell
curl -v -H "Host: www.example.com" -H "X-Forwarded-For: 10.0.1.1" "http://${GATEWAY_HOST}/"
```

The request should be allowed and you should see the response from the backend service.

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy and the ClientTrafficPolicy

```shell
kubectl delete securitypolicy/authorization-client-ip
kubectl delete clientTrafficPolicy/enable-client-ip-detection
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[SecurityPolicy]: ../../../contributions/design/security-policy
[ClientTrafficPolicy]: ../../../api/extension_types#clienttrafficpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
