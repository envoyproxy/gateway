+++
title = "Gateway Helm Chart"
+++

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

## Requirements

| Repository | Name | Version |
|------------|------|---------|
|  | crds | 0.0.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| certgen | object | `{"job":{"affinity":{},"annotations":{},"args":[],"nodeSelector":{},"pod":{"annotations":{},"labels":{},"securityContext":{"fsGroup":65532,"runAsGroup":65532,"runAsNonRoot":true,"runAsUser":65532,"seccompProfile":{"type":"RuntimeDefault"}}},"resources":{},"securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"privileged":false,"readOnlyRootFilesystem":true,"runAsGroup":65532,"runAsNonRoot":true,"runAsUser":65532,"seccompProfile":{"type":"RuntimeDefault"}},"tolerations":[],"ttlSecondsAfterFinished":30},"rbac":{"annotations":{},"labels":{}}}` | Certgen is used to generate the certificates required by EnvoyGateway. If you want to construct a custom certificate, you can generate a custom certificate through Cert-Manager before installing EnvoyGateway. Certgen will not overwrite the custom certificate. Please do not manually modify `values.yaml` to disable certgen, it may cause EnvoyGateway OIDC,OAuth2,etc. to not work as expected. |
| commonLabels | object | `{}` | Labels to apply to all resources |
| config.envoyGateway | object | `{"extensionApis":{},"gateway":{"controllerName":"gateway.envoyproxy.io/gatewayclass-controller"},"logging":{"level":{"default":"info"}},"provider":{"type":"Kubernetes"}}` | EnvoyGateway configuration. Visit https://gateway.envoyproxy.io/docs/api/extension_types/#envoygateway to view all options. |
| crds.enabled | bool | `true` | Install Envoy Gateway CRDs, Gateway API CRDs, and Gateway API safe upgrade policy resources. Set to false when these resources are managed separately. |
| createNamespace | bool | `false` |  |
| deployment.annotations | object | `{}` |  |
| deployment.envoyGateway.extraEnv | list | `[]` | Additional environment variables for the envoy-gateway container. |
| deployment.envoyGateway.image.repository | string | `""` |  |
| deployment.envoyGateway.image.tag | string | `""` |  |
| deployment.envoyGateway.imagePullPolicy | string | `""` |  |
| deployment.envoyGateway.imagePullSecrets | list | `[]` |  |
| deployment.envoyGateway.livenessProbe.httpGet.path | string | `"/healthz"` |  |
| deployment.envoyGateway.livenessProbe.httpGet.port | int | `8081` |  |
| deployment.envoyGateway.livenessProbe.periodSeconds | int | `20` |  |
| deployment.envoyGateway.livenessProbe.successThreshold | int | `1` |  |
| deployment.envoyGateway.livenessProbe.timeoutSeconds | int | `1` |  |
| deployment.envoyGateway.readinessProbe.httpGet.path | string | `"/readyz"` |  |
| deployment.envoyGateway.readinessProbe.httpGet.port | int | `8081` |  |
| deployment.envoyGateway.readinessProbe.periodSeconds | int | `10` |  |
| deployment.envoyGateway.readinessProbe.successThreshold | int | `1` |  |
| deployment.envoyGateway.readinessProbe.timeoutSeconds | int | `1` |  |
| deployment.envoyGateway.resources.limits.memory | string | `"1024Mi"` |  |
| deployment.envoyGateway.resources.requests.cpu | string | `"100m"` |  |
| deployment.envoyGateway.resources.requests.memory | string | `"256Mi"` |  |
| deployment.envoyGateway.securityContext.allowPrivilegeEscalation | bool | `false` |  |
| deployment.envoyGateway.securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| deployment.envoyGateway.securityContext.privileged | bool | `false` |  |
| deployment.envoyGateway.securityContext.readOnlyRootFilesystem | bool | `true` |  |
| deployment.envoyGateway.securityContext.runAsGroup | int | `65532` |  |
| deployment.envoyGateway.securityContext.runAsNonRoot | bool | `true` |  |
| deployment.envoyGateway.securityContext.runAsUser | int | `65532` |  |
| deployment.envoyGateway.securityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| deployment.envoyGateway.startupProbe.failureThreshold | int | `30` |  |
| deployment.envoyGateway.startupProbe.httpGet.path | string | `"/healthz"` |  |
| deployment.envoyGateway.startupProbe.httpGet.port | int | `8081` |  |
| deployment.envoyGateway.startupProbe.periodSeconds | int | `1` |  |
| deployment.envoyGateway.startupProbe.successThreshold | int | `1` |  |
| deployment.envoyGateway.startupProbe.timeoutSeconds | int | `1` |  |
| deployment.envoyGateway.strategy | object | `{}` | Volume source for the Wasm module cache mounted at /var/lib/eg/wasm. Defaults to an emptyDir when left empty. Example: persist the Wasm module cache across controller restarts by backing it with a PersistentVolumeClaim:   wasmCacheVolume:     persistentVolumeClaim:       claimName: envoy-gateway-wasm-cache |
| deployment.envoyGateway.wasmCacheVolume | object | `{}` |  |
| deployment.pod.affinity | object | `{}` |  |
| deployment.pod.annotations."prometheus.io/port" | string | `"19001"` |  |
| deployment.pod.annotations."prometheus.io/scrape" | string | `"true"` |  |
| deployment.pod.extraVolumeMounts | list | `[]` |  |
| deployment.pod.extraVolumes | list | `[]` |  |
| deployment.pod.labels | object | `{}` |  |
| deployment.pod.nodeSelector | object | `{}` |  |
| deployment.pod.securityContext.fsGroup | int | `65532` |  |
| deployment.pod.securityContext.runAsGroup | int | `65532` |  |
| deployment.pod.securityContext.runAsNonRoot | bool | `true` |  |
| deployment.pod.securityContext.runAsUser | int | `65532` |  |
| deployment.pod.securityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| deployment.pod.tolerations | list | `[]` |  |
| deployment.pod.topologySpreadConstraints | list | `[]` |  |
| deployment.ports[0].name | string | `"grpc"` |  |
| deployment.ports[0].port | int | `18000` |  |
| deployment.ports[0].targetPort | int | `18000` |  |
| deployment.ports[1].name | string | `"ratelimit"` |  |
| deployment.ports[1].port | int | `18001` |  |
| deployment.ports[1].targetPort | int | `18001` |  |
| deployment.ports[2].name | string | `"wasm"` |  |
| deployment.ports[2].port | int | `18002` |  |
| deployment.ports[2].targetPort | int | `18002` |  |
| deployment.ports[3].name | string | `"metrics"` |  |
| deployment.ports[3].port | int | `19001` |  |
| deployment.ports[3].targetPort | int | `19001` |  |
| deployment.priorityClassName | string | `nil` |  |
| deployment.replicas | int | `1` |  |
| global.imagePullSecrets | list | `[]` | Global override for image pull secrets |
| global.imageRegistry | string | `""` | Global override for image registry |
| global.images.envoyGateway.image | string | `nil` | Full image for the Envoy Gateway control plane Deployment installed by this chart. |
| global.images.envoyGateway.pullPolicy | string | `nil` | Image pull policy for the Envoy Gateway control plane Deployment. Default behavior: latest images will be Always else IfNotPresent. |
| global.images.envoyGateway.pullSecrets | list | `[]` | Pull secrets for the Envoy Gateway control plane Deployment. |
| global.images.envoyProxy.image | string | `"docker.io/envoyproxy/envoy:distroless-dev"` | Full image for the managed Envoy Proxy data plane. This updates the generated `envoyProxy` config and does not change the `envoy-gateway` control plane Deployment image. If not specified, the default image built into `envoy-gateway` is used. |
| global.images.envoyProxy.pullPolicy | string | `""` | Image pull policy for the managed Envoy Proxy data plane. Default behavior: IfNotPresent. |
| global.images.envoyProxy.pullSecrets | list | `[]` | Pull secrets for the managed Envoy Proxy data plane. |
| global.images.ratelimit.image | string | `"docker.io/envoyproxy/ratelimit:master"` |  |
| global.images.ratelimit.pullPolicy | string | `"IfNotPresent"` |  |
| global.images.ratelimit.pullSecrets | list | `[]` |  |
| hpa.behavior | object | `{}` |  |
| hpa.enabled | bool | `false` |  |
| hpa.maxReplicas | int | `1` |  |
| hpa.metrics | list | `[]` |  |
| hpa.minReplicas | int | `1` |  |
| kubernetesClusterDomain | string | `"cluster.local"` |  |
| namespaceOverride | string | `""` | Override the namespace for resources deployed by the chart. Defaults to the release namespace. |
| podDisruptionBudget.minAvailable | int | `0` |  |
| service.annotations | object | `{}` |  |
| service.trafficDistribution | string | `""` |  |
| service.type | string | `"ClusterIP"` | Service type. Can be set to LoadBalancer with specific IP, e.g.: type: LoadBalancer loadBalancerIP: 10.236.90.20 |
| topologyInjector.annotations | object | `{}` |  |
| topologyInjector.enabled | bool | `true` |  |

