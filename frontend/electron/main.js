import { app, BrowserWindow, shell, Tray, Menu, nativeImage, ipcMain, dialog, protocol, net } from 'electron';
import { spawn, execFileSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const isDev = !app.isPackaged;

// A standard secure origin lets the desktop app use an exact CORS allowlist.
// Serving the renderer as file:// would give it the opaque "null" origin,
// which is also available to sandboxed web content and must not be trusted.
protocol.registerSchemesAsPrivileged([{
  scheme: 'partflow',
  privileges: {
    standard: true,
    secure: true,
    supportFetchAPI: true,
    corsEnabled: true,
    allowServiceWorkers: true,
    codeCache: true,
  },
}]);

// PartFlow is a data-management desktop app and does not require GPU
// acceleration. Disabling it before the app is ready avoids a hard Electron
// shutdown on Windows machines where the GPU process cannot load its driver
// (for example, on a VM or a machine with a broken graphics stack).
app.commandLine.appendSwitch('disable-gpu');
app.commandLine.appendSwitch('disable-gpu-compositing');
app.commandLine.appendSwitch('in-process-gpu');
app.disableHardwareAcceleration();

app.name = 'PartFlow';
// Pin Electron's persistent user-data location so database, image, log, and
// offline-key paths remain stable across install directories and app updates.
app.setPath('userData', path.join(app.getPath('appData'), 'PartFlow'));

const appState = {
  tray: null,
  splashWindow: null,
  mainWindow: null,
  backendProcess: null,
  backendStarting: null,
  backendEnv: null,
  databaseOperation: false,
  rendererRecoveryPromptOpen: false,
};

const backendPort = 8080;
const hasSingleInstanceLock = app.requestSingleInstanceLock();

if (!hasSingleInstanceLock) {
  app.quit();
} else {
  app.on('second-instance', () => {
    const window = appState.mainWindow;
    if (!window) return;
    if (window.isMinimized()) window.restore();
    window.show();
    window.focus();
  });
}

function getListeningBackendPids() {
  if (process.platform !== 'win32') return new Set();

  try {
    const output = execFileSync('netstat.exe', ['-ano', '-p', 'tcp'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    });
    const pids = new Set();
    for (const line of String(output).split(/\r?\n/)) {
      const columns = line.trim().split(/\s+/);
      if (
        columns.length < 5 ||
        columns[0].toUpperCase() !== 'TCP' ||
        !columns[1].endsWith(`:${backendPort}`) ||
        columns[3].toUpperCase() !== 'LISTENING' ||
        !/^\d+$/.test(columns[4])
      ) {
        continue;
      }
      pids.add(columns[4]);
    }
    return pids;
  } catch {
    return new Set();
  }
}

function getWindowsProcessExecutablePath(pid) {
  try {
    const command = `$process = Get-CimInstance -ClassName Win32_Process -Filter "ProcessId = ${pid}" -ErrorAction SilentlyContinue; if ($process) { [Console]::Write($process.ExecutablePath) }`;
    return execFileSync('powershell.exe', ['-NoProfile', '-NonInteractive', '-Command', command], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    }).trim();
  } catch {
    return '';
  }
}

function killExistingBackendProcesses() {
  if (process.platform !== 'win32') return;

  const expectedBackendPath = path.win32.resolve(getBackendPath()).toLowerCase();
  for (const pid of getListeningBackendPids()) {
    const executablePath = getWindowsProcessExecutablePath(pid);
    if (!executablePath || path.win32.resolve(executablePath).toLowerCase() !== expectedBackendPath) {
      continue;
    }

    try {
      execFileSync('taskkill.exe', ['/PID', pid, '/F', '/T'], { stdio: 'ignore' });
    } catch {
      // Ignore cleanup failures; startup will report if the PartFlow backend remains unavailable.
    }
  }
}

function getBackendPath() {
  if (isDev) {
    return path.join(__dirname, '..', '..', 'dist', 'windows-release', 'partflow-api.exe');
  }
  return path.join(process.resourcesPath, 'backend', 'partflow-api.exe');
}

function getLocalDatabasePath() {
  return getUserDataPath('data', 'partflow.db');
}

function getUserDataPath(...segments) {
  return path.join(app.getPath('userData'), ...segments);
}

function getBundledDatabasePath() {
  return isDev
    ? path.join(__dirname, '..', '..', 'local-partflow.db')
    : path.join(process.resourcesPath, 'partflow.db');
}

function getProductImagesPath() {
  return getUserDataPath('data', 'product-images');
}

function getPartTypeImagesPath() {
  return getUserDataPath('data', 'part-type-images');
}

function getCategoryImagesPath() {
  return getUserDataPath('data', 'category-images');
}

function getBackendLogPath() {
  return getUserDataPath('logs', 'backend.log');
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

function runBackendMaintenance(args) {
  const backendPath = getBackendPath();
  if (!fs.existsSync(backendPath)) throw new Error(`Backend executable not found: ${backendPath}`);
  return execFileSync(backendPath, args, {
    env: appState.backendEnv || { ...process.env, PARTFLOW_LOCAL_DB_PATH: getLocalDatabasePath() },
    cwd: path.dirname(backendPath),
    windowsHide: true,
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
    timeout: 120_000,
  }).trim();
}

async function stopBackendForDatabaseOperation() {
  const backend = appState.backendProcess;
  if (!backend || backend.exitCode !== null || backend.signalCode !== null) {
    appState.backendProcess = null;
    return;
  }

  const exited = new Promise((resolve) => {
    backend.once('exit', resolve);
    backend.once('error', resolve);
  });
  backend.kill('SIGTERM');
  const timeout = new Promise((resolve) => setTimeout(resolve, 10_000));
  await Promise.race([exited, timeout]);
  if (backend.exitCode === null && backend.signalCode === null) {
    backend.kill('SIGKILL');
    await Promise.race([exited, new Promise((resolve) => setTimeout(resolve, 2_000))]);
  }
  if (appState.backendProcess === backend) appState.backendProcess = null;
}

async function withStoppedBackend(operation) {
  if (appState.databaseOperation) throw new Error('A database operation is already running');
  appState.databaseOperation = true;
  try {
    await appState.backendStarting?.catch(() => {});
    await stopBackendForDatabaseOperation();
    return await operation();
  } finally {
    try {
      await startBackend();
    } finally {
      appState.databaseOperation = false;
    }
  }
}

function timestampForFileName() {
  return new Date().toISOString().replaceAll(':', '-').replaceAll('.', '-');
}

function assertTrustedDatabaseRequest(event) {
  if (!appState.mainWindow || event.sender !== appState.mainWindow.webContents) {
    throw new Error('Database operation is not available to this renderer');
  }
  const senderUrl = event.senderFrame?.url || event.sender.getURL();
  if (isDev) {
    let parsed;
    try {
      parsed = new URL(senderUrl);
    } catch {
      throw new Error('Database operation is not available to this page');
    }
    if (parsed.protocol !== 'http:' || parsed.hostname !== 'localhost' || parsed.port !== '5174') {
      throw new Error('Database operation is not available to this page');
    }
    return;
  }
  let senderPath;
  try {
    senderPath = path.resolve(fileURLToPath(senderUrl));
  } catch {
    throw new Error('Database operation is not available to this page');
  }
  const applicationPagePath = path.resolve(__dirname, '..', 'dist', 'index.html');
  if (senderPath.toLowerCase() !== applicationPagePath.toLowerCase()) {
    throw new Error('Database operation is not available to this page');
  }
}

ipcMain.handle('database:backup', async (event) => {
  assertTrustedDatabaseRequest(event);
  const result = await dialog.showSaveDialog(appState.mainWindow, {
    title: 'حفظ نسخة احتياطية من قاعدة البيانات',
    defaultPath: path.join(app.getPath('documents'), `PartFlow-backup-${timestampForFileName()}.db`),
    filters: [{ name: 'SQLite database', extensions: ['db', 'sqlite'] }],
  });
  if (result.canceled || !result.filePath) return { canceled: true };

  await withStoppedBackend(async () => {
    runBackendMaintenance(['--backup-local-database', result.filePath]);
  });
  return { canceled: false, filePath: result.filePath };
});

ipcMain.handle('database:restore', async (event) => {
  assertTrustedDatabaseRequest(event);
  const selection = await dialog.showOpenDialog(appState.mainWindow, {
    title: 'اختيار نسخة PartFlow للاستعادة',
    properties: ['openFile'],
    filters: [{ name: 'SQLite database', extensions: ['db', 'sqlite'] }],
  });
  if (selection.canceled || selection.filePaths.length === 0) return { canceled: true };

  const selectedPath = selection.filePaths[0];
  runBackendMaintenance(['--validate-local-database', selectedPath]);
  const confirmation = await dialog.showMessageBox(appState.mainWindow, {
    type: 'warning',
    title: 'تأكيد استعادة قاعدة البيانات',
    message: 'سيتم استبدال بيانات المتجر الحالية بالنسخة المحددة.',
    detail: 'سيُحفظ ملف استعادة تلقائي من قاعدة البيانات الحالية قبل الاستبدال.',
    buttons: ['استعادة النسخة', 'إلغاء'],
    defaultId: 1,
    cancelId: 1,
    noLink: true,
  });
  if (confirmation.response !== 0) return { canceled: true };

  const databasePath = getLocalDatabasePath();
  const directory = path.dirname(databasePath);
  const recoveryPath = path.join(directory, `partflow-before-restore-${timestampForFileName()}.db`);
  const stagingPath = `${databasePath}.restore-${process.pid}-${Date.now()}`;
  const displacedPath = `${databasePath}.previous-${process.pid}-${Date.now()}`;
  let installed = false;

  try {
    await withStoppedBackend(async () => {
      runBackendMaintenance(['--backup-local-database', recoveryPath]);
      runBackendMaintenance(['--backup-database-file', selectedPath, stagingPath]);

      for (const suffix of ['-wal', '-shm']) {
        await fs.promises.rm(`${databasePath}${suffix}`, { force: true });
      }
      await fs.promises.rename(databasePath, displacedPath);
      try {
        await fs.promises.rename(stagingPath, databasePath);
        installed = true;
      } catch (error) {
        await fs.promises.rename(displacedPath, databasePath);
        throw error;
      }
    });
  } catch (error) {
    if (installed) {
      try {
        await stopBackendForDatabaseOperation();
        for (const suffix of ['-wal', '-shm']) {
          await fs.promises.rm(`${databasePath}${suffix}`, { force: true });
        }
        await fs.promises.rm(databasePath, { force: true });
        await fs.promises.rename(displacedPath, databasePath);
        await startBackend();
      } catch (rollbackError) {
        appendBackendLog(`database restore rollback failed: ${rollbackError.message}`);
        throw new Error(`Restore failed and automatic rollback needs attention. Recovery copy: ${recoveryPath}. ${error.message}`);
      }
    }
    throw error;
  } finally {
    await fs.promises.rm(stagingPath, { force: true }).catch(() => {});
    if (installed && fs.existsSync(displacedPath)) {
      await fs.promises.rm(displacedPath, { force: true }).catch(() => {});
    }
  }

  return { canceled: false, recoveryPath };
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

function parseStoreTimestamp(value) {
  if (value instanceof Date) return Number.isNaN(value.getTime()) ? null : value;
  const raw = String(value || '').trim();
  if (!raw) return null;
  let normalized = raw
    .replace(/\s+m=[+-]?\d+(?:\.\d+)?$/, '')
    .replace(/\s+[A-Z]{2,5}$/, '')
    .replace(/\s([+-]\d{2}:?\d{2})$/, '$1')
    .replace(' ', 'T')
    .replace(/([+-]\d{2})(\d{2})$/, '$1:$2');
  if (/^\d{4}-\d{2}-\d{2}$/.test(normalized)) {
    normalized += 'T12:00:00Z';
  } else if (/^\d{4}-\d{2}-\d{2}T/.test(normalized) && !/(?:Z|[+-]\d{2}:\d{2})$/i.test(normalized)) {
    // Legacy timestamp strings without an offset are UTC by the storage contract.
    normalized += 'Z';
  }
  const parsed = new Date(normalized);
  return Number.isNaN(parsed.getTime()) ? null : parsed;
}

function formatStoreDate(value) {
  const parsed = parseStoreTimestamp(value);
  if (!parsed) return '-';
  return new Intl.DateTimeFormat('ar-SA', {
    dateStyle: 'short',
    timeZone: 'Asia/Jerusalem',
  }).format(parsed);
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
  const purchaseDate = formatStoreDate(purchase.purchase_date);
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
  const date = formatStoreDate(invoice.date);
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
      TZ: 'UTC',
      SERVER_PORT: String(backendPort),
      // The embedded backend always serves the local SQLite database. The
      // cloud is used for authentication, subscription checks, and sync.
      DB_CONNECTION_MODE: 'local',
      PARTFLOW_REQUIRE_CLOUD_AUTH: 'true',
      PARTFLOW_CLOUD_API_URL: 'https://partflow-api.onrender.com/api/v1',
      PARTFLOW_LOCAL_DB_PATH: getLocalDatabasePath(),
    };
    appState.backendEnv = backendEnv;
    delete backendEnv.DATABASE_URL;
    delete backendEnv.DATABASE_URL_CLOUD;

    fs.mkdirSync(path.dirname(getLocalDatabasePath()), { recursive: true });
    if (!fs.existsSync(getLocalDatabasePath()) && fs.existsSync(getBundledDatabasePath())) {
      fs.copyFileSync(getBundledDatabasePath(), getLocalDatabasePath());
    }
    const backendProcess = spawn(backendPath, [], {
      env: backendEnv,
      cwd: path.dirname(backendPath),
      windowsHide: true,
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    appState.backendProcess = backendProcess;
    backendProcess.stdout?.on('data', (chunk) => appendBackendLog(String(chunk).trimEnd()));
    backendProcess.stderr?.on('data', (chunk) => appendBackendLog(String(chunk).trimEnd()));
    backendProcess.once('error', (error) => appendBackendLog(`process error: ${error.message}`));
    backendProcess.once('exit', (code, signal) => {
      appendBackendLog(`process exited: code=${code ?? 'null'} signal=${signal ?? 'null'}`);
      appState.backendProcess = null;
    });

    const isHealthy = await waitForBackend();
    const ownsBackendPort =
      process.platform !== 'win32' || getListeningBackendPids().has(String(backendProcess.pid));
    const processIsRunning =
      Boolean(backendProcess.pid) &&
      backendProcess.exitCode === null &&
      backendProcess.signalCode === null &&
      appState.backendProcess === backendProcess;
    if (!isHealthy || !ownsBackendPort || !processIsRunning) {
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
    : mainWindow.loadURL('partflow://app/');

  loadRenderer.catch(() => {
    console.error('Failed to load PartFlow desktop app.');
  });

  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: 'deny' };
  });

  mainWindow.webContents.on('render-process-gone', (_event, details) => {
    appendBackendLog(`renderer process exited: reason=${details.reason} exitCode=${details.exitCode}`);
    if (appState.rendererRecoveryPromptOpen || mainWindow.isDestroyed()) return;
    appState.rendererRecoveryPromptOpen = true;
    void dialog.showMessageBox(mainWindow, {
      type: 'error',
      title: 'توقفت واجهة PartFlow',
      message: 'تعذر على نافذة المتجر الاستمرار.',
      detail: 'أعد فتح الواجهة للمتابعة. قاعدة البيانات المحلية بقيت كما هي.',
      buttons: ['إعادة فتح الواجهة', 'إغلاق البرنامج'],
      defaultId: 0,
      cancelId: 1,
      noLink: true,
    }).then(({ response }) => {
      appState.rendererRecoveryPromptOpen = false;
      if (mainWindow.isDestroyed()) return;
      if (response === 0) mainWindow.webContents.reload();
      else mainWindow.close();
    }).catch(() => {
      appState.rendererRecoveryPromptOpen = false;
    });
  });

  mainWindow.on('unresponsive', () => {
    if (appState.rendererRecoveryPromptOpen || mainWindow.isDestroyed()) return;
    appState.rendererRecoveryPromptOpen = true;
    void dialog.showMessageBox(mainWindow, {
      type: 'warning',
      title: 'واجهة PartFlow لا تستجيب',
      message: 'توقفت الواجهة عن الاستجابة.',
      buttons: ['إعادة تحميل الواجهة', 'الانتظار'],
      defaultId: 1,
      cancelId: 1,
      noLink: true,
    }).then(({ response }) => {
      appState.rendererRecoveryPromptOpen = false;
      if (response === 0 && !mainWindow.isDestroyed()) mainWindow.webContents.reload();
    }).catch(() => {
      appState.rendererRecoveryPromptOpen = false;
    });
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
  if (!hasSingleInstanceLock) return;
  if (!isDev) {
    const rendererRoot = path.resolve(__dirname, '..', 'dist');
    protocol.handle('partflow', async (request) => {
      const requestURL = new URL(request.url);
      if (requestURL.host !== 'app') {
        return new Response('Not found', { status: 404 });
      }

      let requestPath;
      try {
        requestPath = decodeURIComponent(requestURL.pathname);
      } catch {
        return new Response('Bad path', { status: 400 });
      }
      if (!requestPath || requestPath === '/') requestPath = '/index.html';

      const assetPath = path.resolve(rendererRoot, `.${requestPath}`);
      const relativePath = path.relative(rendererRoot, assetPath);
      if (!relativePath || relativePath.startsWith('..') || path.isAbsolute(relativePath)) {
        return new Response('Bad path', { status: 400 });
      }

      try {
        const stats = await fs.promises.stat(assetPath);
        if (!stats.isFile()) return new Response('Not found', { status: 404 });
      } catch {
        return new Response('Not found', { status: 404 });
      }

      return net.fetch(pathToFileURL(assetPath).toString());
    });
  }
  await fs.promises.mkdir(getProductImagesPath(), { recursive: true });
  while (true) {
    try {
      await startBackend();
      break;
    } catch (error) {
      console.error(error);
      const { response } = await dialog.showMessageBox({
        type: 'error',
        title: 'تعذر تشغيل PartFlow',
        message: 'لم تبدأ خدمة المتجر المحلية.',
        detail: error instanceof Error ? error.message : String(error),
        buttons: ['إعادة المحاولة', 'إنهاء البرنامج'],
        defaultId: 0,
        cancelId: 1,
        noLink: true,
      });
      if (response !== 0) {
        app.quit();
        return;
      }
    }
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
