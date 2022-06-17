## Introduction

Envoy Gateway consists of two types of configuration, static and dynamic. Static is used to configure an Envoy Gateway
component at runtime and dynamic configuration is used to manage the proxy infrastructure and route traffic. This
document details the installation and configuration workflow for the two options described in [Issue #79][79]. Note that
both options assume the Kubernetes provider is being used.

## Option A (Operator):

A user installs Envoy Gateway:
```shell
$ kubectl apply -f quickstart.yaml
```

This all-in-one manifest installs all the necessary resources to run the operator and control/data planes, e.g.
Deployment, RBAC, etc. Since the all-in-one manifest includes GatewayClass and Gateway resources, the operator will
read these resources and instantiate control/data planes after it starts-up. The following are the GatewayClass and
Gateway resources from the all-in-one manifest:
```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GatewayClass
metadata:
   name: eg-class
spec:
   controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: Gateway
metadata:
   name: eg-gw
   namespace: default-ns
spec:
   gatewayClassName: eg-class
   listeners:
      - name: http
        protocol: HTTP
        port: 8080
```

The control and data planes have been instantiated and are now running as separate pods. All data plane instances are
listening on port 8080 but are not routing any traffic to backend services until an HTTPRoute is created that references
the Gateway. Since the GatewayClass did not specify a `parametersRef`, the control and data planes were configured with
default parameters. The user could have created an instance of the `GatewayClassParams` Custom Resource Definition (CRD)
and reference it from the GatewayClass. The `GatewayClassParams` resource is used to configure control and data plane
runtime parameters. For example, to have the control and data planes use a Kubernetes NodePort service instead of the
default LoadBalancer service:
```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GatewayClass
metadata:
   name: eg-class
spec:
   controllerName: gateway.envoyproxy.io/gatewayclass-controller
   parametersRef:
      name: eg-params
      group: gateway.envoyproxy.io
      kind: GatewayClassParams
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: GatewayClassParams
metadata:
   name: eg-params
   namespace: default-ns
spec:
   controlplane:
      networkPublishing:
         type: NodePortService
   dataplane:
      networkPublishing:
         type: NodePortService
```

When a user creates additional Gateways referencing the same GatewayClass, the control plane will merge the config from
each Gateway and push it to the data plane. If another control plane is desired, the user repeats the same process above
but changes the value of `gatewayclass.spec.controllerName`. The operator will instantiate new control/data planes and
configure the control plane with a matching `controllerName`. This will allow the new control plane to reconcile
Gateways and begin controlling the data plane.

To perform an in-place upgrade, the user simply upgrades the operator. The operator is encoded with the intelligence to
perform the necessary steps to perform a graceful upgrade of control and data planes.

## Option B (Manually):

A user installs the control and data planes:
```shell
$ kubectl apply -f eg-control-and-data-planes.yaml
```
This all-in-one manifest installs all the necessary resources to run the control and data planes, e.g. GatewayClass,
Gateway, Deployment (control plane), RBAC, etc. Here is what the GatewayClass and Gateway resources look like from
`eg-control-and-data-planes.yaml`:
```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GatewayClass
metadata:
   name: eg-class
spec:
   controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: Gateway
metadata:
   name: eg-gw
   namespace: default-ns
spec:
   gatewayClassName: eg-class
   listeners:
      - name: http
        protocol: HTTP
        port: 8080
```

The control plane has been instantiated and is now running with default configuration parameters. A config file, env
vars, or CLI args could have been passed to the control plane Deployment to change its default runtime behavior, e.g.
`--networkPublishing.type=NodePortService`. When the control plane starts, it creates the data plane dynamically by
reconciling the GatewayClass and Gateway resources that were included in the all-in-one manifest.

All data plane instances are listening on port 8080 but are not routing any traffic to backend services until an
HTTPRoute is created that references the Gateway. Since the GatewayClass did not specify a `parametersRef`, the data
plane is configured with default parameters. The user could have created an instance of the `GatewayClassParams` Custom
Resource Definition (CRD) and reference it from the GatewayClass. The `GatewayClassParams` resource is used to configure
control and data plane parameters. For example, to have the data plane use a Kubernetes NodePort service instead of the
default LoadBalancer service:
```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GatewayClass
metadata:
   name: eg-class
spec:
   controllerName: gateway.envoyproxy.io/gatewayclass-controller
   parametersRef:
      name: data-plane-params
      group: gateway.envoyproxy.io
      kind: GatewayClassParams
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: GatewayClassParams
metadata:
   name: data-plane-params
   namespace: default-ns
spec:
  dataplane:
    networkPublishing:
      type: NodePortService
```
__Note__: See the [open questions](#open-questions) section on whether the control plane should support dynamic
config/reconfig through GatewayClassParams.

When a user creates additional Gateways referencing the same GatewayClass, the control plane will merge the config from
each Gateway and push it to the data plane. If another control plane is desired, the user repeats the same process above
but with a few modifications to the all-in-one manifest:
- Change the value of `gatewayclass.spec.controllerName`. For example:
  ```
  controllerName: gateway.envoyproxy.io/eg-ns-2/gatewayclass-controller
  ```
- Start the new control plane with a matching `className`. For example:
  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
  name: eg-controlplane-2
  namespace: eg-ns-2
  ...
    containers:
      - args:
        - start
        - --className=gateway.envoyproxy.io/eg-ns-2/gatewayclass-controller
  ...
   ```

To perform an in-place upgrade, the user updates the image of the control plane. This causes the control plane to update
the image of the managed data plane deployment. Optionally, the user can transition to using the Gateway Operator to
manage the control and data planes. The operator is encoded with the intelligence to perform the necessary steps to
perform a graceful upgrade of control and data planes.

## Option B (Operator):

A user installs Envoy Gateway:
```shell
$ kubectl apply -f quickstart.yaml
```

This all-in-one manifest installs all the necessary resources to run the operator and control/data planes, e.g.
Deployment, RBAC, etc.

To perform an in-place upgrade, the user simply upgrades the operator. The operator is encoded with the intelligence to
perform the necessary steps to perform a graceful upgrade of the control plane. When the control plane is upgraded by
the operator, the control plane upgrades the data plane.

### Key differences from [Option A](#option-a)
- The control plane will instantiate the data plane instead of the operator performing this task.
- For upgrades, the operator upgrades the control plane and the control plane upgrades the data plane.

## Open Questions:
- For Option B, should `gatewayclassparams.spec.controlplane` be ignored so the control plane can only be configured by
  CLI args, env vars, and conf file, or should it support dynamic config/reconfig by reading these values and having
  them take precedence over the static config options? If dynamic control plane config/reconfig should not be supported,
  then the control plane will ignore `gatewayclassparams.spec.controlplane`.
- For Option B, how can the control plane rollback, recover, etc. from a failed in-place upgrade? The control plane
  could be left in a state where it can no longer manage the data plane.

[79]: https://github.com/envoyproxy/gateway/issues/97
