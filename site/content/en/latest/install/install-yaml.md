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

{{% alert title="Gateway API CRD compatibility" color="warning" %}}
The `install.yaml` manifest includes Gateway API CRDs and Envoy Gateway CRDs. If your Kubernetes provider already
manages compatible Gateway API CRDs for the cluster, do not use `install.yaml`; follow the
[provider-managed Gateway API CRD Helm steps](../install-helm/#clusters-with-compatible-provider-managed-gateway-api-crds)
instead. If the provider-managed Gateway API CRDs are not compatible with your Envoy Gateway release or required
Gateway API resources, use a compatible Gateway API CRD installation method before installing Envoy Gateway.
{{% /alert %}}

1. In your terminal, run the following command:

    ```shell
    kubectl apply --server-side -f https://github.com/envoyproxy/gateway/releases/download/{{< yaml-version >}}/install.yaml
    ```

2. Next Steps

   Envoy Gateway should now be successfully installed and running, but in order to experience more abilities of Envoy Gateway, you can refer to [Tasks](/latest/tasks).

## Upgrading from the previous version

Some manual migration steps are required to upgrade Envoy Gateway.

{{% alert title="Gateway API v1.6 CRD upgrade required before Envoy Gateway upgrade" color="warning" %}}
This release reconciles `TCPRoute` and `UDPRoute` via the `gateway.networking.k8s.io/v1` API group (promoted in Gateway API v1.6).
**You must upgrade the Gateway API CRDs to v1.6 before upgrading Envoy Gateway** to avoid traffic disruption.

**Standard channel users:** `v1alpha2` is no longer served in the Gateway API v1.6 standard channel.
You must update all `TCPRoute` and `UDPRoute` manifests to `apiVersion: gateway.networking.k8s.io/v1`
before upgrading the CRDs, or those routes will stop being served and traffic will be dropped.

**Experimental channel users:** both `v1` and `v1alpha2` are served after the CRD upgrade, so existing
`v1alpha2` manifests continue to work without immediate changes. Updating manifests to `v1` is still recommended.

If the v1.6 CRDs are not installed before Envoy Gateway is upgraded, TCP and UDP routes will be silently skipped until the CRDs are applied.

The stored version of `TCPRoute` and `UDPRoute` moves from `v1alpha2` to `v1`. Plan a storage-version migration
before `v1alpha2` is eventually removed:

```shell
kubectl get tcproutes.gateway.networking.k8s.io -A -o json | kubectl replace -f -
kubectl get udproutes.gateway.networking.k8s.io -A -o json | kubectl replace -f -
```
{{% /alert %}}

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
