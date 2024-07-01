---
---
Follow the steps from the [Quickstart](../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

Envoy Gateway provides an add-ons Helm Chart, which includes all the needing components for observability.
By default, the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) is disabled.

Install the add-ons Helm Chart:

```shell
helm install eg-addons oci://docker.io/envoyproxy/gateway-addons-helm --version v0.0.0-latest --set opentelemetry-collector.enabled=true -n monitoring --create-namespace
```
