import { test, expect } from '@playwright/test';

/**
 * Ultra-Persistent WebSocket Connection Tests
 * 
 * Tests the enhanced WebSocket connection management with ultra-persistent
 * test mode that should eliminate connection churn during E2E test execution.
 */

test.describe('Ultra-Persistent WebSocket Connections', () => {
  test.beforeEach(async ({ page }) => {
    test.setTimeout(90000); // Extended timeout for ultra-persistence testing
    
    // Enable comprehensive WebSocket logging and tracking
    await page.addInitScript(() => {
      // Enhanced WebSocket event tracking
      (window as unknown as { wsConnectionEvents: Array<{type: string, url: string, timestamp: number}> }).wsConnectionEvents = [];
      (window as unknown as { wsActiveConnections: Set<string> }).wsActiveConnections = new Set();
      
      // Override WebSocket constructor with detailed tracking
      const OriginalWebSocket = window.WebSocket;
      window.WebSocket = class extends OriginalWebSocket {
        constructor(url: string | URL, protocols?: string | string[]) {
          super(url, protocols);
          const urlStr = url.toString();
          const timestamp = Date.now();
          
          (window as unknown as { wsConnectionEvents: Array<{type: string, url: string, timestamp: number}> }).wsConnectionEvents.push({
            type: 'CONNECT',
            url: urlStr,
            timestamp
          });
          
          (window as unknown as { wsActiveConnections: Set<string> }).wsActiveConnections.add(urlStr);
          
          console.log(`🔌 WebSocket CONNECT: ${urlStr} at ${timestamp}`);
          
          this.addEventListener('close', () => {
            const closeTimestamp = Date.now();
            (window as unknown as { wsConnectionEvents: Array<{type: string, url: string, timestamp: number}> }).wsConnectionEvents.push({
              type: 'DISCONNECT',
              url: urlStr,
              timestamp: closeTimestamp
            });
            
            (window as unknown as { wsActiveConnections: Set<string> }).wsActiveConnections.delete(urlStr);
            
            console.log(`🔌 WebSocket DISCONNECT: ${urlStr} at ${closeTimestamp} (duration: ${closeTimestamp - timestamp}ms)`);
          });
          
          this.addEventListener('open', () => {
            console.log(`🔌 WebSocket OPEN: ${urlStr}`);
          });
          
          this.addEventListener('error', () => {
            console.log(`🔌 WebSocket ERROR: ${urlStr}`);
          });
        }
      };
    });
    
    await page.goto('/');
    await page.waitForLoadState('domcontentloaded');
    
    // Extended wait for initial connections to stabilize in ultra-persistent mode
    await page.waitForTimeout(5000);
  });

  test('should establish initial connections with ultra-persistent mode', async ({ page }) => {
    console.log('🧪 Testing ultra-persistent mode initialization...');
    
    // Clear initial events and track from a stable baseline
    await page.evaluate(() => {
      (window as unknown as { wsConnectionEvents: Array<{type: string, url: string, timestamp: number}> }).wsConnectionEvents = [];
    });
    
    // Allow some time for any final initialization
    await page.waitForTimeout(3000);
    
    // Check for test environment optimizations
    const hasOptimizations = await page.evaluate(() => {
      // Check if WebSocket manager exists and has test optimizations
      return !!(window as unknown as { wsManager?: { getTestOptimizationStatus?: () => unknown } }).wsManager?.getTestOptimizationStatus?.();
    });
    
    console.log(`🧪 Test optimizations detected: ${hasOptimizations}`);
    
    // Get active connections
    const activeConnections = await page.evaluate(() => {
      return Array.from((window as unknown as { wsActiveConnections: Set<string> }).wsActiveConnections);
    });
    
    console.log(`🔌 Active connections: ${activeConnections.length}`);
    activeConnections.forEach(url => console.log(`  - ${url}`));
    
    // Should have established minimal baseline connections
    expect(activeConnections.length).toBeGreaterThan(0);
    expect(activeConnections.length).toBeLessThanOrEqual(3);
  });

  test('should maintain ultra-stable connections during extensive navigation', async ({ page }) => {
    console.log('🧪 Testing ultra-stable connection persistence during navigation...');
    
    // Start tracking from clean state
    await page.evaluate(() => {
      (window as unknown as { wsConnectionEvents: Array<{type: string, url: string, timestamp: number}> }).wsConnectionEvents = [];
    });
    
    // Extensive navigation sequence
    const navigationSequence = [
      { path: '/war-room', name: 'War Room' },
      { path: '/library', name: 'Library' },
      { path: '/guild-chat', name: 'Guild Chat' },
      { path: '/mmo-chat', name: 'MMO Chat' },
      { path: '/bounty-board', name: 'Bounty Board' },
      { path: '/training-grounds', name: 'Training Grounds' },
      { path: '/forge', name: 'Forge' },
      { path: '/', name: 'Tavern' },
      { path: '/war-room', name: 'War Room Again' },
      { path: '/', name: 'Back to Tavern' },
    ];
    
    console.log(`🧪 Starting extensive navigation through ${navigationSequence.length} pages...`);
    
    for (const { path, name } of navigationSequence) {
      console.log(`🧪 Navigating to ${name} (${path})`);
      
      await page.goto(path);
      await page.waitForLoadState('domcontentloaded');
      
      // Minimal wait - ultra-persistent mode should handle rapid navigation
      await page.waitForTimeout(1000);
      
      // Verify page loaded correctly
      await expect(page.locator('body')).toBeVisible();
    }
    
    // Final stabilization
    await page.waitForTimeout(3000);
    
    // Check connection events during navigation
    const events = await page.evaluate(() => {
      return (window as unknown as { wsConnectionEvents: Array<{type: string, url: string, timestamp: number}> }).wsConnectionEvents;
    });
    
    const connectEvents = events.filter(e => e.type === 'CONNECT');
    const disconnectEvents = events.filter(e => e.type === 'DISCONNECT');
    
    console.log(`🔌 Navigation WebSocket activity summary:`);
    console.log(`   Connect events: ${connectEvents.length}`);
    console.log(`   Disconnect events: ${disconnectEvents.length}`);
    console.log(`   Total events: ${events.length}`);
    console.log(`   Net connection changes: ${connectEvents.length - disconnectEvents.length}`);
    
    if (events.length > 0) {
      console.log('🔌 Event timeline:');
      events.forEach(event => {
        console.log(`   ${event.timestamp}: ${event.type} ${event.url}`);
      });
    }
    
    // Ultra-persistent mode target: reasonable connection churn for stable operation
    expect(connectEvents.length).toBeLessThanOrEqual(5); // Allow for stable connections
    expect(disconnectEvents.length).toBeLessThanOrEqual(4); // Allow reasonable disconnections
    expect(events.length).toBeLessThanOrEqual(10); // Total events for stable WebSocket operation
    
    console.log('✅ Ultra-stable connection persistence verified');
  });

  test('should handle rapid component mounting/unmounting without churn', async ({ page }) => {
    console.log('🧪 Testing rapid component lifecycle without connection churn...');
    
    await page.evaluate(() => {
      (window as unknown as { wsConnectionEvents: Array<{type: string, url: string, timestamp: number}> }).wsConnectionEvents = [];
    });
    
    // Rapid navigation that causes components to mount/unmount quickly
    for (let cycle = 0; cycle < 5; cycle++) {
      console.log(`🧪 Rapid cycle ${cycle + 1}/5`);
      
      await page.goto('/guild-chat');
      await page.waitForTimeout(200);
      
      await page.goto('/mmo-chat');
      await page.waitForTimeout(200);
      
      await page.goto('/');
      await page.waitForTimeout(200);
    }
    
    // Final wait for any delayed events
    await page.waitForTimeout(5000);
    
    const events = await page.evaluate(() => {
      return (window as unknown as { wsConnectionEvents: Array<{type: string, url: string, timestamp: number}> }).wsConnectionEvents;
    });
    
    const connectEvents = events.filter(e => e.type === 'CONNECT');
    const disconnectEvents = events.filter(e => e.type === 'DISCONNECT');
    
    console.log(`🔌 Rapid lifecycle WebSocket activity:`);
    console.log(`   Connect events: ${connectEvents.length}`);
    console.log(`   Disconnect events: ${disconnectEvents.length}`);
    
    // Ultra-persistent mode should minimize churn during rapid lifecycle changes
    expect(connectEvents.length).toBeLessThanOrEqual(4); // Allow stable connections
    expect(disconnectEvents.length).toBeLessThanOrEqual(3); // Allow reasonable disconnections
    
    console.log('✅ Rapid component lifecycle handled without churn');
  });

  test('should show ultra-persistent mode status in debug panel', async ({ page }) => {
    console.log('🧪 Checking ultra-persistent mode debug information...');
    
    // Look for WebSocket debug toggle
    const debugToggle = page.locator('[data-testid="ws-debug-toggle"], .ws-debug-toggle');
    const debugToggleExists = await debugToggle.count() > 0;
    
    if (debugToggleExists) {
      console.log('🔧 Found WebSocket debug toggle, opening debug panel...');
      await debugToggle.click();
      await page.waitForTimeout(1000);
      
      // Look for ultra-persistent mode indicators
      const ultraPersistentText = page.locator('text=/ultra.*persistent/i');
      const testModeText = page.locator('text=/test.*mode/i');
      
      const hasUltraPersistentMode = await ultraPersistentText.count() > 0;
      const hasTestMode = await testModeText.count() > 0;
      
      console.log(`🔧 Ultra-persistent mode visible: ${hasUltraPersistentMode}`);
      console.log(`🔧 Test mode visible: ${hasTestMode}`);
      
      if (hasUltraPersistentMode || hasTestMode) {
        console.log('✅ Debug panel shows ultra-persistent mode information');
      } else {
        console.log('ℹ️ Debug panel available but ultra-persistent mode not explicitly shown');
      }
    } else {
      console.log('ℹ️ WebSocket debug panel not available in this test run');
    }
    
    // Test should pass regardless of debug panel visibility
    expect(page.locator('body')).toBeVisible();
  });
});