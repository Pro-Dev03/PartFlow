import { expect, test } from '@playwright/test';

test('expenses search, export, decimal entry, and next-field navigation work', async ({ page }) => {
  const listRequests: URL[] = [];
  const apiRequests: string[] = [];
  let createdExpense: any;
  const expenseRows = Array.from({ length: 205 }, (_, index) => ({
    id: `expense-${index + 1}`,
    title: `Expense ${index + 1}`,
    description: `Expense ${index + 1}`,
    amount: 2.5,
    expense_date: '2026-09-10T12:00:00Z',
    category_id: 'category-1',
    category_name: 'General',
    status: 'approved',
    is_recurring: false,
  }));

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
    apiRequests.push(`${request.method()} ${path}`);
    const json = (data: unknown, status = 200) => route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({ success: true, ...(data as Record<string, unknown>) }),
    });

    if (url.hostname === 'partflow-api.onrender.com') return json({ data: {} });
    if (path === '/expenses/categories' && request.method() === 'GET') {
      return json({ data: [{ id: 'category-1', name: 'General', is_active: true }], meta: { total: 1 } });
    }
    if (path === '/expenses' && request.method() === 'GET') {
      listRequests.push(url);
      const pageNumber = Number(url.searchParams.get('page') || 1);
      const perPage = Number(url.searchParams.get('per_page') || 20);
      const filtered = url.searchParams.get('search')
        ? expenseRows.filter((expense) => expense.description.toLowerCase().includes((url.searchParams.get('search') || '').toLowerCase()))
        : expenseRows;
      const rows = filtered.slice((pageNumber - 1) * perPage, pageNumber * perPage);
      return json({ data: rows, meta: { page: pageNumber, per_page: perPage, total: filtered.length, total_pages: Math.ceil(filtered.length / perPage) } });
    }
    if (path === '/expenses' && request.method() === 'POST') {
      createdExpense = request.postDataJSON();
      return json({ data: { id: 'created-expense', ...createdExpense } });
    }
    return json({ data: [] });
  });

  await page.goto('/#/app/expenses');
  await expect(page.getByRole('heading', { name: 'المصروفات', exact: true })).toBeVisible();

  const search = page.getByPlaceholder('بحث', { exact: true });
  await search.fill('Expense 2');
  await expect.poll(() => listRequests.some((url) => url.searchParams.get('search') === 'Expense 2')).toBe(true);
  expect(listRequests.some((url) => url.searchParams.get('search') === 'Expense 2')).toBe(true);
  await search.fill('');
  await expect.poll(() => listRequests.some((url) => !url.searchParams.has('search') && !url.searchParams.has('category_id'))).toBe(true);

  const reportMenu = page.getByRole('button', { name: 'تقرير البيانات' });
  await reportMenu.click();
  const downloadPromise = page.waitForEvent('download');
  await page.getByRole('menuitem', { name: 'تصدير كل النتائج' }).click();
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toContain('expenses-all-');
  await expect.poll(() => listRequests.filter((url) => url.searchParams.get('per_page') === '100').length).toBeGreaterThanOrEqual(3);

  await page.getByRole('button', { name: 'إضافة مصروف' }).click();
  const dialog = page.getByRole('dialog');
  const description = dialog.getByLabel('وصف المصروف');
  const amount = dialog.getByLabel('المبلغ');
  await expect(dialog.locator('select')).toHaveValue('category-1');
  await dialog.locator('select').selectOption('category-1');
  await description.fill('Electricity bill');
  await page.getByRole('button', { name: 'انتقل إلى الحقل التالي' }).click();
  await expect(amount).toBeFocused();
  await amount.fill('12.75');
  const saveButton = dialog.getByRole('button', { name: 'حفظ المصروف' });
  await saveButton.click();
  await expect(apiRequests).toContain('POST /expenses');
  await expect.poll(() => createdExpense?.amount).toBe(12.75);
  expect(createdExpense.category_id).toBe('category-1');
  await expect(dialog).toBeHidden();
});
