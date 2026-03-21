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

  console.log('🧪 INITIALIZING TEST WEBSOCKET CONNECTIONS - These will persist for the entire test session');
  
  // Set global flags
  window.__TEST_WEBSOCKET_INITIALIZED__ = true;
  window.__TEST_WEBSOCKET_PERSISTENT_MODE__ = true;

  // Define all WebSocket URLs used in the application
  const webSocketUrls = [
    'ws://localhost:8080/ws',           // Main store connection
    'ws://localhost:8080/api/ws/chat',  // Chat connection
  ];

  // Initialize persistent connections for all URLs
  webSocketUrls.forEach((url, index) => {
    console.log(`🧪 Initializing persistent connection ${index + 1}/${webSocketUrls.length}: ${url}`);
    
    // Add to persistent connections before subscribing
    wsManager.addPersistentConnection(url);
    wsManager.maintainConnection(url);
    
    // Create a persistent subscription that will never be unsubscribed
    const unsubscribe = wsManager.subscribe(
      url,
      `test-persistent-${index}`,
      (data) => {
        console.log(`🧪 Test connection ${url} received data:`, data);
      },
      (connected) => {
        console.log(`🧪 Test connection ${url} state changed: ${connected ? 'CONNECTED' : 'DISCONNECTED'}`);
      }
    );

    // Store unsubscribe function but NEVER call it during tests
    (window as any)[`__TEST_UNSUBSCRIBE_${index}__`] = unsubscribe;
  });

  console.log(`🧪 Test WebSocket initialization complete. ${webSocketUrls.length} persistent connections established.`);
  console.log('🧪 These connections will remain active throughout the entire test session.');

  // Override console.log temporarily to see if connections are being established
  const originalConsoleLog = console.log;
  console.log = (...args) => {
    if (args[0]?.includes?.('WebSocket')) {
      originalConsoleLog('🔍', ...args);
    } else {
      originalConsoleLog(...args);
    }
  };

  // Restore console.log after 5 seconds
  setTimeout(() => {
    console.log = originalConsoleLog;
    console.log('🧪 WebSocket debug logging disabled, connections should now be stable');
  }, 5000);
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