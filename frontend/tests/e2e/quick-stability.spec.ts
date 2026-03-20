import { test, expect } from '@playwright/test';

/**
 * Quick Stability Test
 * 
 * Minimal test suite to verify basic functionality and server stability
 * without the overhead of comprehensive test suites.
 */

test.describe('Quick Stability Check', () => {
  test('should load homepage and have basic functionality', async ({ page }) => {
    // Set moderate timeout
    test.setTimeout(30000);
    
    // Navigate to the app
    await page.goto('/', { waitUntil: 'domcontentloaded', timeout: 15000 });
    
    // Wait for layout to load
    await page.waitForSelector('.layout', { timeout: 10000 });
    
    // Check basic page structure
    await expect(page.locator('.header')).toBeVisible();
    await expect(page.locator('.nav')).toBeVisible();
    await expect(page.locator('.main')).toBeVisible();
    
    // Check that navigation links exist
    const navLinks = page.locator('.nav .nav-link');
    const linkCount = await navLinks.count();
    expect(linkCount).toBeGreaterThan(5);
    
    console.log(`✓ Homepage loaded with ${linkCount} navigation links`);
  });
  
  test('should be able to navigate to one other page', async ({ page }) => {
    // Set moderate timeout
    test.setTimeout(25000);
    
    // Navigate to homepage
    await page.goto('/', { waitUntil: 'domcontentloaded', timeout: 15000 });
    await page.waitForSelector('.nav', { timeout: 10000 });
    
    // Find a simple link to test (stats page)
    const statsLink = page.locator('.nav .nav-link[href="/stats"]');
    await expect(statsLink).toBeVisible();
    
    // Click and navigate
    await statsLink.click();
    await page.waitForURL('**/stats', { timeout: 10000 });
    
    // Check that we're on the stats page
    await expect(page.locator('body')).toBeVisible();
    expect(page.url()).toContain('/stats');
    
    console.log('✓ Navigation to stats page successful');
  });
  
  test('should handle basic interactions without crashes', async ({ page }) => {
    // Set moderate timeout
    test.setTimeout(20000);
    
    // Navigate to homepage
    await page.goto('/', { waitUntil: 'domcontentloaded', timeout: 15000 });
    await page.waitForSelector('.layout', { timeout: 10000 });
    
    // Try clicking a few different elements
    const clickableElements = [
      '.theme-toggle',
      '.nav-link:first-child',
    ];
    
    let successfulClicks = 0;
    for (const selector of clickableElements) {
      try {
        const element = page.locator(selector);
        const isVisible = await element.isVisible();
        if (isVisible) {
          await element.click({ timeout: 3000 });
          await page.waitForTimeout(500);
          successfulClicks++;
        }
      } catch (error) {
        console.log(`Click on ${selector} failed: ${(error as Error).message}`);
      }
    }
    
    expect(successfulClicks).toBeGreaterThan(0);
    console.log(`✓ ${successfulClicks} interactions successful`);
  });
});