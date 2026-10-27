Added `adaptiveConcurrency` to `BackendTrafficPolicy` so Envoy can dynamically
adjust request concurrency from sampled latencies and shed load when backends
slow down. The initial API exposes `sampleAggregatePercentile`,
`minConcurrencyLimit`, `maxConcurrencyLimit`, `concurrencyUpdateInterval`,
and `concurrencyLimitExceededStatus` (Envoy `StatusCode` values only); empty configuration uses Envoy defaults.
The policy attaches to `Gateway`, `ListenerSet`, `HTTPRoute`, or `GRPCRoute`.
