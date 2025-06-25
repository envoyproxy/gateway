
import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { LucideIcon, Settings, Check } from 'lucide-react';

interface TestTypeCardProps {
  icon: LucideIcon;
  title: string;
  description: string;
  scenarios: string[];
  validationPoints: string[];
  bgColor: string;
  iconColor: string;
  scenariosColor: string;
  validationColor: string;
}

const TestTypeCard: React.FC<TestTypeCardProps> = ({
  icon: Icon,
  title,
  description,
  scenarios,
  validationPoints,
  iconColor,
}) => {
  return (
    <Card className="hover:shadow-md transition-shadow">
      <CardHeader className="pb-4">
        <CardTitle className="text-lg flex items-center">
          <Icon className={`h-5 w-5 mr-2 ${iconColor}`} />
          {title}
        </CardTitle>
        <CardDescription className="mt-2">
          {description}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6 pt-2">
        <div>
          <h5 className="font-medium mb-3 flex items-center">
            <Settings className="h-4 w-4 mr-2 text-blue-600" />
            Test Scenarios
          </h5>
          <ul className="text-sm text-muted-foreground space-y-2 ml-6">
            {scenarios.map((scenario, index) => (
              <li key={index}>• {scenario}</li>
            ))}
          </ul>
        </div>
        <div>
          <h5 className="font-medium mb-3 flex items-center">
            <Check className="h-4 w-4 mr-2 text-green-600" />
            Validation Points
          </h5>
          <ul className="text-sm text-muted-foreground space-y-2 ml-6">
            {validationPoints.map((point, index) => (
              <li key={index}>• {point}</li>
            ))}
          </ul>
        </div>
      </CardContent>
    </Card>
  );
};

export default TestTypeCard;
