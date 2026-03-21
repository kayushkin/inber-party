import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import ErrorBoundary from './components/ErrorBoundary'
import './index.css'
import App from './App.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary onError={(error, errorInfo) => {
      // Root-level error logging - could integrate with analytics/error tracking
      console.error('Root-level error caught:', error, errorInfo);
      // In production, you might want to send this to an error tracking service
    }}>
      <App />
    </ErrorBoundary>
  </StrictMode>,
)
