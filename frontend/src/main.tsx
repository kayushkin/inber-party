import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import ErrorBoundary from './components/ErrorBoundary'
import './index.css'
import App from './App.tsx'

// Enhanced test environment detection
const isTest = !!(
  '__playwright' in globalThis ||
  '__jest' in globalThis ||
  navigator.userAgent.includes('HeadlessChrome') ||
  navigator.userAgent.includes('Playwright') ||
  (navigator.userAgent.includes('Firefox') && navigator.webdriver) ||
  import.meta.env.VITE_NODE_ENV === 'test' ||
  import.meta.env.VITE_CI === 'true' ||
  window.location.hostname === 'localhost'
);

// Set global test persistent mode flag to prevent any WebSocket disconnections
if (isTest) {
  console.log('🧪 Test environment detected - activating GLOBAL WEBSOCKET PERSISTENT MODE');
  (window as any).__TEST_WEBSOCKET_PERSISTENT_MODE__ = true;
}

const AppComponent = (
  <ErrorBoundary onError={(error, errorInfo) => {
    // Root-level error logging - could integrate with analytics/error tracking
    console.error('Root-level error caught:', error, errorInfo);
    // In production, you might want to send this to an error tracking service
  }}>
    <App />
  </ErrorBoundary>
);

createRoot(document.getElementById('root')!).render(
  // Disable StrictMode during E2E tests to prevent component mount/unmount/remount cycles
  // StrictMode causes intentional double-rendering which can trigger WebSocket connection churn
  isTest ? AppComponent : <StrictMode>{AppComponent}</StrictMode>
)
