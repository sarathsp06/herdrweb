// Capture real screenshots of every screen with Playwright Chromium against a
// running bridge in fixtures mode. Usage: node scripts/screenshots.mjs [baseURL] [outDir]
import { chromium } from '@playwright/test';
import { mkdir } from 'node:fs/promises';
import { resolve } from 'node:path';

const base = process.argv[2] || 'http://127.0.0.1:7331';
const out = resolve(process.cwd(), process.argv[3] || 'docs/screenshots');
const q = '?fixtures=1';

const phone = { width: 402, height: 860, deviceScaleFactor: 2 };
const desktop = { width: 1440, height: 900, deviceScaleFactor: 2 };

const shots = [
  { name: 'inbox', url: '/', vp: phone, wait: 'aside.sidebar' },
  { name: 'chat-blocked', url: '/pane/w1%3Ap2', vp: phone, wait: '.blocked' },
  { name: 'chat', url: '/pane/w1%3Ap1', vp: phone, wait: '.transcript' },
  { name: 'diff', url: '/pane/w1%3Ap2/diff', vp: phone, wait: '.diffview .row' },
  { name: 'spaces', url: '/spaces', vp: phone, wait: '.list' },
  { name: 'space-detail', url: '/spaces/w1', vp: phone, wait: '.panes' },
  { name: 'settings', url: '/settings', vp: phone, wait: '.themes' },
  { name: 'desktop', url: '/', vp: desktop, wait: 'aside.sidebar' }
];

await mkdir(out, { recursive: true }).catch(() => {});

const browser = await chromium.launch();
for (const s of shots) {
  const ctx = await browser.newContext({ viewport: { width: s.vp.width, height: s.vp.height }, deviceScaleFactor: s.vp.deviceScaleFactor, colorScheme: 'dark' });
  const page = await ctx.newPage();
  await page.goto(base + s.url + q, { waitUntil: 'networkidle' });
  await page.waitForSelector(s.wait, { timeout: 8000 }).catch(() => {});
  await page.waitForTimeout(500);
  const path = `${out}/${s.name}.png`;
  await page.screenshot({ path });
  console.log('wrote', path);
  await ctx.close();
}
await browser.close();
