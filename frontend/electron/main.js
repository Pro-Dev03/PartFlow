import { app, BrowserWindow, shell, Tray, Menu, nativeImage, ipcMain, dialog } from 'electron';
import { spawn, execSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const isDev = !app.isPackaged;

// PartFlow is a data-management desktop app and does not require GPU
// acceleration. Disabling it before the app is ready avoids a hard Electron
// shutdown on Windows machines where the GPU process cannot load its driver
// (for example, on a VM or a machine with a broken graphics stack).
app.commandLine.appendSwitch('disable-gpu');
app.commandLine.appendSwitch('disable-gpu-compositing');
app.commandLine.appendSwitch('in-process-gpu');
app.disableHardwareAcceleration();

app.name = 'PartFlow';

const appState = {
  tray: null,
  splashWindow: null,
  mainWindow: null,
  backendProcess: null,
  backendStarting: null,
};

const backendPort = 8080;

function killExistingBackendProcesses() {
  if (process.platform !== 'win32') return;

  try {
    const output = execSync(`netstat -ano -p tcp | findstr :${backendPort}`, {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    const pids = new Set();
    for (const line of String(output).split(/\r?\n/)) {
      const match = line.trim().match(/\s+(\d+)\s*$/);
      if (match && match[1]) {
        pids.add(match[1]);
      }
    }

    for (const pid of pids) {
      try {
        execSync(`taskkill /PID ${pid} /F /T`, { stdio: 'ignore' });
      } catch {
        // Ignore failed process cleanup; the main backend start will fail fast if a port is still in use.
      }
    }
  } catch {
    // netstat can fail if nothing is listening; this is safe to ignore.
  }
}

function getBackendPath() {
  if (isDev) {
    return path.join(__dirname, '..', '..', 'dist', 'windows-release', 'partflow-api.exe');
  }
  return path.join(process.resourcesPath, 'backend', 'partflow-api.exe');
}

function getLocalDatabasePath() {
  return path.join(app.getPath('userData'), 'data', 'partflow.db');
}

function getBundledDatabasePath() {
  return isDev
    ? path.join(__dirname, '..', '..', 'local-partflow.db')
    : path.join(process.resourcesPath, 'partflow.db');
}

function getProductImagesPath() {
  return path.join(app.getPath('userData'), 'data', 'product-images');
}

function getPartTypeImagesPath() {
  return path.join(app.getPath('userData'), 'data', 'part-type-images');
}

function getCategoryImagesPath() {
  return path.join(app.getPath('userData'), 'data', 'category-images');
}

function getBackendLogPath() {
  return path.join(app.getPath('userData'), 'logs', 'backend.log');
}

function appendBackendLog(message) {
  try {
    const logPath = getBackendLogPath();
    fs.mkdirSync(path.dirname(logPath), { recursive: true });
    fs.appendFileSync(logPath, `[${new Date().toISOString()}] ${message}\n`, 'utf8');
  } catch {
    // Logging must never prevent the desktop app from starting.
  }
}

function safeProductId(productId) {
  const value = String(productId || '').trim();
  return /^[a-zA-Z0-9_-]+$/.test(value) ? value : null;
}

function dataUrlToBuffer(dataUrl) {
  const match = String(dataUrl || '').match(/^data:image\/(jpeg|jpg|png|webp);base64,(.+)$/);
  if (!match) return null;
  return Buffer.from(match[2], 'base64');
}

async function readProductImages() {
  const imagesPath = getProductImagesPath();
  await fs.promises.mkdir(imagesPath, { recursive: true });
  const files = await fs.promises.readdir(imagesPath);
  const result = {};
  for (const file of files) {
    if (!file.endsWith('.jpg')) continue;
    const productId = file.slice(0, -4);
    const image = await fs.promises.readFile(path.join(imagesPath, file));
    result[productId] = `data:image/jpeg;base64,${image.toString('base64')}`;
  }
  return result;
}

ipcMain.handle('product-images:list', () => readProductImages());
ipcMain.handle('product-images:save', async (_event, productId, dataUrl) => {
  const safeId = safeProductId(productId);
  const image = dataUrlToBuffer(dataUrl);
  if (!safeId || !image || image.length > 5 * 1024 * 1024) {
    throw new Error('Invalid product image');
  }

  const imagesPath = getProductImagesPath();
  await fs.promises.mkdir(imagesPath, { recursive: true });
  const target = path.join(imagesPath, `${safeId}.jpg`);
  const temporary = `${target}.tmp`;
  await fs.promises.writeFile(temporary, image);
  await fs.promises.rm(target, { force: true });
  await fs.promises.rename(temporary, target);
  return `data:image/jpeg;base64,${image.toString('base64')}`;
});
ipcMain.handle('product-images:delete', async (_event, productId) => {
  const safeId = safeProductId(productId);
  if (!safeId) return false;
  try {
    await fs.promises.unlink(path.join(getProductImagesPath(), `${safeId}.jpg`));
  } catch (error) {
    if (error.code !== 'ENOENT') throw error;
  }
  return true;
});

async function readPartTypeImages() {
  const imagesPath = getPartTypeImagesPath();
  await fs.promises.mkdir(imagesPath, { recursive: true });
  const files = await fs.promises.readdir(imagesPath);
  const result = {};
  for (const file of files) {
    if (!file.endsWith('.jpg')) continue;
    const partTypeId = file.slice(0, -4);
    const image = await fs.promises.readFile(path.join(imagesPath, file));
    result[partTypeId] = `data:image/jpeg;base64,${image.toString('base64')}`;
  }
  return result;
}

ipcMain.handle('part-type-images:list', () => readPartTypeImages());
ipcMain.handle('part-type-images:save', async (_event, partTypeId, dataUrl) => {
  const safeId = safeProductId(partTypeId);
  const image = dataUrlToBuffer(dataUrl);
  if (!safeId || !image || image.length > 5 * 1024 * 1024) {
    throw new Error('Invalid part type image');
  }

  const imagesPath = getPartTypeImagesPath();
  await fs.promises.mkdir(imagesPath, { recursive: true });
  const target = path.join(imagesPath, `${safeId}.jpg`);
  const temporary = `${target}.tmp`;
  await fs.promises.writeFile(temporary, image);
  await fs.promises.rm(target, { force: true });
  await fs.promises.rename(temporary, target);
  return `data:image/jpeg;base64,${image.toString('base64')}`;
});
ipcMain.handle('part-type-images:delete', async (_event, partTypeId) => {
  const safeId = safeProductId(partTypeId);
  if (!safeId) return false;
  try {
    await fs.promises.unlink(path.join(getPartTypeImagesPath(), `${safeId}.jpg`));
  } catch (error) {
    if (error.code !== 'ENOENT') throw error;
  }
  return true;
});

async function readCategoryImages() {
  const imagesPath = getCategoryImagesPath();
  await fs.promises.mkdir(imagesPath, { recursive: true });
  const files = await fs.promises.readdir(imagesPath);
  const result = {};
  for (const file of files) {
    if (!file.endsWith('.jpg')) continue;
    const categoryId = file.slice(0, -4);
    const image = await fs.promises.readFile(path.join(imagesPath, file));
    result[categoryId] = `data:image/jpeg;base64,${image.toString('base64')}`;
  }
  return result;
}

ipcMain.handle('category-images:list', () => readCategoryImages());
ipcMain.handle('category-images:save', async (_event, categoryId, dataUrl) => {
  const safeId = safeProductId(categoryId);
  const image = dataUrlToBuffer(dataUrl);
  if (!safeId || !image || image.length > 5 * 1024 * 1024) {
    throw new Error('Invalid category image');
  }

  const imagesPath = getCategoryImagesPath();
  await fs.promises.mkdir(imagesPath, { recursive: true });
  const target = path.join(imagesPath, `${safeId}.jpg`);
  const temporary = `${target}.tmp`;
  await fs.promises.writeFile(temporary, image);
  await fs.promises.rm(target, { force: true });
  await fs.promises.rename(temporary, target);
  return `data:image/jpeg;base64,${image.toString('base64')}`;
});
ipcMain.handle('category-images:delete', async (_event, categoryId) => {
  const safeId = safeProductId(categoryId);
  if (!safeId) return false;
  try {
    await fs.promises.unlink(path.join(getCategoryImagesPath(), `${safeId}.jpg`));
  } catch (error) {
    if (error.code !== 'ENOENT') throw error;
  }
  return true;
});

function escapeInvoiceHtml(value) {
  return String(value ?? '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

function formatInvoiceNumber(value) {
  const normalized = String(value || '').trim();
  return normalized || '-';
}

function formatInvoiceAmount(value) {
  return Number(value || 0).toLocaleString('en-US', { maximumFractionDigits: 2 });
}

function buildSupplierInvoiceHtml(payload) {
  const purchase = payload?.purchase || {};
  const supplier = payload?.supplier || {};
  const items = Array.isArray(payload?.items) ? payload.items : [];
  const total = Number(purchase.total_amount || 0);
  const paid = Number(purchase.paid_amount || 0);
  const remaining = Math.max(0, Number(purchase.remaining ?? total - paid));
  const paymentStatus = remaining <= 0 ? 'مدفوعة' : paid > 0 ? 'مدفوعة جزئيًا' : 'غير مدفوعة';
  const invoiceNumber = formatInvoiceNumber(purchase.invoice_number || purchase.purchase_number);
  const purchaseId = formatInvoiceNumber(purchase.id);
  const purchaseDate = purchase.purchase_date
    ? new Date(purchase.purchase_date).toLocaleDateString('ar-SA')
    : '-';
  const rows = items.map((item) => {
    const quantity = Number(item.quantity || 0);
    const unitCost = Number(item.unit_cost || 0);
    const productName = item.product_name || item.product?.name || 'قطعة';
    const sku = item.sku || item.product?.sku || '';
    return `<tr>
      <td><strong>${escapeInvoiceHtml(productName)}</strong>${sku ? `<small>SKU: ${escapeInvoiceHtml(sku)}</small>` : ''}</td>
      <td>${quantity}</td>
      <td>₪${formatInvoiceAmount(unitCost)}</td>
      <td>₪${formatInvoiceAmount(quantity * unitCost)}</td>
    </tr>`;
  }).join('');
  const subtotal = Number(purchase.subtotal ?? items.reduce((sum, item) => sum + Number(item.quantity || 0) * Number(item.unit_cost || 0), 0));
  const tax = Number(purchase.tax_amount || 0);
  const discount = Number(purchase.discount_amount || 0);

  return `<!doctype html>
<html lang="ar" dir="rtl">
<head><meta charset="utf-8"><title>${escapeInvoiceHtml(invoiceNumber)}</title>
<style>
  @page { size: A4; margin: 14mm; }
  * { box-sizing: border-box; }
  body { margin: 0; color: #172033; font-family: "Segoe UI", Tahoma, Arial, sans-serif; font-size: 12px; direction: rtl; }
  .header { display: flex; justify-content: space-between; gap: 24px; border-bottom: 2px solid #172033; padding-bottom: 16px; }
  h1 { margin: 0 0 6px; font-size: 24px; }
  h2 { margin: 0; font-size: 16px; }
  .muted { color: #5d687a; }
  .meta { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin: 18px 0; }
  .meta div { border: 1px solid #d8dee8; padding: 10px; border-radius: 6px; }
  .label { display: block; color: #5d687a; font-size: 10px; margin-bottom: 4px; }
  table { width: 100%; border-collapse: collapse; margin-top: 16px; }
  th, td { border: 1px solid #cbd3df; padding: 9px 10px; text-align: right; }
  th { background: #eef2f7; font-weight: 700; }
  td small { display: block; color: #5d687a; margin-top: 3px; direction: ltr; text-align: right; }
  .totals { width: 310px; margin-inline-start: auto; margin-top: 20px; }
  .total-row { display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px solid #e1e6ee; }
  .total-row.final { font-size: 15px; font-weight: 700; border-bottom: 2px solid #172033; }
  .status { margin-top: 16px; font-weight: 700; }
  .footer { margin-top: 36px; color: #5d687a; border-top: 1px solid #d8dee8; padding-top: 10px; }
  @media print { .no-print { display: none !important; } }
</style></head>
<body>
  <header class="header">
    <div><h1>PartFlow</h1><div class="muted">فاتورة شراء من مورد</div></div>
    <div><h2>رقم فاتورة المورد: ${escapeInvoiceHtml(invoiceNumber)}</h2><div class="muted">رقم عملية الشراء: ${escapeInvoiceHtml(purchaseId)}</div></div>
  </header>
  <section class="meta">
    <div><span class="label">المورد</span><strong>${escapeInvoiceHtml(supplier.name || purchase.supplier_name || '-')}</strong></div>
    <div><span class="label">تاريخ الفاتورة</span><strong>${escapeInvoiceHtml(purchaseDate)}</strong></div>
    <div><span class="label">حالة الدفع</span><strong>${paymentStatus}</strong></div>
    <div><span class="label">العملة</span><strong>₪</strong></div>
  </section>
  <table><thead><tr><th>المنتج</th><th>الكمية</th><th>سعر الشراء</th><th>الإجمالي</th></tr></thead><tbody>${rows}</tbody></table>
  <section class="totals">
    <div class="total-row"><span>الإجمالي قبل الضريبة</span><strong>₪${formatInvoiceAmount(subtotal)}</strong></div>
    ${discount ? `<div class="total-row"><span>الخصم</span><strong>-₪${formatInvoiceAmount(discount)}</strong></div>` : ''}
    <div class="total-row"><span>الضريبة</span><strong>₪${formatInvoiceAmount(tax)}</strong></div>
    <div class="total-row final"><span>الإجمالي النهائي</span><strong>₪${formatInvoiceAmount(total)}</strong></div>
    <div class="total-row"><span>المدفوع</span><strong>₪${formatInvoiceAmount(paid)}</strong></div>
    <div class="total-row"><span>المتبقي</span><strong>₪${formatInvoiceAmount(remaining)}</strong></div>
    <div class="status">الحالة: ${paymentStatus}</div>
  </section>
  <footer class="footer">تم إنشاء هذه الفاتورة بواسطة PartFlow</footer>
</body></html>`;
}

async function prepareInvoiceHtml(html) {
  const source = String(html || '');
  if (!source.includes("/fonts/NotoNaskhArabic.ttf")) return source;

  const fontPaths = [
    path.join(__dirname, '..', 'public', 'fonts', 'NotoNaskhArabic.ttf'),
    path.join(app.getAppPath(), 'dist', 'fonts', 'NotoNaskhArabic.ttf'),
  ];
  for (const fontPath of fontPaths) {
    try {
      const fontData = await fs.promises.readFile(fontPath);
      return source.replaceAll('/fonts/NotoNaskhArabic.ttf', `data:font/ttf;base64,${fontData.toString('base64')}`);
    } catch {
      // Try the next packaged or development font location.
    }
  }
  return source;
}

async function createInvoiceWindow(html) {
  const invoiceWindow = new BrowserWindow({
    show: false,
    width: 900,
    height: 1200,
    webPreferences: { sandbox: true },
  });
  await invoiceWindow.loadURL(`data:text/html;charset=utf-8,${encodeURIComponent(await prepareInvoiceHtml(html))}`);
  return invoiceWindow;
}

ipcMain.handle('document:print-html', async (_event, html) => {
  const printWindow = await createInvoiceWindow(String(html || ''));
  try {
    return await new Promise((resolve, reject) => {
      printWindow.webContents.print({ silent: false, printBackground: true }, (success, failureReason) => {
        if (!success) reject(new Error(failureReason || 'تعذر فتح نظام الطباعة في Windows.'));
        else resolve(true);
      });
    });
  } finally {
    if (!printWindow.isDestroyed()) printWindow.close();
  }
});

function buildInvoiceHtml(document) {
  const invoice = document || {};
  const items = Array.isArray(invoice.items) ? invoice.items : [];
  const rows = items.map((item) => `<tr>
    <td><strong>${escapeInvoiceHtml(item.name || 'منتج')}</strong>${item.sku || item.barcode ? `<small>${escapeInvoiceHtml(item.sku ? `SKU: ${item.sku}` : `Barcode: ${item.barcode}`)}</small>` : ''}</td>
    <td>${formatInvoiceAmount(item.quantity)}</td>
    <td>₪${formatInvoiceAmount(item.unitPrice)}</td>
    <td>₪${formatInvoiceAmount(item.total)}</td>
  </tr>`).join('');
  const date = invoice.date ? new Date(invoice.date).toLocaleDateString('ar-SA') : '-';
  return `<!doctype html><html lang="ar" dir="rtl"><head><meta charset="utf-8"><title>${escapeInvoiceHtml(invoice.invoiceNumber)}</title>
<style>@page{size:A4;margin:14mm}*{box-sizing:border-box}body{margin:0;color:#172033;font-family:"Segoe UI",Tahoma,Arial,sans-serif;font-size:12px;direction:rtl}.header{display:flex;justify-content:space-between;gap:24px;border-bottom:2px solid #172033;padding-bottom:16px}h1{margin:0 0 6px;font-size:24px}h2{margin:0;font-size:16px}.muted{color:#5d687a}.meta{display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin:18px 0}.meta div{border:1px solid #d8dee8;padding:10px;border-radius:6px}.label{display:block;color:#5d687a;font-size:10px;margin-bottom:4px}table{width:100%;border-collapse:collapse;margin-top:16px}th,td{border:1px solid #cbd3df;padding:9px 10px;text-align:right}th{background:#eef2f7;font-weight:700}td small{display:block;color:#5d687a;margin-top:3px;direction:ltr;text-align:right}.totals{width:310px;margin-inline-start:auto;margin-top:20px}.total-row{display:flex;justify-content:space-between;padding:6px 0;border-bottom:1px solid #e1e6ee}.total-row.final{font-size:15px;font-weight:700;border-bottom:2px solid #172033}.footer{margin-top:36px;color:#5d687a;border-top:1px solid #d8dee8;padding-top:10px}</style></head>
<body><header class="header"><div><h1>PartFlow</h1><div class="muted">${escapeInvoiceHtml(invoice.title || 'فاتورة')}</div></div><div><h2>رقم الفاتورة: ${escapeInvoiceHtml(invoice.invoiceNumber)}</h2><div class="muted">رقم العملية: ${escapeInvoiceHtml(invoice.transactionId || '-')}</div></div></header>
<section class="meta"><div><span class="label">${escapeInvoiceHtml(invoice.partyLabel || 'الطرف')}</span><strong>${escapeInvoiceHtml(invoice.partyName || '-')}</strong></div><div><span class="label">التاريخ</span><strong>${escapeInvoiceHtml(date)}</strong></div><div><span class="label">حالة الدفع</span><strong>${escapeInvoiceHtml(invoice.status || '-')}</strong></div><div><span class="label">طريقة الدفع</span><strong>${escapeInvoiceHtml(invoice.paymentMethod || '-')}</strong></div></section>
<table><thead><tr><th>المنتج</th><th>الكمية</th><th>سعر الوحدة</th><th>الإجمالي</th></tr></thead><tbody>${rows}</tbody></table>
<section class="totals"><div class="total-row"><span>الإجمالي قبل الضريبة</span><strong>₪${formatInvoiceAmount(invoice.subtotal)}</strong></div>${invoice.discount ? `<div class="total-row"><span>الخصم</span><strong>-₪${formatInvoiceAmount(invoice.discount)}</strong></div>` : ''}<div class="total-row"><span>الضريبة</span><strong>₪${formatInvoiceAmount(invoice.tax)}</strong></div><div class="total-row final"><span>الإجمالي النهائي</span><strong>₪${formatInvoiceAmount(invoice.total)}</strong></div><div class="total-row"><span>المدفوع</span><strong>₪${formatInvoiceAmount(invoice.paid)}</strong></div><div class="total-row"><span>المتبقي</span><strong>₪${formatInvoiceAmount(invoice.remaining)}</strong></div></section><footer class="footer">تم إنشاء هذه الفاتورة بواسطة PartFlow</footer></body></html>`;
}

function safeInvoiceFileName(payload) {
  const purchase = payload?.purchase || {};
  const invoice = String(purchase.invoice_number || purchase.purchase_number || '').trim();
  const fallback = String(purchase.id || 'XXXXXXXX').replace(/[^a-zA-Z0-9_-]/g, '').slice(0, 12) || 'XXXXXXXX';
  const value = invoice || `PO-${fallback}`;
  return `فاتورة-مورد-${value.replace(/[<>:"/\\|?*]/g, '-')}.pdf`;
}

ipcMain.handle('invoice:print', async (_event, payload) => {
  const invoiceWindow = await createInvoiceWindow(String(payload?.html || ''));
  try {
    return await new Promise((resolve, reject) => {
      invoiceWindow.webContents.print({ silent: false, printBackground: true }, (success, failureReason) => {
        if (!success) reject(new Error(failureReason || 'تعذر فتح نظام الطباعة في Windows.'));
        else resolve(true);
      });
    });
  } finally {
    if (!invoiceWindow.isDestroyed()) invoiceWindow.close();
  }
});

ipcMain.handle('invoice:save-pdf', async (_event, payload) => {
  const defaultName = String(payload?.fileName || safeInvoiceFileName(payload));
  const defaultPath = path.join(app.getPath('documents'), defaultName);
  const result = await dialog.showSaveDialog(appState.mainWindow, {
    title: 'حفظ الفاتورة كـ PDF',
    defaultPath,
    filters: [{ name: 'PDF', extensions: ['pdf'] }],
  });
  if (result.canceled || !result.filePath) return { canceled: true };

  const invoiceWindow = await createInvoiceWindow(String(payload?.html || ''));
  try {
    const pdf = await invoiceWindow.webContents.printToPDF({
      printBackground: true,
      preferCSSPageSize: true,
      pageSize: 'A4',
      margins: { top: 0, bottom: 0, left: 0, right: 0 },
    });
    await fs.promises.writeFile(result.filePath, pdf);
    return { canceled: false, filePath: result.filePath };
  } finally {
    if (!invoiceWindow.isDestroyed()) invoiceWindow.close();
  }
});

async function waitForBackend() {
  for (let attempt = 0; attempt < 30; attempt += 1) {
    try {
      const response = await fetch(`http://127.0.0.1:${backendPort}/health`);
      if (response.ok) return true;
    } catch {
      // The process may still be starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  return false;
}

function stopBackend() {
  if (!appState.backendProcess || appState.backendProcess.killed) return;
  appState.backendProcess.kill();
  appState.backendProcess = null;
}

async function startBackend() {
  if (appState.backendStarting) return appState.backendStarting;

  appState.backendStarting = (async () => {
    stopBackend();
    killExistingBackendProcesses();
    const backendPath = getBackendPath();
    if (!fs.existsSync(backendPath)) {
      throw new Error(`Backend executable not found: ${backendPath}`);
    }

    appendBackendLog(`starting backend: ${backendPath}`);

    const backendEnv = {
      ...process.env,
      SERVER_PORT: String(backendPort),
      // The embedded backend always serves the local SQLite database. The
      // cloud is used for authentication, subscription checks, and sync.
      DB_CONNECTION_MODE: 'local',
      PARTFLOW_REQUIRE_CLOUD_AUTH: 'true',
      PARTFLOW_CLOUD_API_URL: 'https://partflow-api.onrender.com/api/v1',
      PARTFLOW_LOCAL_DB_PATH: getLocalDatabasePath(),
    };
    delete backendEnv.DATABASE_URL;
    delete backendEnv.DATABASE_URL_CLOUD;

    fs.mkdirSync(path.dirname(getLocalDatabasePath()), { recursive: true });
    if (!fs.existsSync(getLocalDatabasePath()) && fs.existsSync(getBundledDatabasePath())) {
      fs.copyFileSync(getBundledDatabasePath(), getLocalDatabasePath());
    }
    appState.backendProcess = spawn(backendPath, [], {
      env: backendEnv,
      cwd: path.dirname(backendPath),
      windowsHide: true,
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    appState.backendProcess.stdout?.on('data', (chunk) => appendBackendLog(String(chunk).trimEnd()));
    appState.backendProcess.stderr?.on('data', (chunk) => appendBackendLog(String(chunk).trimEnd()));
    appState.backendProcess.once('error', (error) => appendBackendLog(`process error: ${error.message}`));
    appState.backendProcess.once('exit', (code, signal) => {
      appendBackendLog(`process exited: code=${code ?? 'null'} signal=${signal ?? 'null'}`);
      appState.backendProcess = null;
    });

    if (!(await waitForBackend())) {
      appendBackendLog('health check timed out after backend start');
      stopBackend();
      throw new Error('تعذر تشغيل خدمة PartFlow المحلية.');
    }
    appendBackendLog(`backend healthy on port ${backendPort}`);
    return true;
  })();

  try {
    return await appState.backendStarting;
  } finally {
    appState.backendStarting = null;
  }
}

function ensureTrayIcon() {
  const iconPath = isDev
    ? path.join(__dirname, '..', 'public', 'partflow-logo.png')
    : path.join(process.resourcesPath, 'partflow-logo.png');

  try {
    return nativeImage.createFromPath(iconPath).resize({ width: 18, height: 18 });
  } catch {
    return null;
  }
}

function updateWindowTitle(window) {
  if (!window || window.isDestroyed()) return;
  window.setTitle('PartFlow - إدارة محلك بطريقة أذكى');
}

function updateTray() {
  if (!appState.tray) return;

  const trayImage = ensureTrayIcon();
  if (trayImage) {
    appState.tray.setImage(trayImage);
  }

  appState.tray.setToolTip('PartFlow');
  const menu = Menu.buildFromTemplate([
    {
      label: 'فتح التطبيق',
      click: () => {
        const [window] = BrowserWindow.getAllWindows();
        if (window) {
          if (window.isMinimized()) window.restore();
          window.show();
          window.focus();
        }
      },
    },
    { type: 'separator' },
    { label: 'خروج', role: 'quit' },
  ]);

  appState.tray.setContextMenu(menu);
}

function createSplashWindow() {
  const splash = new BrowserWindow({
    width: 520,
    height: 300,
    frame: false,
    transparent: true,
    resizable: false,
    alwaysOnTop: true,
    center: true,
    show: false,
    movable: false,
    skipTaskbar: true,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: false,
    },
  });

  const splashHtml = `
    <!DOCTYPE html>
    <html lang="ar" dir="rtl">
      <head>
        <meta charset="UTF-8" />
        <style>
          :root {
            --bg: #0b1120;
            --panel: rgba(15, 23, 42, 0.92);
            --line: rgba(148, 163, 184, 0.22);
            --primary: #60a5fa;
            --accent: #22c55e;
            --text: #e2e8f0;
            --muted: #94a3b8;
          }
          * { box-sizing: border-box; }
          body {
            margin: 0;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            font-family: Tahoma, Arial, sans-serif;
            background: radial-gradient(circle at top, rgba(96, 165, 250, 0.22), transparent 38%), var(--bg);
          }
          .shell {
            width: 480px;
            background: var(--panel);
            border: 1px solid var(--line);
            border-radius: 26px;
            box-shadow: 0 30px 80px rgba(15, 23, 42, 0.5);
            padding: 26px 26px 20px;
            backdrop-filter: blur(12px);
          }
          .brand {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 14px;
            margin-bottom: 18px;
          }
          .logo {
            width: 52px;
            height: 52px;
            border-radius: 16px;
            display: grid;
            place-items: center;
            background: linear-gradient(135deg, #2563eb, #38bdf8);
            color: white;
            font-size: 22px;
            font-weight: 800;
            box-shadow: 0 18px 30px rgba(37, 99, 235, 0.35);
          }
          h1 {
            margin: 0;
            color: var(--text);
            font-size: 28px;
            letter-spacing: 0.02em;
          }
          p {
            text-align: center;
            color: var(--muted);
            margin: 0 0 20px;
            font-size: 14px;
          }
          .bar {
            width: 100%;
            height: 8px;
            border-radius: 999px;
            background: rgba(148, 163, 184, 0.14);
            overflow: hidden;
          }
          .bar > span {
            display: block;
            width: 38%;
            height: 100%;
            border-radius: inherit;
            background: linear-gradient(90deg, var(--primary), var(--accent));
            animation: pulse 1.6s ease-in-out infinite alternate;
          }
          @keyframes pulse {
            from { transform: translateX(-30%); }
            to { transform: translateX(190%); }
          }
        </style>
      </head>
      <body>
        <div class="shell">
          <div class="brand">
            <div class="logo">P</div>
            <h1>PartFlow</h1>
          </div>
          <p>جاري تجهيز متجر التشغيل وتهيئة الاتصال المحلي...</p>
          <div class="bar"><span></span></div>
        </div>
      </body>
    </html>
  `;

  splash.loadURL(`data:text/html;charset=utf-8,${encodeURIComponent(splashHtml)}`);
  splash.once('ready-to-show', () => splash.show());
  appState.splashWindow = splash;
  return splash;
}

function createWindow() {
  const mainWindow = new BrowserWindow({
    width: 1500,
    height: 980,
    minWidth: 1100,
    minHeight: 720,
    backgroundColor: '#0f172a',
    icon: ensureTrayIcon() || undefined,
    title: 'PartFlow - إدارة محلك بطريقة أذكى',
    titleBarStyle: 'hiddenInset',
    autoHideMenuBar: true,
    show: false,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: false,
    },
  });

  appState.mainWindow = mainWindow;
  mainWindow.setMenuBarVisibility(false);
  updateWindowTitle(mainWindow);

  const loadRenderer = isDev
    ? mainWindow.loadURL('http://localhost:5174')
    : mainWindow.loadFile(path.join(__dirname, '..', 'dist', 'index.html'));

  loadRenderer.catch(() => {
    console.error('Failed to load PartFlow desktop app.');
  });

  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: 'deny' };
  });

  mainWindow.on('ready-to-show', () => {
    if (appState.splashWindow && !appState.splashWindow.isDestroyed()) {
      setTimeout(() => {
        appState.splashWindow.close();
        appState.splashWindow = null;
      }, 650);
    }

    mainWindow.show();
    mainWindow.focus();
  });

  mainWindow.on('focus', () => {
    updateWindowTitle(mainWindow);
  });

  return mainWindow;
}

app.whenReady().then(async () => {
  await fs.promises.mkdir(getProductImagesPath(), { recursive: true });
  killExistingBackendProcesses();
  try {
    await startBackend();
  } catch (error) {
    console.error(error);
  }
  createSplashWindow();

  const trayIcon = ensureTrayIcon();
  if (trayIcon) {
    appState.tray = new Tray(trayIcon);
    updateTray();
  }

  const mainWindow = createWindow();

  appState.tray?.on('click', () => {
    if (mainWindow.isMinimized()) mainWindow.restore();
    mainWindow.show();
    mainWindow.focus();
  });

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('before-quit', () => {
  stopBackend();
});
