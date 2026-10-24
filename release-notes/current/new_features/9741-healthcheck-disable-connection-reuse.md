Added a `disableConnectionReuse` field to the active health check configuration in
BackendTrafficPolicy, exposing Envoy's `HealthCheck.reuse_connection`. Setting it
to `true` makes Envoy open a fresh connection for each health check probe, which
is required to interoperate with backends such as haproxy-style agent checks that
do not expect a reused connection. When unset it defaults to `false`, preserving
the existing behavior of reusing the connection across probes.
