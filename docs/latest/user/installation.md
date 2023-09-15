# Installation

## Prerequisites

A Kubernetes cluster.

__Note:__ Refer to the [Compatibility Matrix](../intro/compatibility.rst) for supported Kubernetes versions.

## Installation

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

```
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace --set deployment.replicas=2
```

### Change the kubernetesClusterDomain name
If you have installed your cluster with different domain name you can use below command.

```
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace --set kubernetesClusterDomain=<domain name>
```

**Note**: Above are some of the ways we can directly use for customization of our installation. But if you are looking for more complex changes [values.yaml](https://helm.sh/docs/chart_template_guide/values_files/) comes to rescue.

### Using values.yaml file for complex installation.

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

```
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace -f values.yaml
```

If you want to know all the available fields inside the values.yaml file checkout the `helm` section [here](../helm/api.md)

## Open Ports

These are the ports used by Envoy Gateway and the managed Envoy Proxy.

| Envoy Gateway          | Address   |  Port  |
|:----------------------:|:---------:|:------:|
| Xds EnvoyProxy Server  | 0.0.0.0   | 18000  |
| Xds RateLimit Server   | 0.0.0.0   | 18001  |
| Admin Server           | 127.0.0.1 | 19000  |
               


| Envoy Proxy                       | Address     | Port    |
|:---------------------------------:|:-----------:| :-----: |
| Admin Server                      | 127.0.0.1   | 19000   |
| Heath Check & Prometheus Listener | 0.0.0.0     | 19001   |
