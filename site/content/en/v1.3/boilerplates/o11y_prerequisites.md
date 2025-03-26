---
---

{{< boilerplate prerequisites >}}

Envoy Gateway provides an add-ons Helm Chart, which includes all the needing components for observability.
By default, the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) is disabled.

Install the add-ons Helm Chart:

```shell
helm install eg-addons oci://docker.io/envoyproxy/gateway-addons-helm --version {{< helm-version >}} --set opentelemetry-collector.enabled=true -n monitoring --create-namespace
```
