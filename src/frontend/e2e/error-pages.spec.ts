import { test, expect, Page } from '@playwright/test';
import { injectAuthToken } from './helpers/auth';

// These tests exercise the 404/401 error pages and the route guards that
// redirect to them — no backend calls are needed, so auth state is faked via
// localStorage (matching the shape AuthGuard/the auth reducer read) rather
// than going through a real login flow.
function fakeUser(role: string) {
  return {
    id: 1,
    username: `${role.toLowerCase()}-user`,
    email: `${role.toLowerCase()}@tenantly.test`,
    role,
    organization_id: 1,
    active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };
}

async function loginAs(page: Page, role: string) {
  await injectAuthToken(page, 'fake-jwt-token', fakeUser(role));
}

test.describe('404 — Not Found', () => {
  test('unauthenticated user hitting an unknown URL sees the 404 board', async ({ page }) => {
    await page.goto('/this-route-does-not-exist');

    await expect(page.locator('.code')).toHaveText('404');
    await expect(page.getByText(/doesn't exist/i)).toBeVisible();
    // Not logged in — the board's home link should point at login, not dashboard.
    await expect(page.getByRole('link', { name: /login/i })).toBeVisible();
  });

  test('authenticated user hitting an unknown URL sees the 404 board with a dashboard link', async ({
    page,
  }) => {
    await loginAs(page, 'Admin');
    await page.goto('/this-route-does-not-exist');

    await expect(page.locator('.code')).toHaveText('404');
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible();
  });

  test('unknown nested path under a valid top-level route also falls through to 404', async ({
    page,
  }) => {
    await loginAs(page, 'SUPER_ADMIN');
    await page.goto('/admin/this-does-not-exist');

    await expect(page.locator('.code')).toHaveText('404');
  });

  test('legacy redirects still work and do not land on the 404 page', async ({ page }) => {
    await loginAs(page, 'Admin');

    await page.goto('/shops');
    await page.waitForURL('**/properties');
    await expect(page.locator('.code')).toHaveCount(0);

    await page.goto('/attachments');
    await page.waitForURL('**/documents');
    await expect(page.locator('.code')).toHaveCount(0);
  });

  test('"Go back" returns to the previous page in history', async ({ page }) => {
    await loginAs(page, 'Admin');
    await page.goto('/dashboard');
    await page.goto('/this-route-does-not-exist');

    await page.getByRole('button', { name: /go back/i }).click();
    await page.waitForURL('**/dashboard');
  });

  test('the app shell (sidenav/toolbar) is not rendered on the 404 page', async ({ page }) => {
    await loginAs(page, 'Admin');
    await page.goto('/this-route-does-not-exist');

    await expect(page.locator('mat-sidenav')).toHaveCount(0);
  });
});

test.describe('401 — Unauthorized', () => {
  test('unauthenticated visit to /401 renders without redirect looping', async ({ page }) => {
    await page.goto('/401');

    await expect(page.locator('.code')).toHaveText('401');
    await expect(page.getByRole('link', { name: /login/i })).toBeVisible();
  });

  test('Accountant typing an admin URL directly is redirected to 401', async ({ page }) => {
    await loginAs(page, 'Accountant');
    await page.goto('/admin/organizations');

    await page.waitForURL('**/401');
    await expect(page.locator('.code')).toHaveText('401');
    await expect(page.getByText(/don't have permission/i)).toBeVisible();
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible();
  });

  test('PropertyManager typing the user-management URL directly is redirected to 401', async ({
    page,
  }) => {
    await loginAs(page, 'PropertyManager');
    await page.goto('/users');

    await page.waitForURL('**/401');
    await expect(page.locator('.code')).toHaveText('401');
  });

  test('SUPER_ADMIN typing an org-data URL directly is redirected to 401', async ({ page }) => {
    // SUPER_ADMIN is platform-level only — no org-scoped data access.
    await loginAs(page, 'SUPER_ADMIN');
    await page.goto('/properties');

    await page.waitForURL('**/401');
    await expect(page.locator('.code')).toHaveText('401');
  });

  test('Accountant typing the properties URL directly is redirected to 401', async ({ page }) => {
    // Accountant has VIEW_PROPERTIES but not MANAGE_PROPERTIES, which is what
    // both the sidenav link and the route guard key off of.
    await loginAs(page, 'Accountant');
    await page.goto('/properties');

    await page.waitForURL('**/401');
  });

  test('ORG_ADMIN passes the parent admin guard but is denied on the super-admin-only child route', async ({
    page,
  }) => {
    await loginAs(page, 'ORG_ADMIN');
    await page.goto('/admin/organizations');

    await page.waitForURL('**/401');
  });

  test('a role with the right permission is NOT redirected to 401', async ({ page }) => {
    await loginAs(page, 'PropertyManager');
    await page.goto('/properties');

    await expect(page).not.toHaveURL(/\/401$/);
    await expect(page.locator('.code')).toHaveCount(0);
  });

  test('"Go to dashboard" navigates away from the 401 page', async ({ page }) => {
    await loginAs(page, 'Accountant');
    await page.goto('/admin/organizations');
    await page.waitForURL('**/401');

    await page.getByRole('link', { name: /dashboard/i }).click();
    await page.waitForURL('**/dashboard');
  });

  test('the app shell (sidenav/toolbar) is not rendered on the 401 page', async ({ page }) => {
    await loginAs(page, 'Accountant');
    await page.goto('/admin/organizations');
    await page.waitForURL('**/401');

    await expect(page.locator('mat-sidenav')).toHaveCount(0);
  });
});
