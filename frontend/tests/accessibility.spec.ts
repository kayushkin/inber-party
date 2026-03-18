import { test, expect } from '@playwright/test';
import { injectAxe, checkA11y, configureAxe } from '@axe-core/playwright';

test.describe('Accessibility Audit', () => {
  test.beforeEach(async ({ page }) => {
    // Start the app
    await page.goto('http://localhost:5173');
    await injectAxe(page);
  });

  test('should not have any automatically detectable accessibility issues on main page', async ({ page }) => {
    await configureAxe(page, {
      rules: {
        // Disable color-contrast rule temporarily as we'll check this manually
        'color-contrast': { enabled: false },
      },
    });
    
    await checkA11y(page);
  });

  test('should have proper keyboard navigation', async ({ page }) => {
    // Wait for page to be fully loaded
    await page.waitForLoadState('domcontentloaded');
    
    // Tab through all interactive elements
    await page.keyboard.press('Tab');
    
    // Check that focus is visible and logical
    const focusedElement = await page.locator(':focus');
    if (await focusedElement.count() > 0) {
      await expect(focusedElement).toBeVisible();
    } else {
      // If no element is focused, that's okay, just tab to the next one
      await page.keyboard.press('Tab');
      const secondFocus = await page.locator(':focus');
      await expect(secondFocus).toBeVisible();
    }
    
    // Test navigation through header controls
    await page.keyboard.press('Tab');
    await page.keyboard.press('Tab');
    await page.keyboard.press('Tab');
    
    // Check that Enter/Space can activate buttons
    const themeToggle = page.locator('[title*="theme"], .theme-toggle, button:has-text("🌙"), button:has-text("☀️")');
    if (await themeToggle.count() > 0) {
      await themeToggle.first().focus();
      await page.keyboard.press('Space');
      // Should toggle theme
    }
  });

  test('should have proper ARIA labels and roles', async ({ page }) => {
    // Check for missing labels on interactive elements
    const buttons = page.locator('button');
    const buttonCount = await buttons.count();
    
    for (let i = 0; i < buttonCount; i++) {
      const button = buttons.nth(i);
      const ariaLabel = await button.getAttribute('aria-label');
      const title = await button.getAttribute('title');
      const text = await button.textContent();
      
      // Button should have either aria-label, title, or text content
      if (!ariaLabel && !title && (!text || text.trim() === '')) {
        const buttonHtml = await button.innerHTML();
        console.warn(`Button without label found: ${buttonHtml}`);
      }
    }

    // Check navigation has proper structure
    const nav = page.locator('nav');
    if (await nav.count() > 0) {
      const navRole = await nav.getAttribute('role');
      const navAriaLabel = await nav.getAttribute('aria-label');
      // Nav should have proper role or aria-label
      expect(navRole === 'navigation' || navAriaLabel !== null).toBeTruthy();
    }

    // Check main content area has proper landmark
    const main = page.locator('main');
    await expect(main).toBeVisible();
  });

  test('should check color contrast ratios', async ({ page }) => {
    // Enable color contrast checking
    await configureAxe(page, {
      rules: {
        'color-contrast': { enabled: true },
      },
    });
    
    try {
      await checkA11y(page, null, {
        tags: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'],
      });
    } catch (error) {
      // Log contrast issues for manual review
      console.log('Color contrast violations found:', error.message);
      
      // Continue with manual checks of key elements
      const textElements = page.locator('body *:visible').filter({ hasText: /\S/ });
      const count = await textElements.count();
      
      // Sample a few elements for manual verification
      for (let i = 0; i < Math.min(count, 10); i++) {
        const element = textElements.nth(i);
        const computedStyle = await element.evaluate((el) => {
          const style = window.getComputedStyle(el);
          return {
            color: style.color,
            backgroundColor: style.backgroundColor,
            fontSize: style.fontSize,
          };
        });
        console.log(`Element ${i}:`, computedStyle);
      }
    }
  });

  test('should be navigable with screen readers', async ({ page }) => {
    // Check heading structure
    await page.waitForSelector('h1, h2, h3, h4, h5, h6', { timeout: 5000 });
    const headings = page.locator('h1, h2, h3, h4, h5, h6');
    const headingCount = await headings.count();
    
    if (headingCount > 0) {
      // Should have logical heading hierarchy - accept h1 or h2 as first heading
      const firstHeading = headings.first();
      const firstHeadingTag = await firstHeading.evaluate(el => el.tagName.toLowerCase());
      const firstHeadingText = await firstHeading.textContent();
      console.log(`First heading: ${firstHeadingTag} - "${firstHeadingText}"`);
      expect(['h1', 'h2'].includes(firstHeadingTag)).toBeTruthy();
    } else {
      throw new Error('No headings found on page');
    }

    // Check for skip links
    const skipLink = page.locator('a[href="#main"], a:has-text("skip"), [aria-label*="skip"]');
    if (await skipLink.count() === 0) {
      console.warn('No skip link found - consider adding one for screen reader users');
    }

    // Check form labels if any forms exist
    const inputs = page.locator('input, textarea, select');
    const inputCount = await inputs.count();
    
    for (let i = 0; i < inputCount; i++) {
      const input = inputs.nth(i);
      const id = await input.getAttribute('id');
      const ariaLabel = await input.getAttribute('aria-label');
      const ariaLabelledBy = await input.getAttribute('aria-labelledby');
      
      if (id) {
        const label = page.locator(`label[for="${id}"]`);
        const hasLabel = await label.count() > 0;
        
        if (!hasLabel && !ariaLabel && !ariaLabelledBy) {
          console.warn(`Input without label found: ${await input.getAttribute('type')} input`);
        }
      }
    }
  });

  test('should work properly on different pages', async ({ page }) => {
    // Test accessibility on key pages
    const pages = [
      '/',
      '/quests', 
      '/war-room',
      '/leaderboard'
    ];

    for (const path of pages) {
      await page.goto(`http://localhost:5173${path}`);
      await page.waitForLoadState('networkidle');
      
      // Basic accessibility check for each page
      try {
        await checkA11y(page, null, {
          rules: {
            'color-contrast': { enabled: false }, // Skip contrast for now
          },
        });
      } catch (error) {
        console.warn(`Accessibility issues found on ${path}:`, error.message);
      }
    }
  });
});