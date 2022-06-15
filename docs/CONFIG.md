## Introduction
Envoy Gateway consists of two types of configuration, static and dynamic. Static is used to configure Envoy Gateway at
runtime and dynamic configuration is used to manage the Envoy proxies. This document provides Envoy Gateway
configuration details.

## Static Config
This is the configuration used to configure various internal aspects of Envoy Gateway at runtime.

#### Configuration File
The configuration file is defined by the ControlPlaneSpec API type. At startup, Envoy Gateway searches for the
configuration at "/etc/envoy-gateway/config.yaml". You can override the path to this file with `-c <path string>` or
`--config-path <path string>`. For example:
```shell
envoy-gateway start -c config.yaml
```

#### CLI Arguments
If no value is provided for a given option, a default value applies. If an option has sub-options, and any of these
sub-options are not specified, a default value will also apply. For example, the `--providers.kubernetes` option will
enable the Kubernetes provider, even though sub-options may exist. Once specified, this option sets/resets all the
default values of the `--providers.kubernetes` sub-options.

#### Environment Variables
All static configuration is available through environment variables. The environment variable format is- uppercase the
CLI argument, replace `-`  with `_`, and append `EG_`. For example, the `config-path` CLI argument is represented as
`EG_CONFIG_PATH`. This format applies to all CLI arguments.

## Dynamic Config
Dynamic configuration manages the data plane through [Gateway API][gw_api] objects, e.g. Gateway, HTTPRoute, etc.

### Using the Kubernetes Provider:
Install Envoy Gateway and the data plane:
```shell
$ kubectl apply -f ./examples/kubernetes/quickstart.yaml
```
This all-in-one manifest installs all the necessary resources to run Envoy Gateway and the managed Envoy proxies in a
Kubernetes cluster. Here is what the GatewayClass and Gateway resources look like from this manifest:
```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GatewayClass
metadata:
   name: example-class
spec:
   controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: Gateway
metadata:
   name: example-gateway
   namespace: envoy-gateway-system
spec:
   gatewayClassName: example-class
   listeners:
      - name: http
        protocol: HTTP
        port: 8080
```

Envoy Gateway should now be running with default configuration parameters. A config file, env vars, or CLI args could
have been passed to the Envoy Gateway Deployment to change its default runtime behavior. For example, passing
`--networkPublishing.type=LoadBalancer` causes Envoy Gateway to be exposed by a Kubernetes LoadBalancer service instead
of the default ClusterIP service. When Envoy Gateway starts, it creates the required Kubernetes resources to provision
the managed Envoy proxies dynamically by reconciling the GatewayClass and Gateway resources included in the all-in-one
manifest.

All Envoy proxies are listening on port 8080 but are not routing any traffic to backend services until an HTTPRoute is
created that references the Gateway. Since the GatewayClass did not specify a `parametersRef`, the Envoy proxies are
configured with default parameters. A `GatewayClassParams` Custom Resource Definition (CRD) could be been created and
referenced by the GatewayClass to configure the managed Envoy proxy infrastructure. For example, to have the proxies use
a Kubernetes NodePort service instead of the default LoadBalancer service:
```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GatewayClass
metadata:
   name: example-class
spec:
   controllerName: gateway.envoyproxy.io/gatewayclass-controller
   parametersRef:
      name: example-params
      group: gateway.envoyproxy.io
      kind: GatewayClassParams
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: GatewayClassParams
metadata:
   name: example-params
   namespace: envoy-gateway-system
spec:
  dataplane:
    networkPublishing:
      type: NodePortService
```
When additional Gateways are created that reference the same GatewayClass, Envoy Gateway will merge the config from
each Gateway and push it to the proxies. If another control plane is desired, the user repeats the same process above
but with the modifications to the all-in-one manifest:
- Update the Namespace from `envoy-gateway-system` to a namespace of your choosing, e.g. `my-ns`.
- Update the namespace of all namespaced resources to match the newly created Namespace. 
- Change the GatewayClass `controllerName` value, e.g. `gateway.envoyproxy.io/my-ns/gatewayclass-controller`.
- Start the new control plane with a matching `className`. For example:
  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
  name: my-eg
  namespace: my-ns
  ...
    containers:
      - args:
        - start
        - --className=gateway.envoyproxy.io/my-ns/gatewayclass-controller
  ...
   ```

To perform an in-place upgrade, update the image in the Envoy Gateway deployment resource. This causes the control plane
to update the image used by the managed proxies.

[gw_api]: https://gateway-api.sigs.k8s.io
