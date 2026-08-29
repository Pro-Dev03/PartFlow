import { test, expect, type Page } from '@playwright/test';

async function login(page: Page) {
  await page.goto('/login');
  await page.fill('input[type="email"]', process.env.E2E_EMAIL ?? '');
  await page.fill('input[type="password"]', process.env.E2E_PASSWORD ?? '');
  await page.click('button[type="submit"]');
  await page.waitForURL('**/app', { timeout: 10000 });
}

test.describe('POS hardening scenarios', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.goto('/app/sales');
    await expect(page.locator('.pos-cashier-shell')).toBeVisible({ timeout: 10000 });
  });

  test('keeps barcode scanner and cashier layout ready', async ({ page }) => {
    await expect(page.locator('.pf-barcode-field')).toBeVisible();
    await expect(page.locator('.pos-cashier-products')).toBeVisible();
    await expect(page.locator('.pos-cashier-cart')).toBeVisible();
    await expect(page.locator('.pos-checkout-button')).toBeVisible();
  });

  test('opens manual product recovery flow', async ({ page }) => {
    await page.getByRole('button', { name: 'إضافة يدويًا' }).click();
    await expect(page.getByRole('heading', { name: 'إضافة منتج يدويًا' })).toBeVisible();
    await expect(page.getByPlaceholder('اسم المنتج *')).toBeFocused();
  });

  test('supports cashier keyboard shortcuts without leaving POS', async ({ page }) => {
    await page.keyboard.press('F2');
    await expect(page.locator('.pos-product-search-field')).toBeFocused();
    await page.keyboard.press('Escape');
    await expect(page).toHaveURL(/\/app\/sales$/);
  });
});
