import React from 'react';
import { BarChart3, Database, AlertCircle, RefreshCw } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

interface DataPlaceholderProps {
  type?: 'empty' | 'error' | 'loading' | 'no-data';
  title?: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
  className?: string;
}

const placeholderConfig = {
  empty: {
    icon: Database,
    title: 'No Data Available',
    description: 'There is no data to display at the moment.',
  },
  error: {
    icon: AlertCircle,
    title: 'Failed to Load Data',
    description: 'An error occurred while loading the data. Please try again.',
  },
  loading: {
    icon: RefreshCw,
    title: 'Loading Data',
    description: 'Please wait while we fetch the latest information.',
  },
  'no-data': {
    icon: BarChart3,
    title: 'No Chart Data',
    description: 'No data available to display in the chart.',
  },
};

export const DataPlaceholder: React.FC<DataPlaceholderProps> = ({
  type = 'empty',
  title,
  description,
  action,
  className,
}) => {
  const config = placeholderConfig[type];
  const Icon = config.icon;

  return (
    <Card className={`border-dashed ${className}`}>
      <CardContent className="flex flex-col items-center justify-center py-12 px-6 text-center">
        <div className="rounded-full bg-gray-100 p-3 mb-4">
          <Icon
            className={`h-8 w-8 text-gray-400 ${type === 'loading' ? 'animate-spin' : ''}`}
          />
        </div>
        <CardTitle className="text-lg font-medium text-gray-900 mb-2">
          {title || config.title}
        </CardTitle>
        <p className="text-sm text-gray-500 mb-6 max-w-sm">
          {description || config.description}
        </p>
        {action && (
          <Button
            onClick={action.onClick}
            variant="outline"
            size="sm"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            {action.label}
          </Button>
        )}
      </CardContent>
    </Card>
  );
};

// Specific placeholder components for common use cases
export const ChartPlaceholder: React.FC<{ onRetry?: () => void }> = ({ onRetry }) => (
  <DataPlaceholder
    type="no-data"
    title="No Chart Data"
    description="Unable to generate chart with the current data set."
    action={onRetry ? { label: 'Retry', onClick: onRetry } : undefined}
  />
);

export const ErrorPlaceholder: React.FC<{ onRetry?: () => void }> = ({ onRetry }) => (
  <DataPlaceholder
    type="error"
    action={onRetry ? { label: 'Try Again', onClick: onRetry } : undefined}
  />
);

export const EmptyPlaceholder: React.FC<{
  title?: string;
  description?: string;
  onAction?: () => void;
  actionLabel?: string;
}> = ({ title, description, onAction, actionLabel }) => (
  <DataPlaceholder
    type="empty"
    title={title}
    description={description}
    action={onAction && actionLabel ? { label: actionLabel, onClick: onAction } : undefined}
  />
);
