import React from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
  errorInfo?: React.ErrorInfo;
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<ErrorFallbackProps>;
}

interface ErrorFallbackProps {
  error?: Error;
  resetError: () => void;
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    this.setState({
      error,
      errorInfo,
    });
  }

  resetError = () => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined });
  };

  render() {
    if (this.state.hasError) {
      const FallbackComponent = this.props.fallback || DefaultErrorFallback;
      return <FallbackComponent error={this.state.error} resetError={this.resetError} />;
    }

    return this.props.children;
  }
}

const DefaultErrorFallback: React.FC<ErrorFallbackProps> = ({ error, resetError }) => (
  <Card className="border-red-200 bg-red-50">
    <CardHeader>
      <CardTitle className="flex items-center gap-2 text-red-800">
        <AlertTriangle className="h-5 w-5" />
        Something went wrong
      </CardTitle>
    </CardHeader>
    <CardContent>
      <div className="space-y-4">
        <p className="text-sm text-red-600">
          An error occurred while rendering this component. Please try refreshing or contact support if the problem persists.
        </p>
        {error && process.env.NODE_ENV === 'development' && (
          <details className="text-xs text-red-500 bg-red-100 p-2 rounded">
            <summary className="cursor-pointer font-medium">Error Details</summary>
            <pre className="mt-2 whitespace-pre-wrap">{error.message}</pre>
            {error.stack && (
              <pre className="mt-2 text-xs whitespace-pre-wrap overflow-x-auto">
                {error.stack}
              </pre>
            )}
          </details>
        )}
        <Button
          onClick={resetError}
          variant="outline"
          size="sm"
          className="border-red-300 text-red-700 hover:bg-red-100"
        >
          <RefreshCw className="h-4 w-4 mr-2" />
          Try Again
        </Button>
      </div>
    </CardContent>
  </Card>
);

// Functional error boundary hook for specific components
export const withErrorBoundary = <P extends object>(
  Component: React.ComponentType<P>,
  fallback?: React.ComponentType<ErrorFallbackProps>
) => {
  const WrappedComponent: React.FC<P> = (props) => (
    <ErrorBoundary fallback={fallback}>
      <Component {...props} />
    </ErrorBoundary>
  );

  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`;
  return WrappedComponent;
};
