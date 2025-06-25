import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { CheckCircle } from 'lucide-react';

const ConformanceOverview = () => {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <CheckCircle className="h-6 w-6 mr-2 text-green-600" />
          Executive Summary
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-4">
        <div className="bg-muted p-4 rounded-lg mb-6">
          <p className="text-sm">
            <strong>What Are Conformance Tests?</strong> Conformance tests are standardized test suites that verify whether an implementation
            correctly supports the behaviors defined by the Gateway API specification. They act as both a validation tool and a certification mechanism.
          </p>
        </div>
        <p className="text-muted-foreground leading-relaxed">
          Envoy Gateway's conformance tests serve as a critical quality assurance mechanism that ensures the implementation properly
          adheres to the <strong>Kubernetes Gateway API specification</strong>. These tests validate that Envoy Gateway behaves
          consistently and predictably according to standardized Gateway API semantics, providing confidence to users who deploy
          Gateway resources in production environments.
        </p>
      </CardContent>
    </Card>
  );
};

export default ConformanceOverview;
