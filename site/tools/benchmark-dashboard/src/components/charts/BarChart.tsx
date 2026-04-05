import React from 'react';
import { BarChart as RechartsBarChart, Bar } from 'recharts';
import { ChartConfig } from '@/components/ui/chart';
import { BaseChart } from './BaseChart';
import { CHART_STYLES, CHART_COLORS } from '@/config/chartConfig';

interface BarChartProps {
  data: any[];
  config: ChartConfig;
  bars: {
    dataKey: string;
    fill?: string;
    radius?: number;
    stackId?: string;
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

export const BarChart: React.FC<BarChartProps> = ({
  data,
  config,
  bars,
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
      <RechartsBarChart>
        {bars.map((bar, index) => (
          <Bar
            key={bar.dataKey}
            dataKey={bar.dataKey}
            stackId={bar.stackId}
            fill={bar.fill || Object.values(CHART_COLORS)[index % Object.values(CHART_COLORS).length]}
            radius={bar.radius || CHART_STYLES.BAR.RADIUS}
          />
        ))}
      </RechartsBarChart>
    </BaseChart>
  );
};

// Predefined bar chart variants
export const LatencyDistributionBarChart: React.FC<{
  data: any[];
  config?: ChartConfig;
  height?: number;
}> = ({ data, config, height }) => (
  <BarChart
    data={data}
    config={config || { value: { label: 'Latency', color: CHART_COLORS.PRIMARY } }}
    bars={[{ dataKey: 'value' }]}
    height={height}
    xAxisConfig={{
      dataKey: 'percentile',
      type: 'category',
    }}
    yAxisConfig={{
      formatter: (value) => `${value}ms`,
    }}
    tooltipConfig={{
      formatter: (value, name) => [`${value}ms`, 'Latency'],
      labelFormatter: (value) => value,
    }}
  />
);

export const MetricComparisonBarChart: React.FC<{
  data: any[];
  config?: ChartConfig;
  height?: number;
  dataKeys: string[];
}> = ({ data, config, height, dataKeys }) => (
  <BarChart
    data={data}
    config={config || dataKeys.reduce((acc, key, index) => ({
      ...acc,
      [key]: { label: key, color: Object.values(CHART_COLORS)[index % Object.values(CHART_COLORS).length] }
    }), {})}
    bars={dataKeys.map(key => ({ dataKey: key }))}
    height={height}
    xAxisConfig={{
      dataKey: 'name',
      type: 'category',
    }}
    tooltipConfig={{
      labelFormatter: (value) => value,
    }}
  />
);
