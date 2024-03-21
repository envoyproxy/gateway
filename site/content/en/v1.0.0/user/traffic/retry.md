---
title: Retry
---

A retry setting specifies the maximum number of times an Envoy proxy attempts to connect to a service if the initial call fails. Retries can enhance service availability and application performance by making sure that calls don’t fail permanently because of transient problems such as a temporarily overloaded service or network. The interval between retries prevents the called service from being overwhelmed with requests.

Envoy Gateway supports the following retry settings:
- **NumRetries**: is the number of retries to be attempted. Defaults to 2.
- **RetryOn**: specifies the retry trigger condition.
- **PerRetryPolicy**: is the retry policy to be applied per retry attempt.

Envoy Gateway introduces a new CRD called [BackendTrafficPolicy](../../../api/extension_types#backendtrafficpolicy) that allows the user to describe their desired retry settings. This instantiated resource can be linked to a [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/), [HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/) or [GRPCRoute](https://gateway-api.sigs.k8s.io/api-types/grpcroute/) resource.

**Note**: There are distinct circuit breaker counters for each `BackendReference` in an `xRoute` rule. Even if a `BackendTrafficPolicy` targets a `Gateway`, each `BackendReference` in that gateway still has separate circuit breaker counter.

## Prerequisites

Follow the installation step from the [Quickstart Guide](../../quickstart) to install Envoy Gateway and sample resources.

## Test and customize retry settings

Before applying a `BackendTrafficPolicy` with retry setting to a route, let's test the default retry settings.

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/status/500"
```

It will return `500` response immediately.

```console
*   Trying 172.18.255.200:80...
* Connected to 172.18.255.200 (172.18.255.200) port 80
> GET /status/500 HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.4.0
> Accept: */*
>
< HTTP/1.1 500 Internal Server Error
< date: Fri, 01 Mar 2024 15:12:55 GMT
< content-length: 0
<
* Connection #0 to host 172.18.255.200 left intact
```

Let's create a `BackendTrafficPolicy` with a retry setting.

The request will be retried 5 times with a 100ms base interval and a 10s maximum interval.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: retry-for-route
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
    namespace: default
  retry:
    numRetries: 5
    perRetry:
      backOff:
        baseInterval: 100ms
        maxInterval: 10s
      timeout: 250ms
    retryOn:
      httpStatusCodes:
        - 500
      triggers:
        - connect-failure
        - retriable-status-codes
EOF
```

Execute the test again.

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/status/500"
```

It will return `500` response after a few while.

```console
*   Trying 172.18.255.200:80...
* Connected to 172.18.255.200 (172.18.255.200) port 80
> GET /status/500 HTTP/1.1
> Host: www.example.com
> User-Agent: curl/8.4.0
> Accept: */*
>
< HTTP/1.1 500 Internal Server Error
< date: Fri, 01 Mar 2024 15:15:53 GMT
< content-length: 0
<
* Connection #0 to host 172.18.255.200 left intact
```

Let's check the stats to see the retry behavior.

```shell
egctl x stats envoy-proxy -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-name=eg,gateway.envoyproxy.io/owning-gateway-namespace=default | grep "envoy_cluster_upstream_rq_retry{envoy_cluster_name=\"httproute/default/backend/rule/0\"}"
```

You will expect to see the stats.

```console
envoy_cluster_upstream_rq_retry{envoy_cluster_name="httproute/default/backend/rule/0"} 5
```
