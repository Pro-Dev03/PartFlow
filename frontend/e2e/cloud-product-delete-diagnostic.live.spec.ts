import { expect, test } from '@playwright/test';

const apiBase = 'https://partflow-api.onrender.com/api/v1';

test('diagnose the QA product deletion response', async ({ request }) => {
  const productId = process.env.E2E_QA_PRODUCT_ID;
  const email = process.env.E2E_EMAIL;
  const password = process.env.E2E_PASSWORD;
  expect(productId).toBeTruthy();
  expect(email).toBeTruthy();
  expect(password).toBeTruthy();

  const login = await request.post(`${apiBase}/auth/login`, { data: { email, password } });
  expect(login.status()).toBe(200);
  const loginPayload = await login.json();
  const token = loginPayload?.data?.token ?? loginPayload?.data?.access_token;
  expect(token).toBeTruthy();
  const headers = { Authorization: `Bearer ${token}` };

  const product = await request.get(`${apiBase}/products/${productId}`, { headers });
  console.log(JSON.stringify({ qaProductId: productId, getStatus: product.status(), productName: (await product.json().catch(() => undefined))?.data?.name }));
  const deleted = await request.delete(`${apiBase}/products/${productId}`, { headers });
  console.log(JSON.stringify({ qaProductId: productId, deleteStatus: deleted.status(), deleteBody: await deleted.json().catch(() => undefined) }));
});
