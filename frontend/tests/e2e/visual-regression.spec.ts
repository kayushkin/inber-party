import { test, expect } from '@playwright/test';

test.describe('Visual Regression Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to load
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Wait longer for WebSocket connections and dynamic content to stabilize
    await page.waitForTimeout(3000);
    
    // Hide dynamic elements that change frequently (timestamps, etc)
    await page.addStyleTag({
      content: `
        /* Hide timestamps and other dynamic content for visual regression testing */
        .timestamp, [data-timestamp], .time, .last-seen, .online-status,
        .realtime-indicator, .live-badge, .websocket-status,
        .session-time, .uptime, .last-updated, .duration,
        time, [datetime] {
          visibility: hidden !important;
        }
        
        /* Stabilize animations and transitions */
        *, *::before, *::after {
          animation-duration: 0.01ms !important;
          animation-delay: -0.01ms !important;
          transition-duration: 0.01ms !important;
          transition-delay: 0.01ms !important;
        }
      `
    });
    
    // Wait additional time after style injection
    await page.waitForTimeout(1000);
  });

  test('Tavern page visual regression', async ({ page }) => {
    // Ensure we're on the tavern page
    await expect(page).toHaveURL('/');
    
    // Wait for content to stabilize (agents, animations, etc.)
    await page.waitForTimeout(4000);
    
    // Take a screenshot of the full page
    await expect(page).toHaveScreenshot('tavern-page.png', {
      fullPage: true,
      animations: 'disabled', // Disable animations for consistent screenshots
      timeout: 10000, // Increased timeout for large pages
    });
  });

  test('Navigation header visual regression', async ({ page }) => {
    // Take a screenshot of just the navigation header
    const navigation = page.locator('header, nav').first();
    
    if (await navigation.isVisible()) {
      await expect(navigation).toHaveScreenshot('navigation-header.png');
    } else {
      // If no specific nav element, take a screenshot of the top portion
      await expect(page).toHaveScreenshot('top-section.png', {
        clip: { x: 0, y: 0, width: 1280, height: 200 }
      });
    }
  });

  test('Agent card visual regression', async ({ page }) => {
    // Wait for agents to potentially load
    await page.waitForTimeout(3000);
    
    // Look for agent cards
    const agentCard = page.locator('.agent-card, .card, [data-testid="agent-card"]').first();
    
    if (await agentCard.isVisible()) {
      await expect(agentCard).toHaveScreenshot('agent-card.png');
    } else {
      // If no agent cards found, take a screenshot of the main content area
      const mainContent = page.locator('main, .content, .grid').first();
      if (await mainContent.isVisible()) {
        await expect(mainContent).toHaveScreenshot('main-content-area.png');
      }
    }
  });

  test('War Room page visual regression', async ({ page }) => {
    // Navigate to War Room
    const warRoomLink = page.locator('a[href="/war-room"], nav a:has-text("War Room")').first();
    
    if (await warRoomLink.isVisible()) {
      await warRoomLink.click();
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);
      
      await expect(page).toHaveScreenshot('war-room-page.png', {
        fullPage: true,
        animations: 'disabled',
      });
    } else {
      // Try direct navigation
      await page.goto('/war-room');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);
      
      await expect(page).toHaveScreenshot('war-room-page-direct.png', {
        fullPage: true,
        animations: 'disabled',
      });
    }
  });

  test('Library page visual regression', async ({ page }) => {
    // Navigate to Library
    const libraryLink = page.locator('a[href="/library"], nav a:has-text("Library")').first();
    
    if (await libraryLink.isVisible()) {
      await libraryLink.click();
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);
      
      await expect(page).toHaveScreenshot('library-page.png', {
        fullPage: true,
        animations: 'disabled',
      });
    } else {
      // Try direct navigation
      await page.goto('/library');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);
      
      await expect(page).toHaveScreenshot('library-page-direct.png', {
        fullPage: true,
        animations: 'disabled',
      });
    }
  });

  test('Guild Chat page visual regression', async ({ page }) => {
    // Navigate to Guild Chat
    await page.goto('/guild-chat');
    await page.waitForLoadState('networkidle');
    
    // Wait longer for chat interface to stabilize
    await page.waitForTimeout(2000);
    
    await expect(page).toHaveScreenshot('guild-chat-page.png', {
      fullPage: true,
      animations: 'disabled',
      timeout: 8000, // Increased timeout
    });
  });

  test('Mobile responsive visual regression', async ({ page }) => {
    // Test mobile viewport (iPhone 12 size)
    await page.setViewportSize({ width: 390, height: 844 });
    
    // Reload page to ensure responsive layout
    await page.reload();
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Take screenshot of mobile layout
    await expect(page).toHaveScreenshot('tavern-mobile.png', {
      fullPage: true,
      animations: 'disabled',
    });
    
    // Test if mobile menu exists and capture it
    const mobileMenuButton = page.locator('[data-testid="mobile-menu"], .mobile-menu-button, button:has-text("Menu")').first();
    
    if (await mobileMenuButton.isVisible()) {
      await mobileMenuButton.click();
      await page.waitForTimeout(500);
      
      await expect(page).toHaveScreenshot('mobile-menu-open.png', {
        fullPage: true,
        animations: 'disabled',
      });
    }
  });

  test('Tablet responsive visual regression', async ({ page }) => {
    // Test tablet viewport (iPad size)
    await page.setViewportSize({ width: 768, height: 1024 });
    
    // Reload page to ensure responsive layout
    await page.reload();
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Take screenshot of tablet layout
    await expect(page).toHaveScreenshot('tavern-tablet.png', {
      fullPage: true,
      animations: 'disabled',
    });
  });

  test('Quest board visual regression', async ({ page }) => {
    // Navigate to quest board
    await page.goto('/quests');
    await page.waitForLoadState('networkidle');
    
    // Wait longer for quest data to potentially load and stabilize
    await page.waitForTimeout(3000);
    
    // Wait for any loading states to complete
    await page.waitForSelector('[data-testid="quest-board"], .quest-board, .main-content', { timeout: 10000 });
    
    await expect(page).toHaveScreenshot('quest-board-page.png', {
      fullPage: true,
      animations: 'disabled',
      timeout: 10000, // Increased timeout
    });
  });

  test('Theme consistency visual regression', async ({ page }) => {
    // Test light theme (if theme switching exists)
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Look for theme toggle
    const themeToggle = page.locator('[data-testid="theme-toggle"], .theme-toggle, button:has-text("theme")').first();
    
    if (await themeToggle.isVisible()) {
      // Take screenshot in current theme
      await expect(page).toHaveScreenshot('theme-default.png', {
        fullPage: true,
        animations: 'disabled',
      });
      
      // Switch theme and take another screenshot
      await themeToggle.click();
      await page.waitForTimeout(500);
      
      await expect(page).toHaveScreenshot('theme-switched.png', {
        fullPage: true,
        animations: 'disabled',
      });
    } else {
      // Just take a screenshot of the current theme
      await expect(page).toHaveScreenshot('theme-current.png', {
        fullPage: true,
        animations: 'disabled',
      });
    }
  });
});