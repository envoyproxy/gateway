+++
title = "Install with Helm"
weight = -100
+++

[Helm](https://helm.sh) is a package manager for Kubernetes that automates the release and management of software on Kubernetes. 

Envoy Gateway can be installed via a Helm chart with a few simple steps, depending on if you are deploying for the first time, upgrading Envoy Gateway from an existing installation, or migrating from Envoy Gateway.

## Before you begin

{{% alert title="Compatibility Matrix" color="warning" %}}
Refer to the [Version Compatibility Matrix](./matrix) to learn more.
{{% /alert %}}

The Envoy Gateway Helm chart is hosted by DockerHub.

It is published at `oci://docker.io/envoyproxy/gateway-helm`.

{{% alert title="Note" color="primary" %}}
We use `v0.0.0-latest` as the latest development version.

You can visit [Envoy Gateway Helm Chart](https://hub.docker.com/r/envoyproxy/gateway-helm/tags) for more releases.
{{% /alert %}}

## Install with Helm

Envoy Gateway is typically deployed to Kubernetes from the command line. If you don't have Kubernetes, you should use `kind` to create one.

{{% alert title="Developer Guide" color="primary" %}}
Refer to the [Developer Guide](/latest/contributions/develop) to learn more.
{{% /alert %}}

Install the Gateway API CRDs and Envoy Gateway:

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace
```

Wait for Envoy Gateway to become available:

```shell
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

Install the GatewayClass, Gateway, HTTPRoute and example app:

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/latest/quickstart.yaml -n default
```

**Note**: [`quickstart.yaml`] defines that Envoy Gateway will listen for
traffic on port 80 on its globally-routable IP address, to make it easy to use
browsers to test Envoy Gateway. When Envoy Gateway sees that its Listener is
using a privileged port (<1024), it will map this internally to an
unprivileged port, so that Envoy Gateway doesn't need additional privileges.
It's important to be aware of this mapping, since you may need to take it into
consideration when debugging.

[`quickstart.yaml`]: https://github.com/envoyproxy/gateway/releases/download/latest/quickstart.yaml

## Helm chart customizations

Some of the quick ways of using the helm install command for envoy gateway installation are below. 

### Increase the replicas

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace --set deployment.replicas=2
```

### Change the kubernetesClusterDomain name

If you have installed your cluster with different domain name you can use below command.

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace --set kubernetesClusterDomain=<domain name>
```

**Note**: Above are some of the ways we can directly use for customization of our installation. But if you are looking for more complex changes [values.yaml](https://helm.sh/docs/chart_template_guide/values_files/) comes to rescue.

### Using values.yaml file for complex installation

```yaml
deployment:
  envoyGateway:
    resources:
      limits:
        cpu: 700m
        memory: 128Mi
      requests:
        cpu: 10m
        memory: 64Mi
  ports:
    - name: grpc
      port: 18005
      targetPort: 18000
    - name: ratelimit
      port: 18006
      targetPort: 18001

config:
  envoyGateway:
    logging:
      level:
        default: debug
```

Here we have made three changes to our values.yaml file. Increase the resources limit for cpu to `700m`, changed the port for grpc to `18005` and for ratelimit to `18006` and also updated the logging level to `debug`.

You can use the below command to install the envoy gateway using values.yaml file.

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace -f values.yaml
```

{{% alert title="Helm Chart Values" color="primary" %}}
If you want to know all the available fields inside the values.yaml file, please see the [Helm Chart Values](../api).
{{% /alert %}}

## Open Ports

These are the ports used by Envoy Gateway and the managed Envoy Proxy.

### Envoy Gateway

| Envoy Gateway          | Address   |  Port  |  Configurable  |
|:----------------------:|:---------:|:------:|    :------:    |
| Xds EnvoyProxy Server  | 0.0.0.0   | 18000  |       No       |
| Xds RateLimit Server   | 0.0.0.0   | 18001  |       No       |
| Admin Server           | 127.0.0.1 | 19000  |       Yes      |
| Metrics Server         |  0.0.0.0  | 19001  |       No       |
| Health Check           | 127.0.0.1 |  8081  |       No       |

### EnvoyProxy

| Envoy Proxy                       | Address     | Port    |
|:---------------------------------:|:-----------:| :-----: |
| Admin Server                      | 127.0.0.1   | 19000   |
| Heath Check  | 0.0.0.0     | 19001   |

{{% alert title="Next Steps" color="warning" %}}
Envoy Gateway should now be successfully installed and running, but in order to experience more abilities of Envoy Gateway, you can refer to [User Guides](../user).
{{% /alert %}}
