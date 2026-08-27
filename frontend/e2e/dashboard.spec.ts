import { test, expect, type Page } from '@playwright/test';

async function login(page: Page) {
  await page.goto('/login');
  await page.fill('input[type="email"]', process.env.E2E_EMAIL ?? '');
  await page.fill('input[type="password"]', process.env.E2E_PASSWORD ?? '');
  await page.click('button[type="submit"]');
  await page.waitForURL('**/app', { timeout: 10000 });
}

test.describe('Dashboard Tests', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('should load dashboard successfully', async ({ page }) => {
    // Wait for dashboard to load
    await expect(page).toHaveTitle(/PartFlow/);
    
    // Check for key dashboard elements
    await expect(page.locator('text=الرئيسية')).toBeVisible();
    await expect(page.locator('text=نقطة البيع')).toBeVisible();
    await expect(page.getByRole('button', { name: 'المخزون' })).toBeVisible();
  });

  test('should display attention section', async ({ page }) => {
    // Check for attention section
    const attentionSection = page.getByText('يحتاج انتباهك', { exact: false }).first();
    await expect(attentionSection).toBeVisible({ timeout: 10000 });
  });

  test('should display smart actions', async ({ page }) => {
    // Check for smart actions
    const smartActions = page.getByText('العمليات الذكية', { exact: false }).first();
    await expect(smartActions).toBeVisible({ timeout: 10000 });
  });

  test('should navigate to POS from dashboard', async ({ page }) => {
    // Click on POS button
    const posButton = page.getByRole('button', { name: 'نقطة البيع' }).first();
    await posButton.click();
    
    // Wait for navigation
    await page.waitForURL('**/app/sales', { timeout: 10000 });
    await expect(page).toHaveURL(/.*\/app\/sales/);
  });

  test('should navigate to inventory from dashboard', async ({ page }) => {
    // Click on inventory button
    const inventoryButton = page.getByRole('button', { name: 'المخزون' }).first();
    await inventoryButton.click();
    
    // Wait for navigation
    await page.waitForURL('**/app/inventory', { timeout: 10000 });
    await expect(page).toHaveURL(/.*\/app\/inventory/);
  });
});