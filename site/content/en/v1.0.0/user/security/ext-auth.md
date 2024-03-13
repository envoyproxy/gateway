---
title: "External Authorization"
---

This guide provides instructions for configuring external authentication.

External authorization calls an external HTTP or gRPC service to check whether an incoming HTTP request is authorized
or not. If the request is deemed unauthorized, then the request will be denied with a 403 (Forbidden) response. If the
request is authorized, then the request will be allowed to proceed to the backend service. 

Envoy Gateway introduces a new CRD called [SecurityPolicy][SecurityPolicy] that allows the user to configure external authorization.
This instantiated resource can be linked to a [Gateway][Gateway] and [HTTPRoute][HTTPRoute] resource.

## Prerequisites

Follow the steps from the [Quickstart](../../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

Verify the Gateway status:

```shell
kubectl get gateway/eg -o yaml
```

## HTTP External Authorization Service

### Installation

Install a demo HTTP service that will be used as the external authorization service:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/ext-auth-http-service.yaml
```

Create a new HTTPRoute resource to route traffic on the path `/myapp` to the backend service.  

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: myapp
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /myapp
    backendRefs:
    - name: backend
      port: 3000   
EOF
```

Verify the HTTPRoute status:

```shell
kubectl get httproute/myapp -o yaml
```

### Configuration

Create a new SecurityPolicy resource to configure the external authorization. This SecurityPolicy targets the HTTPRoute
"myApp" created in the previous step. It calls the HTTP external authorization service "http-ext-auth" on port 9002 for
authorization. The `headersToBackend` field specifies the headers that will be sent to the backend service if the request
is successfully authorized.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: ext-auth-example
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: myapp
  extAuth:
    http:
      backendRef:
        name: http-ext-auth
        port: 9002
      headersToBackend: ["x-current-user"]
EOF
```

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/ext-auth-example -o yaml
```

### Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) guide is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Send a request to the backend service without `Authentication` header:

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/myapp"
```

You should see `403 Forbidden` in the response, indicating that the request is not allowed without authentication.

```shell
* Connected to 172.18.255.200 (172.18.255.200) port 80 (#0)
> GET /myapp HTTP/1.1
> Host: www.example.com
> User-Agent: curl/7.68.0
> Accept: */*
...
< HTTP/1.1 403 Forbidden
< date: Mon, 11 Mar 2024 03:41:15 GMT
< x-envoy-upstream-service-time: 0
< content-length: 0
< 
* Connection #0 to host 172.18.255.200 left intact
```

Send a request to the backend service with `Authentication` header:

```shell
curl -v -H "Host: www.example.com" -H "Authorization: Bearer token1" "http://${GATEWAY_HOST}/myapp"
```

The request should be allowed and you should see the response from the backend service. 
Because the `x-current-user` header from the auth response has been sent to the backend service, 
you should see the `x-current-user` header in the response.

```
"X-Current-User": [
   "user1"
  ],
```

## GRPC External Authorization Service

### Installation

Install a demo gRPC service that will be used as the external authorization service. The demo gRPC service is enabled 
with TLS and a BackendTLSConfig is created to configure the communication between the Envoy proxy and the gRPC service.

Note: TLS is optional for HTTP or gRPC external authorization services. However, enabling TLS is recommended for enhanced
security in production environments.

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/ext-auth-grpc-service.yaml
```

The HTTPRoute created in the previous section is still valid and can be used with the gRPC auth service, but if you have
not created the HTTPRoute, you can create it now.

Create a new HTTPRoute resource to route traffic on the path `/myapp` to the backend service.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: myapp
spec:
  parentRefs:
  - name: eg
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /myapp
    backendRefs:
    - name: backend
      port: 3000   
EOF
```

Verify the HTTPRoute status:

```shell
kubectl get httproute/myapp -o yaml
```

### Configuration

Update the SecurityPolicy that was created in the previous section to use the gRPC external authorization service.
It calls the gRPC external authorization service "grpc-ext-auth" on port 9002 for authorization. 

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: ext-auth-example
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: myapp
  extAuth:
    grpc:
      backendRef:
        name: grpc-ext-auth
        port: 9002
EOF
```

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/ext-auth-example -o yaml
```

Because the gRPC external authorization service is enabled with TLS, a BackendTLSConfig needs to be created to configure
the communication between the Envoy proxy and the gRPC auth service.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: BackendTLSPolicy
metadata:
  name: grpc-ext-auth-btls
spec:
  targetRef:
    group: ''
    kind: Service
    name: grpc-ext-auth
    sectionName: "9002"
  tls:
    caCertRefs:
    - name: grpc-ext-auth-ca
      group: ''
      kind: ConfigMap
    hostname: grpc-ext-auth
EOF
```

Verify the BackendTLSPolicy configuration:

```shell
kubectl get backendtlspolicy/grpc-ext-auth-btls -o yaml
```

### Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) guide is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Send a request to the backend service without `Authentication` header:

```shell
curl -v -H "Host: www.example.com" "http://${GATEWAY_HOST}/myapp"
```

You should see `403 Forbidden` in the response, indicating that the request is not allowed without authentication.

```shell
* Connected to 172.18.255.200 (172.18.255.200) port 80 (#0)
> GET /myapp HTTP/1.1
> Host: www.example.com
> User-Agent: curl/7.68.0
> Accept: */*
...
< HTTP/1.1 403 Forbidden
< date: Mon, 11 Mar 2024 03:41:15 GMT
< x-envoy-upstream-service-time: 0
< content-length: 0
< 
* Connection #0 to host 172.18.255.200 left intact
```

Send a request to the backend service with `Authentication` header:

```shell
curl -v -H "Host: www.example.com" -H "Authorization: Bearer token1" "http://${GATEWAY_HOST}/myapp"
```

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) guide to uninstall Envoy Gateway and the example manifest.

Delete the demo auth services, HTTPRoute, SecurityPolicy and BackendTLSPolicy:

```shell
kubectl delete -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/ext-auth-http-service.yaml
kubectl delete -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/ext-auth-grpc-service.yaml
kubectl delete httproute/myapp
kubectl delete securitypolicy/ext-auth-example
kubectl delete backendtlspolicy/grpc-ext-auth-btls
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop/) to get involved in the project.

[SecurityPolicy]: ../../contributions/design/security-policy/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
