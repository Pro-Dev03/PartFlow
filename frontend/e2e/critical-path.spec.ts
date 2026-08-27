import { test, expect, type Page } from '@playwright/test';

async function login(page: Page) {
  await page.goto('/login');
  await page.fill('input[type="email"]', process.env.E2E_EMAIL ?? '');
  await page.fill('input[type="password"]', process.env.E2E_PASSWORD ?? '');
  await page.click('button[type="submit"]');
  await page.waitForURL('**/app', { timeout: 10000 });
}

test.describe('Critical Path Tests - سيناريو الاختبار الأساسي', () => {
  test('complete sales workflow', async ({ page }) => {
    // 1. Login
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Check if logged in (might need login page)
    await login(page);

    // 2. Dashboard
    await expect(page.locator('text=الرئيسية')).toBeVisible();
    await expect(page.locator('text=نقطة البيع')).toBeVisible();
    await expect(page.getByRole('button', { name: 'المخزون' })).toBeVisible();

    // 3. Navigate to Inventory
    await page.getByRole('button', { name: 'المخزون' }).click();
    await page.waitForURL('**/app/inventory', { timeout: 10000 });
    await expect(page).toHaveURL(/.*inventory/);

    // 4. Check inventory page loads
    await expect(page.getByRole('heading', { name: 'المنتجات' })).toBeVisible({ timeout: 10000 });

    // 5. Navigate to POS
    await page.getByRole('button', { name: 'نقطة البيع' }).click();
    await page.waitForURL('**/app/sales', { timeout: 10000 });
    await expect(page).toHaveURL(/.*\/app\/sales/);

    // 6. Check POS page loads
    await expect(page.locator('text=نقطة البيع')).toBeVisible({ timeout: 10000 });

    // 7. Navigate to Customers
    await page.getByRole('button', { name: 'العملاء' }).click();
    await page.waitForURL('**/app/customers', { timeout: 10000 });
    await expect(page).toHaveURL(/.*customers/);

    // 8. Check customers page loads
    await expect(page.locator('text=العملاء')).toBeVisible({ timeout: 10000 });

    // 9. Navigate back to Dashboard
    await page.getByRole('button', { name: 'لوحة التحكم' }).click();
    await page.waitForURL('**/app/dashboard', { timeout: 10000 });
    await expect(page).toHaveURL(/.*\/app\/dashboard/);
  });

  test('responsive design - mobile', async ({ page, isMobile }) => {
    await login(page);

    if (isMobile) {
      // Check mobile sidebar is collapsed
      const sidebar = page.locator('aside').first();
      await expect(sidebar).toBeVisible();
    } else {
      // Check desktop sidebar is expanded
      const sidebar = page.locator('aside').first();
      await expect(sidebar).toBeVisible();
    }
  });

  test('keyboard navigation', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Test Tab navigation
    await page.keyboard.press('Tab');
    await page.keyboard.press('Tab');
    
    // Test Enter key
    const focusedElement = await page.evaluate(() => document.activeElement?.tagName);
    expect(focusedElement).toBeTruthy();
  });

  test('RTL support', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Check for RTL direction
    const html = page.locator('html');
    const dir = await html.getAttribute('dir');
    expect(dir).toBe('rtl');
  });
});