---
title: "Proxy Access Logs"
---

Envoy Gateway provides observability for the ControlPlane and the underlying EnvoyProxy instances.
This task show you how to config proxy access logs.

## Prerequisites

{{< boilerplate o11y_prerequisites >}}

By default, the Service type of `loki` is ClusterIP, you can change it to LoadBalancer type for further usage:

```shell
kubectl patch service loki -n monitoring -p '{"spec": {"type": "LoadBalancer"}}'
```

Expose endpoints:

```shell
LOKI_IP=$(kubectl get svc loki -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Default Access Log

If custom format string is not specified, Envoy Gateway uses the following default format:

```json
{
  "start_time": "%START_TIME%",
  "method": "%REQ(:METHOD)%",
  "x-envoy-origin-path": "%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%",
  "protocol": "%PROTOCOL%",
  "response_code": "%RESPONSE_CODE%",
  "response_flags": "%RESPONSE_FLAGS%",
  "response_code_details": "%RESPONSE_CODE_DETAILS%",
  "connection_termination_details": "%CONNECTION_TERMINATION_DETAILS%",
  "upstream_transport_failure_reason": "%UPSTREAM_TRANSPORT_FAILURE_REASON%",
  "bytes_received": "%BYTES_RECEIVED%",
  "bytes_sent": "%BYTES_SENT%",
  "duration": "%DURATION%",
  "x-envoy-upstream-service-time": "%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)%",
  "x-forwarded-for": "%REQ(X-FORWARDED-FOR)%",
  "user-agent": "%REQ(USER-AGENT)%",
  "x-request-id": "%REQ(X-REQUEST-ID)%",
  ":authority": "%REQ(:AUTHORITY)%",
  "upstream_host": "%UPSTREAM_HOST%",
  "upstream_cluster": "%UPSTREAM_CLUSTER%",
  "upstream_local_address": "%UPSTREAM_LOCAL_ADDRESS%",
  "downstream_local_address": "%DOWNSTREAM_LOCAL_ADDRESS%",
  "downstream_remote_address": "%DOWNSTREAM_REMOTE_ADDRESS%",
  "requested_server_name": "%REQUESTED_SERVER_NAME%",
  "route_name": "%ROUTE_NAME%"
}
```

> Note: Envoy Gateway disable envoy headers by default, you can enable it by setting `EnableEnvoyHeaders` to `true` in the [ClientTrafficPolicy](../../api/extension_types#backendtrafficpolicy) CRD.


Verify logs from loki:

```shell
curl -s "http://$LOKI_IP:3100/loki/api/v1/query_range" --data-urlencode "query={job=\"fluentbit\"}" | jq '.data.result[0].values'
```

## Disable Access Log

If you want to disable it, set the `telemetry.accesslog.disable` to `true` in the `EnvoyProxy` CRD.

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
    name: disable-accesslog
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: disable-accesslog
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      disable: true
EOF
```

## OpenTelemetry Sink

Envoy Gateway can send logs to OpenTelemetry Sink.

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
    name: otel-access-logging
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: otel-access-logging
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      settings:
        - format:
            type: Text
            text: |
              [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
          sinks:
            - type: OpenTelemetry
              openTelemetry:
                host: otel-collector.monitoring.svc.cluster.local
                port: 4317
                resources:
                  k8s.cluster.name: "cluster-1"
EOF
```

Verify logs from loki:

```shell
curl -s "http://$LOKI_IP:3100/loki/api/v1/query_range" --data-urlencode "query={exporter=\"OTLP\"}" | jq '.data.result[0].values'
```

## gGRPC Access Log Service(ALS) Sink

Envoy Gateway can send logs to a backend implemented [gRPC access log service proto](https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/accesslog/v3/als.proto).
There's an example service [here](https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/envoy-als.yaml), which simply count the log and export to prometheus endpoint.

The following configuration sends logs to the gRPC access log service:

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
    name: als
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: als
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      settings:
        - sinks:
            - type: ALS
              als:
                backendRefs:
                  - name: otel-collector
                    namespace: monitoring
                    port: 9000
                type: HTTP
EOF
```

Verify logs from envoy-als:

```shell
curl -s "http://$LOKI_IP:3100/loki/api/v1/query_range" --data-urlencode "query={exporter=\"OTLP\"}" | jq '.data.result[0].values'
```

## CEL Expressions

Envoy Gateway provides [CEL expressions](https://www.envoyproxy.io/docs/envoy/latest/xds/type/v3/cel.proto.html#common-expression-language-cel-proto) to filter access log . 

For example, you can use the expression `'x-envoy-logged' in request.headers` to filter logs that contain the `x-envoy-logged` header.

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
    name: otel-access-logging
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: otel-access-logging
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      settings:
        - format:
            type: Text
            text: |
              [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%" "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
          matches:
            - "'x-envoy-logged' in request.headers"
          sinks:
            - type: OpenTelemetry
              openTelemetry:
                host: otel-collector.monitoring.svc.cluster.local
                port: 4317
                resources:
                  k8s.cluster.name: "cluster-1"
EOF
```

Verify logs from loki:

```shell
curl -s "http://$LOKI_IP:3100/loki/api/v1/query_range" --data-urlencode "query={exporter=\"OTLP\"}" | jq '.data.result[0].values'
```


## Gateway API Metadata

Envoy Gateway provides additional metadata about the K8s resources that were translated to  certain envoy resources.
For example, details about the `HTTPRoute` and `GRPCRoute` (kind, group, name, namespace and annotations) are available
for access log formatter using the `METADATA` operator. 

To enrich logs, users can add log operator that refer to XDS metadata, such as:
- `%METADATA(ROUTE:envoy-gateway:resources)%`
- `%CEL(xds.route_metadata.filter_metadata['envoy-gateway']['resources'][0]['name'])%`

## Access Log Types

By default, Access Log settings would apply to:
- All Routes
- If traffic is not matched by any Route known to Envoy, the Listener would emit the access log instead 

Users may wish to customize this behavior:
- Emit Access Logs by all Listeners for all traffic with specific settings
- Do not emit Route-oriented access logs when a route is not matched.

To achieve this, users can select if Access Log settings follow the default behavior or apply specifically to 
Routes or Listeners by specifying the setting's type. 

**Note**: When users define their own Access Log settings (with or without a type), the default Envoy Gateway 
file access log is no longer configured. It can be re-enabled explicitly by adding empty settings for the desired components. 

In the following example:
- Route Access logs would use the default Envoy Gateway format and sink 
- Listener Access logs are customized to report transport-level failures and connection attributes

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
    name: otel-access-logging
    namespace: envoy-gateway-system
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: otel-access-logging
  namespace: envoy-gateway-system
spec:
  telemetry:
    accessLog:
      settings:
        - type: Route # re-enable default access log for route
        - type: Listener # configure specific access log for listeners
          format:
            type: Text
            text: |
              [%START_TIME%] %DOWNSTREAM_REMOTE_ADDRESS% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DOWNSTREAM_TRANSPORT_FAILURE_REASON%
          sinks:
            - type: OpenTelemetry
              openTelemetry:
                host: otel-collector.monitoring.svc.cluster.local
                port: 4317
                resources:
                  k8s.cluster.name: "cluster-1"
EOF
```