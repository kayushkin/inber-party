import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  timeout: 60 * 1000, // Increased timeout for navigation tests that are having issues
  expect: {
    timeout: 15000, // Increased timeout for visual regression stability and navigation
    // Visual regression test configuration - more tolerant settings for dynamic content and font rendering
    toHaveScreenshot: {
      threshold: 0.15, // Increased threshold to 15% for real-time app with dynamic content
      mode: 'ci',
      animationHandling: 'disabled',
      animations: 'disabled', // Ensure animations are disabled for consistency
      clip: null, // Full page screenshots by default
      fullPage: true,
      omitBackground: false, // Include background for consistent rendering
      scale: 'device', // Use device scale for consistency
    },
    toMatchSnapshot: {
      threshold: 0.15, // Increased threshold to 15% for real-time app with dynamic content
      mode: 'ci',
    },
  },
  fullyParallel: false, // Disable parallel execution for stability
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1, // Moderate retries for stability without excessive wait time
  workers: 1, // Force single worker for consistent test environment
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    // Visual regression specific settings
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    // WebKit disabled for WSL compatibility - uncomment for full cross-browser testing
    // {
    //   name: 'webkit',
    //   use: { ...devices['Desktop Safari'] },
    // },
  ],

  webServer: [
    {
      command: 'cd .. && go run cmd/server/main.go',
      port: 8080,
      reuseExistingServer: true, // Always reuse to avoid restarts between tests
      timeout: 60 * 1000, // Increased timeout for better reliability
      env: {
        'PORT': '8080',
        'NODE_ENV': 'test',
      },
      stdout: 'pipe', // Capture logs for debugging
      stderr: 'pipe',
    },
    {
      command: 'npm run dev',
      port: 5173,
      reuseExistingServer: true, // Always reuse to avoid restarts between tests
      timeout: 45 * 1000,
      stdout: 'pipe', // Capture logs for debugging
      stderr: 'pipe',
    },
  ],
});