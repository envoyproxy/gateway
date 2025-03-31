---
---

### Install Envoy Gateway

{{< boilerplate prerequisites >}}

### Install Add-ons

Envoy Gateway provides an add-ons Helm chart to simplify the installation of observability components.  
The documentation for the add-ons chart can be found
[here](https://gateway.envoyproxy.io/docs/install/gateway-addons-helm-api/).

Follow the instructions below to install the add-ons Helm chart.

```shell
helm install eg-addons oci://docker.io/envoyproxy/gateway-addons-helm --version {{< helm-version >}} -n monitoring --create-namespace
```

By default, the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) is **disabled.**
To install add-ons with OpenTelemetry Collector enabled, use the following command.

```shell
helm install eg-addons oci://docker.io/envoyproxy/gateway-addons-helm --version {{< helm-version >}} --set opentelemetry-collector.enabled=true -n monitoring --create-namespace
```