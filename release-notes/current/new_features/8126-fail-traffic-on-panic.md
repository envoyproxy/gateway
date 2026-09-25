Added a `failTrafficOnPanic` field to the zone-aware `preferLocal` load balancer
settings, exposing Envoy's `fail_traffic_on_panic`. When a cluster enters panic
mode because the share of healthy endpoints dropped below `panicThreshold`,
Envoy load balances across every endpoint regardless of health; setting this to
`true` makes it reject the traffic instead, applying backpressure to clients and
giving the backend room to recover. It sits under `preferLocal` because Envoy
only defines the setting inside `zone_aware_lb_config`. When unset it defaults
to `false`, preserving the existing behavior.
