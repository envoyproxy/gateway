import React from 'react';
import { Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg' | 'xl';
  variant?: 'default' | 'muted' | 'primary';
  className?: string;
  text?: string;
}

const sizeClasses = {
  sm: 'h-4 w-4',
  md: 'h-6 w-6',
  lg: 'h-8 w-8',
  xl: 'h-12 w-12',
};

const variantClasses = {
  default: 'text-gray-500',
  muted: 'text-gray-400',
  primary: 'text-purple-600',
};

export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'md',
  variant = 'default',
  className,
  text,
}) => {
  return (
    <div className={cn('flex items-center justify-center', className)}>
      <div className="flex flex-col items-center gap-2">
        <Loader2
          className={cn(
            'animate-spin',
            sizeClasses[size],
            variantClasses[variant]
          )}
        />
        {text && (
          <span className={cn('text-sm', variantClasses[variant])}>
            {text}
          </span>
        )}
      </div>
    </div>
  );
};

// Full-page loading component
export const PageLoading: React.FC<{ text?: string }> = ({ text = 'Loading...' }) => (
  <div className="min-h-[400px] flex items-center justify-center">
    <LoadingSpinner size="lg" variant="primary" text={text} />
  </div>
);

// Card loading component
export const CardLoading: React.FC<{ text?: string }> = ({ text = 'Loading data...' }) => (
  <div className="p-8 flex items-center justify-center">
    <LoadingSpinner size="md" variant="muted" text={text} />
  </div>
);

// Inline loading component
export const InlineLoading: React.FC<{ text?: string }> = ({ text }) => (
  <span className="inline-flex items-center gap-2">
    <Loader2 className="h-4 w-4 animate-spin text-gray-500" />
    {text && <span className="text-sm text-gray-500">{text}</span>}
  </span>
);
