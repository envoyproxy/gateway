---
title: "GeoIP Authorization"
---

This task provides instructions for configuring GeoIP-based authorization with Envoy Gateway.

GeoIP authorization uses geolocation data derived from the client IP address to determine whether a request should be
allowed or denied before it is forwarded to the backend service.

Envoy Gateway introduces a new CRD called [SecurityPolicy][] that allows the user to configure GeoIP-based authorization.
This instantiated resource can be linked to a [Gateway][], [HTTPRoute][], or [GRPCRoute][] resource.

GeoIP authorization is configured through `SecurityPolicy.spec.authorization.rules[].principal.clientIPGeoLocations`.

GeoIP authorization requires:

- GeoIP provider configuration on [EnvoyProxy][]
- client IP detection on [ClientTrafficPolicy][]
- a [SecurityPolicy][] attached to a [Gateway][], [HTTPRoute][] or [GRPCRoute][]

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

### Prepare the GeoIP database

Envoy reads GeoIP data from a local MaxMind `.mmdb` database file mounted into the proxy container.

This task uses a public MaxMind anonymous-IP test database. Apply the example manifest before continuing:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/{{< yaml-version >}}/examples/kubernetes/geoip-anonymous-ip-db.yaml
```

To keep this guide readable, the full base64-encoded `ConfigMap` is not repeated here.

For production deployments, mount your own supported MaxMind database and update the file path in the `EnvoyProxy` resource accordingly.

### Configure the Gateway and EnvoyProxy

The following resources create a dedicated `Gateway`, mount the anonymous-IP database into the Envoy proxy, and configure `EnvoyProxy.spec.geoIP.provider.maxMind.anonymousIpDbSource` to read it.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: geoip-authz-gateway
spec:
  gatewayClassName: eg
  infrastructure:
    parametersRef:
      group: gateway.envoyproxy.io
      kind: EnvoyProxy
      name: geoip-authz-proxy
  listeners:
  - name: http
    port: 80
    protocol: HTTP
    allowedRoutes:
      namespaces:
        from: Same
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: geoip-authz-proxy
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        pod:
          volumes:
          - name: geoip-db
            configMap:
              name: geoip-anonymous-ip-db
              items:
              - key: GeoIP2-Anonymous-IP-Test.mmdb
                path: GeoIP2-Anonymous-IP-Test.mmdb
        container:
          volumeMounts:
          - name: geoip-db
            mountPath: /etc/envoy/geoip
            readOnly: true
  geoIP:
    provider:
      type: MaxMind
      maxMind:
        anonymousIpDbSource:
          local:
            path: /etc/envoy/geoip/GeoIP2-Anonymous-IP-Test.mmdb
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resources to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: geoip-authz-gateway
spec:
  gatewayClassName: eg
  infrastructure:
    parametersRef:
      group: gateway.envoyproxy.io
      kind: EnvoyProxy
      name: geoip-authz-proxy
  listeners:
  - name: http
    port: 80
    protocol: HTTP
    allowedRoutes:
      namespaces:
        from: Same
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: geoip-authz-proxy
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        pod:
          volumes:
          - name: geoip-db
            configMap:
              name: geoip-anonymous-ip-db
              items:
              - key: GeoIP2-Anonymous-IP-Test.mmdb
                path: GeoIP2-Anonymous-IP-Test.mmdb
        container:
          volumeMounts:
          - name: geoip-db
            mountPath: /etc/envoy/geoip
            readOnly: true
  geoIP:
    provider:
      type: MaxMind
      maxMind:
        anonymousIpDbSource:
          local:
            path: /etc/envoy/geoip/GeoIP2-Anonymous-IP-Test.mmdb
```

{{% /tab %}}
{{< /tabpane >}}

### Create the route and SecurityPolicy

The following resources create an `HTTPRoute` for `/geo-anonymous` and attach a `SecurityPolicy` that denies requests identified as anonymous networks while allowing other traffic.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-with-authorization-geoip-anonymous
spec:
  parentRefs:
  - name: geoip-authz-gateway
  rules:
  - matches:
    - path:
        type: Exact
        value: /geo-anonymous
    backendRefs:
    - name: backend
      port: 3000
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: authorization-geoip-anonymous
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-with-authorization-geoip-anonymous
  authorization:
    defaultAction: Allow
    rules:
    - action: Deny
      principal:
        clientIPGeoLocations:
        - anonymous:
            isAnonymous: true
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
  name: http-with-authorization-geoip-anonymous
spec:
  parentRefs:
  - name: geoip-authz-gateway
  rules:
  - matches:
    - path:
        type: Exact
        value: /geo-anonymous
    backendRefs:
    - name: backend
      port: 3000
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: authorization-geoip-anonymous
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: http-with-authorization-geoip-anonymous
  authorization:
    defaultAction: Allow
    rules:
    - action: Deny
      principal:
        clientIPGeoLocations:
        - anonymous:
            isAnonymous: true
```

{{% /tab %}}
{{< /tabpane >}}

Verify the `SecurityPolicy` configuration:

```shell
kubectl get securitypolicy/authorization-geoip-anonymous -o yaml
```

### Enable client IP detection

GeoIP authorization depends on Envoy Gateway correctly detecting the client IP address. Without `ClientTrafficPolicy.spec.clientIPDetection`, the `clientIPGeoLocations` match will not work as intended.

The following `ClientTrafficPolicy` tells Envoy Gateway to use the `X-Forwarded-For` header and trust one upstream hop:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-client-ip-detection-geoip
spec:
  clientIPDetection:
    xForwardedFor:
      numTrustedHops: 1
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: geoip-authz-gateway
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-client-ip-detection-geoip
spec:
  clientIPDetection:
    xForwardedFor:
      numTrustedHops: 1
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: geoip-authz-gateway
```

{{% /tab %}}
{{< /tabpane >}}

Verify the `ClientTrafficPolicy` configuration:

```shell
kubectl get clienttrafficpolicy/enable-client-ip-detection-geoip -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Send a request with a regular client IP that is not flagged as anonymous by the test database:

```shell
curl -v -H "Host: www.example.com" -H "X-Forwarded-For: 8.8.8.8" "http://${GATEWAY_HOST}/geo-anonymous"
```

The request should be allowed and return `200 OK`.

Send a request with an IP that the anonymous-IP test database marks as anonymous:

```shell
curl -v -H "Host: www.example.com" -H "X-Forwarded-For: 6.1.0.3" "http://${GATEWAY_HOST}/geo-anonymous"
```

The request should be denied and return `403 Forbidden`.

## Clean-Up

Remove the resources created in this task:

```shell
kubectl delete clienttrafficpolicy/enable-client-ip-detection-geoip
kubectl delete securitypolicy/authorization-geoip-anonymous
kubectl delete httproute/http-with-authorization-geoip-anonymous
kubectl delete gateway/geoip-authz-gateway
kubectl delete envoyproxy/geoip-authz-proxy
```

If you applied the test GeoIP database `ConfigMap`, remove it as well:

```shell
kubectl delete configmap/geoip-anonymous-ip-db
```

## Next Steps

Checkout the following related guides:

- [IP Allowlist/Denylist](restrict-ip-access/)
- [SecurityPolicy API Reference](../../api/extension_types#securitypolicy)
- [ClientTrafficPolicy API Reference](../../api/extension_types#clienttrafficpolicy)

[SecurityPolicy]: ../../api/extension_types#securitypolicy
[EnvoyProxy]: ../../api/extension_types#envoyproxy
[ClientTrafficPolicy]: ../../api/extension_types#clienttrafficpolicy
[Gateway]: https://gateway-api.sigs.k8s.io/reference/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/httproute/
[GRPCRoute]: https://gateway-api.sigs.k8s.io/reference/api-types/grpcroute/
