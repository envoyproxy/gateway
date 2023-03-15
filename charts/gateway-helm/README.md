# Usage

[Helm](https://helm.sh) must be installed to use the charts.  Please refer to
Helm's [documentation](https://helm.sh/docs) to get started.

## Install from DockerHub

Once Helm has been set up correctly, install the chart from dockerhub:

``` shell
  helm install eg oci://docker.io/envoyproxy/gateway-helm -n envoy-gateway-system --create-namespace
```

## Install from Source Code

You can also install the helm chart from the source code:

To install the eg chart along with Gateway API CRDs and Envoy Gateway CRDs:

``` shell
    helm install eg --create-namespace charts/gateway-helm -n envoy-gateway-system
```

## Skip install CRDs

You can install the eg chart along without Gateway API CRDs and Envoy Gateway CRDs, make sure CRDs exist in Cluster first if you want to skip to install them, otherwise EG may fail to start:

``` shell
    helm install eg --create-namespace charts/gateway-helm -n envoy-gateway-system --skip-crds
```

To uninstall the chart:

``` shell
    helm delete eg
```
