import { Page } from '@playwright/test';

/**
 * Injects a fake JWT into localStorage so the auth guard passes without a real login flow.
 * Matches the keys used in AuthService.
 */
export async function injectAuthToken(page: Page, token: string, user: object) {
  await page.addInitScript(
    ({ token, user }) => {
      localStorage.setItem('tenantly_token', token);
      localStorage.setItem('tenantly_user', JSON.stringify(user));
      localStorage.setItem('tenantly_expires_at', String(Date.now() + 3600 * 1000));
    },
    { token, user }
  );
}

export async function loginViaUI(page: Page, email: string, password: string) {
  await page.goto('/login');
  await page.getByLabel(/email/i).fill(email);
  await page.getByLabel(/password/i).fill(password);
  await page.getByRole('button', { name: /login|sign in/i }).click();
  await page.waitForURL('**/dashboard', { timeout: 10_000 });
}
