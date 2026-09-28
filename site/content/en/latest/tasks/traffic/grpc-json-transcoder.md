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

The key `proto-descriptor` is used when present; a `binaryData` entry is also found when it is the ConfigMap's only one.
A `data` entry is accepted too, but it must be base64-encoded and must use the `proto-descriptor` key.

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
JSON traffic:

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

## Policies and routing

The transcoder rewrites `:path` to the gRPC method (`/example.Greeter/SayHello`), but the request stays on the HTTPRoute
rule that matched the JSON path; the route table is not matched again. Everything that governs the request — backend,
timeouts, retries, and every policy — comes from that HTTPRoute. A [GRPCRoute][] for the same service on the same
hostname serves native gRPC clients only, and its policies do not apply to transcoded requests. Attach authentication,
authorization, and rate limiting to the HTTPRoute that receives the JSON traffic.

That holds for the default filter order. A filter that `EnvoyProxy.spec.filterOrder` moves after the transcoder can
clear the route cache — ExtAuth or JWT with `recomputeRoute`, an ExtProc server whose response sets
`clear_route_cache`, or a Lua, Wasm, or dynamic module extension — and Envoy then matches the gRPC method path again,
possibly onto a GRPCRoute whose policies never ran.

Anything the router derives from `:path` sees the transcoded gRPC method instead of the JSON path, so these cannot share
a rule with the transcoder: a `URLRewrite` that modifies the path, a `URLRewrite` with a `PathRegex` hostname, and a
`RequestRedirect`. Such a rule is rejected with `IncompatibleFilters` on the HTTPRoute's `Accepted` condition and returns
`500`. Literal, header, and backend hostname rewrites are unaffected.

The transcoded path also carries no query string, so a `BackendTrafficPolicy` that hashes on query parameters
(`loadBalancer.consistentHash.type: QueryParams`) finds none on these routes and spreads requests as if unhashed. It is
not rejected.

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
