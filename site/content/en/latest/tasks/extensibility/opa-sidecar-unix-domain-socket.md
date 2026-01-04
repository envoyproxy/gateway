---
title: "OPA Sidecar with Unix Domain Socket"
---

This task demonstrates how to deploy an Open Policy Agent (OPA) sidecar inside Envoy Proxy pods created by Envoy Gateway, with Envoy Proxy communicating with OPA over a Unix Domain Socket.

The setup provides these advantages over running OPA as its own Kubernetes service:

- Using a sidecar ensures that the OPA is always on the same worker node as Envoy Proxy. This reduces latency and simplifies access control.
- Using a Unix Domain Socket avoids TCP/TLS overhead while improving security by avoiding network exposure and reduces latency.

This production-ready approach uses Envoy Gateway's extension APIs to configure Envoy Proxy's external authorization and standard Kubernetes API objects to configure OPA's Envoy External Authorization gRPC plugin.

```mermaid
graph TB
    subgraph K8sPod ["⎈ Envoy Proxy Pod"]
        subgraph EnvoyContainer ["Envoy Proxy Container"]
            ExtAuthz("🔐 ext_authz Filter")
        end

        UnixSocket[("Unix Domain Socket<br/>/shared/opa/opa.sock")]

        subgraph OPAContainer ["OPA Container"]
            RegoPolicy("⚙️ Rego Policy")
        end
    end

    %% Flow Definitions
    ExtAuthz <-->|"Authz check"| UnixSocket
    UnixSocket <-->|"Authz check"| RegoPolicy

    style UnixSocket fill:#D0D0D0
    style EnvoyContainer fill:#2962FF,color:#FFFFFF
    style OPAContainer fill:#2962FF,color:#FFFFFF
```

Note: In this diagram, the Unix Domain Socket is shown outside of the Proxy and OPA containers, but it is actually mounted to both containers.

The example flow on this page will use Envoy to authenticate JWTs and OPA to validate the ISS claim on the JWT. If the ISS matches a configured env var value then requests will be allowed.

Note that the OPA sidecar with Unix Domain Sockets pattern does _not_ require that JWTs be used. Also, simple ISS claim validation could be done more easily using the "authentication" property of the `SecurityPolicy` spec, but using OPA allows for more complex logic to be applied. See [JWT Claim-Based Authorization](../security/jwt-claim-authorization.md) to see how to use the `SecurityPolicy` "authentication" property.

## Prerequisites

{{< boilerplate prerequisites >}}

## Steps overview

1. Enable the `Backend` extension API so Envoy can reference a Unix Domain Socket.
2. Create an OPA policy in a `ConfigMap` and mount it into an OPA sidecar in the Envoy pods.
3. Configure a `SecurityPolicy` that delegates authorization to OPA via the Unix Domain Socket `Backend`. For demonstration purposes we'll also configure JWT authentication.
4. Create an HTTPRoute that requires protection and verify traffic flow.

## Step 1: Enable the Backend extension API

Enable the `Backend` API in your Envoy Gateway `ConfigMap` to allow Envoy to reference Unix Domain Sockets by setting the "enableBackend" property to "true" in the "envoy-gateway-config" `ConfigMap`:

```yaml
data:
  envoy-gateway.yaml: |
    ...
    extensionApis:
      enableBackend: true
```

Run these commands to update the `ConfigMap`:

```shell
OLD_CONFIG=$(kubectl get cm envoy-gateway-config -n envoy-gateway-system -o jsonpath='{.data.envoy-gateway\.yaml}')
NEW_CONFIG=$(echo "$OLD_CONFIG" | sed "s/extensionApis: {}/extensionApis:\n  enableBackend: true/")
kubectl create configmap envoy-gateway-config -n envoy-gateway-system \
  --from-literal=envoy-gateway.yaml="$NEW_CONFIG" \
  --dry-run=client -o yaml | kubectl apply -f -
```

Ignore the kubectl warning. It happens because `kubectl apply` expects a "receipt" (the `last-applied-configuration` annotation) to be present on the resource so it can track changes over time. If the ConfigMap was originally created using `kubectl create` (without `--save-config`) or by a Helm chart, that receipt is missing. `kubectl` is simply telling you that it's adding that receipt now.

Restart the Envoy Gateway deployment to apply the updated configuration:

```shell
kubectl -n envoy-gateway-system rollout restart deploy/envoy-gateway
```

The Envoy Proxy pod will be restarted later in this guide.

## Step 2: Create the OPA policy

Create a `ConfigMap` with a [Rego](https://www.openpolicyagent.org/docs/policy-language) policy that OPA will load. This example reads the "ALLOWED_ISS" environment variable and compares it to the JWT ISS claim, allowing the request only if they match.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: opa-policy
  namespace: envoy-gateway-system
data:
  policy.rego: |
    package envoy.authz

    import rego.v1

    default allow := false

    opa_env := opa.runtime().env
    allowed_iss := opa_env.ALLOWED_ISS
    jwt_iss := input.attributes.request.http.headers["x-jwt-iss"]

    allow if {
        jwt_iss == allowed_iss
    }
```

On MacOS you can apply that by copying it and running `pbpaste | kubectl apply -f -`.

In this policy we extract a `x-jwt-iss` header from the request. That header is not automatically available on `ext_authz` filter requests. We will configure Envoy to add that header from the decoded JWT. That configuration happens in the SecurityPolicy in step 4.

## Step 3: Configure the Envoy Proxy pods with an OPA sidecar

Create an `EnvoyProxy` resource that mounts the OPA policy `ConfigMap`, shares a Unix Domain Socket between Envoy and OPA, and configures the "ext_authz" filter order so OPA receives JWT headers from the JWT authentication filter.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: eg
  namespace: envoy-gateway-system
spec:
  # Ensure ext_authz runs after jwt_authn so that OPA sees the JWT-derived headers
  # configured in the SecurityPolicy in step 4
  filterOrder:
    - name: envoy.filters.http.ext_authz
      after: envoy.filters.http.jwt_authn
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        patch:
          type: StrategicMerge
          value:
            spec:
              template:
                spec:
                  volumes:
                    # Shared Unix Domain Socket volume for communication between Envoy
                    # and OPA. This will be mounted in both containers.
                    - name: opa-unix-socket
                      emptyDir: {}
                    # OPA Rego policy file from ConfigMap. This will be mounted
                    # in the OPA container.
                    - name: opa-policy
                      configMap:
                        name: opa-policy
                  containers:
                    # Add the Unox domain socket volume to the existing envoy
                    # proxy container named "envoy".
                    - name: envoy
                      volumeMounts:
                        # Mount the Unix Domain Socket volume in the Envoy container
                        - name: opa-unix-socket
                          mountPath: /shared/opa
                    # Create the OPA sidecar container
                    - name: opa
                      # OPA image tag format:
                      # - "x.y.z": OPA version
                      # - "envoy": includes OPA's Envoy External Authorization gRPC plugin
                      # - "5": plugin version
                      # - "static": static build (no glibc dependency, smaller image)
                      image: openpolicyagent/opa:1.10.1-envoy-5-static
                      args:
                        - "run"
                        - "--server"
                        # In addition to listening on the Unix Domain Socket, we'll listen
                        # on port 8181 for k8s probes
                        - "--addr=0.0.0.0:8181"
                        # Configure OPA's Envoy external authorization plugin to listen on a Unix Domain Socket
                        - "--set=plugins.envoy_ext_authz_grpc.addr=unix:///shared/opa/opa.sock"
                        # In the OPA policy we defined the policy as being in a
                        # "envoy.authz" package and set the "allow" variable
                        # based on the policy decision. Here we tell the plugin
                        # to check that value.
                        - "--set=plugins.envoy_ext_authz_grpc.query=data.envoy.authz.allow"
                        - "--ignore=.*"
                        # You can set log-level to "debug" for detailed request logging
                        - "--log-level=error"
                        # Tell OPA where to find the policy file
                        - "/policy/policy.rego"
                      env:
                        # Environment variable used by the OPA policy to validate the JWT issuer
                        - name: ALLOWED_ISS
                          value: "https://example.com"
                      ports:
                        # Port for health checks only; gRPC server listens on the Unix Domain Socket
                        - name: http
                          containerPort: 8181
                          protocol: TCP
                      volumeMounts:
                        # Mount the Unix Domain Socket volume in the OPA container
                        - name: opa-unix-socket
                          mountPath: /shared/opa
                        # Mount the OPA Rego policy from ConfigMap
                        - name: opa-policy
                          mountPath: /policy
                          readOnly: true
                      livenessProbe:
                        httpGet:
                          path: /health
                          port: 8181
                        initialDelaySeconds: 5
                        periodSeconds: 10
                        timeoutSeconds: 3
                        failureThreshold: 3
                      readinessProbe:
                        httpGet:
                          path: /health?bundle=true
                          port: 8181
                        initialDelaySeconds: 5
                        periodSeconds: 5
                        timeoutSeconds: 3
                        failureThreshold: 3
                      securityContext:
                        # Use the same user ID as the Envoy container (65532) so both containers can access the socket
                        # The user ID was found by running "docker pull docker.io/envoyproxy/gateway:v<version> && docker inspect docker.io/envoyproxy/gateway:v<version>"
                        runAsUser: 65532
                        allowPrivilegeEscalation: false
                      resources:
                        # Adjust as necessary
                        requests:
                          cpu: 100m
                          memory: 128Mi
                        limits:
                          memory: 512Mi
```

This patch adds an OPA sidecar container to the Envoy pods, creates a shared Unix Domain Socket at `/shared/opa/opa.sock`, and configures OPA to listen only on this socket instead of TCP. It sets an `ALLOWED_ISS` property that OPA will use to validate the request. We purposefully set the wrong value here so we can see the response when the wrong ISS is used. We'll set it to the correct value later.

Tell Envoy Gateway to use this `EnvoyProxy` spec by patching the `GatewayClass`.

```shell
kubectl patch gatewayclass eg --type merge -p '
{
  "spec": {
    "parametersRef": {
      "group": "gateway.envoyproxy.io",
      "kind": "EnvoyProxy",
      "name": "eg",
      "namespace": "envoy-gateway-system"
    }
  }
}'
```

This will cause the Envoy Proxy pod to be replaced with the new config.

## Step 4: Connect Envoy to OPA with a SecurityPolicy and Backend

Create a Unix Domain Socket `SecurityPolicy` and a `Backend` that authenticates JWTs, maps the ISS claims to the `x-jwt-iss` header, and delegates authorization to OPA.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: opa-jwt-authz
  # The SecurityPolicy must be in the same namespace as the HTTPRoutes that it
  # will apply to
  namespace: default
spec:
  targetSelectors:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      matchLabels:
        auth: opa-jwt
  jwt:
    providers:
      - name: my-jwt-validator
        # Verify that the JWT is signed with the correct key
        remoteJWKS:
          uri: https://raw.githubusercontent.com/envoyproxy/gateway/refs/heads/main/examples/kubernetes/jwt/jwks.json
        # Add the JWT ISS claim to request to the headers sent to OPA. Note: this only works
        # for string claims, not arrays. For example, the "aud" claim cannot be passed
        # this way because the JWT spec defines it as an array.
        claimToHeaders:
          - header: x-jwt-iss
            claim: iss
  # Delegate authorization to an external program after JWT verification
  extAuth:
    grpc:
      backendRefs:
        # Reference the OPA Unix Domain Socket backend, which is the next API object in this code block
        - group: gateway.envoyproxy.io
          kind: Backend
          name: opa-unix-socket
    # Headers to send to OPA for authorization decisions. This should match the
    # "claimToHeaders" values above.
    headersToExtAuth:
      - x-jwt-iss
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: opa-unix-socket
  # The Backend must be in the same namespace as the SecurityPolicy
  namespace: default
spec:
  endpoints:
    # This configures Envoy proxy to send external auth requests to the Unix Domain Socket
    - unix:
        path: /shared/opa/opa.sock
```

## Step 5: Configure route to use the OPA authorization

Attach the policy by labeling routes that require authorization. This `HTTPRoute` uses the `backend` service installed in the Quick Start:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: test-opa
  namespace: default
  labels:
    auth: opa-jwt
spec:
  hostnames:
  - foo.bar.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /test-opa
    backendRefs:
    - name: backend
      port: 3000
```

## Step 5: Test the configuration

Prepare some env vars:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
export VALID_TOKEN="eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImI1MjBiM2MyYzRiZDc1YTEwZTljZWJjOTU3NjkzM2RjIn0.eyJpc3MiOiJodHRwczovL2Zvby5iYXIuY29tIiwic3ViIjoiMTIzNDU2Nzg5MCIsInVzZXIiOnsibmFtZSI6IkpvaG4gRG9lIiwiZW1haWwiOiJqb2huLmRvZUBleGFtcGxlLmNvbSIsInJvbGVzIjpbImFkbWluIiwiZWRpdG9yIl19LCJwcmVtaXVtX3VzZXIiOnRydWUsImlhdCI6MTUxNjIzOTAyMiwic2NvcGUiOiJyZWFkIGFkZCBkZWxldGUgbW9kaWZ5In0.P36iAlmiRCC79OiB3vstF5Q_9OqUYAMGF3a3H492GlojbV6DcuOz8YIEYGsRSWc-BNJaBKlyvUKsKsGVPtYbbF8ajwZTs64wyO-zhd2R8riPkg_HsW7iwGswV12f5iVRpfQ4AG2owmdOToIaoch0aym89He1ZzEjcShr9olgqlAbbmhnk-namd1rP-xpzPnWhhIVI3mCz5hYYgDTMcM7qbokM5FzFttTRXAn5_Luor23U1062Ct_K53QArwxBvwJ-QYiqcBycHf-hh6sMx_941cUswrZucCpa-EwA3piATf9PKAyeeWHfHV9X-y8ipGOFg3mYMMVBuUZ1lBkJCik9f9kboRY6QzpOISARQj9PKMXfxZdIPNuGmA7msSNAXQgqkvbx04jMwb9U7eCEdGZztH4C8LhlRjgj0ZdD7eNbRjeH2F6zrWyMUpGWaWyq6rMuP98W2DWM5ZflK6qvT1c7FuFsWPvWLkgxQwTWQKrHdKwdbsu32Sj8VtUBJ0-ddEb"
```

If Envoy is not running behind a load balancer, then you'll have to start port forwarding whenever the Envoy Proxy pod is reconfigured and the proxy pods are replaced:

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8888:80 &
```

### Test without a JWT

Requests without a valid JWT are denied:

```shell
curl -v -H "Host: foo.bar.com" "http://${GATEWAY_HOST}/test-opa/get"
```

Expected result: HTTP 401 Unauthorized

### Test with a valid JWT that has the wrong ISS claim

Earlier, we configured the "ALLOWED_ISS" env var on the proxy pods to be "https://example.com", but the JWT in the "VALID_TOKEN" env var has an ISS claim of "https://foo.bar.com".

Requests with an incorrect ISS are denied:

```shell
curl -v -H "Host: foo.bar.com" -H "Authorization: Bearer $VALID_TOKEN" "http://${GATEWAY_HOST}/test-opa/get"
```

Expected result: HTTP 403 Forbidden

### Test with a valid JWT and ISS claim

We must update the `EnvoyProxy` spec to use "https://foo.bar.com" for the "ALLOWED_ISS" so that it matches the ISS in "$VALID_TOKEN".

```yaml
...
                  containers:
                    - name: opa
                      env:
                        - name: ALLOWED_ISS
                          value: "https://foo.bar.com" # Change to this
...
```

The quickest way to do that for this tutorial is via `kubectl -n envoy-gateway-system edit envoyproxy eg`.

Wait for the Envoy Proxy pod to be replaced.

When the JWT issuer matches your configured provider and the OPA policy allows it, the request is forwarded to the backend:

```shell
curl -v -H "Host: foo.bar.com" -H "Authorization: Bearer $VALID_TOKEN" "http://${GATEWAY_HOST}/test-opa/get"
```

Expected result: HTTP 200 and a response from the backend service.

## Clean up

Delete all resources created by this task:

```shell
kubectl delete httproute/test-opa securitypolicy/opa-jwt-authz backend/opa-unix-socket
kubectl delete envoyproxy/eg configmap/opa-policy -n envoy-gateway-system
kubectl delete deploy/$(kubectl get deploy -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}') -n envoy-gateway-system
```