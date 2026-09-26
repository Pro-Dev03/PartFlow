import { expect, test } from '@playwright/test';
import { getBrowserAuthHeaders } from './helpers/auth';

const email = process.env.E2E_EMAIL;
const password = process.env.E2E_PASSWORD;
const apiBase = process.env.E2E_API_BASE_URL ?? 'https://partflow-api.onrender.com/api/v1';
const isLiveSite = process.env.E2E_BASE_URL === 'https://partflow-hpv7.onrender.com';
const isLocalCloudUi = process.env.E2E_ALLOW_LOCAL_CLOUD_UI === 'true';

test.use({ serviceWorkers: 'block' });
test.skip(!email || !password || (!isLiveSite && !isLocalCloudUi), 'QA credentials and a live or explicitly enabled local cloud UI are required.');

test('inventory product pagination follows the API total across pages', async ({ page, request }) => {
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
  const apiResponse = await request.get(`${apiBase}/products?page=1&per_page=10`, { headers });
  expect(apiResponse.ok()).toBeTruthy();
  const payload = await apiResponse.json();
  const apiData = payload?.data?.data ?? payload?.data ?? payload;
  const apiTotal = Number(apiData?.total);
  expect(Number.isFinite(apiTotal)).toBeTruthy();

  await page.goto('/#/app/inventory', { waitUntil: 'domcontentloaded' });
  const cards = page.locator('.compact-product-card:visible');
  const firstPageCount = Math.min(10, apiTotal);
  await expect(cards).toHaveCount(firstPageCount, { timeout: 20_000 });

  if (apiTotal > 10) {
    const firstPageFirstProduct = await cards.first().innerText();
    const nextPage = page.locator('button:has(svg.lucide-chevron-left)').last();
    await expect(nextPage).toBeEnabled();
    await nextPage.click();
    await expect.poll(async () => cards.first().innerText()).not.toBe(firstPageFirstProduct);
    const secondPageCount = Math.min(10, apiTotal - 10);
    await expect(cards).toHaveCount(secondPageCount);
    console.log(JSON.stringify({ apiTotal, firstPageCount, secondPageCount }));
  }
});
