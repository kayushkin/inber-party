import React, { useState, useEffect } from 'react';
import ErrorBoundary, { NetworkErrorFallback } from './ErrorBoundary';

interface AsyncErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<{
    error: Error;
    retry: () => void;
  }>;
}

// Hook for handling async errors in functional components
export const useAsyncError = () => {
  const [asyncError, setAsyncError] = useState<Error | null>(null);

  const throwAsyncError = (error: Error) => {
    setAsyncError(error);
  };

  useEffect(() => {
    if (asyncError) {
      throw asyncError;
    }
  }, [asyncError]);

  const resetAsyncError = () => {
    setAsyncError(null);
  };

  return { throwAsyncError, resetAsyncError };
};

// Wrapper for handling async errors
export const AsyncErrorBoundary: React.FC<AsyncErrorBoundaryProps> = ({ 
  children, 
  fallback = NetworkErrorFallback 
}) => {
  return (
    <ErrorBoundary fallback={fallback}>
      {children}
    </ErrorBoundary>
  );
};

// Higher-order component to wrap any component with error boundary
export const withErrorBoundary = <P extends object>(
  Component: React.ComponentType<P>,
  fallbackComponent?: React.ComponentType<{
    error: Error;
    retry: () => void;
  }>
) => {
  const WrappedComponent = (props: P) => (
    <ErrorBoundary fallback={fallbackComponent}>
      <Component {...props} />
    </ErrorBoundary>
  );
  
  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`;
  return WrappedComponent;
};

// Async operation wrapper with automatic error handling
export const safeAsync = async <T,>(
  operation: () => Promise<T>,
  onError?: (error: Error) => void
): Promise<T | null> => {
  try {
    return await operation();
  } catch (error) {
    const errorObj = error instanceof Error ? error : new Error(String(error));
    if (onError) {
      onError(errorObj);
    } else {
      console.error('Async operation failed:', errorObj);
    }
    return null;
  }
};

// Custom hook for API calls with error handling
export const useApiCall = <T,>(
  apiCall: () => Promise<T>,
  dependencies: React.DependencyList = []
) => {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const { throwAsyncError } = useAsyncError();

  const executeCall = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await apiCall();
      setData(result);
    } catch (err) {
      const errorObj = err instanceof Error ? err : new Error(String(err));
      setError(errorObj);
      
      // For network errors, throw to be caught by ErrorBoundary
      if (errorObj.message.includes('fetch') || errorObj.message.includes('network')) {
        throwAsyncError(errorObj);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    executeCall();
  }, dependencies);

  const retry = () => {
    executeCall();
  };

  return { data, loading, error, retry };
};

export default AsyncErrorBoundary;