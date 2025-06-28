---
title: "Deploy Envoy Gateway in Air-Gapped Environments"
---

Deploying the Envoy Gateway in an air-gapped environment using a Helm chart requires careful configuration of the `values.yaml` file as well as adjustments when deploying a Gateway resource.

You will need to specify custom image repositories for the following components in the Helm chart:

- Gateway
- Ratelimit

## Gateway – `values.yaml` Configuration

```yaml
deployment:
  envoyGateway:
    image:
      repository: custom-cr.internal.io/envoyproxy/gateway
      tag: v1.4.1
```

## Ratelimit - `values.yaml` Configuration

```yaml
global:
  images:
    ratelimit:
      image: custom-cr.internal.io/envoyproxy/ratelimit:master
```

## Gateway Requires a Custom EnvoyProxy Reference

The Gateway must reference a custom EnvoyProxy resource that explicitly specifies the location of the distroless Envoy container image. Without this, the image will be pulled implicitly from Docker Hub.

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

## Default LoadBalancer Service Type

By default, Envoy uses a Service of type `LoadBalancer`. Depending on your Kubernetes environment, you might need to add custom annotations. For example, when deploying in Azure, you can configure the service as follows:

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
          service.beta.kubernetes.io/azure-load-balancer-internal: "true"
      envoyDeployment:
        container:
          image: custom-cr.internal.io/envoyproxy/envoy:distroless-v1.34.1
```
