---
title: "Proxy Observability"
---

Envoy Gateway provides observability for the ControlPlane and the underlying EnvoyProxy instances.
This guide show you how to config proxy observability, includes metrics, logs, and traces.

## Prerequisites

Follow the steps from the [Quickstart Guide](../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

[FluentBit](https://fluentbit.io/) is used to collect logs from the EnvoyProxy instances and forward them to Loki. Install FluentBit:

```shell
helm repo add fluent https://fluent.github.io/helm-charts
helm repo update
helm upgrade --install fluent-bit fluent/fluent-bit -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/fluent-bit/helm-values.yaml -n monitoring --create-namespace --version 0.30.4
```

[Loki](https://grafana.com/oss/loki/) is used to store logs. Install Loki:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/loki/loki.yaml -n monitoring 
```

[Tempo](https://grafana.com/oss/tempo/) is used to store traces. Install Tempo:

```shell
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm upgrade --install tempo grafana/tempo -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/tempo/helm-values.yaml -n monitoring --create-namespace --version 1.3.1
```

[OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) offers a vendor-agnostic implementation of how to receive, process and export telemetry data. 
Install OTel-Collector:

```shell
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo update
helm upgrade --install otel-collector open-telemetry/opentelemetry-collector -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/otel-collector/helm-values.yaml -n monitoring --create-namespace --version 0.60.0
```

Expose endpoints:

```shell
LOKI_IP=$(kubectl get svc loki -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
TEMPO_IP=$(kubectl get svc tempo -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Metrics

By default, Envoy Gateway expose metrics with prometheus endpoint. 

Verify metrics:

```shell
export ENVOY_POD_NAME=$(kubectl get pod -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward pod/$ENVOY_POD_NAME -n envoy-gateway-system 19001:19001

# check metrics 
curl localhost:19001/stats/prometheus  | grep "default/backend/rule/0/match/0-www"
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
curl localhost:19001/metrics  | grep "default/backend/rule/0/match/0-www"
```

## Logs

If custom format string is not specified, Envoy Gateway uses the following default format:

```json
{"start_time":"%START_TIME%","method":"%REQ(:METHOD)%","x-envoy-origin-path":"%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%","protocol":"%PROTOCOL%","response_code":"%RESPONSE_CODE%","response_flags":"%RESPONSE_FLAGS%","response_code_details":"%RESPONSE_CODE_DETAILS%","connection_termination_details":"%CONNECTION_TERMINATION_DETAILS%","upstream_transport_failure_reason":"%UPSTREAM_TRANSPORT_FAILURE_REASON%","bytes_received":"%BYTES_RECEIVED%","bytes_sent":"%BYTES_SENT%","duration":"%DURATION%","x-envoy-upstream-service-time":"%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)%","x-forwarded-for":"%REQ(X-FORWARDED-FOR)%","user-agent":"%REQ(USER-AGENT)%","x-request-id":"%REQ(X-REQUEST-ID)%",":authority":"%REQ(:AUTHORITY)%","upstream_host":"%UPSTREAM_HOST%","upstream_cluster":"%UPSTREAM_CLUSTER%","upstream_local_address":"%UPSTREAM_LOCAL_ADDRESS%","downstream_local_address":"%DOWNSTREAM_LOCAL_ADDRESS%","downstream_remote_address":"%DOWNSTREAM_REMOTE_ADDRESS%","requested_server_name":"%REQUESTED_SERVER_NAME%","route_name":"%ROUTE_NAME%"}
```

> Note: Envoy Gateway disable envoy headers by default, you can enable it by setting `EnableEnvoyHeaders` to `true` in the [ClientTrafficPolicy](../../api/extension_types#backendtrafficpolicy) CRD.


Verify logs from loki:

```shell
curl -s "http://$LOKI_IP:3100/loki/api/v1/query_range" --data-urlencode "query={job=\"fluentbit\"}" | jq '.data.result[0].values'
```

If you want to disable it, set the `telemetry.accesslog.disable` to `true` in the `EnvoyProxy` CRD.

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/accesslog/disable-accesslog.yaml
```

Envoy Gateway can send logs to OpenTelemetry Sink.

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/accesslog/otel-accesslog.yaml
```

Verify logs from loki:

```shell
curl -s "http://$LOKI_IP:3100/loki/api/v1/query_range" --data-urlencode "query={exporter=\"OTLP\"}" | jq '.data.result[0].values'
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
