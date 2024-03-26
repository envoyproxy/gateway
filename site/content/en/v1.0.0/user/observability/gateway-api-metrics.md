---
title: "Gateway API Metrics"
---

Resource metrics for Gateway API objects are available using the [Gateway API State Metrics][gasm] project.
The project also provides example dashboard for visualising the metrics using Grafana, and example alerts using Prometheus & Alertmanager.

## Prerequisites

Follow the steps from the [Quickstart Guide](../../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

Run the following commands to install the metrics stack, with the Gateway API State Metrics configuration, on your kubernetes cluster:

```shell
kubectl apply --server-side -f https://raw.githubusercontent.com/Kuadrant/gateway-api-state-metrics/main/config/examples/kube-prometheus/bundle_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/Kuadrant/gateway-api-state-metrics/main/config/examples/kube-prometheus/bundle.yaml
```

## Metrics and Alerts

To access the Prometheus UI, wait for the statefulset to be ready, then use the port-forward command:

```shell
# This first command may fail if the statefulset has not been created yet.
# In that case, try again until you get a message like 'Waiting for 2 pods to be ready...'
# or 'statefulset rolling update complete 2 pods...'
kubectl -n monitoring rollout status --watch --timeout=5m statefulset/prometheus-k8s
kubectl -n monitoring port-forward service/prometheus-k8s 9090:9090 > /dev/null &
```

Navigate to `http://localhost:9090`.
Metrics can be queried from the 'Graph' tab e.g. `gatewayapi_gateway_created`
See the [Gateway API State Metrics README][gasm-readme] for the full list of Gateway API metrics available.

Alerts can be seen in the 'Alerts' tab.
Gateway API specific alerts will be grouped under the 'gateway-api.rules' heading.

***Note:*** Alerts are defined in a PrometheusRules custom resource in the 'monitoring' namespace. You can modify the alert rules by updating this resource.

## Dashboards

To view the dashboards in Grafana, wait for the deployment to be ready, then use the port-forward command:

```shell
kubectl -n monitoring wait --timeout=5m deployment/grafana --for=condition=Available
kubectl -n monitoring port-forward service/grafana 3000:3000 > /dev/null &
```

Navigate to `http://localhost:3000` and sign in with admin/admin.
The Gateway API State dashboards will be available in the 'Default' folder and tagged with 'gateway-api'.
See the [Gateway API State Metrics README][gasm-dashboards] for further information on available dashboards.

***Note:*** Dashboards are loaded from configmaps. You can modify the dashboards in the Grafana UI, however you will need to export them from the UI and update the json in the configmaps to persist changes.


[gasm]: https://github.com/Kuadrant/gateway-api-state-metrics
[gasm-readme]: https://github.com/Kuadrant/gateway-api-state-metrics/tree/main#metrics
[gasm-dashboards]: https://github.com/Kuadrant/gateway-api-state-metrics/tree/main#dashboards
