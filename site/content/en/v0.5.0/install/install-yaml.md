+++
title = "Install with Kubernetes YAML"
weight = -99
+++

In this guide, we'll walk you through installing Envoy Gateway in your Kubernetes cluster.

The manual install process does not allow for as much control over configuration
as the [Helm install method](../install-helm), so if you need more control over your Envoy Gateway
installation, it is recommended that you use helm.

## Before you begin

Envoy Gateway is designed to run in Kubernetes for production. The most essential requirements are:

* Kubernetes 1.25 or later
* The `kubectl` command-line tool

{{% alert title="Compatibility Matrix" color="warning" %}}
Refer to the [Version Compatibility Matrix](/blog/2022/10/01/versions/) to learn more.
{{% /alert %}}

## Install with YAML

Envoy Gateway is typically deployed to Kubernetes from the command line. If you don't have Kubernetes, you should use `kind` to create one.

{{% alert title="Developer Guide" color="primary" %}}
Refer to the [Developer Guide](../../contributions/develop) to learn more.
{{% /alert %}}

1. In your terminal, run the following command:

    ```shell
    kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/latest/install.yaml
    ```

2. Next Steps

   Envoy Gateway should now be successfully installed and running, but in order to experience more abilities of Envoy Gateway, you can refer to [User Guides](../../user).
