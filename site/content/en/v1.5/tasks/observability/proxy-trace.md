---
title: "Proxy Tracing"
---

Envoy Gateway provides observability for the ControlPlane and the underlying EnvoyProxy instances.
This task show you how to config proxy tracing.

## Prerequisites

{{< boilerplate o11y_prerequisites >}}

Expose Tempo endpoints:

```shell
TEMPO_IP=$(kubectl get svc tempo -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Traces

By default, Envoy Gateway doesn't send traces to any sink.
You can enable traces by setting the `telemetry.tracing` in the [EnvoyProxy][envoy-proxy-crd] CRD.
Currently, Envoy Gateway support OpenTelemetry, [Zipkin](../../api/extension_types#zipkintracingprovider) and Datadog tracer.

### Tracing Provider

The following configurations show how to apply proxy with different providers:

{{< tabpane text=true >}}
{{% tab header="OpenTelemetry" %}}

```shell
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: otel
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: otel
  namespace: envoy-gateway-system
spec:
  telemetry:
    tracing:
      # sample 100% of requests
      samplingRate: 100
      provider:
        backendRefs:
        - name: otel-collector
          namespace: monitoring
          port: 4317
        type: OpenTelemetry
      customTags:
        # This is an example of using a literal as a tag value
        provider:
          type: Literal
          literal:
            value: "otel"
        "k8s.pod.name":
          type: Environment
          environment:
            name: ENVOY_POD_NAME
            defaultValue: "-"
        "k8s.namespace.name":
          type: Environment
          environment:
            name: ENVOY_POD_NAMESPACE
            defaultValue: "envoy-gateway-system"
        # This is an example of using a header value as a tag value
        header1:
          type: RequestHeader
          requestHeader:
            name: X-Header-1
            defaultValue: "-"
EOF
```

Verify OpenTelemetry traces from tempo:

```shell
curl -s "http://$TEMPO_IP:3100/api/search?tags=component%3Dproxy+provider%3Dotel" | jq .traces
```

{{% /tab %}}
{{% tab header="Zipkin" %}}

```shell
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: zipkin
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: zipkin
  namespace: envoy-gateway-system
spec:
  telemetry:
    tracing:
      # sample 100% of requests
      samplingRate: 100
      provider:
        backendRefs:
        - name: otel-collector
          namespace: monitoring
          port: 9411
        type: Zipkin
        zipkin:
          enable128BitTraceId: true
      customTags:
        # This is an example of using a literal as a tag value
        provider:
          type: Literal
          literal:
            value: "zipkin"
        "k8s.pod.name":
          type: Environment
          environment:
            name: ENVOY_POD_NAME
            defaultValue: "-"
        "k8s.namespace.name":
          type: Environment
          environment:
            name: ENVOY_POD_NAMESPACE
            defaultValue: "envoy-gateway-system"
        # This is an example of using a header value as a tag value
        header1:
          type: RequestHeader
          requestHeader:
            name: X-Header-1
            defaultValue: "-"
EOF
```

Verify zipkin traces from tempo:

```shell
curl -s "http://$TEMPO_IP:3100/api/search?tags=component%3Dproxy+provider%3Dzipkin" | jq .traces
```

{{% /tab %}}
{{% tab header="Datadog" %}}

```shell
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: datadog
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: datadog
  namespace: envoy-gateway-system
spec:
  telemetry:
    tracing:
      # sample 100% of requests
      samplingRate: 100
      provider:
        backendRefs:
        - name: datadog-agent
          namespace: monitoring
          port: 8126
        type: Datadog
      customTags:
        # This is an example of using a literal as a tag value
        provider:
          type: Literal
          literal:
            value: "datadog"
        "k8s.pod.name":
          type: Environment
          environment:
            name: ENVOY_POD_NAME
            defaultValue: "-"
        "k8s.namespace.name":
          type: Environment
          environment:
            name: ENVOY_POD_NAMESPACE
            defaultValue: "envoy-gateway-system"
        # This is an example of using a header value as a tag value
        header1:
          type: RequestHeader
          requestHeader:
            name: X-Header-1
            defaultValue: "-"
EOF
```

Verify Datadog traces in [Datadog APM](https://docs.datadoghq.com/tracing/)

{{% /tab %}}
{{< /tabpane >}}

Query trace by trace id:

```shell
curl -s "http://$TEMPO_IP:3100/api/traces/<trace_id>" | jq
```


### Sampling Rate

Envoy Gateway use 100% sample rate, which means all requests will be traced.
This may cause performance issues when traffic is very high, you can adjust
the sample rate by setting the `telemetry.tracing.samplingRate` in the [EnvoyProxy][envoy-proxy-crd] CRD.

The following configurations show how to apply proxy with 1% sample rates:

```shell
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: otel
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: otel
  namespace: envoy-gateway-system
spec:
  telemetry:
    tracing:
      # sample 1% of requests
      samplingRate: 1
      provider:
        backendRefs:
        - name: otel-collector
          namespace: monitoring
          port: 4317
        type: OpenTelemetry
      customTags:
        # This is an example of using a literal as a tag value
        provider:
          type: Literal
          literal:
            value: "otel"
        "k8s.pod.name":
          type: Environment
          environment:
            name: ENVOY_POD_NAME
            defaultValue: "-"
        "k8s.namespace.name":
          type: Environment
          environment:
            name: ENVOY_POD_NAMESPACE
            defaultValue: "envoy-gateway-system"
        # This is an example of using a header value as a tag value
        header1:
          type: RequestHeader
          requestHeader:
            name: X-Header-1
            defaultValue: "-"
EOF
```

If you want the sample rate to less than 1%, you can use the `telemetry.tracing.samplingFraction` field in the [EnvoyProxy][envoy-proxy-crd] CRD.

```shell
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: otel
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: otel
  namespace: envoy-gateway-system
spec:
  telemetry:
    tracing:
      # sample 0.1% of requests
      samplingFraction:
        numerator: 1
        denominator: 1000
      provider:
        backendRefs:
        - name: otel-collector
          namespace: monitoring
          port: 4317
        type: OpenTelemetry
      customTags:
        # This is an example of using a literal as a tag value
        provider:
          type: Literal
          literal:
            value: "otel"
        "k8s.pod.name":
          type: Environment
          environment:
            name: ENVOY_POD_NAME
            defaultValue: "-"
        "k8s.namespace.name":
          type: Environment
          environment:
            name: ENVOY_POD_NAMESPACE
            defaultValue: "envoy-gateway-system"
        # This is an example of using a header value as a tag value
        header1:
          type: RequestHeader
          requestHeader:
            name: X-Header-1
            defaultValue: "-"
EOF
```


[envoy-proxy-crd]: ../../api/extension_types#envoyproxy
