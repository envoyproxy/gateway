# gateway-crds-helm

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

A Helm chart for Kubernetes

## Usage

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs) to get started.

If you want to manage the CRDs outside of the Envoy Gateway Helm chart, you can use this chart to install the CRDs separately.
If you do, make sure that you don't install the CRDs again when installing the Envoy Gateway Helm chart, by using `--skip-crds` flag.

### Install from DockerHub

Once Helm has been set up correctly, install the chart from dockerhub:

``` shell
    helm install gateway-crds oci://docker.io/envoyproxy/gateway-crds-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace
```

You can find all helm chart release in [Dockerhub](https://hub.docker.com/r/envoyproxy/gateway-crds-helm/tags)

To uninstall the chart:

``` shell
    helm uninstall gateway-crds -n envoy-gateway-system
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| crds.envoyGateway.enabled | bool | `false` |  |
| crds.gatewayAPI.channel | string | `"experimental"` |  |
| crds.gatewayAPI.enabled | bool | `false` |  |

