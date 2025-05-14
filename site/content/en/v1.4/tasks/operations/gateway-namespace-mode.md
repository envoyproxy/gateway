---
title: "Gateway Namespace Mode"
---

{{% alert title="Notice" color="warning" %}}

Gateway Namespace Mode is currently an **alpha** feature. We recommend against using it in production workloads until it reaches beta status.

For status updates or to provide feedback, please follow our [GitHub issues](https://github.com/envoyproxy/gateway/issues).

{{% /alert %}}

# Overview

In standard deployment mode, Envoy Gateway creates all data plane resources in the controller namespace (typically `envoy-gateway-system`).

Gateway Namespace Mode changes this behavior by placing Envoy Proxy data plane resources like Deployments, Services and ServiceAccounts in each Gateway's namespace, providing stronger isolation and multi-tenancy.

Traditional deployment mode uses mTLS where both the client and server authenticate each other. However, in Gateway Namespace Mode, we've shifted to server-side TLS and JWT token validation between infra and control-plane.

* Only the CA certificate is available in pods running in Gateway namespaces
* Client certificates are not mounted in these namespaces
* The Envoy proxy still validates server certificates using the CA certificate

Gateway Namespace Mode uses projected service account JWT tokens for authentication.
* Use short-lived, audience-specific JWT tokens. These tokens are automatically mounted into pods via the projected volume mechanism
* JWT validation ensures that only authorized proxies can connect to the xDS server

{{% alert title="Note" color="warning" %}}

Currently it is not supported to run Gateway Namespace Mode with Merged Gateways deployments.

{{% /alert %}}

# Configuration

To enable Gateway Namespace Mode, configure the `provider.kubernetes.deploy.type` field in your Envoy Gateway ConfigMap:

```bash
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
metadata:
  name: envoy-gateway
  namespace: envoy-gateway-system
spec:
  provider:
    type: Kubernetes
    kubernetes:
      deploy:
        type: GatewayNamespace
```

To install Envoy Gateway with Gateway Namespace Mode using Helm:

```bash
helm install \
  --set config.envoyGateway.provider.kubernetes.deploy.type=GatewayNamespace \
  eg oci://docker.io/envoyproxy/gateway-helm \
  --version latest -n envoy-gateway-system --create-namespace
```

## RBAC configuration

When using Gateway Namespace Mode, Envoy Gateway needs additional RBAC permissions to create and manage resources across different namespaces. The following RBAC resources are automatically created when installing Envoy Gateway Helm Chart with Gateway Namespace Mode enabled.

```bash
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: gateway-helm-cluster-infra-manager
rules:
- apiGroups: [""]
  resources: ["serviceaccounts", "services", "configmaps"]
  verbs: ["create", "get", "delete", "deletecollection", "patch"]
- apiGroups: ["apps"]
  resources: ["deployments", "daemonsets"]
  verbs: ["create", "get", "delete", "deletecollection", "patch"]
- apiGroups: ["autoscaling", "policy"]
  resources: ["horizontalpodautoscalers", "poddisruptionbudgets"]
  verbs: ["create", "get", "delete", "deletecollection", "patch"]
- apiGroups: ["authentication.k8s.io"]
  resources: ["tokenreviews"]
  verbs: ["create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: gateway-helm-cluster-infra-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: 'gateway-helm-cluster-infra-manager'
subjects:
- kind: ServiceAccount
  name: 'envoy-gateway'
  namespace: 'envoy-gateway-system'
```

Envoy Gateway also supports configuration to you only watch resources in the specific namespaces by assigning
`EnvoyGateway.provider.kubernetes.watch.namespaces` or `EnvoyGateway.provider.kubernetes.watch.namespaceSelector`.
In this case, when you specify this configuration with Gateway Namespace Mode,Envoy Gateway will only watch for Gateway API resources in the specified namespaces and create needed Roles for infrastructure management in the specified namespaces.

# Using Gateway Namespace Mode

The following example demonstrates deploying two Gateways in different namespaces `team-a` and `team-b`.

## Create test namespaces

```shell
kubectl create namespace team-a
kubectl create namespace team-b
```

## Deploy Gateway Namespace Mode Example

Deploy resources on your cluster from the example, it will create two sets of backend deployments, Gateways and their respective HTTPRoutes in the previously created namespaces `team-a` and `team-b`.

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/gateway-namespace-mode.yaml
```

Verify that Gateways are deployed and programmed

```shell
kubectl get gateways -n team-a

NAME        CLASS   ADDRESS        PROGRAMMED   AGE
gateway-a   eg      172.18.0.200   True         67s
```

```shell
kubectl get gateways -n team-b

NAME        CLASS   ADDRESS        PROGRAMMED   AGE
gateway-b   eg      172.18.0.201   True         67s
```

Verify that HTTPRoutes are deployed

```shell
kubectl get httproute -n team-a

NAME           HOSTNAMES            AGE
team-a-route   ["www.team-a.com"]   67s
```

```shell
kubectl get httproute -n team-b

NAME           HOSTNAMES            AGE
team-b-route   ["www.team-b.com"]   67s
```

Envoy Proxy resources should be created now in the namespace of every Gateway.

```shell
kubectl get pods -n team-a

NAME                                              READY   STATUS    RESTARTS   AGE
envoy-team-a-gateway-a-b65c6264-d56f5d989-6dv5s   2/2     Running   0          65s
team-a-backend-6f786fb76f-nx26p                   1/1     Running   0          65s
```

```shell
kubectl get pods -n team-b

NAME                                               READY   STATUS    RESTARTS   AGE
envoy-team-b-gateway-b-0ac91f5a-74f445884f-95pl8   2/2     Running   0          87s
team-b-backend-966b5f47c-zxngl                     1/1     Running   0          87s
```

```shell
kubectl get services -n team-a

NAME                              TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)          AGE
envoy-team-a-gateway-a-b65c6264   LoadBalancer   10.96.191.198   172.18.0.200   8080:30999/TCP   3m2s
team-a-backend                    ClusterIP      10.96.92.226    <none>         3000/TCP         3m2s
```

```shell
kubectl get services -n team-b

NAME                              TYPE           CLUSTER-IP     EXTERNAL-IP    PORT(S)          AGE
envoy-team-b-gateway-b-0ac91f5a   LoadBalancer   10.96.144.13   172.18.0.201   8081:31683/TCP   3m43s
team-b-backend                    ClusterIP      10.96.26.162   <none>         3000/TCP         3m43s
```

## Testing the Configuration

Fetch external IPs of the services:

```shell
export GATEWAY_HOST_A=$(kubectl get gateway/gateway-a -n team-a -o jsonpath='{.status.addresses[0].value}')
```

```shell
export GATEWAY_HOST_B=$(kubectl get gateway/gateway-b -n team-b -o jsonpath='{.status.addresses[0].value}')
```

Curl the route team-a-route through Envoy proxy:

```shell
curl --header "Host: www.team-a.com" http://$GATEWAY_HOST_A:8080/example
```

```shell
{
 "path": "/example",
 "host": "www.team-a.com",
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
   "172.18.0.3"
  ],
  "X-Forwarded-For": [
   "172.18.0.3"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "52f08a5c-7e07-43b7-bd23-44693c60fc0c"
  ]
 },
 "namespace": "team-a",
 "ingress": "",
 "service": "",
 "pod": "team-a-backend-6f786fb76f-nx26p"
```

Curl the route team-b-route through Envoy proxy:

```shell
curl --header "Host: www.team-b.com" http://$GATEWAY_HOST_B:8081/example
```

```shell
{
 "path": "/example",
 "host": "www.team-b.com",
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
   "172.18.0.3"
  ],
  "X-Forwarded-For": [
   "172.18.0.3"
  ],
  "X-Forwarded-Proto": [
   "http"
  ],
  "X-Request-Id": [
   "62a06bd7-4754-475b-854a-dca3fc159e93"
  ]
 },
 "namespace": "team-b",
 "ingress": "",
 "service": "",
 "pod": "team-b-backend-966b5f47c-d6jwj"
```
