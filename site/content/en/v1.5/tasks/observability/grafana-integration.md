---
title: "Visualising metrics using Grafana"
---

Envoy Gateway provides support for exposing Envoy Gateway and Envoy Proxy metrics to a Prometheus instance.
This task shows you how to visualise the metrics exposed to Prometheus using Grafana.

## Prerequisites

{{< boilerplate o11y_prerequisites >}}

Follow the steps from the [Gateway Observability](./gateway-observability) and [Proxy Metrics](./proxy-metric) to enable Prometheus metrics
for both Envoy Gateway (Control Plane) and Envoy Proxy (Data Plane).

Expose endpoints:

```shell
GRAFANA_IP=$(kubectl get svc grafana -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Connecting Grafana with Prometheus datasource

To visualise metrics from Prometheus, we have to connect Grafana with Prometheus. If you installed Grafana follow the command
from prerequisites sections, the Prometheus datasource should be already configured.

You can also add the datasource manually by following the instructions from [Grafana Docs](https://grafana.com/docs/grafana/latest/datasources/prometheus/configure/).

## Accessing Grafana

You can access the Grafana instance by visiting `http://{GRAFANA_IP}`, derived in prerequisites.

To log in to Grafana, use the credentials `admin:admin`.

Envoy Gateway has examples of dashboard for you to get started, you can check them out under `Dashboards/envoy-gateway`.

If you'd like import Grafana dashboards on your own, please refer to Grafana docs for [importing dashboards](https://grafana.com/docs/grafana/latest/dashboards/manage-dashboards/#import-a-dashboard).

### Envoy Proxy Global

This dashboard example shows the overall downstream and upstream stats for each Envoy Proxy instance.

![Envoy Proxy Global](/img/envoy-proxy-global-dashboard.png)

### Envoy Clusters

This dashboard example shows the overall stats for each cluster from Envoy Proxy fleet.

![Envoy Clusters](/img/envoy-clusters-dashboard.png)

### Envoy Gateway Global

This dashboard example shows the overall stats exported by Envoy Gateway fleet.

![Envoy Gateway Global: Watching Components](/img/envoy-gateway-global-watching-components.png)

![Envoy Gateway Global: Status Updater](/img/envoy-gateway-global-status-updater.png)

![Envoy Gateway Global: xDS Server](/img/envoy-gateway-global-xds-server.png)

![Envoy Gateway Global: Infrastructure Manager](/img/envoy-gateway-global-infra-manager.png)

### Resources Monitor

This dashboard example shows the overall resources stats for both Envoy Gateway and Envoy Proxy fleet.

![Envoy Gateway Resources](/img/resources-monitor-dashboard.png)

## Update Dashboards

All dashboards of Envoy Gateway are maintained under `charts/gateway-addons-helm/dashboards`,
feel free to make [contributions](../../../contributions/CONTRIBUTING).

### Grafonnet

Newer dashboards are generated with [Jsonnet](https://jsonnet.org/) with the [Grafonnet](https://grafana.github.io/grafonnet/index.html).
This is the preferred method for any new dashboards.

You can run `make helm-generate.gateway-addons-helm` to generate new version of dashboards.
All the generated dashboards have a `.gen.json` suffix.

### Legacy Dashboards

Many of our older dashboards are manually created in the UI and exported as JSON and checked in.

These example dashboards cannot be updated in-place by default, if you are trying to
make some changes to the older dashboards, you can save them directly as a JSON file
and then re-import.
