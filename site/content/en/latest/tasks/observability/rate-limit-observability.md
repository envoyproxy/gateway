---
title: "RateLimit Observability"
---

Envoy Gateway provides observability for the RateLimit instances.
This guide show you how to config RateLimit observability, includes traces.

## Prerequisites

Follow the steps from the [Quickstart Guide](../quickstart) to install Envoy Gateway and the HTTPRoute example manifest.
Before proceeding, you should be able to query the example backend using HTTP. Follow the steps from the [Global Rate Limit](../traffic/global-rate-limit) to install RateLimit.


[OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) offers a vendor-agnostic implementation of how to receive, process and export telemetry data.
Install OTel-Collector:

```shell
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo update
helm upgrade --install otel-collector open-telemetry/opentelemetry-collector -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/otel-collector/helm-values.yaml -n monitoring --create-namespace --version 0.60.0
```

## Traces

By default, the Envoy Gateway does not configure RateLimit to send traces to the OpenTelemetry Sink.
You can configure the collector in the `rateLimit.telemetry.tracing` of the `EnvoyGateway`CRD.

RateLimit uses the OpenTelemetry Exporter to export traces to the collector.
You can configure a collector that supports the OTLP protocol, which includes but is not limited to: OpenTelemetry Collector, Jaeger, Zipkin, and so on.

***Note:***
* By default, the Envoy Gateway configures a 100% sampling rate for RateLimit, which may lead to performance issues.
* The Envoy Gateway constructs the Kubernetes FQDN using the value of `BackendObjectReference`, which serves as the target endpoint for
  the RateLimit trace collector. The `BackendObjectReference` is configured through the collector Service. Please note, the configuration of collector Service
  using `Service.type=ExternalName` is currently not supported.

Assuming the OpenTelemetry Collector is running in the `observability` namespace, and it has a service named `otel-svc`,
we only want to sample `50%` of the trace data. We would configure it as follows:

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
          backendRef:
            name: otel-svc
            namespace: observability
EOF
```

After updating the ConfigMap, you will need to restart the envoy-gateway deployment so the configuration kicks in

```shell
kubectl rollout restart deployment envoy-gateway -n envoy-gateway-system
```