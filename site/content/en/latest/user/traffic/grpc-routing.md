---
title: "GRPC Routing"
---

The [GRPCRoute][] resource allows users to configure gRPC routing by matching HTTP/2 traffic and forwarding it to backend gRPC servers.
To learn more about gRPC routing, refer to the [Gateway API documentation][].

## Prerequisites

Follow the steps from the [Quickstart](../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Installation

Install the gRPC routing example resources:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/grpc-routing.yaml
```

The manifest installs a [GatewayClass][], [Gateway][], a Deployment, a Service, and a GRPCRoute resource.
The GatewayClass is a cluster-scoped resource that represents a class of Gateways that can be instantiated.

__Note:__ Envoy Gateway is configured by default to manage a GatewayClass with
`controllerName: gateway.envoyproxy.io/gatewayclass-controller`.

## Verification

Check the status of the GatewayClass:

```shell
kubectl get gc --selector=example=grpc-routing
```

The status should reflect "Accepted=True", indicating Envoy Gateway is managing the GatewayClass.

A Gateway represents configuration of infrastructure. When a Gateway is created, [Envoy proxy][] infrastructure is
provisioned or configured by Envoy Gateway. The `gatewayClassName` defines the name of a GatewayClass used by this
Gateway. Check the status of the Gateway:

```shell
kubectl get gateways --selector=example=grpc-routing
```

The status should reflect "Ready=True", indicating the Envoy proxy infrastructure has been provisioned. The status also
provides the address of the Gateway. This address is used later in the guide to test connectivity to proxied backend
services.

Check the status of the GRPCRoute:

```shell
kubectl get grpcroutes --selector=example=grpc-routing -o yaml
```

The status for the GRPCRoute should surface "Accepted=True" and a `parentRef` that references the example Gateway.
The `example-route` matches any traffic for "grpc-example.com" and forwards it to the "yages" Service.

## Testing the Configuration

Before testing GRPC routing to the `yages` backend, get the Gateway's address.

```shell
export GATEWAY_HOST=$(kubectl get gateway/example-gateway -o jsonpath='{.status.addresses[0].value}')
```

Test GRPC routing to the `yages` backend using the [grpcurl][] command.

```shell
grpcurl -plaintext -authority=grpc-example.com ${GATEWAY_HOST}:80 yages.Echo/Ping
```

You should see the below response

```shell
{
  "text": "pong"
}
```

Envoy Gateway also supports [gRPC-Web][] requests for this configuration. The below `curl` command can be used to send a grpc-Web request with over HTTP/2. You should receive the same response seen in the previous command.

The data in the body `AAAAAAA=` is a base64 encoded representation of an empty message (data length 0) that the Ping RPC accepts.

```shell
curl --http2-prior-knowledge -s ${GATEWAY_HOST}:80/yages.Echo/Ping -H 'Host: grpc-example.com'   -H 'Content-Type: application/grpc-web-text'   -H 'Accept: application/grpc-web-text' -XPOST -d'AAAAAAA=' | base64 -d
```

## GRPCRoute Match
The `matches` field can be used to restrict the route to a specific set of requests based on GRPC's service and/or method names.
It supports two match types: `Exact` and `RegularExpression`.

### Exact

`Exact` match is the default match type.

The following example shows how to match a request based on the service and method names for `grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo`,
as well as a match for all services with a method name `Ping` which matches `yages.Echo/Ping` in our deployment.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GRPCRoute
metadata:
  name: yages
  labels:
    example: grpc-routing
spec:
  parentRefs:
    - name: example-gateway
  hostnames:
    - "grpc-example.com"
  rules:
    - matches:
      - method:
          method: ServerReflectionInfo
          service: grpc.reflection.v1alpha.ServerReflection
      - method:
          method: Ping
      backendRefs:
        - group: ""
          kind: Service
          name: yages
          port: 9000
          weight: 1
EOF
```

Verify the GRPCRoute status:

```shell
kubectl get grpcroutes --selector=example=grpc-routing -o yaml
```

Test GRPC routing to the `yages` backend using the [grpcurl][] command.

```shell
grpcurl -plaintext -authority=grpc-example.com ${GATEWAY_HOST}:80 yages.Echo/Ping
```

### RegularExpression

The following example shows how to match a request based on the service and method names
with match type `RegularExpression`. It matches all the services and methods with pattern
`/.*.Echo/Pin.+`, which matches `yages.Echo/Ping` in our deployment.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GRPCRoute
metadata:
  name: yages
  labels:
    example: grpc-routing
spec:
  parentRefs:
    - name: example-gateway
  hostnames:
    - "grpc-example.com"
  rules:
    - matches:
      - method:
          method: ServerReflectionInfo
          service: grpc.reflection.v1alpha.ServerReflection
      - method:
          method: "Pin.+"
          service: ".*.Echo"
          type: RegularExpression
      backendRefs:
        - group: ""
          kind: Service
          name: yages
          port: 9000
          weight: 1
EOF
```

Verify the GRPCRoute status:

```shell
kubectl get grpcroutes --selector=example=grpc-routing -o yaml
```

Test GRPC routing to the `yages` backend using the [grpcurl][] command.

```shell
grpcurl -plaintext -authority=grpc-example.com ${GATEWAY_HOST}:80 yages.Echo/Ping
```

[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute/
[Gateway API documentation]: https://gateway-api.sigs.k8s.io/
[GatewayClass]: https://gateway-api.sigs.k8s.io/api-types/gatewayclass/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[Envoy proxy]: https://www.envoyproxy.io/
[grpcurl]: https://github.com/fullstorydev/grpcurl
[gRPC-Web]: https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md#protocol-differences-vs-grpc-over-http2
