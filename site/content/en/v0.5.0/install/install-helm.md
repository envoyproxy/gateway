+++
title = "Install with Helm"
weight = -100
+++

[Helm](https://helm.sh) is a package manager for Kubernetes that automates the release and management of software on Kubernetes. 

Envoy Gateway can be installed via a Helm chart with a few simple steps, depending on if you are deploying for the first time, upgrading Envoy Gateway from an existing installation, or migrating from Envoy Gateway.

## Before you begin

{{% alert title="Compatibility Matrix" color="warning" %}}
Refer to the [Version Compatibility Matrix](/blog/2022/10/01/versions/) to learn more.
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
Refer to the [Developer Guide](../../contributions/develop) to learn more.
{{% /alert %}}

When you run the Helm chart, it installs Envoy Gateway.

1. Install the Gateway API CRDs and Envoy Gateway with the following command:

    ``` shell
    helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace
    ```

2. Next Steps

   Envoy Gateway should now be successfully installed and running, but in order to experience more abilities of Envoy Gateway, you can refer to [User Guides](../../user).

For more advanced configuration and details about helm values, please see the [Helm Chart Values](../api).
