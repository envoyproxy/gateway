local g = import './g.libsonnet';
local prometheusQuery = g.query.prometheus;

local variables = import './variables.libsonnet';

{
  cpuUsageForEnvoyGateway:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum by (namespace) (
            rate(
                container_cpu_usage_seconds_total{
                    container="envoy-gateway"
                }
            [$__rate_interval])
        )
      |||
    )
    + prometheusQuery.withIntervalFactor(2)
    + prometheusQuery.withLegendFormat(|||
      {{namespace}}
    |||),

  cpuUsageForEnvoyProxy:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum by (pod) (
            rate(
                container_cpu_usage_seconds_total{
                    container="envoy"
                }
            [$__rate_interval])
        )
      |||
    )
    + prometheusQuery.withIntervalFactor(2)
    + prometheusQuery.withLegendFormat(|||
      {{pod}}
    |||),

  memUsageForEnvoyGateway:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum by (namespace) (
          container_memory_working_set_bytes{container="envoy-gateway"}
        )
      |||
    )
    + prometheusQuery.withIntervalFactor(2)
    + prometheusQuery.withLegendFormat(|||
      {{namespace}}
    |||),

  memUsageForEnvoyProxy:
    prometheusQuery.new(
      '$' + variables.datasource.name,
      |||
        sum by (pod) (
          container_memory_working_set_bytes{container="envoy"}
        )
      |||
    )
    + prometheusQuery.withIntervalFactor(2)
    + prometheusQuery.withLegendFormat(|||
      {{pod}}
    |||),

}

// vim: foldmethod=indent shiftwidth=2 foldlevel=1
