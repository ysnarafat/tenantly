import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  forbidOnly: !!process.env['CI'],
  retries: process.env['CI'] ? 2 : 0,
  workers: 1,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:4200',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'setup', testMatch: /auth\.setup\.ts/ },
    {
      // No default storageState here — error-pages.spec.ts relies on starting
      // logged-out. Specs that need a real session opt in via
      // `test.use({ storageState: 'e2e/.auth/user.json' })`.
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
      dependencies: ['setup'],
    },
    {
      name: 'mobile-chromium',
      use: { ...devices['Pixel 7'] },
      dependencies: ['setup'],
    },
    {
      name: 'small-mobile-chromium',
      use: {
        ...devices['Pixel 7'],
        viewport: { width: 360, height: 800 },
      },
      dependencies: ['setup'],
    },
  ],
});
