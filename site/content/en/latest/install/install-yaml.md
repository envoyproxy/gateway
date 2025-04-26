+++
title = "Install with Kubernetes YAML"
weight = -98
+++

This task walks you through installing Envoy Gateway in your Kubernetes cluster.

The manual install process does not allow for as much control over configuration
as the [Helm install method](./install-helm), so if you need more control over your Envoy Gateway
installation, it is recommended that you use helm.

## Before you begin

{{% alert title="Compatibility Matrix" color="warning" %}}
Refer to the [Version Compatibility Matrix](/news/releases/matrix) to learn more.
{{% /alert %}}

Envoy Gateway is typically deployed in a Kubernetes cluster.
If you don’t have one yet, you can use `kind` to create a local cluster for testing purposes.

{{% alert title="Developer Guide" color="primary" %}}
Refer to the [Developer Guide](../../contributions/develop) to learn more.
{{% /alert %}}

## Install with YAML

1. In your terminal, run the following command:

    ```shell
    kubectl apply --server-side -f https://github.com/envoyproxy/gateway/releases/download/{{< yaml-version >}}/install.yaml
    ```

2. Next Steps

   Envoy Gateway should now be successfully installed and running, but in order to experience more abilities of Envoy Gateway, you can refer to [Tasks](/latest/tasks).

## Upgrading from v1.2

Some manual migration steps are required to upgrade Envoy Gateway to v1.3.

1. Update Gateway-API and Envoy Gateway CRDs:

```shell
helm pull oci://docker.io/envoyproxy/gateway-helm --version {{< yaml-version >}} --untar
kubectl apply --force-conflicts --server-side -f ./gateway-helm/crds/gatewayapi-crds.yaml
kubectl apply --force-conflicts --server-side -f ./gateway-helm/crds/generated
```

2. Install Envoy Gateway {{< yaml-version >}}:

```shell
helm upgrade eg oci://docker.io/envoyproxy/gateway-helm --version {{< yaml-version >}} -n envoy-gateway-system
```

## Open Ports

These are the ports used by Envoy Gateway and the managed Envoy Proxy.

### Envoy Gateway

|     Envoy Gateway     |  Address  | Port  | Configurable |
| :-------------------: | :-------: | :---: | :----------: |
| Xds EnvoyProxy Server |  0.0.0.0  | 18000 |      No      |
| Xds RateLimit Server  |  0.0.0.0  | 18001 |      No      |
|     Admin Server      | 127.0.0.1 | 19000 |     Yes      |
|    Metrics Server     |  0.0.0.0  | 19001 |      No      |
|     Health Check      | 127.0.0.1 | 8081  |      No      |

### EnvoyProxy

|   Envoy Proxy    |  Address  | Port  |
| :--------------: | :-------: | :---: |
|   Admin Server   | 127.0.0.1 | 19000 |
|      Stats       |  0.0.0.0  | 19001 |
| Shutdown Manager |  0.0.0.0  | 19002 |
|    Readiness     |  0.0.0.0  | 19003 |

{{% alert title="Next Steps" color="warning" %}}
Envoy Gateway should now be successfully installed and running.  To experience more abilities of Envoy Gateway, refer to [Tasks](../tasks).
{{% /alert %}}
