# Usage

[Helm](https://helm.sh) must be installed to use the charts.  Please refer to
Helm's [documentation](https://helm.sh/docs) to get started.

Once Helm has been set up correctly, add the repo as follows:

``` shell
  helm repo add eg https://gateway.envoyproxy.io/helm-charts
```

If you had already added this repo earlier, run `helm repo update` to retrieve
the latest versions of the packages.  You can then run `helm search repo
eg` to see the charts.

To install the eg chart:

``` shell
    helm install envoy-gateway --create-namespace charts/eg -n envoy-gateway-system
```

To uninstall the chart:

``` shell
    helm delete envoy-gateway
```
