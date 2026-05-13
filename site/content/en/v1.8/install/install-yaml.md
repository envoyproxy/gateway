+++
title = "Install with Kubernetes YAML"
weight = -98
+++

This task walks you through installing Envoy Gateway in your Kubernetes cluster.

The manual installation process does not allow for as much configuration control (e.g. when you are using a custom domain Kubernetes cluster)
as the [Helm install method](./install-helm), so if you need more control over your Envoy Gateway
installation, it is recommended that you use helm.

## Before you begin

{{% alert title="Compatibility Matrix" color="warning" %}}
Refer to the [Version Compatibility Matrix](/news/releases/matrix) to learn more.
{{% /alert %}}

{{< boilerplate kind-cluster >}}

## Install with YAML

1. In your terminal, run the following command:

    ```shell
    kubectl apply --server-side -f https://github.com/envoyproxy/gateway/releases/download/{{< yaml-version >}}/install.yaml
    ```

2. Next Steps

   Envoy Gateway should now be successfully installed and running, but in order to experience more abilities of Envoy Gateway, you can refer to [Tasks](/latest/tasks).

## Upgrading from the previous version

Some manual migration steps are required to upgrade Envoy Gateway.

1. Update Gateway-API and Envoy Gateway CRDs:

```shell
helm template eg-crds oci://docker.io/envoyproxy/gateway-crds-helm \
  --version {{< yaml-version >}} \
  --set crds.gatewayAPI.enabled=true \
  --set crds.envoyGateway.enabled=true \
  | kubectl apply --force-conflicts --server-side -f -
```

2. Install Envoy Gateway {{< yaml-version >}}:

```shell
helm upgrade eg oci://docker.io/envoyproxy/gateway-helm --version {{< yaml-version >}} -n envoy-gateway-system
```

{{< boilerplate open-ports >}}

{{% alert title="Next Steps" color="warning" %}}
Envoy Gateway should now be successfully installed and running.  To experience more abilities of Envoy Gateway, refer to [Tasks](../tasks).
{{% /alert %}}
