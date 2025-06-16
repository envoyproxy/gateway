+++
title = "Install with Helm"
weight = -100
+++

[Helm](https://helm.sh) is a package manager for Kubernetes that automates the release and management of software on Kubernetes.

Envoy Gateway can be installed via a Helm chart with a few simple steps, depending on if you are deploying for the first time, upgrading Envoy Gateway from an existing installation, or migrating from Envoy Gateway.

## Before you begin

{{% alert title="Compatibility Matrix" color="warning" %}}
Refer to the [Version Compatibility Matrix](/news/releases/matrix) to learn more.
{{% /alert %}}

{{< boilerplate kind-cluster >}}

The Envoy Gateway Helm chart is hosted by DockerHub.

It is published at `oci://docker.io/envoyproxy/gateway-helm`.

{{% alert title="Note" color="primary" %}}
We use `v0.0.0-latest` as the latest development version.

You can visit [Envoy Gateway Helm Chart](https://hub.docker.com/r/envoyproxy/gateway-helm/tags) for more releases.
{{% /alert %}}

## Install with Helm

Install the Gateway API CRDs and Envoy Gateway:

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version {{< helm-version >}} -n envoy-gateway-system --create-namespace
```

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

## Installing CRDs separately

The [Envoy Gateway Helm Chart](https://hub.docker.com/r/envoyproxy/gateway-helm/tags) includes two types of CRDs under its crds/ directory:
* [Gateway API CRDs](../concepts/introduction/#gateway-api/) (from the *experimental* channel): This channel includes additional `alpha` resources
such as `TCPRoute` and `BackendTLSPolicy`, which are commonly used by Envoy Gateway users.
* [Envoy Gateway CRDs](../concepts/introduction/#gateway_api_extensions): These are custom CRDs that extend the Gateway API to support additional
Envoy Gateway-specific features.

If you prefer to manage CRDs separately, the [Envoy Gateway CRDs Helm Chart](https://hub.docker.com/r/envoyproxy/gateway-crds-helm/tags) allows you
to install just the CRDs, with fine-grained control over:
* Which Gateway API channel to use (`standard` or `experimental`)
* Whether to include Envoy Gateway-specific CRDs

Use the following command to install the CRDs using `helm template` and `kubectl`

```shell
helm template eg oci://docker.io/envoyproxy/gateway-crds-helm \
  --version {{< helm-version >}} \
  --set crds.gatewayAPI.enabled=true \
  --set crds.gatewayAPI.channel=standard \
  --set crds.envoyGateway.enabled=true \
  | kubectl apply --server-side -f -
```

**Note**: We're using `helm template` piped into `kubectl apply` instead of `helm install` due to a [known Helm limitation](https://github.com/helm/helm/pull/12277)
related to large CRDs in the `templates/` directory.

Once the CRDs are installed, you can install the main Envoy Gateway Helm chart without re-applying CRDs by using the `--skip-crds` flag:

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version {{< helm-version >}} \
  -n envoy-gateway-system \
  --create-namespace \
  --skip-crds
```

## Upgrading from the previous version

[Helm](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/#some-caveats-and-explanations) does not update CRDs
that live in the `/crds` folder in the Helm Chart. So you will manually need to update the CRDs.
Follow the steps outlined in [this](./install-yaml/#upgrading-from-a-previous-version) section if you're upgrading from a previous version.

## Helm chart customizations

Some of the quick ways of using the helm install command for envoy gateway installation are below.

{{% alert title="Helm Chart Values" color="primary" %}}
If you want to know all the available fields inside the values.yaml file, please see the [Helm Chart Values](./gateway-helm-api).
{{% /alert %}}

### Increase the replicas

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version {{< helm-version >}} -n envoy-gateway-system --create-namespace --set deployment.replicas=2
```

### Change the kubernetesClusterDomain name

If you have installed your cluster with different domain name you can use below command.

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version {{< helm-version >}} -n envoy-gateway-system --create-namespace --set kubernetesClusterDomain=<domain name>
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
helm install eg oci://docker.io/envoyproxy/gateway-helm --version {{< helm-version >}} -n envoy-gateway-system --create-namespace -f values.yaml
```

{{< boilerplate open-ports >}}

{{% alert title="Next Steps" color="warning" %}}
Envoy Gateway should now be successfully installed and running.  To experience more abilities of Envoy Gateway, refer to [Tasks](../tasks).
{{% /alert %}}
