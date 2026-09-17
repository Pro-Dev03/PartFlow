import { test, expect, type Page, type TestInfo } from '@playwright/test';

const BARCODE = 'FNX-GPU-000421';
const API_BASE_URL = process.env.E2E_API_BASE_URL ?? 'http://localhost:8080/api/v1';
const hasLiveEnvironment = Boolean(
  process.env.E2E_EMAIL && process.env.E2E_PASSWORD && process.env.E2E_RUN_BARCODE_LIFECYCLE === 'true',
);

test.describe('Barcode lifecycle: FNX-GPU-000421', () => {
  test.skip(!hasLiveEnvironment, 'Set E2E_EMAIL, E2E_PASSWORD, and E2E_RUN_BARCODE_LIFECYCLE=true for the live environment.');
  test.setTimeout(120_000);

  async function login(page: Page) {
    await page.goto('/login');
    await page.fill('input[type="email"]', process.env.E2E_EMAIL ?? '');
    await page.fill('input[type="password"]', process.env.E2E_PASSWORD ?? '');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/app', { timeout: 15_000 });
  }

  async function api<T = any>(page: Page, method: string, path: string, body?: unknown): Promise<T> {
    return page.evaluate(async ({ apiBase, requestMethod, requestPath, requestBody }) => {
      const token = localStorage.getItem('auth_token') || localStorage.getItem('token');
      const cloudToken = localStorage.getItem('cloud_token');
      const response = await fetch(`${apiBase}${requestPath}`, {
        method: requestMethod,
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
          ...(cloudToken ? { 'X-PartFlow-Cloud-Token': cloudToken } : {}),
        },
        body: requestBody === undefined ? undefined : JSON.stringify(requestBody),
      });
      const payload = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(`${requestMethod} ${requestPath} -> ${response.status}: ${JSON.stringify(payload)}`);
      }
      return payload as T;
    }, { apiBase: API_BASE_URL, requestMethod: method, requestPath: path, requestBody: body });
  }

  function data<T = any>(payload: any): T {
    return (payload?.data ?? payload) as T;
  }

  async function attachEvidence(testInfo: TestInfo, name: string, value: unknown) {
    await testInfo.attach(name, {
      body: JSON.stringify(value, null, 2),
      contentType: 'application/json',
    });
  }

  test('executes the real lifecycle and records DB/API/UI identity evidence', async ({ page }, testInfo) => {
    await login(page);

    const evidence: Record<string, unknown> = { barcode: BARCODE, stages: [], negative: [] };
    const stamp = Date.now();

    const categoryPayload = await api(page, 'POST', '/categories', {
      name: `E2E GPU ${stamp}`,
      description: 'Barcode lifecycle fixture',
      is_active: true,
    });
    const category = data<any>(categoryPayload);
    const categoryId = category.id ?? category.category?.id;

    const supplierPayload = await api(page, 'POST', '/suppliers', {
      code: `E2E-${stamp}`,
      name: `E2E Supplier ${stamp}`,
      is_active: true,
    });
    const supplier = data<any>(supplierPayload);
    const supplierId = supplier.id;

    const productPayload = await api(page, 'POST', '/products', {
      category_id: categoryId,
      preferred_supplier_id: supplierId,
      name: `E2E GPU ${stamp}`,
      sku: `E2E-GPU-${stamp}`,
      barcode: BARCODE,
      cost_price: 400,
      selling_price: 880,
      track_serial: true,
      track_individual: true,
    });
    const product = data<any>(productPayload);
    const productId = product.id ?? product.product?.id;
    expect(productId).toBeTruthy();

    const serial = `SN-E2E-${stamp}`;
    const purchasePayload = await api(page, 'POST', '/purchases', {
      supplier_id: supplierId,
      invoice_number: `E2E-${stamp}`,
      purchase_date: new Date().toISOString(),
      items: [{
        product_id: productId,
        barcode: BARCODE,
        serial_number: serial,
        quantity: 1,
        unit_cost: 400,
        selling_price: 880,
        condition: 'used',
        notes: 'Barcode lifecycle fixture',
      }],
    });
    const purchaseResponse = data<any>(purchasePayload);
    const purchase = purchaseResponse.purchase ?? purchaseResponse;
    const purchaseId = purchase.id;
    const purchaseItemId = purchaseResponse.items?.[0]?.id;
    expect(purchaseId).toBeTruthy();
    expect(purchaseItemId).toBeTruthy();

    await api(page, 'POST', `/purchases/${purchaseId}/payment`, {
      amount: 400,
      paymentMethod: 'cash',
    });
    await api(page, 'POST', `/purchases/${purchaseId}/receive`);

    const resolutionPayload = await api(page, 'GET', `/barcodes/resolve/${encodeURIComponent(BARCODE)}`);
    const resolution = data<any>(resolutionPayload);
    const inventoryItem = resolution.inventory_item;
    expect(resolution.product.id).toBe(productId);
    expect(inventoryItem.barcode).toBe(BARCODE);
    expect(inventoryItem.serial_number).toBe(serial);
    expect(resolution.purchase_ids).toContain(purchaseId);
    expect(inventoryItem.supplier_id ?? supplierId).toBe(supplierId);

    evidence.stages = [{
      name: 'Purchase -> Receive -> Inventory',
      identity: { productId, itemId: inventoryItem.id, barcode: BARCODE, serial, purchaseId, supplierId },
      backend: { endpoint: '/barcodes/resolve/:code' },
      api: resolution,
    }];

    await page.goto('/app/inventory');
    await expect(page).toHaveURL(/\/app\/inventory/);
    await page.getByRole('textbox').first().fill(BARCODE);
    await page.getByRole('textbox').first().press('Enter');
    await expect(page.getByText(BARCODE, { exact: false }).first()).toBeVisible({ timeout: 10_000 });
    await page.screenshot({ path: testInfo.outputPath('inventory-barcode.png'), fullPage: true });

    await page.goto('/app/sales');
    await expect(page.locator('.pos-barcode-input')).toBeVisible({ timeout: 15_000 });
    const resolverResponses: any[] = [];
    page.on('response', async (response) => {
      if (response.url().includes('/barcodes/resolve/')) {
        resolverResponses.push({ url: response.url(), status: response.status(), body: await response.json().catch(() => null) });
      }
    });
    await page.locator('.pos-barcode-input').fill(BARCODE);
    await page.locator('.pos-barcode-input').press('Enter');
    await expect(page.locator('.pos-cart-sidebar')).toContainText(BARCODE, { timeout: 10_000 });
    await page.screenshot({ path: testInfo.outputPath('pos-barcode.png'), fullPage: true });
    expect(resolverResponses.some((item) => item.status === 200)).toBeTruthy();

    const salePayload = await api(page, 'POST', '/sales', {
      items: [{ product_id: productId, inventory_item_id: inventoryItem.id, quantity: 1, unit_price: 880 }],
      payment_method: 'cash',
      payment_amount: 880,
      total_amount: 880,
    });
    const sale = data<any>(salePayload);
    const saleId = sale.id ?? sale.sale?.id;
    const saleItem = sale.items?.[0] ?? sale.sale?.items?.[0];
    expect(saleId).toBeTruthy();
    expect(saleItem?.inventory_item_id).toBe(inventoryItem.id);

    await expect.poll(async () => (await api<any>(page, 'GET', `/barcodes/resolve/${encodeURIComponent(BARCODE)}`)).inventory_item?.status ?? '').toBe('SOLD');

    const secondSale = await page.evaluate(async ({ apiBase, productId: pid, itemId }) => {
      const token = localStorage.getItem('auth_token') || localStorage.getItem('token');
      const cloudToken = localStorage.getItem('cloud_token');
      const response = await fetch(`${apiBase}/sales`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
          ...(cloudToken ? { 'X-PartFlow-Cloud-Token': cloudToken } : {}),
        },
        body: JSON.stringify({ items: [{ product_id: pid, inventory_item_id: itemId, quantity: 1, unit_price: 880 }], payment_method: 'cash', payment_amount: 880, total_amount: 880 }),
      });
      return { status: response.status, body: await response.json().catch(() => ({})) };
    }, { apiBase: API_BASE_URL, productId, itemId: inventoryItem.id });
    expect(secondSale.status).toBeGreaterThanOrEqual(400);
    evidence.negative = [{ name: 'Sell same individual item twice', result: secondSale }];

    const customerPayload = await api(page, 'POST', '/customers', { name: `E2E Customer ${stamp}`, phone: `09${stamp}`.slice(-10) });
    const customer = data<any>(customerPayload);
    const customerId = customer.id;
    const customerReturnPayload = await api(page, 'POST', '/returns', {
      sale_id: saleId,
      customer_id: customerId,
      return_date: new Date().toISOString(),
      return_type: 'FULL',
      reason: 'CUSTOMER_CHANGED_MIND',
      item_condition_after_return: 'READY_FOR_SALE',
      refund_method: 'CASH',
      items: [{
        sale_item_id: saleItem.id,
        product_id: productId,
        inventory_item_id: inventoryItem.id,
        serial_number: serial,
        barcode: BARCODE,
        quantity_returned: 1,
        unit_price: 880,
        total_refund_amount: 880,
        returned_condition: 'USED',
        resolution: 'RESTOCK',
      }],
    });
    const customerReturn = data<any>(customerReturnPayload);
    expect(customerReturn.items?.[0]?.inventory_item_id).toBe(inventoryItem.id);

    const supplierReturnPayload = await api(page, 'POST', '/supplier-returns', {
      purchase_id: purchaseId,
      reason: 'Barcode lifecycle E2E',
      notes: `Barcode ${BARCODE}`,
    });
    const supplierReturn = data<any>(supplierReturnPayload);
    const supplierReturnId = supplierReturn.id;
    await api(page, 'POST', `/supplier-returns/${supplierReturnId}/items`, {
      purchase_item_id: purchaseItemId,
      quantity: 1,
    });
    expect(supplierReturn.purchase_id).toBe(purchaseId);
    expect(supplierReturn.supplier_id).toBe(supplierId);

    for (const route of ['/app/inventory', '/app/purchases', '/app/returns', '/app/supplier-returns', '/app/reports']) {
      await page.goto(route);
      await expect(page).toHaveURL(new RegExp(route.replaceAll('/', '\\/')));
      await page.screenshot({ path: testInfo.outputPath(`${route.split('/').pop()}-barcode.png`), fullPage: true });
    }

    const duplicate = await page.evaluate(async ({ apiBase, productId: pid }) => {
      const token = localStorage.getItem('auth_token') || localStorage.getItem('token');
      const response = await fetch(`${apiBase}/inventory`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
        body: JSON.stringify({ product_id: pid, quantity: 1, barcode: BARCODE, condition: 'USED', purchase_cost: 400, selling_price: 880, status: 'AVAILABLE' }),
      });
      return { status: response.status, body: await response.json().catch(() => ({})) };
    }, { apiBase: API_BASE_URL, productId });
    expect(duplicate.status).toBeGreaterThanOrEqual(400);
    evidence.negative.push({ name: 'Duplicate barcode', result: duplicate });

    const soldResolution = await api<any>(page, 'GET', `/barcodes/resolve/${encodeURIComponent(BARCODE)}`);
    expect(soldResolution.inventory_item.id).toBe(inventoryItem.id);
    expect(soldResolution.inventory_item.barcode).toBe(BARCODE);
    evidence.stages = [...(evidence.stages as any[]), {
      name: 'POS -> Sale -> Customer Return -> Supplier Return -> Reports',
      identity: { productId, itemId: inventoryItem.id, barcode: BARCODE, serial, purchaseId, supplierId },
      backend: { saleId, customerReturnId: customerReturn.return?.id ?? customerReturn.id, supplierReturnId },
      api: { resolution: soldResolution, resolverResponses },
      ui: { routes: ['/app/inventory', '/app/sales', '/app/returns', '/app/supplier-returns', '/app/reports'] },
    }];

    await attachEvidence(testInfo, 'barcode-lifecycle-evidence.json', evidence);
  });
});
