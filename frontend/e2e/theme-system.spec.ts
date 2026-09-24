import { expect, test } from '@playwright/test';

const routes = [
  '/app/dashboard', '/app/activity', '/app/sales', '/app/inventory',
  '/app/customers', '/app/customers/1/purchases', '/app/debts', '/app/suppliers',
  '/app/purchases', '/app/purchases/create', '/app/purchases/edit/1', '/app/purchases/1',
  '/app/expenses', '/app/returns', '/app/returns/create', '/app/returns/1',
  '/app/supplier-returns', '/app/reports', '/app/categories', '/app/archive',
  '/app/audit', '/app/settings',
];

test('all sections keep responsive RTL layouts in both themes', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', (error) => errors.push(error.message));
  await page.route('**/api/v1/**', async (route) => {
    const path = new URL(route.request().url()).pathname;
    let data: unknown = [];
    if (path.includes('/settings/regional')) {
      data = {
        profile: { country_code: 'IL', timezone: 'Asia/Jerusalem', locale: 'ar-SA', time_format: '12h', currency: 'ILS' },
        countries: [{ country_code: 'IL', country_name: 'Israel', timezone: 'Asia/Jerusalem' }],
      };
    } else if (path.includes('check-admin') || path.includes('admin-access')) {
      data = { is_admin: false };
    } else if (path.includes('/settings/store_name')) {
      data = { value: 'PartFlow Demo' };
    }
    await route.fulfill({ json: { success: true, data } });
  });
  await page.addInitScript(() => {
    localStorage.setItem('theme', localStorage.getItem('theme') || 'light');
    localStorage.setItem('language', 'ar');
    localStorage.setItem('partflow-connection-mode', 'cloud');
    localStorage.setItem('partflow-cloud-api-url', 'http://127.0.0.1:8082/api/v1');
    localStorage.setItem('auth-storage', JSON.stringify({
      state: {
        isAuthenticated: true,
        sessionVerified: true,
        sessionRestorePending: false,
        cloudVerificationPending: false,
        user: { id: 1, email: 'visual-audit@partflow.test', first_name: 'Visual Audit', subscription_status: 'active' },
        token: 'visual-audit-token',
        cloudToken: 'visual-audit-token',
        isLoading: false,
      },
      version: 0,
    }));
  });
  await page.goto('/');

  for (const viewport of [{ width: 1280, height: 900 }, { width: 390, height: 844 }]) {
    await page.setViewportSize(viewport);
    for (const theme of ['light', 'dark']) {
      await page.evaluate((nextTheme) => {
        localStorage.setItem('theme', nextTheme);
        document.documentElement.classList.toggle('dark', nextTheme === 'dark');
        document.documentElement.classList.toggle('light', nextTheme === 'light');
        document.documentElement.style.colorScheme = nextTheme;
      }, theme);
      for (const path of routes) {
        await page.goto(`/#${path}`);
        await expect(page.locator('#main-content')).toBeVisible({ timeout: 20_000 });
        await expect(page.locator('body')).toContainText(/\S/);
        const layout = await page.evaluate(() => ({
          width: document.documentElement.scrollWidth,
          viewport: window.innerWidth,
          direction: getComputedStyle(document.querySelector('[data-app-main]') || document.body).direction,
          sidebar: (() => { const node = document.querySelector('#app-sidebar'); if (!node) return null; const style = getComputedStyle(node); const rect = node.getBoundingClientRect(); return { position: style.position, visibility: style.visibility, transform: style.transform, left: Math.round(rect.left), right: Math.round(rect.right) }; })(),
          overflow: Array.from(document.querySelectorAll('body *')).map((node) => {
            const rect = node.getBoundingClientRect();
            return { tag: node.tagName, className: typeof node.className === 'string' ? node.className : '', left: Math.round(rect.left), right: Math.round(rect.right), width: Math.round(rect.width) };
          }).filter((node) => node.width > 0 && (node.left < -1 || node.right > window.innerWidth + 1)).slice(0, 12),
        }));
        expect(layout.width, `${path} ${theme} ${viewport.width}: ${JSON.stringify(layout)}`).toBeLessThanOrEqual(layout.viewport + 1);
        expect(layout.direction).toBe('rtl');
        if (path === '/app/sales') await expect(page.locator('button.fixed.bottom-6.right-6.z-50')).toHaveCount(0);
        if (path === '/app/inventory') {
          await page.getByRole('button', { name: 'إضافة', exact: true }).click();
          const dialog = page.getByRole('dialog');
          await expect(dialog).toBeVisible();
          const modal = dialog.locator('.pf-modal-surface');
          const modalRect = await modal.evaluate((node) => node.getBoundingClientRect().width);
          expect(modalRect).toBeLessThanOrEqual(viewport.width);
          await page.keyboard.press('Escape');
          await expect(dialog).toHaveCount(0);
        }
        if (path === '/app/returns/create') {
          await expect(page.locator('#main-content form')).toHaveCount(1);
        }
        if (path === '/app/settings') {
          await expect(page.locator('.settings-regional-card')).toBeVisible();
          const grid = await page.locator('.settings-regional-card .mt-3.grid').evaluate((node) => getComputedStyle(node).gridTemplateColumns);
          if (viewport.width > 640) expect(grid.split(' ').length).toBe(2);
          else expect(grid.split(' ').length).toBe(1);
        }
      }
    }
  }
  expect(errors).toEqual([]);
});
