import { test, expect } from '@playwright/test';

test.use({ storageState: 'e2e/.auth/user.json' });

const organizationRoutes = [
  '/dashboard',
  '/properties',
  '/tenants',
  '/leases',
  '/leases/due',
  '/payments',
  '/payments/collect',
  '/billing-runs',
  '/reports',
  '/documents',
  '/users',
];

async function expectInsideViewport(page: import('@playwright/test').Page, selector: string) {
  const element = page.locator(selector).last();
  const bounds = await element.boundingBox();
  const viewport = page.viewportSize();
  if (!bounds || !viewport) {
    throw new Error(`${selector} is not measurable`);
  }

  expect(bounds.x).toBeGreaterThanOrEqual(0);
  expect(bounds.y).toBeGreaterThanOrEqual(0);
  expect(bounds.x + bounds.width).toBeLessThanOrEqual(viewport.width);
  expect(bounds.y + bounds.height).toBeLessThanOrEqual(viewport.height);
}

test('authenticated app shell fits a phone viewport and the navigation drawer toggles', async ({
  page,
  isMobile,
}) => {
  test.skip(!isMobile, 'This assertion is specific to touch-sized mobile layouts');

  await page.goto('/dashboard');
  await expect(page.locator('.app-toolbar')).toBeVisible();

  await expect
    .poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth))
    .toBe(true);

  const menuButton = page.locator('.menu-button');
  const menuBounds = await menuButton.boundingBox();
  if (!menuBounds) {
    throw new Error('The mobile navigation button is not visible in the viewport');
  }
  expect(menuBounds.width).toBeGreaterThanOrEqual(44);
  expect(menuBounds.height).toBeGreaterThanOrEqual(44);

  const sidenav = page.locator('mat-sidenav');
  await menuButton.click();
  await expect(sidenav).toHaveClass(/mat-drawer-opened/);

  await menuButton.click();
  await expect(sidenav).not.toHaveClass(/mat-drawer-opened/);
});

test('organization dropdown and popup menu stay within the mobile viewport', async ({
  page,
  isMobile,
}) => {
  test.skip(!isMobile, 'This assertion is specific to mobile header layout');

  await page.setViewportSize({ width: 384, height: 860 });
  await page.goto('/dashboard');
  for (const selector of [
    '.menu-button',
    '.app-branding',
    'app-organization-selector .org-button',
    '.language-toggle',
    '.avatar-btn',
  ]) {
    await expectInsideViewport(page, selector);
  }
  const viewport = page.viewportSize();
  const avatarBounds = await page.locator('.avatar-btn').boundingBox();
  if (!viewport || !avatarBounds) {
    throw new Error('The mobile user avatar is not measurable');
  }
  expect(avatarBounds.x + avatarBounds.width).toBeLessThanOrEqual(viewport.width - 16);

  const organizationButton = page.getByRole('button', { name: /switch organization/i });
  await expect(organizationButton).toBeVisible();
  await organizationButton.click();

  const menu = page.getByRole('menu').last();
  await expect(menu).toBeVisible();
  await expect(menu.getByText('Organizations', { exact: true })).toBeVisible();
  await expectInsideViewport(page, '[role="menu"]');
  await page.keyboard.press('Escape');
  await expect(menu).toBeHidden();
});

test('organization pages render without page-level horizontal overflow on mobile', async ({
  page,
}) => {
  for (const route of organizationRoutes) {
    await page.goto(route);
    await expect(
      page.locator('.app-toolbar'),
      `${route} should remain authenticated`
    ).toBeVisible();
    await expect(
      page.locator('.main-content h1').first(),
      `${route} should render its page heading`
    ).toBeVisible();
    await expect
      .poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), {
        message: `${route} should not create page-level horizontal scrolling`,
      })
      .toBe(true);
  }
});

test('tenant rows stack as cards and their action menu fits on mobile', async ({
  page,
  isMobile,
}) => {
  test.skip(!isMobile, 'This assertion is specific to mobile table layout');

  await page.setViewportSize({ width: 384, height: 860 });
  await page.route('**/api/v1/tenants**', (route) =>
    route.fulfill({
      json: {
        tenants: [
          {
            id: 9201,
            name: 'A Long Tenant Name For Narrow Mobile Screens',
            tenant_type: 'Individual',
            phone_number: '+8801712345678',
            email: 'tenant@example.test',
            nid_last_four: '1234',
            active: true,
            created_at: new Date(0).toISOString(),
            updated_at: new Date(0).toISOString(),
          },
        ],
        pagination: {
          current_page: 1,
          page_size: 100,
          total_items: 1,
          total_pages: 1,
          has_next: false,
          has_prev: false,
        },
      },
    })
  );

  await page.goto('/tenants');
  const card = page.locator('.tenant-mobile-card').first();
  await expect(card).toBeVisible();
  await expect(page.locator('.tenant-desktop-table')).toBeHidden();
  await expect(card.locator('.tenant-mobile-name')).toHaveText(
    'A Long Tenant Name For Narrow Mobile Screens'
  );
  await expect(card.locator('.tenant-mobile-email')).toHaveAttribute(
    'href',
    'mailto:tenant@example.test'
  );
  await expect(card.locator('.tenant-mobile-meta-grid')).toHaveCSS('display', 'grid');

  const expectCardSectionsToStack = async () => {
    const header = await card.locator('.tenant-mobile-card-header').boundingBox();
    const contacts = await card.locator('.tenant-mobile-contacts').boundingBox();
    const details = await card.locator('.tenant-mobile-meta-grid').boundingBox();
    if (!header || !contacts || !details) {
      throw new Error('The mobile tenant card sections are not measurable');
    }

    expect(header.y + header.height).toBeLessThanOrEqual(contacts.y);
    expect(contacts.y + contacts.height).toBeLessThanOrEqual(details.y);
    await expect
      .poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth))
      .toBe(true);
    await expect
      .poll(() =>
        card.evaluate((element) => element.scrollWidth <= element.clientWidth)
      )
      .toBe(true);
  };

  await expectCardSectionsToStack();
  await page.setViewportSize({ width: 360, height: 800 });
  await expectCardSectionsToStack();

  await card.locator('.tenant-mobile-actions').click();
  const menu = page.getByRole('menu').last();
  await expect(menu).toBeVisible();
  await expectInsideViewport(page, '[role="menu"]');
  await page.keyboard.press('Escape');
});

test('property cards and actions fit a narrow mobile viewport', async ({ page, isMobile }) => {
  test.skip(!isMobile, 'This assertion is specific to mobile property layout');

  await page.route('**/api/v1/properties', (route) =>
    route.fulfill({
      json: {
        properties: [
          {
            id: 9201,
            property_name: 'Long Property Name For Narrow Mobile Screens',
            property_code: 'MOBILE-PROPERTY-01',
            address: 'A long street address in Dhaka that should wrap instead of widening the page',
            city: 'Dhaka',
            property_type: 'Residential',
            total_buildings: 0,
            active: true,
            created_at: new Date(0).toISOString(),
            updated_at: new Date(0).toISOString(),
          },
        ],
        total: 1,
      },
    })
  );

  await page.goto('/properties');
  const card = page.locator('app-property-card .property-card').first();
  await expect(card).toBeVisible();
  await expect(card.locator('mat-card-actions')).toHaveCSS('display', 'grid');
  await expect
    .poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth))
    .toBe(true);
});

test('delete confirmation actions remain clickable on a small phone', async ({ page }) => {
  await page.route('**/api/v1/properties/1**', async (route) => {
    const pathname = new URL(route.request().url()).pathname;
    if (pathname.endsWith('/buildings')) {
      await route.fulfill({ json: { buildings: [], total: 0, page: 1, page_size: 10 } });
      return;
    }

    await route.fulfill({
      json: {
        property: {
          id: 1,
          property_name: 'Mobile Test Property',
          property_code: 'MOBILE-1',
          address: 'Dhaka',
          property_type: 'Residential',
          total_buildings: 0,
          active: true,
          created_at: new Date(0).toISOString(),
          updated_at: new Date(0).toISOString(),
          building_count: 0,
          unit_count: 0,
          occupied_units: 0,
          total_revenue: 0,
        },
      },
    });
  });

  await page.goto('/properties/1');
  await expect(page.locator('.detail-header')).toBeVisible();
  await page.getByRole('button', { name: /delete/i }).click();

  const dialog = page.locator('mat-dialog-container');
  await expect(dialog).toBeVisible();
  await expect(dialog.getByRole('button', { name: /cancel/i })).toBeVisible();
  const dialogBounds = await dialog.boundingBox();
  const viewport = page.viewportSize();
  if (!dialogBounds || !viewport) {
    throw new Error('The mobile confirmation dialog is not measurable');
  }
  expect(dialogBounds.height).toBeLessThanOrEqual(viewport.height);

  await dialog.getByRole('button', { name: /cancel/i }).click();
  await expect(dialog).toBeHidden();
});
