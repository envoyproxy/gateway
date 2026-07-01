# Remote Infrastructure Provider Example

This example is a reference implementation of the Envoy Gateway [Remote
Infrastructure Provider][docs] gRPC service. It receives the Envoy Gateway
Infrastructure IR over gRPC and reconciles a `Deployment` and a `Service` per
proxy fleet against the Kubernetes API.

It is the server used by the `e2e-remote-infra` test target and is intended as
a starting point for writing your own provider — not a production-ready
implementation.

For the full conceptual model and a quickstart that wires this image into an
Envoy Gateway install, see [the task documentation][docs].

## What it does

The provider implements the four RPCs defined in
[`proto/remoteinfra/service.proto`](../../proto/remoteinfra/service.proto):

| RPC                            | Behavior                                                         |
| ------------------------------ | ---------------------------------------------------------------- |
| `CreateOrUpdateProxyInfra`     | Reconciles a `Deployment` and `Service` for the IR's proxy fleet |
| `DeleteProxyInfra`             | Deletes the `Deployment` and `Service` for the IR's proxy fleet  |
| `CreateOrUpdateRateLimitInfra` | No-op. Rate limiting must be provisioned out of band, see below  |
| `DeleteRateLimitInfra`         | No-op                                                            |

The IR is delivered as JSON in the request's `ir_bytes` field. The example
ignores fields it doesn't understand, which is the same forward-compatibility
posture you'll want in your own provider.

## Layout

| Path                | Purpose                                                                  |
| ------------------- | ------------------------------------------------------------------------ |
| `main.go`           | Binds a Unix domain socket and starts the gRPC server                    |
| `synthesizer/`      | Translates the IR into desired `Deployment` / `Service` objects          |
| `synthesizer/*.tpl` | Embedded Envoy bootstrap template                                        |
| `pb/`               | Generated gRPC stubs (regenerated from the proto in this repo)           |
| `Dockerfile`        | Builds the server image. Alpine base — see note below                    |
| `Makefile`          | `docker-buildx` target produces `envoyproxy/gateway-remote-infra:latest` |

## Build

```shell
make docker-buildx
```

If you're running on `kind`, load the image into the cluster:

```shell
kind load docker-image --name envoy-gateway envoyproxy/gateway-remote-infra:latest
```

## Run

The expected deployment shape is as a native sidecar of the `envoy-gateway`
Pod, sharing an `emptyDir` volume that holds the UDS. The full patch is in
`test/e2e/remote_infra/sidecar-patch.yaml` and is documented in [the task
guide][docs]. By default the server listens on
`/var/run/remote-infra/server.sock` with mode `0660`; both are tunable via
flags:

```text
--socket-path=/var/run/remote-infra/server.sock
--socket-mode=0660
```

The `NAMESPACE` environment variable selects which Kubernetes namespace the
provider writes resources into. The sidecar patch wires this to the Pod's own
namespace via the downward API.

## Limitations

This is a teaching example, not a production provider. In particular:

- It only renders one Deployment and one Service per IR. Real providers will
  often want a HorizontalPodAutoscaler, PodDisruptionBudget, and so on.
- It does not provision rate limit infrastructure.
- It does not surface reconcile errors back to Envoy Gateway through the gRPC
  response in a structured way; errors are returned as plain strings.
