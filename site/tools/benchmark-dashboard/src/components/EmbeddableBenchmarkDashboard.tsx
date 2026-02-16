import React, { useState, useEffect } from 'react';
import VersionSelector from './VersionSelector';
import SummaryCards from './SummaryCards';
import OverviewTab from './OverviewTab';
import LatencyTab from './LatencyTab';
import ResourcesTab from './ResourcesTab';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useVersionData } from '@/hooks/useVersionData';

export interface EmbeddableDashboardConfig {
  apiBase?: string;
  initialVersion?: string;
  theme?: 'light' | 'dark';
  containerClassName?: string;
  features?: {
    header?: boolean;
    versionSelector?: boolean;
    summaryCards?: boolean;
    tabs?: string[]; // ['overview', 'latency', 'resources']
  };
}

export const EmbeddableBenchmarkDashboard: React.FC<EmbeddableDashboardConfig> = ({
  apiBase = 'https://envoy-gateway-benchmark-report.netlify.app/api',
  initialVersion,
  theme = 'light',
  containerClassName = '',
  features = {
    header: false, // Let Hugo handle the header
    versionSelector: true,
    summaryCards: true,
    tabs: ['overview', 'latency', 'resources']
  }
}) => {
  // Use the existing useVersionData hook
  const versionData = useVersionData();

  // Apply theme class to container
  useEffect(() => {
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [theme]);

  // Add CSS to ensure proper z-index for dropdowns
  useEffect(() => {
    const style = document.createElement('style');
    style.textContent = `
      .benchmark-dashboard [data-radix-popper-content-wrapper] {
        z-index: 9999 !important;
      }
      .benchmark-dashboard .relative.z-50 {
        z-index: 9999 !important;
      }
    `;
    document.head.appendChild(style);

    return () => {
      document.head.removeChild(style);
    };
  }, []);

  return (
    <div className={`benchmark-dashboard ${theme} ${containerClassName}`} data-theme={theme}>
      {/* Conditional header - only show if Hugo requests it */}
      {features.header && (
        <div className="mb-8">
          <h2 className="text-3xl font-bold">Performance Benchmark Report Explorer</h2>
          <p className="text-xl text-gray-600 dark:text-gray-300">Detailed performance analysis</p>
        </div>
      )}

      {/* Version selector */}
      {features.versionSelector && (
        <div className="mb-6">
          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-4 w-full">
            <VersionSelector
              selectedVersion={versionData.selectedVersion}
              availableVersions={versionData.availableVersions}
              onVersionChange={versionData.setSelectedVersion}
              metadata={versionData.metadata}
            />
          </div>
        </div>
      )}

      {/* Summary cards */}
      {features.summaryCards && versionData.performanceSummary && (
        <div className="mb-8">
          <SummaryCards
            performanceSummary={versionData.performanceSummary}
            benchmarkResults={versionData.benchmarkResults}
          />
        </div>
      )}

      {/* Tabs */}
      {features.tabs && features.tabs.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-lg overflow-hidden">
          <Tabs defaultValue={features.tabs[0]} className="w-full">
            <TabsList className={`grid w-full ${features.tabs.length === 1 ? 'grid-cols-1' : features.tabs.length === 2 ? 'grid-cols-2' : 'grid-cols-3'} bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 h-auto p-0 rounded-none`}>
              {features.tabs?.includes('overview') && (
                <TabsTrigger
                  value="overview"
                  className="data-[state=active]:bg-gradient-to-r data-[state=active]:from-purple-600 data-[state=active]:to-indigo-600 data-[state=active]:text-white data-[state=active]:shadow-lg data-[state=active]:border-b-2 data-[state=active]:border-purple-600 hover:bg-gray-50 dark:hover:bg-gray-700 text-sm sm:text-base py-4 px-6 rounded-t-lg border-b-2 border-transparent transition-all duration-200 font-medium"
                >
                  Overview
                </TabsTrigger>
              )}
              {features.tabs?.includes('latency') && (
                <TabsTrigger
                  value="latency"
                  className="data-[state=active]:bg-gradient-to-r data-[state=active]:from-purple-600 data-[state=active]:to-indigo-600 data-[state=active]:text-white data-[state=active]:shadow-lg data-[state=active]:border-b-2 data-[state=active]:border-purple-600 hover:bg-gray-50 dark:hover:bg-gray-700 text-sm sm:text-base py-4 px-6 rounded-t-lg border-b-2 border-transparent transition-all duration-200 font-medium"
                >
                  Request RTT Analysis
                </TabsTrigger>
              )}
              {features.tabs?.includes('resources') && (
                <TabsTrigger
                  value="resources"
                  className="data-[state=active]:bg-gradient-to-r data-[state=active]:from-purple-600 data-[state=active]:to-indigo-600 data-[state=active]:text-white data-[state=active]:shadow-lg data-[state=active]:border-b-2 data-[state=active]:border-purple-600 hover:bg-gray-50 dark:hover:bg-gray-700 text-sm sm:text-base py-4 px-6 rounded-t-lg border-b-2 border-transparent transition-all duration-200 font-medium"
                >
                  Resource Usage
                </TabsTrigger>
              )}
            </TabsList>

            <div className="p-6">
              {features.tabs?.includes('overview') && (
                <TabsContent value="overview">
                  <OverviewTab
                    performanceMatrix={versionData.performanceMatrix}
                    benchmarkResults={versionData.benchmarkResults}
                    testConfiguration={versionData.testConfiguration}
                    performanceSummary={versionData.performanceSummary}
                    latencyPercentileComparison={versionData.latencyPercentileComparison}
                  />
                </TabsContent>
              )}

              {features.tabs?.includes('latency') && (
                <TabsContent value="latency">
                  <LatencyTab
                    latencyPercentileComparison={versionData.latencyPercentileComparison}
                    benchmarkResults={versionData.benchmarkResults}
                  />
                </TabsContent>
              )}

              {features.tabs?.includes('resources') && (
                <TabsContent value="resources">
                  <ResourcesTab
                    resourceTrends={versionData.resourceTrends}
                    benchmarkResults={versionData.benchmarkResults}
                  />
                </TabsContent>
              )}
            </div>
          </Tabs>
        </div>
      )}
    </div>
  );
};
