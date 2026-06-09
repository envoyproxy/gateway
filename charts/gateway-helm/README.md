# gateway-helm

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

## Usage

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs) to get started.

### Install from DockerHub

Once Helm has been set up correctly, install the chart from dockerhub:

``` shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace
```

Image overrides target different components. `global.images.envoyGateway.*` configures the Envoy Gateway control plane Deployment rendered by this chart. `global.images.envoyProxy.*` configures the managed Envoy Proxy data plane through the generated `EnvoyGateway` config.

This command installs both Gateway API CRDs and Envoy Gateway CRDs. If your Kubernetes provider already manages
Gateway API CRDs for the cluster, confirm that the provider-installed Gateway API version and channel are compatible
with the Envoy Gateway release and the Gateway API resources you plan to use. If they are compatible, install only the
Envoy Gateway CRDs separately and use `--set crds.enabled=false` when installing this chart.

You can find all helm chart release in [Dockerhub](https://hub.docker.com/r/envoyproxy/gateway-helm/tags)

### Install from Source Code

You can also install the helm chart from the source code:

To install the eg chart along with Gateway API CRDs and Envoy Gateway CRDs:

``` shell
make kube-deploy TAG=latest
```

### Skip install CRDs

You can install the eg chart without the bundled Gateway API CRDs, Envoy Gateway CRDs, and Gateway API safe upgrade
policy resources by setting `crds.enabled=false`. Make sure these resources exist in the cluster before installing the
chart with `--set crds.enabled=false`, otherwise Envoy Gateway may fail to start.

If your Kubernetes provider manages compatible Gateway API CRDs, install only the Envoy Gateway CRDs from the
`gateway-crds-helm` chart first:

``` shell
helm template eg-crds oci://docker.io/envoyproxy/gateway-crds-helm --set 'crds.gatewayAPI.enabled=false' --set 'crds.envoyGateway.enabled=true' \
    --version v0.0.0-latest | kubectl apply --server-side -f -
```

If the provider-managed Gateway API CRDs are not compatible, use a compatible Gateway API CRD installation method for
the cluster first, then install this chart with `--set crds.enabled=false`.

After the required CRDs are installed, install the eg chart with `--set crds.enabled=false`. Setting `crds.enabled=false`
also skips the Gateway API safe upgrade policy resources (the safe-upgrades ValidatingAdmissionPolicy and binding shipped
with the Gateway API bundle), so manage them outside this chart when these resources are owned elsewhere:

``` shell
helm install eg --create-namespace oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --set crds.enabled=false
```

To uninstall the chart:

``` shell
helm uninstall eg -n envoy-gateway-system
```

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
| global.images.envoyProxy.image | string | `""` | Full image for the managed Envoy Proxy data plane. This updates the generated `envoyProxy` config and does not change the `envoy-gateway` control plane Deployment image. If not specified, the default image built into `envoy-gateway` is used. |
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
| monitoring.podMonitor.annotations | object | `{}` | Annotations to add to the PodMonitor |
| monitoring.podMonitor.authorization | object | `{}` | Authorization configuration for the scrape endpoint. |
| monitoring.podMonitor.enabled | bool | `false` | Enable podMonitor for Envoy Gateway |
| monitoring.podMonitor.interval | string | `""` | Interval at which metrics should be scraped. Defaults to Prometheus default. |
| monitoring.podMonitor.labels | object | `{}` | Labels to add to the PodMonitor |
| monitoring.podMonitor.metricRelabelings | list | `[]` | MetricRelabelConfigs to apply to samples before storing. |
| monitoring.podMonitor.namespace | string | `""` | Namespace for the PodMonitor. Defaults to the release namespace. |
| monitoring.podMonitor.podTargetLabels | list | `[]` | Additional labels from pod metadata to include as metric labels. |
| monitoring.podMonitor.relabelings | list | `[]` | RelabelConfigs to apply to samples before scraping. |
| monitoring.podMonitor.telemetryPath | string | `"/metrics"` | Path to scrape metrics from. |
| monitoring.podMonitor.timeout | string | `""` | Timeout for scrape requests. |
| namespaceOverride | string | `""` | Override the namespace for resources deployed by the chart. Defaults to the release namespace. |
| podDisruptionBudget.minAvailable | int | `0` |  |
| service.annotations | object | `{}` |  |
| service.trafficDistribution | string | `""` |  |
| service.type | string | `"ClusterIP"` | Service type. Can be set to LoadBalancer with specific IP, e.g.: type: LoadBalancer loadBalancerIP: 10.236.90.20 |
| topologyInjector.annotations | object | `{}` |  |
| topologyInjector.enabled | bool | `true` |  |

