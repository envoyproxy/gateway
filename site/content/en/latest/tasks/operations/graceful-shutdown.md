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
     when no readiness delay is configured.
   - Envoy continues to serve accepted connections while listeners are draining.
3. Shutdown manager fails readiness and health checks via `/healthcheck/fail`.
   - By default this happens immediately and also starts listener drain.
   - When `readinessFailureDelay` is configured, this step is delayed without
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
- `readinessFailureDelay`: 0 seconds - Optional delay before failing readiness after drain starts

Configure `readinessFailureDelay` when failing readiness immediately would leave
the node without ready local endpoints before upstream load balancers have
stopped sending traffic to it. This can be important when using
`externalTrafficPolicy: Local` on the Kubernetes Service, where a node without
ready local endpoints can stop receiving traffic even though the terminating
Envoy pod is still able to serve connections during its drain period.

`readinessFailureDelay` does not extend the drain sequence or keep the pod's
containers running. If the drain completes before `readinessFailureDelay`
elapses, `/healthcheck/fail` is not called. This can happen when connections
drop below `exitAtConnections` after `minDrainDuration`, or when
`readinessFailureDelay` is greater than or equal to `drainTimeout`.

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
    readinessFailureDelay: "40s"     # Override default 0s
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
    readinessFailureDelay: "40s"     # Override default 0s
```

{{% /tab %}}
{{< /tabpane >}}
