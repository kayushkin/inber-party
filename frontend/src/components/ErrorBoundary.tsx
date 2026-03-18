import React from 'react';
import './ErrorBoundary.css';

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<{
    error: Error;
    retry: () => void;
  }>;
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({ errorInfo });
    
    // Log error details
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    
    // Call optional error callback
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }
  }

  retry = () => {
    this.setState({ hasError: false, error: null, errorInfo: null });
  };

  render() {
    if (this.state.hasError && this.state.error) {
      if (this.props.fallback) {
        const FallbackComponent = this.props.fallback;
        return <FallbackComponent error={this.state.error} retry={this.retry} />;
      }

      return <DefaultErrorFallback error={this.state.error} retry={this.retry} />;
    }

    return this.props.children;
  }
}

interface ErrorFallbackProps {
  error: Error;
  retry: () => void;
}

export const DefaultErrorFallback: React.FC<ErrorFallbackProps> = ({ error, retry }) => {
  return (
    <div className="error-boundary">
      <div className="error-boundary-content">
        <div className="error-icon">🛡️</div>
        <h2>Something went wrong in the guild halls!</h2>
        <p className="error-message">
          The magic that powers this interface has encountered an unexpected problem.
        </p>
        <details className="error-details">
          <summary>Technical details (for the guild's scribes)</summary>
          <pre className="error-stack">{error.toString()}</pre>
        </details>
        <div className="error-actions">
          <button onClick={retry} className="retry-button">
            🔄 Try Again
          </button>
          <button onClick={() => window.location.reload()} className="reload-button">
            🏠 Return to Tavern
          </button>
        </div>
      </div>
    </div>
  );
};

// Specific error fallbacks for different contexts
export const AgentErrorFallback: React.FC<ErrorFallbackProps> = ({ retry }) => {
  return (
    <div className="error-boundary agent-error">
      <div className="error-boundary-content">
        <div className="error-icon">😵</div>
        <h3>This adventurer seems confused...</h3>
        <p>Something went wrong while loading their information.</p>
        <button onClick={retry} className="retry-button">
          🔄 Check on them again
        </button>
      </div>
    </div>
  );
};

export const QuestErrorFallback: React.FC<ErrorFallbackProps> = ({ retry }) => {
  return (
    <div className="error-boundary quest-error">
      <div className="error-boundary-content">
        <div className="error-icon">📜</div>
        <h3>Quest board is illegible</h3>
        <p>The quest details couldn't be loaded properly.</p>
        <button onClick={retry} className="retry-button">
          🔄 Check the quest board again
        </button>
      </div>
    </div>
  );
};

export const NetworkErrorFallback: React.FC<ErrorFallbackProps> = ({ retry }) => {
  return (
    <div className="error-boundary network-error">
      <div className="error-boundary-content">
        <div className="error-icon">🌐</div>
        <h3>Lost connection to the guild</h3>
        <p>Check your connection to the magical network.</p>
        <button onClick={retry} className="retry-button">
          🔄 Reconnect
        </button>
      </div>
    </div>
  );
};

export default ErrorBoundary;