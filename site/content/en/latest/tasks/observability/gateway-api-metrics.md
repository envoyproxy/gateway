---
title: "Gateway API Metrics"
---

Resource metrics for **Kubernetes Gateway API** objects are available through the [Gateway API State Metrics][gasm] project. The project also includes an example dashboard for visualizing the metrics with Grafana, along with sample alerts using Prometheus and Alertmanager.

## Prerequisites

### Install Envoy Gateway

{{< boilerplate prerequisites >}}

### Install Add-ons

Envoy Gateway provides an add-ons Helm chart to simplify the installation of observability components.  
The documentation for the add-ons chart can be found
[here](https://gateway.envoyproxy.io/docs/install/gateway-addons-helm-api/).

Follow the instructions below to install the add-ons Helm chart.

```shell
helm install eg-addons oci://docker.io/envoyproxy/gateway-addons-helm --version {{< helm-version >}} --set prometheus.kube-state-metrics.enabled=true -n monitoring --create-namespace
```

### Install CRDs

```shell
kubectl apply --server-side -f https://raw.githubusercontent.com/Kuadrant/gateway-api-state-metrics/main/config/examples/kube-prometheus/bundle_crd.yaml
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


Refer to the [Gateway API State Metrics README][gasm-readme] for the complete list of available Gateway API metrics.

[//]: # (TDOD: Alerts)
[//]: # (## Alerts)

[//]: # (To view the alerts, navigate to the **Alerts** tab at `http://localhost:9090/alerts`. Gateway API-specific alerts will be grouped under the `gateway-api.rules` heading.  )

[//]: # (Alternatively, you can use the following command to view the alerts via the Prometheus API:)

[//]: # ()
[//]: # (```shell)

[//]: # (curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select&#40;.labels.rule_group and &#40;.labels.rule_group | test&#40;"gateway-api.rules"&#41;&#41;&#41;')

[//]: # (```)

[//]: # ()
[//]: # (***Note:*** Alerts are defined in a _PrometheusRules_ custom resource within the **monitoring** namespace. You can modify the alert rules by updating this resource.)

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

A set of Grafana dashboards is provided by [Gateway API State Metrics][gasm]. These dashboards are available
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

[gasm]: https://github.com/Kuadrant/gateway-api-state-metrics
[gasm-readme]: https://github.com/Kuadrant/gateway-api-state-metrics/tree/main#metrics
[gasm-dashboards]: https://github.com/Kuadrant/gateway-api-state-metrics/tree/main#dashboards
