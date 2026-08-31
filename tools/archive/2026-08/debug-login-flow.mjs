import { chromium } from 'playwright';

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
const events = [];

page.on('console', (msg) => events.push(`console:${msg.type()}: ${msg.text()}`));
page.on('pageerror', (err) => events.push(`pageerror:${err.message}`));
page.on('request', (req) => {
  if (req.url().includes('/auth/login') || req.url().includes('/api/v1')) {
    events.push(`request:${req.method()}:${req.url()}`);
  }
});
page.on('response', (res) => {
  if (res.url().includes('/auth/login') || res.url().includes('/api/v1')) {
    events.push(`response:${res.status()}:${res.url()}`);
  }
});

await page.goto('http://localhost:5175/#/login', { waitUntil: 'domcontentloaded', timeout: 30000 });
await page.locator('#email').fill('owner@partflow.com');
await page.locator('#password').fill('Owner123456');

const button = page.locator('button[type="submit"]');
console.log('BUTTON_DISABLED_BEFORE', await button.evaluate((el) => el.disabled));
await button.click({ force: true });
await page.waitForTimeout(6000);

console.log('URL_AFTER', page.url());
console.log('LOCALSTORE', JSON.stringify({
  mode: await page.evaluate(() => localStorage.getItem('partflow-operating-mode')),
  authStorage: await page.evaluate(() => localStorage.getItem('auth-storage')),
  authToken: await page.evaluate(() => localStorage.getItem('auth_token')),
  refreshToken: await page.evaluate(() => localStorage.getItem('refresh_token')),
  offlineSession: await page.evaluate(() => localStorage.getItem('partflow-offline-session')),
}, null, 2));
console.log('EVENTS', JSON.stringify(events.slice(-80), null, 2));

await browser.close();
