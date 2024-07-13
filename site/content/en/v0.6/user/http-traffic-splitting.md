---
title: "HTTPRoute Traffic Splitting"
---

The [HTTPRoute][] resource allows one or more [backendRefs][] to be provided. Requests will be routed to these upstreams
if they match the rules of the HTTPRoute. If an invalid backendRef is configured, then HTTP responses will be returned
with status code `500` for all requests that would have been sent to that backend.

## Installation

Follow the steps from the [Quickstart Guide](../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Single backendRef

When a single backendRef is configured in a HTTPRoute, it will receive 100% of the traffic.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-headers
spec:
  parentRefs:
  - name: eg
  hostnames:
  - backends.example
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
EOF
```

The HTTPRoute status should indicate that it has been accepted and is bound to the example Gateway.

```shell
kubectl get httproute/http-headers -o yaml
```

Get the Gateway's address:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Querying `backends.example/get` should result in a `200` response from the example Gateway and the output from the
example app should indicate which pod handled the request. There is only one pod in the deployment for the example app
from the quickstart, so it will be the same on all subsequent requests.

```console
$ curl -vvv --header "Host: backends.example" "http://${GATEWAY_HOST}/get"
...
> GET /get HTTP/1.1
> Host: backends.example
> User-Agent: curl/7.81.0
> Accept: */*
> add-header: something
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< content-length: 474
< x-envoy-upstream-service-time: 0
< server: envoy
<
...
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-79665566f5-s589f"
...
```

## Multiple backendRefs

If multiple backendRefs are configured, then traffic will be split between the backendRefs equally unless a weight is
configured.

First, create a second instance of the example app from the quickstart:

```shell
cat <<EOF | kubectl apply -f -
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: backend-2
---
apiVersion: v1
kind: Service
metadata:
  name: backend-2
  labels:
    app: backend-2
    service: backend-2
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: backend-2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend-2
      version: v1
  template:
    metadata:
      labels:
        app: backend-2
        version: v1
    spec:
      serviceAccountName: backend-2
      containers:
        - image: gcr.io/k8s-staging-ingressconformance/echoserver:v20221109-7ee2f3e
          imagePullPolicy: IfNotPresent
          name: backend-2
          ports:
            - containerPort: 3000
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
EOF
```

Then create an HTTPRoute that uses both the app from the quickstart and the second instance that was just created

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-headers
spec:
  parentRefs:
  - name: eg
  hostnames:
  - backends.example
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
    - group: ""
      kind: Service
      name: backend-2
      port: 3000
EOF
```

Querying `backends.example/get` should result in `200` responses from the example Gateway and the output from the
example app that indicates which pod handled the request should switch between the first pod and the second one from the
new deployment on subsequent requests.

```console
$ curl -vvv --header "Host: backends.example" "http://${GATEWAY_HOST}/get"
...
> GET /get HTTP/1.1
> Host: backends.example
> User-Agent: curl/7.81.0
> Accept: */*
> add-header: something
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< content-length: 474
< x-envoy-upstream-service-time: 0
< server: envoy
<
...
 "namespace": "default",
 "ingress": "",
 "service": "",
 "pod": "backend-75bcd4c969-lsxpz"
...
```

## Weighted backendRefs

If multiple backendRefs are configured and an un-even traffic split between the backends is desired, then the `weight`
field can be used to control the weight of requests to each backend. If weight is not configured for a backendRef it is
assumed to be `1`.

The [weight field in a backendRef][backendRefs] controls the distribution of the traffic split. The proportion of
requests to a single backendRef is calculated by dividing its `weight` by the sum of all backendRef weights in the
HTTPRoute. The weight is not a percentage and the sum of all weights does not need to add up to 100.

The HTTPRoute below will configure the gateway to send 80% of the traffic to the backend service, and 20% to the
backend-2 service.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-headers
spec:
  parentRefs:
  - name: eg
  hostnames:
  - backends.example
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
      weight: 8
    - group: ""
      kind: Service
      name: backend-2
      port: 3000
      weight: 2
EOF
```

## Invalid backendRefs

backendRefs can be considered invalid for the following reasons:

- The `group` field is configured to something other than `""`. Currently, only the core API group (specified by
  omitting the group field or setting it to an empty string) is supported
- The `kind` field is configured to anything other than `Service`. Envoy Gateway currently only supports Kubernetes
  Service backendRefs
- The backendRef configures a service with a `namespace` not permitted by any existing ReferenceGrants
- The `port` field is not configured or is configured to a port that does not exist on the Service
- The named Service configured by the backendRef cannot be found

Modifying the above example to make the backend-2 backendRef invalid by using a port that does not exist on the Service
will result in 80% of the traffic being sent to the backend service, and 20% of the traffic receiving an HTTP response
with status code `500`.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-headers
spec:
  parentRefs:
  - name: eg
  hostnames:
  - backends.example
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
      weight: 8
    - group: ""
      kind: Service
      name: backend-2
      port: 9000
      weight: 2
EOF
```

Querying `backends.example/get` should result in `200` responses 80% of the time, and `500` responses 20% of the time.

```console
$ curl -vvv --header "Host: backends.example" "http://${GATEWAY_HOST}/get"
> GET /get HTTP/1.1
> Host: backends.example
> User-Agent: curl/7.81.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 500 Internal Server Error
< server: envoy
< content-length: 0
<
```

[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[backendRefs]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.BackendRef
