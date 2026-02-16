---
title: 'Deploy Envoy Gateway in Air-Gapped Environments'
---

Deploying the Envoy Gateway in an air-gapped environment using a Helm chart
requires careful configuration of the `values.yaml` file as well as adjustments
when deploying a Gateway resource.

You will need to specify custom image repositories for the following components
in the Helm chart. This can be done on a global level or image level.

- Gateway
- Ratelimit

## Gateway – `values.yaml` Configuration

Example done in image level:

```yaml
deployment:
  envoyGateway:
    image:
      repository: custom-cr.internal.io/envoyproxy/gateway
      tag: v1.4.1
```

It's also possible to define the registry on a global level:

```yaml
# Global settings
global:
  # If set, these take highest precedence and change both envoyGateway and ratelimit's container registry and pull secrets.
  # -- Global override for image registry
  imageRegistry: 'custom-cr.internal.io'
```

## Ratelimit - `values.yaml` Configuration

Example done on global level:

```yaml
global:
  images:
    ratelimit:
      image: custom-cr.internal.io/envoyproxy/ratelimit:master
```

Furthermore for private registries you might need to define imagePullSecrets.

On global level:

```yaml
global:
  imagePullSecrets:
    - my-private-registry-secret
```

or per image

```yaml
global:
  images:
    ratelimit:
      pullSecrets:
        - name: my-private-registry-secret
```

## Gateway Requires a Custom EnvoyProxy Reference

Either the Gateway or GatewayClass must reference a custom EnvoyProxy resource
that explicitly specifies the location of the distroless Envoy container image.
Without this, the image will be pulled implicitly from Docker Hub.

For air-gapped deployments, you must configure the EnvoyProxy to use your internal container registry:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-envoy-proxy
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          image: custom-cr.internal.io/envoyproxy/envoy:distroless-v1.34.1
```

For comprehensive EnvoyProxy configuration options including deployment settings, resource limits, annotations, and other customizations, see [Customize EnvoyProxy](customize-envoyproxy).

## Default LoadBalancer Service Type

By default, Envoy uses a Service of type `LoadBalancer`. In air-gapped environments, 
you may need to configure service annotations or change the service type depending 
on your Kubernetes environment and network restrictions.

For detailed service configuration options including annotations, service types, and other networking customizations, see [Customize EnvoyProxy](customize-envoyproxy).
