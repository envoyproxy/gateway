---
title: "Visualising metrics using Grafana"
---

Envoy Gateway provides support for exposing Envoy Gateway and Envoy Proxy metrics to a Prometheus instance.
This task shows you how to visualise the metrics exposed to Prometheus using Grafana.

## Prerequisites

Follow the steps from the [Quickstart](../../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

Follow the steps from the [Gateway Observability](../gateway-observability) and [Proxy Observability](../proxy-observability#metrics) to enable Prometheus metrics
for both Envoy Gateway (Control Plane) and Envoy Proxy (Data Plane).

[Prometheus](https://prometheus.io) is used to scrape metrics from the Envoy Gateway and Envoy Proxy instances. Install Prometheus:

```shell
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade --install prometheus prometheus-community/prometheus -n monitoring --create-namespace
```

[Grafana](https://grafana.com/grafana/) is used to visualise the metrics exposed by the Envoy Gateway and Envoy Proxy instances.
Install Grafana:

```shell
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm upgrade --install grafana grafana/grafana -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/grafana/helm-values.yaml -n monitoring --create-namespace
```

Expose endpoints:

```shell
GRAFANA_IP=$(kubectl get svc grafana -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Connecting Grafana with Prometheus datasource

To visualise metrics from Prometheus, we have to connect Grafana with Prometheus. If you installed Grafana follow the command
from prerequisites sections, the Prometheus datasource should be already configured.

You can also add the datasource manually by following the instructions from [Grafana Docs](https://grafana.com/docs/grafana/latest/datasources/prometheus/configure-prometheus-data-source/).

## Accessing Grafana

You can access the Grafana instance by visiting `http://{GRAFANA_IP}`, derived in prerequisites.

To log in to Grafana, use the credentials `admin:admin`.

Envoy Gateway has examples of dashboard for you to get started, you can check them out under `Dashboards/envoy-gateway`.

If you'd like import Grafana dashboards on your own, please refer to Grafana docs for [importing dashboards](https://grafana.com/docs/grafana/latest/dashboards/manage-dashboards/#import-a-dashboard).

### [Envoy Global](https://raw.githubusercontent.com/envoyproxy/gateway/main/charts/gateway-addons-helm/dashboards/envoy-global.json)

This dashboard example shows the overall downstram and upstream stats for each Envoy Proxy instance.

![Envoy Global](/img/envoy-global-dashboard.png)

### [Envoy Clusters](https://raw.githubusercontent.com/envoyproxy/gateway/main/charts/gateway-addons-helm/dashboards/envoy-clusters.json)

This dashboard example shows the overall stats for each cluster from Envoy Proxy fleet.

![Envoy Clusters](/img/envoy-clusters-dashboard.png)

### [Envoy Pod Resources](https://raw.githubusercontent.com/envoyproxy/gateway/main/charts/gateway-addons-helm/dashboards/envoy-pod-resource.json)

This dashboard example shows the overall pod resources stats for each Envoy Proxy instance.

![Envoy Pod Resources](/img/envoy-pod-resources-dashboard.png)

### [Envoy Gateway Global](https://raw.githubusercontent.com/envoyproxy/gateway/main/charts/gateway-addons-helm/dashboards/envoy-gateway-global.json)

This dashboard example shows the overall stats exported by Envoy Gateway fleet.

![Envoy Gateway Global: Watching Components](/img/envoy-gateway-global-watching-components.png)

![Envoy Gateway Global: Status Updater](/img/envoy-gateway-global-status-updater.png)

![Envoy Gateway Global: xDS Server](/img/envoy-gateway-global-xds-server.png)

![Envoy Gateway Global: Infrastructure Manager](/img/envoy-gateway-global-infra-manager.png)

## Update Dashboards

The example dashboards cannot be updated in-place by default, if you are trying to
make some changes to current dashboards, you can save them directly as a JSON file.

All dashboards of Envoy Gateway are maintained under `charts/gateway-addons-helm/dashboards`, 
feel free to make [contributions](../../../contributions/CONTRIBUTING).
