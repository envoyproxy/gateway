# Visualising metrics using Grafana

Envoy Gateway provides support for exposing Envoy Proxy metrics to a Prometheus instance.
This guide shows you how to visualise the metrics exposed to prometheus using grafana.

## Prerequisites

Follow the steps from the [Quickstart Guide](quickstart.md) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

Follow the steps from the [Proxy Observability](proxy-observability.md#Metrics) to enable prometheus metrics.

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
helm upgrade --install grafana grafana/grafana -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/grafana/helm-values.yaml -n monitoring --create-namespace
```

Expose endpoints:

```shell
GRAFANA_IP=$(kubectl get svc grafana -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Configure Prometheus for scraping metrics
Following the steps in [Proxy Observability](proxy-observability.md#Metrics), we have exposed the metrics from
envoy proxies on port `19001` and path `/stats/prometheus`.

Now we need to configure Prometheus to scrape the metrics from this port and path for all the envoy proxy pods.
There are many ways to configure prometheus: by adding annotations to envoy proxy pods,
by using prometheus-operator and associated CRDs to generate scrape configs or by writing the scrape configs manually.

In this guide, we will add the required annotations on Envoy Proxy pods by updating EnvoyProxy configuration.

Update the Envoy Proxy spec by:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/metric/prometheus-annotations.yaml
```

## Accessing Grafana
If you installed Grafana in your cluster using the command from prerequisites section,
you should have prometheus configured as a data source automatically.

You can access the Grafana instance by visiting http://{GRAFANA_IP}

To log in to Grafana, use the credentials `admin:admin`.

Envoy Gateway has examples of dashboard for you to get started:
- [Envoy Global](https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/dashboards/envoy-global.json)
- [Envoy Clusters]((https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/dashboards/envoy-clusters.json))
- [Envoy Pod Resources]((https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/dashboards/envoy-pod-resource.json))

You can load the above dashboards in your Grafana to get started.
