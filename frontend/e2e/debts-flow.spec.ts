import { expect, test } from '@playwright/test';

test('debt search and status filters reach the server and payment records update the balance', async ({ page }) => {
  const debtRequests: URL[] = [];
  let paymentPayload: any;
  let remainingAmount = 120;
  let paymentAttempts = 0;
  const recordedPaymentReferences = new Map<string, { amount: number; method: string }>();

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
    if (path === '/debts' && request.method() === 'GET') {
      debtRequests.push(url);
      const search = url.searchParams.get('search') || '';
      const status = url.searchParams.get('status') || '';
      const tab = url.searchParams.get('tab') || 'all';
      const customerId = url.searchParams.get('customer_id') || '';
      const paid = tab === 'paid' || status === 'paid';
      const row = customerId === 'customer-outside'
        ? { id: 'debt-outside', customer_id: 'customer-outside', customer_name: 'Outside Customer', customer_code: 'C-OUT', customer_phone: '0594444444', invoice_number: 'INV-OUT', amount: 80, remaining_amount: 80, due_date: '2026-10-01', status: status || 'pending' }
        : search.toLowerCase().includes('far')
        ? { id: 'debt-far', customer_id: 'customer-far', customer_name: 'Far Customer', customer_code: 'C-FAR', customer_phone: '0591111111', invoice_number: 'INV-FAR', amount: 200, remaining_amount: 200, due_date: '2026-10-01', status: status || 'pending' }
        : paid
          ? { id: 'debt-paid', customer_id: 'customer-paid', customer_name: 'Paid Customer', customer_code: 'C-PAID', customer_phone: '0592222222', invoice_number: 'INV-PAID', amount: 40, remaining_amount: 0, due_date: '2026-09-01', status: 'paid' }
          : { id: 'debt-open', customer_id: 'customer-open', customer_name: 'Open Customer', customer_code: 'C-OPEN', customer_phone: '0593333333', invoice_number: 'INV-OPEN', amount: 120, remaining_amount: remainingAmount, due_date: '2026-10-01', status: 'pending' };
      const matchesStatus = !status || status === row.status || (status === 'overdue' && row.status === 'overdue');
      const data = matchesStatus && (tab !== 'paid' || row.status === 'paid') && (tab !== 'open' || row.status !== 'paid') ? [row] : [];
      return reply({ data, meta: { page: 1, per_page: 10, total: data.length, summary: { total_debt: 120, paid_amount: 0, remaining_amount: remainingAmount, customer_count: 1 } } });
    }
    if (path === '/customers/customer-open/debt-payments' && request.method() === 'POST') {
      paymentPayload = request.postDataJSON();
      paymentAttempts += 1;
      const priorPayment = recordedPaymentReferences.get(paymentPayload.reference);
      if (priorPayment && (priorPayment.amount !== Number(paymentPayload.amount) || priorPayment.method !== paymentPayload.method)) {
        return reply({ error: { message: 'Duplicate payment reference' } }, 409);
      }
      if (!priorPayment) {
        recordedPaymentReferences.set(paymentPayload.reference, { amount: Number(paymentPayload.amount), method: paymentPayload.method });
        remainingAmount -= Number(paymentPayload.amount);
      }
      if (paymentAttempts === 1) return reply({ error: { message: 'Payment response lost after commit' } }, 500);
      return reply({ data: { success: true } }, 201);
    }
    return reply({ data: [] });
  });

  await page.goto('/#/app/debts');
  await expect(page.getByRole('heading', { name: 'قائمة الديون' })).toBeVisible();
  await expect.poll(() => debtRequests.some((url) => url.searchParams.get('tab') === 'open')).toBe(true);

  await page.getByRole('button', { name: 'تسجيل دفعة' }).first().click();
  await page.getByPlaceholder('أدخل المبلغ...').fill('50');
  await page.locator('select').selectOption('bank_transfer');
  await page.getByRole('button', { name: 'تسجيل الدفعة' }).click();
  await expect(page.getByText('تعذر تسجيل الدفعة. بقي النموذج مفتوحًا؛ أعد المحاولة دون إعادة إدخال البيانات.')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'تسجيل دفعة' })).toBeVisible();
  const firstPaymentReference = paymentPayload.reference;
  await page.getByRole('button', { name: 'تسجيل الدفعة' }).click();
  await expect(page.getByRole('heading', { name: 'تم تسجيل الدفعة بنجاح!' })).toBeVisible();
  expect(paymentPayload).toMatchObject({ amount: 50, method: 'bank_transfer' });
  expect(paymentPayload.reference).toBe(firstPaymentReference);
  await page.getByRole('button', { name: 'إغلاق', exact: true }).click();
  await expect.poll(() => debtRequests.filter((url) => url.searchParams.get('tab') === 'open').length).toBeGreaterThan(1);
  await expect(page.getByRole('button', { name: /غير مسدد ₪70/ })).toBeVisible();

  await page.getByPlaceholder('ابحث بالاسم أو الكود أو الهاتف...').fill('Far Customer');
  await expect.poll(() => debtRequests.some((url) => url.searchParams.get('search') === 'Far Customer'
    && url.searchParams.get('tab') === 'open')).toBe(true);
  await expect(page.getByRole('heading', { name: 'Far Customer' }).first()).toBeVisible();

  await page.getByRole('button', { name: 'خيارات متقدمة' }).click();
  await page.getByRole('button', { name: 'متأخر', exact: true }).click();
  await expect.poll(() => debtRequests.some((url) => url.searchParams.get('status') === 'overdue'
    && url.searchParams.get('search') === 'Far Customer')).toBe(true);

  await page.goto('/#/app/debts?customer_id=customer-outside');
  await expect.poll(() => debtRequests.some((url) => url.searchParams.get('customer_id') === 'customer-outside')).toBe(true);
  await expect(page.getByRole('heading', { name: 'تسجيل دفعة' })).toBeVisible();
  await expect(page.getByPlaceholder('أدخل المبلغ...')).toHaveValue('');
});
