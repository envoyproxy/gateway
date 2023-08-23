# Proxy Enhance

Envoy Gateway provides basic system debugging for the ControlPlane and the underlying EnvoyProxy instances.
This guide show you how to get a core file for debugging in proxy.

## Prerequisites

Follow the steps from the [Quickstart Guide](quickstart.md) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

### Add GatewayClass ParametersRef
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

### Custommize EnvoyProxy CoreDump
You can customize the EnvoyProxy CoreDump via EnvoyProxy Config like:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  enableCoreDump: true
EOF
```

After you apply the config,you will find the generated core file in the local `/var/gateway/proxy/data/` directory when envoy proxy crashes.


## Debug coredump file
### GDB
