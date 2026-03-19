import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  timeout: 30 * 1000,
  expect: {
    timeout: 5000,
    // Visual regression test configuration - more tolerant settings
    toHaveScreenshot: {
      threshold: 0.3, // Increased threshold for viewport differences
      mode: 'ci',
    },
    toMatchSnapshot: {
      threshold: 0.3, // Increased threshold for better stability
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