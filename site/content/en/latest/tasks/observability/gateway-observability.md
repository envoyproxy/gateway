---
title: "Gateway Observability"
---

Envoy Gateway provides observability for the ControlPlane and the underlying EnvoyProxy instances.
This task show you how to config gateway control-plane observability, includes metrics.

## Prerequisites

Follow the steps from the [Quickstart](../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

[OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) offers a vendor-agnostic implementation of how to receive, process and export telemetry data.
Install OTel-Collector:

```shell
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo update
helm upgrade --install otel-collector open-telemetry/opentelemetry-collector -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/otel-collector/helm-values.yaml -n monitoring --create-namespace --version 0.60.0
```

## Metrics

The default installation of Envoy Gateway installs a default [EnvoyGateway][] configuration and attaches it
using a `ConfigMap`. In this section, we will update this resource to enable various ways to retrieve metrics
from Envoy Gateway.

### Retrieve Prometheus Metrics from Envoy Gateway

By default, prometheus metric is enabled. You can directly retrieve metrics from Envoy Gateway:

```shell
export ENVOY_POD_NAME=$(kubectl get pod -n envoy-gateway-system --selector=control-plane=envoy-gateway,app.kubernetes.io/instance=eg -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward pod/$ENVOY_POD_NAME -n envoy-gateway-system 19001:19001

# check metrics 
curl localhost:19001/metrics
```

The following is an example to disable prometheus metric for Envoy Gateway.

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
    telemetry:
      metrics:
        prometheus:
          disable: true
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
    telemetry:
      metrics:
        prometheus:
          disable: true
```

{{% /tab %}}
{{< /tabpane >}}

After updating the `ConfigMap`, you will need to restart the `envoy-gateway` deployment so the configuration kicks in:

```shell
kubectl rollout restart deployment envoy-gateway -n envoy-gateway-system
```

### Enable Open Telemetry sink in Envoy Gateway

The following is an example to send metric via Open Telemetry sink to OTEL gRPC Collector.

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
    telemetry:
      metrics:
        sinks:
          - type: OpenTelemetry
            openTelemetry:
              host: otel-collector.monitoring.svc.cluster.local
              port: 4317
              protocol: grpc
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
    telemetry:
      metrics:
        sinks:
          - type: OpenTelemetry
            openTelemetry:
              host: otel-collector.monitoring.svc.cluster.local
              port: 4317
              protocol: grpc
```

{{% /tab %}}
{{< /tabpane >}}

After updating the `ConfigMap`, you will need to restart the `envoy-gateway` deployment so the configuration kicks in:

```shell
kubectl rollout restart deployment envoy-gateway -n envoy-gateway-system
```

Verify OTel-Collector metrics:

```shell
export OTEL_POD_NAME=$(kubectl get pod -n monitoring --selector=app.kubernetes.io/name=opentelemetry-collector -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward pod/$OTEL_POD_NAME -n monitoring 19001:19001

# check metrics 
curl localhost:19001/metrics
```

[EnvoyGateway]: ../../api/extension_types#envoygateway
