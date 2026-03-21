import { test, expect } from '@playwright/test';

test.describe('WebSocket Duration Analysis', () => {
  test('should measure WebSocket connection duration during normal navigation', async ({ page }) => {
    const connectionEvents: Array<{type: 'connect' | 'disconnect', timestamp: number, url?: string}> = [];
    
    // Monitor console messages for WebSocket events
    page.on('console', (msg) => {
      const text = msg.text();
      if (text.includes('WebSocket connected')) {
        connectionEvents.push({ type: 'connect', timestamp: Date.now() });
      } else if (text.includes('WebSocket disconnected')) {
        connectionEvents.push({ type: 'disconnect', timestamp: Date.now() });
      }
    });

    console.log('🔍 Starting WebSocket duration analysis...');
    const startTime = Date.now();
    
    // Go to main page and stay there for a while
    await page.goto('/');
    await page.waitForSelector('h1', { timeout: 10000 });
    console.log('✓ Main page loaded');
    
    // Wait a bit to see how long connections last on a stable page
    await page.waitForTimeout(5000);
    console.log('✓ Waited 5 seconds on main page');
    
    // Navigate to quests page
    await page.click('a[href="/quests"]');
    await page.waitForSelector('h1', { timeout: 5000 });
    console.log('✓ Navigated to quests page');
    
    // Wait again
    await page.waitForTimeout(3000);
    console.log('✓ Waited 3 seconds on quests page');
    
    // Navigate back
    await page.click('a[href="/"]');
    await page.waitForSelector('h1', { timeout: 5000 });
    console.log('✓ Navigated back to main page');
    
    // Final wait
    await page.waitForTimeout(2000);
    console.log('✓ Final wait completed');
    
    const totalTestTime = Date.now() - startTime;
    
    // Calculate connection statistics
    const connectCount = connectionEvents.filter(e => e.type === 'connect').length;
    const disconnectCount = connectionEvents.filter(e => e.type === 'disconnect').length;
    
    // Calculate average connection duration (rough approximation)
    let totalConnectionTime = 0;
    let connectionDurations: number[] = [];
    
    for (let i = 0; i < connectionEvents.length - 1; i++) {
      const current = connectionEvents[i];
      const next = connectionEvents[i + 1];
      
      if (current.type === 'connect' && next.type === 'disconnect') {
        const duration = next.timestamp - current.timestamp;
        connectionDurations.push(duration);
        totalConnectionTime += duration;
      }
    }
    
    const averageConnectionDuration = connectionDurations.length > 0 ? 
      totalConnectionTime / connectionDurations.length : 0;
    
    console.log('📊 WebSocket Connection Analysis:');
    console.log(`  Total test time: ${totalTestTime}ms (${(totalTestTime/1000).toFixed(1)}s)`);
    console.log(`  Total connections: ${connectCount}`);
    console.log(`  Total disconnections: ${disconnectCount}`);
    console.log(`  Connection durations: ${connectionDurations.map(d => (d/1000).toFixed(1) + 's').join(', ')}`);
    console.log(`  Average connection duration: ${(averageConnectionDuration/1000).toFixed(1)}s`);
    console.log(`  Connection churn rate: ${(connectCount / (totalTestTime/1000)).toFixed(2)} connections/second`);
    
    // Define acceptable thresholds
    const maxChurnRate = 0.5; // No more than 0.5 connections per second
    const minConnectionDuration = 1000; // Connections should last at least 1 second
    
    const actualChurnRate = connectCount / (totalTestTime / 1000);
    
    console.log('\n🎯 Performance Expectations:');
    console.log(`  Expected churn rate: < ${maxChurnRate} connections/second`);
    console.log(`  Actual churn rate: ${actualChurnRate.toFixed(2)} connections/second`);
    console.log(`  Expected min duration: > ${minConnectionDuration/1000}s`);
    console.log(`  Actual avg duration: ${(averageConnectionDuration/1000).toFixed(1)}s`);
    
    // For now, just log the results - we'll use this to establish reasonable baselines
    // expect(actualChurnRate).toBeLessThan(maxChurnRate);
    // expect(averageConnectionDuration).toBeGreaterThan(minConnectionDuration);
    
    console.log('\n✅ Analysis complete - check the logs above to determine if connection behavior is reasonable');
  });
  
  test('should test WebSocket connection stability on single page', async ({ page }) => {
    const connectionEvents: Array<{type: string, timestamp: number}> = [];
    
    page.on('console', (msg) => {
      const text = msg.text();
      if (text.includes('WebSocket') && (text.includes('connected') || text.includes('disconnected'))) {
        connectionEvents.push({ type: text.includes('connected') ? 'connect' : 'disconnect', timestamp: Date.now() });
      }
    });

    console.log('🔍 Testing single-page WebSocket stability...');
    
    await page.goto('/');
    await page.waitForSelector('h1', { timeout: 10000 });
    
    // Stay on the same page for a longer time to see if connections churn unnecessarily
    await page.waitForTimeout(10000); // 10 seconds on same page
    
    const connectCount = connectionEvents.filter(e => e.type === 'connect').length;
    const disconnectCount = connectionEvents.filter(e => e.type === 'disconnect').length;
    
    console.log(`📊 Single-page stability (10s): ${connectCount} connects, ${disconnectCount} disconnects`);
    
    // On a single stable page, we should have very few connections and even fewer disconnections
    expect(connectCount).toBeLessThanOrEqual(2); // Max 2 initial connections (main + chat)
    expect(disconnectCount).toBeLessThanOrEqual(1); // Should have minimal disconnections on stable page
  });
});