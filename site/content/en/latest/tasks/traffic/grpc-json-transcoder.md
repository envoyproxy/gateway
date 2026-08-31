---
title: "gRPC-JSON Transcoding"
---

This task shows how to use [HTTPRouteFilter][HTTPRouteFilter] to transcode JSON/HTTP requests into gRPC calls, so
REST clients can talk to a gRPC backend without a separate gateway service.

Envoy derives the mapping from the `google.api.http` options in your protobuf definitions, so the JSON paths a client
calls are the ones already declared in the `.proto`.

The filter is only supported at the rule level of an [HTTPRoute][HTTPRoute]. Referencing it from a [GRPCRoute][] or from
a `backendRef` makes the route unresolvable: the incoming request has to be JSON/HTTP for there to be anything to
transcode, and a `backendRef` filter has no route table on which to enable the transcoder.

## Prerequisites

{{< boilerplate prerequisites >}}

## Generate the proto descriptor

The transcoder needs a binary `FileDescriptorSet` describing your services. Generate it with `--include_imports`, which
bundles the files your protos import — without them Envoy cannot build a descriptor pool and rejects the configuration
with only `Unable to build proto descriptor pool` to go on.

```shell
protoc --include_imports --descriptor_set_out=proto-descriptor.pb path/to/your.proto
```

Store it in a ConfigMap in the same namespace as the HTTPRoute. `--from-file` puts the bytes in `binaryData`, which the
transcoder reads as-is:

```shell
kubectl create configmap greeter-proto-descriptor --from-file=proto-descriptor=proto-descriptor.pb
```

The key `proto-descriptor` is used when present; otherwise the ConfigMap must hold exactly one entry. A `data` entry is
also accepted, but it must be base64-encoded.

## Reach the backend over HTTP/2

What leaves the transcoder is gRPC, which requires HTTP/2 upstream. On an HTTPRoute the upstream protocol defaults to
HTTP/1.1, so the backend has to declare HTTP/2 explicitly — either `appProtocol: kubernetes.io/h2c` on the Service port,
or `appProtocols: [gateway.envoyproxy.io/h2c]` on a [Backend][]. Nothing in the route status reports the mismatch: the
route stays `Accepted` and every transcoded request fails at runtime.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: grpc-service
spec:
  selector:
    app: grpc-service
  ports:
  - name: grpc
    protocol: TCP
    port: 9000
    targetPort: 9000
    appProtocol: kubernetes.io/h2c
```

## Configuration

Create an `HTTPRouteFilter` that points at the ConfigMap, and reference it from the HTTPRoute rule that receives the
JSON traffic. `matchIncomingRequestRoute: true` keeps the request on the route that matched it, so one rule is enough:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: HTTPRouteFilter
metadata:
  name: grpc-transcoder
spec:
  grpcJSONTranscoder:
    protoDescriptor:
      valueRef:
        group: ""
        kind: ConfigMap
        name: greeter-proto-descriptor
    services:
    - example.Greeter
    matchIncomingRequestRoute: true
```

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: grpc-route
spec:
  parentRefs:
  - name: eg
  rules:
  # The JSON path clients call, from the google.api.http option on SayHello.
  - matches:
    - path:
        type: PathPrefix
        value: /v1/hello
    filters:
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: HTTPRouteFilter
        name: grpc-transcoder
    backendRefs:
    - name: grpc-service
      port: 9000
```

Then call the JSON path:

```shell
curl -H "Host: grpc.example.com" http://${GATEWAY_HOST}/v1/hello/world
```

When `services` is omitted, every service declared by the descriptor's own proto files is transcoded, excluding services
that come from imported files. Naming them explicitly is worth doing when the descriptor carries more than you want to
expose.

## Routing the rewritten path separately

`matchIncomingRequestRoute` defaults to `false`, which is Envoy's own default. In that mode the transcoder rewrites
`:path` to the gRPC method (`/example.Greeter/SayHello`) and Envoy matches the routing table again, so a route for the
rewritten path must exist or the request gets a 404.

Leave it unset when you want the gRPC method to route on its own terms — a different backend, a different timeout, or
policies attached to a [GRPCRoute][]. Either a second HTTPRoute rule or a GRPCRoute for the service satisfies the
re-match:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: grpc-route
spec:
  parentRefs:
  - name: eg
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /v1/hello
    filters:
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: HTTPRouteFilter
        name: grpc-transcoder
    backendRefs:
    - name: grpc-service
      port: 9000
  # Serves the rewritten path. A GRPCRoute for example.Greeter works here too.
  - matches:
    - path:
        type: PathPrefix
        value: /example.Greeter
    backendRefs:
    - name: grpc-service
      port: 9000
```

Only the rule receiving the JSON request needs the filter — the response is still transcoded back to JSON even when the
re-match lands on a route with no transcoder config, because Envoy resolves filter enablement once, against the route
matched at `decodeHeaders`.

## Diagnosing a bad descriptor

The descriptor is parsed and validated when the route is translated, not when Envoy loads the listener. A descriptor that
is malformed, missing its imports, or does not declare a service you named in `services` makes the rule return `500` and
records the reason on the HTTPRoute's `Accepted` condition:

```shell
kubectl get httproute grpc-route -o yaml
```

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

```shell
kubectl delete httproutefilter/grpc-transcoder
kubectl delete configmap/greeter-proto-descriptor
```

## Next Steps

Check out the [Developer Guide](/community/develop) to get involved in the project.

[HTTPRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/grpcroute/
[HTTPRouteFilter]: ../../../api/extension_types#httproutefilter
[Backend]: ../../../api/extension_types#backend
