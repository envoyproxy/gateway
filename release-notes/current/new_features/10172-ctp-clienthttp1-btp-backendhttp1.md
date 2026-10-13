Added `BackendTrafficPolicy.spec.http1` (`BackendHTTP1Settings`) for configuring HTTP/1 settings
on upstream backend connections, including `enableTrailers` and `preserveHeaderCase`. When both
`BackendTrafficPolicy.spec.http1` and `ClientTrafficPolicy.spec.http1` are set, `BackendTrafficPolicy`
takes precedence for backend traffic. Setting `http1: {}` (empty) forces explicit HTTP/1 on the
upstream cluster.

When targeting a GRPCRoute, TCPRoute, TLSRoute, or UDPRoute directly via `targetRef`, `targetRefs`,
or `targetSelectors`, the policy remains accepted but receives a `Warning` condition since those route
kinds do not use HTTP/1 and the field has no effect. When attached to a broader target (Gateway, ListenerSet, or a specific
listener via `sectionName`), `http1` is silently ignored for any GRPCRoute, TCPRoute, TLSRoute, or
UDPRoute destinations beneath that target and still applies to HTTPRoute destinations. For gRPC-only
external service backends (ExtProc, gRPC ExtAuth, ALS, OpenTelemetry access log/metrics/tracing),
`http1` on their own `BackendSettings` is rejected at admission by existing validations on
those types.
