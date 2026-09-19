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

export async function loginViaUI(page: Page, username: string, password: string) {
  await page.goto('/login');
  // The login field asks for a username (LOGIN.FIELDS.USERNAME), and the
  // backend looks the account up by username (UserService.Login →
  // GetByUsername) — not by email, even though some seed accounts also have
  // an email address that looks like a login.
  await page.getByLabel(/username/i).fill(username);
  await page.getByLabel(/password/i).fill(password);
  await page.getByRole('button', { name: /login|sign in/i }).click();
  await page.waitForURL('**/dashboard', { timeout: 10_000 });
}
