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

Then, reference the `custom-envoy-proxy` in your Gateway manifest

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: envoy-gateway
  namespace: default
  annotations:
    cert-manager.io/cluster-issuer: cluster-wide-ca-issuer
    cert-manager.io/duration: 8760h
    cert-manager.io/renew-before: 360h
    cert-manager.io/usages: server auth, client auth
spec:
  gatewayClassName: envoy-gateway-class
  infrastructure:
    parametersRef:
      group: gateway.envoyproxy.io
      kind: EnvoyProxy
      name: custom-envoy-proxy
  listeners:
    - hostname: example.com
      name: https
      port: 443
      protocol: HTTPS
      tls:
        certificateRefs:
          - name: example-tls
        mode: Terminate
```

or reference it in your GatewayClass, so that each new Gateway uses the
EnvoyProxy automatically:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: envoy-gateway-class
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: custom-envoy-proxy
    namespace: default
```

## Default LoadBalancer Service Type

By default, Envoy uses a Service of type `LoadBalancer`. Depending on your
Kubernetes environment, you might need to add custom annotations. For example,
when deploying in Azure, you can configure the service as follows:

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
      envoyService:
        type: LoadBalancer
        annotations:
          service.beta.kubernetes.io/azure-load-balancer-internal: 'true'
      envoyDeployment:
        container:
          image: custom-cr.internal.io/envoyproxy/envoy:distroless-v1.34.1
```
