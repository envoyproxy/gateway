---
title: "Request Header Size"
---

This task explains how to configure the maximum request header size
using the [ClientTrafficPolicy][] API.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuring Maximum Request Header Size

By default, Envoy rejects HTTP requests whose total header size exceeds 60 KiB, responding
with a `431 Request Header Fields Too Large` status. Use `headers.maxRequestHeaderBytes` in
a `ClientTrafficPolicy` to raise or lower this limit.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: request-header-size
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  headers:
    maxRequestHeaderBytes: 80Ki
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
  name: request-header-size
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  headers:
    maxRequestHeaderBytes: 80Ki
```

{{% /tab %}}
{{< /tabpane >}}

Verify the policy is accepted:

```shell
kubectl get clienttrafficpolicies.gateway.envoyproxy.io -n default
```

```
NAME                   STATUS     AGE
request-header-size    Accepted   5s
```

[ClientTrafficPolicy]: ../../../api/extension_types#clienttrafficpolicy
