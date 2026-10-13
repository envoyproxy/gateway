Added `BackendTrafficPolicy.spec.http1` (`BackendHTTP1Settings`) for configuring HTTP/1 settings
on upstream backend connections, including `enableTrailers` and `preserveHeaderCase`. When both
`BackendTrafficPolicy.spec.http1` and `ClientTrafficPolicy.spec.http1` are set, `BackendTrafficPolicy`
takes precedence for backend traffic. Setting `http1: {}` (empty) forces explicit HTTP/1 on the
upstream cluster.

This field is not supported for gRPC destinations: attaching a `BackendTrafficPolicy` with `http1`
directly to a `GRPCRoute` is rejected at admission. When the policy targets a `Gateway`, a specific listener via `sectionName`, or a
`ListenerSet`, `http1` is silently ignored for any `GRPCRoute` destinations beneath that target.
gRPC-only external services (ExtProc, gRPC ExtAuth, ALS, OpenTelemetry access log/metrics, and
tracing providers) reject `http1` on their own `backendSettings` via separate validation rules.
