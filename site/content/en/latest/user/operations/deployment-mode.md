---
title: "Deployment Mode"
---
## Deployment modes

### One GatewayClass per Envoy Gateway Controller
* An Envoy Gateway is associated with a single [GatewayClass][] resource under one controller.
This is the simplest deployment mode and is suitable for scenarios where each Gateway needs to have its own dedicated set of resources and configurations.

### Multiple GatewayClasses per Envoy Gateway Controller
* An Envoy Gateway is associated with multiple [GatewayClass][] resources under one controller.
* Support for accepting multiple GatewayClasses was added [here][issue1231].

### Separate Envoy Gateway Controllers
If you've instantiated multiple GatewayClasses, you can also run separate Envoy Gateway controllers in different namespaces,
linking a GatewayClass to each of them for multi-tenancy.
Please follow the example [Multi-tenancy](#multi-tenancy).

### Merged Gateways onto a single EnvoyProxy fleet
By default, each Gateway has its own dedicated set of Envoy Proxy and its configurations.
However, for some deployments, it may be more convenient to merge listeners across multiple Gateways and deploy a single Envoy Proxy fleet.
This can help to efficiently utilize the infra resources in the cluster and manage them in a centralized manner, or have a single IP address for all of the listeners.
Setting the `mergeGateways` field in the EnvoyProxy resource linked to GatewayClass will result in merging all Gateway listeners under one GatewayClass resource.
* The tuple of port, protocol, and hostname must be unique across all Listeners.
Please follow the example [Merged gateways deployment](#Merged gateways deployment).

### Supported Modes

#### Kubernetes

* The default deployment model is - Envoy Gateway **watches** for resources such a `Service` & `HTTPRoute` in **all** namespaces
and **creates** managed data plane resources such as EnvoyProxy `Deployment` in the **namespace where Envoy Gateway is running**.
* Envoy Gateway also supports [Namespaced deployment mode][], you can watch resources in the specific namespaces by assigning
`EnvoyGateway.provider.kubernetes.watch.namespaces` or `EnvoyGateway.provider.kubernetes.watch.namespaceSelector` and **creates** managed data plane resources in the **namespace where Envoy Gateway is running**.
* Support for alternate deployment modes is being tracked [here][issue1117].

### Multi-tenancy

#### Kubernetes

* A `tenant` is a group within an organization (e.g. a team or department) who shares organizational resources. We recommend
each `tenant` deploy their own Envoy Gateway controller in their respective `namespace`. Below is an example of deploying Envoy Gateway
by the `marketing` and `product` teams in separate namespaces.

* Lets deploy Envoy Gateway in the `marketing` namespace and also watch resources only in this namespace. We are also setting the controller name to a unique string here `gateway.envoyproxy.io/marketing-gatewayclass-controller`.

```shell
helm install \
--set config.envoyGateway.gateway.controllerName=gateway.envoyproxy.io/marketing-gatewayclass-controller \
--set config.envoyGateway.provider.kubernetes.watch.type=Namespaces \
--set config.envoyGateway.provider.kubernetes.watch.namespaces={marketing} \
eg-marketing oci://docker.io/envoyproxy/gateway-helm \
--version v0.0.0-latest -n marketing --create-namespace
```

Lets create a `GatewayClass` linked to the marketing team's Envoy Gateway controller, and as well other resources linked to it, so the `backend` application operated by this team can be exposed to external clients.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg-marketing
spec:
  controllerName: gateway.envoyproxy.io/marketing-gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
  namespace: marketing
spec:
  gatewayClassName: eg-marketing
  listeners:
    - name: http
      protocol: HTTP
      port: 8080
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: backend
  namespace: marketing
---
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: marketing
  labels:
    app: backend
    service: backend
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: backend
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: marketing
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend
      version: v1
  template:
    metadata:
      labels:
        app: backend
        version: v1
    spec:
      serviceAccountName: backend
      containers:
        - image: gcr.io/k8s-staging-gateway-api/echo-basic:v20231214-v1.0.0-140-gf544a46e
          imagePullPolicy: IfNotPresent
          name: backend
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
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
  namespace: marketing
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.marketing.example.com"
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
            value: /
EOF
```

Lets port forward to the generated envoy proxy service in the `marketing` namespace and send a request to it.

```shell
export ENVOY_SERVICE=$(kubectl get svc -n marketing --selector=gateway.envoyproxy.io/owning-gateway-namespace=marketing,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
kubectl -n marketing port-forward service/${ENVOY_SERVICE} 8888:8080 &
```

```shell
curl --verbose --header "Host: www.marketing.example.com" http://localhost:8888/get
```

```console
*   Trying 127.0.0.1:8888...
* Connected to localhost (127.0.0.1) port 8888 (#0)
> GET /get HTTP/1.1
> Host: www.marketing.example.com
> User-Agent: curl/7.86.0
> Accept: */*
>
Handling connection for 8888
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Thu, 20 Apr 2023 19:19:42 GMT
< content-length: 521
< x-envoy-upstream-service-time: 0
< server: envoy
<
{
 "path": "/get",
 "host": "www.marketing.example.com",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "User-Agent": [
   "curl/7.86.0"
  ],
  "X-Envoy-Expected-Rq-Timeout-Ms": [
   "15000"
  ],
  "X-Envoy-Internal": [
   "true"
  ],
  "X-Forwarded-For": [
   "10.1.0.157"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "c637977c-458a-48ae-92b3-f8c429849322"
  ]
 },
 "namespace": "marketing",
 "ingress": "",
 "service": "",
 "pod": "backend-74888f465f-bcs8f"
* Connection #0 to host localhost left intact
```

* Lets deploy Envoy Gateway in the `product` namespace and also watch resources only in this namespace.

```shell
helm install \
--set config.envoyGateway.gateway.controllerName=gateway.envoyproxy.io/product-gatewayclass-controller \
--set config.envoyGateway.provider.kubernetes.watch.type=Namespaces \
--set config.envoyGateway.provider.kubernetes.watch.namespaces={product} \
eg-product oci://docker.io/envoyproxy/gateway-helm \
--version v0.0.0-latest -n product --create-namespace
```

Lets create a `GatewayClass` linked to the product team's Envoy Gateway controller, and as well other resources linked to it, so the `backend` application operated by this team can be exposed to external clients.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg-product
spec:
  controllerName: gateway.envoyproxy.io/product-gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
  namespace: product
spec:
  gatewayClassName: eg-product
  listeners:
    - name: http
      protocol: HTTP
      port: 8080
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: backend
  namespace: product
---
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: product
  labels:
    app: backend
    service: backend
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: backend
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: product
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend
      version: v1
  template:
    metadata:
      labels:
        app: backend
        version: v1
    spec:
      serviceAccountName: backend
      containers:
        - image: gcr.io/k8s-staging-gateway-api/echo-basic:v20231214-v1.0.0-140-gf544a46e
          imagePullPolicy: IfNotPresent
          name: backend
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
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
  namespace: product
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.product.example.com"
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
            value: /
EOF
```

Lets port forward to the generated envoy proxy service in the `product` namespace and send a request to it.

```shell
export ENVOY_SERVICE=$(kubectl get svc -n product --selector=gateway.envoyproxy.io/owning-gateway-namespace=product,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
kubectl -n product port-forward service/${ENVOY_SERVICE} 8889:8080 &
```

```shell
curl --verbose --header "Host: www.product.example.com" http://localhost:8889/get
```

```shell
*   Trying 127.0.0.1:8889...
* Connected to localhost (127.0.0.1) port 8889 (#0)
> GET /get HTTP/1.1
> Host: www.product.example.com
> User-Agent: curl/7.86.0
> Accept: */*
>
Handling connection for 8889
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< content-type: application/json
< x-content-type-options: nosniff
< date: Thu, 20 Apr 2023 19:20:17 GMT
< content-length: 517
< x-envoy-upstream-service-time: 0
< server: envoy
<
{
 "path": "/get",
 "host": "www.product.example.com",
 "method": "GET",
 "proto": "HTTP/1.1",
 "headers": {
  "Accept": [
   "*/*"
  ],
  "User-Agent": [
   "curl/7.86.0"
  ],
  "X-Envoy-Expected-Rq-Timeout-Ms": [
   "15000"
  ],
  "X-Envoy-Internal": [
   "true"
  ],
  "X-Forwarded-For": [
   "10.1.0.156"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "39196453-2250-4331-b756-54003b2853c2"
  ]
 },
 "namespace": "product",
 "ingress": "",
 "service": "",
 "pod": "backend-74888f465f-64fjs"
* Connection #0 to host localhost left intact
```

With the below command you can ensure that you are no able to access the marketing team's backend exposed using the `www.marketing.example.com` hostname
and the product team's data plane.

```shell
curl --verbose --header "Host: www.marketing.example.com" http://localhost:8889/get
```

```console
*   Trying 127.0.0.1:8889...
* Connected to localhost (127.0.0.1) port 8889 (#0)
> GET /get HTTP/1.1
> Host: www.marketing.example.com
> User-Agent: curl/7.86.0
> Accept: */*
>
Handling connection for 8889
* Mark bundle as not supporting multiuse
< HTTP/1.1 404 Not Found
< date: Thu, 20 Apr 2023 19:22:13 GMT
< server: envoy
< content-length: 0
<
* Connection #0 to host localhost left intact
```

### Merged gateways deployment

In this example, we will deploy GatewayClass
```shell
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: merged-eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: custom-proxy-config
    namespace: envoy-gateway-system
```
with a referenced [EnvoyProxy][] resource configured to enable merged Gateways deployment mode.
```shell
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  mergeGateways: true
```

#### Deploy merged-gateways example

Deploy resources from the example to your cluster.
```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/merged-gateways.yaml
```

[EnvoyProxy]: ../../../api/extension_types#envoyproxy
[GatewayClass]: https://gateway-api.sigs.k8s.io/api-types/gatewayclass/
[Namespaced deployment mode]: ../../../api/extension_types#kuberneteswatchmode
[issue1231]: https://github.com/envoyproxy/gateway/issues/1231
[issue1117]: https://github.com/envoyproxy/gateway/issues/1117
