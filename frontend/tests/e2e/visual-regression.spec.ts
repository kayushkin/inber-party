import { test, expect, type Page } from '@playwright/test';

// Helper function to wait for visual stability
async function waitForVisualStability(page: Page, timeout = 15000) {
  // Wait for network to be idle
  await page.waitForLoadState('networkidle');
  
  // Wait for all fonts to load
  await page.waitForFunction(() => document.fonts.ready);
  
  // Wait for any remaining lazy loading or dynamic content
  await page.waitForTimeout(3000);
  
  // Wait for no visible loading indicators
  await page.waitForFunction(() => {
    const loadingElements = document.querySelectorAll('.loading, .spinner, .skeleton, [data-loading="true"]');
    return loadingElements.length === 0 || 
           Array.from(loadingElements).every(el => !el.offsetParent);
  }, { timeout });
  
  // Additional stability wait
  await page.waitForTimeout(1000);
}

test.describe('Visual Regression Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for it to load
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Mock Date.now to freeze timestamps for consistent screenshots
    await page.addInitScript(() => {
      const now = new Date('2026-03-20T05:42:00.000Z').getTime();
      Date.now = () => now;
      Date.prototype.getTime = function() { return now; };
    });
    
    // Wait longer for WebSocket connections and dynamic content to stabilize
    await page.waitForTimeout(5000);
    
    // Enhanced CSS to hide ALL dynamic content that could cause visual differences
    await page.addStyleTag({
      content: `
        /* Hide timestamps and other dynamic content for visual regression testing */
        .timestamp, [data-timestamp], .time, .last-seen, .online-status,
        .realtime-indicator, .live-badge, .websocket-status,
        .session-time, .uptime, .last-updated, .duration,
        time, [datetime], .status-text, .connection-status,
        .heartbeat-indicator, .pulse, .blink, .typing-indicator,
        .load-time, .render-time, .performance-stats,
        .websocket-debug, .debug-panel {
          visibility: hidden !important;
          opacity: 0 !important;
        }
        
        /* Replace dynamic text content with fixed placeholders */
        .agent-status::after { content: "idle" !important; }
        .connection-count::after { content: "0" !important; }
        
        /* Completely disable all animations, transitions, and transforms */
        *, *::before, *::after {
          animation-duration: 0ms !important;
          animation-delay: 0ms !important;
          transition-duration: 0ms !important;
          transition-delay: 0ms !important;
          transform: none !important;
          transition-property: none !important;
          animation-name: none !important;
        }
        
        /* Force consistent font rendering */
        * {
          font-feature-settings: normal !important;
          text-rendering: geometricPrecision !important;
          -webkit-font-smoothing: antialiased !important;
          -moz-osx-font-smoothing: grayscale !important;
        }
        
        /* Hide WebSocket connection indicators and live data */
        .ws-status, .live-data, .realtime-counter,
        .agent-activity, .quest-progress, .live-stats {
          display: none !important;
        }
      `
    });
    
    // Wait for styles to apply and content to stabilize
    await page.waitForTimeout(2000);
    
    // Wait for any lazy-loaded content
    await page.waitForFunction(() => {
      return document.readyState === 'complete' && 
             !document.querySelector('.loading, .spinner, [data-loading="true"]');
    }, { timeout: 10000 });
  });

  test('Tavern page visual regression', async ({ page }) => {
    // Ensure we're on the tavern page
    await expect(page).toHaveURL('/');
    
    // Wait for complete visual stability
    await waitForVisualStability(page);
    
    // Take a screenshot of the full page
    await expect(page).toHaveScreenshot('tavern-page.png', {
      fullPage: true,
      animations: 'disabled',
      timeout: 15000,
      maxDiffPixels: 5000, // Allow up to 5000 pixel differences for dynamic content
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