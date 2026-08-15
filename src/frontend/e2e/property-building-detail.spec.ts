import { test, expect } from '@playwright/test';

// Reuses the session from the "setup" project (see auth.setup.ts) instead of
// logging in per-test — the backend throttles login to 5 attempts per
// account per 15 minutes (accountLoginLimiter in user_handler.go).
test.use({ storageState: 'e2e/.auth/user.json' });

// These tests target the seeded dev property (id=1, "Default Property") and
// its one seeded building (id=1, "Main Building") — the only property/building
// rows that reliably exist in this environment. Both currently have zero
// units, so unit-list assertions are written against the empty state.
const PROPERTY_ID = 1;
const BUILDING_ID = 1;

test.describe('Property detail page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(`/properties/${PROPERTY_ID}`);
    await page.waitForSelector('.detail-header', { timeout: 10_000 });
    // Buildings load asynchronously after the property itself resolves — wait
    // for that second fetch to settle before asserting on the list.
    await page.waitForSelector('.buildings-list, .empty-tab', { timeout: 10_000 });
  });

  test('shows property header with name, status badge, and meta line', async ({ page }) => {
    await expect(page.locator('h1')).not.toBeEmpty();
    await expect(page.locator('.status-badge')).toBeVisible();
    await expect(page.locator('.property-meta')).toBeVisible();
  });

  test('shows back link to properties list', async ({ page }) => {
    const back = page.locator('a.back-link');
    await expect(back).toBeVisible();
    await back.click();
    await page.waitForURL('**/properties');
  });

  test('shows four stat cards', async ({ page }) => {
    await expect(page.locator('.stats-grid .stat-card')).toHaveCount(4);
  });

  test('view full financial report link points at reports', async ({ page }) => {
    const link = page.getByRole('link', { name: /view full financial report/i });
    await expect(link).toBeVisible();
    await expect(link).toHaveAttribute('href', '/reports');
  });

  test('Edit button opens the property form dialog pre-filled', async ({ page }) => {
    await page.getByRole('button', { name: /edit/i }).click();
    const dialog = page.locator('mat-dialog-container');
    await expect(dialog).toBeVisible({ timeout: 5_000 });

    // Pre-filled with the real property, not blank
    await expect(dialog.locator('input[formcontrolname="property_name"]')).not.toHaveValue('');

    await page.getByRole('button', { name: /cancel/i }).click();
    await expect(dialog).toBeHidden();
  });

  test('Delete button opens the confirm-delete dialog, cancel leaves the property intact', async ({
    page,
  }) => {
    await page.getByRole('button', { name: /delete/i }).click();
    const dialog = page.locator('mat-dialog-container');
    await expect(dialog).toBeVisible({ timeout: 5_000 });
    await expect(dialog.getByRole('button', { name: /delete/i })).toBeDisabled();

    await dialog.getByRole('button', { name: /cancel/i }).click();
    await expect(dialog).toBeHidden();

    // Still on the same property afterwards — nothing was deleted
    await expect(page).toHaveURL(new RegExp(`/properties/${PROPERTY_ID}$`));
  });

  test('Add Building button opens the building form dialog, cancel adds nothing', async ({
    page,
  }) => {
    const buildingCountBefore = await page.locator('.buildings-list .building-card').count();

    await page.getByRole('button', { name: /add building/i }).click();
    const dialog = page.locator('mat-dialog-container');
    await expect(dialog).toBeVisible({ timeout: 5_000 });
    await expect(dialog.locator('input[formcontrolname="building_name"]')).toHaveValue('');

    await dialog.getByRole('button', { name: /cancel/i }).click();
    await expect(dialog).toBeHidden();

    await expect(page.locator('.buildings-list .building-card')).toHaveCount(buildingCountBefore);
  });

  test('Buildings section lists the seeded building and expands to its units tree', async ({
    page,
  }) => {
    const card = page.locator('.building-card').first();
    await expect(card).toBeVisible();
    await expect(card.locator('.building-name')).not.toBeEmpty();

    const header = card.locator('.building-card-header');
    await header.click();
    await expect(card.locator('.units-tree')).toBeVisible();

    // Zero units seeded — tree shows the empty state, not a crash
    const unitNodes = card.locator('.unit-node');
    const emptyState = card.locator('.empty-units');
    await expect(unitNodes.or(emptyState).first()).toBeVisible({ timeout: 5_000 });

    // Collapses back on second click
    await header.click();
    await expect(card.locator('.units-tree')).toBeHidden();
  });

  test('"View full details" navigates to the building detail page', async ({ page }) => {
    const card = page.locator('.building-card').first();
    // The card header div is itself role="button" (click-to-expand) and has
    // no accessible name of its own, so getByRole('button') would resolve
    // both it and the icon button — a plain <button> tag selector is
    // unambiguous since the header div isn't one.
    await card.locator('.building-card-header button').click();
    await page.waitForURL(`**/properties/${PROPERTY_ID}/buildings/**`);
    await expect(page.locator('.breadcrumb')).toBeVisible();
  });
});

test.describe('Building detail page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(`/properties/${PROPERTY_ID}/buildings/${BUILDING_ID}`);
    await page.waitForSelector('.detail-header', { timeout: 10_000 });
    // Units load asynchronously after the building itself resolves — wait for
    // that second fetch to settle before asserting on the list.
    await page.waitForSelector('.units-grid, .empty-tab', { timeout: 10_000 });
  });

  test('shows breadcrumb with property name and building name', async ({ page }) => {
    const breadcrumb = page.locator('.breadcrumb');
    await expect(breadcrumb).toBeVisible();
    await expect(breadcrumb.locator('.breadcrumb-link')).not.toBeEmpty();
    await expect(breadcrumb.locator('.breadcrumb-current')).not.toBeEmpty();
  });

  test('breadcrumb link navigates back to the parent property', async ({ page }) => {
    await page.locator('.breadcrumb-link').click();
    await page.waitForURL(`**/properties/${PROPERTY_ID}`);
  });

  test('shows building header with name, status badge, and meta line', async ({ page }) => {
    await expect(page.locator('h1')).not.toBeEmpty();
    await expect(page.locator('.status-badge')).toBeVisible();
    await expect(page.locator('.building-meta')).toBeVisible();
  });

  test('shows three stat cards', async ({ page }) => {
    await expect(page.locator('.stats-grid .stat-card')).toHaveCount(3);
  });

  test('Edit button opens the building form dialog pre-filled', async ({ page }) => {
    await page.getByRole('button', { name: /edit/i }).click();
    const dialog = page.locator('mat-dialog-container');
    await expect(dialog).toBeVisible({ timeout: 5_000 });
    await expect(dialog.locator('input[formcontrolname="building_name"]')).not.toHaveValue('');

    await dialog.getByRole('button', { name: /cancel/i }).click();
    await expect(dialog).toBeHidden();
  });

  test('Delete button opens the confirm-delete dialog, cancel leaves the building intact', async ({
    page,
  }) => {
    await page.getByRole('button', { name: /delete/i }).click();
    const dialog = page.locator('mat-dialog-container');
    await expect(dialog).toBeVisible({ timeout: 5_000 });

    await dialog.getByRole('button', { name: /cancel/i }).click();
    await expect(dialog).toBeHidden();

    await expect(page).toHaveURL(
      new RegExp(`/properties/${PROPERTY_ID}/buildings/${BUILDING_ID}$`)
    );
  });

  test('Add Unit button opens the unit form dialog, cancel adds nothing', async ({ page }) => {
    await page.getByRole('button', { name: /add unit/i }).click();
    const dialog = page.locator('mat-dialog-container');
    await expect(dialog).toBeVisible({ timeout: 5_000 });
    await expect(dialog.locator('input[formcontrolname="unit_number"]')).toHaveValue('');

    await dialog.getByRole('button', { name: /cancel/i }).click();
    await expect(dialog).toBeHidden();
  });

  test('Units section shows empty state when the building has no units', async ({ page }) => {
    const grid = page.locator('.units-grid');
    const empty = page.locator('.empty-tab');
    await expect(grid.or(empty).first()).toBeVisible({ timeout: 5_000 });
  });
});
