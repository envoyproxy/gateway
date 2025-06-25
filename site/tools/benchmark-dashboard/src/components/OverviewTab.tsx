
import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart';
import { ChartPlaceholder } from '@/components/common/DataPlaceholder';
import ChartWatermark from '@/components/common/ChartWatermark';
import {
  AreaChart,
  Area,
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer
} from 'recharts';
import { TrendingUp, Clock, CheckCircle, Lightbulb, MemoryStick, Cpu, Zap, Shield } from 'lucide-react';

interface OverviewTabProps {
  performanceMatrix: any[];
  benchmarkResults: any[];
  testConfiguration: any;
  performanceSummary: any;
  latencyPercentileComparison?: any[]; // Add optional latency data
}

const OverviewTab = ({
  performanceMatrix,
  benchmarkResults,
  testConfiguration,
  performanceSummary,
  latencyPercentileComparison
}: OverviewTabProps) => {
  // Clean performance data for throughput over route scaling
  const throughputData = performanceMatrix
    .filter(item => item.phase === 'scaling-up')
    .map(item => ({
      routes: item.routes,
      throughput: Math.round(item.throughput),
      latency: Number(item.meanLatency.toFixed(1))
    }));

  // Memory usage data
  const memoryData = benchmarkResults
    .filter(item => item.phase === 'scaling-up')
    .map(item => ({
      routes: item.routes,
      gateway: Math.round(item.resources.envoyGateway.memory.mean),
      proxy: Math.round(item.resources.envoyProxy.memory.mean),
      total: Math.round(item.resources.envoyGateway.memory.mean + item.resources.envoyProxy.memory.mean)
    }));

  // Generate latency distribution dynamically from benchmark results at max scale (worst case)
  const generateLatencyDistribution = () => {
    // If we have latencyPercentileComparison data, use it
    if (latencyPercentileComparison && latencyPercentileComparison.length > 0) {
      const maxScaleData = latencyPercentileComparison
        .filter(item => item.phase === 'scaling-up')
        .sort((a, b) => b.routes - a.routes)[0]; // Get highest route count

      if (maxScaleData) {
        return [
          { percentile: 'P50', value: Number(maxScaleData.p50.toFixed(1)), status: 'excellent' },
          { percentile: 'P75', value: Number(maxScaleData.p75.toFixed(1)), status: 'excellent' },
          { percentile: 'P90', value: Number(maxScaleData.p90.toFixed(1)), status: 'good' },
          { percentile: 'P95', value: Number(maxScaleData.p95.toFixed(1)), status: 'acceptable' },
          { percentile: 'P99', value: Number(maxScaleData.p99.toFixed(1)), status: 'watch' }
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
          { percentile: 'P50', value: Number((percentiles.p50 / 1000).toFixed(1)), status: 'excellent' },
          { percentile: 'P75', value: Number((percentiles.p75 / 1000).toFixed(1)), status: 'excellent' },
          { percentile: 'P90', value: Number((percentiles.p90 / 1000).toFixed(1)), status: 'good' },
          { percentile: 'P95', value: Number((percentiles.p95 / 1000).toFixed(1)), status: 'acceptable' },
          { percentile: 'P99', value: Number((percentiles.p99 / 1000).toFixed(1)), status: 'watch' }
        ];
      }
    }

    // Return null when no data is available
    return null;
  };

  const latencyDistribution = generateLatencyDistribution();

  // Memory efficiency data (from ResourcesTab)
  const memoryEfficiencyData = memoryData.map(item => ({
    routes: item.routes,
    gatewayPerRoute: Number((item.gateway / item.routes).toFixed(2)),
    proxyPerRoute: Number((item.proxy / item.routes).toFixed(2)),
    totalPerRoute: Number((item.total / item.routes).toFixed(2))
  }));

  const chartConfig = {
    throughput: {
      label: "Throughput",
      color: "#8b5cf6", // Purple-500
    },
    latency: {
      label: "Latency",
      color: "#6366f1", // Indigo-500
    },
    gateway: {
      label: "Gateway",
      color: "#a855f7", // Purple-400
    },
    proxy: {
      label: "Proxy",
      color: "#4f46e5", // Indigo-600
    },
    totalPerRoute: {
      label: "Total per Route",
      color: "#8b5cf6", // Purple-500
    },
    gatewayPerRoute: {
      label: "Gateway per Route",
      color: "#a855f7", // Purple-400
    },
    proxyPerRoute: {
      label: "Proxy per Route",
      color: "#4f46e5", // Indigo-600
    },
  };

  return (
    <div className="space-y-6">
      {/* Key Performance Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg Throughput</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{Math.round(performanceSummary.avgThroughput)} RPS</div>
            <p className="text-xs text-muted-foreground">
              Consistent across all scales
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Mean Response Time</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{(performanceSummary.avgLatency / 1000).toFixed(1)}ms</div>
            <p className="text-xs text-muted-foreground">
              End-to-end as measured by Nighthawk
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Max Routes in Test</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{performanceSummary.maxRoutes}</div>
            <p className="text-xs text-muted-foreground">
              Routes tested successfully
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Reliability</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">100%</div>
            <p className="text-xs text-muted-foreground">
              Perfect system reliability
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Main Performance Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Throughput Performance */}
        <Card>
          <CardHeader>
            <CardTitle>Throughput Consistency</CardTitle>
            <CardDescription>
              Throughput remains stable across different route scales
            </CardDescription>
          </CardHeader>
          <CardContent>
            {throughputData && throughputData.length > 0 ? (
              <div className="relative">
                <ChartWatermark />
                <ChartContainer config={chartConfig}>
                  <AreaChart data={throughputData}>
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
                      domain={[0, 'dataMax']}
                      tickLine={false}
                      axisLine={false}
                      tickMargin={8}
                      tickFormatter={(value) => `${value}`}
                    />
                    <ChartTooltip
                      content={<ChartTooltipContent
                        formatter={(value, name) => [
                          `${value} RPS`,
                          "Throughput"
                        ]}
                        labelFormatter={(value) => `${value} routes`}
                      />}
                    />
                    <Area
                      dataKey="throughput"
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

        {/* Memory Scaling */}
        <Card>
          <CardHeader>
            <CardTitle>Memory Usage</CardTitle>
            <CardDescription>
              Memory scaling patterns for Gateway and Proxy components
            </CardDescription>
          </CardHeader>
          <CardContent>
            {memoryData && memoryData.length > 0 ? (
              <div className="relative">
                <ChartWatermark />
                <ChartContainer config={chartConfig}>
                  <AreaChart data={memoryData}>
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
                      tickFormatter={(value) => `${value}MB`}
                    />
                    <ChartTooltip
                      content={<ChartTooltipContent />}
                    />
                    <Area
                      dataKey="gateway"
                      stackId="memory"
                      type="monotone"
                      fill="#a855f7"
                      stroke="#a855f7"
                    />
                    <Area
                      dataKey="proxy"
                      stackId="memory"
                      type="monotone"
                      fill="#4f46e5"
                      stroke="#4f46e5"
                    />
                  </AreaChart>
                </ChartContainer>
              </div>
            ) : (
              <ChartPlaceholder />
            )}
          </CardContent>
        </Card>
      </div>

      {/* Secondary Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Latency Distribution */}
        <Card>
          <CardHeader>
            <CardTitle>Request RTT Distribution</CardTitle>
            <CardDescription>
              Latency percentiles at 1000 routes (worst case)
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
                          "Latency"
                        ]}
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

        {/* Memory Efficiency */}
        <Card>
          <CardHeader>
            <CardTitle>Memory Efficiency</CardTitle>
            <CardDescription>
              Memory usage per route shows how efficiently memory scales with route count
            </CardDescription>
          </CardHeader>
          <CardContent>
            {memoryEfficiencyData && memoryEfficiencyData.length > 0 ? (
              <div className="relative">
                <ChartWatermark />
                <ChartContainer config={chartConfig}>
                  <LineChart data={memoryEfficiencyData}>
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
                      tickFormatter={(value) => `${value}MB`}
                    />
                    <ChartTooltip
                      content={<ChartTooltipContent
                        formatter={(value, name) => [
                          `${value}MB per route`,
                          name === 'gatewayPerRoute' ? 'Gateway' :
                          name === 'proxyPerRoute' ? 'Proxy' : 'Total'
                        ]}
                        labelFormatter={(value) => `${value} routes`}
                      />}
                    />
                    <Line
                      dataKey="totalPerRoute"
                      type="monotone"
                      stroke="#8b5cf6"
                      strokeWidth={3}
                      dot={{ fill: "#8b5cf6", strokeWidth: 2, r: 4 }}
                    />
                    <Line
                      dataKey="gatewayPerRoute"
                      type="monotone"
                      stroke="#a855f7"
                      strokeWidth={2}
                      strokeDasharray="5 5"
                      dot={{ fill: "#a855f7", strokeWidth: 2, r: 3 }}
                    />
                    <Line
                      dataKey="proxyPerRoute"
                      type="monotone"
                      stroke="#4f46e5"
                      strokeWidth={2}
                      strokeDasharray="3 3"
                      dot={{ fill: "#4f46e5", strokeWidth: 2, r: 3 }}
                    />
                  </LineChart>
                </ChartContainer>
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

export default OverviewTab;
