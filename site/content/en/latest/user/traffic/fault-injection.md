---
title: "Fault Injection"
---

[Envoy fault injection] can be used to inject delays and abort requests to mimic failure scenarios such as service failures and overloads.

Envoy Gateway supports the following fault scenarios:
- **delay fault**: inject a custom fixed delay into the request with a certain probability to simulate delay failures.
- **abort fault**: inject a custom response code into the response with a certain probability to simulate abort failures.

Envoy Gateway introduces a new CRD called [BackendTrafficPolicy][] that allows the user to describe their desired fault scenarios.
This instantiated resource can be linked to a [Gateway][], [HTTPRoute][] or [GRPCRoute][] resource.

## Prerequisites

Follow the steps from the [Quickstart](../../quickstart) guide to install Envoy Gateway and the example manifest.
For GRPC - follow the steps from the [GRPC Routing](../grpc-routing) example.
Before proceeding, you should be able to query the example backend using HTTP or GRPC.

### Install the hey load testing tool
* The `hey` CLI will be used to generate load and measure response times. Follow the installation instruction from the [Hey project] docs.

## Configuration

Allow requests with a valid faultInjection by creating an [BackendTrafficPolicy][BackendTrafficPolicy] and attaching it to the example HTTPRoute or GRPCRoute.

### HTTPRoute

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: fault-injection-50-percent-abort
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: foo
  faultInjection:
    abort:
      httpStatus: 501
      percentage: 50
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: fault-injection-delay
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: bar
  faultInjection:
    delay:
      fixedDelay: 2s
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: foo
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
      weight: 1
    matches:
    - path:
        type: PathPrefix
        value: /foo
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: bar
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
      weight: 1
    matches:
    - path:
        type: PathPrefix
        value: /bar

EOF
```

Two HTTPRoute has been created, one for `/foo` and another for `/bar`.  `fault-injection-abort` BackendTrafficPolicy has been created and targeted HTTPRoute foo to abort requests for `/foo`. `fault-injection-delay` BackendTrafficPolicy has been created and targeted HTTPRoute foo to delay `2s` requests for `/bar`. 

Verify the HTTPRoute configuration and status:

```shell
kubectl get httproute/foo -o yaml
kubectl get httproute/bar -o yaml
```

Verify the BackendTrafficPolicy configuration:

```shell
kubectl get backendtrafficpolicy/fault-injection-50-percent-abort -o yaml
kubectl get backendtrafficpolicy/fault-injection-delay -o yaml
```

### GRPCRoute

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: fault-injection-abort
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: GRPCRoute
    name: yages
  faultInjection:
    abort:
      grpcStatus: 14
---
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
  - backendRefs:
    - group: ""
      kind: Service
      name: yages
      port: 9000
      weight: 1
EOF
```

A BackendTrafficPolicy has been created and targeted GRPCRoute yages to abort requests for `yages` service..

Verify the GRPCRoute configuration and status:

```shell
kubectl get grpcroute/yages -o yaml
```

Verify the SecurityPolicy configuration:

```shell
kubectl get backendtrafficpolicy/fault-injection-abort -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) guide is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

### HTTPRoute

Verify that requests to `foo` route are aborted.

```shell
hey -n 1000 -c 100 -host "www.example.com"  http://${GATEWAY_HOST}/foo
```

```console
Status code distribution:
  [200]	501 responses
  [501]	499 responses
```

Verify that requests to `bar` route are delayed.

```shell
hey -n 1000 -c 100 -host "www.example.com"  http://${GATEWAY_HOST}/bar
```

```console
Summary:
  Total:	20.1493 secs
  Slowest:	2.1020 secs
  Fastest:	1.9940 secs
  Average:	2.0123 secs
  Requests/sec:	49.6295

  Total data:	557000 bytes
  Size/request:	557 bytes

Response time histogram:
  1.994 [1]	|
  2.005 [475]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  2.016 [419]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  2.026 [5]	|
  2.037 [0]	|
  2.048 [0]	|
  2.059 [30]	|■■■
  2.070 [0]	|
  2.080 [0]	|
  2.091 [11]	|■
  2.102 [59]	|■■■■■
```

### GRPCRoute

Verify that requests to `yages`service are aborted.

```shell
grpcurl -plaintext -authority=grpc-example.com ${GATEWAY_HOST}:80 yages.Echo/Ping
```

You should see the below response

```shell
Error invoking method "yages.Echo/Ping": rpc error: code = Unavailable desc = failed to query for service descriptor "yages.Echo": fault filter abort
```

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) guide to uninstall Envoy Gateway and the example manifest.

Delete the BackendTrafficPolicy:

```shell
kubectl delete BackendTrafficPolicy/fault-injection-abort
```

[Envoy fault injection]: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/fault_filter.html
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute/
[Hey project]: https://github.com/rakyll/hey
