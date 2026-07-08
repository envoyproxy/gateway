---
title: "Backend Utilization Load Balancing"
---

BackendUtilization load balancing uses [Open Resource Cost Application (ORCA)][ORCA] load metrics reported by the backend to dynamically weight endpoints. Under the hood it is implemented as [Envoy's client-side weighted round-robin][client-side-wrr] policy: each endpoint's weight is derived from the utilization metrics it emits, so instances running hot receive proportionally less traffic than those with headroom.

If no ORCA metrics are received from an endpoint, that endpoint is treated as evenly weighted.

See the [Load Balancing concepts page][concepts-lb] for a deeper explanation of ORCA metric formats.

## Prerequisites

* Your backend (or a sidecar in front of it) must emit ORCA load metrics as response headers or trailers. See [Backend instrumentation](#backend-instrumentation) below.
* {{< boilerplate prerequisites >}}

## Build and Deploy the Example Backend

The Envoy Gateway repository includes a small HTTP server under `examples/backend-utilization/` that emits a fixed ORCA `cpu_utilization` value (set via the `ORCA_CPU_UTILIZATION` environment variable) on every response. The example manifest deploys two sets of pods — one reporting `0.1` (idle) and one reporting `0.9` (hot) — behind a single Service. This lets you observe the weighting effect without wiring real load into a backend.

**Note:** The `envoyproxy/gateway-backend-utilization` image is not published to a public registry — you need to build it locally from a checkout of the Envoy Gateway repository.

* Build the example backend image

  ```shell
  make -C examples/backend-utilization docker-buildx
  ```

* Make the image available to your cluster

  {{< tabpane text=true >}}
  {{% tab header="local kind server" %}}

  ```shell
  kind load docker-image --name envoy-gateway envoyproxy/gateway-backend-utilization:latest
  ```

  {{% /tab %}}
  {{% tab header="other Kubernetes server" %}}

  ```shell
  docker tag envoyproxy/gateway-backend-utilization:latest $YOUR_DOCKER_REPO/gateway-backend-utilization:latest
  docker push $YOUR_DOCKER_REPO/gateway-backend-utilization:latest
  ```

  If you push to your own registry, update the `image:` field in `examples/kubernetes/backend-utilization.yaml` to match before applying.

  {{% /tab %}}
  {{< /tabpane >}}

* Apply the example manifest (Service, two Deployments, HTTPRoute)

  ```shell
  kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/backend-utilization.yaml -n default
  ```

Verify the two Deployments are ready:

```shell
kubectl get deployment/backend-utilization-low deployment/backend-utilization-high -n default
```

## Configure BackendUtilization

Apply a [BackendTrafficPolicy][BackendTrafficPolicy] with `loadBalancer.type: BackendUtilization`:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}
```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: backend-utilization
  namespace: default
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend-utilization
  loadBalancer:
    type: BackendUtilization
    backendUtilization:
      blackoutPeriod: 1s      # shorten so the demo shifts traffic quickly
      weightUpdatePeriod: 500ms
EOF
```
{{% /tab %}}
{{% tab header="Apply from file" %}}
```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: backend-utilization
  namespace: default
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend-utilization
  loadBalancer:
    type: BackendUtilization
    backendUtilization:
      blackoutPeriod: 1s      # shorten so the demo shifts traffic quickly
      weightUpdatePeriod: 500ms
```
{{% /tab %}}
{{< /tabpane >}}

Leaving `backendUtilization: {}` empty accepts the defaults, but the 10 s default `blackoutPeriod` means traffic will appear evenly split for the first 10 seconds of the test. The shorter values above make the weighting visible immediately. The `backendUtilization` field itself is required when `type: BackendUtilization` — omitting it will fail CEL validation.

## Configuration Fields

All fields on `backendUtilization` are optional.

| Field | Default | Purpose |
|---|---|---|
| `blackoutPeriod` | `10s` | How long an endpoint must report metrics before its reported weight is trusted. Prevents traffic from shifting based on a single noisy sample. |
| `weightExpirationPeriod` | `3m` | If an endpoint stops reporting for this long, its reported weight is discarded and it reverts to the default weight. |
| `weightUpdatePeriod` | `1s` | How often Envoy recomputes the weight table. Values below `100ms` are capped at `100ms`. |
| `errorUtilizationPenaltyPercent` | `0` | Multiplier (as `percent × 100`) applied to an endpoint's effective utilization based on its error rate (eps/qps). `100` = 1.0×, `150` = 1.5×, `200` = 2.0×. Higher values push errant endpoints out of rotation faster. |
| `metricNamesForComputingUtilization` | _unset_ | Custom ORCA metric keys to feed into the weight formula when `application_utilization` isn't reported. Use `named_metrics.<key>` for keys inside the ORCA proto's `named_metrics` map. |
| `keepResponseHeaders` | `false` | By default Envoy strips the ORCA headers/trailers before forwarding the response. Set to `true` to let downstream clients see them (useful for chained load balancers or debugging). |

### Example: Tuned for a Bursty Backend

```yaml
loadBalancer:
  type: BackendUtilization
  backendUtilization:
    blackoutPeriod: 30s              # ignore reports during slow-start
    weightExpirationPeriod: 1m       # shorter memory — react faster to silent endpoints
    weightUpdatePeriod: 500ms        # faster reweighting
    errorUtilizationPenaltyPercent: 150  # 1.5× penalty for errant endpoints
```

### Example: Application-Defined Utilization

If your backend reports a custom metric (for example, queue depth) instead of CPU utilization, wire it in through `metricNamesForComputingUtilization`:

```yaml
loadBalancer:
  type: BackendUtilization
  backendUtilization:
    metricNamesForComputingUtilization:
    - named_metrics.queue_depth
```

The backend would then emit:

```http
endpoint-load-metrics: TEXT named_metrics.queue_depth=0.42
```

## Backend Instrumentation

Your backend must emit ORCA load metrics. Envoy accepts metrics in three formats on response **headers or trailers**:

| Format | Header | Payload |
|---|---|---|
| Binary | `endpoint-load-metrics-bin` | Base64-encoded serialized [`OrcaLoadReport`][orca-proto] proto |
| JSON | `endpoint-load-metrics` | `JSON {"cpu_utilization": 0.3, "mem_utilization": 0.8}` |
| TEXT | `endpoint-load-metrics` | `TEXT cpu=0.3,mem=0.8,named_metrics.queue_depth=0.42` |

For gRPC backends, the [xDS ORCA][grpc-orca] libraries emit these automatically via the `orca_load_report` service. For HTTP backends, add a response middleware that measures and serializes your CPU/memory/custom metrics on each response.

## Combining With Zone-Aware Routing

`BackendUtilization` composes with `weightedZones` to produce locality-aware weighted round-robin (Envoy's `wrr_locality` policy). See the [WeightedZones example][zone-aware-weighted] on the zone-aware routing page.

`preferLocal` is **not** supported with `BackendUtilization`.

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the Quickstart instructions to set the variable.

Give Envoy a few seconds after applying the policy to collect ORCA samples and compute endpoint weights — until then, traffic will appear roughly even. Then send 200 requests and tally which deployment handled each. Because `backend-utilization-low` reports `cpu_utilization=0.1` and `backend-utilization-high` reports `0.9`, Envoy should weight the `low` pods roughly 9× more heavily.

```shell
for i in $(seq 1 200); do
  curl -s -H "Host: www.example.com" "http://${GATEWAY_HOST}/backend-utilization" | jq -r '.pod'
done | sort | uniq -c
```

Expected output (exact counts will vary, but `low` should dominate ~9:1):

```console
  90 backend-utilization-low-6b9cf46b59-l7df7
  87 backend-utilization-low-6b9cf46b59-xxrw2
  12 backend-utilization-high-5fdb65cb87-mctlp
  11 backend-utilization-high-5fdb65cb87-rrdvq
```

If you instead see a roughly even split, the weights may not have stabilized yet — wait a few seconds and retry. You can verify the per-endpoint weights directly through the Envoy admin interface:

```shell
ENVOY_POD=$(kubectl get pods -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
kubectl -n envoy-gateway-system port-forward pod/${ENVOY_POD} 19000:19000 &
curl -s localhost:19000/clusters | grep "backend-utilization" | grep weight
```

You should see weights roughly `10000` for the `low` pods and `1111` for the `high` pods (the inverse of the reported utilization).

## Clean-Up

```shell
kubectl delete backendtrafficpolicy/backend-utilization
kubectl delete -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/backend-utilization.yaml -n default
```

[ORCA]: https://docs.google.com/document/d/1NSnK3346BkBo1JUU3I9I5NYYnaJZQPt8_Z_XCBCI3uA
[orca-proto]: https://www.envoyproxy.io/docs/envoy/latest/xds/data/orca/v3/orca_load_report.proto
[client-side-wrr]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/load_balancing_policies/client_side_weighted_round_robin/v3/client_side_weighted_round_robin.proto
[grpc-orca]: https://github.com/grpc/proposal/blob/master/A51-custom-backend-metrics.md
[concepts-lb]: ../../../concepts/load-balancing#backend-utilization-orca
[zone-aware-weighted]: ../zone-aware-routing#weightedzones
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
