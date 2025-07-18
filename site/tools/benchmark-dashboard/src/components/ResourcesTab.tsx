import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart';
import { ChartPlaceholder } from '@/components/common/DataPlaceholder';
import ChartWatermark from '@/components/common/ChartWatermark';
import { AreaChart, Area, LineChart, Line, BarChart, Bar, ComposedChart, XAxis, YAxis, CartesianGrid } from 'recharts';
import { MemoryStick, Cpu, Server, Database, TrendingUp, Zap, Activity, CheckCircle } from 'lucide-react';

interface ResourcesTabProps {
  resourceTrends: any[];
  benchmarkResults: any[];
}

const ResourcesTab = ({ resourceTrends, benchmarkResults }: ResourcesTabProps) => {
  // Clean memory data for scaling up phase with proper numeric scaling
  const memoryData = benchmarkResults
    .filter(item => item.phase === 'scaling-up')
    .map(item => ({
      routes: item.routes, // Use actual route values
      gateway: Math.round(item.resources.envoyGateway.memory.mean),
      proxy: Math.round(item.resources.envoyProxy.memory.mean),
      total: Math.round(item.resources.envoyGateway.memory.mean + item.resources.envoyProxy.memory.mean)
    }));

  // CPU usage data with proper numeric scaling
  const cpuData = benchmarkResults
    .filter(item => item.phase === 'scaling-up')
    .map(item => ({
      routes: item.routes, // Use actual route values
      gatewayMean: Number(item.resources.envoyGateway.cpu.mean.toFixed(1)),
      gatewayMax: Number(item.resources.envoyGateway.cpu.max.toFixed(1)),
      proxyMean: Number(item.resources.envoyProxy.cpu.mean.toFixed(1)),
      proxyMax: Number(item.resources.envoyProxy.cpu.max.toFixed(1))
    }));

  // Resource efficiency data with proper numeric scaling
  const efficiencyData = memoryData.map(item => ({
    routes: item.routes, // Use actual route values
    gatewayPerRoute: Number((item.gateway / item.routes).toFixed(2)),
    proxyPerRoute: Number((item.proxy / item.routes).toFixed(2)),
    totalPerRoute: Number((item.total / item.routes).toFixed(2))
  }));

  // Resource scaling summary (calculated from actual data)
  const calculateScalingSummary = () => {
    if (memoryData.length === 0) {
      return [
        { component: 'Gateway', min: 0, max: 0, scaling: 'No Data', efficiency: 'Unknown' },
        { component: 'Proxy', min: 0, max: 0, scaling: 'No Data', efficiency: 'Unknown' },
        { component: 'Total', min: 0, max: 0, scaling: 'No Data', efficiency: 'Unknown' }
      ];
    }

    const gatewayMemories = memoryData.map(d => d.gateway);
    const proxyMemories = memoryData.map(d => d.proxy);
    const totalMemories = memoryData.map(d => d.total);

    return [
      {
        component: 'Gateway',
        min: Math.min(...gatewayMemories),
        max: Math.max(...gatewayMemories),
        scaling: analyzeMemoryScalingPattern(memoryData.map(d => d.routes), gatewayMemories),
        efficiency: analyzeMemoryEfficiency(memoryData.map(d => d.routes), gatewayMemories)
      },
      {
        component: 'Proxy',
        min: Math.min(...proxyMemories),
        max: Math.max(...proxyMemories),
        scaling: analyzeMemoryScalingPattern(memoryData.map(d => d.routes), proxyMemories),
        efficiency: analyzeMemoryEfficiency(memoryData.map(d => d.routes), proxyMemories)
      },
      {
        component: 'Total',
        min: Math.min(...totalMemories),
        max: Math.max(...totalMemories),
        scaling: analyzeMemoryScalingPattern(memoryData.map(d => d.routes), totalMemories),
        efficiency: analyzeMemoryEfficiency(memoryData.map(d => d.routes), totalMemories)
      }
    ];
  };

  // Function to analyze memory scaling patterns from actual data
  const analyzeMemoryScalingPattern = (routes: number[], memoryValues: number[]) => {
    if (routes.length < 3) return 'Insufficient Data';

    // Calculate correlation coefficient for memory linearity
    const n = routes.length;
    const sumX = routes.reduce((sum, x) => sum + x, 0);
    const sumY = memoryValues.reduce((sum, y) => sum + y, 0);
    const sumXY = routes.reduce((sum, x, i) => sum + x * memoryValues[i], 0);
    const sumX2 = routes.reduce((sum, x) => sum + x * x, 0);
    const sumY2 = memoryValues.reduce((sum, y) => sum + y * y, 0);

    const numerator = n * sumXY - sumX * sumY;
    const denominator = Math.sqrt((n * sumX2 - sumX * sumX) * (n * sumY2 - sumY * sumY));
    const correlation = denominator === 0 ? 0 : numerator / denominator;
    const rSquared = correlation * correlation;

    // Calculate differences between consecutive memory measurements to detect step-wise behavior
    const memoryDifferences = [];
    for (let i = 1; i < memoryValues.length; i++) {
      memoryDifferences.push(memoryValues[i] - memoryValues[i - 1]);
    }

    // Coefficient of variation in memory differences (high = step-wise, low = smooth)
    const avgDiff = memoryDifferences.reduce((sum, d) => sum + d, 0) / memoryDifferences.length;
    const varianceDiff = memoryDifferences.reduce((sum, d) => sum + Math.pow(d - avgDiff, 2), 0) / memoryDifferences.length;
    const stdDevDiff = Math.sqrt(varianceDiff);
    const coefficientOfVariation = avgDiff === 0 ? 0 : Math.abs(stdDevDiff / avgDiff);

    // Determine memory scaling pattern
    if (rSquared > 0.95) {
      return 'Highly Linear';
    } else if (rSquared > 0.85) {
      return 'Linear';
    } else if (coefficientOfVariation > 1.5) {
      return 'Step-wise';
    } else if (rSquared > 0.7) {
      return 'Moderately Linear';
    } else {
      return 'Variable';
    }
  };

  // Function to analyze memory efficiency characteristics
  const analyzeMemoryEfficiency = (routes: number[], memoryValues: number[]) => {
    if (routes.length < 2) return 'Unknown';

    // Calculate memory per route at different scales
    const memoryPerRouteRatios = routes.map((route, i) => memoryValues[i] / route);

    // Check if memory efficiency improves (memory-per-route decreases) with scale
    const firstRatio = memoryPerRouteRatios[0];
    const lastRatio = memoryPerRouteRatios[memoryPerRouteRatios.length - 1];
    const memoryEfficiencyImprovement = (firstRatio - lastRatio) / firstRatio;

    // Calculate consistency in memory usage (low standard deviation = more consistent)
    const avgRatio = memoryPerRouteRatios.reduce((sum, r) => sum + r, 0) / memoryPerRouteRatios.length;
    const variance = memoryPerRouteRatios.reduce((sum, r) => sum + Math.pow(r - avgRatio, 2), 0) / memoryPerRouteRatios.length;
    const coefficientOfVariation = Math.sqrt(variance) / avgRatio;

    if (memoryEfficiencyImprovement > 0.3) {
      return 'Excellent';
    } else if (memoryEfficiencyImprovement > 0.1 || coefficientOfVariation < 0.2) {
      return 'Good';
    } else if (coefficientOfVariation < 0.4) {
      return 'Moderate';
    } else {
      return 'Variable';
    }
  };

  const scalingSummary = calculateScalingSummary();

    const chartConfig = {
    gateway: {
      label: "Gateway",
      color: "#a855f7", // Purple-400
    },
    proxy: {
      label: "Proxy",
      color: "#4f46e5", // Indigo-600
    },
    gatewayMean: {
      label: "Gateway Mean",
      color: "#8b5cf6", // Purple-500
    },
    gatewayMax: {
      label: "Gateway Peak",
      color: "#c084fc", // Purple-300
    },
    proxyMean: {
      label: "Proxy Mean",
      color: "#6366f1", // Indigo-500
    },
    proxyMax: {
      label: "Proxy Peak",
      color: "#818cf8", // Indigo-400
    },
  };

  // Calculate key resource metrics from actual data
  const calculateResourceMetrics = () => {
    if (memoryData.length === 0 || cpuData.length === 0 || efficiencyData.length === 0) {
      return {
        gatewayMemoryRange: '0-0MB',
        proxyMemoryRange: '0-0MB',
        peakCPU: 0,
        memoryPerRouteAtScale: 0
      };
    }

    // Calculate memory ranges
    const gatewayMemories = memoryData.map(d => d.gateway);
    const proxyMemories = memoryData.map(d => d.proxy);

    const gatewayMin = Math.min(...gatewayMemories);
    const gatewayMax = Math.max(...gatewayMemories);
    const proxyMin = Math.min(...proxyMemories);
    const proxyMax = Math.max(...proxyMemories);

    // Calculate peak CPU from all components
    const allCPUMaxValues = cpuData.flatMap(d => [d.gatewayMax, d.proxyMax]);
    const peakCPU = Math.max(...allCPUMaxValues);

    // Get memory per route at highest scale (most efficient point)
    const highestScaleEfficiency = efficiencyData[efficiencyData.length - 1];
    const memoryPerRouteAtScale = highestScaleEfficiency?.totalPerRoute || 0;

    return {
      gatewayMemoryRange: `${gatewayMin}-${gatewayMax}MB`,
      proxyMemoryRange: `${proxyMin}-${proxyMax}MB`,
      peakCPU: Math.round(peakCPU),
      memoryPerRouteAtScale: memoryPerRouteAtScale
    };
  };

  const resourceMetrics = calculateResourceMetrics();

  return (
    <div className="space-y-6">
      {/* Key Resource Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Gateway Memory</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{resourceMetrics.gatewayMemoryRange}</div>
            <p className="text-xs text-muted-foreground">
              Linear scaling pattern
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Proxy Memory</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{resourceMetrics.proxyMemoryRange}</div>
            <p className="text-xs text-muted-foreground">
              Efficient step-wise scaling
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Peak CPU</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{resourceMetrics.peakCPU}%</div>
            <p className="text-xs text-muted-foreground">
              Brief spikes, stable avg
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Efficiency</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{resourceMetrics.memoryPerRouteAtScale.toFixed(2)}MB</div>
            <p className="text-xs text-muted-foreground">
              Memory per route at scale
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Main Resource Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Memory Usage */}
        <Card>
          <CardHeader>
            <CardTitle>Memory Scaling</CardTitle>
            <CardDescription>
              How Gateway and Proxy memory usage grows with route count
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
                    content={<ChartTooltipContent
                      labelFormatter={(value) => `${value} routes`}
                    />}
                  />
                  <Area
                    dataKey="gateway"
                    type="monotone"
                    fill="#a855f7"
                    fillOpacity={0.4}
                    stroke="#a855f7"
                    strokeWidth={2}
                  />
                  <Area
                    dataKey="proxy"
                    type="monotone"
                    fill="#4f46e5"
                    fillOpacity={0.4}
                    stroke="#4f46e5"
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

        {/* CPU Usage */}
        <Card>
          <CardHeader>
            <CardTitle>CPU Usage Patterns</CardTitle>
            <CardDescription>
              Mean vs peak CPU usage showing burst characteristics.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {cpuData && cpuData.length > 0 ? (
              <div className="relative">
                <ChartWatermark />
                <ChartContainer config={chartConfig}>
                <ComposedChart data={cpuData}>
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
                    tickFormatter={(value) => `${value}%`}
                  />
                  <ChartTooltip
                    content={<ChartTooltipContent
                      labelFormatter={(value) => `${value} routes`}
                    />}
                  />
                  <Area
                    dataKey="gatewayMax"
                    type="monotone"
                    fill="#c084fc"
                    fillOpacity={0.1}
                    stroke="#c084fc"
                    strokeWidth={1}
                    strokeDasharray="2 2"
                  />
                  <Area
                    dataKey="proxyMax"
                    type="monotone"
                    fill="#818cf8"
                    fillOpacity={0.1}
                    stroke="#818cf8"
                    strokeWidth={1}
                    strokeDasharray="2 2"
                  />
                  <Line
                    dataKey="gatewayMean"
                    type="monotone"
                    stroke="#8b5cf6"
                    strokeWidth={3}
                    dot={{ fill: "#8b5cf6", strokeWidth: 2, r: 4 }}
                  />
                  <Line
                    dataKey="proxyMean"
                    type="monotone"
                    stroke="#6366f1"
                    strokeWidth={3}
                    dot={{ fill: "#6366f1", strokeWidth: 2, r: 4 }}
                  />
                </ComposedChart>
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
        {/* Memory Efficiency */}
        <Card>
          <CardHeader>
            <CardTitle>Memory Efficiency</CardTitle>
            <CardDescription>
              Memory usage per route shows how efficiently memory scales with route count
            </CardDescription>
          </CardHeader>
          <CardContent>
            {efficiencyData && efficiencyData.length > 0 ? (
              <div className="relative">
                <ChartWatermark />
                <ChartContainer config={chartConfig}>
                <LineChart data={efficiencyData}>
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

        {/* Scaling Summary */}
        <Card>
          <CardHeader>
            <CardTitle>Memory Scaling Summary</CardTitle>
            <CardDescription>
              Memory usage characteristics and scaling patterns across different components
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {scalingSummary.map((item, index) => (
                <div key={index} className="flex items-center justify-between p-3 bg-muted/30 rounded-lg">
                  <div className="space-y-1">
                    <p className="font-medium">{item.component}</p>
                    <div className="flex space-x-4 text-sm text-muted-foreground">
                      <span>Range: {item.min}-{item.max}MB</span>
                      <span>Pattern: {item.scaling}</span>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <CheckCircle className="h-4 w-4 text-green-600" />
                    <span className="text-sm font-medium text-green-700">{item.efficiency}</span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default ResourcesTab;
