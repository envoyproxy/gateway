import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { CheckCircle, Zap, Settings, Shield, Globe, Lock, Route, Server } from 'lucide-react';

const TestCoverageSection = () => {
  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900 flex items-center">
        <CheckCircle className="h-6 w-6 mr-2 text-green-600" />
        Test Coverage Overview
      </h2>

      {/* Standard Conformance Tests */}
      <Card>
        <CardHeader className="pb-4">
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg flex items-center">
              <Shield className="h-5 w-5 mr-2 text-blue-600" />
              Standard Gateway API Conformance Tests
            </CardTitle>
            <Badge variant="default" className="bg-blue-600">Core Features</Badge>
          </div>
          <CardDescription className="mt-2">
            Validates core Gateway API functionality that every Gateway API implementation must support
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6 pt-2">
          <div>
            <h5 className="font-medium mb-3 flex items-center">
              <CheckCircle className="h-4 w-4 mr-2 text-green-600" />
              What is tested
            </h5>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <ul className="text-sm text-muted-foreground space-y-2 ml-6">
                <li className="flex items-center">
                  <Settings className="h-3 w-3 mr-2 text-gray-400" />
                  Gateway and GatewayClass lifecycle
                </li>
                <li className="flex items-center">
                  <Route className="h-3 w-3 mr-2 text-gray-400" />
                  HTTPRoute functionality and routing
                </li>
                <li className="flex items-center">
                  <Lock className="h-3 w-3 mr-2 text-gray-400" />
                  Cross-namespace references with ReferenceGrant
                </li>
              </ul>
              <ul className="text-sm text-muted-foreground space-y-2 ml-6">
                <li className="flex items-center">
                  <Shield className="h-3 w-3 mr-2 text-gray-400" />
                  TLS termination and certificate management
                </li>
                <li className="flex items-center">
                  <CheckCircle className="h-3 w-3 mr-2 text-gray-400" />
                  Status conditions and reporting
                </li>
                <li className="flex items-center">
                  <Server className="h-3 w-3 mr-2 text-gray-400" />
                  Backend selection and health checking
                </li>
              </ul>
            </div>
          </div>

          <div>
            <h5 className="font-medium mb-3">Test profiles included</h5>
            <div className="flex flex-wrap gap-2">
              <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">HTTP Profile</Badge>
              <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">TLS Profile</Badge>
              <Badge variant="outline" className="bg-purple-50 text-purple-700 border-purple-200">GRPC Profile</Badge>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Experimental Conformance Tests */}
      <Card>
        <CardHeader className="pb-4">
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg flex items-center">
              <Zap className="h-5 w-5 mr-2 text-orange-600" />
              Experimental Conformance Tests
            </CardTitle>
            <Badge variant="secondary" className="bg-orange-100 text-orange-700">Experimental Features</Badge>
          </div>
          <CardDescription className="mt-2">
            Validates cutting-edge Gateway API features that are still in experimental status but supported by Envoy Gateway
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6 pt-2">
          <div>
            <h5 className="font-medium mb-3 flex items-center">
              <Zap className="h-4 w-4 mr-2 text-orange-600" />
              What is tested
            </h5>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <ul className="text-sm text-muted-foreground space-y-2 ml-6">
                <li>• Advanced HTTP features and capabilities</li>
                <li>• Emerging Gateway API specifications</li>
                <li>• Extended TLS configurations</li>
              </ul>
              <ul className="text-sm text-muted-foreground space-y-2 ml-6">
                <li>• Experimental GRPC routing features</li>
                <li>• Multiple conformance profiles simultaneously</li>
                <li>• YAML conformance report generation</li>
              </ul>
            </div>
          </div>

          <div>
            <h5 className="font-medium mb-3">Key differences</h5>
            <div className="bg-orange-50 p-4 rounded-lg">
              <ul className="text-sm text-orange-800 space-y-1">
                <li>• Tests experimental Gateway API channel features</li>
                <li>• Generates detailed conformance reports in YAML format</li>
                <li>• Supports multiple conformance profiles simultaneously</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Feature Support Matrix */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="text-lg flex items-center">
            <Globe className="h-5 w-5 mr-2 text-purple-600" />
            Feature Support Matrix
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-2">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <h4 className="font-semibold text-green-700">Core Features (Always Supported)</h4>
              <div className="space-y-2">
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 mr-2 text-green-600" />
                  <span className="text-sm">Gateway resource lifecycle and configuration</span>
                </div>
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 mr-2 text-green-600" />
                  <span className="text-sm">GatewayClass definitions and controller binding</span>
                </div>
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 mr-2 text-green-600" />
                  <span className="text-sm">HTTPRoute traffic routing and manipulation</span>
                </div>
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 mr-2 text-green-600" />
                  <span className="text-sm">ReferenceGrant cross-namespace security</span>
                </div>
              </div>
            </div>

            <div className="space-y-4">
              <h4 className="font-semibold text-blue-700">Extended Features (Configurable Support)</h4>
              <div className="space-y-2">
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 mr-2 text-blue-600" />
                  <span className="text-sm">TLS termination and certificate management</span>
                </div>
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 mr-2 text-blue-600" />
                  <span className="text-sm">GRPC routing and traffic management</span>
                </div>
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 mr-2 text-blue-600" />
                  <span className="text-sm">Header modification and path rewriting</span>
                </div>
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 mr-2 text-blue-600" />
                  <span className="text-sm">Request mirroring and cross-namespace routing</span>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default TestCoverageSection;
