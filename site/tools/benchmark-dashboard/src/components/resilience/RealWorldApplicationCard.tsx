
import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Target } from 'lucide-react';

const RealWorldApplicationCard = () => {
  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center">
          <Target className="h-6 w-6 mr-2 text-teal-600" />
          Real-World Application Scenarios
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-2">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="space-y-3">
            <h4 className="font-semibold">For Platform Teams</h4>
            <ul className="text-sm text-muted-foreground space-y-1">
              <li>• Disaster Recovery Planning</li>
              <li>• High Availability Design</li>
              <li>• Maintenance Window Planning</li>
              <li>• Fault Tolerance Validation</li>
            </ul>
          </div>
          <div className="space-y-3">
            <h4 className="font-semibold">For Application Teams</h4>
            <ul className="text-sm text-muted-foreground space-y-1">
              <li>• Service Continuity Assurance</li>
              <li>• Configuration Safety</li>
              <li>• Extension Server Development</li>
              <li>• Multi-tenant Isolation</li>
            </ul>
          </div>
          <div className="space-y-3">
            <h4 className="font-semibold">For Site Reliability Engineers</h4>
            <ul className="text-sm text-muted-foreground space-y-1">
              <li>• Incident Response Planning</li>
              <li>• Recovery Procedures</li>
              <li>• Monitoring & Alerting</li>
              <li>• Chaos Engineering</li>
            </ul>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default RealWorldApplicationCard;
