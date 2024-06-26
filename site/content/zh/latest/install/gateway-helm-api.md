---
title: "Gateway Helm Chart"
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
| certgen | object | `{"job":{"annotations":{},"resources":{},"ttlSecondsAfterFinished":30},"rbac":{"annotations":{},"labels":{}}}` | Certgen is used to generate the certificates required by EnvoyGateway. If you want to construct a custom certificate, you can generate a custom certificate through Cert-Manager before installing EnvoyGateway. Certgen will not overwrite the custom certificate. Please do not manually modify `values.yaml` to disable certgen, it may cause EnvoyGateway OIDC,OAuth2,etc. to not work as expected. |
| config.envoyGateway.gateway.controllerName | string | `"gateway.envoyproxy.io/gatewayclass-controller"` |  |
| config.envoyGateway.logging.level.default | string | `"info"` |  |
| config.envoyGateway.provider.type | string | `"Kubernetes"` |  |
| createNamespace | bool | `false` |  |
| deployment.envoyGateway.image.repository | string | `""` |  |
| deployment.envoyGateway.image.tag | string | `""` |  |
| deployment.envoyGateway.imagePullPolicy | string | `""` |  |
| deployment.envoyGateway.imagePullSecrets | list | `[]` |  |
| deployment.envoyGateway.resources.limits.cpu | string | `"500m"` |  |
| deployment.envoyGateway.resources.limits.memory | string | `"1024Mi"` |  |
| deployment.envoyGateway.resources.requests.cpu | string | `"100m"` |  |
| deployment.envoyGateway.resources.requests.memory | string | `"256Mi"` |  |
| deployment.pod.affinity | object | `{}` |  |
| deployment.pod.annotations."prometheus.io/port" | string | `"19001"` |  |
| deployment.pod.annotations."prometheus.io/scrape" | string | `"true"` |  |
| deployment.pod.labels | object | `{}` |  |
| deployment.pod.tolerations | list | `[]` |  |
| deployment.pod.topologySpreadConstraints | list | `[]` |  |
| deployment.ports[0].name | string | `"grpc"` |  |
| deployment.ports[0].port | int | `18000` |  |
| deployment.ports[0].targetPort | int | `18000` |  |
| deployment.ports[1].name | string | `"ratelimit"` |  |
| deployment.ports[1].port | int | `18001` |  |
| deployment.ports[1].targetPort | int | `18001` |  |
| deployment.ports[2].name | string | `"metrics"` |  |
| deployment.ports[2].port | int | `19001` |  |
| deployment.ports[2].targetPort | int | `19001` |  |
| deployment.replicas | int | `1` |  |
| global.images.envoyGateway.image | string | `nil` |  |
| global.images.envoyGateway.pullPolicy | string | `nil` |  |
| global.images.envoyGateway.pullSecrets | list | `[]` |  |
| global.images.ratelimit.image | string | `"docker.io/envoyproxy/ratelimit:master"` |  |
| global.images.ratelimit.pullPolicy | string | `"IfNotPresent"` |  |
| global.images.ratelimit.pullSecrets | list | `[]` |  |
| kubernetesClusterDomain | string | `"cluster.local"` |  |
| podDisruptionBudget.minAvailable | int | `0` |  |

