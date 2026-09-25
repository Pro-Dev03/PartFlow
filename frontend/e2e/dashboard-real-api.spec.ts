import { expect, test } from '@playwright/test';

const apiBase = process.env.E2E_LOCAL_API_BASE_URL;

test.skip(!apiBase, 'Set E2E_LOCAL_API_BASE_URL to run this live API integration check.');

test('dashboard renders and reconciles its empty-store metrics against the real API', async ({ page, request }) => {
  const email = process.env.E2E_EMAIL ?? 'owner@partflow.com';
  const password = process.env.E2E_PASSWORD;
  expect(password, 'E2E_PASSWORD is required for the temporary QA account').toBeTruthy();

  const loginResponse = await request.post(`${apiBase}/auth/login`, {
    data: { email, password },
  });
  expect(loginResponse.status(), 'local SQLite login should issue a real JWT').toBe(200);
  const login = await loginResponse.json();
  const accessToken = login?.data?.token ?? login?.data?.access_token;
  const userId = login?.data?.user?.id;
  expect(accessToken).toBeTruthy();
  expect(userId).toBeTruthy();

  const cloudToken = 'dashboard-qa-cloud-session';
  const apiResponses: Array<{ path: string; status: number }> = [];
  const runtimeErrors: string[] = [];
  let dashboardStats: Record<string, unknown> | undefined;

  page.on('pageerror', (error) => runtimeErrors.push(error.message));
  page.on('response', async (response) => {
    const url = new URL(response.url());
    if (!url.pathname.startsWith('/api/v1/')) return;
    apiResponses.push({ path: `${url.pathname}${url.search}`, status: response.status() });
    if (url.pathname === '/api/v1/dashboard/stats' && response.ok()) {
      const payload = await response.json().catch(() => undefined);
      dashboardStats = payload?.data;
    }
  });

  await page.route('https://partflow-api.onrender.com/api/v1/**', async (route) => {
    const source = new URL(route.request().url());
    if (source.pathname === '/api/v1/auth/validate') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            token: cloudToken,
            user: {
              id: userId,
              email,
              is_active: true,
              subscription_status: 'active',
              subscription_expires_at: '2099-12-31T23:59:59Z',
            },
          },
        }),
      });
    }

    const response = await route.fetch({
      url: `${apiBase}${source.pathname.slice('/api/v1'.length)}${source.search}`,
    });
    return route.fulfill({ response });
  });

  await page.addInitScript(({ token, id, userEmail }) => {
    localStorage.setItem('partflow-connection-mode', 'cloud');
    localStorage.setItem('auth-storage', JSON.stringify({
      state: {
        isAuthenticated: true,
        sessionVerified: true,
        sessionRestorePending: false,
        cloudVerificationPending: false,
        user: { id, email: userEmail, name: 'Dashboard QA Owner', role: 'owner' },
        token,
        cloudToken: 'dashboard-qa-cloud-session',
        refreshTokenValue: null,
      },
      version: 0,
    }));
  }, { token: accessToken, id: userId, userEmail: email });

  await page.goto('/#/app/dashboard');
  await expect(page.locator('#main-content')).toBeVisible();
  await expect(page.locator('.dashboard-attention')).toBeVisible();
  await expect(page.locator('.dashboard-smart-actions')).toBeVisible();
  await expect(page.locator('#dashboard-core-metrics')).toBeVisible();
  await expect(page.locator('.dashboard-performance-card')).toBeVisible();
  await expect.poll(() => dashboardStats).toBeTruthy();

  expect(dashboardStats).toMatchObject({
    total_sales: 0,
    total_purchases: 0,
    total_expenses: 0,
    total_products: 0,
    total_customers: 0,
    total_suppliers: 0,
    todaySales: 0,
    todayProfit: 0,
    outstandingDebts: 0,
    lowStockCount: 0,
  });
  expect(apiResponses.filter((response) => response.status >= 400)).toEqual([]);

  const rangeButtons = page.locator('.dashboard-performance-range-button');
  await expect(rangeButtons).toHaveCount(4);
  for (let index = 0; index < 4; index += 1) {
    await rangeButtons.nth(index).click();
    await expect(page.locator('.dashboard-performance-card')).toBeVisible();
  }

  expect(runtimeErrors).toEqual([]);
});
