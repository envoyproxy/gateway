import React from 'react';
import { AreaChart as RechartsAreaChart, Area, ComposedChart, Line } from 'recharts';
import { ChartConfig } from '@/components/ui/chart';
import { BaseChart } from './BaseChart';
import { CHART_STYLES, CHART_COLORS } from '@/config/chartConfig';
import { APP_CONSTANTS } from '@/config/constants';

interface AreaChartProps {
  data: any[];
  config: ChartConfig;
  areas: {
    dataKey: string;
    stackId?: string;
    fill?: string;
    stroke?: string;
    fillOpacity?: number;
    strokeWidth?: number;
  }[];
  lines?: {
    dataKey: string;
    stroke?: string;
    strokeWidth?: number;
    strokeDasharray?: string;
    dot?: boolean | object;
  }[];
  height?: number;
  xAxisConfig?: {
    dataKey: string;
    type?: 'number' | 'category';
    domain?: [number | string, number | string];
    ticks?: number[];
    formatter?: (value: any) => string;
  };
  yAxisConfig?: {
    domain?: [number | string, number | string];
    formatter?: (value: any) => string;
  };
  tooltipConfig?: {
    labelFormatter?: (value: any) => string;
    formatter?: (value: any, name: string) => [string, string];
  };
  showGrid?: boolean;
  className?: string;
}

export const AreaChart: React.FC<AreaChartProps> = ({
  data,
  config,
  areas,
  lines,
  height,
  xAxisConfig,
  yAxisConfig,
  tooltipConfig,
  showGrid = true,
  className,
}) => {
  const ChartComponent = lines && lines.length > 0 ? ComposedChart : RechartsAreaChart;

  return (
    <BaseChart
      data={data}
      config={config}
      height={height}
      xAxisConfig={xAxisConfig}
      yAxisConfig={yAxisConfig}
      tooltipConfig={tooltipConfig}
      showGrid={showGrid}
      className={className}
    >
      <ChartComponent>
        {/* Areas */}
        {areas.map((area, index) => (
          <Area
            key={area.dataKey}
            dataKey={area.dataKey}
            stackId={area.stackId}
            type="monotone"
            fill={area.fill || Object.values(CHART_COLORS)[index % Object.values(CHART_COLORS).length]}
            fillOpacity={area.fillOpacity || CHART_STYLES.AREA.FILL_OPACITY}
            stroke={area.stroke || area.fill || Object.values(CHART_COLORS)[index % Object.values(CHART_COLORS).length]}
            strokeWidth={area.strokeWidth || CHART_STYLES.AREA.STROKE_WIDTH}
          />
        ))}

        {/* Lines (for ComposedChart) */}
        {lines && lines.map((line, index) => (
          <Line
            key={line.dataKey}
            dataKey={line.dataKey}
            type="monotone"
            stroke={line.stroke || Object.values(CHART_COLORS)[index % Object.values(CHART_COLORS).length]}
            strokeWidth={line.strokeWidth || CHART_STYLES.LINE.STROKE_WIDTH}
            strokeDasharray={line.strokeDasharray}
            dot={line.dot !== undefined ? line.dot : {
              fill: line.stroke || Object.values(CHART_COLORS)[index % Object.values(CHART_COLORS).length],
              strokeWidth: CHART_STYLES.LINE.DOT_STROKE_WIDTH,
              r: CHART_STYLES.LINE.DOT_RADIUS,
            }}
          />
        ))}
      </ChartComponent>
    </BaseChart>
  );
};

// Predefined area chart variants
export const ThroughputAreaChart: React.FC<{
  data: any[];
  config?: ChartConfig;
  height?: number;
}> = ({ data, config, height }) => (
  <AreaChart
    data={data}
    config={config || { throughput: { label: 'Throughput', color: CHART_COLORS.PRIMARY } }}
    areas={[{ dataKey: 'throughput' }]}
    height={height}
    xAxisConfig={{
      dataKey: 'routes',
      type: 'number',
      domain: APP_CONSTANTS.CHART.ROUTE_DOMAIN as [number, number],
      ticks: APP_CONSTANTS.CHART.ROUTE_TICKS,
    }}
    yAxisConfig={{
      formatter: (value) => `${value}`,
    }}
    tooltipConfig={{
      formatter: (value, name) => [`${value} RPS`, 'Throughput'],
      labelFormatter: (value) => `${value} routes`,
    }}
  />
);

export const MemoryAreaChart: React.FC<{
  data: any[];
  config?: ChartConfig;
  height?: number;
}> = ({ data, config, height }) => (
  <AreaChart
    data={data}
    config={config || {
      gateway: { label: 'Gateway', color: CHART_COLORS.ACCENT },
      proxy: { label: 'Proxy', color: CHART_COLORS.DARK },
    }}
    areas={[
      { dataKey: 'gateway', stackId: 'memory', fillOpacity: 0.6 },
      { dataKey: 'proxy', stackId: 'memory', fillOpacity: 0.6 },
    ]}
    height={height}
    xAxisConfig={{
      dataKey: 'routes',
      type: 'number',
      domain: APP_CONSTANTS.CHART.ROUTE_DOMAIN as [number, number],
      ticks: APP_CONSTANTS.CHART.ROUTE_TICKS,
    }}
    yAxisConfig={{
      formatter: (value) => `${value}MB`,
    }}
    tooltipConfig={{
      labelFormatter: (value) => `${value} routes`,
    }}
  />
);
