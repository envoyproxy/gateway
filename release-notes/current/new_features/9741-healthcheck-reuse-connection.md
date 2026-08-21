Added a `reuseConnection` field to the active health check configuration in
BackendTrafficPolicy, exposing Envoy's `HealthCheck.reuse_connection`. Setting it
to `false` makes Envoy open a fresh connection for each health check probe, which
is required to interoperate with backends such as haproxy-style agent checks that
do not expect a reused connection. When unset, Envoy's default of `true` is
preserved.
