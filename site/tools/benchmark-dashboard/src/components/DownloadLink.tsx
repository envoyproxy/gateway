import React from 'react';
import { Download, ExternalLink } from 'lucide-react';
import { TestSuite } from '../data/types';

interface DownloadLinkProps {
  testSuite: TestSuite;
  className?: string;
  showIcon?: boolean;
}

export function DownloadLink({ testSuite, className = '', showIcon = true }: DownloadLinkProps) {
  if (!testSuite.metadata.downloadUrl) {
    return null;
  }

  return (
    <a
      href={testSuite.metadata.downloadUrl}
      target="_blank"
      rel="noopener noreferrer"
      className={`inline-flex items-center gap-2 text-blue-600 hover:text-blue-800 hover:underline transition-colors ${className}`}
      title={`Download raw benchmark data for version ${testSuite.metadata.version}`}
    >
      {showIcon && <Download size={16} />}
      <span>Download Raw Data</span>
      <ExternalLink size={14} className="opacity-60" />
    </a>
  );
}

interface DownloadButtonProps {
  testSuite: TestSuite;
  className?: string;
  variant?: 'primary' | 'secondary' | 'outline';
}

export function DownloadButton({
  testSuite,
  className = '',
  variant = 'outline'
}: DownloadButtonProps) {
  if (!testSuite.metadata.downloadUrl) {
    return null;
  }

  const variantClasses = {
    primary: 'bg-blue-600 text-white hover:bg-blue-700',
    secondary: 'bg-gray-600 text-white hover:bg-gray-700',
    outline: 'border border-gray-300 text-gray-700 hover:bg-gray-50'
  };

  return (
    <a
      href={testSuite.metadata.downloadUrl}
      target="_blank"
      rel="noopener noreferrer"
      className={`inline-flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${variantClasses[variant]} ${className}`}
      title={`Download raw benchmark data for version ${testSuite.metadata.version}`}
    >
      <Download size={16} />
      <span>Download Data</span>
    </a>
  );
}

// Usage examples:
//
// Simple link:
// <DownloadLink testSuite={testSuite} />
//
// Button style:
// <DownloadButton testSuite={testSuite} variant="primary" />
//
// Custom styling:
// <DownloadLink
//   testSuite={testSuite}
//   className="text-green-600 font-medium"
//   showIcon={false}
// />
