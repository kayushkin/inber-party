import React from 'react';

// Development-only component for testing error boundaries
const ErrorTestButtons: React.FC = () => {
  if (import.meta.env.PROD) {
    return null; // Don't show in production
  }

  const throwJSError = () => {
    throw new Error('Test JavaScript Error - Error boundary should catch this!');
  };

  const throwAsyncError = () => {
    setTimeout(() => {
      throw new Error('Test Async Error - This should appear in console');
    }, 100);
  };

  const throwChunkError = () => {
    const error = new Error('Loading chunk 42 failed');
    error.name = 'ChunkLoadError';
    throw error;
  };

  const simulateNetworkError = () => {
    const error = new Error('Failed to fetch data from server');
    error.message = 'Network error occurred';
    throw error;
  };

  return (
    <div 
      style={{ 
        position: 'fixed', 
        bottom: '20px', 
        right: '20px', 
        zIndex: 9999,
        backgroundColor: 'rgba(255, 0, 0, 0.1)',
        border: '2px dashed red',
        padding: '10px',
        borderRadius: '8px',
        fontSize: '12px'
      }}
    >
      <div style={{ marginBottom: '8px', fontWeight: 'bold' }}>
        🚨 Dev Error Testing
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
        <button onClick={throwJSError} style={{ fontSize: '11px', padding: '2px 6px' }}>
          Throw JS Error
        </button>
        <button onClick={throwAsyncError} style={{ fontSize: '11px', padding: '2px 6px' }}>
          Throw Async Error
        </button>
        <button onClick={throwChunkError} style={{ fontSize: '11px', padding: '2px 6px' }}>
          Throw Chunk Error
        </button>
        <button onClick={simulateNetworkError} style={{ fontSize: '11px', padding: '2px 6px' }}>
          Throw Network Error
        </button>
      </div>
    </div>
  );
};

export default ErrorTestButtons;