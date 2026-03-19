import React from 'react';
import ErrorBoundary, { Props as ErrorBoundaryProps } from '../components/ErrorBoundary';

// Higher-order component for easier usage
export function withErrorBoundary<T extends object>(
  Component: React.ComponentType<T>,
  errorBoundaryProps?: Omit<ErrorBoundaryProps, 'children'>
) {
  const WrappedComponent = (props: T) => (
    React.createElement(ErrorBoundary, errorBoundaryProps, React.createElement(Component, props))
  );
  
  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`;
  return WrappedComponent;
}