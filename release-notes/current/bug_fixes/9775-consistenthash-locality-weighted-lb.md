Fixed backendRef weights being silently ignored when a route's backendRefs collapse into a
single cluster (no per-backendRef filters) with a ConsistentHash load balancer configured.
Envoy Gateway now always enables locality weighted load balancing on the generated Maglev
policy, so the configured weights are honored instead of being treated as equal.
