/**
 * POS Cash Sale E2E Test - Full Monitoring
 * Monitors: Console errors, React errors, Network requests (400/401/404/500), Backend responses
 */
import { chromium } from '@playwright/test';

const BASE_URL = 'http://localhost:5173';

// Monitoring collectors
const consoleErrors = [];
const consoleWarnings = [];
const networkLog = [];
const pageErrors = [];

async function runTest() {
  console.log('🚀 Starting POS Cash Sale test with full monitoring...\n');

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    locale: 'ar',
  });
  const page = await context.newPage();

  // ===== MONITORING SETUP =====

  // 1. Console messages (JS errors, React warnings)
  page.on('console', (msg) => {
    const type = msg.type();
    const text = msg.text();
    if (type === 'error') {
      consoleErrors.push(text);
      console.log(`  ❌ [CONSOLE ERROR] ${text.substring(0, 400)}`);
    } else if (type === 'warning') {
      consoleWarnings.push(text);
      console.log(`  ⚠️  [CONSOLE WARN] ${text.substring(0, 250)}`);
    }
  });

  // 2. Page errors (uncaught exceptions / React crashes)
  page.on('pageerror', (err) => {
    pageErrors.push(err.message);
    console.log(`  💥 [PAGE ERROR] ${err.message.substring(0, 400)}`);
  });

  // 3. Network requests - capture ALL API calls with status + response body
  page.on('response', async (response) => {
    const url = response.url();
    if (url.includes('/api/') || url.includes(':8080')) {
      const status = response.status();
      let body = '';
      try {
        body = await response.text();
      } catch { /* ignore */ }
      const entry = { method: response.request().method(), url, status, body: body.substring(0, 500) };
      networkLog.push(entry);
      const icon = status >= 400 ? '🔴' : '🟢';
      console.log(`  ${icon} [NETWORK] ${response.request().method()} ${url.replace(BASE_URL, '')} → ${status}`);
      if (status >= 400 && body) {
        console.log(`     └─ Response: ${body.substring(0, 400)}`);
      }
    }
  });

  // 4. Failed requests (network-level failures)
  page.on('requestfailed', (request) => {
    const failure = request.failure()?.errorText || 'unknown';
    console.log(`  🔴 [REQUEST FAILED] ${request.method()} ${request.url()} → ${failure}`);
    networkLog.push({ method: request.method(), url: request.url(), status: 'FAILED', body: failure });
  });

  // ===== TEST SCENARIO =====
  try {
    // Step 1: Open POS page
    console.log('\n📄 Step 1: Opening POS page (/app/sales)...');
    await page.goto(`${BASE_URL}/app/sales`, { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(2000);
    await page.screenshot({ path: '/tmp/pos-1-initial.png', fullPage: false });
    console.log('   ✅ Page loaded. Screenshot: /tmp/pos-1-initial.png');

    // Step 2: Add a product to cart by clicking first product card
    console.log('\n🛒 Step 2: Adding product to cart...');
    // Try clicking on a product card in the ProductSearch grid
    const productCards = page.locator('[class*="cursor-pointer"]').filter({ hasText: /₪|شيكل/ });
    const cardCount = await productCards.count();
    console.log(`   Found ${cardCount} clickable product cards`);

    let added = false;
    if (cardCount > 0) {
      await productCards.first().click();
      added = true;
    } else {
      // Fallback: look for buttons with "إضافة" text or product names
      const addButtons = page.getByRole('button', { name: /إضافة|add/i });
      const btnCount = await addButtons.count();
      console.log(`   Found ${btnCount} add buttons`);
      if (btnCount > 0) {
        await addButtons.first().click();
        added = true;
      }
    }

    if (!added) {
      // Last resort: use barcode input
      console.log('   Using barcode fallback...');
      const barcodeInput = page.locator('input[placeholder*="باركود"], input[placeholder*="barcode" i]').first();
      if (await barcodeInput.count() > 0) {
        await barcodeInput.fill('1234567890125'); // RAM barcode
        await barcodeInput.press('Enter');
        added = true;
      }
    }

    await page.waitForTimeout(1500);
    await page.screenshot({ path: '/tmp/pos-2-after-add.png', fullPage: false });
    console.log('   ✅ Product added. Screenshot: /tmp/pos-2-after-add.png');

    // Check cart state - look for total amount display
    const totalText = await page.locator('text=/الإجمالي/').locator('..').textContent().catch(() => null);
    console.log(`   Total section: ${totalText ? totalText.substring(0, 100) : 'NOT FOUND'}`);

    // Step 3: Fill paid amount for cash payment
    console.log('\n💰 Step 3: Filling paid amount (cash payment)...');
    const paidInput = page.locator('input[type="number"][placeholder*="المبلغ"]').first();
    if (await paidInput.count() > 0) {
      // Extract total from the total display, or just enter a large enough amount
      const totalMatch = totalText ? totalText.match(/₪?([\d.]+)/) : null;
      const totalValue = totalMatch ? parseFloat(totalMatch[1]) : 250;
      await paidInput.fill(String(totalValue));
      console.log(`   Entered paid amount: ${totalValue}`);
    } else {
      console.log('   ⚠️ Paid amount input NOT FOUND!');
    }
    await page.waitForTimeout(800);
    await page.screenshot({ path: '/tmp/pos-3-payment-filled.png', fullPage: false });

    // Step 4: Click checkout button
    console.log('\n✅ Step 4: Clicking "إتمام البيع" (Complete Sale)...');
    const checkoutBtn = page.getByRole('button', { name: /إتمام البيع/ }).first();
    const isDisabled = await checkoutBtn.isDisabled().catch(() => 'not-found');
    console.log(`   Checkout button disabled: ${isDisabled}`);

    if (isDisabled === true) {
      console.log('   ❌ BUTTON IS DISABLED — investigating why...');
      // Dump button state and surrounding info
      const btnHTML = await checkoutBtn.evaluate((el) => el.outerHTML).catch(() => 'n/a');
      console.log(`   Button HTML: ${btnHTML.substring(0, 300)}`);
    } else if (isDisabled === 'not-found') {
      console.log('   ❌ CHECKOUT BUTTON NOT FOUND!');
      const allButtons = await page.locator('button').allTextContents();
      console.log(`   All buttons: ${JSON.stringify(allButtons).substring(0, 500)}`);
    } else {
      await checkoutBtn.click();
      console.log('   ✅ Clicked checkout button');
    }

    // Step 5: Wait for result - success modal OR error
    console.log('\n⏳ Step 5: Waiting for result (success invoice modal or error)...');
    await page.waitForTimeout(5000);
    await page.screenshot({ path: '/tmp/pos-4-result.png', fullPage: false });

    // Check for invoice modal (success indicator)
    const invoiceModal = await page.locator('text=/فاتورة البيع/').count();
    const errorMsg = await page.locator('text=/خطأ|فشل|حدث خطأ|تعذر/').count();
    const toastError = await page.locator('[class*="toast"], [role="alert"], [data-sonner-toast]').allTextContents().catch(() => []);

    console.log(`\n📊 RESULTS:`);
    console.log(`   Invoice modal visible: ${invoiceModal > 0 ? '✅ YES' : '❌ NO'}`);
    console.log(`   Error text on page: ${errorMsg > 0 ? '❌ YES' : '✅ NO'}`);
    console.log(`   Toasts/alerts: ${toastError.length > 0 ? JSON.stringify(toastError).substring(0, 300) : 'none'}`);

    // ===== SUMMARY =====
    console.log('\n' + '='.repeat(60));
    console.log('📋 MONITORING SUMMARY');
    console.log('='.repeat(60));
    console.log(`Console Errors: ${consoleErrors.length}`);
    console.log(`Page Errors (uncaught): ${pageErrors.length}`);
    console.log(`API Requests captured: ${networkLog.length}`);
    const failedApi = networkLog.filter(r => r.status >= 400 || r.status === 'FAILED');
    console.log(`Failed API requests (4xx/5xx): ${failedApi.length}`);
    if (failedApi.length > 0) {
      console.log('\nFailed requests detail:');
      failedApi.forEach(r => console.log(`  ${r.method} ${r.url} → ${r.status}\n  Body: ${r.body}`));
    }
    if (consoleErrors.length > 0) {
      console.log('\nAll console errors:');
      consoleErrors.forEach(e => console.log(`  - ${e.substring(0, 300)}`));
    }
    if (pageErrors.length > 0) {
      console.log('\nAll page errors:');
      pageErrors.forEach(e => console.log(`  - ${e.substring(0, 300)}`));
    }

    // Verdict
    const saleSucceeded = invoiceModal > 0;
    console.log('\n' + '='.repeat(60));
    console.log(saleSucceeded
      ? '🎉 VERDICT: Cash sale COMPLETED successfully in browser!'
      : '🚨 VERDICT: Cash sale FAILED in browser — see details above');
    console.log('='.repeat(60));

  } catch (err) {
    console.error('\n💥 Test scenario error:', err.message);
    await page.screenshot({ path: '/tmp/pos-error-state.png', fullPage: true }).catch(() => {});
  } finally {
    await browser.close();
  }
}

runTest().catch((e) => { console.error('FATAL:', e); process.exit(1); });