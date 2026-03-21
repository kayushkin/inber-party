import { test, expect, type Page } from '@playwright/test';

// Enhanced helper function to wait for complete visual stability
async function waitForVisualStability(page: Page, timeout = 20000) {
  // Force consistent viewport size
  await page.setViewportSize({ width: 1280, height: 720 });
  
  // Inject comprehensive stability CSS before any content loads
  await page.addInitScript(() => {
    // Mock all time-related functions for consistency
    const fixedTime = new Date('2026-03-20T12:00:00.000Z').getTime();
    Date.now = () => fixedTime;
    Date.prototype.getTime = function() { return fixedTime; };
    
    // Override performance.now() for consistent timing
    performance.now = () => 0;
    
    // Create a stable mock WebSocket that doesn't trigger reconnection loops
    window.WebSocket = class StableMockWebSocket {
      constructor(url) {
        this.url = url;
        this.readyState = 1; // OPEN state - stable connection
        this.onopen = null;
        this.onclose = null;
        this.onmessage = null;
        this.onerror = null;
        
        // Immediately call onopen to simulate successful connection
        setTimeout(() => {
          if (this.onopen) this.onopen({});
        }, 10);
      }
      
      close() {
        this.readyState = 3; // CLOSED
        // Don't trigger onclose to prevent reconnection attempts
      }
      
      send() {
        // Silently accept messages without processing
      }
    };
    
    // Also prevent the optimized WebSocket manager from attempting any real connections
    window.__VISUAL_REGRESSION_TEST__ = true;
  });
  
  // Wait for network to be idle
  await page.waitForLoadState('networkidle');
  
  // Wait for all fonts to load
  await page.waitForFunction(() => document.fonts.ready, { timeout });
  
  // Apply comprehensive CSS to eliminate ALL sources of visual variance
  await page.addStyleTag({
    content: `
      /* Force consistent font rendering */
      * {
        font-feature-settings: "kern" 0 !important;
        text-rendering: geometricPrecision !important;
        -webkit-font-smoothing: antialiased !important;
        -moz-osx-font-smoothing: grayscale !important;
        font-variant-ligatures: none !important;
        font-variant-numeric: normal !important;
        -webkit-text-stroke: 0 !important;
      }
      
      /* Disable ALL animations, transitions, and transforms */
      *, *::before, *::after {
        animation: none !important;
        transition: none !important;
        transform: none !important;
        animation-duration: 0s !important;
        animation-delay: 0s !important;
        transition-duration: 0s !important;
        transition-delay: 0s !important;
        animation-fill-mode: none !important;
        transform-origin: initial !important;
      }
      
      /* Hide ALL dynamic and time-sensitive content */
      .timestamp, [data-timestamp], .time, .last-seen, .online-status,
      .realtime-indicator, .live-badge, .websocket-status, .ws-status,
      .session-time, .uptime, .last-updated, .duration, .load-time,
      time, [datetime], .status-text, .connection-status, .debug-info,
      .heartbeat-indicator, .pulse, .blink, .typing-indicator,
      .render-time, .performance-stats, .websocket-debug, .debug-panel,
      .live-data, .realtime-counter, .agent-activity, .quest-progress,
      .live-stats, .connection-count, .active-connections, .ping-time,
      .version-info, .build-info, .commit-hash, .deployment-time {
        visibility: hidden !important;
        opacity: 0 !important;
        display: none !important;
      }
      
      /* Force static placeholder text for dynamic content areas */
      .agent-status::after { content: "idle" !important; }
      .connection-count::after { content: "0" !important; }
      .current-time::after { content: "12:00 PM" !important; }
      
      /* Ensure consistent box sizing and positioning */
      * {
        box-sizing: border-box !important;
        backface-visibility: hidden !important;
        perspective: none !important;
      }
      
      /* Disable hover effects that might interfere */
      *:hover {
        transform: none !important;
        transition: none !important;
      }
      
      /* Force consistent scrollbar behavior */
      ::-webkit-scrollbar {
        width: 16px !important;
      }
      
      /* Hide loading indicators and skeletons */
      .loading, .spinner, .skeleton, [data-loading="true"],
      .loading-skeleton, .shimmer, .placeholder-loading {
        display: none !important;
      }
    `
  });
  
  // Wait for dynamic content to finish loading
  await page.waitForTimeout(5000);
  
  // Wait for no visible loading indicators
  await page.waitForFunction(() => {
    const loadingElements = document.querySelectorAll('.loading, .spinner, .skeleton, [data-loading="true"], .loading-skeleton');
    return loadingElements.length === 0 || 
           Array.from(loadingElements).every(el => !el.offsetParent);
  }, { timeout: timeout / 2 });
  
  // Additional stability wait to ensure all rendering is complete
  await page.waitForTimeout(2000);
  
  // Final check that page is fully rendered and stable
  await page.waitForFunction(() => {
    return document.readyState === 'complete' && 
           !document.querySelector('.loading, .spinner, [data-loading="true"]') &&
           document.body.offsetHeight > 0;
  }, { timeout: 5000 });
}

test.describe('Visual Regression Tests', () => {
  // Set global test timeout for visual regression tests
  test.setTimeout(60000); // 60 seconds
  
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and wait for comprehensive stability
    await page.goto('/');
    await waitForVisualStability(page, 30000);
  });

  test('Tavern page visual regression', async ({ page }) => {
    // Ensure we're on the tavern page
    await expect(page).toHaveURL('/');
    
    // Additional stability wait for this specific test
    await waitForVisualStability(page);
    
    // Take a screenshot of the full page with generous tolerance
    await expect(page).toHaveScreenshot('tavern-page.png', {
      fullPage: true,
      animations: 'disabled',
      timeout: 30000,
      threshold: 0.15, // Allow 15% pixel difference
      maxDiffPixels: 25000, // Allow up to 25,000 pixel differences
    });
  });

  test('Navigation header visual regression', async ({ page }) => {
    // Wait for navigation to be fully rendered
    await page.waitForSelector('header, nav', { timeout: 10000 });
    await page.waitForTimeout(1000);
    
    // Take a screenshot of just the navigation header
    const navigation = page.locator('header, nav').first();
    
    if (await navigation.isVisible()) {
      await expect(navigation).toHaveScreenshot('navigation-header.png', {
        animations: 'disabled',
        threshold: 0.1, // Allow 10% pixel difference
        maxDiffPixels: 5000,
      });
    } else {
      // If no specific nav element, take a screenshot of the top portion
      await expect(page).toHaveScreenshot('top-section.png', {
        clip: { x: 0, y: 0, width: 1280, height: 200 },
        threshold: 0.1,
        maxDiffPixels: 3000,
      });
    }
  });

  test('Agent card visual regression', async ({ page }) => {
    // Wait for agents to potentially load with extended timeout
    await page.waitForTimeout(5000);
    
    // Look for agent cards with multiple selectors
    const agentCard = page.locator('.agent-card, .card, [data-testid="agent-card"], .agent-item').first();
    
    if (await agentCard.isVisible()) {
      await expect(agentCard).toHaveScreenshot('agent-card.png', {
        animations: 'disabled',
        threshold: 0.1,
        maxDiffPixels: 3000,
      });
    } else {
      // If no agent cards found, take a screenshot of the main content area
      const mainContent = page.locator('main, .content, .grid, .main-content').first();
      if (await mainContent.isVisible()) {
        await expect(mainContent).toHaveScreenshot('main-content-area.png', {
          animations: 'disabled',
          threshold: 0.15,
          maxDiffPixels: 10000,
        });
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
    
    // Wait for comprehensive stability
    await waitForVisualStability(page);
    
    // Wait for any loading states to complete
    try {
      await page.waitForSelector('[data-testid="quest-board"], .quest-board, .main-content', { timeout: 15000 });
    } catch (error) {
      // If specific quest board elements aren't found, continue with general page
      console.log('Quest board specific elements not found, proceeding with general page screenshot:', error.message);
    }
    
    await expect(page).toHaveScreenshot('quest-board-page.png', {
      fullPage: true,
      animations: 'disabled',
      timeout: 30000,
      threshold: 0.15, // Allow 15% pixel difference for dynamic quest content
      maxDiffPixels: 25000,
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