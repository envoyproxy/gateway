+++
title = "Install with Argo CD"
weight = -99
+++

[Argo CD](https://argo-cd.readthedocs.io) is a declarative, GitOps continuous delivery tool for Kubernetes.
Argo CD can be used to manage the deployment of Envoy Gateway on Kubernetes clusters.

## Before you begin

{{% alert title="Compatibility Matrix" color="warning" %}}
Refer to the [Version Compatibility Matrix](/news/releases/matrix) to learn more.
{{% /alert %}}

{{< boilerplate kind-cluster >}}

Argo CD must be installed in your Kubernetes cluster, and the argocd CLI must be available on your local machine.
If you haven’t set it up yet, you can follow the [official installation guide](https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/) to install Argo CD.

## Install with Argo CD

Create a new Argo CD Application that pulls the Envoy Gateway Helm chart as its source.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: envoy-gateway
  namespace: argocd
spec:
  project: default
  source:
    chart: gateway-helm
    repoURL: docker.io/envoyproxy
    targetRevision: {{< helm-version >}}
  destination:
    namespace: envoy-gateway-system
    server: https://kubernetes.default.svc
  syncPolicy:
    syncOptions:
    - CreateNamespace=true
    - ServerSideApply=true
EOF
```

**Note**:

* Set `ServerSideApply` to `true` to enable Kubernetes [server-side apply](https://kubernetes.io/docs/reference/using-api/server-side-apply/). This helps avoid the 262,144-byte annotation size limit.
* For simplicity, we apply the Application resource directly to the cluster.
In a production environment, it’s recommended to store this configuration in a Git repository and manage it using another Argo CD Application that uses Git as its source — following a GitOps workflow.

Sync the application:

```shell
argocd app sync envoy-gateway
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


## Helm chart customizations

You can customize the Envoy Gateway installation by using the Helm chart values.

{{% alert title="Helm Chart Values" color="primary" %}}
If you want to know all the available fields inside the values.yaml file, please see the [Helm Chart Values](./gateway-helm-api).
{{% /alert %}}

Below is an example of how to customize the Envoy Gateway installation by using the `valuesObject` field in the Argo CD Application.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: envoy-gateway
  namespace: argocd
spec:
  project: default
  source:
    helm:
      valuesObject:
        deployment:
          envoyGateway:
            resources:
              limits:
                cpu: 700m
                memory: 256Mi
    chart: gateway-helm
    path: gateway-helm
    repoURL: docker.io/envoyproxy
    targetRevision: {{< helm-version >}}
  destination:
    namespace: envoy-gateway-system
    server: https://kubernetes.default.svc
  syncPolicy:
    syncOptions:
    - CreateNamespace=true
    - ServerSideApply=true
```

Argo CD supports multiple ways of specifying Helm chart values, you can find more details in the [Argo CD documentation](https://argo-cd.readthedocs.io/en/stable/user-guide/helm/#helm).

{{< boilerplate open-ports >}}

{{% alert title="Next Steps" color="warning" %}}
Envoy Gateway should now be successfully installed and running.  To experience more abilities of Envoy Gateway, refer to [Tasks](../tasks).
{{% /alert %}}
