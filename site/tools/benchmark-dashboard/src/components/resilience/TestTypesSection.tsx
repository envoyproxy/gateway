
import React from 'react';
import { Target, Settings, Activity, Monitor } from 'lucide-react';
import TestTypeCard from './TestTypeCard';

const TestTypesSection = () => {
  const testTypes = [
    {
      icon: Settings,
      title: "Extension Server Resilience",
      description: "Validates extension server fault tolerance and configuration cache preservation",
      scenarios: [
        "XDS Translation Error Handling",
        "Configuration Cache Preservation", 
        "Control Plane Restart Resilience",
        "Extension Server Recovery"
      ],
      validationPoints: [
        "Last known good config preservation",
        "Prometheus error metrics accuracy",
        "Automatic recovery capabilities",
        "Multi-tenant isolation safety"
      ],
      bgColor: "from-purple-50 to-indigo-50",
      iconColor: "text-purple-600",
      scenariosColor: "bg-green-50 border border-green-200",
      validationColor: "bg-blue-50 border border-blue-200"
    },
    {
      icon: Activity,
      title: "Control Plane High Availability",
      description: "Tests control plane HA, leader election, and API server connectivity resilience",
      scenarios: [
        "Multi-Instance Failover Testing",
        "API Server Connectivity Loss",
        "Leader Election Transitions",
        "Network Partition Simulation"
      ],
      validationPoints: [
        "Secondary instance xDS serving",
        "Configuration reconciliation",
        "Graceful leadership handoff",
        "Service continuity maintenance"
      ],
      bgColor: "from-blue-50 to-indigo-50",
      iconColor: "text-blue-600",
      scenariosColor: "bg-orange-50 border border-orange-200",
      validationColor: "bg-purple-50 border border-purple-200"
    },
    {
      icon: Monitor,
      title: "Data Plane Continuity",
      description: "Validates traffic processing continues even when control plane is offline",
      scenarios: [
        "Complete control plane outage",
        "Extended disconnection periods",
        "HTTP traffic continuity validation"
      ],
      validationPoints: [
        "Existing routes remain functional",
        "No end-user service interruption",
        "Traffic processing continues normally"
      ],
      bgColor: "from-green-50 to-teal-50",
      iconColor: "text-green-600",
      scenariosColor: "bg-red-50 border border-red-200",
      validationColor: "bg-teal-50 border border-teal-200"
    }
  ];

  return (
    <div className="space-y-6">
      <div className="text-center mb-8">
        <h2 className="text-3xl font-bold text-gray-900 mb-4 flex items-center justify-center">
          <Target className="h-8 w-8 mr-3 text-purple-600" />
          Types of Resilience Tests
        </h2>
        <p className="text-lg text-gray-600 max-w-3xl mx-auto">
          Three comprehensive test categories that validate different aspects of fault tolerance
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {testTypes.map((testType, index) => (
          <TestTypeCard key={index} {...testType} />
        ))}
      </div>
    </div>
  );
};

export default TestTypesSection;
