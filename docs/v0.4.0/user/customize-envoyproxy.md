# Customize EnvoyProxy

Envoy Gateway provides a [EnvoyProxy][] CRD that can be linked to the ParametersRef
in GatewayClass y cluster admins to customize the managed EnvoyProxy Deployment and
Service. To learn more about GatewayClass and ParametersRef, please refer to [Gateway API documentation][].

## Installation

Follow the steps from the [Quickstart Guide](quickstart.md) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Add GatewayClass ParametersRef

First, you need to add ParametersRef in GatewayClass, and refer to EnvoyProxy Config:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: config.gateway.envoyproxy.io
    kind: EnvoyProxy
    name: custom-proxy-config
    namespace: envoy-gateway-system
EOF
```

## Customize EnvoyProxy Deployment Replicas

You can customize the EnvoyProxy Deployment Replicas via EnvoyProxy Config like:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        replicas: 2
EOF
```

After you apply the config, you should see the replicas of envoyproxy changes to 2.
And also you can dynamically change the value.

``` shell
kubectl get deployment envoy-gateway
```

## Customize EnvoyProxy Image

You can customize the EnvoyProxy Image via EnvoyProxy Config like:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          image: envoyproxy/envoy:v1.25-v0.4.0
EOF
```

After applying the config, you can get the deployment image, and see it has changed.

## Customize EnvoyProxy Pod Annotations

You can customize the EnvoyProxy Pod Annotations via EnvoyProxy Config like:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        pod:
          annotations:
            custom1: deploy-annotation1
            custom2: deploy-annotation2
EOF
```

After applying the config, you can get the envoyproxy pods, and see new annotations has been added.

## Customize EnvoyProxy Deployment Resources

You can customize the EnvoyProxy Deployment Resources via EnvoyProxy Config like:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          resources:
            requests:
              cpu: 150m
              memory: 640Mi
            limits:
              cpu: 500m
              memory: 1Gi
EOF
```

After applying the config, you can get the envoyproxy deployment, and see resources has been changed.

## Customize EnvoyProxy Service Annotations

You can customize the EnvoyProxy Service Annotations via EnvoyProxy Config like:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyService:
        annotations:
          custom1: svc-annotation1
          custom2: svc-annotation2

EOF
```

After applying the config, you can get the envoyproxy service, and see annotations has been added.

## Customize EnvoyProxy Bootstrap Config

You can customize the EnvoyProxy Bootstrap Config via EnvoyProxy Config like:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  bootstrap: |
    xxxxxxxxxx

EOF
```

After applying the config, the bootstrap config will be overridden by the new config you provided.
Envoy Gateway will use Webhook to validate the bootstrap config you provided.

[Gateway API documentation]: https://gateway-api.sigs.k8s.io/
[EnvoyProxy]: https://www.envoyproxy.io/
