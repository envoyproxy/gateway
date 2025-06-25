
import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Users, ArrowRight, Shield } from 'lucide-react';

const BusinessValueSection = () => {
  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle className="text-xl flex items-center">
          <Users className="h-6 w-6 mr-2 text-blue-600" />
          Why Resilience Testing Matters for Your Operations
        </CardTitle>
        <CardDescription className="mt-2">
          Understanding how resilience tests protect your mission-critical applications from real-world failures
        </CardDescription>
      </CardHeader>
      <CardContent className="pt-2">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div className="space-y-4">
            <h4 className="font-semibold text-lg">Service Availability Guarantee</h4>
            <div className="space-y-3">
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-green-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Zero-downtime operations</strong> during platform maintenance</span>
              </div>
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-green-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Fault tolerance</strong> during unexpected infrastructure failures</span>
              </div>
              <div className="flex items-start">
                <ArrowRight className="h-5 w-5 mr-3 text-green-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Configuration safety</strong> against human and system errors</span>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h4 className="font-semibold text-lg">Operational Confidence</h4>
            <div className="space-y-3">
              <div className="flex items-start">
                <Shield className="h-5 w-5 mr-3 text-purple-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Multi-tenancy safety:</strong> Issues with one tenant don't affect others</span>
              </div>
              <div className="flex items-start">
                <Shield className="h-5 w-5 mr-3 text-purple-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Disaster recovery:</strong> Validated backup and recovery processes</span>
              </div>
              <div className="flex items-start">
                <Shield className="h-5 w-5 mr-3 text-purple-600 mt-0.5 flex-shrink-0" />
                <span className="text-muted-foreground"><strong>Configuration integrity:</strong> Protection against corrupted deployments</span>
              </div>
            </div>
          </div>
        </div>

        <div className="mt-8 p-6 bg-muted rounded-lg">
          <h5 className="font-medium mb-3 text-lg">Real-World Impact</h5>
          <p className="text-muted-foreground leading-relaxed">
            These resilience tests ensure that your mission-critical applications remain available during Kubernetes API server
            maintenance, control plane upgrades, network partitions, and extension server malfunctions. The tests provide confidence
            that Envoy Gateway can handle real-world operational challenges while maintaining service availability for your users.
          </p>
        </div>
      </CardContent>
    </Card>
  );
};

export default BusinessValueSection;
