// web/tests/snap.spec.js
// Tests: Snap toggle (F3 and button), per-mode checkboxes, localStorage persistence.

const { test, expect } = require('@playwright/test');
const { waitForAppReady, clearAppStorage } = require('./helpers');

test.describe('Snap settings', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await clearAppStorage(page);
    await page.reload();
    await waitForAppReady(page);
    const overlay = page.locator('#welcome-overlay');
    if (await overlay.isVisible()) {
      await page.locator('#btn-welcome-close').click();
    }
  });

  test('snap is ON by default (button has snap-on class)', async ({ page }) => {
    const btn = page.locator('#btn-snap');
    await expect(btn).toHaveClass(/snap-on/);
    await expect(btn).toContainText('Snap');
  });

  test('clicking snap button toggles snap OFF', async ({ page }) => {
    const btn = page.locator('#btn-snap');
    await btn.click();
    await expect(btn).not.toHaveClass(/snap-on/);
  });

  test('clicking snap button twice restores snap ON', async ({ page }) => {
    const btn = page.locator('#btn-snap');
    await btn.click();
    await btn.click();
    await expect(btn).toHaveClass(/snap-on/);
  });

  test('pressing F3 toggles snap OFF', async ({ page }) => {
    const btn = page.locator('#btn-snap');
    await expect(btn).toHaveClass(/snap-on/);
    await page.keyboard.press('F3');
    await expect(btn).not.toHaveClass(/snap-on/);
  });

  test('pressing F3 twice restores snap ON', async ({ page }) => {
    const btn = page.locator('#btn-snap');
    await page.keyboard.press('F3');
    await page.keyboard.press('F3');
    await expect(btn).toHaveClass(/snap-on/);
  });

  test('snap toggle persists across page reload', async ({ page }) => {
    // Turn snap OFF
    await page.locator('#btn-snap').click();
    await expect(page.locator('#btn-snap')).not.toHaveClass(/snap-on/);

    // Verify it was saved to localStorage
    const stored = await page.evaluate(() => localStorage.getItem('snapEnabled'));
    expect(stored).toBe('false');

    // Reload and verify snap is still OFF
    await page.reload();
    await waitForAppReady(page);
    await expect(page.locator('#btn-snap')).not.toHaveClass(/snap-on/);
  });

  test('snap mode checkboxes are visible in settings popover', async ({ page }) => {
    await page.locator('#btn-snap-settings').click();
    const panel = page.locator('#snap-settings');
    await expect(panel).toHaveClass(/open/);

    // All 8 snap modes should be present and checked by default
    const modes = [
      'snp-endpoint', 'snp-midpoint', 'snp-center', 'snp-quadrant',
      'snp-intersection', 'snp-perpendicular', 'snp-tangent', 'snp-nearest',
    ];
    for (const id of modes) {
      await expect(page.locator(`#${id}`)).toBeChecked();
    }
  });

  test('"None" button unchecks all snap modes', async ({ page }) => {
    await page.locator('#btn-snap-settings').click();
    await page.locator('#btn-snap-none').click();

    const modes = [
      'snp-endpoint', 'snp-midpoint', 'snp-center', 'snp-quadrant',
      'snp-intersection', 'snp-perpendicular', 'snp-tangent', 'snp-nearest',
    ];
    for (const id of modes) {
      await expect(page.locator(`#${id}`)).not.toBeChecked();
    }
  });

  test('"All" button rechecks all snap modes after unchecking', async ({ page }) => {
    await page.locator('#btn-snap-settings').click();
    await page.locator('#btn-snap-none').click();
    await page.locator('#btn-snap-all').click();

    const modes = [
      'snp-endpoint', 'snp-midpoint', 'snp-center', 'snp-quadrant',
      'snp-intersection', 'snp-perpendicular', 'snp-tangent', 'snp-nearest',
    ];
    for (const id of modes) {
      await expect(page.locator(`#${id}`)).toBeChecked();
    }
  });

  test('individual snap mode persists after page reload', async ({ page }) => {
    // Open settings and uncheck 'midpoint'
    await page.locator('#btn-snap-settings').click();
    await page.locator('#snp-midpoint').uncheck();
    await expect(page.locator('#snp-midpoint')).not.toBeChecked();

    // Verify localStorage saved the updated mask (midpoint bit 0x02 cleared)
    const mask = await page.evaluate(() => localStorage.getItem('snapMask'));
    expect(mask).not.toBeNull();
    expect(parseInt(mask, 10) & 0x02).toBe(0);

    // Reload and verify the checkbox is still unchecked
    await page.reload();
    await waitForAppReady(page);
    await page.locator('#btn-snap-settings').click();
    await expect(page.locator('#snp-midpoint')).not.toBeChecked();
  });
});
