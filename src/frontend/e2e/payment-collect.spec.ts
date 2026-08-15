import { test, expect, Page } from '@playwright/test';

// Reuses the session from the "setup" project (see auth.setup.ts) instead of
// logging in per-test — the backend throttles login to 5 attempts per
// account per 15 minutes (accountLoginLimiter in user_handler.go).
test.use({ storageState: 'e2e/.auth/user.json' });

// The dev database never has real "Due"/"Partial"/"Overdue" payment rows (see
// property-building-detail.spec.ts's note on the seed data), so exercising the
// populated table, save, and skip flows requires mocking the payments API
// rather than depending on backend fixtures that don't exist here.
const API_PAYMENTS = '**/api/v1/payments**';

const now = new Date();
const CURRENT_MONTH = now.getMonth() + 1;
const CURRENT_YEAR = now.getFullYear();

function makePayment(id: number, overrides: Record<string, unknown> = {}) {
  return {
    id,
    unit_id: id,
    tenant_id: id,
    building_id: 1,
    property_id: 1,
    month: CURRENT_MONTH,
    year: CURRENT_YEAR,
    amount_due: 15000,
    amount_paid: 0,
    status: 'Due',
    payment_method: undefined,
    payment_date: undefined,
    due_date: `${CURRENT_YEAR}-${String(CURRENT_MONTH).padStart(2, '0')}-05`,
    created_at: new Date(0).toISOString(),
    updated_at: new Date(0).toISOString(),
    property_name: 'Default Property',
    building_name: 'Main Building',
    building_code: 'BLD-001',
    unit_number: `10${id}`,
    unit_type: 'Apartment',
    tenant_name: `Tenant ${id}`,
    ...overrides,
  };
}

// Mocks both the three parallel GET calls PaymentService.getAllDuePayments
// makes (one per status: Due, Partial, Overdue — splitting `rows` by their
// status) and, optionally, the PUT calls saveAll() makes per row. A single
// route registration is used for both verbs so tests that need to change the
// PUT behavior mid-test (re-registering the route) don't end up stacking
// handlers where a fallback could slip through to the real backend.
async function mockPaymentsApi(
  page: Page,
  rows: ReturnType<typeof makePayment>[],
  onUpdate?: (id: number) => { status: number; body?: unknown }
) {
  await page.route(API_PAYMENTS, async (route) => {
    const request = route.request();

    if (request.method() === 'PUT') {
      const id = Number(new URL(request.url()).pathname.split('/').pop());
      const { status, body } = onUpdate?.(id) ?? { status: 200 };
      if (status >= 400) {
        await route.fulfill({ status, json: body ?? { error: 'Failed to save' } });
      } else {
        await route.fulfill({ json: body ?? { id, status: 'Paid' } });
      }
      return;
    }

    const url = new URL(request.url());
    const status = url.searchParams.get('status');
    const matching = rows.filter((r) => r.status === status);
    await route.fulfill({
      json: {
        payments: matching,
        total: matching.length,
        page: 1,
        page_size: matching.length || 1,
        total_pages: 1,
      },
    });
  });
}

test.describe('Payment collect page — empty state', () => {
  test('shows empty state with a working Generate link when there is nothing due', async ({
    page,
  }) => {
    await mockPaymentsApi(page, []);
    await page.goto('/payments/collect');
    await page.waitForSelector('.payment-collect-container');

    await expect(page.locator('.empty-state')).toBeVisible();
    await expect(page.locator('.collect-table')).toHaveCount(0);

    await page.locator('.empty-state button').click();
    await expect(page.locator('mat-dialog-container')).toBeVisible({ timeout: 5_000 });
    await page.keyboard.press('Escape');
  });

  test('shows month/year selectors and a Generate button in the period card', async ({ page }) => {
    await mockPaymentsApi(page, []);
    await page.goto('/payments/collect');
    await page.waitForSelector('.payment-collect-container');

    await expect(page.locator('.period-field')).toHaveCount(2);
    // Scoped to the period row — the empty state below also renders its own
    // "Generate" button, and this page has no data yet in this test.
    await expect(
      page.locator('.period-row').getByRole('button', { name: /generate/i })
    ).toBeVisible();
  });
});

test.describe('Payment collect page — populated table', () => {
  test.beforeEach(async ({ page }) => {
    await mockPaymentsApi(page, [
      makePayment(1, { amount_due: 15000, amount_paid: 0, status: 'Due' }),
      makePayment(2, { amount_due: 20000, amount_paid: 5000, status: 'Partial' }),
      makePayment(3, { amount_due: 10000, amount_paid: 0, status: 'Overdue' }),
    ]);
    await page.goto('/payments/collect');
    await page.waitForSelector('.collect-table', { timeout: 10_000 });
  });

  test('renders one row per due payment with tenant/unit info and defaults', async ({ page }) => {
    const rows = page.locator('.collect-table tbody tr');
    await expect(rows).toHaveCount(3);

    const first = rows.first();
    await expect(first.locator('.tenant-name')).toHaveText('Tenant 1');
    await expect(first.locator('.property-name')).toHaveText('Default Property');
    await expect(first.locator('.unit-info')).toContainText('101');

    // amountPaid defaults to the outstanding balance (amount_due - amount_paid)
    await expect(first.locator('.amount-input')).toHaveValue('15000');
    await expect(first.locator('.method-select')).toHaveValue('');
  });

  test('shows the pending/outstanding stat chips', async ({ page }) => {
    await expect(page.locator('.stat-chip.pending')).toContainText('3');
    await expect(page.locator('.stat-chip.outstanding')).toContainText('40,000');
    await expect(page.locator('.stat-chip.saved')).toHaveCount(0);
    await expect(page.locator('.stat-chip.failed')).toHaveCount(0);
  });

  test('Save All is disabled until a row has an amount and is not skipped', async ({ page }) => {
    const saveButton = page.getByRole('button', { name: /save/i }).last();
    // Rows default to their full outstanding amount, so Save All starts enabled.
    await expect(saveButton).toBeEnabled();

    // Skipping every row disables it again.
    const checkboxes = page.locator('.collect-table mat-checkbox');
    const count = await checkboxes.count();
    for (let i = 0; i < count; i++) {
      await checkboxes.nth(i).click();
    }
    await expect(saveButton).toBeDisabled();
  });

  test('editing the amount input updates the row', async ({ page }) => {
    const firstAmount = page.locator('.collect-table tbody tr').first().locator('.amount-input');
    await firstAmount.fill('5000');
    await expect(firstAmount).toHaveValue('5000');
  });

  test('skip checkbox marks the row skipped and drops the pending count', async ({ page }) => {
    await expect(page.locator('.stat-chip.pending')).toContainText('3');

    const firstRow = page.locator('.collect-table tbody tr').first();
    await firstRow.locator('mat-checkbox').click();

    await expect(firstRow).toHaveClass(/is-skipped/);
    await expect(page.locator('.stat-chip.pending')).toContainText('2');
  });

  test('changing the month reloads the table', async ({ page }) => {
    let requestCount = 0;
    page.on('request', (req) => {
      if (req.method() === 'GET' && req.url().includes('/api/v1/payments')) requestCount++;
    });

    await page.locator('.period-field mat-select').first().click();
    await page.locator('mat-option').filter({ hasText: 'January' }).click();

    await expect(page.locator('.collect-table')).toBeVisible();
    expect(requestCount).toBeGreaterThan(0);
  });
});

test.describe('Payment collect page — saving', () => {
  test('Save All records every pending row on full success', async ({ page }) => {
    await mockPaymentsApi(
      page,
      [
        makePayment(1, { amount_due: 15000, amount_paid: 0, status: 'Due' }),
        makePayment(2, { amount_due: 8000, amount_paid: 0, status: 'Due' }),
      ],
      () => ({ status: 200 })
    );
    await page.goto('/payments/collect');
    await page.waitForSelector('.collect-table', { timeout: 10_000 });

    await page.getByRole('button', { name: /save/i }).last().click();

    await expect(page.locator('.collect-table tbody tr')).toHaveCount(2);
    await expect(page.locator('.collect-table tbody tr.is-saved')).toHaveCount(2);
    await expect(page.locator('.status-icon.saved')).toHaveCount(2);
    await expect(page.locator('.stat-chip.saved')).toContainText('2');
  });

  test('Save All reports a partial failure without touching the successful row', async ({
    page,
  }) => {
    await mockPaymentsApi(
      page,
      [
        makePayment(1, { amount_due: 15000, amount_paid: 0, status: 'Due' }),
        makePayment(2, { amount_due: 8000, amount_paid: 0, status: 'Due' }),
      ],
      (id) => (id === 2 ? { status: 500, body: { error: 'Failed to save' } } : { status: 200 })
    );
    await page.goto('/payments/collect');
    await page.waitForSelector('.collect-table', { timeout: 10_000 });

    await page.getByRole('button', { name: /save/i }).last().click();

    await expect(page.locator('.collect-table tbody tr.is-saved')).toHaveCount(1);
    await expect(page.locator('.collect-table tbody tr.has-error')).toHaveCount(1);
    await expect(page.locator('.stat-chip.saved')).toContainText('1');
    await expect(page.locator('.stat-chip.failed')).toContainText('1');
  });
});
