---
title: "Session Persistence"
---

Session Persistence allows client requests to be consistently routed to the same backend service instance. This is useful in many scenarios, such as when an application needs to maintain state across multiple requests. In Envoy Gateway, session persistence can be enabled by configuring [HTTPRoute][].

Envoy Gateway supports following session persistence types:
- **Cookie-based** Session Persistence: Session persistence is achieved based on specific cookie information in the request.
- **Header-based** Session Persistence: Session persistence is achieved based on specific header information in the request.

## Prerequisites

### Install Envoy Gateway

{{< boilerplate prerequisites >}}

For better testing the session persistence, you can add more hosts in upstream cluster by increasing the replicas of one deployment:

```shell
kubectl patch deployment backend -n default -p '{"spec": {"replicas": 4}}'
```

## Cookie-based Session Persistence

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: cookie
spec:
  parentRefs:
    - name: eg
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: backend
          port: 3000
      sessionPersistence:
        sessionName: Session-A
        type: Cookie
        absoluteTimeout: 10s
        cookieConfig:
          lifetimeType: Permanent
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: cookie
spec:
  parentRefs:
    - name: eg
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: backend
          port: 3000
      sessionPersistence:
        sessionName: Session-A
        type: Cookie
        absoluteTimeout: 10s
        cookieConfig:
          lifetimeType: Permanent
```

{{% /tab %}}
{{< /tabpane >}}

{{< boilerplate rollout-envoy-gateway >}}

### Testing

Send a request to get a cookie and test:

```shell
COOKIE=$(curl --verbose http://localhost:8888/get 2>&1 | grep "set-cookie" | awk '{print $3}')
for i in `seq 5`; do
    curl -H "Cookie: $COOKIE" http://localhost:8888/get 2>/dev/null | grep pod
done
```

You can see all responses are from the same pod:

```console
 "pod": "backend-765694d47f-lxwdf"
 "pod": "backend-765694d47f-lxwdf"
 "pod": "backend-765694d47f-lxwdf"
 "pod": "backend-765694d47f-lxwdf"
 "pod": "backend-765694d47f-lxwdf"
```

Wait for cookie expiration, test again:

```shell
sleep 10 && for i in `seq 5`; do
    curl -H "Cookie: $COOKIE" -H "Host: www.example.com" http://localhost:8888/get 2>/dev/null | grep pod
done
```

```console
 "pod": "backend-765694d47f-kvwqb"
 "pod": "backend-765694d47f-lxwdf"
 "pod": "backend-765694d47f-kvwqb"
 "pod": "backend-765694d47f-2ff9s"
 "pod": "backend-765694d47f-lxwdf"
```

Due to cookie expiration, no session any more.

## Header-based Session Persistence

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: header
spec:
  parentRefs:
    - name: eg
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: backend
          port: 3000
      sessionPersistence:
        sessionName: Session-A
        type: Header
        absoluteTimeout: 10s
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: header
spec:
  parentRefs:
    - name: eg
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: backend
          port: 3000
      sessionPersistence:
        sessionName: Session-A
        type: Header
        absoluteTimeout: 10s
```

{{% /tab %}}
{{< /tabpane >}}

### Testing


Send a request to get a cookie and test:

```shell
HEADER=$(curl --verbose http://localhost:8888/get 2>&1 | grep "session-a" | awk '{print $3}')
for i in `seq 5`; do
    curl -H "Session-A: $HEADER" http://localhost:8888/get 2>/dev/null | grep pod
done
```

You can see all responses are from the same pod:

```console
 "pod": "backend-765694d47f-gn7q2"
 "pod": "backend-765694d47f-gn7q2"
 "pod": "backend-765694d47f-gn7q2"
 "pod": "backend-765694d47f-gn7q2"
 "pod": "backend-765694d47f-gn7q2"
```

We remove the header and test again:

```shell
for i in `seq 5`; do
    curl http://localhost:8888/get 2>/dev/null | grep pod
done
```

```console
 "pod": "backend-765694d47f-2ff9s"
 "pod": "backend-765694d47f-kvwqb"
 "pod": "backend-765694d47f-2ff9s"
 "pod": "backend-765694d47f-lxwdf"
 "pod": "backend-765694d47f-kvwqb"

```

[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute