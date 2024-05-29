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

## Usage

[Helm](https://helm.sh) must be installed to use the charts.  Please refer to
Helm's [documentation](https://helm.sh/docs) to get started.

### Install from DockerHub

Once Helm has been set up correctly, install the chart from dockerhub:

``` shell
  helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace
```
You can find all helm chart release in [Dockerhub](https://hub.docker.com/r/envoyproxy/gateway-helm/tags)

### Install from Source Code

You can also install the helm chart from the source code:

To install the eg chart along with Gateway API CRDs and Envoy Gateway CRDs:

``` shell
    make kube-deploy TAG=latest
```

### Skip install CRDs

You can install the eg chart along without Gateway API CRDs and Envoy Gateway CRDs, make sure CRDs exist in Cluster first if you want to skip to install them, otherwise EG may fail to start:

``` shell
    helm install eg --create-namespace oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --skip-crds
```

To uninstall the chart:

``` shell
    helm delete eg
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| certgen.job.annotations | object | `{}` |  |
| certgen.job.resources | object | `{}` |  |
| certgen.job.ttlSecondsAfterFinished | int | `30` |  |
| certgen.rbac.annotations | object | `{}` |  |
| certgen.rbac.labels | object | `{}` |  |
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

