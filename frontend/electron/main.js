import { app, BrowserWindow, shell, Tray, Menu, nativeImage } from 'electron';
import { spawn, execSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const isDev = !app.isPackaged;

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
      windowsHide: true,
      stdio: 'ignore',
    });
    appState.backendProcess.once('exit', () => {
      appState.backendProcess = null;
    });

    if (!(await waitForBackend())) {
      stopBackend();
      throw new Error('تعذر تشغيل خدمة PartFlow المحلية.');
    }
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
