import React from 'react';
import { LineChart as RechartsLineChart, Line } from 'recharts';
import { ChartConfig } from '@/components/ui/chart';
import { BaseChart } from './BaseChart';
import { CHART_STYLES, CHART_COLORS } from '@/config/chartConfig';
import { APP_CONSTANTS, STROKE_PATTERNS } from '@/config/constants';

interface LineChartProps {
  data: any[];
  config: ChartConfig;
  lines: {
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

export const LineChart: React.FC<LineChartProps> = ({
  data,
  config,
  lines,
  height,
  xAxisConfig,
  yAxisConfig,
  tooltipConfig,
  showGrid = true,
  className,
}) => {
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
      <RechartsLineChart>
        {lines.map((line, index) => (
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
      </RechartsLineChart>
    </BaseChart>
  );
};

// Predefined line chart variants
export const EfficiencyLineChart: React.FC<{
  data: any[];
  config?: ChartConfig;
  height?: number;
}> = ({ data, config, height }) => (
  <LineChart
    data={data}
    config={config || {
      totalPerRoute: { label: 'Total per Route', color: CHART_COLORS.PRIMARY },
      gatewayPerRoute: { label: 'Gateway per Route', color: CHART_COLORS.ACCENT },
      proxyPerRoute: { label: 'Proxy per Route', color: CHART_COLORS.DARK },
    }}
    lines={[
      { dataKey: 'totalPerRoute', strokeWidth: 3, dot: { r: 4 } },
      { dataKey: 'gatewayPerRoute', strokeWidth: 2, strokeDasharray: STROKE_PATTERNS.DASHED, dot: { r: 3 } },
      { dataKey: 'proxyPerRoute', strokeWidth: 2, strokeDasharray: STROKE_PATTERNS.DOTTED, dot: { r: 3 } },
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
      formatter: (value, name) => [
        `${value}MB per route`,
        name === 'gatewayPerRoute' ? 'Gateway' :
        name === 'proxyPerRoute' ? 'Proxy' : 'Total'
      ],
      labelFormatter: (value) => `${value} routes`,
    }}
  />
);

export const LatencyLineChart: React.FC<{
  data: any[];
  config?: ChartConfig;
  height?: number;
}> = ({ data, config, height }) => (
  <LineChart
    data={data}
    config={config || {
      latency: { label: 'Latency', color: CHART_COLORS.PRIMARY },
    }}
    lines={[{ dataKey: 'latency' }]}
    height={height}
    xAxisConfig={{
      dataKey: 'routes',
      type: 'number',
      domain: APP_CONSTANTS.CHART.ROUTE_DOMAIN as [number, number],
      ticks: APP_CONSTANTS.CHART.ROUTE_TICKS,
    }}
    yAxisConfig={{
      formatter: (value) => `${value}ms`,
    }}
    tooltipConfig={{
      formatter: (value, name) => [`${value}ms`, 'Latency'],
      labelFormatter: (value) => `${value} routes`,
    }}
  />
);
