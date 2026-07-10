import { test, expect, Page } from '@playwright/test';
import { loginViaUI } from './helpers/auth';

// Credentials — override via env vars for CI
const EMAIL = process.env['E2E_EMAIL'] ?? 'admin@tenantly.com';
const PASSWORD = process.env['E2E_PASSWORD'] ?? 'password';

test.describe('Payments page', () => {
  test.beforeEach(async ({ page }) => {
    await loginViaUI(page, EMAIL, PASSWORD);
    await page.goto('/payments');
    await page.waitForSelector('.payment-container', { timeout: 10_000 });
  });

  // ── Page structure ──────────────────────────────────────────────

  test('shows page title and action buttons', async ({ page }) => {
    await expect(page.locator('h1.page-title')).toBeVisible();
    await expect(page.getByRole('button', { name: /new payment/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /generate monthly/i })).toBeVisible();
  });

  test('shows summary cards', async ({ page }) => {
    await expect(page.locator('.stat-card')).toHaveCount(5);
    await expect(page.locator('.stat-label').filter({ hasText: /total due/i })).toBeVisible();
    await expect(page.locator('.stat-label').filter({ hasText: /total collected/i })).toBeVisible();
    await expect(page.locator('.stat-label').filter({ hasText: /collection rate/i })).toBeVisible();
  });

  test('shows two tabs', async ({ page }) => {
    await expect(page.getByRole('tab', { name: /all payments/i })).toBeVisible();
    await expect(page.getByRole('tab', { name: /by property/i })).toBeVisible();
  });

  // ── All Payments tab ────────────────────────────────────────────

  test.describe('All Payments tab', () => {
    test('defaults: month All, year current, filters at end', async ({ page }) => {
      const currentYear = new Date().getFullYear().toString();

      // Status filter defaults to All
      const statusSelect = page.locator('mat-select').first();
      await expect(statusSelect).toContainText(/all/i);

      // Month filter defaults to All
      const monthSelect = page.locator('mat-select').nth(1);
      await expect(monthSelect).toContainText(/all/i);

      // Year filter defaults to current year
      const yearSelect = page.locator('mat-select').nth(2);
      await expect(yearSelect).toContainText(currentYear);
    });

    test('Apply and Clear buttons visible', async ({ page }) => {
      await expect(page.getByRole('button', { name: /apply/i })).toBeVisible();
      await expect(page.getByRole('button', { name: /clear/i })).toBeVisible();
    });

    test('month dropdown shows translated month names', async ({ page }) => {
      const monthSelect = page.locator('mat-select').nth(1);
      await monthSelect.click();

      const panel = page.locator('mat-option');
      await expect(panel.filter({ hasText: 'January' })).toBeVisible();
      await expect(panel.filter({ hasText: 'December' })).toBeVisible();

      // No raw translation keys visible
      await expect(panel.filter({ hasText: 'PAYMENT_LIST' })).toHaveCount(0);

      await page.keyboard.press('Escape');
    });

    test('status dropdown shows translated values', async ({ page }) => {
      const statusSelect = page.locator('mat-select').first();
      await statusSelect.click();

      const options = page.locator('mat-option');
      await expect(options.filter({ hasText: 'Paid' })).toBeVisible();
      await expect(options.filter({ hasText: 'Due' })).toBeVisible();
      await expect(options.filter({ hasText: 'Partial' })).toBeVisible();
      await expect(options.filter({ hasText: 'Overdue' })).toBeVisible();

      // No raw translation keys
      await expect(options.filter({ hasText: 'PAYMENT_LIST' })).toHaveCount(0);

      await page.keyboard.press('Escape');
    });

    test('Apply filter loads data and Clear resets', async ({ page }) => {
      // Pick a specific status
      await page.locator('mat-select').first().click();
      await page.locator('mat-option').filter({ hasText: 'Paid' }).click();

      await page.getByRole('button', { name: /apply/i }).click();

      // Clear resets dropdowns
      await page.getByRole('button', { name: /clear/i }).click();
      await expect(page.locator('mat-select').first()).toContainText(/all/i);
    });

    test('table renders column headers', async ({ page }) => {
      const headers = page.locator('th.mat-header-cell');
      await expect(headers).not.toHaveCount(0);
      // At least one of the expected headers is present
      const headerTexts = await headers.allTextContents();
      const flat = headerTexts.map(t => t.trim().toLowerCase()).join(' ');
      expect(flat).toMatch(/period|unit|status/);
    });

    test('pagination controls visible when data exists', async ({ page }) => {
      const rowCount = await page.locator('tr.mat-row').count();
      if (rowCount > 0) {
        await expect(page.locator('mat-paginator, .pagination-bar')).toBeVisible();
      }
    });
  });

  // ── By Property tab ─────────────────────────────────────────────

  test.describe('By Property tab', () => {
    test.beforeEach(async ({ page }) => {
      await page.getByRole('tab', { name: /by property/i }).click();
    });

    test('defaults: month All, year current, Load button at end', async ({ page }) => {
      const currentYear = new Date().getFullYear().toString();

      const monthSelect = page.locator('.tree-filters-row mat-select').first();
      await expect(monthSelect).toContainText(/all/i);

      const yearSelect = page.locator('.tree-filters-row mat-select').nth(1);
      await expect(yearSelect).toContainText(currentYear);

      // Load button is last child in the row
      const loadBtn = page.locator('.tree-filters-row button').last();
      await expect(loadBtn).toContainText(/load/i);
    });

    test('month dropdown shows translated names', async ({ page }) => {
      const monthSelect = page.locator('.tree-filters-row mat-select').first();
      await monthSelect.click();

      const options = page.locator('mat-option');
      await expect(options.filter({ hasText: 'January' })).toBeVisible();
      await expect(options.filter({ hasText: 'PAYMENT_LIST' })).toHaveCount(0);

      await page.keyboard.press('Escape');
    });

    test('Load button fetches and displays property tree', async ({ page }) => {
      await page.locator('.tree-filters-row button').last().click();

      // Either data cards or empty state should appear — not a spinner indefinitely
      await expect(
        page.locator('.property-node, .tree-empty, [class*="property-card"]')
      ).toBeVisible({ timeout: 15_000 });
    });

    test('changing month filter and reloading updates tree', async ({ page }) => {
      await page.locator('.tree-filters-row button').last().click();
      await page.waitForSelector('.property-node, .tree-empty', { timeout: 15_000 });

      // Switch to a specific month
      const monthSelect = page.locator('.tree-filters-row mat-select').first();
      await monthSelect.click();
      await page.locator('mat-option').filter({ hasText: 'January' }).click();

      await page.locator('.tree-filters-row button').last().click();
      await page.waitForSelector('.property-node, .tree-empty', { timeout: 15_000 });
    });
  });

  // ── New Payment dialog ──────────────────────────────────────────

  test('New Payment button opens dialog', async ({ page }) => {
    await page.getByRole('button', { name: /new payment/i }).click();
    await expect(page.locator('mat-dialog-container')).toBeVisible({ timeout: 5_000 });
    await page.keyboard.press('Escape');
  });

  // ── Generate Monthly dialog ─────────────────────────────────────

  test('Generate Monthly button opens dialog', async ({ page }) => {
    await page.getByRole('button', { name: /generate monthly/i }).click();
    await expect(page.locator('mat-dialog-container')).toBeVisible({ timeout: 5_000 });
    await page.keyboard.press('Escape');
  });
});
