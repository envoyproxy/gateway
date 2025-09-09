---
title: Failover 
---

Active-passive failover in an API gateway setup is like having a backup plan in place to keep things
running smoothly if something goes wrong. Here’s why it’s valuable:

* Staying Online: When the main (or "active") backend has issues or goes offline,
the fallback (or "passive") backend is ready to step in instantly.
This helps keep your API accessible and your services running, so users don’t even notice any interruptions.

* Automatic Switch Over: If a problem occurs, the system can automatically switch traffic over to the fallback backend. 
This avoids needing someone to jump in and fix things manually, which could take time and might even lead to mistakes.

* Lower Costs: In an active-passive setup, the fallback backend doesn’t need to work all the time—it’s just on standby. 
This can save on costs (like cloud egress costs) compared to setups where both backend are running at full capacity.

* Peace of Mind with Redundancy: Although the fallback backend isn’t handling traffic daily, it's there as a safety net.
If something happens with the primary backend, the backup can take over immediately, ensuring your service doesn’t skip a beat.

## Prerequisites

{{< boilerplate prerequisites >}}

## Test

* We'll first create two [services][Service] & [deployments][Deployment], called `active` and `passive`, representing an `active` and `passive` backend application.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: active 
  labels:
    app: active
    service: active
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: active
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: active
spec:
  replicas: 1
  selector:
    matchLabels:
      app: active
      version: v1
  template:
    metadata:
      labels:
        app: active
        version: v1
    spec:
      containers:
        - image: gcr.io/k8s-staging-gateway-api/echo-basic:v20231214-v1.0.0-140-gf544a46e
          imagePullPolicy: IfNotPresent
          name: active 
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
apiVersion: v1
kind: Service
metadata:
  name: passive 
  labels:
    app: passive
    service: passive
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: passive
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: passive
spec:
  replicas: 1
  selector:
    matchLabels:
      app: passive
      version: v1
  template:
    metadata:
      labels:
        app: passive
        version: v1
    spec:
      containers:
        - image: gcr.io/k8s-staging-gateway-api/echo-basic:v20231214-v1.0.0-140-gf544a46e
          imagePullPolicy: IfNotPresent
          name: passive 
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

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: active 
  labels:
    app: active
    service: active
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: active
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: active
spec:
  replicas: 1
  selector:
    matchLabels:
      app: active
      version: v1
  template:
    metadata:
      labels:
        app: active
        version: v1
    spec:
      containers:
        - image: gcr.io/k8s-staging-gateway-api/echo-basic:v20231214-v1.0.0-140-gf544a46e
          imagePullPolicy: IfNotPresent
          name: active 
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
apiVersion: v1
kind: Service
metadata:
  name: passive 
  labels:
    app: passive
    service: passive
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: passive
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: passive
spec:
  replicas: 1
  selector:
    matchLabels:
      app: passive
      version: v1
  template:
    metadata:
      labels:
        app: passive
        version: v1
    spec:
      containers:
        - image: gcr.io/k8s-staging-gateway-api/echo-basic:v20231214-v1.0.0-140-gf544a46e
          imagePullPolicy: IfNotPresent
          name: passive 
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
```

{{% /tab %}}
{{< /tabpane >}}


* Follow the instructions [here](./../../tasks/traffic/backend/#enable-backend) to enable the Backend API

* Create two [Backend][] resources that are used to represent the `active` backend and `passive` backend.
Note, we've set `fallback: true` for the `passive` backend to indicate its a passive backend


{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: passive 
spec:
  fallback: true
  endpoints:
    - fqdn:
        hostname: passive.default.svc.cluster.local
        port: 3000 
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: active
spec:
  endpoints:
  - fqdn:
      hostname: active.default.svc.cluster.local 
      port: 3000
---
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: passive 
spec:
  fallback: true
  endpoints:
    - fqdn:
        hostname: passive.default.svc.cluster.local
        port: 3000 
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: active
spec:
  endpoints:
  - fqdn:
      hostname: active.default.svc.cluster.local 
      port: 3000
---
```

{{% /tab %}}
{{< /tabpane >}}

* Lets create an [HTTPRoute][] that can route to both these backends

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: ha-example
  namespace: default
spec:
  hostnames:
  - www.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg 
    namespace: default
  rules:
  - backendRefs:
    - group: gateway.envoyproxy.io
      kind: Backend
      name: active
      namespace: default
      port: 3000
    - group: gateway.envoyproxy.io
      kind: Backend
      name: passive 
      namespace: default
      port: 3000
    matches:
    - path:
        type: PathPrefix
        value: /test
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: ha-example
  namespace: default
spec:
  hostnames:
  - www.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg 
    namespace: default
  rules:
  - backendRefs:
    - group: gateway.envoyproxy.io
      kind: Backend
      name: active
      namespace: default
      port: 3000
    - group: gateway.envoyproxy.io
      kind: Backend
      name: passive 
      namespace: default
      port: 3000
    matches:
    - path:
        type: PathPrefix
        value: /test
```

{{% /tab %}}
{{< /tabpane >}}

* Lets configure a [BackendTrafficPolicy][] with a passive health check setting to detect an transient errors.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: passive-health-check
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: ha-example 
  healthCheck:
    passive:
      baseEjectionTime: 10s
      interval: 2s
      maxEjectionPercent: 100
      consecutive5XxErrors: 1 
      consecutiveGatewayErrors: 0
      consecutiveLocalOriginFailures: 1
      splitExternalLocalOriginErrors: false
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
  name: passive-health-check
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: ha-example 
  healthCheck:
    passive:
      baseEjectionTime: 10s
      interval: 2s
      maxEjectionPercent: 100
      consecutive5XxErrors: 1 
      consecutiveGatewayErrors: 0
      consecutiveLocalOriginFailures: 1
      splitExternalLocalOriginErrors: false
```

{{% /tab %}}
{{< /tabpane >}}



* Lets send 10 requests. You should see that they all go to the `active` backend.

```shell
for i in {1..10; do curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/test 2>/dev/null | jq .pod; done
```

```console
"active-5bb896774f-lz8s9"
"active-5bb896774f-lz8s9"
"active-5bb896774f-lz8s9"
"active-5bb896774f-lz8s9"
"active-5bb896774f-lz8s9"
"active-5bb896774f-lz8s9"
"active-5bb896774f-lz8s9"
"active-5bb896774f-lz8s9"
"active-5bb896774f-lz8s9"
"active-5bb896774f-lz8s9"
```

* Lets simulate a failure in the `active` backend by changing the server listening port to `5000`

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: active
spec:
  replicas: 1
  selector:
    matchLabels:
      app: active
      version: v1
  template:
    metadata:
      labels:
        app: active
        version: v1
    spec:
      containers:
        - image: gcr.io/k8s-staging-gateway-api/echo-basic:v20231214-v1.0.0-140-gf544a46e
          imagePullPolicy: IfNotPresent
          name: active 
          ports:
            - containerPort: 3000
          env:
            - name: HTTP_PORT
              value: "5000"
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

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: active
spec:
  replicas: 1
  selector:
    matchLabels:
      app: active
      version: v1
  template:
    metadata:
      labels:
        app: active
        version: v1
    spec:
      containers:
        - image: gcr.io/k8s-staging-gateway-api/echo-basic:v20231214-v1.0.0-140-gf544a46e
          imagePullPolicy: IfNotPresent
          name: active 
          ports:
            - containerPort: 3000
          env:
            - name: HTTP_PORT
              value: "5000"
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
```

{{% /tab %}}
{{< /tabpane >}}

* Lets send 10 requests again. You should see them all being sent to the `passive` backend

```shell
for i in {1..10; do curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/test 2>/dev/null | jq .pod; done
```

```console
parse error: Invalid numeric literal at line 1, column 9
"passive-7ddbf945c9-wkc4f"
"passive-7ddbf945c9-wkc4f"
"passive-7ddbf945c9-wkc4f"
"passive-7ddbf945c9-wkc4f"
"passive-7ddbf945c9-wkc4f"
"passive-7ddbf945c9-wkc4f"
"passive-7ddbf945c9-wkc4f"
"passive-7ddbf945c9-wkc4f"
"passive-7ddbf945c9-wkc4f"
```

The first error can be avoided by configuring [retries](./../../tasks/traffic/retry.md).

[Backend]: ../../../api/extension_types#backend
[BackendTrafficPolicy]: ../../../api/extension_types#backendtrafficpolicy
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[Service]: https://kubernetes.io/docs/concepts/services-networking/service/
[Deployment]: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/
