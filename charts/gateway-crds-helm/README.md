# gateway-crds-helm

![Version: v0.0.0-latest](https://img.shields.io/badge/Version-v0.0.0--latest-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: latest](https://img.shields.io/badge/AppVersion-latest-informational?style=flat-square)

A Helm chart for Envoy Gateway CRDs

**Homepage:** <https://gateway.envoyproxy.io/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| envoy-gateway-steering-committee |  | <https://github.com/envoyproxy/gateway/blob/main/GOVERNANCE.md> |
| envoy-gateway-maintainers |  | <https://github.com/envoyproxy/gateway/blob/main/CODEOWNERS> |

## Source Code

* <https://github.com/envoyproxy/gateway>

## Usage

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs) to get started.

If you want to manage the CRDs outside of the Envoy Gateway Helm chart, you can use this chart to install the CRDs separately.
If you do, make sure that you don't install the CRDs again when installing the Envoy Gateway Helm chart, by using `--set crds.enabled=false`.
If your Kubernetes provider already manages Gateway API CRDs for the cluster, compare the provider-installed Gateway API
version and channel with the [Envoy Gateway compatibility matrix](https://gateway.envoyproxy.io/news/releases/matrix/)
and the Gateway API resources you plan to use. If they are compatible, leave those CRDs disabled in this chart and
install only the Envoy Gateway CRDs.
You can check the installed Gateway API version and channel from the CRD annotations:

``` shell
kubectl get crd gateways.gateway.networking.k8s.io \
  -o go-template='version={{ index .metadata.annotations "gateway.networking.k8s.io/bundle-version" }} channel={{ index .metadata.annotations "gateway.networking.k8s.io/channel" }}{{ "\n" }}'
```

### Install from DockerHub

Once Helm has been set up correctly, install the chart from dockerhub:

``` shell
helm template eg-crds oci://docker.io/envoyproxy/gateway-crds-helm --set 'crds.gatewayAPI.enabled=true' --set 'crds.envoyGateway.enabled=true' \
    --version v0.0.0-latest | kubectl apply --server-side -f -
```

For clusters with compatible provider-managed Gateway API CRDs, install only the Envoy Gateway CRDs:

``` shell
helm template eg-crds oci://docker.io/envoyproxy/gateway-crds-helm --set 'crds.gatewayAPI.enabled=false' --set 'crds.envoyGateway.enabled=true' \
    --version v0.0.0-latest | kubectl apply --server-side -f -
```

If the provider-managed Gateway API CRDs are not compatible, use a compatible Gateway API CRD installation method for
the cluster instead of mixing provider-managed CRDs with another Gateway API CRD copy from this chart.

**Note**: We're using `helm template` piped into `kubectl apply` instead of `helm install` due to a [known Helm limitation](https://github.com/helm/helm/pull/12277)
related to large CRDs in the `templates/` directory.

You can find all helm chart release in [Dockerhub](https://hub.docker.com/r/envoyproxy/gateway-crds-helm/tags)

To uninstall the chart:

``` shell
helm template eg-crds oci://docker.io/envoyproxy/gateway-crds-helm \
    --version v0.0.0-latest | kubectl delete --server-side -f -
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| crds.envoyGateway.enabled | bool | `false` |  |
| crds.gatewayAPI.channel | string | `"experimental"` |  |
| crds.gatewayAPI.enabled | bool | `false` |  |

