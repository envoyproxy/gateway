---
title: "Zone Aware Routing"
---

EnvoyGateway makes use of [Envoy Zone Aware Routing][Envoy Zone Aware Routing] to keep network traffic in the originating zone.
Preferring same-zone traffic between Pods in your cluster can help with reliability, performance (network latency and throughput), or cost.

Zone-aware routing may be enabled in one of two ways:
1. Configuring a [BackendTrafficPolicy][BackendTrafficPolicy] with the `loadbalancer.zoneAware` field
2. Configuring a backendRef Kubernetes `Service` with  [Traffic Distribution][Traffic Distribution] or [Topology Aware Routing][Topology Aware Routing]

When both a backendRef and a [BackendTrafficPolicy][BackendTrafficPolicy] include a configuration for zone awareness, the [BackendTrafficPolicy][BackendTrafficPolicy] takes precedence.

## Prerequisites
* The Kubernetes cluster's nodes must indicate topology information via the `topology.kubernetes.io/zone` [well-known label][Kubernetes well-known metadata].
* There must be at least two valid topology zones for scheduling.
* {{< boilerplate prerequisites >}}

## Configuration

Choose one of the following configuration options.

### Option 1: Kubernetes Service
Create the example Kubernetes Service with either topology aware routing or traffic distribution enabled.

#### Topology Aware Routing
{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}
```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.kubernetes.io/topology-mode: Auto
  name: zone-aware-routing-backend
  labels:
    app: zone-aware-routing-backend
    service: zone-aware-routing-backend
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: zone-aware-routing-backend
EOF
```
{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the below service manifest to your cluster:
```yaml
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.kubernetes.io/topology-mode: Auto
  name: zone-aware-routing-backend
  labels:
    app: zone-aware-routing-backend
    service: zone-aware-routing-backend
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: zone-aware-routing-backend
```
{{% /tab %}}
{{< /tabpane >}}

#### Traffic Distribution
{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}
```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: zone-aware-routing-backend
  labels:
    app: zone-aware-routing-backend
    service: zone-aware-routing-backend
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: zone-aware-routing-backend
  trafficDistribution: PreferClose
EOF
```
{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the below service manifest to your cluster:
```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: zone-aware-routing-backend
  labels:
    app: zone-aware-routing-backend
    service: zone-aware-routing-backend
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: zone-aware-routing-backend
  trafficDistribution: PreferClose
```
{{% /tab %}}
{{< /tabpane >}}

### Option 2: BackendTrafficPolicy
Zone aware routing can also be enabled directly with a [BackendTrafficPolicy][BackendTrafficPolicy].
The example below configures similar behavior to Kubernetes Traffic Distribution and forces all traffic to the local zone via the `force` field instead
of Envoy's default behavior which _prefers_ routing locally as much as possible while still achieving overall equal request distribution across all endpoints.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}
```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: zone-aware-routing
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: zone-aware-routing
  loadBalancer:
    type: RoundRobin
    zoneAware:
      preferLocal: # Enables Envoy to prefer local zone endpoints while maintaining overall traffic balance across zones
        minEndpointsThreshold: 1 # Zone-aware routing is disabled if total number of endpoints is less than this threshold
        force: # Forces all traffic to stay within the local zone, regardless of upstream endpoint distribution between zones
          minEndpointsInZoneThreshold: 1  # If fewer local endpoints exist than this threshold, fallback to standard zone-aware routing behavior
EOF
```
{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:
```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: zone-aware-routing
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: zone-aware-routing
  loadBalancer:
    type: RoundRobin
    zoneAware:
      preferLocal: # Enables Envoy to prefer local zone endpoints while maintaining overall traffic balance across zones
        minEndpointsThreshold: 1 # Zone-aware routing is disabled if total number of endpoints is less than this threshold
        force: # Forces all traffic to stay within the local zone, regardless of upstream endpoint distribution between zones
          minEndpointsInZoneThreshold: 1  # If fewer local endpoints exist than this threshold, fallback to standard zone-aware routing behavior
```
{{% /tab %}}
{{< /tabpane >}}


### Example deployments and HTTPRoute
Next apply the example manifests to create two Deployments and an HTTPRoute. For the test configuration one Deployment
(zone-aware-routing-backend-local) includes affinity for EnvoyProxy Pods to ensure its Pods are scheduled in the same
zone and the second Deployment (zone-aware-routing-backend-nonlocal) uses anti-affinity to ensure its Pods _don't_
schedule to the same zone in order to demonstrate functionality.
```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/zone-aware-routing.yaml -n default
```

Ensure that both example Deployments are marked as ready and produces the following output:
```shell
kubectl get deployment/zone-aware-routing-backend-local deployment/zone-aware-routing-backend-nonlocal -n default
NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
zone-aware-routing-backend-local      3/3     3            3           9m1s
zone-aware-routing-backend-nonlocal   3/3     3            3           9m1s

```

An HTTPRoute resource is created for the `/zone-aware-routing` path prefix along with two example Deployment resources that
are respectively configured with pod affinity/anti-affinity targeting the Envoy Proxy Pods for testing.

Verify the HTTPRoute configuration and status:

```shell
kubectl get httproute/zone-aware-routing -o yaml
```

If used during configuration, verify the [BackendTrafficPolicy][BackendTrafficPolicy]:

```shell
kubectl get backendtrafficpolicy/zone-aware-routing -o yaml
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
curl -H "Host: www.example.com" "http://${GATEWAY_HOST}/zone-aware-routing" -XPOST -d '{}'
```

We will see the following output. The `pod` header identifies which backend received the request and should have
a name prefix of `zone-aware-routing-backend-local-`. We should see this behavior every time and the nonlocal backend should
not receive requests.

```
{
  "path": "/zone-aware-routing",
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
  "pod": "zone-aware-routing-backend-local-5586cd668d-md8sf"
}
```


## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the zone-aware-routing example resources and HTTPRoute:

```shell
kubectl delete service/zone-aware-routing-backend 
kubectl delete deployment/zone-aware-routing-backend-local
kubectl delete deployment/zone-aware-routing-backend-nonlocal
kubectl delete httproute/zone-aware-routing
kubectl delete backendtrafficpolicy/zone-aware-routing
```

[Traffic Distribution]: https://kubernetes.io/docs/concepts/services-networking/service/#traffic-distribution
[Topology Aware Routing]: https://kubernetes.io/docs/concepts/services-networking/topology-aware-routing/
[Kubernetes well-known metadata]: https://kubernetes.io/docs/reference/labels-annotations-taints/#topologykubernetesiozone
[Envoy Zone Aware Routing]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/zone_aware
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy