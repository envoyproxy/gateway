
import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Activity, Clock, MemoryStick, Server } from 'lucide-react';

interface SummaryCardsProps {
  performanceSummary: {
    avgThroughput: number;
    avgLatency: number;
    maxRoutes: number;
    totalTests: number;
  };
  benchmarkResults: any[];
}

const SummaryCards = ({ performanceSummary, benchmarkResults }: SummaryCardsProps) => {
  // Calculate dynamic values from the actual data
  const maxThroughput = Math.round(Math.max(...benchmarkResults.map(r => r.throughput)));
  const meanLatencyMs = Math.round(performanceSummary.avgLatency / 1000);
  const maxMemoryMB = Math.round(Math.max(...benchmarkResults.map(r =>
    r.resources.envoyGateway.memory.mean + r.resources.envoyProxy.memory.mean
  )));
  const maxRoutes = performanceSummary.maxRoutes;
  return (
    <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
      <Card className="bg-gradient-to-r from-purple-600 to-indigo-600 text-white">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium flex items-center">
            <Activity className="h-4 w-4 mr-2" />
            Max Throughput in Test
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{maxThroughput.toLocaleString()} RPS</div>
          <p className="text-purple-100 text-sm">Requests per second</p>
        </CardContent>
      </Card>

      <Card className="bg-gradient-to-r from-purple-600 to-indigo-600 text-white">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium flex items-center">
            <Clock className="h-4 w-4 mr-2" />
            Mean Response Time
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{meanLatencyMs}ms</div>
          <p className="text-purple-100 text-sm">End-to-end via Nighthawk</p>
        </CardContent>
      </Card>

      <Card className="bg-gradient-to-r from-purple-600 to-indigo-600 text-white">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium flex items-center">
            <MemoryStick className="h-4 w-4 mr-2" />
            Memory Usage
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{maxMemoryMB}MB</div>
          <p className="text-purple-100 text-sm">Peak at {maxRoutes} routes</p>
        </CardContent>
      </Card>

      <Card className="bg-gradient-to-r from-purple-600 to-indigo-600 text-white">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium flex items-center text">
            <Server className="h-4 w-4 mr-2" />
            Max Routes in Test
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{maxRoutes.toLocaleString()}</div>
          <p className="text-purple-100 text-sm">HTTPRoutes tested</p>
        </CardContent>
      </Card>
    </div>
  );
};

export default SummaryCards;
