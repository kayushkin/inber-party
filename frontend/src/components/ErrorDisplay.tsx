import React from 'react';
import './ErrorDisplay.css';

interface ErrorDisplayProps {
  error: Error | string | null;
  title?: string;
  showRetry?: boolean;
  onRetry?: () => void;
  className?: string;
  variant?: 'inline' | 'card' | 'banner';
}

const ErrorDisplay: React.FC<ErrorDisplayProps> = ({
  error,
  title = "Something went wrong",
  showRetry = true,
  onRetry,
  className = "",
  variant = 'card'
}) => {
  if (!error) return null;

  const errorMessage = error instanceof Error ? error.message : error;
  
  // Determine appropriate icon and user-friendly message based on error type
  const getErrorDetails = (message: string) => {
    if (message.includes('fetch') || message.includes('network') || message.includes('Failed to fetch')) {
      return {
        icon: '🌐',
        userMessage: 'Unable to connect to the guild network',
        suggestion: 'Check your connection and try again'
      };
    }
    if (message.includes('404')) {
      return {
        icon: '🗺️',
        userMessage: 'The requested page or resource was not found',
        suggestion: 'The adventurer or quest you are looking for might not exist'
      };
    }
    if (message.includes('401') || message.includes('403')) {
      return {
        icon: '🚪',
        userMessage: 'You do not have permission to access this',
        suggestion: 'Check with the guild master for proper credentials'
      };
    }
    if (message.includes('500')) {
      return {
        icon: '⚡',
        userMessage: 'The guild servers are having problems',
        suggestion: 'Please try again in a few moments'
      };
    }
    if (message.includes('timeout')) {
      return {
        icon: '⏱️',
        userMessage: 'The request took too long to complete',
        suggestion: 'The guild network may be slow, try again'
      };
    }
    return {
      icon: '⚠️',
      userMessage: 'An unexpected error occurred',
      suggestion: 'Try refreshing the page or contact support if this persists'
    };
  };

  const { icon, userMessage, suggestion } = getErrorDetails(errorMessage);

  return (
    <div className={`error-display error-display--${variant} ${className}`}>
      <div className="error-display__content">
        <div className="error-display__icon">{icon}</div>
        <div className="error-display__text">
          <h3 className="error-display__title">{title}</h3>
          <p className="error-display__message">{userMessage}</p>
          <p className="error-display__suggestion">{suggestion}</p>
          
          {(import.meta.env?.DEV) && (
            <details className="error-display__details">
              <summary>Technical details</summary>
              <code className="error-display__code">{errorMessage}</code>
            </details>
          )}
        </div>
        
        {showRetry && onRetry && (
          <button className="error-display__retry" onClick={onRetry}>
            🔄 Try Again
          </button>
        )}
      </div>
    </div>
  );
};

// Specific error display variants for common use cases
export const InlineError: React.FC<Omit<ErrorDisplayProps, 'variant'>> = (props) => (
  <ErrorDisplay {...props} variant="inline" />
);

export const BannerError: React.FC<Omit<ErrorDisplayProps, 'variant'>> = (props) => (
  <ErrorDisplay {...props} variant="banner" />
);

export const CardError: React.FC<Omit<ErrorDisplayProps, 'variant'>> = (props) => (
  <ErrorDisplay {...props} variant="card" />
);

export default ErrorDisplay;