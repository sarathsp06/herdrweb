// Capture real screenshots of every screen with Playwright Chromium against a
// LIVE bridge (make run). It discovers a real space + pane from the running
// session, so no fixtures and no hard-coded ids.
// Usage: node scripts/screenshots.mjs [baseURL] [outDir]
import { chromium } from '@playwright/test';
import { mkdir } from 'node:fs/promises';
import { resolve } from 'node:path';

const base = process.argv[2] || 'http://127.0.0.1:7331';
const out = resolve(process.cwd(), process.argv[3] || 'docs/screenshots');
const phone = { viewport: { width: 402, height: 860 }, deviceScaleFactor: 2, colorScheme: 'dark' };
const desktop = { viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2, colorScheme: 'dark' };

await mkdir(out, { recursive: true }).catch(() => {});
const browser = await chromium.launch();

async function shoot(ctx, name, url, wait) {
  const page = await ctx.newPage();
  await page.goto(base + url, { waitUntil: 'networkidle' });
  if (wait) await page.waitForSelector(wait, { timeout: 8000 }).catch(() => {});
  await page.waitForTimeout(500);
  await page.screenshot({ path: `${out}/${name}.png` });
  console.log('wrote', `${out}/${name}.png`);
  await page.close();
}

// Discover a real pane (prefer an agent) by clicking the live inbox, then derive
// its space id from the pane id (`<space>:<pane>`).
const scout = await browser.newContext(phone);
const sp = await scout.newPage();
await sp.goto(base + '/', { waitUntil: 'networkidle' });
await sp.waitForSelector('aside.sidebar', { timeout: 8000 });
const target = (await sp.$('.arow')) || (await sp.$('.srow-main'));
if (!target) throw new Error('no space/agent in the live session — start one, then rerun');
await target.click();
await sp.waitForURL('**/pane/**', { timeout: 8000 });
const paneUrl = new URL(sp.url()).pathname;
const paneId = decodeURIComponent(paneUrl.slice('/pane/'.length));
const spaceId = paneId.split(':')[0];
await scout.close();

const phoneCtx = await browser.newContext(phone);
await shoot(phoneCtx, 'inbox', '/', 'aside.sidebar');
await shoot(phoneCtx, 'spaces', '/spaces', '.list');
await shoot(phoneCtx, 'space-detail', `/spaces/${spaceId}`, '.panes');
await shoot(phoneCtx, 'pane', paneUrl, 'pre.raw');
await shoot(phoneCtx, 'settings', '/settings', '.themes');
await phoneCtx.close();

const deskCtx = await browser.newContext(desktop);
await shoot(deskCtx, 'desktop', paneUrl, 'aside.sidebar');
await deskCtx.close();

await browser.close();
