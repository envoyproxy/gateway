+++
title = "Gateway Addons Helm Chart"
+++

![Version: v0.0.0-latest](https://img.shields.io/badge/Version-v0.0.0--latest-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: latest](https://img.shields.io/badge/AppVersion-latest-informational?style=flat-square)

An Add-ons Helm chart for Envoy Gateway

**Homepage:** <https://gateway.envoyproxy.io/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| envoy-gateway-steering-committee |  | <https://github.com/envoyproxy/gateway/blob/main/GOVERNANCE.md> |
| envoy-gateway-maintainers |  | <https://github.com/envoyproxy/gateway/blob/main/CODEOWNERS> |

## Source Code

* <https://github.com/envoyproxy/gateway>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://grafana.github.io/helm-charts | grafana | 6.58.4 |
| https://prometheus-community.github.io/helm-charts | prometheus | 23.1.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| grafana.adminPassword | string | `"admin"` |  |
| grafana.datasources."datasources.yaml".apiVersion | int | `1` |  |
| grafana.datasources."datasources.yaml".datasources[0].name | string | `"Prometheus"` |  |
| grafana.datasources."datasources.yaml".datasources[0].type | string | `"prometheus"` |  |
| grafana.datasources."datasources.yaml".datasources[0].url | string | `"http://prometheus-server"` |  |
| grafana.service.type | string | `"LoadBalancer"` |  |
| prometheus.alertmanager.enabled | bool | `false` |  |
| prometheus.kube-state-metrics.enabled | bool | `false` |  |
| prometheus.prometheus-node-exporter.enabled | bool | `false` |  |
| prometheus.prometheus-pushgateway.enabled | bool | `false` |  |
| prometheus.server.fullnameOverride | string | `"prometheus"` |  |
| prometheus.server.global.scrape_interval | string | `"15s"` |  |
| prometheus.server.image.repository | string | `"prom/prometheus"` |  |
| prometheus.server.persistentVolume.enabled | bool | `false` |  |
| prometheus.server.readinessProbeInitialDelay | int | `0` |  |
| prometheus.server.securityContext | string | `nil` |  |
| prometheus.server.service.type | string | `"LoadBalancer"` |  |

