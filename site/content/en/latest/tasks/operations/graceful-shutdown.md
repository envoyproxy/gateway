---
title: "Graceful Shutdown and Hitless Upgrades"
---

Envoy Gateway enables zero-downtime deployments through graceful connection draining during pod termination.

## Overview

The shutdown manager sidecar coordinates graceful connection draining during pod termination, providing:

- Zero-downtime rolling updates
- Configurable drain timeouts
- Automatic health check failure to remove pods from load balancer rotation

### Shutdown Process

1. Kubernetes sends SIGTERM to the pod's containers and marks the pod as
   terminating.
2. Shutdown manager starts Envoy listener drain.
   - Drain is initiated directly via
     `/drain_listeners?graceful&skip_exit` or indirectly via `/healthcheck/fail`
     when no health check failure delay is configured.
   - Envoy continues to serve accepted connections while listeners are draining.
3. Shutdown manager fails health checks via `/healthcheck/fail`, causing the
   pod's readiness probe to fail.
   - By default this happens immediately and also starts listener drain.
   - When `healthCheckFailureDelay` is configured, this step is delayed without
     delaying listener drain, connection monitoring, or the drain timeout.
4. Connection monitoring begins, polling `server.total_connections`
5. Process exits when connections reach zero or drain timeout is exceeded

## Configuration

Graceful shutdown behavior includes default values that can be overridden using the EnvoyProxy resource. The EnvoyProxy resource can be referenced in two ways:
1. **Gateway-level**: Referenced from a Gateway via `infrastructure.parametersRef`
2. **GatewayClass-level**: Referenced from a GatewayClass via `parametersRef`

**Default Values:**
- `drainTimeout`: 60 seconds - Maximum time for connection draining
- `minDrainDuration`: 10 seconds - Minimum wait before allowing exit
- `healthCheckFailureDelay`: 0 seconds - Optional delay before failing health checks after drain starts

`healthCheckFailureDelay` does not extend the drain sequence or keep the pod's
containers running. If the drain completes before `healthCheckFailureDelay`
elapses, `/healthcheck/fail` is not called. This can happen when connections
drop below `exitAtConnections` after `minDrainDuration`, or when
`healthCheckFailureDelay` is greater than or equal to `drainTimeout`.

{{< tabpane text=true >}}
{{% tab header="Gateway-Level Configuration" %}}

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  infrastructure:
    parametersRef:
      group: gateway.envoyproxy.io
      kind: EnvoyProxy
      name: graceful-shutdown-config
  listeners:
  - name: http
    port: 80
    protocol: HTTP
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: graceful-shutdown-config
spec:
  shutdown:
    drainTimeout: "90s"              # Override default 60s
    minDrainDuration: "15s"          # Override default 10s
    healthCheckFailureDelay: "40s"   # Override default 0s
```

{{% /tab %}}
{{% tab header="GatewayClass-Level Configuration" %}}

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: graceful-shutdown-config
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: graceful-shutdown-config
spec:
  shutdown:
    drainTimeout: "90s"              # Override default 60s
    minDrainDuration: "15s"          # Override default 10s
    healthCheckFailureDelay: "40s"   # Override default 0s
```

{{% /tab %}}
{{< /tabpane >}}

## Known Limitations

### hostNetwork deployments

When the Envoy proxy Deployment/DaemonSet is patched to run with `hostNetwork: true`, the envoy
container's PreStop lifecycle hook (an HTTP GET to the shutdown-manager's `/shutdown/ready`
endpoint) can fail with a kubelet error like `failed to find networking container`. This is
caused by a known kubelet bug
([kubernetes/kubernetes#134285](https://github.com/kubernetes/kubernetes/issues/134285)): for
hostNetwork pods, kubelet cannot resolve an implicit target address for the PreStop `httpGet`
action because the pod IP reported by the CRI is empty.

Because the PreStop hook never completes, shutdown-manager doesn't get a chance to drain
connections before Envoy exits, so in-flight connections can be dropped on pod termination.

As a workaround, you can patch the envoy container's PreStop `httpGet` action to target
`127.0.0.1` explicitly, which is reachable because the pod shares the node's network namespace:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: eg
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        patch:
          type: StrategicMerge
          value:
            spec:
              template:
                spec:
                  hostNetwork: true
                  containers:
                  - name: envoy
                    lifecycle:
                      preStop:
                        httpGet:
                          host: 127.0.0.1
```

Only set `host: 127.0.0.1` for hostNetwork pods. For the default (non-hostNetwork) case, kubelet
runs `httpGet` lifecycle hooks from the node's network namespace, so hardcoding `127.0.0.1` would
target the node instead of the pod and break the hook.
