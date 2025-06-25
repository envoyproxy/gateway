
import React from 'react';
import { Link } from 'react-router-dom';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Settings, Activity, BarChart3, Monitor, ExternalLink, TrendingUp, TrendingDown, GitBranch } from 'lucide-react';

const TestingMethodology = () => {
  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <CardTitle className="flex items-center">
              <Settings className="h-5 w-5 mr-2 text-blue-600" />
              Testing Methodology
            </CardTitle>
            <CardDescription>Performance benchmarking ran for every PR and release, explore the methodology used.</CardDescription>
          </div>
          <Link to="/testing-details" className="flex items-center text-blue-600 hover:text-blue-800 text-sm font-medium transition-colors self-start sm:self-auto">
            View Detailed Methodology
            <ExternalLink className="h-4 w-4 ml-1" />
          </Link>
        </div>
        
        {/* Environment Badge */}
        <div className="flex items-center justify-center sm:justify-start gap-2 mt-4 p-3 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-200">
          <GitBranch className="h-4 w-4 text-blue-600" />
          <Badge variant="outline" className="bg-white border-blue-300 text-blue-800 font-semibold px-3 py-1">
            Kind Cluster Environment
          </Badge>
          <span className="text-sm text-blue-700 font-medium">via GitHub CI</span>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Test Execution Overview */}
        <div className="p-4 bg-gray-50 rounded-lg border">
          <h4 className="font-semibold text-gray-900 mb-4 flex items-center text-sm sm:text-base">
            <BarChart3 className="h-4 w-4 mr-2 text-blue-600" />
            12 Benchmark Runs: Bidirectional Scale Testing
          </h4>
          <div className="space-y-4">
            <div className="space-y-3">
              <div className="flex items-center space-x-2">
                <TrendingUp className="h-4 w-4 text-green-600" />
                <span className="font-medium text-gray-900 text-sm sm:text-base">Scale UP (6 runs)</span>
              </div>
              <div className="flex flex-wrap items-center gap-1 sm:gap-2 text-xs sm:text-sm">
                {[10, 50, 100, 300, 500, 1000].map((routes, index) => (
                  <div key={routes} className="flex items-center space-x-1 sm:space-x-2">
                    <Badge variant="outline" className="text-xs px-1 sm:px-2">
                      {routes}
                    </Badge>
                    {index < 5 && <span className="text-gray-400 hidden sm:inline">→</span>}
                  </div>
                ))}
              </div>
            </div>
            <div className="space-y-3">
              <div className="flex items-center space-x-2">
                <TrendingDown className="h-4 w-4 text-red-600" />
                <span className="font-medium text-gray-900 text-sm sm:text-base">Scale DOWN (6 runs)</span>
              </div>
              <div className="flex flex-wrap items-center gap-1 sm:gap-2 text-xs sm:text-sm">
                {[1000, 500, 300, 100, 50, 10].map((routes, index) => (
                  <div key={routes} className="flex items-center space-x-1 sm:space-x-2">
                    <Badge variant="outline" className="text-xs px-1 sm:px-2">
                      {routes}
                    </Badge>
                    {index < 5 && <span className="text-gray-400 hidden sm:inline">→</span>}
                  </div>
                ))}
              </div>
            </div>
          </div>
          <div className="text-xs sm:text-sm text-gray-600 mt-4 flex flex-wrap items-center gap-1">
            <span>Each test:</span>
            <Badge variant="outline" className="text-xs mx-1">10,000 RPS</Badge>
            <span>for</span>
            <Badge variant="outline" className="text-xs mx-1">30 seconds</Badge>
            <span className="block sm:inline mt-1 sm:mt-0">measuring control plane and data plane performance</span>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
          {/* Infrastructure */}
          <div className="space-y-3">
            <div className="flex items-center space-x-2">
              <Settings className="h-4 w-4 text-gray-600" />
              <h4 className="font-semibold text-gray-900 text-sm sm:text-base">Infrastructure</h4>
            </div>
            <ul className="text-xs sm:text-sm text-gray-600 space-y-1">
              <li className="flex items-center gap-2">
                <span>•</span>
                <Badge variant="outline" className="text-xs">Nighthawk</Badge>
                <span>load generator</span>
              </li>
              <li>• Go-based benchmark harness</li>
              <li>• Kubernetes with Kind clusters</li>
              <li>• Prometheus monitoring</li>
            </ul>
          </div>

          {/* Metrics */}
          <div className="space-y-3">
            <div className="flex items-center space-x-2">
              <BarChart3 className="h-4 w-4 text-gray-600" />
              <h4 className="font-semibold text-gray-900 text-sm sm:text-base">Metrics Collected</h4>
            </div>
            <ul className="text-xs sm:text-sm text-gray-600 space-y-1">
              <li>• Latency percentiles (P50, P95, P99)</li>
              <li>• CPU & memory utilization</li>
              <li>• Request success rates</li>
              <li>• Connection pool efficiency</li>
            </ul>
          </div>

          {/* Quality */}
          <div className="space-y-3 sm:col-span-2 lg:col-span-1">
            <div className="flex items-center space-x-2">
              <Monitor className="h-4 w-4 text-gray-600" />
              <h4 className="font-semibold text-gray-900 text-sm sm:text-base">Quality Assurance</h4>
            </div>
            <ul className="text-xs sm:text-sm text-gray-600 space-y-1">
              <li>• Automated CI/CD execution</li>
              <li>• Isolated test environments</li>
              <li>• Multiple test iterations</li>
              <li>• Statistical validation</li>
            </ul>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default TestingMethodology;
