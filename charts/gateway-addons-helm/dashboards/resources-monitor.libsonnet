local g = import 'lib/g.libsonnet';

local row = g.panel.row;

local panels = import 'lib/panels.libsonnet';
local variables = import 'lib/variables.libsonnet';
local queries = import 'lib/queries.libsonnet';

g.dashboard.new('Resources Monitor')
+ g.dashboard.withDescription(|||
  Memory and CPU Usage Monitor for Envoy Gateway and Envoy Proxy.
|||)
+ g.dashboard.graphTooltip.withSharedCrosshair()
+ g.dashboard.withVariables([
  variables.datasource,
])
+ g.dashboard.withPanels(
  g.util.grid.makeGrid([
    row.new('Envoy Gateway')
    + row.withPanels([
      panels.timeSeries.cpuUsage('CPU Usage', queries.cpuUsageForEnvoyGateway),
      panels.timeSeries.memoryUsage('Memory Usage', queries.memUsageForEnvoyGateway),
    ]),
    row.new('Envoy Proxy')
    + row.withPanels([
      panels.timeSeries.cpuUsage('CPU Usage', queries.cpuUsageForEnvoyProxy),
      panels.timeSeries.memoryUsage('Memory Usage', queries.memUsageForEnvoyProxy),
    ]),
  ], panelWidth=8)
)
+ g.dashboard.withUid(std.md5('resources-monitor.json'))
