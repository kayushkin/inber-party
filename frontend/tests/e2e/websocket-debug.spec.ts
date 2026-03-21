import { test, expect } from '@playwright/test';

// Type definitions for test environment WebSocket manager
interface WebSocketManager {
  getTestOptimizationStatus(): {
    isTestEnvironment: boolean;
    testPersistenceMode: boolean;
    persistentConnections: string[];
  };
  getStats(): {
    totalConnections: number;
    activeConnections: number;
    reconnectAttempts: number;
  };
}

declare global {
  interface Window {
    wsManager: WebSocketManager;
  }
}

test.describe('WebSocket Debug Tests', () => {
  test('should show WebSocket manager status', async ({ page }) => {
    // Go to main page
    await page.goto('/');
    
    // Wait for app to load
    await page.waitForSelector('h1', { timeout: 10000 });
    
    // Inject debug script to check WebSocket manager status
    const wsManagerStatus = await page.evaluate(() => {
      // Access the WebSocket manager instance
      const wsManager = window.wsManager;
      
      if (!wsManager) {
        return { error: 'WebSocket manager not found in window' };
      }
      
      // Get test optimization status
      const testStatus = wsManager.getTestOptimizationStatus();
      const stats = wsManager.getStats();
      
      return {
        testOptimizationStatus: testStatus,
        connectionStats: stats,
        isTestEnvironmentDetected: testStatus?.isTestEnvironment || false,
        testPersistenceMode: testStatus?.testPersistenceMode || false,
        persistentConnections: testStatus?.persistentConnections || []
      };
    });
    
    console.log('WebSocket Manager Status:', JSON.stringify(wsManagerStatus, null, 2));
    
    // Check if test environment was detected
    expect(wsManagerStatus.isTestEnvironmentDetected).toBe(true);
    expect(wsManagerStatus.testPersistenceMode).toBe(true);
    
    // Navigate to different page and check if connections persist
    await page.click('a[href="/quests"]');
    await page.waitForSelector('h1', { timeout: 5000 });
    
    // Wait a moment for any connection changes
    await page.waitForTimeout(2000);
    
    const wsManagerStatusAfterNav = await page.evaluate(() => {
      const wsManager = window.wsManager;
      return wsManager ? {
        connectionStats: wsManager.getStats(),
        testOptimizationStatus: wsManager.getTestOptimizationStatus()
      } : { error: 'WebSocket manager not found' };
    });
    
    console.log('WebSocket Manager Status After Navigation:', JSON.stringify(wsManagerStatusAfterNav, null, 2));
  });
  
  test('should check WebSocket connection behavior', async ({ page }) => {
    let wsConnectCount = 0;
    let wsDisconnectCount = 0;
    
    // Monitor console logs for WebSocket connection events
    page.on('console', (msg) => {
      const text = msg.text();
      if (text.includes('WebSocket connected')) {
        wsConnectCount++;
      }
      if (text.includes('WebSocket disconnected')) {
        wsDisconnectCount++;
      }
    });
    
    // Go to main page
    await page.goto('/');
    await page.waitForSelector('h1', { timeout: 10000 });
    
    // Navigate through several pages quickly
    await page.click('a[href="/quests"]');
    await page.waitForSelector('h1', { timeout: 5000 });
    
    await page.click('a[href="/"]');
    await page.waitForSelector('h1', { timeout: 5000 });
    
    await page.click('a[href="/quests"]');
    await page.waitForSelector('h1', { timeout: 5000 });
    
    // Wait for any delayed operations
    await page.waitForTimeout(3000);
    
    console.log(`WebSocket Connections: ${wsConnectCount}, Disconnections: ${wsDisconnectCount}`);
    
    // In ultra-persistent mode, we should have minimal disconnections
    // Allow some initial connections but expect very few disconnections
    expect(wsDisconnectCount).toBeLessThan(5); // Should be much less in ultra-persistent mode
  });
});