local g = import 'g.libsonnet';

{
  timeSeries: {
    local timeSeries = g.panel.timeSeries,
    local fieldOverride = g.panel.timeSeries.fieldOverride,
    local custom = timeSeries.fieldConfig.defaults.custom,
    local options = timeSeries.options,

    base(title, targets):
      timeSeries.new(title)
      + timeSeries.queryOptions.withTargets(targets)
      + timeSeries.queryOptions.withInterval('1m')
      + options.legend.withDisplayMode('table')
      + options.legend.withCalcs([
        'lastNotNull',
        'max',
      ])
      + custom.withFillOpacity(10)
      + custom.withShowPoints('never'),

    short(title, targets):
      self.base(title, targets)
      + timeSeries.standardOptions.withUnit('short')
      + timeSeries.standardOptions.withDecimals(0),

    seconds(title, targets):
      self.base(title, targets)
      + timeSeries.standardOptions.withUnit('s')
      + custom.scaleDistribution.withType('log')
      + custom.scaleDistribution.withLog(10),

    cpuUsage: self.seconds,

    bytes(title, targets):
      self.base(title, targets)
      + timeSeries.standardOptions.withUnit('bytes')
      + custom.scaleDistribution.withType('log')
      + custom.scaleDistribution.withLog(2),

    memoryUsage: self.bytes,

    durationQuantile(title, targets):
      self.base(title, targets)
      + timeSeries.standardOptions.withUnit('s')
      + custom.withDrawStyle('bars')
      + timeSeries.standardOptions.withOverrides([
        fieldOverride.byRegexp.new('/mean/i')
        + fieldOverride.byRegexp.withProperty(
          'custom.fillOpacity',
          0
        )
        + fieldOverride.byRegexp.withProperty(
          'custom.lineStyle',
          {
            dash: [8, 10],
            fill: 'dash',
          }
        ),
      ]),
  },

  heatmap: {
    local heatmap = g.panel.heatmap,
    local options = heatmap.options,

    base(title, targets):
      heatmap.new(title)
      + heatmap.queryOptions.withTargets(targets)
      + heatmap.queryOptions.withInterval('1m')
      + options.withCalculate()
      + options.calculation.xBuckets.withMode('size')
      + options.calculation.xBuckets.withValue('1min')
      + options.withCellGap(2)
      + options.color.withMode('scheme')
      + options.color.withScheme('Spectral')
      + options.color.withSteps(128)
      + options.yAxis.withDecimals(0)
      + options.yAxis.withUnit('s'),
  },
}

// vim: foldmethod=marker foldmarker=local,;
