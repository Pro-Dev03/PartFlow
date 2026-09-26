import { expect, test } from '@playwright/test';
import { getBrowserAuthHeaders } from './helpers/auth';

const email = process.env.E2E_EMAIL;
const password = process.env.E2E_PASSWORD;
const productId = process.env.E2E_BLOCKED_DELETE_PRODUCT_ID;
const apiBase = process.env.E2E_API_BASE_URL ?? 'https://partflow-api.onrender.com/api/v1';
const isLiveSite = process.env.E2E_BASE_URL === 'https://partflow-hpv7.onrender.com';
const isLocalCloudUi = process.env.E2E_ALLOW_LOCAL_CLOUD_UI === 'true';

test.use({ serviceWorkers: 'block' });
test.skip(!email || !password || !productId || (!isLiveSite && !isLocalCloudUi), 'A known product with inventory history and QA credentials are required.');

test('deleting a product with inventory history is rejected without closing confirmation', async ({ page, request }) => {
  await page.addInitScript(() => localStorage.setItem('partflow-connection-mode', 'cloud'));
  await page.goto('/#/login', { waitUntil: 'domcontentloaded' });
  await page.locator('#email').fill(email!);
  await page.locator('#password').fill(password!);
  await page.locator('form button[type="submit"]').click();
  await expect(page).toHaveURL(/#\/app(?:\/|$)/, { timeout: 30_000 });

  const browserHeaders = await getBrowserAuthHeaders(page);
  const cloudToken = browserHeaders['X-PartFlow-Cloud-Token'];
  const headers = cloudToken
    ? { Authorization: `Bearer ${cloudToken}`, 'X-PartFlow-Cloud-Token': cloudToken }
    : browserHeaders;
  const productResponse = await request.get(`${apiBase}/products/${productId}`, { headers });
  expect(productResponse.ok()).toBeTruthy();
  const payload = await productResponse.json();
  const data = payload?.data?.product ?? payload?.data ?? payload;
  const productName = String(data?.name ?? '');
  expect(productName).toBeTruthy();

  await page.goto('/#/app/inventory', { waitUntil: 'domcontentloaded' });
  const search = page.locator('input.pf-inventory-search-input');
  await search.fill(productName);
  const card = page.locator('.compact-product-card:visible').filter({ hasText: productName });
  await expect(card).toHaveCount(1);
  await card.getByRole('button', { name: 'خيارات المنتج' }).click();
  await page.getByRole('menuitem', { name: 'حذف', exact: true }).click({ force: true });

  const dialog = page.getByRole('dialog', { name: 'حذف المنتج' });
  await expect(dialog).toBeVisible();
  const deleteResponse = page.waitForResponse((response) =>
    response.url() === `${apiBase}/products/${productId}` && response.request().method() === 'DELETE',
  );
  await dialog.getByRole('button', { name: 'حذف المنتج', exact: true }).click();
  const response = await deleteResponse;
  expect(response.status()).toBe(409);
  await expect(dialog).toBeVisible();
  await dialog.getByRole('button', { name: 'إلغاء', exact: true }).click();
  await expect(dialog).toBeHidden();
});
