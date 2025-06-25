
import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Target, Settings } from 'lucide-react';

const ImplementationDetailsGrid = () => {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Test Execution Details */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="text-lg flex items-center">
            <Target className="h-5 w-5 mr-2 text-orange-600" />
            Test Execution & Monitoring
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-2">
          <div className="grid grid-cols-1 gap-6">
            <div className="space-y-3">
              <h4 className="font-semibold">Execution Context</h4>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• CI/CD Pipeline integration</li>
                <li>• Pre-release validation</li>
                <li>• Regression testing</li>
                <li>• Isolated test namespace</li>
              </ul>
            </div>
            <div className="space-y-3">
              <h4 className="font-semibold">Key Metrics</h4>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• Translation error rates</li>
                <li>• Control plane connectivity</li>
                <li>• HTTP response validation</li>
                <li>• Leader election status</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Infrastructure Requirements */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="text-lg flex items-center">
            <Settings className="h-5 w-5 mr-2 text-indigo-600" />
            Infrastructure Requirements
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-2">
          <div className="space-y-4">
            <p className="text-muted-foreground">The resilience tests require:</p>
            <ul className="text-muted-foreground space-y-2">
              <li>• Kubernetes cluster with Envoy Gateway installed</li>
              <li>• Multiple replica configurations for HA testing</li>
              <li>• Network policy support for connectivity simulation</li>
              <li>• Prometheus for metrics validation</li>
              <li>• Extension server testing capabilities</li>
            </ul>
            <div className="mt-6 p-4 bg-muted rounded-lg">
              <p className="text-sm">
                <strong>Note:</strong> This comprehensive resilience testing approach provides confidence that Envoy Gateway
                can handle the unpredictable nature of production environments.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default ImplementationDetailsGrid;
