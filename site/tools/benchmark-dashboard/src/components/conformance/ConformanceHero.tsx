import React from 'react';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ArrowLeft, CheckCircle } from 'lucide-react';

const ConformanceHero = () => {
  return (
    <div className="bg-gradient-to-r from-purple-600 to-indigo-600 text-white py-8 sm:py-12">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col sm:flex-row sm:items-center">
          <div className="bg-white/10 p-2 sm:p-3 rounded-full mr-0 sm:mr-4 mb-4 sm:mb-0 self-start">
            <CheckCircle className="h-6 w-6 sm:h-8 sm:w-8 text-white" />
          </div>
          <div>
            <h1 className="text-2xl sm:text-4xl font-bold mb-2">Envoy Gateway Conformance Tests</h1>
            <p className="text-base sm:text-xl text-purple-100 max-w-3xl">
              Standardized validation testing that ensures Envoy Gateway properly implements the Kubernetes Gateway API specification
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ConformanceHero;
