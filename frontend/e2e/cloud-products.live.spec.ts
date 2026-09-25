import { expect, test } from '@playwright/test';

const email = process.env.E2E_EMAIL;
const password = process.env.E2E_PASSWORD;
const apiBase = 'https://partflow-api.onrender.com/api/v1';

test.use({ serviceWorkers: 'block' });
test.skip(
  !email || !password || process.env.E2E_BASE_URL !== 'https://partflow-hpv7.onrender.com',
  'Live QA credentials and E2E_BASE_URL=https://partflow-hpv7.onrender.com are required',
);

test('cloud product create, search, edit, cancel, and delete persist correctly', async ({ page, request }) => {
  const runId = Date.now().toString();
  const name = `QA-PF-PRODUCT-${runId}`;
  const editedName = `${name}-EDIT`;
  const sku = `QA-${runId}`;
  const barcode = `98${runId.slice(-11)}`;
  let productId: string | undefined;
  let authorization: string | undefined;
  let deletedThroughUI = false;

  page.on('request', (req) => {
    if (!req.url().startsWith(apiBase)) return;
    const header = req.headers().authorization;
    if (header?.startsWith('Bearer ')) authorization = header;
  });

  const apiGet = async (id: string) => request.get(`${apiBase}/products/${id}`, {
    headers: authorization ? { Authorization: authorization } : {},
  });

  try {
    await page.addInitScript(() => localStorage.setItem('partflow-connection-mode', 'cloud'));
    await page.goto('/#/login', { waitUntil: 'domcontentloaded' });
    await page.locator('#email').fill(email!);
    await page.locator('#password').fill(password!);
    await page.locator('form button[type="submit"]').click();
    await expect(page).toHaveURL(/#\/app(?:\/|$)/, { timeout: 30_000 });
    await page.goto('/#/app/inventory', { waitUntil: 'domcontentloaded' });

    await page.getByRole('button', { name: 'إضافة', exact: true }).first().click();
    await page.locator('.pf-entry-option').filter({ hasText: 'إنشاء صنف جديد' }).click();
    const categoryDialog = page.getByRole('dialog', { name: 'اختر تصنيف المنتج' });
    await expect(categoryDialog).toBeVisible();
    await expect.poll(() => categoryDialog.locator('select option').count()).toBeGreaterThan(1);
    await categoryDialog.locator('select').selectOption({ index: 1 });
    await categoryDialog.getByRole('button', { name: 'متابعة للمنتج' }).click();

    const createDialog = page.getByRole('dialog', { name: 'إضافة منتج جديد' });
    await expect(createDialog).toBeVisible();
    await createDialog.getByRole('button', { name: 'التالي', exact: true }).click();
    await createDialog.locator('.product-create-stage input[type="number"]').first().fill('11.5');
    await createDialog.getByPlaceholder('مثال: CPU-001').fill(sku);
    await createDialog.getByPlaceholder('امسح أو أدخل الباركود').fill(barcode);
    await page.waitForTimeout(400);
    await createDialog.getByPlaceholder('أدخل اسم المنتج').fill(name);
    await createDialog.getByRole('button', { name: 'التالي', exact: true }).click();
    const pricingInputs = createDialog.locator('.product-create-stage input[type="number"]');
    await pricingInputs.nth(0).fill('17');
    await pricingInputs.nth(1).fill('0');
    await pricingInputs.nth(2).fill('2');
    await createDialog.getByRole('button', { name: 'التالي', exact: true }).click();

    const createdResponse = page.waitForResponse((response) =>
      response.url() === `${apiBase}/products` && response.request().method() === 'POST'
    );
    await createDialog.getByRole('button', { name: 'إضافة المنتج' }).click();
    const created = await createdResponse;
    expect(created.status()).toBeGreaterThanOrEqual(200);
    expect(created.status()).toBeLessThan(300);
    const createdPayload = await created.json();
    productId = String(createdPayload?.data?.product?.id ?? createdPayload?.data?.id ?? '');
    expect(productId).toBeTruthy();

    const initialProduct = await apiGet(productId!);
    expect(initialProduct.status()).toBe(200);
    const initialData = (await initialProduct.json()).data;
    const initialRecord = initialData?.product ?? initialData;
    expect(initialRecord.name).toBe(name);
    expect(initialRecord.sku).toBe(sku);
    expect(Number(initialRecord.selling_price)).toBe(17);
    expect(Number(initialRecord.cost_price)).toBe(11.5);

    const search = page.locator('input.pf-inventory-search-input');
    await search.fill(name);
    const card = page.locator('.compact-product-card:visible').filter({ hasText: name });
    await expect(card).toHaveCount(1);

    await card.getByRole('button', { name: 'خيارات المنتج' }).click();
    await page.getByRole('menuitem', { name: 'تعديل', exact: true }).click();
    const editDialog = page.getByRole('dialog', { name: 'تعديل المنتج' });
    await editDialog.getByPlaceholder('أدخل اسم المنتج').fill(`${name}-CANCEL`);
    await editDialog.getByRole('button', { name: 'إلغاء' }).click();
    const afterCancel = (await (await apiGet(productId!)).json()).data;
    expect((afterCancel?.product ?? afterCancel).name).toBe(name);

    await card.getByRole('button', { name: 'خيارات المنتج' }).click();
    await page.getByRole('menuitem', { name: 'تعديل', exact: true }).click();
    await editDialog.getByPlaceholder('أدخل اسم المنتج').fill(editedName);
    const updatedResponse = page.waitForResponse((response) =>
      response.url() === `${apiBase}/products/${productId}` && response.request().method() === 'PUT'
    );
    await editDialog.getByRole('button', { name: 'حفظ التغييرات' }).click();
    expect((await updatedResponse).ok()).toBeTruthy();
    const updatedData = (await (await apiGet(productId!)).json()).data;
    expect((updatedData?.product ?? updatedData).name).toBe(editedName);

    await search.fill(editedName);
    const editedCard = page.locator('.compact-product-card:visible').filter({ hasText: editedName });
    await expect(editedCard).toHaveCount(1);
    await editedCard.getByRole('button', { name: 'خيارات المنتج' }).click();
    await page.getByRole('menuitem', { name: 'حذف', exact: true }).click();
    const deletedResponse = page.waitForResponse((response) =>
      response.url() === `${apiBase}/products/${productId}` && response.request().method() === 'DELETE'
    );
    await page.getByRole('button', { name: 'حذف المنتج', exact: true }).click();
    const deleted = await deletedResponse;
    if (!deleted.ok()) {
      console.log(JSON.stringify({ qaProductId: productId, deleteStatus: deleted.status(), deleteBody: await deleted.json().catch(() => undefined) }));
    }
    expect(deleted.ok()).toBeTruthy();
    deletedThroughUI = true;
    await expect(editedCard).toHaveCount(0);

    console.log(JSON.stringify({ product: 'PASS', created: true, searched: true, cancelPreserved: true, edited: true, deletedThroughUI }));
  } finally {
    if (productId && !deletedThroughUI && authorization) {
      const cleanup = await request.delete(`${apiBase}/products/${productId}`, {
        headers: { Authorization: authorization },
      });
      console.log(JSON.stringify({ qaCleanupStatus: cleanup.status(), qaProductId: productId }));
    }
  }
});
