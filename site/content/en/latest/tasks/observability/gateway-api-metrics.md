---
title: "Gateway API Metrics"
---

Resource metrics for **Kubernetes Gateway API** objects are available through the [Gateway API State Metrics][gasm] project. The project also includes an example dashboard for visualizing the metrics with Grafana, along with sample alerts using Prometheus and Alertmanager.

## Prerequisites

### Install Envoy Gateway

{{< boilerplate prerequisites >}}

### Install Metrics Stack

Run the following commands to install the metrics stack, with the _Gateway API State Metrics_ configuration, 
on your kubernetes cluster:

```shell
kubectl apply --server-side -f https://raw.githubusercontent.com/Kuadrant/gateway-api-state-metrics/main/config/examples/kube-prometheus/bundle_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/Kuadrant/gateway-api-state-metrics/main/config/examples/kube-prometheus/bundle.yaml
```

## Metrics

To query metrics using Prometheus API, follow the steps below. Make sure to wait for the statefulset to be ready before port-forwarding.

```shell
export PROMETHEUS_PORT=$(kubectl get service prometheus-k8s -n monitoring -o jsonpath='{.spec.ports[0].port}')
kubectl port-forward service/prometheus-k8s -n monitoring 9090:$PROMETHEUS_PORT
```

The example query below fetches the `gatewayapi_gateway_created` metric.
Alternatively, access the Prometheus UI at `http://localhost:9090`.

```shell
curl -s 'http://localhost:9090/api/v1/query?query=gatewayapi_gateway_created' | jq . 
```


Refer to the [Gateway API State Metrics README][gasm-readme] for the complete list of available Gateway API metrics.

## Alerts
To view the alerts, navigate to the **Alerts** tab at `http://localhost:9090/alerts`. Gateway API-specific alerts will be grouped under the `gateway-api.rules` heading.  
Alternatively, you can use the following command to view the alerts via the Prometheus API:

```shell
curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.labels.rule_group and (.labels.rule_group | test("gateway-api.rules")))'
```

***Note:*** Alerts are defined in a _PrometheusRules_ custom resource within the **monitoring** namespace. You can modify the alert rules by updating this resource.

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

The Gateway API State dashboards are located in the 'Default' folder and are tagged with `gateway-api`.
For detailed information about available dashboards, refer to the [Gateway API State Metrics README][gasm-dashboards].

**Note:**  
Dashboards are loaded from ConfigMaps. While you can modify dashboards directly in the Grafana UI, to persist these changes you must:
1. Export the modified dashboards from the UI
2. Update the corresponding JSON in the ConfigMaps

## Next Steps

Check out the [Gateway Exported Metrics](./grafana-integration.md) section to learn more about the metrics exported by the Envoy Gateway.

[gasm]: https://github.com/Kuadrant/gateway-api-state-metrics
[gasm-readme]: https://github.com/Kuadrant/gateway-api-state-metrics/tree/main#metrics
[gasm-dashboards]: https://github.com/Kuadrant/gateway-api-state-metrics/tree/main#dashboards
