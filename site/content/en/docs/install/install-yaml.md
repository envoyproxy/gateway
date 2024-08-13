+++
title = "Install with Kubernetes YAML"
weight = -99
+++

This task walks you through installing Envoy Gateway in your Kubernetes cluster.

The manual install process does not allow for as much control over configuration
as the [Helm install method](./install-helm), so if you need more control over your Envoy Gateway
installation, it is recommended that you use helm.

## Before you begin

Envoy Gateway is designed to run in Kubernetes for production. The most essential requirements are:

* Kubernetes 1.27 or later
* The `kubectl` command-line tool

{{% alert title="Compatibility Matrix" color="warning" %}}
Refer to the [Version Compatibility Matrix](/news/releases/matrix) to learn more.
{{% /alert %}}

## Install with YAML

Envoy Gateway is typically deployed to Kubernetes from the command line. If you don't have Kubernetes, you should use `kind` to create one.

{{% alert title="Developer Guide" color="primary" %}}
Refer to the [Developer Guide](../../contributions/develop) to learn more.
{{% /alert %}}

1. In your terminal, run the following command:

    ```shell
    kubectl apply --server-side -f https://github.com/envoyproxy/gateway/releases/download/{{< yaml-version >}}/install.yaml
    ```

2. Next Steps

   Envoy Gateway should now be successfully installed and running, but in order to experience more abilities of Envoy Gateway, you can refer to [Tasks](/latest/tasks).

## Upgrading from v1.0

Due to breaking changes in Gateway API v1.1, some manual migration steps are required to upgrade Envoy Gateway to v1.1.

1. Delete `BackendTLSPolicy` CRD (and resources):

```shell
kubectl delete crd backendtlspolicies.gateway.networking.k8s.io
```

2. Update Gateway-API and Envoy Gateway CRDs:

```shell
helm pull oci://docker.io/envoyproxy/gateway-helm --version {{< yaml-version >}} --untar
kubectl apply --force-conflicts --server-side -f ./gateway-helm/crds/gatewayapi-crds.yaml
kubectl apply --force-conflicts --server-side -f ./gateway-helm/crds/generated
```

3. Update your `BackendTLSPolicy` and `GRPCRoute` resources according to Gateway-API [v1.1 Upgrade Notes](https://gateway-api.sigs.k8s.io/guides/#v11-upgrade-notes)

4. Update your Envoy Gateway xPolicy resources: remove the namespace section from targetRef.

5. Install Envoy Gateway {{< yaml-version >}}:

```shell
helm upgrade eg oci://docker.io/envoyproxy/gateway-helm --version {{< yaml-version >}} -n envoy-gateway-system
```
