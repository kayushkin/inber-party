import { useState, useCallback, useRef } from 'react';

interface ErrorState {
  error: Error | null;
  hasError: boolean;
  retryCount: number;
}

interface UseErrorHandlerOptions {
  maxRetries?: number;
  onError?: (error: Error, retryCount: number) => void;
  onRetry?: (retryCount: number) => void;
  resetOnSuccess?: boolean;
}

export const useErrorHandler = (options: UseErrorHandlerOptions = {}) => {
  const {
    maxRetries = 3,
    onError,
    onRetry,
    resetOnSuccess = true
  } = options;

  const [errorState, setErrorState] = useState<ErrorState>({
    error: null,
    hasError: false,
    retryCount: 0
  });

  const errorTimeoutRef = useRef<number | null>(null);

  const setError = useCallback((error: Error | string | null) => {
    const errorObj = error instanceof Error ? error : error ? new Error(error) : null;
    
    setErrorState(prev => {
      const newState = {
        error: errorObj,
        hasError: !!errorObj,
        retryCount: errorObj ? prev.retryCount : 0
      };

      if (errorObj && onError) {
        onError(errorObj, newState.retryCount);
      }

      return newState;
    });
  }, [onError]);

  const clearError = useCallback(() => {
    if (errorTimeoutRef.current) {
      clearTimeout(errorTimeoutRef.current);
      errorTimeoutRef.current = null;
    }
    setErrorState({
      error: null,
      hasError: false,
      retryCount: 0
    });
  }, []);

  const retry = useCallback(() => {
    if (errorState.retryCount >= maxRetries) {
      return false;
    }

    setErrorState(prev => {
      const newRetryCount = prev.retryCount + 1;
      if (onRetry) {
        onRetry(newRetryCount);
      }
      return {
        ...prev,
        error: null,
        hasError: false,
        retryCount: newRetryCount
      };
    });

    return true;
  }, [errorState.retryCount, maxRetries, onRetry]);

  // Auto-clear error after a timeout
  const setTemporaryError = useCallback((error: Error | string, timeoutMs = 5000) => {
    setError(error);
    
    if (errorTimeoutRef.current) {
      clearTimeout(errorTimeoutRef.current);
    }
    
    errorTimeoutRef.current = setTimeout(() => {
      clearError();
    }, timeoutMs);
  }, [setError, clearError]);

  // Wrapper for async operations with automatic error handling
  const withErrorHandling = useCallback(async <T>(
    operation: () => Promise<T>,
    errorMessage?: string
  ): Promise<T | null> => {
    try {
      if (resetOnSuccess) {
        clearError();
      }
      
      const result = await operation();
      
      // Reset retry count on successful operation
      if (resetOnSuccess && errorState.retryCount > 0) {
        setErrorState(prev => ({ ...prev, retryCount: 0 }));
      }
      
      return result;
    } catch (err) {
      const error = err instanceof Error ? err : new Error(errorMessage || String(err));
      setError(error);
      return null;
    }
  }, [resetOnSuccess, clearError, errorState.retryCount, setError]);

  // Check if we can retry
  const canRetry = errorState.hasError && errorState.retryCount < maxRetries;

  return {
    error: errorState.error,
    hasError: errorState.hasError,
    retryCount: errorState.retryCount,
    canRetry,
    setError,
    clearError,
    retry,
    setTemporaryError,
    withErrorHandling,
    maxRetries
  };
};

// Hook for handling network request errors specifically
export const useNetworkErrorHandler = () => {
  const errorHandler = useErrorHandler({
    maxRetries: 3,
    onError: (error, retryCount) => {
      console.warn(`Network error (attempt ${retryCount + 1}):`, error.message);
    }
  });

  const handleNetworkRequest = useCallback(async <T>(
    request: () => Promise<Response>,
    parseResponse?: (response: Response) => Promise<T>
  ): Promise<T | null> => {
    return errorHandler.withErrorHandling(async () => {
      const response = await request();
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      if (parseResponse) {
        return await parseResponse(response);
      }

      return await response.json() as T;
    });
  }, [errorHandler]);

  return {
    ...errorHandler,
    handleNetworkRequest
  };
};

export default useErrorHandler;