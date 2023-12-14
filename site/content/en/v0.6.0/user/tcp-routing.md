---
title: "TCP Routing"
---

[TCPRoute][] provides a way to route TCP requests. When combined with a Gateway listener, it can be used to forward
connections on the port specified by the listener to a set of backends specified by the TCPRoute. To learn more about
HTTP routing, refer to the [Gateway API documentation][].

## Installation

Install Envoy Gateway:

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.6.0 -n envoy-gateway-system --create-namespace
```

Wait for Envoy Gateway to become available:

```shell
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

## Configuration

In this example, we have one Gateway resource and two TCPRoute resources that distribute the traffic with the following
rules:

All TCP streams on port `8088` of the Gateway are forwarded to port 3001 of `foo` Kubernetes Service.
All TCP streams on port `8089` of the Gateway are forwarded to port 3002 of `bar` Kubernetes Service.
In this example two TCP listeners will be applied to the Gateway in order to route them to two separate backend
TCPRoutes, note that the protocol set for the listeners on the Gateway is TCP:

Install the GatewayClass and a `tcp-gateway` Gateway first.

```shell
cat <<EOF | kubectl apply -f -
kind: GatewayClass
apiVersion: gateway.networking.k8s.io/v1
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: tcp-gateway
spec:
  gatewayClassName: eg
  listeners:
  - name: foo
    protocol: TCP
    port: 8088
    allowedRoutes:
      kinds:
      - kind: TCPRoute
  - name: bar
    protocol: TCP
    port: 8089
    allowedRoutes:
      kinds:
      - kind: TCPRoute
EOF
```

Install two services `foo` and `bar`, which are binded to `backend-1` and `backend-2`.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: foo
  labels:
    app: backend-1
spec:
  ports:
    - name: http
      port: 3001
      targetPort: 3000
  selector:
    app: backend-1
---
apiVersion: v1
kind: Service
metadata:
  name: bar
  labels:
    app: backend-2
spec:
  ports:
    - name: http
      port: 3002
      targetPort: 3000
  selector:
    app: backend-2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend-1
      version: v1
  template:
    metadata:
      labels:
        app: backend-1
        version: v1
    spec:
      containers:
        - image: gcr.io/k8s-staging-ingressconformance/echoserver:v20221109-7ee2f3e
          imagePullPolicy: IfNotPresent
          name: backend-1
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
            - name: SERVICE_NAME
              value: foo
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
            - name: SERVICE_NAME
              value: bar
EOF
```

Install two TCPRoutes `tcp-app-1` and `tcp-app-2` with different `sectionName`:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: tcp-app-1
spec:
  parentRefs:
  - name: tcp-gateway
    sectionName: foo
  rules:
  - backendRefs:
    - name: foo
      port: 3001
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: tcp-app-2
spec:
  parentRefs:
  - name: tcp-gateway
    sectionName: bar
  rules:
  - backendRefs:
    - name: bar
      port: 3002
EOF
```

In the above example we separate the traffic for the two separate backend TCP Services by using the sectionName field in
the parentRefs:

``` yaml
spec:
  parentRefs:
  - name: tcp-gateway
    sectionName: foo
```

This corresponds directly with the name in the listeners in the Gateway:

``` yaml
  listeners:
  - name: foo
    protocol: TCP
    port: 8088
  - name: bar
    protocol: TCP
    port: 8089
```

In this way each TCPRoute "attaches" itself to a different port on the Gateway so that the `foo` service
is taking traffic for port `8088` from outside the cluster and `bar` service takes the port `8089` traffic.

Before testing, please get the tcp-gateway Gateway's address first:

```shell
export GATEWAY_HOST=$(kubectl get gateway/tcp-gateway -o jsonpath='{.status.addresses[0].value}')
```

You can try to use nc to test the TCP connections of envoy gateway with different ports, and you can see them succeeded:

```shell
nc -zv ${GATEWAY_HOST} 8088

nc -zv ${GATEWAY_HOST} 8089
```

You can also try to send requests to envoy gateway and get responses as shown below:

```shell
curl -i "http://${GATEWAY_HOST}:8088"

HTTP/1.1 200 OK
Content-Type: application/json
X-Content-Type-Options: nosniff
Date: Tue, 03 Jan 2023 10:18:36 GMT
Content-Length: 267

{
 "path": "/",
 "host": "xxx.xxx.xxx.xxx:8088",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "User-Agent": [
   "curl/7.85.0"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "foo",
 "pod": "backend-1-c6c5fb958-dl8vl"
}
```

You can see that the traffic routing to `foo` service when sending request to `8088` port.

```shell
curl -i "http://${GATEWAY_HOST}:8089"

HTTP/1.1 200 OK
Content-Type: application/json
X-Content-Type-Options: nosniff
Date: Tue, 03 Jan 2023 10:19:28 GMT
Content-Length: 267

{
 "path": "/",
 "host": "xxx.xxx.xxx.xxx:8089",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "User-Agent": [
   "curl/7.85.0"
  ]
 },
 "namespace": "default",
 "ingress": "",
 "service": "bar",
 "pod": "backend-2-98fcff498-hcmgb"
}                                            
```

You can see that the traffic routing to `bar` service when sending request to `8089` port.

[TCPRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TCPRoute
[Gateway API documentation]: https://gateway-api.sigs.k8s.io/
