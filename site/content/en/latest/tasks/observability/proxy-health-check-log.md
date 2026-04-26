---
title: "Proxy Health Check Logs"
---

Envoy Gateway can log health check events for upstream clusters using the [`healthCheckLog`][ProxyHealthCheckLog]
field in the [EnvoyProxy][] CRD's [telemetry][ProxyTelemetry] section. Events are written as JSON to a
configured file sink using Envoy's
[`event_logger`](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/health_check.proto#envoy-v3-api-field-config-core-v3-healthcheck-event-logger)
mechanism and the
[`HealthCheckEventFileSink`](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/health_check/event_sinks/file/v3/file.proto)
extension.

> **Note:** Health check event logging only applies to xRoutes that have **active health
> checks** configured via a [BackendTrafficPolicy][].

## Prerequisites

{{< boilerplate prerequisites >}}

## Configure Active Health Checks

Health check event logs require active health checks to be running.
Configure a [BackendTrafficPolicy][] targeting your `HTTPRoute` with an active health
check. The example below polls `/healthz` every three seconds:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: backend-health-check
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend-route
  healthCheck:
    active:
      type: HTTP
      http:
        path: /healthz
      interval: 3s
      timeout: 1s
      unhealthyThreshold: 3
      healthyThreshold: 1
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: backend-health-check
  namespace: default
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: backend-route
  healthCheck:
    active:
      type: HTTP
      http:
        path: /healthz
      interval: 3s
      timeout: 1s
      unhealthyThreshold: 3
      healthyThreshold: 1
```

{{% /tab %}}
{{< /tabpane >}}

## Enable Health Check Event Logging

Configure health check event logging in the [EnvoyProxy][] CRD.
When no sinks are specified, events are written to `/dev/stdout` by default.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: hc-event-logging
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: hc-event-logging
  namespace: envoy-gateway-system
spec:
  telemetry:
    healthCheckLog: {}
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: hc-event-logging
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: hc-event-logging
  namespace: envoy-gateway-system
spec:
  telemetry:
    healthCheckLog: {}
```

{{% /tab %}}
{{< /tabpane >}}

To write events to a specific file instead, configure an explicit sink:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: hc-event-logging-file
  namespace: envoy-gateway-system
spec:
  telemetry:
    healthCheckLog:
      sinks:
        - type: File
          file:
            path: /var/log/envoy/health-check-events.log
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: hc-event-logging-file
  namespace: envoy-gateway-system
spec:
  telemetry:
    healthCheckLog:
      sinks:
        - type: File
          file:
            path: /var/log/envoy/health-check-events.log
```

{{% /tab %}}
{{< /tabpane >}}

Health check events will now appear in the Envoy proxy container's standard output
in JSON format, for example:

```json
{
  "health_checker_type": "HTTP",
  "host": {
    "socket_address": { "protocol": "TCP", "address": "1.2.3.4", "port_value": 8080 }
  },
  "cluster_name": "default/backend-route/rule/0/match/0/backend-route",
  "timestamp": "2024-01-15T10:23:00.123Z",
  "health_check_failure_event": {
    "failure_type": "ACTIVE",
    "first_check": false
  }
}
```

## Log All Events

When `matches` is omitted (the default), all health check probe outcomes are
logged. To log only specific outcomes, set `matches` to one or more values;
they are ORed together. At least one failure variant and one success variant must
be specified together.

| Value | Logged when |
|---|---|
| `Failure` | Every failed probe, regardless of current health state |
| `FailureTransition` | Only when a host transitions from healthy → unhealthy |
| `Success` | Every successful probe, regardless of current health state |
| `SuccessTransition` | Only when a host transitions from unhealthy → healthy |

To log only on state transitions (the most conservative setting):

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: hc-event-logging-transitions
  namespace: envoy-gateway-system
spec:
  telemetry:
    healthCheckLog:
      matches:
        - FailureTransition
        - SuccessTransition
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: hc-event-logging-transitions
  namespace: envoy-gateway-system
spec:
  telemetry:
    healthCheckLog:
      matches:
        - FailureTransition
        - SuccessTransition
```

{{% /tab %}}
{{< /tabpane >}}

To log every probe result regardless of outcome:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: hc-event-logging-verbose
  namespace: envoy-gateway-system
spec:
  telemetry:
    healthCheckLog:
      matches:
        - Failure
        - Success
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: hc-event-logging-verbose
  namespace: envoy-gateway-system
spec:
  telemetry:
    healthCheckLog:
      matches:
        - Failure
        - Success
```

{{% /tab %}}
{{< /tabpane >}}

## Verify

Trigger a health check failure (e.g. by scaling the backend deployment to zero
replicas) and confirm events appear in the proxy logs:

```shell
kubectl logs -l gateway.envoyproxy.io/owning-gateway-name=eg -n envoy-gateway-system -c envoy | grep health_checker_type
```

[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[EnvoyProxy]: ../../../api/extension_types#envoyproxy
[ProxyHealthCheckLog]: ../../../api/extension_types#proxyhealthchecklog
[ProxyTelemetry]: ../../../api/extension_types#proxytelemetry
