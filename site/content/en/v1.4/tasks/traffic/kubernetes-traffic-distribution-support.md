---
title: "Kubernetes Traffic Distribution Support"
---

EnvoyGateway supports Kubernetes [Traffic Distribution][Traffic Distribution] and [Topology Aware Routing][Topology Aware Routing] 
which is useful for keeping network traffic in the originating zone. Preferring same-zone traffic between Pods in your 
cluster can help with reliability, performance (network latency and throughput), or cost.

{{% alert title="Note" color="primary" %}}
The current implementation for Topology Aware Routing support doesn't respect the minimum of 3 endpoints per zone
requirement and applies the same logic as `spec.TrafficDistribution=PreferClose`. This will be fixed in the next release.
{{% /alert %}}


## Prerequisites
* The Kubernetes cluster's nodes must indicate topology information via the `topology.kubernetes.io/zone` [well-known label][Kubernetes well-known metadata].
* There must be at least two valid topology zones for scheduling.
* {{< boilerplate prerequisites >}}

## Configuration

Apply the example manifests with zonal based routing enabled:
```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/latest/zone-routing.yaml -n default
```

Ensure that both example Deployments are marked as ready and produces the following output:
```shell
kubectl get deployment/zone-routing-backend-local deployment/zone-routing-backend-nonlocal -n default
NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
zone-routing-backend-local      1/1     1            1           9m1s
zone-routing-backend-nonlocal   1/1     1            1           9m1s

```

An HTTPRoute resource is created for the `/zone-routing` path prefix along with two example Deployment resources that 
are respectively configured with pod affinity/anti-affinity targeting the Envoy Proxy Pods for testing.

Verify the HTTPRoute configuration and status:

```shell
kubectl get httproute/zone-routing -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

### HTTPRoute

We will now send a request and expect to be routed to backend in the same local zone as the Envoy Proxy Pods.

```shell
curl -H "Host: www.example.com" "http://${GATEWAY_HOST}/zone-routing" -XPOST -d '{}'
```

We will see the following output. The `pod` header identifies which backend received the request and should have
a name prefix of `zone-routing-backend-local-`. We should see this behavior every time and the nonlocal backend should
not receive requests.

```
{
  "path": "/zone-routing",
  "host": "www.example.com",
  "method": "GET",
  "proto": "HTTP/1.1",
  "headers": {
    "Accept": [
      "*/*"
    ],
    "User-Agent": [
      "curl/8.7.1"
    ],
    "X-Envoy-External-Address": [
      "127.0.0.1"
    ],
    "X-Forwarded-For": [
      "10.244.1.5"
    ],
    "X-Forwarded-Proto": [
      "http"
    ],
    "X-Request-Id": [
      "0f4c5d28-52d0-4727-881f-abf2ac12b3b7"
    ]
  },
  "namespace": "default",
  "ingress": "",
  "service": "",
  "pod": "zone-routing-backend-local-5586cd668d-md8sf"
}
```


## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the zone-routing example resources BackendTrafficPolicy and HTTPRoute:

```shell
kubectl delete service/zone-routing-backend 
kubectl delete deployment/zone-routing-backend-local
kubectl delete deployment/zone-routing-backend-nonlocal
kubectl delete httproute/zone-routing
```

[Traffic Distribution]: https://kubernetes.io/docs/concepts/services-networking/service/#traffic-distribution
[Topology Aware Routing]: https://kubernetes.io/docs/concepts/services-networking/topology-aware-routing/
[Kubernetes well-known metadata]: https://kubernetes.io/docs/reference/labels-annotations-taints/#topologykubernetesiozone