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
    await page.getByRole('button', { name: /^نقطة البيع$/ }).first().click();
    await page.waitForURL('**/app/sales', { timeout: 10000 });
    await expect(page.locator('text=نقطة البيع').first()).toBeVisible({ timeout: 20000 });
  });

  test('keeps barcode scanner and cashier layout ready', async ({ page }) => {
    await expect(page.locator('.pos-barcode-input')).toBeVisible();
    await expect(page.locator('.pos-products-area')).toBeVisible();
    await expect(page.locator('.pos-cart-sidebar')).toBeVisible();
    await expect(page.locator('.checkout-btn')).toBeVisible();
  });

  test('opens manual product recovery flow', async ({ page }) => {
    await page.getByRole('button', { name: /إضافة يدوي/i }).click();
    await expect(page.getByRole('heading', { name: /إضافة منتج يدوي/i })).toBeVisible();
    await expect(page.getByPlaceholder('اسم المنتج *')).toBeFocused();
  });

  test('supports cashier keyboard shortcuts without leaving POS', async ({ page }) => {
    await page.keyboard.press('F2');
    await expect(page.locator('.pos-barcode-input')).toBeFocused();
    await page.keyboard.press('Escape');
    await expect(page).toHaveURL(/\/app\/sales$/);
  });
});
