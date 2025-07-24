
import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart';
import { ChartPlaceholder } from '@/components/common/DataPlaceholder';
import ChartWatermark from '@/components/common/ChartWatermark';
import { AreaChart, Area, LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid } from 'recharts';
import { Clock, Gauge, Target, TrendingUp, AlertTriangle, CheckCircle } from 'lucide-react';

interface LatencyTabProps {
  latencyPercentileComparison: any[];
  benchmarkResults: any[];
}

const LatencyTab = ({ latencyPercentileComparison, benchmarkResults }: LatencyTabProps) => {
  // Clean RTT data for scaling up phase
  const latencyData = latencyPercentileComparison
    .filter(item => item.phase === 'scaling-up')
    .map(item => ({
      routes: item.routes,
      p50: Number(item.p50.toFixed(1)),
      p95: Number(item.p95.toFixed(1)),
      p99: Number(item.p99.toFixed(1))
    }));

  // Generate RTT distribution dynamically from benchmark results at max scale (worst case)
  const generateLatencyDistribution = () => {
    // If we have latencyPercentileComparison data, use it
    if (latencyPercentileComparison && latencyPercentileComparison.length > 0) {
      const maxScaleData = latencyPercentileComparison
        .filter(item => item.phase === 'scaling-up')
        .sort((a, b) => b.routes - a.routes)[0]; // Get highest route count

      if (maxScaleData) {
        return [
          { percentile: 'P50', value: Number(maxScaleData.p50.toFixed(1)), category: 'Median', status: 'excellent' },
          { percentile: 'P75', value: Number(maxScaleData.p75.toFixed(1)), category: '75th', status: 'excellent' },
          { percentile: 'P90', value: Number(maxScaleData.p90.toFixed(1)), category: '90th', status: 'good' },
          { percentile: 'P95', value: Number(maxScaleData.p95.toFixed(1)), category: '95th', status: 'watch' },
          { percentile: 'P99', value: Number(maxScaleData.p99.toFixed(1)), category: '99th', status: 'alert' }
        ];
      }
    }

    // If we have benchmarkResults, extract from there
    if (benchmarkResults && benchmarkResults.length > 0) {
      const maxScaleData = benchmarkResults
        .filter(item => item.phase === 'scaling-up')
        .sort((a, b) => b.routes - a.routes)[0]; // Get highest route count

      if (maxScaleData && maxScaleData.latency && maxScaleData.latency.percentiles) {
        const percentiles = maxScaleData.latency.percentiles;
        return [
          { percentile: 'P50', value: Number((percentiles.p50 / 1000).toFixed(1)), category: 'Median', status: 'excellent' },
          { percentile: 'P75', value: Number((percentiles.p75 / 1000).toFixed(1)), category: '75th', status: 'excellent' },
          { percentile: 'P90', value: Number((percentiles.p90 / 1000).toFixed(1)), category: '90th', status: 'good' },
          { percentile: 'P95', value: Number((percentiles.p95 / 1000).toFixed(1)), category: '95th', status: 'watch' },
          { percentile: 'P99', value: Number((percentiles.p99 / 1000).toFixed(1)), category: '99th', status: 'alert' }
        ];
      }
    }

    // Return null when no data is available
    return null;
  };

  const latencyDistribution = generateLatencyDistribution();

  // RTT consistency across routes
  const latencyConsistency = benchmarkResults
    .filter(item => item.phase === 'scaling-up')
    .map(item => ({
      routes: item.routes,
      mean: Number((item.latency.mean / 1000).toFixed(1)),
      p95: Number((item.latency.percentiles.p95 / 1000).toFixed(1)),
      ratio: Number((item.latency.percentiles.p95 / item.latency.mean).toFixed(1))
    }));

  // Generate performance comparison data dynamically
  const generatePerformanceComparison = () => {
    if (latencyPercentileComparison && latencyPercentileComparison.length > 0) {
      return latencyPercentileComparison
        .filter(item => item.phase === 'scaling-up')
        .sort((a, b) => a.routes - b.routes) // Sort by routes ascending
        .map(item => ({
          scale: `${item.routes} Routes`,
          p50: Number(item.p50.toFixed(1)),
          p95: Number(item.p95.toFixed(1)),
          p99: Number(item.p99.toFixed(1))
        }));
    }

    // Return null when no data is available
    return null;
  };

  const performanceComparison = generatePerformanceComparison();

  const chartConfig = {
    p50: {
      label: "P50 (Median)",
      color: "#8b5cf6", // Purple-500
    },
    p95: {
      label: "P95",
      color: "#6366f1", // Indigo-500
    },
    p99: {
      label: "P99",
      color: "#4f46e5", // Indigo-600
    },
    mean: {
      label: "Mean",
      color: "#a855f7", // Purple-400
    },
    ratio: {
      label: "P95/Mean Ratio",
      color: "#7c3aed", // Violet-600
    },
  };

  // Calculate key metrics from actual data
  const calculateKeyMetrics = () => {
    if (latencyData.length === 0) {
      return {
        medianLatency: 0,
        p95Latency: 0,
        p99Latency: 0,
        consistencyRatio: 0
      };
    }

    // Use the highest scale data point (worst case) for display
    const maxScaleData = latencyData[latencyData.length - 1];

    // Calculate average P95/Mean ratio across all scales for consistency metric
    const avgRatio = latencyConsistency.length > 0
      ? latencyConsistency.reduce((sum, item) => sum + item.ratio, 0) / latencyConsistency.length
      : 0;

    return {
      medianLatency: maxScaleData?.p50 || 0,
      p95Latency: maxScaleData?.p95 || 0,
      p99Latency: maxScaleData?.p99 || 0,
      consistencyRatio: avgRatio
    };
  };

  const keyMetrics = calculateKeyMetrics();

  return (
    <div className="space-y-6">
      {/* Key RTT Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Median RTT</CardTitle>
            <Target className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{keyMetrics.medianLatency.toFixed(1)}ms</div>
            <p className="text-xs text-muted-foreground">
              Consistent across all scales
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">P95 RTT</CardTitle>
            <Gauge className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{keyMetrics.p95Latency.toFixed(1)}ms</div>
            <p className="text-xs text-muted-foreground">
              95% of requests under {keyMetrics.p95Latency.toFixed(0)}ms
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Tail RTT</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{keyMetrics.p99Latency.toFixed(1)}ms</div>
            <p className="text-xs text-muted-foreground">
              P99 RTT (1% of requests)
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Consistency</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{keyMetrics.consistencyRatio.toFixed(1)}:1</div>
            <p className="text-xs text-muted-foreground">
              P95/Mean ratio
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Main RTT Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Key Percentiles Over Scale */}
        <Card>
          <CardHeader>
            <CardTitle>Request RTT Scaling Behavior</CardTitle>
            <CardDescription>
              How key percentiles perform as route count increases
            </CardDescription>
          </CardHeader>
          <CardContent>
            {latencyData && latencyData.length > 0 ? (
              <div className="relative">
                <ChartWatermark />
                <ChartContainer config={chartConfig}>
                  <AreaChart data={latencyData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="routes"
                      type="number"
                      scale="linear"
                      domain={[0, 1100]}
                      ticks={[10, 50, 100, 300, 500, 1000]}
                      tickLine={false}
                      axisLine={false}
                      tickMargin={8}
                    />
                    <YAxis
                      tickLine={false}
                      axisLine={false}
                      tickMargin={8}
                      tickFormatter={(value) => `${value}ms`}
                    />
                    <ChartTooltip
                      content={<ChartTooltipContent
                        labelFormatter={(value) => `${value} routes`}
                      />}
                    />
                    <Area
                      dataKey="p99"
                      type="monotone"
                      fill="#4f46e5"
                      fillOpacity={0.2}
                      stroke="#4f46e5"
                      strokeWidth={2}
                    />
                    <Area
                      dataKey="p95"
                      type="monotone"
                      fill="#6366f1"
                      fillOpacity={0.3}
                      stroke="#6366f1"
                      strokeWidth={2}
                    />
                    <Area
                      dataKey="p50"
                      type="monotone"
                      fill="#8b5cf6"
                      fillOpacity={0.4}
                      stroke="#8b5cf6"
                      strokeWidth={2}
                    />
                  </AreaChart>
                </ChartContainer>
              </div>
            ) : (
              <ChartPlaceholder />
            )}
          </CardContent>
        </Card>

        {/* RTT Distribution at Scale */}
        <Card>
          <CardHeader>
            <CardTitle>Request RTT Distribution</CardTitle>
            <CardDescription>
              Percentile breakdown at 1000 routes (worst case scenario)
            </CardDescription>
          </CardHeader>
          <CardContent>
            {latencyDistribution ? (
              <div className="relative">
                <ChartWatermark />
                <ChartContainer config={chartConfig}>
                  <BarChart data={latencyDistribution}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="percentile"
                      tickLine={false}
                      axisLine={false}
                      tickMargin={8}
                    />
                    <YAxis
                      tickLine={false}
                      axisLine={false}
                      tickMargin={8}
                      tickFormatter={(value) => `${value}ms`}
                    />
                    <ChartTooltip
                      content={<ChartTooltipContent
                        formatter={(value, name) => [
                          `${value}ms`,
                          "RTT"
                        ]}
                        labelFormatter={(value) => `${value} Percentile`}
                      />}
                    />
                    <Bar
                      dataKey="value"
                      fill="#6366f1"
                      radius={[4, 4, 0, 0]}
                    />
                  </BarChart>
                </ChartContainer>
              </div>
            ) : (
              <ChartPlaceholder />
            )}
          </CardContent>
        </Card>
      </div>

      {/* Secondary Analysis */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* RTT Consistency */}
        <Card>
          <CardHeader>
            <CardTitle>Request RTT Consistency</CardTitle>
            <CardDescription>
              P95/Mean ratio shows how predictable RTT is
            </CardDescription>
          </CardHeader>
          <CardContent>
            {latencyConsistency && latencyConsistency.length > 0 ? (
              <div className="relative">
                <ChartWatermark />
                <ChartContainer config={chartConfig}>
                  <LineChart data={latencyConsistency}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="routes"
                      type="number"
                      scale="linear"
                      domain={[0, 1100]}
                      ticks={[10, 50, 100, 300, 500, 1000]}
                      tickLine={false}
                      axisLine={false}
                      tickMargin={8}
                    />
                    <YAxis
                      yAxisId="ratio"
                      orientation="left"
                      tickLine={false}
                      axisLine={false}
                      tickMargin={8}
                      tickFormatter={(value) => `${value}x`}
                    />
                    <YAxis
                      yAxisId="latency"
                      orientation="right"
                      tickLine={false}
                      axisLine={false}
                      tickMargin={8}
                      tickFormatter={(value) => `${value}ms`}
                    />
                    <ChartTooltip
                      content={<ChartTooltipContent
                        labelFormatter={(value) => `${value} routes`}
                      />}
                    />
                    <Line
                      yAxisId="latency"
                      dataKey="mean"
                      type="monotone"
                      stroke="#a855f7"
                      strokeWidth={2}
                      dot={{ fill: "#a855f7", strokeWidth: 2, r: 3 }}
                    />
                    <Line
                      yAxisId="ratio"
                      dataKey="ratio"
                      type="monotone"
                      stroke="#7c3aed"
                      strokeWidth={3}
                      strokeDasharray="5 5"
                      dot={{ fill: "#7c3aed", strokeWidth: 2, r: 4 }}
                    />
                  </LineChart>
                </ChartContainer>
              </div>
            ) : (
              <ChartPlaceholder />
            )}
          </CardContent>
        </Card>

        {/* Performance Summary Table */}
        <Card>
          <CardHeader>
            <CardTitle>Performance Summary</CardTitle>
            <CardDescription>
              RTT performance at different scales
            </CardDescription>
          </CardHeader>
          <CardContent>
            {performanceComparison && performanceComparison.length > 0 ? (
              <div className="space-y-4">
                {performanceComparison.map((item, index) => (
                  <div key={index} className="flex items-center justify-between p-3 bg-muted/30 rounded-lg">
                    <div className="space-y-1">
                      <p className="font-medium">{item.scale}</p>
                      <div className="flex space-x-4 text-sm text-muted-foreground">
                        <span>P50: {item.p50}ms</span>
                        <span>P95: {item.p95}ms</span>
                        <span>P99: {item.p99}ms</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <ChartPlaceholder />
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default LatencyTab;
