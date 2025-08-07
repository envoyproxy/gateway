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

1. Kubernetes sends SIGTERM to the pod
2. Shutdown manager fails health checks via `/healthcheck/fail`
   - This causes Kubernetes readiness probes to fail
   - External load balancers and services stop routing new traffic to the pod
   - Existing connections continue to be served while draining
3. Connection monitoring begins, polling `server.total_connections`
4. Process exits when connections reach zero or drain timeout is exceeded

## Configuration

Graceful shutdown behavior includes default values that can be overridden using the EnvoyProxy resource referenced from a Gateway via `infrastructure.parametersRef`.

**Default Values:**
- `drainTimeout`: 60 seconds - Maximum time for connection draining
- `minDrainDuration`: 10 seconds - Minimum wait before allowing exit

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
    drainTimeout: "90s"      # Override default 60s
    minDrainDuration: "15s"  # Override default 10s
```
