---
title: "Standalone Deployment Mode"
---

{{% alert title="Notice" color="warning" %}}

Standalone mode is an experimental feature, please **DO NOT** use it in production.

{{% /alert %}}

Envoy Gateway also supports running in standalone mode. In this mode, you can configure the
Envoy Gateway to use different file provider and infrastructure provider combinations.

- The file provider will configure the Envoy Gateway where to get all gateway-api resources.
- The infrastructure provider will configure the Envoy Gateway how to deploy the data-plane components.

Currently, the types supported by file provider are:

- **File**, retrieve gateway-api resources from file system.

The infrastructure provider supports the following types:

- **Host**, run data-plane component as a host process.

## Quick Start

In this quick-start, we will run Envoy Gateway in standalone mode with **File** type file provider
and **Host** type infrastructure provider.

### Prerequisites

Create a local directory just for testing:

```shell
mkdir -p /tmp/envoy-gateway-test
```

Compile Envoy Gateway binary from project by using command:

```shell
make build
```

The compiled binary lies in `bin/{os}/{arch}/envoy-gateway`.

### Create Certificates

All runners in Envoy Gateway are using TLS connection, so create these TLS certificates locally to 
ensure the Envoy Gateway works properly.

```shell
envoy-gateway certgen --local
```

### Start Envoy Gateway

Start Envoy Gateway by the following command:

```shell
envoy-gateway server --config-path standalone.yaml
```

with `standalone.yaml` configuration:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
provider:
  type: Custom
  custom:
    resource:
      type: File
      file:
        paths: ["/tmp/envoy-gateway-test"]
    infrastructure:
      type: Host
      host: {}
logging:
  level:
    default: info
```

Update gateway-api resources:

```shell
cp examples/standalone/quickstart.yaml /tmp/envoy-gateway-test/quickstart.yaml
```

From Envoy Gateway log, you should be able to observe a Envoy Proxy has been started.

### Test Connection

Starts a simple local server as an upstream service:

```shell
python3 -m http.server 3000
```
