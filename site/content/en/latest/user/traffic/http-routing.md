---
title: "HTTP Routing"
---

The [HTTPRoute][] resource allows users to configure HTTP routing by matching HTTP traffic and forwarding it to
Kubernetes backends. Currently, the only supported backend supported by Envoy Gateway is a Service resource. This guide
shows how to route traffic based on host, header, and path fields and forward the traffic to different Kubernetes
Services. To learn more about HTTP routing, refer to the [Gateway API documentation][].

## Prerequisites

Follow the steps from the [Quickstart](../../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Installation

Install the HTTP routing example resources:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/http-routing.yaml
```

The manifest installs a [GatewayClass][], [Gateway][], four Deployments, four Services, and three HTTPRoute resources.
The GatewayClass is a cluster-scoped resource that represents a class of Gateways that can be instantiated.

__Note:__ Envoy Gateway is configured by default to manage a GatewayClass with
`controllerName: gateway.envoyproxy.io/gatewayclass-controller`.

## Verification

Check the status of the GatewayClass:

```shell
kubectl get gc --selector=example=http-routing
```

The status should reflect "Accepted=True", indicating Envoy Gateway is managing the GatewayClass.

A Gateway represents configuration of infrastructure. When a Gateway is created, [Envoy proxy][] infrastructure is
provisioned or configured by Envoy Gateway. The `gatewayClassName` defines the name of a GatewayClass used by this
Gateway. Check the status of the Gateway:

```shell
kubectl get gateways --selector=example=http-routing
```

The status should reflect "Ready=True", indicating the Envoy proxy infrastructure has been provisioned. The status also
provides the address of the Gateway. This address is used later in the guide to test connectivity to proxied backend
services.

The three HTTPRoute resources create routing rules on the Gateway. In order to receive traffic from a Gateway,
an HTTPRoute must be configured with `parentRefs` which reference the parent Gateway(s) that it should be attached to.
An HTTPRoute can match against a [single set of hostnames][spec]. These hostnames are matched before any other matching
within the HTTPRoute takes place. Since `example.com`, `foo.example.com`, and `bar.example.com` are separate hosts with
different routing requirements, each is deployed as its own HTTPRoute - `example-route, ``foo-route`, and `bar-route`.

Check the status of the HTTPRoutes:

```shell
kubectl get httproutes --selector=example=http-routing -o yaml
```

The status for each HTTPRoute should surface "Accepted=True" and a `parentRef` that references the example Gateway.
The `example-route` matches any traffic for "example.com" and forwards it to the "example-svc" Service.

## Testing the Configuration

Before testing HTTP routing to the `example-svc` backend, get the Gateway's address.

```shell
export GATEWAY_HOST=$(kubectl get gateway/example-gateway -o jsonpath='{.status.addresses[0].value}')
```

Test HTTP routing to the `example-svc` backend.

```shell
curl -vvv --header "Host: example.com" "http://${GATEWAY_HOST}/"
```

A `200` status code should be returned and the body should include `"pod": "example-backend-*"` indicating the traffic
was routed to the example backend service. If you change the hostname to a hostname not represented in any of the
HTTPRoutes, e.g. "www.example.com", the HTTP traffic will not be routed and a `404` should be returned.

The `foo-route` matches any traffic for `foo.example.com` and applies its routing rules to forward the traffic to the
"foo-svc" Service. Since there is only one path prefix match for `/login`, only `foo.example.com/login/*` traffic will
be forwarded. Test HTTP routing to the `foo-svc` backend.

```shell
curl -vvv --header "Host: foo.example.com" "http://${GATEWAY_HOST}/login"
```

A `200` status code should be returned and the body should include `"pod": "foo-backend-*"` indicating the traffic
was routed to the foo backend service. Traffic to any other paths that do not begin with `/login` will not be matched by
this HTTPRoute. Test this by removing `/login` from the request.

```shell
curl -vvv --header "Host: foo.example.com" "http://${GATEWAY_HOST}/"
```

The HTTP traffic will not be routed and a `404` should be returned.

Similarly, the `bar-route` HTTPRoute matches traffic for `bar.example.com`. All traffic for this hostname will be
evaluated against the routing rules. The most specific match will take precedence which means that any traffic with the
`env:canary` header will be forwarded to `bar-svc-canary` and if the header is missing or not `canary` then it'll be
forwarded to `bar-svc`. Test HTTP routing to the `bar-svc` backend.

```shell
curl -vvv --header "Host: bar.example.com" "http://${GATEWAY_HOST}/"
```

A `200` status code should be returned and the body should include `"pod": "bar-backend-*"` indicating the traffic
was routed to the foo backend service.

Test HTTP routing to the `bar-canary-svc` backend by adding the `env: canary` header to the request.

```shell
curl -vvv --header "Host: bar.example.com" --header "env: canary" "http://${GATEWAY_HOST}/"
```

A `200` status code should be returned and the body should include `"pod": "bar-canary-backend-*"` indicating the
traffic was routed to the foo backend service.

### JWT Claims Based Routing

Users can route to a specific backend by matching on JWT claims.
This can be achieved, by defining a SecurityPolicy with a jwt configuration that does the following
* Converts jwt claims to headers, which can be used for header based routing
* Sets the recomputeRoute field to `true`. This is required so that the incoming request matches on a fallback/catch all route where the JWT can be authenticated, the claims from the JWT can be converted to headers, and then the route match can be recomputed to match based on the updated headers.

For this feature to work please make sure
* you have a fallback route rule defined, the backend for this route rule can be invalid.
* The SecurityPolicy is applied to both the fallback route as well as the route with the claim header matches, to avoid spoofing.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: jwt-example
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: jwt-claim-routing
  jwt:
    providers:
      - name: example
        recomputeRoute: true
        claimToHeaders:
          - claim: sub
            header: x-sub
          - claim: admin
            header: x-admin
          - claim: name
            header: x-name
        remoteJWKS:
          uri: https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/jwks.json
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: jwt-claim-routing
spec:
  parentRefs:
    - name: eg
  rules:
    - backendRefs:
        - kind: Service
          name: foo-svc
          port: 8080
          weight: 1
      matches:
        - headers:
            - name: x-name
              value: John Doe
    - backendRefs:
        - kind: Service
          name: bar-svc
          port: 8080
          weight: 1
      matches:
        - headers:
            - name: x-name
              value: Tom
    # catch all
    - backendRefs:
        - kind: Service
          name: infra-backend-invalid
          port: 8080
          weight: 1
      matches:
        - path:
            type: PathPrefix
            value: /
EOF
```

Get the JWT used for testing request authentication:

```shell
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/test.jwt -s) && echo "$TOKEN" | cut -d '.' -f2 - | base64 --decode -
```

Test routing to the `foo-svc` backend by specifying a JWT Token with a claim `name: John Doe`.

```shell
curl -sS -H "Authorization: Bearer $TOKEN" "http://${GATEWAY_HOST}/" | jq .pod
"foo-backend-6df8cc6b9f-fmwcg"
```

Get another JWT used for testing request authentication:

```shell
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/with-different-claim.jwt -s) && echo "$TOKEN" | cut -d '.' -f2 - | base64 --decode -
```

Test HTTP routing to the `bar-svc` backenbackend by specifying a JWT Token with a claim `name: Tom`.

```shell
curl -sS -H "Authorization: Bearer $TOKEN" "http://${GATEWAY_HOST}/" | jq .pod
"bar-backend-6688b8944c-s8htr"
```

[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[Gateway API documentation]: https://gateway-api.sigs.k8s.io/
[GatewayClass]: https://gateway-api.sigs.k8s.io/api-types/gatewayclass/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[Envoy proxy]: https://www.envoyproxy.io/
[spec]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteSpec
