import React, { useState, useMemo } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart';
import { ChartPlaceholder } from '@/components/common/DataPlaceholder';
import ChartWatermark from '@/components/common/ChartWatermark';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer,
  Legend
} from 'recharts';
import {
  getAvailableVersions,
  generatePerformanceComparison,
  generateLatencyComparison,
  generateResourceComparison
} from '@/data';

const VersionComparisonTab = () => {
  const availableVersions = getAvailableVersions();
  const [selectedVersions, setSelectedVersions] = useState<string[]>(availableVersions);
  const [selectedPhase, setSelectedPhase] = useState<'scaling-up' | 'scaling-down' | 'both'>('both');

  // Generate comparison data based on selections
  const performanceData = generatePerformanceComparison(selectedVersions);
  const latencyData = generateLatencyComparison(selectedVersions);
  const resourceData = generateResourceComparison(selectedVersions);

  // Filter by phase
  const filteredPerformanceData = selectedPhase === 'both'
    ? performanceData
    : performanceData.filter(d => d.phase === selectedPhase);

  const filteredLatencyData = selectedPhase === 'both'
    ? latencyData
    : latencyData.filter(d => d.phase === selectedPhase);

  // Prepare data for throughput comparison chart
  const throughputComparisonData = filteredPerformanceData
    .reduce((acc, curr) => {
      const key = `${curr.routes}-${curr.phase}`;
      if (!acc[key]) {
        acc[key] = { routes: curr.routes, phase: curr.phase };
      }
      acc[key][curr.version] = curr.throughput;
      return acc;
    }, {} as any);

  const throughputChartData = Object.values(throughputComparisonData);

  // Prepare data for latency comparison chart (P95)
  const latencyComparisonData = filteredLatencyData
    .reduce((acc, curr) => {
      const key = `${curr.routes}-${curr.phase}`;
      if (!acc[key]) {
        acc[key] = { routes: curr.routes, phase: curr.phase };
      }
      acc[key][curr.version] = Math.round(curr.p95 * 100) / 100; // Round to 2 decimals
      return acc;
    }, {} as any);

  const latencyChartData = Object.values(latencyComparisonData);

  // Color palette for different versions
  const versionColors = [
    '#8884d8', '#82ca9d', '#ffc658', '#ff7300', '#00ff00',
    '#ff00ff', '#00ffff', '#ffff00', '#ff0000', '#0000ff'
  ];

  const handleVersionToggle = (version: string, checked: boolean) => {
    if (checked) {
      setSelectedVersions([...selectedVersions, version]);
    } else {
      setSelectedVersions(selectedVersions.filter(v => v !== version));
    }
  };

  return (
    <div className="space-y-6">
      {/* Version Selection Controls */}
      <Card>
        <CardHeader>
          <CardTitle>Version Comparison Controls</CardTitle>
          <CardDescription>
            Select versions and phases to compare performance metrics
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Version Selection */}
            <div>
              <label className="text-sm font-medium mb-3 block">Select Versions to Compare</label>
              <div className="space-y-2">
                {availableVersions.map((version, index) => (
                  <div key={version} className="flex items-center space-x-2">
                    <Checkbox
                      id={version}
                      checked={selectedVersions.includes(version)}
                      onCheckedChange={(checked) => handleVersionToggle(version, checked as boolean)}
                    />
                    <label htmlFor={version} className="text-sm">
                      <span className="inline-block w-3 h-3 mr-2 rounded"
                            style={{ backgroundColor: versionColors[index % versionColors.length] }}></span>
                      Version {version}
                    </label>
                  </div>
                ))}
              </div>
            </div>

            {/* Phase Selection */}
            <div>
              <label className="text-sm font-medium mb-3 block">Test Phase</label>
              <Select value={selectedPhase} onValueChange={(value: any) => setSelectedPhase(value)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="both">Both Scaling Up & Down</SelectItem>
                  <SelectItem value="scaling-up">Scaling Up Only</SelectItem>
                  <SelectItem value="scaling-down">Scaling Down Only</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Throughput Comparison */}
      <Card>
        <CardHeader>
          <CardTitle>Throughput Comparison Across Versions</CardTitle>
          <CardDescription>
            Compare requests per second performance across different Envoy Gateway versions
          </CardDescription>
        </CardHeader>
        <CardContent>
          {throughputChartData && throughputChartData.length > 0 ? (
            <div className="relative">
              <ChartWatermark />
              <ChartContainer config={{}}>
              <LineChart data={throughputChartData} height={400}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="routes"
                  tickLine={false}
                  axisLine={false}
                  tickMargin={8}
                  label={{ value: 'Number of Routes', position: 'insideBottom', offset: -5 }}
                />
                <YAxis
                  tickLine={false}
                  axisLine={false}
                  tickMargin={8}
                  label={{ value: 'Requests/sec', angle: -90, position: 'insideLeft' }}
                />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Legend />
                {selectedVersions.map((version, index) => (
                  <Line
                    key={version}
                    type="monotone"
                    dataKey={version}
                    stroke={versionColors[index % versionColors.length]}
                    strokeWidth={2}
                    dot={{ r: 4 }}
                    name={`v${version}`}
                  />
                ))}
              </LineChart>
            </ChartContainer>
          </div>
        ) : (
          <ChartPlaceholder />
        )}
        </CardContent>
      </Card>

      {/* Latency Comparison */}
      <Card>
        <CardHeader>
          <CardTitle>P95 Latency Comparison</CardTitle>
          <CardDescription>
            Compare 95th percentile latency across versions (lower is better)
          </CardDescription>
        </CardHeader>
        <CardContent>
          {latencyChartData && latencyChartData.length > 0 ? (
            <div className="relative">
              <ChartWatermark />
              <ChartContainer config={{}}>
              <LineChart data={latencyChartData} height={400}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="routes"
                  tickLine={false}
                  axisLine={false}
                  tickMargin={8}
                  label={{ value: 'Number of Routes', position: 'insideBottom', offset: -5 }}
                />
                <YAxis
                  tickLine={false}
                  axisLine={false}
                  tickMargin={8}
                  label={{ value: 'P95 Latency (ms)', angle: -90, position: 'insideLeft' }}
                />
                <ChartTooltip
                  content={<ChartTooltipContent
                    formatter={(value, name) => [`${value}ms`, `v${name}`]}
                  />}
                />
                <Legend />
                {selectedVersions.map((version, index) => (
                  <Line
                    key={version}
                    type="monotone"
                    dataKey={version}
                    stroke={versionColors[index % versionColors.length]}
                    strokeWidth={2}
                    dot={{ r: 4 }}
                    name={`v${version}`}
                  />
                ))}
              </LineChart>
            </ChartContainer>
            </div>
          ) : (
            <ChartPlaceholder />
          )}
        </CardContent>
      </Card>

      {/* Summary Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {selectedVersions.map((version, index) => {
          const versionPerformanceData = filteredPerformanceData.filter(d => d.version === version);
          const avgThroughput = versionPerformanceData.length > 0
            ? Math.round(versionPerformanceData.reduce((sum, d) => sum + d.throughput, 0) / versionPerformanceData.length)
            : 0;
          const avgLatency = versionPerformanceData.length > 0
            ? Math.round((versionPerformanceData.reduce((sum, d) => sum + d.meanLatency, 0) / versionPerformanceData.length) * 10) / 10
            : 0;

          return (
            <Card key={version}>
              <CardHeader className="pb-2">
                <CardTitle className="text-base flex items-center">
                  <span className="inline-block w-3 h-3 mr-2 rounded"
                        style={{ backgroundColor: versionColors[index % versionColors.length] }}></span>
                  Version {version}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Avg Throughput:</span>
                    <span className="font-medium">{avgThroughput} RPS</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Avg Latency:</span>
                    <span className="font-medium">{avgLatency} ms</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Tests:</span>
                    <span className="font-medium">{versionPerformanceData.length}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
};

export default VersionComparisonTab;
