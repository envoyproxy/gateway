---
title: "Visualising metrics using Grafana"
---

Envoy Gateway provides support for exposing Envoy Proxy metrics to a Prometheus instance.
This guide shows you how to visualise the metrics exposed to prometheus using grafana.

## Prerequisites

Follow the steps from the [Quickstart Guide](../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

Follow the steps from the [Proxy Observability](../proxy-observability#Metrics) to enable prometheus metrics.

[Prometheus](https://prometheus.io) is used to scrape metrics from the Envoy Proxy instances. Install Prometheus:

```shell
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade --install prometheus prometheus-community/prometheus -n monitoring --create-namespace
```

[Grafana](https://grafana.com/grafana/) is used to visualise the metrics exposed by the envoy proxy instances.
Install Grafana:

```shell
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm upgrade --install grafana grafana/grafana -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.6.0/examples/grafana/helm-values.yaml -n monitoring --create-namespace
```

Expose endpoints:

```shell
GRAFANA_IP=$(kubectl get svc grafana -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Connecting Grafana with Prometheus datasource

To visualise metrics from Prometheus, we have to connect Grafana with Prometheus. If you installed Grafana from the command
from prerequisites sections, the prometheus datasource should be already configured.

You can also add the data source manually by following the instructions from [Grafana Docs](https://grafana.com/docs/grafana/v0.6.0/datasources/prometheus/configure-prometheus-data-source/).

## Accessing Grafana

You can access the Grafana instance by visiting `http://{GRAFANA_IP}`, derived in prerequisites.

To log in to Grafana, use the credentials `admin:admin`.

Envoy Gateway has examples of dashboard for you to get started:

### [Envoy Global](https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/grafana/dashboards/envoy-global.json)

![Envoy Global](/img/envoy-global-dashboard.png)

### [Envoy Clusters](https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/grafana/dashboards/envoy-clusters.json)

![Envoy Clusters](/img/envoy-clusters-dashboard.png)

### [Envoy Pod Resources](https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/grafana/dashboards/envoy-pod-resource.json)

![Envoy Pod Resources](/img/envoy-pod-resources-dashboard.png)

You can load the above dashboards in your Grafana to get started. Please refer to Grafana docs for [importing dashboards](https://grafana.com/docs/grafana/v0.6.0/dashboards/manage-dashboards/#import-a-dashboard).
