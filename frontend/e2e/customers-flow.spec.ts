import { expect, test } from '@playwright/test';

test('customer search, sorting, export, save retry, archive retry, and next navigation work', async ({ page }) => {
  const customerRequests: URL[] = [];
  const allCustomers = Array.from({ length: 205 }, (_, index) => ({
    id: `customer-${index + 1}`,
    code: `C-${String(index + 1).padStart(3, '0')}`,
    name: index === 0 ? 'No Phone Customer' : `Customer ${String(index + 1).padStart(3, '0')}`,
    phone: index === 0 ? null : `059900${String(index).padStart(4, '0')}`,
    totalPurchases: index * 10,
    paidAmount: index * 8,
    outstanding: index * 2,
    is_active: true,
  }));
  let createShouldFail = true;
  let deleteShouldFail = true;
  let createdPayload: any;

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
    const reply = (data: Record<string, unknown>, status = 200) => route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({ success: status < 400, ...data }),
    });

    if (url.hostname === 'partflow-api.onrender.com') return reply({ data: {} });
    if (path === '/dashboard/stats' && request.method() === 'GET') {
      return reply({ data: { total_customers: 205, activeCustomers: 80, outstandingDebtorCount: 14, outstandingDebts: 725 } });
    }
    if (path === '/customers' && request.method() === 'GET') {
      customerRequests.push(url);
      const search = (url.searchParams.get('search') || '').toLowerCase();
      let rows = allCustomers.filter((customer) => !search
        || customer.name.toLowerCase().includes(search)
        || customer.code.toLowerCase().includes(search)
        || String(customer.phone || '').includes(search));
      const sortBy = url.searchParams.get('sort_by');
      const sortOrder = url.searchParams.get('sort_order');
      if (sortBy === 'name') rows = [...rows].sort((left, right) => left.name.localeCompare(right.name));
      if (sortOrder === 'desc') rows.reverse();
      const pageNumber = Number(url.searchParams.get('page') || 1);
      const perPage = Number(url.searchParams.get('per_page') || 20);
      return reply({
        data: rows.slice((pageNumber - 1) * perPage, pageNumber * perPage),
        meta: { total: rows.length, page: pageNumber, per_page: perPage, total_pages: Math.ceil(rows.length / perPage) },
      });
    }
    if (path === '/customers' && request.method() === 'POST') {
      createdPayload = request.postDataJSON();
      if (createShouldFail) return reply({ error: { message: 'Customer save failed' } }, 400);
      return reply({ data: { id: 'created-customer', ...createdPayload } }, 201);
    }
    if (path.startsWith('/customers/customer-') && request.method() === 'DELETE') {
      if (deleteShouldFail) return reply({ error: { message: 'Customer archive failed' } }, 400);
      return reply({ data: { id: 'customer-1' } });
    }
    return reply({ data: [] });
  });

  await page.goto('/#/app/customers');
  await expect(page.getByRole('heading', { name: 'العملاء', exact: true })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'No Phone Customer' })).toBeVisible();

  const search = page.getByPlaceholder('ابحث بالاسم أو رقم الهاتف...', { exact: true });
  await search.fill('No Phone');
  await expect.poll(() => customerRequests.some((url) => url.searchParams.get('search') === 'No Phone')).toBe(true);
  await expect(page.getByRole('heading', { name: 'No Phone Customer' })).toBeVisible();
  await search.fill('');
  await expect.poll(() => customerRequests.some((url) => !url.searchParams.has('search'))).toBe(true);

  await page.getByRole('button', { name: 'ترتيب حسب الاسم' }).click();
  await expect.poll(() => customerRequests.some((url) => url.searchParams.get('sort_by') === 'name' && url.searchParams.get('sort_order') === 'asc')).toBe(true);

  await page.getByRole('button', { name: 'تقرير البيانات' }).click();
  const downloadPromise = page.waitForEvent('download');
  await page.getByRole('menuitem', { name: 'تصدير كل النتائج' }).click();
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toContain('customers-all-');
  await expect.poll(() => customerRequests.filter((url) => url.searchParams.get('per_page') === '100').length).toBeGreaterThanOrEqual(3);

  await page.getByRole('button', { name: 'إضافة زبون' }).click();
  const formDialog = page.getByRole('dialog');
  const code = formDialog.getByLabel('كود العميل');
  const name = formDialog.getByLabel('الاسم الكامل');
  const phone = formDialog.getByLabel('رقم الهاتف');
  await expect(code).toBeFocused();
  await page.getByRole('button', { name: 'انتقل إلى الحقل التالي' }).click();
  await expect(name).toBeFocused();
  await name.fill('Retry Customer');
  await page.getByRole('button', { name: 'انتقل إلى الحقل التالي' }).click();
  await expect(phone).toBeFocused();
  await phone.fill('0599111222');
  await formDialog.getByRole('button', { name: 'إضافة العميل' }).click();
  await expect(formDialog).toBeVisible();
  expect(createdPayload.name).toBe('Retry Customer');
  createShouldFail = false;
  await formDialog.getByRole('button', { name: 'إضافة العميل' }).click();
  await expect(formDialog).toBeHidden();

  await page.getByRole('button', { name: 'خيارات العميل' }).first().click();
  await page.getByRole('menuitem', { name: 'حذف' }).click();
  const confirmDialog = page.getByRole('dialog');
  await confirmDialog.getByRole('button', { name: 'أرشفة العميل' }).click();
  await expect(confirmDialog).toBeVisible();
  deleteShouldFail = false;
  await confirmDialog.getByRole('button', { name: 'أرشفة العميل' }).click();
  await expect(confirmDialog).toBeHidden();
});
