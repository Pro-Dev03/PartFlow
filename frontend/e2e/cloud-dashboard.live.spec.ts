import { expect, test } from '@playwright/test';

const email = process.env.E2E_EMAIL;
const password = process.env.E2E_PASSWORD;
const apiOrigin = 'https://partflow-api.onrender.com';

test.use({ serviceWorkers: 'block' });
test.skip(
  !email || !password || process.env.E2E_BASE_URL !== 'https://partflow-hpv7.onrender.com',
  'Live QA credentials and E2E_BASE_URL=https://partflow-hpv7.onrender.com are required',
);

test('cloud dashboard reconciles visible metrics with live API data', async ({ page }) => {
  const apiFailures: Array<{ path: string; status: number; phase: string }> = [];
  const pageErrors: string[] = [];
  let phase = 'startup';
  let stats: Record<string, unknown> | undefined;
  let activity: Array<Record<string, unknown>> | undefined;

  page.on('pageerror', (error) => pageErrors.push(error.message));
  page.on('response', async (response) => {
    if (!response.url().startsWith(`${apiOrigin}/api/v1/`)) return;
    const path = new URL(response.url()).pathname;
    if (response.status() >= 400) apiFailures.push({ path, status: response.status(), phase });
    if (path === '/api/v1/dashboard/stats' && response.ok()) {
      stats = (await response.json().catch(() => undefined))?.data;
    }
    if (path === '/api/v1/dashboard/activity' && response.ok()) {
      activity = (await response.json().catch(() => undefined))?.data?.items;
    }
  });

  await page.addInitScript(() => localStorage.setItem('partflow-connection-mode', 'cloud'));
  await page.goto('/#/login', { waitUntil: 'domcontentloaded' });
  await page.locator('#email').fill(email!);
  await page.locator('#password').fill(password!);
  await page.locator('form button[type="submit"]').click();
  await expect(page).toHaveURL(/#\/app(?:\/|$)/, { timeout: 30_000 });
  phase = 'dashboard';
  await page.goto('/#/app/dashboard', { waitUntil: 'domcontentloaded' });

  await expect(page.locator('#dashboard-core-metrics')).toBeVisible({ timeout: 30_000 });
  await expect(page.locator('#dashboard-financial-operations')).toBeVisible();
  await expect(page.locator('.dashboard-attention')).toBeVisible();
  await expect(page.locator('.dashboard-smart-actions')).toBeVisible();
  await expect(page.locator('.dashboard-performance-card')).toBeVisible();
  await expect.poll(() => stats, { timeout: 30_000 }).toBeTruthy();

  const values = stats!;
  const coreValues = await page.locator('section[aria-labelledby="dashboard-core-metrics"] .numeric-metric').allTextContents();
  const financeValues = await page.locator('section[aria-labelledby="dashboard-financial-operations"] .numeric-metric').allTextContents();
  expect(coreValues).toHaveLength(4);
  expect(financeValues).toHaveLength(5);

  const formatted = await page.evaluate((numbers) => numbers.map((value) =>
    `₪${Number(value ?? 0).toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 })}`
  ), [
    values.todaySales,
    values.todayProfit,
    values.outstandingDebts,
    values.todaySupplierReturns,
    values.todayCollected,
    values.todayDebtCollected,
    values.todaySupplierPaid,
    values.todayExpenses,
    values.todayCashDifference,
  ]);
  expect([...coreValues, ...financeValues].map((value) => value.trim())).toEqual(formatted);
  await expect(page.locator('section[aria-labelledby="dashboard-core-metrics"] .numeric-quantity')
    .filter({ hasText: /^\s*\d+\s*$/ }))
    .toHaveText(String(Number(values.lowStockCount ?? 0)));

  const rangeButtons = page.locator('.dashboard-performance-range-button');
  await expect(rangeButtons).toHaveCount(4);
  for (let index = 0; index < 4; index += 1) {
    await rangeButtons.nth(index).click();
    await expect(rangeButtons.nth(index)).toHaveAttribute('class', /primary/);
    await expect(page.locator('.dashboard-performance-card')).toBeVisible();
  }

  await expect.poll(() => activity, { timeout: 30_000 }).toBeDefined();
  if (activity!.length > 0) {
    await expect(page.getByText(String(activity![0].title), { exact: true }).first()).toBeVisible();
  }

  const followDashboardAction = async (button: ReturnType<typeof page.locator>, expectedPath: string) => {
    phase = 'navigation';
    await button.click();
    await expect.poll(() => new URL(page.url()).hash).toBe(`#${expectedPath}`);
    await page.goBack();
    await expect(page).toHaveURL(/#\/app\/dashboard$/);
    await expect(page.locator('#dashboard-core-metrics')).toBeVisible();
    phase = 'dashboard';
  };

  const coreCards = page.locator('section[aria-labelledby="dashboard-core-metrics"] .unified-stat-card');
  const corePaths = [
    '/app/reports?report=net-sales&range=today',
    '/app/reports?report=profit&range=today',
    '/app/debts',
    '/app/inventory',
    '/app/supplier-returns',
  ];
  await expect(coreCards).toHaveCount(corePaths.length);
  for (let index = 0; index < corePaths.length; index += 1) {
    await followDashboardAction(coreCards.nth(index), corePaths[index]);
  }

  const financeCards = page.locator('section[aria-labelledby="dashboard-financial-operations"] .unified-stat-card');
  const financePaths = [
    '/app/reports?report=net-sales&range=today',
    '/app/debts',
    '/app/purchases',
    '/app/expenses',
  ];
  await expect(financeCards).toHaveCount(financePaths.length + 1);
  for (let index = 0; index < financePaths.length; index += 1) {
    await followDashboardAction(financeCards.nth(index), financePaths[index]);
  }
  await financeCards.last().click();
  await expect(page).toHaveURL(/#\/app\/dashboard$/);

  const smartActions = page.locator('.dashboard-smart-actions button');
  const smartActionPaths = [
    ...(Number(values.lowStockCount ?? 0) > 0 ? ['/app/inventory?low_stock_only=true'] : []),
    '/app/debts',
    '/app/sales',
    '/app/inventory',
    '/app/customers',
    '/app/debts',
    '/app/expenses',
  ];
  await expect(smartActions).toHaveCount(smartActionPaths.length);
  for (let index = 0; index < smartActionPaths.length; index += 1) {
    await followDashboardAction(smartActions.nth(index), smartActionPaths[index]);
  }

  await followDashboardAction(page.getByRole('button', { name: 'عرض كل النشاط' }), '/app/activity');

  const attentionButtons = page.locator('.dashboard-attention button');
  const attentionCount = await attentionButtons.count();
  for (let index = 0; index < attentionCount; index += 1) {
    phase = 'navigation';
    await attentionButtons.nth(index).click();
    await expect.poll(() => new URL(page.url()).hash).toMatch(/^#\/app\/(inventory|debts)(\?|$)/);
    await page.goBack();
    await expect(page).toHaveURL(/#\/app\/dashboard$/);
    phase = 'dashboard';
  }

  expect(apiFailures.filter((failure) => failure.phase === 'dashboard')).toEqual([]);
  expect(pageErrors).toEqual([]);
  console.log(JSON.stringify({
    dashboard: 'PASS',
    checkedMetrics: formatted.length + 1,
    recentActivityItems: activity!.length,
    chartRanges: 4,
    checkedActionRoutes: corePaths.length + financePaths.length + smartActionPaths.length + 1 + attentionCount,
    startupApiFailures: apiFailures.filter((failure) => failure.phase === 'startup'),
    dashboardApiFailures: apiFailures.filter((failure) => failure.phase === 'dashboard'),
    pageErrors,
  }));
});
