import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Users, ArrowRight, Shield, Target, Zap, Settings, TrendingUp, Globe } from 'lucide-react';

const BusinessValueSection = () => {
  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle className="text-xl flex items-center">
          <Users className="h-6 w-6 mr-2 text-blue-600" />
          Why Conformance Testing Matters: The "So What" for End Users
        </CardTitle>
        <CardDescription className="mt-2">
          Understanding how conformance tests provide reliability, portability, and production readiness guarantees
        </CardDescription>
      </CardHeader>
      <CardContent className="pt-2">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div className="space-y-4">
            <h4 className="font-semibold text-lg flex items-center">
              <Shield className="h-5 w-5 mr-2 text-green-600" />
              Reliability Assurance
            </h4>
            <div className="space-y-3">
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-green-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Predictable behavior:</strong> Gateway resources behave consistently across environments</span>
              </div>
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-green-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Reduced outages:</strong> Fewer unexpected behaviors in production</span>
              </div>
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-green-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Debugging confidence:</strong> Known behavior patterns reduce troubleshooting time</span>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="font-semibold text-lg flex items-center">
              <Globe className="h-5 w-5 mr-2 text-blue-600" />
              Portability Guarantee
            </h4>
            <div className="space-y-3">
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-blue-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Environment consistency:</strong> Same configurations work across different deployments</span>
              </div>
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-blue-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Vendor freedom:</strong> Prevents lock-in to specific implementations</span>
              </div>
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-blue-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Migration safety:</strong> Easier transitions between cloud providers</span>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="font-semibold text-lg flex items-center">
              <Target className="h-5 w-5 mr-2 text-purple-600" />
              Feature Completeness
            </h4>
            <div className="space-y-3">
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-purple-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Documented features work:</strong> No implementation gaps or surprises</span>
              </div>
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-purple-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Specification compliance:</strong> Full Gateway API feature support</span>
              </div>
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-purple-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Version compatibility:</strong> Clear understanding of supported features</span>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="font-semibold text-lg flex items-center">
              <Zap className="h-5 w-5 mr-2 text-orange-600" />
              Production Readiness
            </h4>
            <div className="space-y-3">
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-orange-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Standardized testing:</strong> Validates real-world usage patterns</span>
              </div>
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-orange-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Risk reduction:</strong> Fewer critical issues discovered post-deployment</span>
              </div>
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-orange-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Interoperability:</strong> Works with standard Kubernetes resources</span>
              </div>
            </div>
          </div>
        </div>

        <div className="mt-8 p-6 bg-muted rounded-lg">
          <h5 className="font-medium mb-3 text-lg flex items-center">
            <TrendingUp className="h-5 w-5 mr-2 text-green-600" />
            Compliance & Certification Benefits
          </h5>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h6 className="font-medium mb-2">Enterprise Benefits</h6>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• Satisfies enterprise compliance requirements</li>
                <li>• Provides audit trails for regulatory purposes</li>
                <li>• Reduces vendor evaluation time</li>
                <li>• Enables confident architectural decisions</li>
              </ul>
            </div>
            <div>
              <h6 className="font-medium mb-2">Operational Benefits</h6>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>• Faster incident resolution</li>
                <li>• Reduced training requirements</li>
                <li>• Simplified toolchain integration</li>
                <li>• Improved change management confidence</li>
              </ul>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default BusinessValueSection;
