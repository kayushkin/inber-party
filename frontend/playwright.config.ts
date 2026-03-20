import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  timeout: 30 * 1000,
  expect: {
    timeout: 5000,
    // Visual regression test configuration - more tolerant settings for font rendering variations
    toHaveScreenshot: {
      threshold: 0.05, // Threshold as ratio (0-1), allowing for small font rendering differences
      mode: 'ci',
      animationHandling: 'disabled',
    },
    toMatchSnapshot: {
      threshold: 0.05, // Threshold as ratio (0-1), allowing for small font rendering differences
    },
  },
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
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
      reuseExistingServer: !process.env.CI,
      timeout: 30 * 1000,
      env: {
        'PORT': '8080',
        'NODE_ENV': 'test',
      },
    },
    {
      command: 'npm run dev',
      port: 5173,
      reuseExistingServer: !process.env.CI,
      timeout: 30 * 1000,
    },
  ],
});