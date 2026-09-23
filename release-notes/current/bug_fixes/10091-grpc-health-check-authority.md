Fixed active gRPC health checks sending an invalid `:authority` header. Envoy Gateway only set the
service name on the gRPC health checker, so Envoy fell back to the name of the cluster, which is
generated per route rule and contains slashes (`grpcroute/my-ns/my-route/rule/0`). Slashes are not
valid in an authority, and gRPC servers built on `golang.org/x/net` v0.59.0 or later reset the stream
with `PROTOCOL_ERROR`, leaving every endpoint marked `failed_active_hc` and clients receiving
`no healthy upstream` from a healthy backend. The authority now defaults to the effective route
hostname, mirroring the HTTP health checker, and `BackendTrafficPolicy` gained
`healthCheck.active.grpc.hostname` to set it explicitly.
