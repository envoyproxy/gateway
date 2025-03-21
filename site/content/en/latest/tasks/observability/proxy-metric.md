---
title: "Proxy Metrics"
---

Envoy Gateway offers observability for both the Control Plane and the underlying Envoy Proxy instances.
This task shows you how to configure proxy metrics.

## Prerequisites

### Install Envoy Gateway

{{< boilerplate prerequisites >}}

### Install Add-ons

Envoy Gateway provides an add-ons Helm chart to simplify the installation of observability components.  
The documentation for the add-ons chart can be found
[here](https://gateway.envoyproxy.io/docs/install/gateway-addons-helm-api/).

Follow the instructions below to install the add-ons Helm chart.

```shell
helm install eg-addons oci://docker.io/envoyproxy/gateway-addons-helm --version {{< helm-version >}} -n monitoring --create-namespace
```

By default, the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) is **disabled.** 
To install add-ons with OpenTelemetry Collector enabled, use the following command.

```shell
helm install eg-addons oci://docker.io/envoyproxy/gateway-addons-helm --version {{< helm-version >}} --set opentelemetry-collector.enabled=true -n monitoring --create-namespace
```

## Metrics

### Prometheus Metrics

To query metrics using Prometheus API, follow the steps below.

```shell
export PROMETHEUS_PORT=$(kubectl get service prometheus -n monitoring -o jsonpath='{.spec.ports[0].port}')
kubectl port-forward service/prometheus -n monitoring 19001:$PROMETHEUS_PORT
```

Query metrics using Prometheus API:

```shell
curl -s 'http://localhost:19001/api/v1/query?query=topk(1,envoy_cluster_upstream_cx_connect_ms_sum)' | jq . 
```

To directly view the metrics in Prometheus format from the Envoy's `/stats/prometheus` 
[admin endpoint](https://www.envoyproxy.io/docs/envoy/latest/operations/admin), follow the steps below.

```shell
export ENVOY_POD_NAME=$(kubectl get pod -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward pod/$ENVOY_POD_NAME -n envoy-gateway-system 19001:19001
```

View the metrics:

```shell
curl localhost:19001/stats/prometheus  | grep "default/backend/rule/0"
```

If you are only using the OpenTelemetry sink, you might want to set the `telemetry.metrics.prometheus.disable` to `true`
in the _EnvoyProxy CRD_ as shown in the following command.

```yaml
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
    name: prometheus
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: prometheus
  namespace: envoy-gateway-system
spec:
  telemetry:
    metrics:
      prometheus:
        disable: true
EOF
```

To completely remove Prometheus resources from the cluster, set the `prometheus.enabled` Helm value to `false`.

```shell
helm upgrade eg-addons oci://docker.io/envoyproxy/gateway-addons-helm --version v0.0.0-latest -n monitoring --set prometheus.enabled=false 
```

### OpenTelemetry Metrics

Envoy Gateway can export metrics to an OpenTelemetry sink. Use the following command to send metrics to the 
OpenTelemetry Collector. Ensure that the OpenTelemetry components are enabled, 
as mentioned in the [Prerequisites](#prerequisites).

```yaml
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
    name: otel-sink
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: otel-sink
  namespace: envoy-gateway-system
spec:
  telemetry:
    metrics:
      sinks:
        - type: OpenTelemetry
          openTelemetry:
            host: otel-collector.monitoring.svc.cluster.local
            port: 4317
EOF
```

Verify OTel-Collector metrics:

```shell
export OTEL_POD_NAME=$(kubectl get pod -n monitoring --selector=app.kubernetes.io/name=opentelemetry-collector -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward pod/$OTEL_POD_NAME -n monitoring 19001:19001

# check metrics 
curl localhost:19001/metrics  | grep "default/backend/rule/0"
```
