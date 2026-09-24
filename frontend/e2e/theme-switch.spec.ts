import { expect, test } from '@playwright/test';

const largeProducts = Array.from({ length: 1000 }, (_, index) => ({
  id: `audit-product-${index + 1}`,
  name: `Audit Product ${index + 1}`,
  sku: `AUD-${index + 1}`,
  sellingPrice: 12.5 + index,
  costPrice: 8 + index,
  stock: 20 + (index % 40),
  current_quantity: 20 + (index % 40),
  condition: 'new',
  min_stock_level: 3,
}));

const largeInventory = largeProducts.map((product) => ({
  id: `audit-item-${product.id}`,
  product_id: product.id,
  product_name: product.name,
  condition: 'NEW',
  selling_price: product.sellingPrice,
  purchase_cost: product.costPrice,
  available_quantity: product.stock,
  status: 'AVAILABLE',
  created_at: '2026-01-01T00:00:00Z',
}));

test('measure theme toggle rendering costs with 1,000 inventory records', async ({ page, browser }) => {
  let apiRequestCount = 0;
  const largePayloadRequests = { products: 0, inventory: 0 };
  page.on('request', (request) => {
    if (request.url().includes('/api/v1/')) apiRequestCount += 1;
  });
  await page.route('**/api/v1/**', async (route) => {
    const requestUrl = new URL(route.request().url());
    const pathname = requestUrl.pathname;
    const pageNumber = Number(requestUrl.searchParams.get('page') || 1);
    const pageSize = Number(requestUrl.searchParams.get('per_page') || 10);
    const slicePage = <T,>(items: T[]) => items.slice((pageNumber - 1) * pageSize, pageNumber * pageSize);
    if (pathname.endsWith('/products')) {
      if (pageSize >= 1000) largePayloadRequests.products += 1;
      return route.fulfill({ json: { success: true, data: { products: slicePage(largeProducts) }, meta: { total: largeProducts.length } } });
    }
    if (pathname.endsWith('/inventory/items-with-supplier')) {
      if (pageSize >= 1000) largePayloadRequests.inventory += 1;
      return route.fulfill({ json: { success: true, data: { items: slicePage(largeInventory) }, meta: { total: largeInventory.length } } });
    }
    return route.fulfill({ json: { success: true, data: [] } });
  });
  await page.addInitScript(() => {
    localStorage.setItem('theme', 'light');
    localStorage.setItem('language', 'ar');
    localStorage.setItem('partflow-connection-mode', 'cloud');
    localStorage.setItem('partflow-cloud-api-url', 'http://127.0.0.1:8082/api/v1');
    localStorage.setItem('auth-storage', JSON.stringify({ state: { isAuthenticated: true, sessionVerified: true, sessionRestorePending: false, cloudVerificationPending: false, user: { id: 1, email: 'audit@test', first_name: 'Audit', subscription_status: 'active' }, token: 'audit', cloudToken: 'audit', isLoading: false }, version: 0 }));
    (window as any).__firstThemeFrame = undefined;
    const sampleFirstFrame = () => {
      if ((!document.documentElement.classList.contains('dark') && !document.documentElement.classList.contains('light')) || !document.body) {
        requestAnimationFrame(sampleFirstFrame);
        return;
      }
      const overlay = document.getElementById('loading-overlay');
      (window as any).__firstThemeFrame = {
        rootClass: document.documentElement.className,
        bodyBackground: getComputedStyle(document.body).backgroundColor,
        overlayBackground: overlay ? getComputedStyle(overlay).backgroundColor : null,
      };
    };
    requestAnimationFrame(sampleFirstFrame);
  });
  const client = await page.context().newCDPSession(page);
  await client.send('Performance.enable');
  await page.goto('/#/app/dashboard');
  await expect(page.locator('#main-content')).toBeVisible({ timeout: 20_000 });
  await page.evaluate(() => { window.location.hash = '/app/inventory'; });
  await expect(page.locator('#main-content')).toBeVisible({ timeout: 20_000 });
  await expect(page.locator('button[title="Toggle theme"]')).toBeVisible();
  await expect(page.locator('#main-content')).toContainText('Audit Product 1', { timeout: 20_000 });
  await expect(page.locator('#main-content .compact-product-card')).toHaveCount(10);
  expect(largePayloadRequests.products).toBeGreaterThan(0);
  expect(largePayloadRequests.inventory).toBeGreaterThan(0);
  await page.locator('#main-content button:has(svg.lucide-chevron-left)').last().click();
  await expect(page.locator('#main-content .compact-product-card').first()).toContainText('Audit Product 11');
  await expect(page.locator('#main-content .compact-product-card')).toHaveCount(10);
  await page.waitForTimeout(1200);
  const startup = await page.evaluate(() => (window as any).__firstThemeFrame);
  expect(startup.rootClass).toContain('light');
  expect(startup.bodyBackground).toBe('rgb(247, 245, 240)');
  expect(startup.overlayBackground).toBe('rgb(247, 245, 240)');
  const darkContext = await browser.newContext({ viewport: { width: 1280, height: 900 } });
  const darkPage = await darkContext.newPage();
  await darkPage.route('**/api/v1/**', async (route) => route.fulfill({ json: { success: true, data: [] } }));
  await darkPage.addInitScript(() => {
    localStorage.setItem('theme', 'dark');
    localStorage.setItem('language', 'ar');
    localStorage.setItem('partflow-connection-mode', 'cloud');
    localStorage.setItem('partflow-cloud-api-url', 'http://127.0.0.1:8082/api/v1');
    localStorage.setItem('auth-storage', JSON.stringify({ state: { isAuthenticated: true, sessionVerified: true, sessionRestorePending: false, cloudVerificationPending: false, user: { id: 1, email: 'dark-audit@test', first_name: 'Dark Audit', subscription_status: 'active' }, token: 'audit', cloudToken: 'audit', isLoading: false }, version: 0 }));
    (window as any).__firstThemeFrame = undefined;
    const sampleFirstFrame = () => {
      if ((!document.documentElement.classList.contains('dark') && !document.documentElement.classList.contains('light')) || !document.body) {
        requestAnimationFrame(sampleFirstFrame);
        return;
      }
      const overlay = document.getElementById('loading-overlay');
      (window as any).__firstThemeFrame = {
        rootClass: document.documentElement.className,
        bodyBackground: getComputedStyle(document.body).backgroundColor,
        overlayBackground: overlay ? getComputedStyle(overlay).backgroundColor : null,
      };
    };
    requestAnimationFrame(sampleFirstFrame);
  });
  const darkUrl = new URL(page.url());
  darkUrl.hash = '/app/dashboard';
  await darkPage.goto(darkUrl.toString());
  await expect(darkPage.locator('#main-content')).toBeVisible({ timeout: 20_000 });
  const darkStartup = await darkPage.evaluate(() => (window as any).__firstThemeFrame);
  expect(darkStartup.rootClass).toContain('dark');
  expect(darkStartup.bodyBackground).toBe('rgb(17, 23, 33)');
  expect(darkStartup.overlayBackground).toBe('rgb(15, 23, 42)');
  await darkContext.close();
  await page.evaluate(() => {
    (window as any).__themeAudit = { frames: [], changedDomNodes: 0, layoutShifts: [] };
    const main = document.querySelector('#main-content')!;
    new MutationObserver((records) => {
      (window as any).__themeAudit.changedDomNodes += records.reduce((count, record) => count + record.addedNodes.length + record.removedNodes.length, 0);
    }).observe(main, { childList: true, subtree: true });
    new PerformanceObserver((entries) => {
      entries.getEntries().forEach((entry: any) => {
        (window as any).__themeAudit.layoutShifts.push(entry.value);
      });
    }).observe({ type: 'layout-shift', buffered: false });
    document.addEventListener('click', (event) => {
      const target = event.target as Element | null;
      if (!target?.closest('button[title="Toggle theme"]')) return;
      const started = performance.now();
      requestAnimationFrame(() => {
        const rect = main.getBoundingClientRect();
        const style = getComputedStyle(main);
        (window as any).__themeAudit.frames.push({
          clickToFrameMs: +(performance.now() - started).toFixed(2),
          rootClass: document.documentElement.className,
          background: getComputedStyle(document.body).backgroundColor,
          width: rect.width,
          height: rect.height,
          scrollWidth: document.documentElement.scrollWidth,
          transitionDuration: style.transitionDuration,
        });
      });
    }, true);
  });
  const before = await client.send('Performance.getMetrics');
  const apiRequestsBeforeToggles = apiRequestCount;
  for (let i = 0; i < 4; i++) {
    await page.locator('button[title="Toggle theme"]').click();
    await page.waitForTimeout(100);
  }
  const after = await client.send('Performance.getMetrics');
  const audit = await page.evaluate(() => (window as any).__themeAudit);
  const pick = (metrics: any[], name: string) => metrics.find((metric) => metric.name === name)?.value ?? 0;
  console.log(JSON.stringify({
    toggles: audit.frames.map((frame: any) => ({ clickToFrameMs: frame.clickToFrameMs, theme: frame.rootClass.includes('dark') ? 'dark' : 'light', background: frame.background, width: frame.width, height: frame.height, scrollWidth: frame.scrollWidth })),
    changedDomNodes: audit.changedDomNodes,
    layoutShiftValues: audit.layoutShifts,
    chrome: {
      taskDurationSeconds: +(pick(after.metrics, 'TaskDuration') - pick(before.metrics, 'TaskDuration')).toFixed(4),
      layoutDurationSeconds: +(pick(after.metrics, 'LayoutDuration') - pick(before.metrics, 'LayoutDuration')).toFixed(4),
      recalcStyleDurationSeconds: +(pick(after.metrics, 'RecalcStyleDuration') - pick(before.metrics, 'RecalcStyleDuration')).toFixed(4),
      layoutCount: pick(after.metrics, 'LayoutCount') - pick(before.metrics, 'LayoutCount'),
      recalcStyleCount: pick(after.metrics, 'RecalcStyleCount') - pick(before.metrics, 'RecalcStyleCount'),
      jsHeapUsedDeltaMb: +((pick(after.metrics, 'JSHeapUsedSize') - pick(before.metrics, 'JSHeapUsedSize')) / (1024 * 1024)).toFixed(2),
    },
  }, null, 2));
  expect(audit.frames.map((frame: any) => frame.rootClass.includes('dark'))).toEqual([true, false, true, false]);
  expect(audit.changedDomNodes).toBe(0);
  expect(audit.layoutShifts).toEqual([]);
  expect(apiRequestCount - apiRequestsBeforeToggles).toBe(0);

  await page.setViewportSize({ width: 390, height: 844 });
  await page.waitForTimeout(700);
  await page.evaluate(() => {
    (window as any).__themeAudit = { frames: [], changedDomNodes: 0, layoutShifts: [] };
  });
  const mobileBefore = await client.send('Performance.getMetrics');
  const apiRequestsBeforeMobileToggles = apiRequestCount;
  for (let i = 0; i < 4; i++) {
    await page.locator('button[title="Toggle theme"]').click();
    await page.waitForTimeout(100);
  }
  const mobileAfter = await client.send('Performance.getMetrics');
  const mobileAudit = await page.evaluate(() => (window as any).__themeAudit);
  console.log(JSON.stringify({
    mobileToggles: mobileAudit.frames.map((frame: any) => ({ clickToFrameMs: frame.clickToFrameMs, theme: frame.rootClass.includes('dark') ? 'dark' : 'light', background: frame.background, width: frame.width, height: frame.height, scrollWidth: frame.scrollWidth })),
    mobileChangedDomNodes: mobileAudit.changedDomNodes,
    mobileLayoutShifts: mobileAudit.layoutShifts,
    mobileChrome: {
      taskDurationSeconds: +(pick(mobileAfter.metrics, 'TaskDuration') - pick(mobileBefore.metrics, 'TaskDuration')).toFixed(4),
      layoutDurationSeconds: +(pick(mobileAfter.metrics, 'LayoutDuration') - pick(mobileBefore.metrics, 'LayoutDuration')).toFixed(4),
      recalcStyleDurationSeconds: +(pick(mobileAfter.metrics, 'RecalcStyleDuration') - pick(mobileBefore.metrics, 'RecalcStyleDuration')).toFixed(4),
      layoutCount: pick(mobileAfter.metrics, 'LayoutCount') - pick(mobileBefore.metrics, 'LayoutCount'),
      recalcStyleCount: pick(mobileAfter.metrics, 'RecalcStyleCount') - pick(mobileBefore.metrics, 'RecalcStyleCount'),
      jsHeapUsedDeltaMb: +((pick(mobileAfter.metrics, 'JSHeapUsedSize') - pick(mobileBefore.metrics, 'JSHeapUsedSize')) / (1024 * 1024)).toFixed(2),
    },
  }, null, 2));
  expect(mobileAudit.frames.map((frame: any) => frame.rootClass.includes('dark'))).toEqual([true, false, true, false]);
  expect(mobileAudit.frames.every((frame: any) => frame.scrollWidth <= 391)).toBeTruthy();
  expect(mobileAudit.changedDomNodes).toBe(0);
  expect(mobileAudit.layoutShifts).toEqual([]);
  expect(apiRequestCount - apiRequestsBeforeMobileToggles).toBe(0);
});
