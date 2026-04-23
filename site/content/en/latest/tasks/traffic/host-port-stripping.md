---
title: "Host Normalization"
---

This task explains how to configure host normalization settings
using the [ClientTrafficPolicy][] API.

## Prerequisites

{{< boilerplate prerequisites >}}

## Stripping Port from the Host Header

Some clients include the port in the `Host` (or `:authority`) header
(e.g. `example.com:443`). Use `host.stripPortMode` to have Envoy strip the port
before route matching.

- `Any` strips the port unconditionally.
- `Matching` strips only when the port matches the listener port.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: strip-host-port
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  host:
    stripPortMode: Any
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
  name: strip-host-port
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  host:
    stripPortMode: Any
```

{{% /tab %}}
{{< /tabpane >}}

Verify the policy is accepted:

```shell
kubectl get clienttrafficpolicies.gateway.envoyproxy.io -n default
```

```
NAME               STATUS     AGE
strip-host-port    Accepted   5s
```

## Stripping Trailing Dot from the Host Header

Envoy does not strip trailing dots from the `Host` header by default, unlike some
other proxies (e.g. NGINX). This means requests with `Host: example.com.` will not
match routes with domains set to `example.com`. Use `host.stripTrailingHostDot` to
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
