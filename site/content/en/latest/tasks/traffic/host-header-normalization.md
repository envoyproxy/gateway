---
title: "Host Normalization"
---

This task explains how to configure host normalization settings
using the [ClientTrafficPolicy][] API.

## Prerequisites

{{< boilerplate prerequisites >}}

## Stripping Trailing Dot from the Host Header

Envoy does not strip trailing dots from the `Host` header by default, unlike some
other proxies (e.g. NGINX). This means requests with `Host: example.com.` will not
match routes with domains set to `example.com`. Use `headers.host.stripTrailingHostDot` to
normalize these requests.

When the host includes a port (e.g. `example.com.:443`), only the trailing dot from
the host section is stripped, leaving the port as-is (`example.com:443`).

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: strip-trailing-dot
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  headers:
    host:
      stripTrailingHostDot: true
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
  name: strip-trailing-dot
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  headers:
    host:
      stripTrailingHostDot: true
```

{{% /tab %}}
{{< /tabpane >}}

Verify the policy is accepted:

```shell
kubectl get clienttrafficpolicies.gateway.envoyproxy.io -n default
```

```
NAME                  STATUS     AGE
strip-trailing-dot    Accepted   5s
```

[ClientTrafficPolicy]: ../../../api/extension_types#clienttrafficpolicy
