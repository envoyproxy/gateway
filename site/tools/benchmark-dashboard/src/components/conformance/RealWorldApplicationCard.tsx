import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Target, Users, Briefcase, Wrench } from 'lucide-react';

const RealWorldApplicationCard = () => {
  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center">
          <Target className="h-6 w-6 mr-2 text-red-600" />
          Real-World Application
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-2">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="space-y-3">
            <h4 className="font-semibold text-gray-900 flex items-center">
              <Users className="h-5 w-5 mr-2 text-blue-600" />
              For Platform Teams
            </h4>
            <ul className="text-sm text-gray-700 space-y-1">
              <li>• <strong>Architecture Decisions:</strong> Confident selection of Gateway implementations</li>
              <li>• <strong>Migration Planning:</strong> Risk assessment for Gateway API adoption</li>
              <li>• <strong>Multi-cloud Strategy:</strong> Consistent behavior across providers</li>
              <li>• <strong>Compliance Reporting:</strong> Documentation for enterprise requirements</li>
            </ul>
          </div>
          <div className="space-y-3">
            <h4 className="font-semibold text-gray-900 flex items-center">
              <Briefcase className="h-5 w-5 mr-2 text-green-600" />
              For Application Teams
            </h4>
            <ul className="text-sm text-gray-700 space-y-1">
              <li>• <strong>Configuration Confidence:</strong> Trust in Gateway resource behavior</li>
              <li>• <strong>Development Velocity:</strong> Faster iteration with reliable APIs</li>
              <li>• <strong>Environment Parity:</strong> Same configs work dev to production</li>
              <li>• <strong>Troubleshooting Efficiency:</strong> Predictable debugging workflows</li>
            </ul>
          </div>
          <div className="space-y-3">
            <h4 className="font-semibold text-gray-900 flex items-center">
              <Wrench className="h-5 w-5 mr-2 text-purple-600" />
              For Site Reliability Engineers
            </h4>
            <ul className="text-sm text-gray-700 space-y-1">
              <li>• <strong>Incident Response:</strong> Known failure modes and behaviors</li>
              <li>• <strong>Change Management:</strong> Confidence in configuration updates</li>
              <li>• <strong>Monitoring Setup:</strong> Understanding of expected behaviors</li>
              <li>• <strong>Disaster Recovery:</strong> Validated backup and restore processes</li>
            </ul>
          </div>
        </div>

        <div className="mt-8 p-6 bg-blue-50 rounded-lg">
          <h5 className="font-medium mb-3 text-lg text-blue-900">
            Impact on the Gateway API Ecosystem
          </h5>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h6 className="font-medium mb-2 text-blue-800">Setting Standards</h6>
              <p className="text-sm text-blue-700 leading-relaxed">
                Envoy Gateway's conformance testing contributes to the broader Gateway API ecosystem by validating
                specification completeness, identifying implementation challenges, and demonstrating best practices
                for other Gateway API implementations.
              </p>
            </div>
            <div>
              <h6 className="font-medium mb-2 text-blue-800">Building Trust</h6>
              <p className="text-sm text-blue-700 leading-relaxed">
                The comprehensive test suite builds confidence in the Gateway API specification itself,
                Envoy Gateway as a reliable implementation, and the cloud-native ecosystem's ability
                to deliver on standardized API promises.
              </p>
            </div>
          </div>
        </div>

        <div className="mt-6 p-6 bg-green-50 rounded-lg">
          <h5 className="font-medium mb-3 text-lg text-green-900">
            Conclusion: From Specification to Production Reality
          </h5>
          <p className="text-sm text-green-800 leading-relaxed">
            Envoy Gateway's conformance tests are not just technical validation—they represent a commitment to reliability,
            interoperability, and user confidence. By ensuring that Gateway API resources behave consistently and predictably,
            these tests enable users to <strong>deploy with confidence</strong>, <strong>scale reliably</strong>,
            <strong>integrate seamlessly</strong>, and <strong>maintain compliance</strong> with industry standards
            and enterprise requirements.
          </p>
        </div>
      </CardContent>
    </Card>
  );
};

export default RealWorldApplicationCard;
