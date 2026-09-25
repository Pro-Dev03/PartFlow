import { expect, test } from '@playwright/test';

test('inventory item barcode can be edited and verified in details', async ({ page }) => {
  const updatedBarcode = 'INV-BARCODE-UPDATED-01';
  const updatedCalls: Array<{ method: string; body: any }> = [];
  const requestedProducts: Array<{ search: string | null }> = [];
  const barcodeLookups: string[] = [];
  const product = {
    id: 'product-1',
    name: 'قطعة اختبار المخزون',
    sku: 'SKU-INV-1',
    barcode: 'INV-BARCODE-OLD-01',
    cost_price: 12,
    selling_price: 18,
    stock: 1,
  };
  const inventoryItem: Record<string, any> = {
    id: 'inventory-item-1',
    product_id: product.id,
    product_name: product.name,
    product,
    barcode: product.barcode,
    condition: 'NEW',
    status: 'AVAILABLE',
    quantity: 1,
    purchase_cost: 12,
    selling_price: 18,
    purchase_date: '2026-09-20T12:00:00Z',
  };

  const json = (data: unknown) => ({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ success: true, data }),
  });

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
    if (path === '/products' && request.method() === 'GET') {
      requestedProducts.push({ search: url.searchParams.get('search') });
    }
    if (path.startsWith('/barcodes/resolve/')) {
      barcodeLookups.push(decodeURIComponent(path.slice('/barcodes/resolve/'.length)));
      return route.fulfill(json({ product, inventory_item: null }));
    }
    if (url.hostname === 'partflow-api.onrender.com') {
      return route.fulfill(json({}));
    }
    if (path === '/categories') return route.fulfill(json([]));
    if (path === '/suppliers') return route.fulfill(json({ suppliers: [], total: 0 }));
    if (path === '/products' && request.method() === 'GET') {
      return route.fulfill({
        ...json({ products: [product] }),
        body: JSON.stringify({ success: true, data: { products: [product] }, meta: { total: 1 } }),
      });
    }
    if (path === '/inventory/items-with-supplier' && request.method() === 'GET') {
      return route.fulfill({
        ...json({ items: [inventoryItem] }),
        body: JSON.stringify({ success: true, data: { items: [inventoryItem] }, meta: { total: 1 } }),
      });
    }
    if (path === '/inventory/items/inventory-item-1' && request.method() === 'PUT') {
      const body = request.postDataJSON();
      updatedCalls.push({ method: request.method(), body });
      Object.assign(inventoryItem, body);
      return route.fulfill(json({ item: inventoryItem }));
    }
    return route.fulfill(json([]));
  });

  await page.goto('/#/app/inventory');
  await expect(page.getByRole('heading', { name: 'المنتجات' })).toBeVisible();
  const search = page.getByPlaceholder('ابحث عن منتج أو امسح الباركود');
  await search.fill('قطعة اختبار المخزون');
  await search.press('Enter');
  await expect(search).toHaveValue('قطعة اختبار المخزون');
  await expect.poll(() => requestedProducts.some((request) => request.search === 'قطعة اختبار المخزون')).toBe(true);
  expect(barcodeLookups).toEqual([]);
  await search.fill('');
  await search.pressSequentially('SCN-INV-0001', { delay: 8 });
  await search.press('Enter');
  await expect(search).toHaveValue('INV-BARCODE-OLD-01');
  await expect.poll(() => barcodeLookups).toEqual(['SCN-INV-0001']);
  await search.fill('');

  await page.getByRole('button', { name: 'عناصر المخزون' }).click();
  await page.getByRole('button', { name: 'عرض الجدول' }).click();
  await expect(page.getByText(product.name, { exact: true }).first()).toBeVisible();
  await page.getByRole('button', { name: 'خيارات العنصر' }).click();
  await page.getByRole('menuitem', { name: 'تعديل الباركود' }).click();

  const barcodeInput = page.getByLabel('باركود عنصر المخزون');
  await expect(barcodeInput).toHaveValue(product.barcode);
  await barcodeInput.fill(updatedBarcode);
  const nextButton = page.getByRole('button', { name: 'انتقل إلى الحقل التالي' });
  await expect(nextButton).toBeVisible();
  await nextButton.click();
  await expect(page.getByRole('dialog').getByRole('button', { name: 'حفظ' })).toBeFocused();
  await page.getByRole('dialog').getByRole('button', { name: 'حفظ' }).click();
  await expect(page.getByRole('dialog')).toBeHidden();

  expect(updatedCalls).toEqual([{ method: 'PUT', body: { barcode: updatedBarcode } }]);
  await expect(page.getByText(updatedBarcode, { exact: true }).first()).toBeVisible();
  await page.getByRole('button', { name: 'عرض العنصر' }).click();
  await expect(page.getByRole('dialog')).toContainText(updatedBarcode);
});
