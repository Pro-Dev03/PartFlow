import { test, expect, type Page, type TestInfo } from '@playwright/test';
import { getBrowserAuthHeaders } from './helpers/auth';

const API_BASE_URL = process.env.E2E_API_BASE_URL ?? 'http://localhost:8080/api/v1';
const hasLiveEnvironment = Boolean(
  process.env.E2E_EMAIL && process.env.E2E_PASSWORD && process.env.E2E_RUN_OPENING_STOCK_LIFECYCLE === 'true',
);

test.use({ trace: 'on' });

test.describe('Opening Stock supplier-less lifecycle', () => {
  test.skip(!hasLiveEnvironment, 'Set E2E_EMAIL, E2E_PASSWORD, and E2E_RUN_OPENING_STOCK_LIFECYCLE=true for the live environment.');
  test.setTimeout(120_000);

  async function login(page: Page) {
    await page.goto('/login');
    await page.fill('input[type="email"]', process.env.E2E_EMAIL ?? '');
    await page.fill('input[type="password"]', process.env.E2E_PASSWORD ?? '');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/app', { timeout: 15_000 });
  }

  async function api<T = any>(page: Page, method: string, path: string, body?: unknown): Promise<T> {
    const authHeaders = await getBrowserAuthHeaders(page);
    return page.evaluate(async ({ apiBase, requestMethod, requestPath, requestBody, authHeaders }) => {
      const response = await fetch(`${apiBase}${requestPath}`, {
        method: requestMethod,
        headers: {
          'Content-Type': 'application/json',
          ...authHeaders,
        },
        body: requestBody === undefined ? undefined : JSON.stringify(requestBody),
      });
      const payload = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(`${requestMethod} ${requestPath} -> ${response.status}: ${JSON.stringify(payload)}`);
      return payload as T;
    }, { apiBase: API_BASE_URL, requestMethod: method, requestPath: path, requestBody: body, authHeaders });
  }

  function data<T = any>(payload: any): T {
    return (payload?.data ?? payload) as T;
  }

  async function attachEvidence(testInfo: TestInfo, value: unknown) {
    await testInfo.attach('opening-stock-lifecycle-evidence.json', {
      body: JSON.stringify(value, null, 2),
      contentType: 'application/json',
    });
  }

  test('creates Used Individual Opening Stock and completes POS/customer return lifecycle', async ({ page }, testInfo) => {
    await login(page);
    const stamp = `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
    const barcode = `OPEN-E2E-${stamp}`;
    const serial = `OPEN-SERIAL-${stamp}`;
    const networkEvidence: unknown[] = [];
    page.on('response', async (response) => {
      if (/\/products|\/inventory\/opening-stock|\/barcodes\/resolve|\/sales|\/returns/.test(response.url())) {
        networkEvidence.push({ url: response.url(), status: response.status(), body: await response.json().catch(() => null) });
      }
    });
    const evidence: Record<string, unknown> = { barcode, serial, stages: [], network: networkEvidence };

    const productPayload = await api(page, 'POST', '/products', {
      name: `Opening Stock E2E ${stamp}`,
      sku: `OPEN-E2E-${stamp}`,
      barcode,
      cost_price: 25,
      selling_price: 60,
      track_serial: true,
      track_individual: true,
    });
    const product = data<any>(productPayload);
    const productId = product.id ?? product.product?.id;
    expect(productId).toBeTruthy();

    await page.goto('/app/inventory');
    await page.getByRole('button', { name: 'Opening Stock / الرصيد الافتتاحي' }).first().click();
    const openingStockDialog = page.locator('[role="dialog"]').filter({ hasText: 'Opening Stock / الرصيد الافتتاحي' });
    await expect(openingStockDialog).toBeVisible();
    await openingStockDialog.getByLabel('نوع الإدخال').selectOption('individual');
    await openingStockDialog.getByLabel('المنتج').selectOption(productId);
    await openingStockDialog.getByLabel('الكمية').fill('1');
    await openingStockDialog.getByLabel('Barcode / الباركود').fill(barcode);
    await openingStockDialog.getByLabel('Serial / الرقم التسلسلي').fill(serial);
    await openingStockDialog.getByLabel('الحالة').selectOption('USED');
    await openingStockDialog.getByLabel('Business Date / تاريخ الرصيد').fill('2026-09-16');
    await openingStockDialog.getByRole('button', { name: 'حفظ الرصيد الافتتاحي' }).click();
    await expect(page.getByText('تم تسجيل الرصيد الافتتاحي دون إنشاء شراء.')).toBeVisible();

    const openingResolution = data<any>(await api(page, 'GET', `/barcodes/resolve/${encodeURIComponent(barcode)}`));
    const item = openingResolution.inventory_item;
    expect(openingResolution.product.id).toBe(productId);
    expect(item.barcode).toBe(barcode);
    expect(item.serial_number).toBe(serial);
    expect(item.supplier_id ?? null).toBeNull();
    expect(item.condition).toBe('USED');
    evidence.stages = [{ name: 'UI Opening Stock -> Inventory', identity: { productId, itemId: item.id, barcode, serial }, api: openingResolution }];

    await page.getByPlaceholder('مسح أو أدخل الباركود...').fill(barcode);
    await page.getByPlaceholder('مسح أو أدخل الباركود...').press('Enter');
    await expect(page.getByPlaceholder('مسح أو أدخل الباركود...')).toHaveValue('');
    await page.screenshot({ path: testInfo.outputPath('opening-stock-inventory.png'), fullPage: true });

    await page.goto('/app/sales');
    await page.locator('.pos-barcode-input').fill(barcode);
    await page.locator('.pos-barcode-input').press('Enter');
    await expect(page.locator('.pos-cart-sidebar')).toContainText(serial, { timeout: 10_000 });
    await page.screenshot({ path: testInfo.outputPath('opening-stock-pos.png'), fullPage: true });

    const salePayload = data<any>(await api(page, 'POST', '/sales', {
      items: [{ product_id: productId, inventory_item_id: item.id, quantity: 1, unit_price: 60 }],
      payment_method: 'cash',
      payment_amount: 60,
      total_amount: 60,
    }));
    const saleId = salePayload.id ?? salePayload.sale?.id;
    expect(saleId).toBeTruthy();
    const sale = data<any>(await api(page, 'GET', `/sales/${saleId}`));
    const saleItem = sale.items?.[0] ?? sale.sale?.items?.[0];
    expect(saleItem.inventory_item_id).toBe(item.id);
    const soldResolution = data<any>(await api(page, 'GET', `/barcodes/resolve/${encodeURIComponent(barcode)}`));
    expect(soldResolution.inventory_item.id).toBe(item.id);
    expect(soldResolution.inventory_item.status).toBe('SOLD');

    const returnPayload = data<any>(await api(page, 'POST', '/returns', {
      sale_id: saleId,
      return_date: new Date().toISOString(),
      return_type: 'FULL',
      reason: 'CUSTOMER_CHANGED_MIND',
      item_condition_after_return: 'READY_FOR_SALE',
      refund_method: 'CASH',
      items: [{
        sale_item_id: saleItem.id,
        product_id: productId,
        inventory_item_id: item.id,
        serial_number: serial,
        barcode,
        quantity_returned: 1,
        unit_price: 60,
        total_refund_amount: 60,
        returned_condition: 'USED',
        resolution: 'RESTOCK',
      }],
    }));
    const returnId = returnPayload.return?.id ?? returnPayload.id;
    expect(returnId).toBeTruthy();
    await api(page, 'POST', `/returns/${returnId}/approve`);
    await api(page, 'POST', `/returns/${returnId}/complete`);

    const returnedResolution = data<any>(await api(page, 'GET', `/barcodes/resolve/${encodeURIComponent(barcode)}`));
    expect(returnedResolution.inventory_item.id).toBe(item.id);
    expect(returnedResolution.inventory_item.barcode).toBe(barcode);
    expect(returnedResolution.inventory_item.serial_number).toBe(serial);
    expect(returnedResolution.inventory_item.supplier_id ?? null).toBeNull();
    expect(returnedResolution.inventory_item.status).toBe('AVAILABLE');

    await page.goto('/app/returns');
    await expect(page.getByText('المرتجعات')).toBeVisible();
    await page.screenshot({ path: testInfo.outputPath('opening-stock-customer-return.png'), fullPage: true });
    evidence.stages = [...(evidence.stages as any[]), {
      name: 'Inventory -> POS -> Sale -> Customer Return',
      identity: { productId, itemId: item.id, barcode, serial, supplierId: null },
      backend: { saleId, returnId },
      api: { soldResolution, returnedResolution },
      ui: { routes: ['/app/inventory', '/app/sales', '/app/returns'] },
    }];
    await attachEvidence(testInfo, evidence);
  });
});
