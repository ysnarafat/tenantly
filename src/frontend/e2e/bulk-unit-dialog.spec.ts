import { test, expect, Locator, Page } from '@playwright/test';
import { injectAuthToken } from './helpers/auth';

const PROPERTY_ID = 1;
const BUILDING_ID = 1;

// Everything this dialog needs is client-side, so auth is faked via
// localStorage (as error-pages.spec.ts does) and every API call it makes is
// mocked — no real session, and therefore no login-rate-limit exposure.
const ADMIN_USER = {
  id: 1,
  username: 'admin-user',
  email: 'admin@tenantly.test',
  role: 'Admin',
  organization_id: 1,
  active: true,
  created_at: new Date(0).toISOString(),
  updated_at: new Date(0).toISOString(),
};

// The building and its unit list are mocked rather than read from the dev seed:
// the dialog's behaviour depends on the building type (which unit types are
// offered) and on which unit numbers already exist (the collision check), and
// neither is guaranteed by the seed data. The bulk POST is mocked too so these
// tests never write units into the dev database.
const BUILDING = {
  id: BUILDING_ID,
  property_id: PROPERTY_ID,
  building_name: 'Main Building',
  building_code: 'BLD-001',
  building_type: 'Residential',
  total_floors: 6,
  has_elevator: true,
  construction_year: 2015,
  active_status: true,
  created_at: new Date(0).toISOString(),
  updated_at: new Date(0).toISOString(),
  property_name: 'Default Property',
  unit_count: 0,
  occupied_units: 0,
  total_revenue: 0,
  occupancy_rate: 0,
};

function makeUnit(id: number, unitNumber: string) {
  return {
    id,
    building_id: BUILDING_ID,
    property_id: PROPERTY_ID,
    unit_number: unitNumber,
    unit_type: 'Apartment',
    active: true,
    created_at: new Date(0).toISOString(),
    updated_at: new Date(0).toISOString(),
    property_name: BUILDING.property_name,
    building_name: BUILDING.building_name,
    building_code: BUILDING.building_code,
    lease_active: false,
  };
}

const isBuildingRequest = (url: URL) => url.pathname.endsWith(`/buildings/${BUILDING_ID}`);
const isUnitListRequest = (url: URL) =>
  url.pathname.endsWith(`/buildings/${BUILDING_ID}/units/list`);
const isBulkRequest = (url: URL) => url.pathname.endsWith(`/buildings/${BUILDING_ID}/units/bulk`);

/** Serves the building, its (optionally pre-populated) unit list, and a
 *  no-op bulk-create endpoint. */
async function mockBuildingApi(page: Page, existingUnitNumbers: string[] = []) {
  const units = existingUnitNumbers.map((number, index) => makeUnit(index + 1, number));

  await page.route(isBuildingRequest, (route) => route.fulfill({ json: { building: BUILDING } }));
  await page.route(isUnitListRequest, (route) =>
    route.fulfill({
      json: { data: units, meta: { total: units.length, page: 1, page_size: 20 } },
    })
  );
  await page.route(isBulkRequest, async (route) => {
    const body = route.request().postDataJSON() as { units: unknown[] };
    await route.fulfill({
      json: {
        message: 'Units created successfully',
        building_id: BUILDING_ID,
        units: [],
        summary: { units_created: body.units.length },
      },
    });
  });
}

async function openBulkDialog(page: Page, existingUnitNumbers: string[] = []): Promise<Locator> {
  await injectAuthToken(page, 'fake-jwt-token', ADMIN_USER);
  await mockBuildingApi(page, existingUnitNumbers);
  await page.goto(`/properties/${PROPERTY_ID}/buildings/${BUILDING_ID}`);
  await page.waitForSelector('.units-grid, .empty-tab', { timeout: 10_000 });

  await page.getByRole('button', { name: /bulk add units/i }).click();
  const dialog = page.locator('mat-dialog-container');
  await expect(dialog).toBeVisible({ timeout: 5_000 });
  return dialog;
}

const dataRows = (dialog: Locator) => dialog.locator('.unit-rows .unit-row:not(.unit-row--head)');
const numberInput = (dialog: Locator, index: number) =>
  dataRows(dialog).nth(index).locator('.cell-input--number');
const generateButton = (dialog: Locator) => dialog.getByRole('button', { name: /^generate$/i });
const createButton = (dialog: Locator) => dialog.getByRole('button', { name: /create \d+ units/i });

/** Fills the generator fields. Left as separate steps because each field is a
 *  distinct Material form field rather than one composite control. */
async function fillGenerator(
  dialog: Locator,
  values: { pattern?: string; start?: string; count?: string }
) {
  if (values.pattern !== undefined) {
    await dialog.locator('.pattern-field input').fill(values.pattern);
  }
  if (values.start !== undefined) {
    await dialog.locator('.start-field input').fill(values.start);
  }
  if (values.count !== undefined) {
    await dialog.locator('.count-field input').fill(values.count);
  }
}

test.describe('Bulk unit dialog — layout', () => {
  test('opens with generator, shared-details and units panels', async ({ page }) => {
    const dialog = await openBulkDialog(page);

    await expect(dialog.locator('.panel')).toHaveCount(3);
    await expect(dialog.locator('.title-main')).toHaveText(/bulk add units/i);
    // Breadcrumb names the property and building the batch lands in
    await expect(dialog.locator('.title-context')).toContainText('Default Property');
    await expect(dialog.locator('.title-context')).toContainText('Main Building');
  });

  test('starts with one blank row, submit disabled and a prompt in the footer', async ({
    page,
  }) => {
    const dialog = await openBulkDialog(page);

    await expect(dataRows(dialog)).toHaveCount(1);
    await expect(numberInput(dialog, 0)).toHaveValue('');
    await expect(createButton(dialog)).toBeDisabled();
    await expect(dialog.locator('.status')).toContainText(/add at least one unit number/i);
    await expect(dialog.locator('.row-error')).toHaveCount(0);
  });

  test('offers only the unit types allowed for a residential building', async ({ page }) => {
    const dialog = await openBulkDialog(page);

    await dialog.locator('mat-select[formcontrolname="unit_type"]').click();
    const options = page.locator('mat-option');
    // UNIT_TYPES_BY_BUILDING_TYPE.Residential — Apartment, Parking, Storage
    await expect(options).toHaveCount(3);
    await expect(options.first()).toContainText('Apartment');
    await page.keyboard.press('Escape');
  });
});

test.describe('Bulk unit dialog — generator', () => {
  test('previews the head and tail of the sequence before generating', async ({ page }) => {
    const dialog = await openBulkDialog(page);
    await fillGenerator(dialog, { pattern: 'A-{n}', start: '101', count: '10' });

    // Three leading numbers, an ellipsis, then the last one
    const chips = dialog.locator('.preview-chip');
    await expect(chips).toHaveCount(4);
    await expect(chips.nth(0)).toHaveText('A-101');
    await expect(chips.nth(1)).toHaveText('A-102');
    await expect(chips.nth(2)).toHaveText('A-103');
    await expect(dialog.locator('.preview-ellipsis')).toBeVisible();
    await expect(chips.nth(3)).toHaveText('A-110');

    // Preview only — nothing added to the list yet
    await expect(dataRows(dialog)).toHaveCount(1);
  });

  test('Generate fills the rows and replaces the blank starter row', async ({ page }) => {
    const dialog = await openBulkDialog(page);
    await fillGenerator(dialog, { pattern: 'A-{n}', start: '101', count: '10' });
    await generateButton(dialog).click();

    await expect(dataRows(dialog)).toHaveCount(10);
    await expect(numberInput(dialog, 0)).toHaveValue('A-101');
    await expect(numberInput(dialog, 9)).toHaveValue('A-110');
    await expect(dialog.locator('.status')).toContainText(/10 unit\(s\) ready/i);
    await expect(createButton(dialog)).toBeEnabled();
    await expect(createButton(dialog)).toContainText('10');
  });

  test('a pattern without {n} gets the number appended, and leading zeros are padded', async ({
    page,
  }) => {
    const dialog = await openBulkDialog(page);
    await fillGenerator(dialog, { pattern: 'B', start: '01', count: '3' });
    await generateButton(dialog).click();

    await expect(dataRows(dialog)).toHaveCount(3);
    await expect(numberInput(dialog, 0)).toHaveValue('B01');
    await expect(numberInput(dialog, 1)).toHaveValue('B02');
    await expect(numberInput(dialog, 2)).toHaveValue('B03');
  });

  test('generating twice keeps the first batch and appends the second', async ({ page }) => {
    const dialog = await openBulkDialog(page);
    await fillGenerator(dialog, { pattern: 'A-{n}', start: '101', count: '2' });
    await generateButton(dialog).click();

    await fillGenerator(dialog, { pattern: 'B-{n}', start: '201', count: '2' });
    await generateButton(dialog).click();

    await expect(dataRows(dialog)).toHaveCount(4);
    await expect(numberInput(dialog, 0)).toHaveValue('A-101');
    await expect(numberInput(dialog, 3)).toHaveValue('B-202');
  });

  test('Generate is disabled when the sequence would be empty', async ({ page }) => {
    const dialog = await openBulkDialog(page);

    await fillGenerator(dialog, { count: '0' });
    await expect(generateButton(dialog)).toBeDisabled();

    await fillGenerator(dialog, { start: 'abc', count: '5' });
    await expect(generateButton(dialog)).toBeDisabled();
  });

  test('caps the batch at 100 rows and says so', async ({ page }) => {
    const dialog = await openBulkDialog(page);

    await fillGenerator(dialog, { pattern: 'A-{n}', start: '1', count: '60' });
    await generateButton(dialog).click();
    await expect(dataRows(dialog)).toHaveCount(60);
    await expect(dialog.locator('.capped-notice')).toHaveCount(0);

    // Only 40 of the next 60 fit
    await fillGenerator(dialog, { pattern: 'B-{n}', start: '1', count: '60' });
    await generateButton(dialog).click();

    await expect(dataRows(dialog)).toHaveCount(100);
    await expect(dialog.locator('.capped-notice')).toContainText(/100/);
    await expect(dialog.getByRole('button', { name: /add unit/i })).toBeDisabled();
  });
});

test.describe('Bulk unit dialog — row editing', () => {
  test('pasting a column fills this row and the ones below it', async ({ page }) => {
    const dialog = await openBulkDialog(page);

    // A real clipboard write would need browser permissions; dispatching the
    // paste event with its own DataTransfer exercises the same handler.
    await numberInput(dialog, 0).evaluate((element, text) => {
      const transfer = new DataTransfer();
      transfer.setData('text/plain', text);
      element.dispatchEvent(
        new ClipboardEvent('paste', { clipboardData: transfer, bubbles: true, cancelable: true })
      );
    }, 'C-1\nC-2\nC-3');

    await expect(dataRows(dialog)).toHaveCount(3);
    await expect(numberInput(dialog, 0)).toHaveValue('C-1');
    await expect(numberInput(dialog, 1)).toHaveValue('C-2');
    await expect(numberInput(dialog, 2)).toHaveValue('C-3');
    await expect(dialog.locator('.status')).toContainText(/3 unit\(s\) ready/i);
  });

  test('Enter on the last row adds a row and moves focus into it', async ({ page }) => {
    const dialog = await openBulkDialog(page);

    await numberInput(dialog, 0).fill('A-101');
    await numberInput(dialog, 0).press('Enter');

    await expect(dataRows(dialog)).toHaveCount(2);
    await expect(numberInput(dialog, 1)).toBeFocused();
  });

  test('Add unit appends an empty row', async ({ page }) => {
    const dialog = await openBulkDialog(page);

    await dialog.getByRole('button', { name: /add unit/i }).click();

    await expect(dataRows(dialog)).toHaveCount(2);
    await expect(numberInput(dialog, 1)).toHaveValue('');
  });

  test('the row remove button drops just that row', async ({ page }) => {
    const dialog = await openBulkDialog(page);
    await fillGenerator(dialog, { pattern: 'A-{n}', start: '101', count: '3' });
    await generateButton(dialog).click();

    await dataRows(dialog).nth(1).locator('button.cell-remove').click();

    await expect(dataRows(dialog)).toHaveCount(2);
    await expect(numberInput(dialog, 0)).toHaveValue('A-101');
    await expect(numberInput(dialog, 1)).toHaveValue('A-103');
  });

  test('Clear all resets the list to a single blank row', async ({ page }) => {
    const dialog = await openBulkDialog(page);
    await fillGenerator(dialog, { pattern: 'A-{n}', start: '101', count: '5' });
    await generateButton(dialog).click();
    await expect(dataRows(dialog)).toHaveCount(5);

    await dialog.getByRole('button', { name: /clear all/i }).click();

    await expect(dataRows(dialog)).toHaveCount(1);
    await expect(numberInput(dialog, 0)).toHaveValue('');
    await expect(createButton(dialog)).toBeDisabled();
  });

  test('empty rows are skipped rather than flagged as errors', async ({ page }) => {
    const dialog = await openBulkDialog(page);
    await fillGenerator(dialog, { pattern: 'A-{n}', start: '101', count: '2' });
    await generateButton(dialog).click();
    await dialog.getByRole('button', { name: /add unit/i }).click();

    await expect(dataRows(dialog)).toHaveCount(3);
    await expect(dialog.locator('.row-error')).toHaveCount(0);
    await expect(dialog.locator('.status')).toContainText(/2 unit\(s\) ready/i);
    await expect(dialog.locator('.status')).toContainText(/1 empty row\(s\) skipped/i);
    await expect(createButton(dialog)).toBeEnabled();
    await expect(createButton(dialog)).toContainText('2');
  });
});

test.describe('Bulk unit dialog — validation', () => {
  test('a duplicate unit number flags the row and blocks submit', async ({ page }) => {
    const dialog = await openBulkDialog(page);
    await fillGenerator(dialog, { pattern: 'A-{n}', start: '101', count: '3' });
    await generateButton(dialog).click();

    await numberInput(dialog, 1).fill('A-101');

    // Both occurrences are marked, not just the second
    await expect(dialog.locator('.unit-row.is-invalid')).toHaveCount(2);
    await expect(dialog.locator('.row-error').first()).toContainText(/duplicate unit number/i);
    await expect(dialog.locator('.status')).toContainText(/duplicate/i);
    await expect(dialog.locator('.status.is-error')).toBeVisible();
    await expect(createButton(dialog)).toBeDisabled();
  });

  test('a unit number that already exists in the building is flagged before submit', async ({
    page,
  }) => {
    const dialog = await openBulkDialog(page, ['A-101', 'A-102']);

    await numberInput(dialog, 0).fill('A-101');

    await expect(dialog.locator('.unit-row.is-invalid')).toHaveCount(1);
    await expect(dialog.locator('.row-error')).toContainText(/already exists/i);
    await expect(dialog.locator('.status')).toContainText(/already exist/i);
    await expect(createButton(dialog)).toBeDisabled();
  });

  test('an unused unit number passes even when others exist in the building', async ({ page }) => {
    const dialog = await openBulkDialog(page, ['A-101', 'A-102']);

    await numberInput(dialog, 0).fill('A-103');

    await expect(dialog.locator('.unit-row.is-invalid')).toHaveCount(0);
    await expect(createButton(dialog)).toBeEnabled();
  });

  test('fixing the duplicate re-enables submit', async ({ page }) => {
    const dialog = await openBulkDialog(page);
    await fillGenerator(dialog, { pattern: 'A-{n}', start: '101', count: '2' });
    await generateButton(dialog).click();

    await numberInput(dialog, 1).fill('A-101');
    await expect(createButton(dialog)).toBeDisabled();

    await numberInput(dialog, 1).fill('A-102');
    await expect(dialog.locator('.unit-row.is-invalid')).toHaveCount(0);
    await expect(createButton(dialog)).toBeEnabled();
  });
});

test.describe('Bulk unit dialog — submit', () => {
  test('Cancel closes the dialog without posting anything', async ({ page }) => {
    const dialog = await openBulkDialog(page);
    await fillGenerator(dialog, { pattern: 'A-{n}', start: '101', count: '3' });
    await generateButton(dialog).click();

    let posted = false;
    page.on('request', (request) => {
      if (request.method() === 'POST' && isBulkRequest(new URL(request.url()))) posted = true;
    });

    await dialog.getByRole('button', { name: /cancel/i }).click();
    await expect(dialog).toBeHidden();
    expect(posted).toBe(false);
  });

  test('Create posts every filled row with the shared details applied', async ({ page }) => {
    const dialog = await openBulkDialog(page);
    await fillGenerator(dialog, { pattern: 'A-{n}', start: '101', count: '3' });
    await generateButton(dialog).click();

    // Shared details apply to the whole batch; one row also gets its own name
    await dialog.locator('input[formcontrolname="floor"]').fill('3');
    await dialog.locator('input[formcontrolname="section"]').fill('East');
    await dataRows(dialog).nth(0).locator('.cell-input').nth(1).fill('Corner flat');
    // Trailing blank row must not reach the payload
    await dialog.getByRole('button', { name: /add unit/i }).click();

    const [request] = await Promise.all([
      page.waitForRequest(
        (candidate) => candidate.method() === 'POST' && isBulkRequest(new URL(candidate.url()))
      ),
      createButton(dialog).click(),
    ]);

    expect(request.postDataJSON()).toEqual({
      units: [
        {
          unit_number: 'A-101',
          unit_name: 'Corner flat',
          unit_type: 'Apartment',
          floor: 3,
          section: 'East',
        },
        { unit_number: 'A-102', unit_type: 'Apartment', floor: 3, section: 'East' },
        { unit_number: 'A-103', unit_type: 'Apartment', floor: 3, section: 'East' },
      ],
    });

    await expect(dialog).toBeHidden();
  });
});
