import React, { Component } from 'react';
import type { ReactNode } from 'react';
import { ApiError, OfflineError } from '../services/apiClient';
import './ErrorBoundary.css';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

export default class ErrorBoundary extends Component<Props, State> {
  private retryTimeoutId: number | null = null;

  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error: Error): State {
    // Update state so the next render will show the fallback UI
    return { hasError: true, error, errorInfo: null };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // Log error details
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    
    this.setState({
      error,
      errorInfo
    });

    // Call custom error handler if provided
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }
  }

  componentWillUnmount() {
    if (this.retryTimeoutId) {
      clearTimeout(this.retryTimeoutId);
    }
  }

  handleRetry = () => {
    // Clear error state
    this.setState({ hasError: false, error: null, errorInfo: null });
    
    // Optional: Add delay before retry to prevent rapid retries
    this.retryTimeoutId = setTimeout(() => {
      // Force a re-render of children
      this.forceUpdate();
    }, 100);
  };

  handleReload = () => {
    window.location.reload();
  };

  private getErrorMessage(): string {
    const { error } = this.state;
    
    if (!error) return 'An unknown error occurred';
    
    if (error instanceof OfflineError) {
      return 'You appear to be offline. Please check your connection and try again.';
    }
    
    if (error instanceof ApiError) {
      if (error.isNetworkError) {
        return 'Network error. Please check your connection and try again.';
      }
      if (error.status === 404) {
        return 'The requested resource was not found.';
      }
      if (error.status === 500) {
        return 'Server error. Please try again in a few minutes.';
      }
      if (error.status === 429) {
        return 'Too many requests. Please wait a moment and try again.';
      }
      return error.message;
    }
    
    // Generic JavaScript errors
    if (error.name === 'ChunkLoadError') {
      return 'Failed to load application resources. Please refresh the page.';
    }
    
    return error.message || 'An unexpected error occurred';
  }

  private shouldShowRetryButton(): boolean {
    const { error } = this.state;
    
    if (!error) return false;
    
    // Show retry for network errors, API errors, and other recoverable errors
    return (
      error instanceof OfflineError ||
      error instanceof ApiError ||
      error.name === 'ChunkLoadError' ||
      error.message.toLowerCase().includes('network') ||
      error.message.toLowerCase().includes('fetch')
    );
  }

  render() {
    if (this.state.hasError) {
      // Custom fallback UI
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // Default error UI
      return (
        <div className="error-boundary">
          <div className="error-content">
            <div className="error-icon">⚠️</div>
            <h2 className="error-title">Something went wrong</h2>
            <p className="error-message">{this.getErrorMessage()}</p>
            
            <div className="error-actions">
              {this.shouldShowRetryButton() && (
                <button 
                  className="retry-button"
                  onClick={this.handleRetry}
                >
                  Try Again
                </button>
              )}
              
              <button 
                className="reload-button"
                onClick={this.handleReload}
              >
                Reload Page
              </button>
            </div>

            {/* Show error details in development */}
            {import.meta.env.DEV && this.state.error && (
              <details className="error-details">
                <summary>Error Details (Development)</summary>
                <pre className="error-stack">
                  {this.state.error.stack}
                  {this.state.errorInfo?.componentStack}
                </pre>
              </details>
            )}
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

// Higher-order component for easier usage
export function withErrorBoundary<T extends object>(
  Component: React.ComponentType<T>,
  errorBoundaryProps?: Omit<Props, 'children'>
) {
  const WrappedComponent = (props: T) => (
    <ErrorBoundary {...errorBoundaryProps}>
      <Component {...props} />
    </ErrorBoundary>
  );

  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`;
  
  return WrappedComponent;
}

// Backward compatibility fallback components
interface ErrorFallbackProps {
  error?: Error;
  retry?: () => void;
}

export const NetworkErrorFallback: React.FC<ErrorFallbackProps> = ({ retry }) => {
  return (
    <div className="error-boundary network-error">
      <div className="error-content">
        <div className="error-icon">🌐</div>
        <h3 className="error-title">Lost connection to the guild</h3>
        <p className="error-message">Check your connection to the magical network.</p>
        {retry && (
          <button onClick={retry} className="retry-button">
            🔄 Reconnect
          </button>
        )}
      </div>
    </div>
  );
};

export const AgentErrorFallback: React.FC<ErrorFallbackProps> = ({ retry }) => {
  return (
    <div className="error-boundary agent-error">
      <div className="error-content">
        <div className="error-icon">🗡️</div>
        <h3 className="error-title">Adventurer data unavailable</h3>
        <p className="error-message">Failed to load adventurer information.</p>
        {retry && (
          <button onClick={retry} className="retry-button">
            🔄 Try Again
          </button>
        )}
      </div>
    </div>
  );
};

export const QuestErrorFallback: React.FC<ErrorFallbackProps> = ({ retry }) => {
  return (
    <div className="error-boundary quest-error">
      <div className="error-content">
        <div className="error-icon">📜</div>
        <h3 className="error-title">Quest data unavailable</h3>
        <p className="error-message">Failed to load quest information.</p>
        {retry && (
          <button onClick={retry} className="retry-button">
            🔄 Try Again
          </button>
        )}
      </div>
    </div>
  );
};