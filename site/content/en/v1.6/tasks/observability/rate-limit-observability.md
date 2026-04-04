---
title: "RateLimit Observability"
---

Envoy Gateway provides observability for the RateLimit instances.
This guide show you how to config RateLimit observability, includes traces.

## Prerequisites

{{< boilerplate o11y_prerequisites >}}

Follow the steps from the [Global Rate Limit](../traffic/global-rate-limit) to install RateLimit.

## Metrics

Envoy Gateway’s RateLimit service exposes Prometheus metrics that help you answer:

- Are requests being rate-limited?
- Which *domain*/*descriptor keys* are hottest?
- Are you getting close to limits (before you start denying traffic)?

### Retrieve Prometheus metrics

1) Find the RateLimit Pod (names can vary by installation):

```bash
kubectl get pods -n envoy-gateway-system | grep -i ratelimit
```

2) Port-forward the metrics port (default is **19001**):

```bash
export RATELIMIT_POD="<paste a Pod name from the previous command>"
kubectl -n envoy-gateway-system port-forward pod/${RATELIMIT_POD} 19001:19001
```

3) Verify the metrics exist and check the exact metric names:

```bash
curl -s localhost:19001/metrics | grep -E '^ratelimit_service_rate_limit_' | head
```

> Tip: keep `/metrics` open while you write PromQL. It’s the fastest way to confirm whether a metric is exported as a **counter** (single series) or a **histogram family** (`_bucket`, `_sum`, `_count`).

### Metric names and meanings

These metrics are exported per RateLimit *domain* and (by default) up to two descriptor-key labels:

- `ratelimit_service_rate_limit_total_hits`  
  Total number of requests evaluated by RateLimit (allowed + denied).

- `ratelimit_service_rate_limit_over_limit`  
  Number of requests that were **denied** (rate-limited).

- `ratelimit_service_rate_limit_near_limit`  
  Number of requests that were **allowed**, but close enough to the configured limit to be considered “near limit”.
  (Useful as an early warning signal.)

- `ratelimit_service_rate_limit_within_limit`  
  Number of requests that were **allowed** and not considered “near limit”.
  (Good for dashboards tracking allowed request QPS.)

- `ratelimit_service_rate_limit_shadow_mode`  
  Number of requests evaluated under **shadow mode** (dry-run / not enforced), when shadow mode is enabled in the rate limit service config.

### Labels

Typical labels you’ll see:

- `domain`: the RateLimit domain (a namespace for rate limit configs)
- `key1`: first descriptor key (when present)
- `key2`: second descriptor key (when present)

> Note: The exported label depth is controlled by the StatsD → Prometheus mapping configuration. By default you should expect **only `domain`, `key1`, and `key2`** to be labeled. Deeper descriptor segments typically won’t appear as additional labels unless the mapping is extended.
>
> **Warning: High Cardinality Risk**
>
> Descriptor keys can include unbounded values (e.g., client IPs, user IDs, unique headers). Grouping or filtering on those labels can create a time series per unique value and cause cardinality explosion in Prometheus. Avoid grouping by high-cardinality labels unless the value set is known and bounded.

---

## PromQL examples

### Histogram vs counter metric names

By default, Envoy Gateway’s StatsD → Prometheus mapping emits **histogram-family metrics**. If you customize the mapping, you may instead see **counter-style metrics**.

- **Histogram-family metrics** (multiple series, default):
  - `ratelimit_service_rate_limit_total_hits_bucket`
  - `ratelimit_service_rate_limit_total_hits_sum`
  - `ratelimit_service_rate_limit_total_hits_count`
- **Counter-style metrics** (single series, custom mapping):
  - `ratelimit_service_rate_limit_total_hits`

If you see `*_count` in `/metrics`, treat that as the primary “counter-like” series and use it with `rate()`. Use the counter-style examples only if the histogram family is not present.

### RateLimit request rate (QPS)

**Histogram-style (preferred when `_count` exists):**
```promql
sum(rate(ratelimit_service_rate_limit_total_hits_count[5m])) by (domain, key1, key2)
```

**Counter-style (fallback):**
```promql
sum(rate(ratelimit_service_rate_limit_total_hits[5m])) by (domain, key1, key2)
```

### Denied request rate (QPS, over-limit)

**Histogram-style:**
```promql
sum(rate(ratelimit_service_rate_limit_over_limit_count[5m])) by (domain, key1, key2)
```

**Counter-style:**
```promql
sum(rate(ratelimit_service_rate_limit_over_limit[5m])) by (domain, key1, key2)
```

### Allowed request rate (QPS, within-limit)

**Histogram-style:**
```promql
sum(rate(ratelimit_service_rate_limit_within_limit_count[5m])) by (domain, key1, key2)
```

**Counter-style:**
```promql
sum(rate(ratelimit_service_rate_limit_within_limit[5m])) by (domain, key1, key2)
```

### Near-limit request rate (QPS, early warning)

**Histogram-style:**
```promql
sum(rate(ratelimit_service_rate_limit_near_limit_count[5m])) by (domain, key1, key2)
```

**Counter-style:**
```promql
sum(rate(ratelimit_service_rate_limit_near_limit[5m])) by (domain, key1, key2)
```

### Over-limit ratio (denied/total)

This shows the fraction of requests denied by RateLimit.

**Histogram-style:**
```promql
sum(rate(ratelimit_service_rate_limit_over_limit_count[5m])) by (domain, key1, key2)
/
sum(rate(ratelimit_service_rate_limit_total_hits_count[5m])) by (domain, key1, key2)
```

**Counter-style:**
```promql
sum(rate(ratelimit_service_rate_limit_over_limit[5m])) by (domain, key1, key2)
/
sum(rate(ratelimit_service_rate_limit_total_hits[5m])) by (domain, key1, key2)
```

## Quick self-check while reviewing

After you apply RateLimit and send traffic, you should be able to see series for at least `total_hits` (and usually `within_limit`/`over_limit`) in `/metrics`.

If your PromQL query returns nothing:
- Confirm the metric name in `/metrics` (do you have `_count`?).
- Confirm label names in `/metrics` (do you have `key1`/`key2`?).
- Confirm traffic is actually hitting RateLimit (look for `total_hits` moving).

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

## Next steps

- See the full list of supported configuration fields in the
  [BackendTrafficPolicy API reference](../../../api/extension_types#backendtrafficpolicy)
