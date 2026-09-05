import { test, expect } from '@playwright/test';

const q = '?fixtures=1';

test.describe('routes reachable (phone)', () => {
  test.use({ viewport: { width: 402, height: 860 } });

  test('inbox shows spaces and agents sections', async ({ page }) => {
    await page.goto('/' + q);
    await expect(page.getByText('spaces', { exact: true })).toBeVisible();
    await expect(page.getByText('agents', { exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: /hedr-web/ }).first()).toBeVisible();
  });

  test('agent row opens the pane (raw terminal)', async ({ page }) => {
    await page.goto('/' + q);
    await page.getByRole('button', { name: /codex w1:p2/ }).first().click();
    await expect(page).toHaveURL(/\/pane\/w1(%3A|:)p2/);
    await expect(page.getByPlaceholder('Message the agent')).toBeVisible();
  });

  test('composer shows nav keys and no free-text chips', async ({ page }) => {
    await page.goto('/pane/w1%3Ap2' + q);
    await expect(page.getByRole('button', { name: 'up' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'enter' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'run tests' })).toHaveCount(0);
  });

  test('pane shows the pane.read caption when dev captions on', async ({ page }) => {
    await page.goto('/settings' + q);
    await page.locator('.row', { hasText: 'Developer captions' }).getByRole('switch').click();
    await page.goto('/pane/w1%3Ap2' + q);
    await expect(page.getByText(/pane\.read . source=recent_unwrapped/)).toBeVisible();
  });

  test('diff viewer toggles wrap', async ({ page }) => {
    await page.goto('/pane/w1%3Ap2/diff' + q);
    const wrap = page.getByRole('button', { name: /wrap (on|off)/ });
    await expect(wrap).toContainText('off');
    await wrap.click();
    await expect(wrap).toContainText('on');
  });

  test('space mutation goes through the sheet then fires a toast', async ({ page }) => {
    await page.goto('/spaces' + q);
    await page.getByRole('button', { name: 'Close' }).first().click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText(/Running processes are killed/)).toBeVisible();
    await page.getByRole('button', { name: 'Close space' }).click();
    await expect(page.getByRole('status')).toBeVisible();
  });

  test('settings and spaces routes render', async ({ page }) => {
    await page.goto('/settings' + q);
    await expect(page.getByText('Theme', { exact: true })).toBeVisible();
    await page.goto('/spaces/w1' + q);
    await expect(page.getByText('AGENTS w1:t1')).toBeVisible();
  });

  test('navigating resets the new route scroll container to top', async ({ page }) => {
    await page.goto('/spaces' + q);
    const content = page.locator('main.content');
    // Force overflow so the offset actually sticks (the fixture list is short —
    // this simulates a user who scrolled a longer real list).
    await content.evaluate((el) => {
      el.style.setProperty('padding-bottom', '2000px');
      el.scrollTop = 200;
    });
    await expect(content).toHaveJSProperty('scrollTop', 200);
    await page.getByRole('button', { name: 'toggle navigation' }).click();
    await page.getByRole('button', { name: 'Agents' }).click();
    await expect(page).toHaveURL(/\/$|\/pane\//);
    await expect(content).toHaveJSProperty('scrollTop', 0);
  });
});

test.describe('desktop layout (>=880px)', () => {
  test.use({ viewport: { width: 1200, height: 900 } });

  test('desktop shows the sidebar and resolves / to a chat', async ({ page }) => {
    await page.goto('/' + q);
    await expect(page.locator('aside.sidebar')).toBeVisible();
    await expect(page).toHaveURL(/\/pane\//);
    // no floating tab bar on desktop
    await expect(page.locator('nav.bar')).toHaveCount(0);
  });
});
