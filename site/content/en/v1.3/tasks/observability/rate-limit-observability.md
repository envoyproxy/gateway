---
title: "RateLimit Observability"
---

Envoy Gateway provides observability for the RateLimit instances.
This guide show you how to config RateLimit observability, includes traces.

## Prerequisites

{{< boilerplate o11y_prerequisites >}}

Follow the steps from the [Global Rate Limit](../traffic/global-rate-limit) to install RateLimit.

## Traces

By default, the Envoy Gateway does not configure RateLimit to send traces to the OpenTelemetry Sink.
You can configure the collector in the `rateLimit.telemetry.tracing` of the `EnvoyGateway`CRD.

RateLimit uses the OpenTelemetry Exporter to export traces to the collector.
You can configure a collector that supports the OTLP protocol, which includes but is not limited to: OpenTelemetry Collector, Jaeger, Zipkin, and so on.

***Note:***

* By default, the Envoy Gateway configures a `100%` sampling rate for RateLimit, which may lead to performance issues.

Assuming the OpenTelemetry Collector is running in the `observability` namespace, and it has a service named `otel-svc`,
we only want to sample `50%` of the trace data. We would configure it as follows:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-gateway-config
  namespace: envoy-gateway-system
data:
  envoy-gateway.yaml: |
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyGateway
    provider:
      type: Kubernetes
    gateway:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
    rateLimit:
      backend:
        type: Redis
        redis:
          url: redis-service.default.svc.cluster.local:6379
      telemetry:
        tracing:
          sampleRate: 50
          provider:
            url: otel-svc.observability.svc.cluster.local:4318
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-gateway-config
  namespace: envoy-gateway-system
data:
  envoy-gateway.yaml: |
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyGateway
    provider:
      type: Kubernetes
    gateway:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
    rateLimit:
      backend:
        type: Redis
        redis:
          url: redis-service.default.svc.cluster.local:6379
      telemetry:
        tracing:
          sampleRate: 50
          provider:
            url: otel-svc.observability.svc.cluster.local:4318
```

{{% /tab %}}
{{< /tabpane >}}

{{< boilerplate rollout-envoy-gateway >}}
