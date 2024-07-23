---
title: "HTTPRoute Request Mirroring"
---

The [HTTPRoute][] resource allows one or more [backendRefs][] to be provided. Requests will be routed to these upstreams. It is possible to divide the traffic between these backends using [Traffic Splitting](http-traffic-splitting.md), but it is also possible to mirror requests to another Service instead. Request mirroring is accomplished using Gateway API's [HTTPRequestMirrorFilter][] on the `HTTPRoute`.

When requests are made to a `HTTPRoute` that uses a `HTTPRequestMirrorFilter`, the response will never come from the `backendRef` defined in the filter. Responses from the mirror `backendRef` are always ignored.

## Installation

Follow the steps from the [Quickstart Guide][] to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Mirroring the Traffic

Next, create a new `Deployment` and `Service` to mirror requests to. The following example will use
a second instance of the application deployed in the quickstart.

```shell
kubectl apply -f - <<EOF
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

Then create an `HTTPRoute` that uses a `HTTPRequestMirrorFilter` to send requests to the original
service from the quickstart, and mirror request to the service that was just deployed.

```shell
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: http-mirror
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
    filters:
    - type: RequestMirror
      requestMirror:
        backendRef:
          kind: Service
          name: backend-2
          port: 3000
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
EOF
```

The HTTPRoute status should indicate that it has been accepted and is bound to the example Gateway.

```shell
kubectl get httproute/http-mirror -o yaml
```

Get the Gateway's address:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Querying `backends.example/get` should result in a `200` response from the example Gateway and the output from the
example app should indicate which pod handled the request. There is only one pod in the deployment for the example app
from the quickstart, so it will be the same on all subsequent requests.

```console
$ curl -v --header "Host: backends.example" "http://${GATEWAY_HOST}/get"
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

Check the logs of the pods and you will see that the original deployment and the new deployment each got a request:

```shell
$ kubectl logs deploy/backend && kubectl logs deploy/backend-2
...
Starting server, listening on port 3000 (http)
Echoing back request made to /get to client (10.42.0.10:41566)
Starting server, listening on port 3000 (http)
Echoing back request made to /get to client (10.42.0.10:45096)
```

## Multiple BackendRefs

When an `HTTPRoute` has multiple `backendRefs` and an `HTTPRequestMirrorFilter`, traffic splitting will still behave the same as it normally would for the main `backendRefs` while the `backendRef` of the `HTTPRequestMirrorFilter` will continue receiving mirrored copies of the incoming requests.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: http-mirror
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
    filters:
    - type: RequestMirror
      requestMirror:
        backendRef:
          kind: Service
          name: backend-2
          port: 3000
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
    - group: ""
      kind: Service
      name: backend-3
      port: 3000
EOF
```

## Multiple HTTPRequestMirrorFilters

Multiple `HTTPRequestMirrorFilters` are not supported on the same `HTTPRoute` `rule`. When attempting to do so, the admission webhook will reject the configuration.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: http-mirror
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
    filters:
    - type: RequestMirror
      requestMirror:
        backendRef:
          kind: Service
          name: backend-2
          port: 3000
    - type: RequestMirror
      requestMirror:
        backendRef:
          kind: Service
          name: backend-3
          port: 3000
    backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
EOF
```

```console
Error from server: error when creating "STDIN": admission webhook "validate.gateway.networking.k8s.io" denied the request: spec.rules[0].filters: Invalid value: "RequestMirror": cannot be used multiple times in the same rule
```

[Quickstart Guide]: quickstart.md
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[backendRefs]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.BackendRef
[HTTPRequestMirrorFilter]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRequestMirrorFilter
