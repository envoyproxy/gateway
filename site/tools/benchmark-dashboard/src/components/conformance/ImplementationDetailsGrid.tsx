import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Target, Settings, Cloud, Network } from 'lucide-react';

const ImplementationDetailsGrid = () => {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Test Infrastructure and Execution */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="text-lg flex items-center">
            <Target className="h-5 w-5 mr-2 text-orange-600" />
            Test Infrastructure & Execution
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-2">
          <div className="grid grid-cols-1 gap-6">
            <div className="space-y-3">
              <h4 className="font-semibold">Automated Test Pipeline</h4>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• Continuous Integration: Tests run on every code change</li>
                <li>• Multi-environment Testing: Various Kubernetes versions</li>
                <li>• Regression Prevention: Ensures new features don't break existing functionality</li>
                <li>• Certification Process: Results submitted for official recognition</li>
              </ul>
            </div>
            <div className="space-y-3">
              <h4 className="font-semibold">Test Methodology</h4>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• Behavior-Driven Development: Human-readable specifications</li>
                <li>• Real-world Scenarios: Mirror actual deployment patterns</li>
                <li>• Comprehensive Coverage: Happy path and error scenarios</li>
                <li>• Version Tracking: Per Gateway API version compatibility</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Environment-Specific Handling */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="text-lg flex items-center">
            <Cloud className="h-5 w-5 mr-2 text-indigo-600" />
            Environment-Specific Handling
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-2">
          <div className="space-y-4">
            <div className="space-y-3">
              <h4 className="font-semibold flex items-center">
                <Settings className="h-4 w-4 mr-2 text-blue-600" />
                Gateway Namespace Mode
              </h4>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• Skips GatewayStaticAddresses tests</li>
                <li>• Skips GatewayInfrastructure tests</li>
                <li>• Focuses on namespace-scoped operations</li>
              </ul>
            </div>
            <div className="space-y-3">
              <h4 className="font-semibold flex items-center">
                <Network className="h-4 w-4 mr-2 text-green-600" />
                IP Family Support
              </h4>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• Auto-detects IPv4, IPv6, or dual-stack</li>
                <li>• Skips unsupported network configurations</li>
                <li>• Adapts tests to cluster capabilities</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Reporting and Certification */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="text-lg flex items-center">
            <Target className="h-5 w-5 mr-2 text-purple-600" />
            Reporting & Certification
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-2">
          <div className="space-y-4">
            <div className="space-y-3">
              <h4 className="font-semibold">Output Formats</h4>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• YAML Reports: Machine-readable conformance reports</li>
                <li>• CI/CD Integration: Pipeline-ready output formats</li>
                <li>• Version Tracking: Conformance per Gateway API version</li>
              </ul>
            </div>
            <div className="space-y-3">
              <h4 className="font-semibold">Ecosystem Impact</h4>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• Validates specification completeness</li>
                <li>• Identifies implementation challenges</li>
                <li>• Demonstrates best practices</li>
                <li>• Builds ecosystem trust</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Requirements */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="text-lg flex items-center">
            <Settings className="h-5 w-5 mr-2 text-red-600" />
            Infrastructure Requirements
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-2">
          <div className="space-y-4">
            <p className="text-muted-foreground">The conformance tests require:</p>
            <ul className="text-muted-foreground space-y-2">
              <li>• Kubernetes cluster with Envoy Gateway installed</li>
              <li>• Standard Gateway API resources support</li>
              <li>• Network connectivity for cross-namespace testing</li>
              <li>• TLS certificate management capabilities</li>
              <li>• Access to create test namespaces and resources</li>
            </ul>
            <div className="mt-6 p-4 bg-muted rounded-lg">
              <p className="text-sm">
                <strong>Note:</strong> These conformance tests transform the Gateway API from a specification document
                into a reliable, production-ready platform for managing ingress traffic in Kubernetes environments.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default ImplementationDetailsGrid;
