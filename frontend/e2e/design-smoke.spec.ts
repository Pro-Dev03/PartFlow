import { expect, test, type Page } from '@playwright/test';

const routes = [
  '/app/dashboard', '/app/activity', '/app/sales', '/app/inventory',
  '/app/customers', '/app/debts', '/app/suppliers', '/app/purchases',
  '/app/purchases/create', '/app/expenses', '/app/returns',
  '/app/returns/create', '/app/supplier-returns', '/app/reports',
  '/app/categories', '/app/archive', '/app/settings',
];

async function login(page: Page) {
  const email = process.env.E2E_EMAIL;
  const password = process.env.E2E_PASSWORD;
  if (!email || !password) throw new Error('E2E_EMAIL and E2E_PASSWORD are required');

  // The isolated test API serves authentication and data for this session.
  await page.addInitScript(() => {
    localStorage.setItem('partflow-connection-mode', 'cloud');
    localStorage.setItem('partflow-cloud-api-url', 'http://127.0.0.1:8082/api/v1');
  });
  await page.goto('/#/login');
  await page.locator('input[type="email"]').fill(email);
  await page.locator('input[type="password"]').fill(password);
  await page.locator('button[type="submit"]').click();
  await page.waitForURL(/#\/app(?:\/dashboard)?(?:\?.*)?$/, { timeout: 20_000 });
}

test('every routed page renders without runtime errors', async ({ page }) => {
  const errors: string[] = [];
  const failedRequests = new Set<string>();
  page.on('pageerror', (error) => errors.push(error.message));
  page.on('response', (response) => {
    if (response.status() >= 500) failedRequests.add(`${response.status()} ${new URL(response.url()).pathname}`);
  });

  await login(page);
  for (const route of routes) {
    await page.evaluate((nextRoute) => { window.location.hash = nextRoute; }, route);
    await expect(page.locator('#main-content')).toBeVisible();
    await expect(page.locator('#main-content')).toContainText(/\S/);
    await expect(page.locator('body')).not.toContainText('Failed to resolve import');
    await page.waitForTimeout(350);
  }

  expect(errors).toEqual([]);
  expect([...failedRequests]).toEqual([]);
});

test('mobile menu keeps section labels and search reaches inventory', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await login(page);
  await page.locator('button[aria-controls="app-sidebar"]').click();
  const sidebar = page.locator('#app-sidebar');
  await expect(sidebar).toBeVisible();
  await expect(sidebar.locator('.nav-label').filter({ hasText: /المشتريات/ })).toBeVisible();
  await sidebar.locator('.sidebar-item').filter({ hasText: /^المخزون$/ }).click();
  await expect(page).toHaveURL(/#\/app\/inventory$/);
  await expect(sidebar).toBeHidden();

  await page.setViewportSize({ width: 1280, height: 800 });
  const search = page.locator('.pf-header-search input');
  await search.fill('sample search');
  await search.press('Enter');
  await expect(page).toHaveURL(/#\/app\/inventory\?search=sample%20search$/);
});

test('key work surfaces render at desktop and mobile sizes in both themes', async ({ page }) => {
  await login(page);
  await expect(page.locator('#app-sidebar')).toHaveCSS('width', '248px');
  const desktopBounds = await page.locator('#app-sidebar').evaluate((node) => ({
    left: node.getBoundingClientRect().left,
    right: node.getBoundingClientRect().right,
    viewport: window.innerWidth,
    page: document.documentElement.scrollWidth,
  }));
  expect(desktopBounds.right).toBeLessThanOrEqual(desktopBounds.viewport);
  for (const theme of ['light', 'dark']) {
    await page.evaluate((value) => localStorage.setItem('theme', value), theme);
    for (const route of ['dashboard', 'sales', 'reports', 'settings']) {
      await page.evaluate((nextRoute) => { window.location.hash = nextRoute; }, `/app/${route}`);
      await expect(page.locator('#main-content')).toBeVisible();
      await expect(page.locator('#main-content')).toContainText(/\S/);
      await page.waitForTimeout(500);
      await page.screenshot({ path: `test-results/${theme}-${route}-desktop.png`, fullPage: true });
      await page.setViewportSize({ width: 390, height: 844 });
      await expect(page.locator('#app-sidebar')).toBeHidden();
      if (route === 'dashboard' || route === 'settings') {
        const headingHeight = await page.locator('.pf-page-header').first().evaluate((node) => node.getBoundingClientRect().height);
        expect(headingHeight).toBeLessThan(170);
      }
      await page.screenshot({ path: `test-results/${theme}-${route}-mobile.png`, fullPage: true });
      if (route === 'sales') {
        const bounds = await page.evaluate(() => {
          const header = document.querySelector('.pos-modern-header')?.getBoundingClientRect();
          const right = document.querySelector('.pos-header-right')?.getBoundingClientRect();
          const products = document.querySelector('.pos-products-area')?.getBoundingClientRect();
          const headerStyle = getComputedStyle(document.querySelector('.pos-modern-header')!);
          const rightStyle = getComputedStyle(document.querySelector('.pos-header-right')!);
          return {
            headerTop: header?.top, headerBottom: header?.bottom, headerHeight: header?.height,
            rightTop: right?.top, rightBottom: right?.bottom, productsTop: products?.top,
            display: headerStyle.display, rows: headerStyle.gridTemplateRows,
            height: headerStyle.height, minHeight: headerStyle.minHeight,
            rightPosition: rightStyle.position,
          };
        });
        expect(bounds.rightBottom, JSON.stringify(bounds)).toBeLessThanOrEqual(bounds.headerBottom ?? 0);
        expect(bounds.productsTop, JSON.stringify(bounds)).toBeGreaterThanOrEqual(bounds.headerBottom ?? 0);
      }
      if (route === 'settings') {
        const bounds = await page.evaluate(() => {
          const box = (selector: string) => {
            const node = document.querySelector(selector);
            const rect = node?.getBoundingClientRect();
            return { left: rect?.left, right: rect?.right, width: rect?.width };
          };
          return {
            viewport: window.innerWidth, scrollWidth: document.documentElement.scrollWidth,
            main: box('#main-content'), page: box('.settings-page'),
            account: box('.settings-account-overview'), nav: box('.settings-section-nav'),
            sidebar: box('#app-sidebar'), sidebarClass: document.querySelector('#app-sidebar')?.className,
            sidebarVisibility: getComputedStyle(document.querySelector('#app-sidebar')!).visibility,
          };
        });
        expect(bounds.scrollWidth, JSON.stringify(bounds)).toBeLessThanOrEqual(bounds.viewport);
        expect(bounds.account.right, JSON.stringify(bounds)).toBeLessThanOrEqual(bounds.viewport);
        expect(bounds.nav.right, JSON.stringify(bounds)).toBeLessThanOrEqual(bounds.viewport);
        expect(bounds.sidebarVisibility, JSON.stringify(bounds)).toBe('hidden');
      }
      await page.setViewportSize({ width: 1280, height: 800 });
    }
  }
});
