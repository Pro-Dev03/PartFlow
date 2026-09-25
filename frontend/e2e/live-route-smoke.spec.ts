import { expect, test, type Page } from '@playwright/test';

const routes = [
  '/app/dashboard',
  '/app/activity',
  '/app/sales',
  '/app/inventory',
  '/app/customers',
  '/app/debts',
  '/app/suppliers',
  '/app/purchases',
  '/app/purchases/create',
  '/app/expenses',
  '/app/returns',
  '/app/returns/create',
  '/app/supplier-returns',
  '/app/reports',
  '/app/settings',
  '/app/categories',
  '/app/audit',
  '/app/archive',
];

async function signInToLiveSite(page: Page) {
  const email = process.env.E2E_EMAIL;
  const password = process.env.E2E_PASSWORD;
  if (!email || !password) throw new Error('E2E_EMAIL and E2E_PASSWORD are required');

  await page.goto('/#/login');
  const cloudMode = page.getByRole('button', { name: /سحابي Online|Cloud Online/i });
  if (await cloudMode.count()) await cloudMode.click();
  await page.locator('input[type="email"]').fill(email);
  await page.locator('input[type="password"]').fill(password);
  await page.locator('button[type="submit"]').click();
  await page.waitForURL(/#\/app(?:\/|$)/, { timeout: 35_000 });
}

test('Render website routes load and their cloud API requests do not fail', async ({ page }) => {
  test.setTimeout(180_000);
  const runtimeErrors: string[] = [];
  const apiFailures = new Set<string>();
  const pageChecks: Array<{ route: string; heading: string; contentLength: number }> = [];

  page.on('pageerror', (error) => runtimeErrors.push(error.message));
  page.on('response', (response) => {
    const url = new URL(response.url());
    if (url.pathname.startsWith('/api/v1/') && response.status() >= 400) {
      apiFailures.add(`${response.status()} ${url.pathname}`);
    }
  });

  await signInToLiveSite(page);

  for (const route of routes) {
    await page.evaluate((path) => { window.location.hash = path; }, route);
    await expect(page.locator('#main-content')).toBeVisible({ timeout: 15_000 });
    await page.waitForTimeout(500);
    const content = await page.locator('#main-content').innerText().catch(() => '');
    const heading = await page.locator('#main-content h1, #main-content h2').first().innerText().catch(() => '');
    pageChecks.push({ route, heading: heading.trim().slice(0, 80), contentLength: content.trim().length });
  }

  console.log(`[LIVE_ROUTE_SMOKE] ${JSON.stringify({ pageChecks, apiFailures: [...apiFailures], runtimeErrors })}`);
  expect(pageChecks.filter((check) => check.contentLength < 10).map((check) => check.route)).toEqual([]);
  expect(apiFailures).toEqual(new Set());
  expect(runtimeErrors).toEqual([]);
});
