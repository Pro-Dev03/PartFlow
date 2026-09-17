import { test, expect, type Page, type TestInfo } from '@playwright/test';

const API_BASE_URL = process.env.E2E_API_BASE_URL ?? 'http://localhost:8080/api/v1';
const hasLiveEnvironment = Boolean(
  process.env.E2E_EMAIL && process.env.E2E_PASSWORD && process.env.E2E_RUN_RETURN_SUPPLIER_BRIDGE === 'true',
);

test.describe('Customer Return -> Supplier Return bridge', () => {
  test.skip(!hasLiveEnvironment, 'Set E2E_EMAIL, E2E_PASSWORD, and E2E_RUN_RETURN_SUPPLIER_BRIDGE=true for the live environment.');
  test.setTimeout(180_000);

  async function login(page: Page) {
    await page.goto('/login');
    await page.fill('input[type="email"]', process.env.E2E_EMAIL ?? '');
    await page.fill('input[type="password"]', process.env.E2E_PASSWORD ?? '');
    await page.click('button[type="submit"]');
    try {
      await page.waitForURL('**/app', { timeout: 8_000 });
      return;
    } catch {
      await page.evaluate(async ({ email, password }) => {
        const cloudResponse = await fetch('https://partflow-api.onrender.com/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password }),
        });
        const cloudPayload = await cloudResponse.json();
        if (!cloudResponse.ok) throw new Error(`cloud login failed: ${cloudResponse.status}`);
        const cloudToken = cloudPayload.data?.access_token || cloudPayload.access_token;
        const localResponse = await fetch(`${window.location.protocol}//${window.location.hostname}:8080/api/v1/auth/cloud-session`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ cloud_token: cloudToken }),
        });
        const localPayload = await localResponse.json();
        if (!localResponse.ok) throw new Error(`local cloud session failed: ${localResponse.status}`);
        const localToken = localPayload.data?.access_token || localPayload.data?.token || localPayload.access_token || localPayload.token;
        localStorage.setItem('cloud_token', cloudToken);
        localStorage.setItem('auth_token', localToken);
        if (cloudPayload.data?.refresh_token || cloudPayload.refresh_token) localStorage.setItem('cloud_refresh_token', cloudPayload.data?.refresh_token || cloudPayload.refresh_token);
      }, { email: process.env.E2E_EMAIL ?? '', password: process.env.E2E_PASSWORD ?? '' });
      await page.goto('/app');
      await page.waitForURL('**/app', { timeout: 15_000 });
    }
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
      if (!response.ok) throw new Error(`${requestMethod} ${requestPath} -> ${response.status}: ${JSON.stringify(payload)}`);
      return payload as T;
    }, { apiBase: API_BASE_URL, requestMethod: method, requestPath: path, requestBody: body });
  }

  async function apiStatus(page: Page, method: string, path: string, body?: unknown) {
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
      return { status: response.status, body: await response.json().catch(() => ({})) };
    }, { apiBase: API_BASE_URL, requestMethod: method, requestPath: path, requestBody: body });
  }

  function data<T = any>(payload: any): T {
    return (payload?.data ?? payload) as T;
  }
  function fieldSelect(page: Page, label: string) {
    return page.locator('label').filter({ hasText: label }).locator('..').locator('select');
  }

  async function attachEvidence(testInfo: TestInfo, evidence: unknown) {
    await testInfo.attach('customer-return-supplier-return-bridge.json', {
      body: JSON.stringify(evidence, null, 2),
      contentType: 'application/json',
    });
  }

  test('creates the active supplier request from the Customer Returns UI and completes it once', async ({ page }, testInfo) => {
    await login(page);
    const stamp = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
    const barcode = `RET-SUP-${stamp}`;
    const serial = `SERIAL-${stamp}`;
    const evidence: Record<string, unknown> = { stages: [], negative: [] };

    const category = data<any>(await api(page, 'POST', '/categories', { name: `Return Bridge ${stamp}`, is_active: true }));
    const supplier = data<any>(await api(page, 'POST', '/suppliers', { code: `RET-${stamp}`, name: `Return Supplier ${stamp}`, is_active: true }));
    const product = data<any>(await api(page, 'POST', '/products', {
      category_id: category.id ?? category.category?.id,
      preferred_supplier_id: supplier.id,
      name: `Return Bridge Product ${stamp}`,
      sku: `RET-SUP-${stamp}`,
      barcode,
      cost_price: 400,
      selling_price: 800,
      track_serial: true,
      track_individual: true,
    }));
    const productId = product.id ?? product.product?.id;

    const purchaseResponse = data<any>(await api(page, 'POST', '/purchases', {
      supplier_id: supplier.id,
      invoice_number: `RET-${stamp}`,
      purchase_date: new Date().toISOString(),
      items: [{ product_id: productId, barcode, serial_number: serial, quantity: 1, unit_cost: 400, selling_price: 800, condition: 'used' }],
    }));
    const purchase = purchaseResponse.purchase ?? purchaseResponse;
    const purchaseId = purchase.id;
    const purchaseItemId = purchaseResponse.items?.[0]?.id ?? purchase.items?.[0]?.id;
    await api(page, 'POST', `/purchases/${purchaseId}/payment`, { amount: 400, paymentMethod: 'cash' });
    await api(page, 'POST', `/purchases/${purchaseId}/receive`);

    let resolution: any;
    await expect.poll(async () => {
      resolution = data<any>(await api(page, 'GET', `/barcodes/resolve/${encodeURIComponent(barcode)}`));
      return resolution.inventory_item?.serial_number ?? '';
    }, { timeout: 30_000 }).toBe(serial);
    const inventoryItem = resolution.inventory_item;
    expect(inventoryItem.barcode).toBe(barcode);
    expect(inventoryItem.serial_number).toBe(serial);
    const inventoryRecord = data<any>(await api(page, 'GET', `/inventory/items/${inventoryItem.id}`));
    expect(inventoryRecord.supplier_id ?? inventoryRecord.item?.supplier_id ?? inventoryItem.supplier_id).toBe(supplier.id);

    const customer = data<any>(await api(page, 'POST', '/customers', { name: `Return Customer ${stamp}`, phone: `09${stamp}`.slice(-10) }));
    const sale = data<any>(await api(page, 'POST', '/sales', {
      customer_id: customer.id,
      items: [{ product_id: productId, inventory_item_id: inventoryItem.id, quantity: 1, unit_price: 800 }],
      payment_method: 'cash',
      payment_amount: 800,
      total_amount: 800,
    }));
    const saleId = sale.id ?? sale.sale?.id;
    const saleDetails = data<any>(await api(page, 'GET', `/sales/${saleId}`));
    const saleItem = saleDetails.items?.[0] ?? sale.items?.[0] ?? sale.sale?.items?.[0];
    expect(saleItem.inventory_item_id).toBe(inventoryItem.id);

    await page.goto('/#/app/returns/create');
  await fieldSelect(page, 'فاتورة البيع').selectOption(saleId);
  await expect(fieldSelect(page, 'العنصر')).toBeEnabled({ timeout: 15_000 });
  await fieldSelect(page, 'العنصر').selectOption(saleItem.id);
  await fieldSelect(page, 'ماذا يحدث للمنتج بعد الإرجاع؟').selectOption('RETURN_TO_SUPPLIER');
  await fieldSelect(page, 'سبب المرتجع').selectOption('DEFECTIVE');
    await page.getByRole('button', { name: 'حفظ المرتجع' }).click();
    await page.waitForURL('**/app/returns', { timeout: 15_000 });

    const returns = data<any>(await api(page, 'GET', `/returns/sale/${saleId}`));
    const customerReturn = Array.isArray(returns) ? returns[0] : returns.returns?.[0] ?? returns.data?.[0];
    const customerReturnId = customerReturn.id;
    expect(customerReturnId).toBeTruthy();
    await page.goto(`/#/app/returns/${customerReturnId}`);
    await expect(page.getByText('تفاصيل المرتجع')).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText(serial, { exact: false })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText(barcode, { exact: false })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText('إرجاع للمورد', { exact: false }).first()).toBeVisible({ timeout: 15_000 });

    const selectedAgain = await apiStatus(page, 'PUT', `/returns/${customerReturnId}`, {
      item_condition_after_return: 'RETURN_TO_SUPPLIER',
      reason: 'DEFECTIVE',
    });
    expect(selectedAgain.status).toBe(200);
    const selectedThirdTime = await apiStatus(page, 'PUT', `/returns/${customerReturnId}`, {
      item_condition_after_return: 'RETURN_TO_SUPPLIER',
      reason: 'DEFECTIVE',
    });
    expect(selectedThirdTime.status).toBe(200);

    await api(page, 'POST', `/returns/${customerReturnId}/approve`);
    await api(page, 'POST', `/returns/${customerReturnId}/complete`);

    const supplierReturnsPayload = data<any>(await api(page, 'GET', '/supplier-returns'));
    const supplierReturns = Array.isArray(supplierReturnsPayload) ? supplierReturnsPayload : supplierReturnsPayload.data ?? [];
    const linked = supplierReturns.filter((item: any) => item.customer_return_id === customerReturnId);
    expect(linked).toHaveLength(1);
    const linkedReturn = linked[0];
    expect(linkedReturn.sale_id).toBe(saleId);
    expect(linkedReturn.purchase_id).toBe(purchaseId);
    expect(linkedReturn.supplier_id).toBe(supplier.id);
    expect(linkedReturn.inventory_item_id).toBe(inventoryItem.id);
    expect(linkedReturn.barcode).toBe(barcode);
    expect(linkedReturn.serial_number).toBe(serial);
    expect(linkedReturn.status).toBe('PENDING');

    const supplierBeforeComplete = data<any>(await api(page, 'GET', `/suppliers/${supplier.id}`));
    const balanceBeforeComplete = Number(supplierBeforeComplete.current_balance ?? supplierBeforeComplete.outstanding ?? supplierBeforeComplete.supplier?.current_balance ?? 0);

    await page.goto('/#/app/supplier-returns');
    const activeCard = page.locator('div.rounded.border.p-3').filter({ hasText: linkedReturn.return_number }).first();
    await expect(activeCard).toBeVisible({ timeout: 15_000 });
    await expect(activeCard).toContainText(customerReturnId);
    await expect(activeCard).toContainText(saleId);
    await expect(activeCard).toContainText(purchaseId);
    await expect(activeCard).toContainText(supplier.id);
    await expect(activeCard).toContainText(inventoryItem.id);
    await expect(activeCard).toContainText(barcode);
    await expect(activeCard).toContainText(serial);
    await page.screenshot({ path: testInfo.outputPath('active-supplier-return-linked.png'), fullPage: true });

    const duplicateCompletion = await apiStatus(page, 'POST', `/returns/${customerReturnId}/complete`);
    expect(duplicateCompletion.status).toBeGreaterThanOrEqual(400);
    const afterDuplicateAttempt = data<any>(await api(page, 'GET', '/supplier-returns'));
    const sameReturnRows = (Array.isArray(afterDuplicateAttempt) ? afterDuplicateAttempt : afterDuplicateAttempt.data ?? [])
      .filter((item: any) => item.customer_return_id === customerReturnId);
    expect(sameReturnRows).toHaveLength(1);
    evidence.negative = [{ name: 'Repeated Return to Supplier selection/completion', selectedAgain, selectedThirdTime, duplicateCompletion, linkedRows: sameReturnRows.length }];

    const completionResponsePromise = page.waitForResponse((response) => response.url().includes(`/supplier-returns/${linkedReturn.id}/complete`));
    await activeCard.getByRole('button', { name: 'إكمال الإرجاع' }).click();
    const completionResponse = await completionResponsePromise;
    expect(completionResponse.status(), await completionResponse.text()).toBe(200);
    await page.reload();
    await page.getByRole('button', { name: 'الأرشيف', exact: true }).click();
    const completedCard = page.locator('div.rounded.border.p-3').filter({ hasText: linkedReturn.return_number }).first();
    await expect(completedCard).toBeVisible({ timeout: 15_000 });
    await expect(completedCard).toContainText('مكتمل', { timeout: 15_000 });

    const completedPayload = data<any>(await api(page, 'GET', '/supplier-returns'));
    const completedRows = (Array.isArray(completedPayload) ? completedPayload : completedPayload.data ?? [])
      .filter((item: any) => item.customer_return_id === customerReturnId);
    expect(completedRows).toHaveLength(1);
    expect(completedRows[0].status).toBe('COMPLETED');

    const supplierAfterComplete = data<any>(await api(page, 'GET', `/suppliers/${supplier.id}`));
    const balanceAfterComplete = Number(supplierAfterComplete.current_balance ?? supplierAfterComplete.outstanding ?? supplierAfterComplete.supplier?.current_balance ?? 0);
    const supplierLedgerPayload = data<any>(await api(page, 'GET', `/suppliers/${supplier.id}/ledger`));
    const supplierLedger = Array.isArray(supplierLedgerPayload) ? supplierLedgerPayload : supplierLedgerPayload.data ?? [];
    const supplierReturnCredits = supplierLedger.filter((entry: any) =>
      String(entry.transaction_type || '').toUpperCase() === 'SUPPLIER_RETURN'
      && String(entry.reference_id || '') === linkedReturn.id,
    );
    expect(supplierReturnCredits).toHaveLength(1);
    expect(Number(supplierReturnCredits[0].amount)).toBeCloseTo(400, 2);

    await expect(completedCard).toContainText(barcode);
    await expect(completedCard).toContainText(serial);

    evidence.stages = [
      { name: 'DB fixture -> Sale', identity: { purchaseId, supplierId: supplier.id, inventoryItemId: inventoryItem.id, barcode, serial, saleId } },
      { name: 'Customer Returns UI -> Return to Supplier', identity: { customerReturnId, saleId, inventoryItemId: inventoryItem.id, barcode, serial } },
      { name: 'API -> Active Supplier Return', identity: linkedReturn },
      { name: 'Supplier Returns UI -> Completed', identity: { ...completedRows[0], balanceBeforeComplete, balanceAfterComplete, supplierReturnCredits } },
    ];
    await attachEvidence(testInfo, evidence);
  });
});
