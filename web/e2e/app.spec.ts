import { test, expect } from '@playwright/test';

const q = '?fixtures=1';

test.describe('routes reachable (phone)', () => {
  test.use({ viewport: { width: 402, height: 860 } });

  test('inbox shows space chips, urgency sections and the tab bar', async ({ page }) => {
    await page.goto('/' + q);
    await expect(page.getByRole('listitem').filter({ hasText: 'hedr-web' })).toBeVisible();
    await expect(page.getByText('needs you', { exact: true })).toBeVisible();
    await expect(page.locator('.section-label', { hasText: 'working' })).toBeVisible();
    const bar = page.getByRole('navigation', { name: 'primary' });
    await expect(bar.getByRole('button', { name: /Agents/ })).toBeVisible();
    await expect(bar.getByRole('button', { name: /Spaces/ })).toBeVisible();
    await expect(bar.getByRole('button', { name: /Settings/ })).toBeVisible();
  });

  test('agent row opens the pane and the tab bar yields to the composer', async ({ page }) => {
    await page.goto('/' + q);
    await page.getByRole('button', { name: /codex w1:p2/ }).first().click();
    await expect(page).toHaveURL(/\/pane\/w1(%3A|:)p2/);
    await expect(page.getByPlaceholder('Message the agent')).toBeVisible();
    await expect(page.getByRole('navigation', { name: 'primary' })).toHaveCount(0);
  });

  test('pane header switches tabs through the bottom sheet', async ({ page }) => {
    await page.goto('/pane/w1%3Ap2' + q);
    await page.getByRole('button', { name: 'switch tab' }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await dialog.getByRole('menuitem', { name: /server/ }).click();
    await expect(page).toHaveURL(/\/pane\/w1(%3A|:)p3/);
  });

  test('pane header back returns to the inbox', async ({ page }) => {
    await page.goto('/pane/w1%3Ap2' + q);
    await page.getByRole('button', { name: 'back to agents' }).click();
    await expect(page).toHaveURL(/\/$|\/\?fixtures=1$/);
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

  test('space mutation goes through overflow, then the sheet, then a toast', async ({ page }) => {
    await page.goto('/spaces' + q);
    await page.getByRole('button', { name: 'actions for hedr-web' }).click();
    await page.getByRole('menuitem', { name: 'Close space' }).click();
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
    await page.getByRole('navigation', { name: 'primary' }).getByRole('button', { name: /Agents/ }).click();
    await expect(page).toHaveURL(/\/$|\/\?fixtures=1$/);
    await expect(content).toHaveJSProperty('scrollTop', 0);
  });

  test('direct control toggles from the key row and one swipe surfaces the direction toast', async ({ page }) => {
    await page.goto('/pane/w1%3Ap2' + q);
    const toggle = page.getByRole('button', { name: 'direct control' });
    await expect(toggle).toBeVisible();
    await expect(toggle).toHaveAttribute('aria-pressed', 'false');
    await toggle.click();
    await expect(toggle).toHaveAttribute('aria-pressed', 'true');

    const scroll = page.locator('.scroll');
    const box = (await scroll.boundingBox())!;
    const x = box.x + box.width / 2;
    // One swipe, one direction: drag well past the gesture threshold in a single stroke.
    await page.evaluate(
      ([x, yStart, yEnd]) => {
        const el = document.querySelector('.scroll')!;
        const fire = (type: string, y: number) => {
          const t = new Touch({ identifier: 1, target: el, clientX: x, clientY: y });
          el.dispatchEvent(new TouchEvent(type, { touches: type === 'touchend' ? [] : [t], changedTouches: [t], bubbles: true, cancelable: true }));
        };
        fire('touchstart', yStart);
        fire('touchmove', yEnd);
        fire('touchend', yEnd);
      },
      [x, box.y + box.height - 20, box.y + 20]
    );
    await expect(page.getByText('swipe → up')).toBeVisible();
  });

  test('direct control toggle is absent on a non-agent pane', async ({ page }) => {
    await page.goto('/pane/w1%3Ap3' + q);
    await expect(page.getByRole('button', { name: 'direct control' })).toHaveCount(0);
  });
});

test.describe('desktop layout (>=880px)', () => {
  test.use({ viewport: { width: 1200, height: 900 } });

  test('desktop shows the sidebar, resolves / to a chat, and hides the tab bar', async ({ page }) => {
    await page.goto('/' + q);
    await expect(page.locator('aside.sidebar')).toBeVisible();
    await expect(page).toHaveURL(/\/pane\//);
    await expect(page.getByRole('navigation', { name: 'primary' })).toHaveCount(0);
  });
});
