// web/tests/print.spec.js
// Tests: Print/Plot dialog — open, format selection, PNG and SVG export.

const { test, expect } = require('@playwright/test');
const { waitForAppReady, clearAppStorage } = require('./helpers');

test.describe('Print / Plot dialog', () => {
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

  test('Print dialog opens via toolbar button', async ({ page }) => {
    await page.locator('#btn-print').click();
    await expect(page.locator('#print-modal')).toHaveClass(/open/);
  });

  test('Print dialog closes via Cancel button', async ({ page }) => {
    await page.locator('#btn-print').click();
    await expect(page.locator('#print-modal')).toHaveClass(/open/);
    await page.locator('#btn-close-print').click();
    await expect(page.locator('#print-modal')).not.toHaveClass(/open/);
  });

  test('Print dialog closes via Escape key', async ({ page }) => {
    await page.locator('#btn-print').click();
    await page.keyboard.press('Escape');
    await expect(page.locator('#print-modal')).not.toHaveClass(/open/);
  });

  test('DPI row is hidden for PDF format', async ({ page }) => {
    await page.locator('#btn-print').click();
    await page.locator('#print-fmt').selectOption('pdf');
    await expect(page.locator('#print-dpi-row')).toBeHidden();
  });

  test('DPI row is visible for PNG format', async ({ page }) => {
    await page.locator('#btn-print').click();
    await page.locator('#print-fmt').selectOption('png');
    await expect(page.locator('#print-dpi-row')).toBeVisible();
  });

  test('DPI row is hidden for SVG format', async ({ page }) => {
    await page.locator('#btn-print').click();
    await page.locator('#print-fmt').selectOption('svg');
    await expect(page.locator('#print-dpi-row')).toBeHidden();
  });

  test('PNG export triggers a download with non-empty data', async ({ page }) => {
    // Draw a line so there is something on the canvas
    await page.locator('#tool-line').click();
    const canvas = page.locator('#canvas');
    const box = await canvas.boundingBox();
    await page.mouse.click(box.x + box.width * 0.3, box.y + box.height * 0.4);
    await page.mouse.click(box.x + box.width * 0.7, box.y + box.height * 0.6);
    await page.keyboard.press('Escape');

    // Open print dialog and choose PNG
    await page.locator('#btn-print').click();
    await page.locator('#print-fmt').selectOption('png');
    await page.locator('#print-area').selectOption('view');

    // Intercept the download
    const [download] = await Promise.all([
      page.waitForEvent('download', { timeout: 10_000 }),
      page.locator('#btn-execute-print').click(),
    ]);

    expect(download.suggestedFilename()).toMatch(/drawing\.png/i);

    const stream = await download.createReadStream();
    const chunks = [];
    for await (const chunk of stream) chunks.push(chunk);
    const size = chunks.reduce((s, c) => s + c.length, 0);
    expect(size).toBeGreaterThan(0);
  });

  test('SVG export triggers a download with non-empty SVG content', async ({ page }) => {
    // Draw something so the SVG export has content
    await page.locator('#tool-circle').click();
    const canvas = page.locator('#canvas');
    const box = await canvas.boundingBox();
    await page.mouse.click(box.x + box.width * 0.5, box.y + box.height * 0.5);
    await page.mouse.click(box.x + box.width * 0.5 + 60, box.y + box.height * 0.5);
    await page.keyboard.press('Escape');

    // Open print dialog and choose SVG
    await page.locator('#btn-print').click();
    await page.locator('#print-fmt').selectOption('svg');

    const [download] = await Promise.all([
      page.waitForEvent('download', { timeout: 10_000 }),
      page.locator('#btn-execute-print').click(),
    ]);

    expect(download.suggestedFilename()).toMatch(/drawing\.svg/i);

    const stream = await download.createReadStream();
    const chunks = [];
    for await (const chunk of stream) chunks.push(chunk);
    const content = Buffer.concat(chunks).toString('utf8');
    expect(content.length).toBeGreaterThan(0);
    expect(content).toContain('<svg');
  });

  test('status bar shows export confirmation after SVG export', async ({ page }) => {
    await page.locator('#btn-print').click();
    await page.locator('#print-fmt').selectOption('svg');

    await Promise.all([
      page.waitForEvent('download', { timeout: 10_000 }),
      page.locator('#btn-execute-print').click(),
    ]);

    const status = page.locator('#status');
    await expect(status).toContainText('svg', { ignoreCase: true, timeout: 5_000 });
  });
});
