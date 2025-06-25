
import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ExternalLink, Users, GitBranch, Code, Heart, Shield, Eye } from 'lucide-react';
import { Link } from 'react-router-dom';

const NextSteps = () => {
  return (
    <div className="space-y-6 sm:space-y-8 w-full">
      {/* Main Next Steps Card */}
      <Card className="border-gray-200">
        <CardHeader className="text-center pb-4 sm:pb-6 px-4 sm:px-6">
          <CardTitle className="text-xl sm:text-2xl font-bold text-gray-900">Ready to Get Started?</CardTitle>
          <CardDescription className="text-base sm:text-lg text-gray-600 max-w-2xl mx-auto">
            Apply these insights to your deployment and join our collaborative testing community
          </CardDescription>
        </CardHeader>
        <CardContent className="px-4 sm:px-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 sm:gap-8">
            {/* Deployment Planning */}
            <div className="bg-gradient-to-br from-blue-50 to-indigo-50 rounded-xl p-4 sm:p-6 border border-blue-100">
              <div className="flex items-center mb-3 sm:mb-4">
                <div className="p-2 bg-blue-100 rounded-lg mr-3 flex-shrink-0">
                  <Shield className="h-4 w-4 sm:h-5 sm:w-5 text-blue-700" />
                </div>
                <h4 className="font-semibold text-gray-900 text-base sm:text-lg">Planning Your Deployment</h4>
              </div>
              <div className="space-y-2 sm:space-y-3">
                <div className="flex items-start space-x-2 sm:space-x-3">
                  <div className="w-5 h-5 sm:w-6 sm:h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-xs font-medium mt-0.5 flex-shrink-0">1</div>
                  <span className="text-xs sm:text-sm text-gray-700">Estimate your expected route count and traffic patterns</span>
                </div>
                <div className="flex items-start space-x-2 sm:space-x-3">
                  <div className="w-5 h-5 sm:w-6 sm:h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-xs font-medium mt-0.5 flex-shrink-0">2</div>
                  <span className="text-xs sm:text-sm text-gray-700">Choose resource limits based on our community recommendations</span>
                </div>
                <div className="flex items-start space-x-2 sm:space-x-3">
                  <div className="w-5 h-5 sm:w-6 sm:h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-xs font-medium mt-0.5 flex-shrink-0">3</div>
                  <span className="text-xs sm:text-sm text-gray-700">Set up monitoring for memory usage and latency metrics</span>
                </div>
                <div className="flex items-start space-x-2 sm:space-x-3">
                  <div className="w-5 h-5 sm:w-6 sm:h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-xs font-medium mt-0.5 flex-shrink-0">4</div>
                  <span className="text-xs sm:text-sm text-gray-700">Plan for gradual scaling with continuous validation</span>
                </div>
              </div>
            </div>

            {/* Monitoring Setup */}
            <div className="bg-gradient-to-br from-green-50 to-emerald-50 rounded-xl p-4 sm:p-6 border border-green-100">
              <div className="flex items-center mb-3 sm:mb-4">
                <div className="p-2 bg-green-100 rounded-lg mr-3 flex-shrink-0">
                  <Eye className="h-4 w-4 sm:h-5 sm:w-5 text-green-700" />
                </div>
                <h4 className="font-semibold text-gray-900 text-base sm:text-lg">Monitoring & Observability</h4>
              </div>
              <div className="space-y-2 sm:space-y-3">
                <div className="flex items-start space-x-2 sm:space-x-3">
                  <div className="w-5 h-5 sm:w-6 sm:h-6 bg-green-600 text-white rounded-full flex items-center justify-center text-xs font-medium mt-0.5 flex-shrink-0">1</div>
                  <span className="text-xs sm:text-sm text-gray-700">Configure Prometheus metrics collection</span>
                </div>
                <div className="flex items-start space-x-2 sm:space-x-3">
                  <div className="w-5 h-5 sm:w-6 sm:h-6 bg-green-600 text-white rounded-full flex items-center justify-center text-xs font-medium mt-0.5 flex-shrink-0">2</div>
                  <span className="text-xs sm:text-sm text-gray-700">Set up alerts for P95 latency {'>'} 50ms</span>
                </div>
                <div className="flex items-start space-x-2 sm:space-x-3">
                  <div className="w-5 h-5 sm:w-6 sm:h-6 bg-green-600 text-white rounded-full flex items-center justify-center text-xs font-medium mt-0.5 flex-shrink-0">3</div>
                  <span className="text-xs sm:text-sm text-gray-700">Monitor memory usage trends and connection pools</span>
                </div>
                <div className="flex items-start space-x-2 sm:space-x-3">
                  <div className="w-5 h-5 sm:w-6 sm:h-6 bg-green-600 text-white rounded-full flex items-center justify-center text-xs font-medium mt-0.5 flex-shrink-0">4</div>
                  <span className="text-xs sm:text-sm text-gray-700">Track performance degradation patterns</span>
                </div>
              </div>

              <div className="mt-3 sm:mt-4 pt-3 sm:pt-4 border-t border-green-200">
                <Button
                  variant="outline"
                  size="sm"
                  className="w-full border-green-300 text-green-700 hover:bg-green-50 text-xs sm:text-sm"
                  onClick={() => window.open('https://gateway.envoyproxy.io/docs/tasks/observability/', '_blank')}
                >
                  <ExternalLink className="h-3 w-3 sm:h-4 sm:w-4 mr-2" />
                  Observability Documentation
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Community Collaboration Card */}
      <Card className="border-purple-200 bg-gradient-to-br from-purple-50 via-white to-indigo-50">
        <CardHeader className="text-center pb-4 sm:pb-6 px-4 sm:px-6">
          <div className="flex flex-col sm:flex-row items-center justify-center space-y-2 sm:space-y-0 sm:space-x-3 mb-3 sm:mb-4">
            <div className="p-3 bg-purple-100 rounded-xl">
              <Heart className="h-5 w-5 sm:h-6 sm:w-6 text-purple-700" />
            </div>
            <CardTitle className="text-xl sm:text-2xl font-bold bg-gradient-to-r from-purple-700 to-indigo-700 bg-clip-text text-transparent text-center sm:text-left">
              Built by the Community, For the Community
            </CardTitle>
          </div>
          <CardDescription className="text-sm sm:text-lg text-gray-600 max-w-3xl mx-auto leading-relaxed">
            These benchmarks and insights are the result of collective effort from contributors across the industry.
            Join us in making Envoy Gateway better for everyone.
          </CardDescription>
        </CardHeader>
        <CardContent className="px-4 sm:px-6">
          {/* Community Values */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6 mb-6 sm:mb-8">
            <div className="text-center p-3 sm:p-4 bg-white rounded-lg border border-gray-100 shadow-sm">
              <div className="p-2 sm:p-3 bg-blue-100 rounded-full w-fit mx-auto mb-2 sm:mb-3">
                <Eye className="h-5 w-5 sm:h-6 sm:w-6 text-blue-700" />
              </div>
              <h4 className="font-semibold text-gray-900 mb-1 sm:mb-2 text-sm sm:text-base">Transparent</h4>
              <p className="text-xs sm:text-sm text-gray-600">All testing methodologies, results, and code are open source and publicly available</p>
            </div>

            <div className="text-center p-3 sm:p-4 bg-white rounded-lg border border-gray-100 shadow-sm">
              <div className="p-2 sm:p-3 bg-green-100 rounded-full w-fit mx-auto mb-2 sm:mb-3">
                <Users className="h-5 w-5 sm:h-6 sm:w-6 text-green-700" />
              </div>
              <h4 className="font-semibold text-gray-900 mb-1 sm:mb-2 text-sm sm:text-base">Collaborative</h4>
              <p className="text-xs sm:text-sm text-gray-600">Built together by maintainers, contributors, and users from around the world</p>
            </div>

            <div className="text-center p-3 sm:p-4 bg-white rounded-lg border border-gray-100 shadow-sm sm:col-span-2 lg:col-span-1">
              <div className="p-2 sm:p-3 bg-purple-100 rounded-full w-fit mx-auto mb-2 sm:mb-3">
                <Shield className="h-5 w-5 sm:h-6 sm:w-6 text-purple-700" />
              </div>
              <h4 className="font-semibold text-gray-900 mb-1 sm:mb-2 text-sm sm:text-base">Vendor-Neutral</h4>
              <p className="text-xs sm:text-sm text-gray-600">Independent testing with no commercial influence or bias</p>
            </div>
          </div>

          {/* Call to Action Buttons */}
          <div className="space-y-4">
            <div className="text-center mb-4 sm:mb-6">
              <h4 className="text-base sm:text-lg font-semibold text-gray-900 mb-1 sm:mb-2">Ready to Contribute?</h4>
              <p className="text-sm sm:text-base text-gray-600">There are many ways to get involved and help improve Envoy Gateway testing</p>
            </div>

            <div className="flex flex-col space-y-3 sm:space-y-0 sm:flex-row sm:gap-4 justify-center max-w-4xl mx-auto">
              <a
                href="https://github.com/envoyproxy/gateway"
                target="_blank"
                rel="noopener noreferrer"
                className="bg-gray-900 hover:bg-gray-800 text-white px-4 sm:px-6 py-3 rounded-lg text-sm sm:text-base font-medium transition-all inline-flex items-center justify-center shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 duration-200 flex-1"
              >
                <GitBranch className="h-4 w-4 sm:h-5 sm:w-5 mr-2" />
                Explore the Code
              </a>

              <Link
                to="/test-instructions"
                className="bg-blue-600 hover:bg-blue-700 text-white px-4 sm:px-6 py-3 rounded-lg text-sm sm:text-base font-medium transition-all inline-flex items-center justify-center shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 duration-200 flex-1"
              >
                <Code className="h-4 w-4 sm:h-5 sm:w-5 mr-2" />
                Run Tests Yourself
              </Link>

              <a
                href="https://gateway.envoyproxy.io/contributions/"
                target="_blank"
                rel="noopener noreferrer"
                className="bg-white hover:bg-gray-50 text-gray-900 border-2 border-gray-300 hover:border-gray-400 px-4 sm:px-6 py-3 rounded-lg text-sm sm:text-base font-medium transition-all inline-flex items-center justify-center shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 duration-200 flex-1"
              >
                <Users className="h-4 w-4 sm:h-5 sm:w-5 mr-2" />
                Join the Community
              </a>
            </div>

            {/* Additional Community Info */}
            <div className="mt-6 sm:mt-8 pt-4 sm:pt-6 border-t border-purple-200">
              <div className="text-center text-xs sm:text-sm text-gray-600 space-y-1 sm:space-y-2">
                <p>
                  <strong>No corporate agenda.</strong> These tests are maintained by the community,
                  ensuring unbiased and reliable performance insights.
                </p>
                <p>
                  Every contribution, from code improvements to documentation, helps make Envoy Gateway better for everyone.
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default NextSteps;
