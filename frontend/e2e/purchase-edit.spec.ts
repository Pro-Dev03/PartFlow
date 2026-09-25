import { expect, test } from '@playwright/test';

test('purchase edit supports next-field navigation, saving, payment, barcode and receiving', async ({ page }) => {
  const browserErrors: string[] = [];
  const failedRequests: string[] = [];
  const observedCalls: string[] = [];
  let purchase = {
    id: 'purchase-1',
    supplier_id: 'supplier-1',
    invoice_number: 'INV-001',
    purchase_date: '2026-09-20T12:00:00Z',
    expected_delivery_date: '2026-09-25T12:00:00Z',
    status: 'pending',
    total_amount: 20,
    paid_amount: 0,
    remaining_amount: 20,
    notes: '',
  };
  let purchaseItems = [{
    id: 'item-1',
    purchase_id: 'purchase-1',
    product_id: 'product-1',
    product_name: 'فلتر زيت',
    quantity: 2,
    unit_cost: 10,
    selling_price: 15,
    category_id: '',
    condition: 'new',
  }];
  let updatePayload: any;
  let paymentPayload: any;
  let receiveCalled = false;

  const json = (data: unknown, status = 200) => ({
    status,
    contentType: 'application/json',
    body: JSON.stringify(data),
  });

  page.on('pageerror', (error) => browserErrors.push(error.message));
  page.on('console', (message) => {
    if (message.type() === 'error' && !message.text().includes('Failed to load resource')) {
      browserErrors.push(message.text());
    }
  });
  page.on('requestfailed', (request) => failedRequests.push(request.url()));

  await page.addInitScript(() => {
    localStorage.setItem('partflow-connection-mode', 'local');
    localStorage.setItem('auth-storage', JSON.stringify({
      state: {
        isAuthenticated: true,
        sessionVerified: true,
        user: { id: 'test-owner', name: 'Test Owner', role: 'admin', email: 'test@example.invalid' },
        token: 'local-test-token',
        cloudToken: 'cloud-test-token',
        refreshTokenValue: null,
      },
      version: 0,
    }));
  });

  await page.route('**/api/v1/**', async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname.replace('/api/v1', '');
    const method = request.method();
    observedCalls.push(method + ' ' + path + url.search);

    if (url.hostname === 'partflow-api.onrender.com') {
      if (path === '/auth/validate') return route.fulfill(json({ success: true, data: { valid: true } }));
      if (path === '/auth/refresh') return route.fulfill(json({ success: true, data: { access_token: 'cloud-test-token' } }));
      return route.fulfill(json({ success: true, data: {} }));
    }
    if (path === '/suppliers' && method === 'GET') {
      return route.fulfill(json({ success: true, data: [{ id: 'supplier-1', name: 'مورد تجريبي', phone: '0599000000' }], meta: { total: 1 } }));
    }
    if (path === '/categories' && method === 'GET') return route.fulfill(json({ success: true, data: [] }));
    if (path === '/purchases/purchase-1' && method === 'GET') {
      return route.fulfill(json({ success: true, data: { purchase: { ...purchase }, items: purchaseItems.map((item) => ({ ...item })) } }));
    }
    if (path === '/purchases/purchase-1' && method === 'PUT') {
      updatePayload = request.postDataJSON();
      purchaseItems = updatePayload.items.map((item: any, index: number) => ({
        ...purchaseItems[index],
        ...item,
        product_name: purchaseItems[index]?.product_name || 'منتج جديد',
      }));
      purchase = {
        ...purchase,
        supplier_id: updatePayload.supplier_id,
        invoice_number: updatePayload.invoice_number,
        purchase_date: updatePayload.purchase_date,
        expected_delivery_date: updatePayload.expected_delivery_date,
        notes: updatePayload.notes || '',
        total_amount: purchaseItems.reduce((sum, item) => sum + Number(item.quantity) * Number(item.unit_cost), 0),
      };
      purchase.remaining_amount = Math.max(0, purchase.total_amount - purchase.paid_amount);
      return route.fulfill(json({ success: true, data: { id: purchase.id } }));
    }
    if (path === '/purchases/purchase-1/payment' && method === 'POST') {
      paymentPayload = request.postDataJSON();
      if (paymentPayload.amount > purchase.total_amount - purchase.paid_amount) {
        return route.fulfill(json({ success: false, error: { message: 'payment amount exceeds remaining balance' } }, 400));
      }
      purchase.paid_amount += paymentPayload.amount;
      purchase.remaining_amount = Math.max(0, purchase.total_amount - purchase.paid_amount);
      return route.fulfill(json({ success: true, data: { purchase: { ...purchase } } }));
    }
    if (path === '/purchases/purchase-1/receive' && method === 'POST') {
      receiveCalled = true;
      return route.fulfill(json({ success: true, data: { id: 'purchase-1', status: 'received' } }));
    }
    if (path === '/products' && method === 'GET') {
      const pageNumber = Number(url.searchParams.get('page') || 1);
      const product = {
        id: 'product-search-' + pageNumber,
        name: 'فلتر تجريبي ' + pageNumber,
        barcode: '88000000' + pageNumber,
        sku: 'SKU-' + pageNumber,
        cost_price: 4,
        selling_price: 7,
      };
      return route.fulfill(json({ success: true, data: { products: [product], total: 25 } }));
    }
    if (path.startsWith('/products/barcode/') && method === 'GET') {
      return route.fulfill(json({ success: false, error: { code: '404', message: 'barcode not found' } }, 404));
    }
    return route.fulfill(json({ success: true, data: [] }));
  });

  await page.goto('/#/app/purchases/edit/purchase-1');
  await expect(page.getByRole('heading', { name: 'تعديل الشراء' })).toBeVisible();

  const nextField = page.getByRole('button', { name: 'انتقل إلى الحقل التالي' });
  const productSearch = page.getByLabel('البحث عن منتج');
  await expect(nextField).toBeVisible();
  await nextField.click();
  await expect(productSearch).toBeFocused();

  const quantity = page.getByLabel('كمية فلتر زيت');
  const cost = page.getByLabel('تكلفة فلتر زيت');
  const sellingPrice = page.getByLabel('سعر بيع فلتر زيت');
  const paymentAmount = page.getByLabel('مبلغ الدفعة');
  await quantity.fill('3');
  await quantity.press('Enter');
  await expect(cost).toBeFocused();
  await cost.fill('12');
  await cost.press('Enter');
  await expect(sellingPrice).toBeFocused();
  await sellingPrice.fill('18');
  await sellingPrice.press('Enter');
  await expect(page.getByLabel('مسح باركود المنتج')).toBeFocused();
  await expect(paymentAmount).toBeDisabled();

  await page.getByRole('button', { name: 'حفظ التعديلات' }).click();
  await expect(paymentAmount).toBeEnabled();
  expect(updatePayload.items[0]).toMatchObject({ quantity: 3, unit_cost: 12 });
  await paymentAmount.fill('36');
  await page.getByRole('button', { name: 'تسجيل الدفعة' }).click();
  await expect.poll(() => paymentPayload?.amount).toBe(36);
  await expect(page.getByText('₪36.00', { exact: false }).first()).toBeVisible();
  expect(purchase.remaining_amount).toBe(0);

  await productSearch.fill('فلتر');
  await expect(page.getByRole('button', { name: /فلتر تجريبي 1/ })).toBeVisible();
  await page.getByRole('button', { name: 'التالي', exact: true }).click();
  await expect.poll(() => observedCalls.some((call) => call.includes('/products?') && call.includes('page=2'))).toBe(true);
  await expect(page.getByRole('button', { name: /فلتر تجريبي 2/ })).toBeVisible();
  await page.getByRole('button', { name: /فلتر تجريبي 2/ }).click();
  await expect(page.getByText(/2 صنف/)).toBeVisible();

  const barcode = page.getByLabel('مسح باركود المنتج');
  await barcode.fill('880000009999');
  await barcode.press('Enter');
  const manualProductDialog = page.getByRole('dialog');
  await expect(manualProductDialog.getByText('إضافة قطعة جديدة')).toBeVisible();
  const manualBarcode = page.getByPlaceholder('أدخل الباركود');
  await expect(manualBarcode).toHaveValue('880000009999');
  await expect(manualProductDialog.getByRole('button', { name: 'انتقل إلى الحقل التالي' })).toBeVisible();
  await manualProductDialog.getByRole('button', { name: 'انتقل إلى الحقل التالي' }).click();
  await expect(manualBarcode).toBeFocused();
  const manualDescription = manualProductDialog.getByPlaceholder('أدخل وصف المنتج');
  await manualDescription.focus();
  await manualProductDialog.getByRole('button', { name: 'انتقل إلى الحقل التالي' }).click();
  await expect(manualProductDialog.getByRole('button', { name: 'إضافة للشراء والمخزون' })).toBeFocused();
  await manualProductDialog.getByRole('button', { name: 'إلغاء' }).click();
  await expect(manualProductDialog).toBeHidden();
  await page.getByRole('button', { name: 'حذف فلتر تجريبي 2 من الشراء' }).click();
  await expect(page.getByText(/1 صنف/)).toBeVisible();

  await page.getByRole('checkbox', { name: /استلام مباشر بعد الحفظ/ }).check();
  await page.getByRole('button', { name: 'تحديث واستلام' }).click();
  await expect.poll(() => receiveCalled).toBe(true);
  await expect(page).toHaveURL(/purchases/);

  expect(browserErrors).toEqual([]);
  expect(failedRequests).toEqual([]);
});
