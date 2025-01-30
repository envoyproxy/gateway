---
title: "Standalone Deployment Mode"
---

{{% alert title="Notice" color="warning" %}}

Standalone mode is an experimental feature, please **DO NOT** use it in production.

{{% /alert %}}

Envoy Gateway also supports running in standalone mode. In this mode, Envoy Gateway
does not need to rely on Kubernetes and can be deployed directly on bare metal or virtual machines.

Currently, Envoy Gateway only support the file provider and the host infrastructure provider combinations.

- The file provider will configure the Envoy Gateway to get all gateway-api resources from file system.
- The host infrastructure provider will configure the Envoy Gateway to deploy one Envoy Proxy as a host process.

# Quick Start

## Running locally on the host machine

In this quick-start, we will run Envoy Gateway in standalone mode with the file provider
and the host infrastructure provider.

### Prerequisites

Create a local directory just for testing:

```shell
mkdir -p /tmp/envoy-gateway-test
```

As we do not provide the Envoy Gateway binary in latest release,
you can compile this binary on your own from project by using command:

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
extensionApis:
  enableBackend: true
```

As you can see, we have enabled the [Backend][] API, this API will be used to represent our local endpoints.

### Trigger an Update

Any changes under watched `paths` will be considered as an update by the file provider.

For instance, copying example file into `/tmp/envoy-gateway-test/` will trigger an update of gateway-api resources:

```shell
cp examples/standalone/quickstart.yaml /tmp/envoy-gateway-test/quickstart.yaml
```

From the Envoy Gateway log, you should be able to observe that the Envoy Proxy has been started, and its admin address has been returned.

### Test Connection

Starts a simple local server as an endpoint:

```shell
python3 -m http.server 3000
```

Curl the example server through Envoy Proxy:

```shell
curl --verbose --header "Host: www.example.com" http://0.0.0.0:8888/
```

```console
*   Trying 0.0.0.0:8888...
* Connected to 0.0.0.0 (127.0.0.1) port 8888 (#0)
> GET / HTTP/1.1
> Host: www.example.com
> User-Agent: curl/7.81.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< server: SimpleHTTP/0.6 Python/3.10.12
< date: Sat, 26 Oct 2024 13:20:34 GMT
< content-type: text/html; charset=utf-8
< content-length: 1870
<
...
* Connection #0 to host 0.0.0.0 left intact
```

## Running in a Container

In this quick-start, we will run Envoy Gateway in standalone mode with the file provider
and the host infrastructure provider.

### Prerequisites

Create a local directory just for testing:

```shell
mkdir -p /tmp/envoy-gateway-test/config
chmod -R 777 /tmp/envoy-gateway-test
```

Create a container network to run Envoy Gateway and a local server:

```shell
docker network create envoy-gateway-test
```

It's important to widen permissions of a created directory to avoid `Permission denied` error

### Create Certificates

All runners in Envoy Gateway are using TLS connection, so create these TLS certificates locally to
ensure the Envoy Gateway works properly.

```shell
docker run --rm --volume /tmp/envoy-gateway-test:/tmp/envoy-gateway envoyproxy/gateway:{{< helm-version >}} certgen --local
```

### Start Envoy Gateway

The following configuration should be placed into `/tmp/envoy-gateway-test/standalone.yaml` on the host:

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
        paths: ["/tmp/envoy-gateway/config"]
    infrastructure:
      type: Host
      host: {}
logging:
  level:
    default: info
extensionApis:
  enableBackend: true
```

Start Envoy Gateway by the following command:

```shell
$ docker run \
  --name envoy-gateway \
  --network envoy-gateway-test \
  --publish 8888:8888 \
  --volume /tmp/envoy-gateway-test:/tmp/envoy-gateway \
  --detach \
  envoyproxy/gateway:{{< helm-version >}} \
  server --config-path /tmp/envoy-gateway/standalone.yaml
```

As you can see, we have enabled the [Backend][] API, this API will be used to represent our local endpoints.

### Trigger an Update

Any changes under watched `paths` will be considered as an update by the file provider.

For instance, copying example file into `/tmp/envoy-gateway/config` will trigger an update of gateway-api resources:

```shell
cp examples/standalone/quickstart-containers.yaml /tmp/envoy-gateway-test/config/
```

From the Envoy Gateway log, you should be able to observe that the Envoy Proxy has been started, and its admin address has been returned.

### Test Connection

Starts a simple local server in a same container network:

```shell
$ docker run \
  --name local-server \
  --hostname local-server.local \
  --network envoy-gateway-test \
  --detach \
  python:3 \
  python3 -m http.server 3000
```

The `--hostname` field values is used in `Backend` object of Envoy Gateway as FQDN.
This way there is no need to update `Backend` object if IP address of container changed.

Curl the example server through Envoy Proxy:

```shell
curl --verbose --header "Host: www.example.com" http://0.0.0.0:8888/
```

```console
*   Trying 0.0.0.0:8888...
* Connected to 0.0.0.0 (0.0.0.0) port 8888
* using HTTP/1.x
> GET / HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.11.1
> Accept: */*
>
* Request completely sent off
< HTTP/1.1 200 OK
< server: SimpleHTTP/0.6 Python/3.13.1
< date: Wed, 29 Jan 2025 17:04:11 GMT
< content-type: text/html; charset=utf-8
< content-length: 877
<
...
* Connection #0 to host 0.0.0.0 left intact
```

[Backend]: ../../../api/extension_types#backend
