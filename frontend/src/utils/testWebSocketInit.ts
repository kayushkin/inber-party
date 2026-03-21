/**
 * Test Environment WebSocket Initialization
 * 
 * This module initializes WebSocket connections once at page load in test environments,
 * completely independent of React component lifecycle to eliminate connection churn.
 */

import { wsManager } from './websocket';

declare global {
  interface Window {
    __TEST_WEBSOCKET_INITIALIZED__: boolean;
    __TEST_WEBSOCKET_PERSISTENT_MODE__: boolean;
    __INBER_PARTY_TEST_INITIALIZED__: boolean;
    __WEBSOCKET_PERMANENT_LOCK__: boolean;
  }
}

/**
 * Initialize WebSocket connections for test environment
 * This runs once when the page loads and maintains connections throughout the test session
 */
export function initTestWebSockets() {
  // Enhanced test environment detection
  const isTestEnvironment = !!(
    '__playwright' in globalThis ||
    '__jest' in globalThis ||
    navigator.userAgent.includes('HeadlessChrome') ||
    navigator.userAgent.includes('Playwright') ||
    (navigator.userAgent.includes('Firefox') && navigator.webdriver) ||
    import.meta.env.VITE_NODE_ENV === 'test' ||
    import.meta.env.VITE_CI === 'true' ||
    (typeof window !== 'undefined' && window.location.hostname === 'localhost' && 
     (window.location.port === '5173' || window.location.port === '8080'))
  );

  if (!isTestEnvironment) {
    console.log('Not a test environment, skipping test WebSocket initialization');
    return;
  }

  // Prevent multiple initializations
  if (window.__TEST_WEBSOCKET_INITIALIZED__) {
    console.log('🧪 Test WebSockets already initialized, skipping');
    return;
  }

  console.log('🧪 ENABLING COMPONENT WEBSOCKET BLOCKING - All component-level WebSocket operations will be disabled');
  console.log('🧪 Only the store will manage WebSocket connections during tests');
  
  // Set global flags to block component WebSocket operations
  window.__TEST_WEBSOCKET_INITIALIZED__ = true;
  window.__TEST_WEBSOCKET_PERSISTENT_MODE__ = true;
  
  // Additional global flag for component blocking
  window.__INBER_PARTY_TEST_INITIALIZED__ = true;
  window.__WEBSOCKET_PERMANENT_LOCK__ = true;

  console.log('🧪 Test WebSocket initialization complete. Component WebSocket operations are now BLOCKED.');
  console.log('🧪 Store will handle all WebSocket connections exclusively.');
}

/**
 * Clean up test WebSocket connections (only call this when the test session ends)
 */
export function cleanupTestWebSockets() {
  if (!window.__TEST_WEBSOCKET_INITIALIZED__) return;

  console.log('🧪 Cleaning up test WebSocket connections');
  
  // Close all connections
  wsManager.closeAll();
  
  // Reset flags
  window.__TEST_WEBSOCKET_INITIALIZED__ = false;
  window.__TEST_WEBSOCKET_PERSISTENT_MODE__ = false;
}