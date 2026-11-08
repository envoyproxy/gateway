---
title: 'Remote Infrastructure Provider'
---

This task explains how to configure Envoy Gateway to defer Envoy proxy and rate
limit infrastructure management to a remote service of your own. With the Remote
infrastructure provider, Envoy Gateway translates Gateway API resources into its
into an infrastructure IR format and forwards it over gRPC to a provider you
operate. The provider is then responsible for reconciling the data plane.

This is the right choice when the built-in Kubernetes infrastructure provider
doesn't fit your environment. Common cases include running proxies in a
different cluster from the control plane, deploying onto a non-Kubernetes
substrate (VMs, an internal scheduler, etc.), or imposing custom organizational
rules on how proxy workloads are shaped.

**Note:** Most users should use the built-in Kubernetes infrastructure provider.
The Remote provider is an extension point; running it means owning the lifecycle
of every proxy and rate limit deployment yourself.

## Security Warning

{{% alert title="Security Warning" color="warning" %}} The remote infrastructure
provider is a privileged component. It receives the complete Infrastructure IR
for every Gateway managed by Envoy Gateway and is trusted to faithfully
reconcile it. A compromised or misconfigured provider can stop creating proxies,
deploy proxies with attacker-controlled configuration, or expose the data plane
to networks it shouldn't reach.

When deploying a remote infrastructure provider, you should:

- Run the provider as a sidecar of `envoy-gateway`, or otherwise on a private
  channel that is not reachable from untrusted networks. The reference
  deployment uses a Unix domain socket on a shared volume.
- Restrict the provider's permissions to the minimum needed to manage the proxy
  data plane.
- Audit the provider on the same cadence as Envoy Gateway itself; bumping Envoy
  Gateway may change the IR shape. {{% /alert %}}

## How it works

When the Remote provider is enabled, Envoy Gateway defers infrastructure
management using an gRPC client to send updates to your provider. For every
reconcile that mutates a proxy or rate limit deployment, Envoy Gateway issues a
corresponding RPC carrying the IR as structured protobuf data.

The service contract is defined in `proto/remoteinfra/service.proto`:

```proto
service EnvoyGatewayRemoteInfrastructureProvider {
    rpc CreateOrUpdateProxyInfra(CreateOrUpdateProxyInfraRequest) returns (CreateOrUpdateProxyInfraResponse);
    rpc DeleteProxyInfra(DeleteProxyInfraRequest) returns (DeleteProxyInfraResponse);
    rpc CreateOrUpdateRateLimitInfra(CreateOrUpdateRateLimitInfraRequest) returns (CreateOrUpdateRateLimitInfraResponse);
    rpc DeleteRateLimitInfra(DeleteRateLimitInfraRequest) returns (DeleteRateLimitInfraResponse);
}
```

The proxy RPCs carry the IR as a structured `Infra` message that mirrors
`internal/ir/infra.go`, including proxy metadata (labels, annotations, and the
owner reference) and each metric sink's `destination` (endpoints and upstream
TLS). Most fields are typed protobuf fields; the proxy `config` (an `EnvoyProxy`
resource) is the only field carried as JSON-encoded `bytes`, because that CRD
schema is large and evolves independently of this contract. The rate limit RPCs
are parameterless. The provider is expected to be idempotent — Envoy Gateway
calls the `CreateOrUpdate` RPCs every time the desired state changes and may
retry on transient errors.

## Configuration

The Remote provider is selected under `provider.custom.infrastructure` in the
`EnvoyGateway` config. The `service` field is required and uses the standard
[ExtensionService] shape, so you can reach the provider over a Unix socket, a
hostname, or an IP.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
provider:
  type: Custom
  custom:
    resource:
      type: Kubernetes
    infrastructure:
      type: Remote
      remote:
        service:
          unix:
            path: /var/run/remote-infra/server.sock
```

To dial a remote service over the network instead, use `fqdn` or `ip`:

```yaml
service:
  fqdn:
    hostname: remote-infra.envoy-gateway-system.svc.cluster.local
    port: 5005
  tls:
    certificateRef:
      name: remote-infra-ca
      namespace: envoy-gateway-system
```

When TLS is configured, Envoy Gateway verifies the server certificate against
the CA in the referenced Secret. Add `clientCertificateRef` if your provider
requires mTLS.

### Rate limit service address

The Remote provider tells Envoy Gateway _what_ rate limit infrastructure to
stand up via the `CreateOrUpdateRateLimitInfra` RPC, but Envoy proxies still
need to know _where_ to send rate limit checks. Set `rateLimit.url` on the
`EnvoyGateway` config to the address your provider exposes. The value must be a
`grpc://` URL with an explicit host and port:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
rateLimit:
  url: grpc://my-hosted-rl-service.com:443
```

If `rateLimit.url` is unset, Envoy Gateway falls back to the address of the
Service the Kubernetes infrastructure manager would have created.

## Quickstart

This walks through deploying the example provider in `examples/remote-infra` as
a native sidecar of the `envoy-gateway` Deployment, communicating over a Unix
domain socket. The example provider reconciles a `Deployment` and a `Service`
per Gateway and is intended as a starting point — not a production-ready
provider.

### Prerequisites

{{< boilerplate prerequisites >}}

### Build and load the example provider image

The example lives in `examples/remote-infra/` and ships with a Makefile target
that builds a multi-arch image tagged `envoyproxy/gateway-remote-infra:latest`.

```shell
make -C examples/remote-infra docker-buildx
```

If you're running on `kind`, load the image into the cluster so the sidecar can
pull it:

```shell
kind load docker-image --name envoy-gateway envoyproxy/gateway-remote-infra:latest
```

### Inject the provider as a sidecar

The provider needs to share a volume with `envoy-gateway` so they can talk over
a UDS, and it needs to start before `envoy-gateway` does. Native sidecars give
us both for free on Kubernetes 1.29+.

```shell
cat <<'EOF' | kubectl patch deployment envoy-gateway -n envoy-gateway-system --patch-file /dev/stdin
apiVersion: apps/v1
kind: Deployment
metadata:
  name: envoy-gateway
  namespace: envoy-gateway-system
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
        runAsGroup: 65532
        # fsGroup ensures the shared emptyDir is group-owned by 65532 so the
        # non-root sidecar can create the UDS socket file in it.
        fsGroup: 65532
      volumes:
        - name: remote-infra-socket
          emptyDir: {}
      containers:
        - name: envoy-gateway
          volumeMounts:
            - name: remote-infra-socket
              mountPath: /var/run/remote-infra
      initContainers:
        - name: remote-infra
          image: envoyproxy/gateway-remote-infra:latest
          imagePullPolicy: IfNotPresent
          restartPolicy: Always
          args:
            - --socket-path=/var/run/remote-infra/server.sock
          env:
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          # Gates startup of envoy-gateway: it will not start until the UDS
          # socket exists.
          startupProbe:
            exec:
              command: ["test", "-S", "/var/run/remote-infra/server.sock"]
            periodSeconds: 1
            failureThreshold: 30
          volumeMounts:
            - name: remote-infra-socket
              mountPath: /var/run/remote-infra
EOF
```

### Switch Envoy Gateway to the Remote provider

Patch the `envoy-gateway-config` ConfigMap to dial the sidecar over the shared
socket:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-gateway-config
  namespace: envoy-gateway-system
data:
  envoy-gateway.yaml: |
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyGateway
    gateway:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
    provider:
      type: Custom
      custom:
        resource:
          type: Kubernetes
        infrastructure:
          type: Remote
          remote:
            service:
              unix:
                path: /var/run/remote-infra/server.sock
EOF
```

Restart Envoy Gateway so the new config is picked up:

```shell
kubectl rollout restart -n envoy-gateway-system deployment/envoy-gateway
kubectl rollout status  -n envoy-gateway-system deployment/envoy-gateway --timeout=2m
```

### Verify

Apply a Gateway and an HTTPRoute, then watch the example provider's logs to
confirm it received the IR:

```shell
kubectl logs -n envoy-gateway-system deployment/envoy-gateway -c remote-infra
```

You should see `Creating proxy infra [...]` lines as Gateways are reconciled.
The example provider creates a `Deployment` and `Service` for each Gateway in
the same namespace as Envoy Gateway; check them with:

```shell
kubectl get deploy,svc -n envoy-gateway-system -l app.kubernetes.io/managed-by=envoy-gateway
```

## Writing your own provider

The example in `examples/remote-infra/` is a useful starting point but only
renders a `Deployment` and a `Service`. A production-grade provider will
typically need to:

- **Be idempotent.** Envoy Gateway calls `CreateOrUpdate` whenever the IR
  changes, which in practice can be many times per second during churn. The
  example uses a server-side apply pattern; you should do the same or its
  equivalent for your data plane.
- **Plan for IR evolution.** The IR is JSON for forward-compatibility, but new
  fields will be added as Envoy Gateway evolves. Design your provider to ignore
  unknown fields rather than fail on them.
- **Communicate data plane status**. Envoy Gateway needs metadata about your
  data plane, such as the DNS address to find the fleet at. Make sure your
  provider communicates that information back via a Service object. The example
  infra server provides a basic example.

The IR types are defined in `internal/ir/`. The `ir.Infra` value Envoy Gateway
sends is documented by the Go types `ir.ProxyInfra` and `ir.RateLimitInfra`.
