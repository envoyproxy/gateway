---
title: "Helm Chart Values"
---

![Version: v0.0.0-latest](https://img.shields.io/badge/Version-v0.0.0--latest-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: latest](https://img.shields.io/badge/AppVersion-latest-informational?style=flat-square)

The Helm chart for Envoy Gateway

**Homepage:** <https://gateway.envoyproxy.io/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| envoy-gateway-steering-committee |  | <https://github.com/envoyproxy/gateway/blob/main/GOVERNANCE.md> |
| envoy-gateway-maintainers |  | <https://github.com/envoyproxy/gateway/blob/main/CODEOWNERS> |

## Source Code

* <https://github.com/envoyproxy/gateway>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| config.envoyGateway.gateway.controllerName | string | `"gateway.envoyproxy.io/gatewayclass-controller"` |  |
| config.envoyGateway.logging.level.default | string | `"info"` |  |
| config.envoyGateway.provider.type | string | `"Kubernetes"` |  |
| createNamespace | bool | `false` |  |
| deployment.envoyGateway.image.repository | string | `"${ImageRepository}"` |  |
| deployment.envoyGateway.image.tag | string | `"${ImageTag}"` |  |
| deployment.envoyGateway.imagePullPolicy | string | `"Always"` |  |
| deployment.envoyGateway.resources.limits.cpu | string | `"500m"` |  |
| deployment.envoyGateway.resources.limits.memory | string | `"1024Mi"` |  |
| deployment.envoyGateway.resources.requests.cpu | string | `"100m"` |  |
| deployment.envoyGateway.resources.requests.memory | string | `"256Mi"` |  |
| deployment.kubeRbacProxy.image.repository | string | `"gcr.io/kubebuilder/kube-rbac-proxy"` |  |
| deployment.kubeRbacProxy.image.tag | string | `"v0.11.0"` |  |
| deployment.kubeRbacProxy.resources.limits.cpu | string | `"500m"` |  |
| deployment.kubeRbacProxy.resources.limits.memory | string | `"128Mi"` |  |
| deployment.kubeRbacProxy.resources.requests.cpu | string | `"5m"` |  |
| deployment.kubeRbacProxy.resources.requests.memory | string | `"64Mi"` |  |
| deployment.pod.annotations | object | `{}` |  |
| deployment.pod.labels | object | `{}` |  |
| deployment.ports[0].name | string | `"grpc"` |  |
| deployment.ports[0].port | int | `18000` |  |
| deployment.ports[0].targetPort | int | `18000` |  |
| deployment.ports[1].name | string | `"ratelimit"` |  |
| deployment.ports[1].port | int | `18001` |  |
| deployment.ports[1].targetPort | int | `18001` |  |
| deployment.replicas | int | `1` |  |
| envoyGatewayMetricsService.ports[0].name | string | `"https"` |  |
| envoyGatewayMetricsService.ports[0].port | int | `8443` |  |
| envoyGatewayMetricsService.ports[0].protocol | string | `"TCP"` |  |
| envoyGatewayMetricsService.ports[0].targetPort | string | `"https"` |  |
| kubernetesClusterDomain | string | `"cluster.local"` |  |

