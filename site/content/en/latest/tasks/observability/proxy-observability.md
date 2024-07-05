---
title: "Proxy Observability"
---

Envoy Gateway provides observability for the ControlPlane and the underlying EnvoyProxy instances.
This task show you how to config proxy observability, includes metrics, and traces.

## Prerequisites

{{< boilerplate o11y_prerequisites >}}

Expose endpoints:

```shell
TEMPO_IP=$(kubectl get svc tempo -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Metrics

By default, Envoy Gateway expose metrics with prometheus endpoint. 

Verify metrics:

```shell
export ENVOY_POD_NAME=$(kubectl get pod -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward pod/$ENVOY_POD_NAME -n envoy-gateway-system 19001:19001

# check metrics 
curl localhost:19001/stats/prometheus  | grep "default/backend/rule/0"
```

You can disable metrics by setting the `telemetry.metrics.prometheus.disable` to `true` in the `EnvoyProxy` CRD.

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/metric/disable-prometheus.yaml
```

Envoy Gateway can send metrics to OpenTelemetry Sink.
Send metrics to OTel-Collector:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/metric/otel-sink.yaml
```

Verify OTel-Collector metrics:

```shell
export OTEL_POD_NAME=$(kubectl get pod -n monitoring --selector=app.kubernetes.io/name=opentelemetry-collector -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward pod/$OTEL_POD_NAME -n monitoring 19001:19001

# check metrics 
curl localhost:19001/metrics  | grep "default/backend/rule/0"
```

## Traces

By default, Envoy Gateway doesn't send traces to OpenTelemetry Sink.
You can enable traces by setting the `telemetry.tracing` in the `EnvoyProxy` CRD.

***Note:*** Envoy Gateway use 100% sample rate, which means all requests will be traced. This may cause performance issues.

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/tracing/default.yaml
```

Verify traces from tempo:

```shell
curl -s "http://$TEMPO_IP:3100/api/search" --data-urlencode "q={ component=envoy }" | jq .traces
```

```shell
curl -s "http://$TEMPO_IP:3100/api/traces/<trace_id>" | jq
```
