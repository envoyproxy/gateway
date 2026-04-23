import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { RefreshCw, Download, AlertCircle, CheckCircle, Clock } from 'lucide-react';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';

interface BenchmarkData {
  timestamp: string;
  version: string;
  report: {
    title: string;
    sections: Array<{
      title: string;
      content: string;
      level: number;
    }>;
    metrics: Record<string, number>;
    summary?: string;
  };
  rawMarkdown?: string;
}

interface BenchmarkReportProps {
  className?: string;
}

export const BenchmarkReport: React.FC<BenchmarkReportProps> = ({ className }) => {
  const [benchmarkData, setBenchmarkData] = useState<BenchmarkData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);

  // Fetch the latest benchmark report
  const fetchLatestReport = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('/api/benchmark-report/latest');

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to fetch benchmark report');
      }

      const data = await response.json();
      setBenchmarkData(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unknown error occurred');
    } finally {
      setLoading(false);
    }
  };

  // Trigger a fresh fetch of the benchmark report from GitHub
  const refreshReport = async () => {
    setRefreshing(true);
    setError(null);

    try {
      const response = await fetch('/api/benchmark-report', {
        method: 'POST'
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to refresh benchmark report');
      }

      const result = await response.json();

      // After successful refresh, fetch the latest data
      await fetchLatestReport();

    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unknown error occurred');
    } finally {
      setRefreshing(false);
    }
  };

  // Load data on component mount
  useEffect(() => {
    fetchLatestReport();
  }, []);

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const renderMetrics = (metrics: Record<string, number>) => {
    return Object.entries(metrics).map(([key, value]) => (
      <div key={key} className="flex justify-between items-center py-2">
        <span className="text-sm font-medium capitalize">
          {key.replace(/_/g, ' ')}
        </span>
        <Badge variant="outline">{value}</Badge>
      </div>
    ));
  };

  const renderSections = (sections: Array<{ title: string; content: string; level: number }>) => {
    return sections.map((section, index) => (
      <div key={index} className="mb-6">
        <h3 className="text-lg font-semibold mb-3">{section.title}</h3>
        <div className="prose prose-sm max-w-none">
          <pre className="whitespace-pre-wrap text-sm bg-gray-50 p-4 rounded-lg">
            {section.content}
          </pre>
        </div>
      </div>
    ));
  };

  return (
    <div className={`space-y-6 ${className}`}>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <span>Envoy Gateway Benchmark Report</span>
                {benchmarkData && (
                  <Badge variant="secondary">
                    {benchmarkData.version}
                  </Badge>
                )}
              </CardTitle>
              <CardDescription>
                Latest performance benchmarks from Envoy Gateway releases
              </CardDescription>
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={fetchLatestReport}
                disabled={loading}
              >
                <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                Reload
              </Button>
              <Button
                variant="default"
                size="sm"
                onClick={refreshReport}
                disabled={refreshing}
              >
                <Download className={`h-4 w-4 ${refreshing ? 'animate-spin' : ''}`} />
                Fetch Latest
              </Button>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          {error && (
            <Alert className="mb-4">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {loading && (
            <div className="flex items-center justify-center py-8">
              <RefreshCw className="h-6 w-6 animate-spin mr-2" />
              <span>Loading benchmark report...</span>
            </div>
          )}

          {refreshing && (
            <Alert className="mb-4">
              <Clock className="h-4 w-4" />
              <AlertTitle>Fetching Latest Report</AlertTitle>
              <AlertDescription>
                Downloading and processing the latest benchmark report from GitHub...
              </AlertDescription>
            </Alert>
          )}

          {benchmarkData && !loading && (
            <div className="space-y-6">
              {/* Report Header */}
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-bold">{benchmarkData.report.title}</h2>
                  <p className="text-sm text-muted-foreground mt-1">
                    Last updated: {formatTimestamp(benchmarkData.timestamp)}
                  </p>
                </div>
                <Alert className="w-auto">
                  <CheckCircle className="h-4 w-4" />
                  <AlertDescription>Report loaded successfully</AlertDescription>
                </Alert>
              </div>

              <Separator />

              {/* Metrics Overview */}
              {benchmarkData.report.metrics && Object.keys(benchmarkData.report.metrics).length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">Key Metrics</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {renderMetrics(benchmarkData.report.metrics)}
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Report Sections */}
              {benchmarkData.report.sections && benchmarkData.report.sections.length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">Detailed Report</CardTitle>
                  </CardHeader>
                  <CardContent>
                    {renderSections(benchmarkData.report.sections)}
                  </CardContent>
                </Card>
              )}

              {/* Raw Markdown */}
              {benchmarkData.rawMarkdown && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">Raw Report</CardTitle>
                    <CardDescription>
                      Original markdown content from the benchmark report
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <pre className="whitespace-pre-wrap text-xs bg-gray-50 p-4 rounded-lg max-h-96 overflow-y-auto">
                      {benchmarkData.rawMarkdown}
                    </pre>
                  </CardContent>
                </Card>
              )}
            </div>
          )}

          {!benchmarkData && !loading && !error && (
            <div className="text-center py-8">
              <p className="text-muted-foreground mb-4">
                No benchmark report available yet.
              </p>
              <Button onClick={refreshReport} disabled={refreshing}>
                <Download className="h-4 w-4 mr-2" />
                Fetch First Report
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
