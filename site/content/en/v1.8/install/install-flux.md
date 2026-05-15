+++
title = "Install with Flux CD"
weight = -98
+++

[Flux](https://fluxcd.io) is a CNCF-graduated, GitOps-based continuous delivery tool for Kubernetes that reconciles cluster state from a Git repository or OCI registry.
Flux can be used to manage the deployment of Envoy Gateway on Kubernetes clusters.

## Before you begin

{{% alert title="Compatibility Matrix" color="warning" %}}
Refer to the [Version Compatibility Matrix](/news/releases/matrix) to learn more.
{{% /alert %}}

{{< boilerplate kind-cluster >}}

Flux must be installed in your Kubernetes cluster.
If you haven't set it up yet, follow the [Flux installation guide](https://fluxcd.io/flux/installation/).
You can use the `flux` CLI, the [Flux Operator](https://fluxoperator.dev/), or any other supported method.

## Install with Flux

The Envoy Gateway Helm chart is published as an OCI artifact at `oci://docker.io/envoyproxy/gateway-helm`.
Create an `OCIRepository` source and a `HelmRelease` that installs the chart into the `envoy-gateway-system` namespace.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: envoy-gateway-system
---
apiVersion: source.toolkit.fluxcd.io/v1
kind: OCIRepository
metadata:
  name: gateway-helm
  namespace: envoy-gateway-system
spec:
  interval: 1h
  url: oci://docker.io/envoyproxy/gateway-helm
  layerSelector:
    mediaType: "application/vnd.cncf.helm.chart.content.v1.tar+gzip"
    operation: copy
  ref:
    tag: {{< helm-version >}}
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: envoy-gateway
  namespace: envoy-gateway-system
spec:
  interval: 5m
  releaseName: eg
  chartRef:
    kind: OCIRepository
    name: gateway-helm
  upgrade:
    strategy:
      name: RetryOnFailure
      retryInterval: 5m
EOF
```

**Note**: For simplicity, we apply these manifests directly to the cluster.
In a production environment, it's recommended to store this configuration in a Git or OCI source that Flux reconciles, following a GitOps workflow.

Wait for Envoy Gateway to become available:

```shell
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

Install the GatewayClass, Gateway, HTTPRoute and example app:

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/{{< yaml-version >}}/quickstart.yaml -n default
```

**Note**: [`quickstart.yaml`] defines that Envoy Gateway will listen for
traffic on port 80 on its globally-routable IP address, to make it easy to use
browsers to test Envoy Gateway. When Envoy Gateway sees that its Listener is
using a privileged port (<1024), it will map this internally to an
unprivileged port, so that Envoy Gateway doesn't need additional privileges.
It's important to be aware of this mapping, since you may need to take it into
consideration when debugging.

[`quickstart.yaml`]: https://github.com/envoyproxy/gateway/releases/download/{{< yaml-version >}}/quickstart.yaml


## Helm chart customizations

You can customize the Envoy Gateway installation by setting Helm chart values on the `HelmRelease`.

{{% alert title="Helm Chart Values" color="primary" %}}
If you want to know all the available fields inside the values.yaml file, please see the [Helm Chart Values](./gateway-helm-api).
{{% /alert %}}

Below is an example of how to customize the Envoy Gateway installation by using the `values` field on the `HelmRelease`.

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: envoy-gateway
  namespace: envoy-gateway-system
spec:
  interval: 5m
  releaseName: eg
  chartRef:
    kind: OCIRepository
    name: gateway-helm
  upgrade:
    strategy:
      name: RetryOnFailure
      retryInterval: 5m
  values:
    deployment:
      envoyGateway:
        resources:
          limits:
            cpu: 700m
            memory: 256Mi
```

For values stored in a `ConfigMap` or `Secret`, or for advanced merge strategies, see the [Flux HelmRelease values reference](https://fluxcd.io/flux/components/helm/helmreleases/#values).

{{< boilerplate open-ports >}}

{{% alert title="Next Steps" color="warning" %}}
Envoy Gateway should now be successfully installed and running.  To experience more abilities of Envoy Gateway, refer to [Tasks](../tasks).
{{% /alert %}}
