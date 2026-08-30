import { test as setup } from '@playwright/test';
import { loginViaUI } from './helpers/auth';

// Backend throttles login to 5 attempts per account per 15 minutes
// (accountLoginLimiter in user_handler.go) to mitigate credential stuffing.
// Logging in once here and sharing the resulting storageState across every
// spec keeps the whole suite well under that limit regardless of how many
// tests or files run.
const authFile = 'e2e/.auth/user.json';

const USERNAME = process.env['E2E_USERNAME'] ?? 'admin';
const PASSWORD = process.env['E2E_PASSWORD'] ?? 'admin123';

setup('authenticate', async ({ page }) => {
  await loginViaUI(page, USERNAME, PASSWORD);
  await page.context().storageState({ path: authFile });
});
