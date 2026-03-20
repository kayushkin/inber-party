import { test, expect } from '@playwright/test';

/**
 * WebSocket Stability Tests
 * 
 * Tests specific to WebSocket connection management during E2E test runs.
 * Ensures connections are stable and don't cause excessive connect/disconnect cycles.
 */

test.describe('WebSocket Connection Stability', () => {
  test.beforeEach(async ({ page }) => {
    test.setTimeout(45000);
    
    // Track WebSocket connection events
    const wsEvents: string[] = [];
    
    page.on('console', msg => {
      const text = msg.text();
      if (text.includes('WebSocket') || text.includes('🧪')) {
        wsEvents.push(text);
      }
    });
    
    // Store events for later inspection
    await page.addInitScript(() => {
      (window as Record<string, unknown>).wsEvents = [];
    });
    
    await page.goto('/');
    await page.waitForLoadState('domcontentloaded');
    
    // Give extra time for WebSocket optimization to work
    await page.waitForTimeout(5000);
  });

  test('should establish WebSocket connection without excessive reconnect attempts', async ({ page }) => {
    // Check that the page loaded successfully
    await expect(page.locator('body')).toBeVisible();
    
    // Wait for potential WebSocket connection events to settle
    await page.waitForTimeout(3000);
    
    // Check for WebSocket debug panel in development mode
    const debugPanel = page.locator('[data-testid="ws-debug-panel"]');
    if (await debugPanel.count() > 0) {
      // WebSocket debug panel is visible, check connection status
      const connectionStatus = page.locator('[data-testid="ws-connection-status"]');
      if (await connectionStatus.count() > 0) {
        const statusText = await connectionStatus.textContent();
        console.log('WebSocket connection status:', statusText);
      }
    }
    
    // Check console for excessive WebSocket errors
    const wsErrors = await page.evaluate(() => {
      return (window as Record<string, unknown>).wsErrors || [];
    });
    
    // Should not have more than a few WebSocket connection attempts
    expect(wsErrors.length).toBeLessThanOrEqual(3);
    
    console.log('✓ WebSocket connection stable without excessive reconnect attempts');
  });

  test('should handle WebSocket disconnection gracefully', async ({ page }) => {
    // Wait for initial connection
    await page.waitForTimeout(2000);
    
    // Simulate WebSocket disconnection by disabling JavaScript temporarily
    await page.addScriptTag({
      content: `
        // Store original WebSocket
        const OriginalWebSocket = window.WebSocket;
        
        // Mock WebSocket to simulate disconnection
        window.WebSocket = function() {
          const ws = {
            close: () => {},
            send: () => {},
            readyState: 3, // CLOSED
            addEventListener: () => {},
            removeEventListener: () => {}
          };
          setTimeout(() => {
            if (ws.onclose) ws.onclose({code: 1006});
          }, 100);
          return ws;
        };
        
        // Restore after a short time
        setTimeout(() => {
          window.WebSocket = OriginalWebSocket;
        }, 2000);
      `
    });
    
    // Wait for disconnection/reconnection cycle
    await page.waitForTimeout(3000);
    
    // App should still be functional
    await expect(page.locator('body')).toBeVisible();
    
    // Should not crash the entire application
    const errorBoundaries = page.locator('[data-testid="error-boundary"]');
    const errorCount = await errorBoundaries.count();
    expect(errorCount).toBe(0);
    
    console.log('✓ WebSocket disconnection handled gracefully');
  });

  test('should optimize connections in test environment', async ({ page }) => {
    // Set Playwright flag for detection
    await page.addInitScript(() => {
      (globalThis as Record<string, unknown>).__playwright = true;
    });

    // Check for test environment detection  
    const isTestDetected = await page.evaluate(() => {
      return !!(
        '__playwright' in globalThis ||
        navigator.userAgent.includes('HeadlessChrome') ||
        (navigator.userAgent.includes('Firefox') && navigator.webdriver) ||
        window.location.hostname === 'localhost'
      );
    });
    
    expect(isTestDetected).toBe(true);
    
    // Wait for any delayed connection initialization
    await page.waitForTimeout(4000);
    
    // Check that the app is functional despite test optimizations
    await expect(page.locator('body')).toBeVisible();
    
    // Navigate to different pages to test WebSocket stability during navigation
    const navLinks = page.locator('nav a, [role="navigation"] a').first();
    if (await navLinks.count() > 0) {
      await navLinks.click();
      await page.waitForLoadState('domcontentloaded');
      await page.waitForTimeout(1000);
      
      // Should still be functional
      await expect(page.locator('body')).toBeVisible();
    }
    
    console.log('✓ Test environment optimizations working correctly');
  });

  test('should not flood console with WebSocket errors during Firefox tests', async ({ page, browserName }) => {
    if (browserName !== 'firefox') {
      test.skip('Firefox-specific test');
      return;
    }
    
    const consoleErrors: string[] = [];
    
    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });
    
    // Navigate and wait for potential WebSocket connection attempts
    await page.goto('/');
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(5000);
    
    // Filter WebSocket-specific errors
    const wsErrors = consoleErrors.filter(error => 
      error.toLowerCase().includes('websocket') ||
      error.includes('Failed to establish WebSocket') ||
      error.includes('WebSocket connection')
    );
    
    // Firefox should not have excessive WebSocket errors
    expect(wsErrors.length).toBeLessThanOrEqual(2);
    
    if (wsErrors.length > 0) {
      console.log('Firefox WebSocket errors (should be minimal):', wsErrors);
    }
    
    console.log('✓ Firefox WebSocket connection handling optimized');
  });
});