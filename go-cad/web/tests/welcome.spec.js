// web/tests/welcome.spec.js
// Tests: Welcome overlay visibility and dismissal flows.

const { test, expect } = require('@playwright/test');
const { waitForAppReady, clearAppStorage } = require('./helpers');

test.describe('Welcome overlay', () => {
  test.beforeEach(async ({ page }) => {
    // Load app, clear storage so welcome always appears, then reload.
    await page.goto('/');
    await clearAppStorage(page);
    await page.reload();
    await waitForAppReady(page);
  });

  test('overlay is visible on first load', async ({ page }) => {
    const overlay = page.locator('#welcome-overlay');
    await expect(overlay).toBeVisible();
    await expect(overlay.locator('#welcome-title')).toHaveText('go-cad');
  });

  test('clicking "Skip" hides the overlay', async ({ page }) => {
    await page.locator('#btn-welcome-close').click();
    await expect(page.locator('#welcome-overlay')).toBeHidden();
  });

  test('clicking "New Drawing" hides the overlay', async ({ page }) => {
    await page.locator('#btn-welcome-new').click();
    await expect(page.locator('#welcome-overlay')).toBeHidden();
  });

  test('pressing Escape hides the overlay', async ({ page }) => {
    await page.keyboard.press('Escape');
    await expect(page.locator('#welcome-overlay')).toBeHidden();
  });

  test('clicking the backdrop hides the overlay', async ({ page }) => {
    const overlay = page.locator('#welcome-overlay');
    const box = await overlay.boundingBox();
    if (box) {
      // Click top-left corner of the backdrop (outside the welcome box)
      await page.mouse.click(box.x + 5, box.y + 5);
    }
    await expect(overlay).toBeHidden();
  });

  test('overlay stays hidden on subsequent loads after dismissal', async ({ page }) => {
    await page.locator('#btn-welcome-close').click();
    await expect(page.locator('#welcome-overlay')).toBeHidden();

    await page.reload();
    await waitForAppReady(page);
    await expect(page.locator('#welcome-overlay')).toBeHidden();
  });
});
