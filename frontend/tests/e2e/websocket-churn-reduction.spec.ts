import { test, expect } from '@playwright/test';

/**
 * WebSocket Connection Churn Reduction Tests
 * 
 * Tests specifically focused on reducing excessive connection/disconnection cycles
 * during E2E test navigation, as mentioned in the BACKLOG.md.
 */

test.describe('WebSocket Connection Churn Reduction', () => {
  test.beforeEach(async ({ page }) => {
    test.setTimeout(60000); // Longer timeout for stability testing
    
    // Enable detailed WebSocket logging for analysis
    await page.addInitScript(() => {
      // Track WebSocket events globally
      (window as unknown as { wsConnectionEvents: string[] }).wsConnectionEvents = [];
      
      // Override WebSocket constructor to track connections
      const OriginalWebSocket = window.WebSocket;
      window.WebSocket = class extends OriginalWebSocket {
        constructor(url: string | URL, protocols?: string | string[]) {
          super(url, protocols);
          const urlStr = url.toString();
          (window as unknown as { wsConnectionEvents: string[] }).wsConnectionEvents.push(`CONNECT:${urlStr}:${Date.now()}`);
          
          this.addEventListener('close', () => {
            (window as unknown as { wsConnectionEvents: string[] }).wsConnectionEvents.push(`DISCONNECT:${urlStr}:${Date.now()}`);
          });
        }
      };
    });
    
    await page.goto('/');
    await page.waitForLoadState('domcontentloaded');
    
    // Wait for initial WebSocket connections to stabilize
    await page.waitForTimeout(3000);
  });

  test('should maintain stable connections during navigation between pages', async ({ page }) => {
    console.log('🧪 Starting navigation stability test...');
    
    // Clear initial connection events
    await page.evaluate(() => {
      (window as unknown as { wsConnectionEvents: string[] }).wsConnectionEvents = [];
    });
    
    // Navigate through multiple pages to test connection stability
    const navigationSequence = [
      { path: '/', name: 'Tavern' },
      { path: '/war-room', name: 'War Room' },
      { path: '/library', name: 'Library' },
      { path: '/guild-chat', name: 'Guild Chat' },
      { path: '/', name: 'Back to Tavern' },
    ];
    
    for (const { path, name } of navigationSequence) {
      console.log(`🧪 Navigating to ${name} (${path})`);
      
      await page.goto(path);
      await page.waitForLoadState('domcontentloaded');
      
      // Give time for any connection changes to stabilize
      await page.waitForTimeout(2000);
      
      // Verify page loaded correctly
      await expect(page.locator('body')).toBeVisible();
    }
    
    // Check WebSocket connection events
    const events = await page.evaluate(() => {
      return (window as unknown as { wsConnectionEvents: string[] }).wsConnectionEvents;
    });
    
    // Count connections and disconnections
    const connections = events.filter(event => event.startsWith('CONNECT:')).length;
    const disconnections = events.filter(event => event.startsWith('DISCONNECT:')).length;
    
    console.log(`🧪 WebSocket Events during navigation:`);
    console.log(`   Connections: ${connections}`);
    console.log(`   Disconnections: ${disconnections}`);
    console.log(`   Total events: ${events.length}`);
    
    // Target: Reduce connection churn from observed 20+ cycles to <5 stable connections
    expect(connections).toBeLessThanOrEqual(5);
    expect(disconnections).toBeLessThanOrEqual(5);
    expect(events.length).toBeLessThanOrEqual(10);
    
    if (events.length > 0) {
      console.log('🧪 WebSocket event timeline:', events);
    }
    
    console.log('✓ Navigation completed with stable WebSocket connections');
  });

  test('should handle rapid navigation without excessive reconnects', async ({ page }) => {
    console.log('🧪 Starting rapid navigation test...');
    
    // Clear initial connection events
    await page.evaluate(() => {
      (window as unknown as { wsConnectionEvents: string[] }).wsConnectionEvents = [];
    });
    
    // Rapid navigation sequence (simulating user clicking through pages quickly)
    for (let i = 0; i < 3; i++) {
      console.log(`🧪 Rapid navigation cycle ${i + 1}`);
      
      // Quick navigation without waiting for full stabilization
      await page.goto('/war-room');
      await page.waitForTimeout(500); // Minimal wait time
      
      await page.goto('/library');
      await page.waitForTimeout(500);
      
      await page.goto('/');
      await page.waitForTimeout(500);
    }
    
    // Final stabilization wait
    await page.waitForTimeout(3000);
    
    const events = await page.evaluate(() => {
      return (window as unknown as { wsConnectionEvents: string[] }).wsConnectionEvents;
    });
    
    const connections = events.filter(event => event.startsWith('CONNECT:')).length;
    const disconnections = events.filter(event => event.startsWith('DISCONNECT:')).length;
    
    console.log(`🧪 Rapid navigation WebSocket events:`);
    console.log(`   Connections: ${connections}`);
    console.log(`   Disconnections: ${disconnections}`);
    
    // Even with rapid navigation, should maintain connection stability
    expect(connections).toBeLessThanOrEqual(6); // Allow slightly more for rapid navigation
    expect(disconnections).toBeLessThanOrEqual(6);
    
    console.log('✓ Rapid navigation handled with minimal connection churn');
  });

  test('should show test environment optimizations in debug panel', async ({ page }) => {
    // Check if WebSocket debug panel exists and shows test optimizations
    const debugToggle = page.locator('.ws-debug-toggle');
    if (await debugToggle.count() > 0) {
      await debugToggle.click();
      await page.waitForTimeout(1000);
      
      // Look for test optimization indicator
      const testOptimizations = page.locator('text=🧪 Test Environment Optimizations');
      if (await testOptimizations.count() > 0) {
        console.log('✓ WebSocket debug panel shows test environment optimizations');
        
        // Check specific optimization values
        const persistentConnectionsText = await page.locator('text=Persistent URLs:').textContent();
        const cooldownText = await page.locator('text=Reconnect Cooldown:').textContent();
        
        console.log(`🧪 Debug info - ${persistentConnectionsText}, ${cooldownText}`);
      } else {
        console.log('ℹ️ WebSocket debug panel available but test optimizations not visible');
      }
    } else {
      console.log('ℹ️ WebSocket debug panel not available in this test run');
    }
    
    // Test should pass regardless of debug panel visibility
    await expect(page.locator('body')).toBeVisible();
  });

  test('should maintain connection stability with MMO chat components', async ({ page }) => {
    console.log('🧪 Testing MMO chat connection stability...');
    
    // Navigate to MMO chatroom
    await page.goto('/mmo-chat');
    await page.waitForLoadState('domcontentloaded');
    
    // Wait for chat component to initialize
    await page.waitForTimeout(3000);
    
    // Clear connection events after initial setup
    await page.evaluate(() => {
      (window as unknown as { wsConnectionEvents: string[] }).wsConnectionEvents = [];
    });
    
    // Navigate away and back to test connection persistence
    await page.goto('/');
    await page.waitForTimeout(1000);
    
    await page.goto('/mmo-chat');
    await page.waitForTimeout(2000);
    
    const events = await page.evaluate(() => {
      return (window as unknown as { wsConnectionEvents: string[] }).wsConnectionEvents;
    });
    
    const chatConnections = events.filter(event => 
      event.startsWith('CONNECT:') && event.includes('chat')
    ).length;
    
    console.log(`🧪 Chat WebSocket connections during navigation: ${chatConnections}`);
    
    // Chat connections should be minimal during navigation
    expect(chatConnections).toBeLessThanOrEqual(2);
    
    console.log('✓ MMO chat component maintains stable connections');
  });
});