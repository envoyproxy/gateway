
import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Shield } from 'lucide-react';

const ResilienceOverview = () => {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Shield className="h-6 w-6 mr-2 text-purple-600" />
          Resilience Testing Overview
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-4">
        <div className="bg-muted p-4 rounded-lg mb-6">
          <p className="text-sm">
            <strong>Important:</strong> These resilience tests are completely separate from the performance benchmark tests.
            While benchmark tests measure <em>speed and throughput</em>, resilience tests validate <em>fault tolerance and recovery</em>.
          </p>
        </div>
        <p className="text-muted-foreground leading-relaxed">
          The Envoy Gateway resilience tests are a comprehensive fault tolerance test suite designed to validate
          operational stability under various failure scenarios. These tests ensure that Envoy Gateway can handle
          real-world operational challenges gracefully while maintaining service availability for end users during
          infrastructure failures, maintenance windows, and configuration errors.
        </p>
      </CardContent>
    </Card>
  );
};

export default ResilienceOverview;
