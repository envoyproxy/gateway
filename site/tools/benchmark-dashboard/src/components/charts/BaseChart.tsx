
import React from 'react';
import { ChartContainer, ChartTooltip, ChartTooltipContent, ChartConfig } from '@/components/ui/chart';
import { CartesianGrid, XAxis, YAxis, ResponsiveContainer } from 'recharts';
import { APP_CONSTANTS } from '@/config/constants';
import { CHART_STYLES } from '@/config/chartConfig';
import { ErrorBoundary } from '@/components/common/ErrorBoundary';
import { ChartPlaceholder } from '@/components/common/DataPlaceholder';
import ChartWatermark from '@/components/common/ChartWatermark';

interface BaseChartProps {
  data: any[];
  config: ChartConfig;
  height?: number;
  children: React.ReactNode;
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
  showWatermark?: boolean;
}

export const BaseChart: React.FC<BaseChartProps> = ({
  data,
  config,
  height = APP_CONSTANTS.CHART.DEFAULT_HEIGHT,
  children,
  xAxisConfig,
  yAxisConfig,
  tooltipConfig,
  showGrid = true,
  className,
  showWatermark = false,
}) => {
  // Handle empty data
  if (!data || data.length === 0) {
    return <ChartPlaceholder />;
  }

  return (
    <ErrorBoundary>
      <div className={`relative ${className}`}>
        {showWatermark && <ChartWatermark />}

        <ChartContainer config={config}>
          <ResponsiveContainer width="100%" height={height}>
            {React.cloneElement(children as React.ReactElement, {
              data,
              children: [
                // Grid
                showGrid && (
                  <CartesianGrid
                    key="grid"
                    strokeDasharray={CHART_STYLES.GRID.STROKE_DASH_ARRAY}
                  />
                ),

                // X Axis
                <XAxis
                  key="xaxis"
                  dataKey={xAxisConfig?.dataKey || 'x'}
                  type={xAxisConfig?.type || 'category'}
                  domain={xAxisConfig?.domain}
                  ticks={xAxisConfig?.ticks}
                  tickLine={false}
                  axisLine={false}
                  tickMargin={APP_CONSTANTS.CHART.DEFAULT_TICK_MARGIN}
                  tickFormatter={xAxisConfig?.formatter}
                />,

                // Y Axis
                <YAxis
                  key="yaxis"
                  domain={yAxisConfig?.domain || [0, 'dataMax']}
                  tickLine={false}
                  axisLine={false}
                  tickMargin={APP_CONSTANTS.CHART.DEFAULT_TICK_MARGIN}
                  tickFormatter={yAxisConfig?.formatter}
                />,

                // Tooltip
                <ChartTooltip
                  key="tooltip"
                  content={
                    <ChartTooltipContent
                      formatter={tooltipConfig?.formatter}
                      labelFormatter={tooltipConfig?.labelFormatter}
                    />
                  }
                />,

                // Chart content
                ...(React.Children.toArray(
                  (children as React.ReactElement).props.children
                ) || []),
              ].filter(Boolean),
            })}
          </ResponsiveContainer>
        </ChartContainer>
      </div>
    </ErrorBoundary>
  );
};
