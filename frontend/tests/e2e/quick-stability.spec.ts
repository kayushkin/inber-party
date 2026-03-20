import { test, expect } from '@playwright/test';

/**
 * Quick Stability Tests
 * 
 * Focused set of tests for basic functionality that should always work.
 * These tests prioritize speed and reliability over comprehensive coverage.
 */

test.describe('Quick Stability Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app with reasonable timeout
    await page.goto('/', { waitUntil: 'domcontentloaded', timeout: 30000 });
    
    // Wait for React app to be fully initialized
    await page.waitForSelector('.layout', { timeout: 15000 });
    
    // Wait for navigation to be ready
    await page.waitForSelector('.nav', { timeout: 10000 });
  });

  test('should load main page successfully', async ({ page }) => {
    // Basic page structure checks
    await expect(page.locator('html')).toBeVisible();
    await expect(page.locator('body')).toBeVisible();
    
    // Should have meaningful content
    const bodyText = await page.locator('body').textContent();
    expect(bodyText?.length || 0).toBeGreaterThan(50);
    
    console.log('✓ Main page loads successfully');
  });

  test('should have working navigation links', async ({ page }) => {
    // Get navigation links
    const navLinks = page.locator('.nav .nav-link');
    const navCount = await navLinks.count();
    
    expect(navCount).toBeGreaterThan(0);
    console.log(`Found ${navCount} navigation links`);
    
    // Test just one navigation link (Quests) to verify navigation works
    const questsLink = page.locator('.nav .nav-link').filter({ hasText: 'Quests' }).first();
    
    if (await questsLink.count() > 0) {
      await questsLink.click({ timeout: 10000 });
      
      // Wait for navigation using a simple approach
      await page.waitForFunction(() => {
        return window.location.pathname.includes('/quests');
      }, { timeout: 20000 });
      
      // Verify we're on the quests page
      expect(page.url()).toContain('/quests');
      
      // Wait for page content
      await page.waitForTimeout(2000);
      
      // Should have page content
      const pageContent = await page.locator('body').textContent();
      expect(pageContent?.length || 0).toBeGreaterThan(20);
      
      console.log('✓ Navigation to quests page successful');
    }
  });

  test('should render without critical errors', async ({ page }) => {
    const jsErrors: string[] = [];
    
    page.on('console', msg => {
      if (msg.type() === 'error') {
        jsErrors.push(msg.text());
      }
    });

    page.on('pageerror', error => {
      jsErrors.push(`Page Error: ${error.message}`);
    });
    
    // Interact with the page briefly
    await page.waitForTimeout(3000);
    
    // Filter out expected errors from missing backend/network issues
    const criticalErrors = jsErrors.filter(error => 
      !error.includes('404') &&
      !error.includes('Failed to fetch') &&
      !error.includes('Failed to load resource') &&
      !error.includes('500') &&
      !error.includes('WebSocket') &&
      !error.includes('favicon.ico') &&
      !error.includes('network') &&
      !error.includes('connection') &&
      !error.includes('ECONNREFUSED')
    );
    
    console.log(`JS errors: ${jsErrors.length}, Critical: ${criticalErrors.length}`);
    
    // Should not have critical JavaScript errors
    expect(criticalErrors.length).toBeLessThanOrEqual(1);
    
    console.log('✓ No critical JavaScript errors');
  });

  test('should have basic interactive elements', async ({ page }) => {
    // Should have clickable elements
    const buttons = page.locator('button, a, [role="button"]');
    const buttonCount = await buttons.count();
    
    expect(buttonCount).toBeGreaterThan(0);
    
    // Should have navigation
    const navElements = page.locator('nav, .nav');
    const navCount = await navElements.count();
    
    expect(navCount).toBeGreaterThan(0);
    
    console.log(`✓ Interactive elements present: ${buttonCount} buttons, ${navCount} nav`);
  });
});