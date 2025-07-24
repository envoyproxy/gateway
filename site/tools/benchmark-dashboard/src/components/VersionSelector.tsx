import React from 'react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Calendar, GitBranch, ExternalLink, Download, BarChart3 } from 'lucide-react';

interface VersionSelectorProps {
  selectedVersion: string;
  availableVersions: string[];
  onVersionChange: (version: string) => void;
  metadata?: {
    version: string;
    runId: string;
    date: string;
    environment?: string;
    description?: string;
    gitCommit?: string;
  } | null;
  onDownloadReport?: () => void;
}

const VersionSelector = ({
  selectedVersion,
  availableVersions,
  onVersionChange,
  metadata,
  onDownloadReport
}: VersionSelectorProps) => {
  const formatDate = (dateString: string) => {
    try {
      return new Date(dateString).toLocaleDateString();
    } catch {
      return dateString;
    }
  };

  const getReleaseUrl = (version: string) => {
    return `https://github.com/envoyproxy/gateway/releases/tag/v${version}`;
  };

  const getBenchmarkDownloadUrl = (version: string) => {
    return `https://github.com/envoyproxy/gateway/releases/download/v${version}/benchmark_report.zip`;
  };

  const handleViewRelease = () => {
    const releaseUrl = getReleaseUrl(selectedVersion);
    window.open(releaseUrl, '_blank', 'noopener,noreferrer');
  };

  const handleDownloadBenchmarkReport = () => {
    const downloadUrl = getBenchmarkDownloadUrl(selectedVersion);
    window.open(downloadUrl, '_blank', 'noopener,noreferrer');
  };

  return (
    <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
      <div className="flex items-center gap-2">
        <label className="text-sm font-medium text-gray-700">Version:</label>
        <Select value={selectedVersion} onValueChange={onVersionChange}>
          <SelectTrigger className="w-32">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {availableVersions.map((version) => (
              <SelectItem key={version} value={version}>
                v{version}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {metadata && (
        <div className="flex flex-wrap items-center gap-2 text-xs text-gray-600">
          <div className="flex items-center gap-1">
            <Calendar className="h-3 w-3" />
            <span>{formatDate(metadata.date)}</span>
          </div>

          {metadata.environment && (
            <Badge variant="secondary" className="text-xs">
              {metadata.environment}
            </Badge>
          )}

          {metadata.gitCommit && (
            <div className="flex items-center gap-1">
              <GitBranch className="h-3 w-3" />
              <span className="font-mono">{metadata.gitCommit.substring(0, 7)}</span>
            </div>
          )}

          {metadata.description && (
            <span className="hidden sm:inline max-w-xs truncate" title={metadata.description}>
              {metadata.description}
            </span>
          )}
        </div>
      )}

      <div className="flex items-center gap-2 ml-auto">
        <Button
          variant="outline"
          size="sm"
          onClick={handleViewRelease}
          className="flex items-center gap-1 text-xs"
        >
          <ExternalLink className="h-3 w-3" />
          View Release
        </Button>

        <Button
          variant="outline"
          size="sm"
          onClick={handleDownloadBenchmarkReport}
          className="flex items-center gap-1 text-xs"
        >
          <BarChart3 className="h-3 w-3" />
          Download Benchmark
        </Button>
      </div>
    </div>
  );
};

export default VersionSelector;
