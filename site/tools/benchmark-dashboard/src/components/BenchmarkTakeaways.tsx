
import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Separator } from '@/components/ui/separator';
import {
  CheckCircle,
  Clock,
  MemoryStick,
  Cpu,
  Zap,
  TrendingUp,
  Target,
  Shield,
  Database,
  Server,
  Users,
  BookOpen,
  Lightbulb,
  Info
} from 'lucide-react';

interface BenchmarkTakeawaysProps {
  version?: string;
}

export const BenchmarkTakeaways: React.FC<BenchmarkTakeawaysProps> = ({
  version = '1.4.2'
}) => {
  const performanceInsights = [
    {
      icon: CheckCircle,
      title: 'Throughput Stays Consistent',
      description: 'Your request processing performance won\'t degrade as you add more routes (10-1000 routes tested)',
      category: 'Performance'
    },
    {
      icon: Clock,
      title: 'User Experience Stays Reliable',
      description: 'Total response time remains low (6-8ms mean, ~50ms P95) regardless of route count',
      category: 'Performance'
    },
    {
      icon: Target,
      title: 'Consistently Fast Response Times',
      description: 'Median response time stays around 3ms whether you have 10 routes or 1000 routes',
      category: 'Latency'
    },
    {
      icon: TrendingUp,
      title: 'No Performance Degradation',
      description: 'Adding more routes doesn\'t slow down your existing traffic',
      category: 'Latency'
    }
  ];

  const resourceInsights = [
    {
      icon: MemoryStick,
      title: 'Predictable Infrastructure Costs',
      description: 'Control plane memory grows predictably (~0.09MB per route) making capacity planning simple',
      category: 'Planning'
    },
    {
      icon: Zap,
      title: 'Data Plane Gets More Efficient per Route',
      description: 'Proxy memory plateaus around 128MB, so your per-route costs actually decrease as you add more routes',
      category: 'Cost'
    },
    {
      icon: Cpu,
      title: 'Standard Kubernetes Resources Work Fine',
      description: 'CPU stays manageable (<4% control plane, <17% data plane average) with 1 CPU core limits',
      category: 'Infrastructure'
    },
    {
      icon: TrendingUp,
      title: 'Rolling Updates & Cleanup Are Low Risk',
      description: 'Scale-down performs even better than scale-up, making route updates and cleanup safe',
      category: 'Operations'
    }
  ];

  const productionRecommendations = [
    {
      component: 'Envoy Gateway (Control Plane)',
      icon: Database,
      memory: '200-250MB (linear scaling ~0.09MB/route)',
      cpu: '1 core (allow 2-4 cores for config update spikes)',
      pattern: 'Predictable linear growth'
    },
    {
      component: 'Envoy Proxy (Data Plane)',
      icon: Server,
      memory: '130MB plateau (step increases, then stable)',
      cpu: '1 core avg (allow bursts to 100% during updates)',
      pattern: 'Step-wise growth with plateau effect'
    }
  ];

  return (
    <div className="bg-muted/30 border-t">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Header */}
        <div className="text-center mb-12">
          <div className="flex items-center justify-center mb-4">
            <div className="bg-gradient-to-r from-purple-600 to-indigo-600 p-3 rounded-xl mr-4 shadow-sm">
              <BookOpen className="h-6 w-6 text-white" />
            </div>
            <div className="text-left">
              <h2 className="text-2xl font-bold text-foreground mb-1">
                Key Takeaways
              </h2>
              <p className="text-sm text-muted-foreground">
                Envoy Gateway v{version} Benchmark Report
              </p>
            </div>
          </div>
          <div className="max-w-4xl mx-auto">
            <p className="text-base text-muted-foreground">
              Essential insights and recommendations for production deployments
            </p>
          </div>
        </div>

        {/* Quick Summary Banner */}
        <div className="mb-10">
          <Alert className="border-2 border-blue-200 bg-blue-50/50">
            <CheckCircle className="h-4 w-4 text-blue-600" />
            <AlertDescription className="text-sm">
              <strong className="font-semibold text-blue-800">Production Ready Summary:</strong>
              <span className="text-blue-700"> Envoy Gateway v{version} demonstrates consistent performance across all scales (10-1000 routes)
              with predictable resource usage and reliable latency characteristics suitable for enterprise production environments.</span>
            </AlertDescription>
          </Alert>
        </div>

        {/* Performance & Latency Insights */}
        <div className="mb-10">
          <div className="flex items-center gap-3 mb-6">
            <Zap className="h-5 w-5 text-purple-600" />
            <h3 className="text-lg font-bold text-foreground">Performance & User Experience</h3>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {performanceInsights.map((insight, index) => {
              const Icon = insight.icon;
              return (
                <Card key={index} className="hover:shadow-md transition-shadow duration-200">
                  <CardContent className="p-4">
                    <div className="flex items-start space-x-3">
                      <div className="p-2 rounded-lg bg-gradient-to-r from-purple-100 to-indigo-100 flex-shrink-0">
                        <Icon className="h-4 w-4 text-purple-600" />
                      </div>
                      <div className="flex-grow">
                        <div className="flex items-center gap-2 mb-2">
                          <h4 className="text-sm font-semibold text-foreground">
                            {insight.title}
                          </h4>
                          <Badge variant="secondary" className="text-xs">
                            {insight.category}
                          </Badge>
                        </div>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                          {insight.description}
                        </p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </div>

        {/* Resource Planning */}
        <div className="mb-10">
          <div className="flex items-center gap-3 mb-6">
            <MemoryStick className="h-5 w-5 text-indigo-600" />
            <h3 className="text-lg font-bold text-foreground">Resource Planning & Cost Optimization</h3>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {resourceInsights.map((insight, index) => {
              const Icon = insight.icon;
              return (
                <Card key={index} className="hover:shadow-md transition-shadow duration-200">
                  <CardContent className="p-4">
                    <div className="flex items-start space-x-3">
                      <div className="p-2 rounded-lg bg-gradient-to-r from-indigo-100 to-blue-100 flex-shrink-0">
                        <Icon className="h-4 w-4 text-indigo-600" />
                      </div>
                      <div className="flex-grow">
                        <div className="flex items-center gap-2 mb-2">
                          <h4 className="text-sm font-semibold text-foreground">
                            {insight.title}
                          </h4>
                          <Badge variant="secondary" className="text-xs">
                            {insight.category}
                          </Badge>
                        </div>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                          {insight.description}
                        </p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </div>

        {/* Production Sizing Recommendations */}
        <Card className="mb-10 shadow-sm">
          <CardHeader className="bg-gradient-to-r from-blue-50 to-indigo-50">
            <div className="flex items-center gap-3">
              <Server className="h-5 w-5 text-blue-600" />
              <div>
                <CardTitle className="text-lg">Production Sizing Recommendations</CardTitle>
                <CardDescription className="text-sm mt-1">
                  Concrete resource allocations based on benchmark results
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="p-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {productionRecommendations.map((rec, index) => {
                const Icon = rec.icon;
                return (
                  <div key={index} className="p-4 rounded-lg border bg-gradient-to-r from-blue-50/50 to-indigo-50/50 hover:shadow-sm transition-shadow">
                    <div className="flex items-center gap-3 mb-4">
                      <Icon className="h-5 w-5 text-blue-600" />
                      <h4 className="text-base font-semibold text-foreground">{rec.component}</h4>
                    </div>
                    <div className="space-y-3">
                      <div className="flex justify-between items-start">
                        <span className="font-medium text-foreground text-sm">Memory:</span>
                        <span className="text-muted-foreground text-right flex-1 ml-3 text-sm">{rec.memory}</span>
                      </div>
                      <div className="flex justify-between items-start">
                        <span className="font-medium text-foreground text-sm">CPU:</span>
                        <span className="text-muted-foreground text-right flex-1 ml-3 text-sm">{rec.cpu}</span>
                      </div>
                      <div className="flex justify-between items-start">
                        <span className="font-medium text-foreground text-sm">Pattern:</span>
                        <span className="text-muted-foreground text-right flex-1 ml-3 text-sm">{rec.pattern}</span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>

        {/* Bottom Line */}
        <Card className="shadow-sm border-2 border-purple-200 bg-gradient-to-r from-purple-50/50 to-indigo-50/50">
          <CardHeader className="text-center pb-3">
            <div className="flex items-center justify-start gap-3 mb-3">
              <div className="bg-gradient-to-r from-purple-600 to-indigo-600 p-3 rounded-xl shadow-sm">
                <Lightbulb className="h-6 w-6 text-white" />
              </div>
              <div className="text-left">
                <CardTitle className="text-xl text-foreground">Bottom Line</CardTitle>
                <CardDescription className="text-sm text-muted-foreground mt-1">
                  What this means for your production deployment
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="px-6 pb-6">

            <div className="bg-white rounded-lg p-5 shadow-sm border-2 border-purple-200">
              <div className="flex items-start gap-3">
                <div className="bg-gradient-to-r from-purple-600 to-indigo-600 p-2 rounded-lg shadow-sm flex-shrink-0">
                  <CheckCircle className="h-4 w-4 text-white" />
                </div>
                <div>
                  <h4 className="text-base font-bold text-foreground mb-2">Ready for Production</h4>
                  <p className="text-muted-foreground text-sm leading-relaxed">
                    Envoy Gateway v{version} demonstrates production-ready performance with predictable resource usage,
                    consistent latency, and reliable throughput across all tested scales (10-1000 routes).
                  </p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};
