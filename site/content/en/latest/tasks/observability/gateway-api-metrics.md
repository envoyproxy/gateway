---
title: "Gateway API Metrics"
---

Resource metrics for **Kubernetes Gateway API** objects are available through the [Kube State Metrics](https://github.com/kubernetes/kube-state-metrics) project.

The example Grafana dashboards, along with the sample alerts configured in this task, can be found in the [gateway-api-state-metrics](https://github.com/Kuadrant/gateway-api-state-metrics) repository.

## Prerequisites

{{< boilerplate o11y_prerequisites >}}

### Enable kube-state-metrics

The `kube-state-metrics` service is required to collect metrics from the Kubernetes API server. Use the following command to enable it:

```shell
helm upgrade eg-addons oci://docker.io/envoyproxy/gateway-addons-helm \
--version {{< helm-version >}} \
--reuse-values \
--set prometheus.kube-state-metrics.enabled=true \
-n monitoring
```

## Metrics

To query metrics using Prometheus API, follow the steps below. Make sure to wait for the statefulset to be ready before port-forwarding.

```shell
export PROMETHEUS_PORT=$(kubectl get service prometheus -n monitoring -o jsonpath='{.spec.ports[0].port}')
kubectl port-forward service/prometheus -n monitoring 9090:$PROMETHEUS_PORT
```

The example query below fetches the `gatewayapi_gateway_created` metric.
Alternatively, access the Prometheus UI at `http://localhost:9090`.

```shell
curl -s 'http://localhost:9090/api/v1/query?query=gatewayapi_gateway_created' | jq . 
```


## Alerts

A set of example alert rules are available in
[config/examples/rules](https://github.com/Kuadrant/gateway-api-state-metrics/tree/main/config/examples/rules). To create alert use the following command:

```shell
cat <<EOF | helm upgrade eg-addons oci://docker.io/envoyproxy/gateway-addons-helm \
  --version v0.0.0-latest \
  -n monitoring --reuse-values -f -
prometheus:
  serverFiles:
    alerting_rules.yml:
      groups:
        - name: gateway-api.rules
          rules:
            - alert: UnhealthyGateway
              expr: (gatewayapi_gateway_status{type="Accepted"} == 0) or (gatewayapi_gateway_status{type="Programmed"} == 0)
              for: 10m
              labels:
                severity: critical
              annotations:
                summary: "Either the Accepted or Programmed status is not True"
                description: "Gateway {{ \$labels.namespace }}/{{ \$labels.name }} has an unhealthy status"
            - alert: InsecureHTTPListener
              expr: gatewayapi_gateway_listener_info{protocol="HTTP"}
              for: 10m
              labels:
                severity: critical
              annotations:
                summary: "Listeners must use HTTPS"
                description: "Gateway {{ \$labels.namespace }}/{{ \$labels.name }} has an insecure listener {{ \$labels.protocol }}/{{ \$labels.port }}"
EOF
```

To view the alerts, navigate to the **Alerts** tab at `http://localhost:9090/alerts`.

Alternatively, you can use the following command to view the alerts via the Prometheus API:


```shell
curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[]'

```

## Dashboards

To access the Grafana dashboards, follow these steps:

1. Wait for the deployment to complete, then set up port forwarding using the following commands:

    ```shell
    export GRAFANA_PORT=$(kubectl get service grafana -n monitoring -o jsonpath='{.spec.ports[0].port}')
    kubectl port-forward service/grafana -n monitoring 3000:$GRAFANA_PORT
    ```

2. Access Grafana by navigating to `http://localhost:3000` in your web browser
3. Log in using the default credentials:
   - Username: `admin`
   - Password: `admin`

A set of Grafana dashboards is provided by [Gateway API State Metrics](https://github.com/Kuadrant/gateway-api-state-metrics/tree/main/src/dashboards). These dashboards are available
in [./config/examples/dashboards](https://github.com/Kuadrant/gateway-api-state-metrics/tree/main/config/examples/dashboards)
and on [grafana.com](https://grafana.com/grafana/dashboards/?search=Gateway+API+State).
To import them manually navigate to the Grafana UI and select **Dashboards** > **New** > **Import**.

Alternatively, use the following command to import dashboards using the Grafana API:


```shell
export GRAFANA_API_KEY="your-api-key"

urls=(
  "https://grafana.com/api/dashboards/19433/revisions/1/download"
  "https://grafana.com/api/dashboards/19432/revisions/1/download"
  "https://grafana.com/api/dashboards/19434/revisions/1/download"
  "https://grafana.com/api/dashboards/19570/revisions/1/download"
)

for url in "${urls[@]}"; do
  dashboard_data=$(curl -s "$url")
  curl -X POST \
    -H "Authorization: Bearer $GRAFANA_API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"dashboard\": $dashboard_data, \"overwrite\": true}" \
    "http://localhost:3000/api/dashboards/db"
done
```

## Next Steps

Check out the [Gateway Exported Metrics](./grafana-integration.md) section to learn more about the metrics exported by the Envoy Gateway.
