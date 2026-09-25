import { expect, test } from '@playwright/test';

test('customer purchase history searches across server pages, retries failed data, and opens invoices', async ({ page }) => {
  const salesRequests: URL[] = [];
  const allSales = Array.from({ length: 25 }, (_, index) => ({
    id: `sale-${index + 1}`,
    invoice_number: `INV-${String(index + 1).padStart(3, '0')}`,
    customer_id: 'customer-1',
    sale_date: '2026-09-20T10:00:00Z',
    total_amount: 100 + index,
    paid_amount: 100 + index,
    remaining_amount: 0,
    payment_status: 'paid',
  }));
  let debtRequestShouldFail = true;

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
    if (path === '/customers/customer-1' && request.method() === 'GET') {
      return reply({ data: { id: 'customer-1', name: 'عميل الاختبار', totalPurchases: 2500, current_balance: 0 } });
    }
    if (path === '/customers/customer-1/ledger' && request.method() === 'GET') {
      return reply({ data: [] });
    }
    if (path === '/customers/customer-1/debts' && request.method() === 'GET') {
      if (debtRequestShouldFail) {
        return reply({ error: { message: 'Debt load failed' } }, 400);
      }
      return reply({ data: [] });
    }
    if (path === '/sales' && request.method() === 'GET') {
      salesRequests.push(url);
      const search = (url.searchParams.get('search') || '').toLowerCase();
      const customerId = url.searchParams.get('customer_id');
      const rows = allSales.filter((sale) => sale.customer_id === customerId
        && (!search || sale.invoice_number.toLowerCase().includes(search)));
      const currentPage = Number(url.searchParams.get('page') || 1);
      const perPage = Number(url.searchParams.get('per_page') || 20);
      return reply({ data: {
        sales: rows.slice((currentPage - 1) * perPage, currentPage * perPage),
        total: rows.length,
        page: currentPage,
        per_page: perPage,
      } });
    }
    if (path === '/sales/sale-21' && request.method() === 'GET') {
      return reply({ data: {
        sale: {
          id: 'sale-21', invoice_number: 'INV-021', sale_date: '2026-09-20T10:00:00Z',
          subtotal: 120, total_amount: 120, paid_amount: 120, payment_status: 'paid', payment_method: 'cash',
        },
        items: [{ product_name: 'قطعة اختبار', quantity: 1, unit_price: 120, total_amount: 120 }],
      } });
    }
    return reply({ data: [] });
  });

  await page.goto('/#/app/customers/customer-1/purchases');
  await expect(page.getByRole('button', { name: 'إعادة المحاولة' })).toBeVisible();
  debtRequestShouldFail = false;
  await page.getByRole('button', { name: 'إعادة المحاولة' }).click();
  await expect(page.getByRole('heading', { name: 'عميل الاختبار' })).toBeVisible();
  await expect(page.getByText('25 فاتورة')).toBeVisible();
  await expect(page.getByText('صفحة 1 من 2')).toBeVisible();

  await page.getByRole('button', { name: /التالي/ }).click();
  await expect(page.getByText('صفحة 2 من 2')).toBeVisible();
  await page.getByPlaceholder('ابحث برقم الفاتورة...').fill('INV-021');
  await expect.poll(() => salesRequests.some((url) => url.searchParams.get('search') === 'INV-021'
    && url.searchParams.get('page') === '1')).toBe(true);
  await expect(page.getByText('صفحة 1 من 1')).toBeVisible();
  await expect(page.getByText('INV-021')).toBeVisible();

  await page.getByRole('button', { name: 'عرض الفاتورة INV-021' }).click();
  await expect(page.getByTitle('قالب الفاتورة')).toBeVisible();
});
