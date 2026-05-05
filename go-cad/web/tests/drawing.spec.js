// web/tests/drawing.spec.js
// Tests: Tool switching, drawing a line, confirming it appears in the entity list.

const { test, expect } = require('@playwright/test');
const { waitForAppReady, clearAppStorage } = require('./helpers');

test.describe('Drawing tools and entity list', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await clearAppStorage(page);
    await page.reload();
    await waitForAppReady(page);
    // Dismiss welcome overlay so it doesn't block interactions
    const overlay = page.locator('#welcome-overlay');
    if (await overlay.isVisible()) {
      await page.locator('#btn-welcome-close').click();
    }
  });

  test('Select tool button is active by default', async ({ page }) => {
    const selectBtn = page.locator('#tool-select');
    await expect(selectBtn).toHaveClass(/active/);
  });

  test('clicking Line button activates Line tool', async ({ page }) => {
    await page.locator('#tool-line').click();
    await expect(page.locator('#tool-line')).toHaveClass(/active/);
    await expect(page.locator('#tool-select')).not.toHaveClass(/active/);
  });

  test('drawing a line adds it to the entity list returned by cadEntities()', async ({ page }) => {
    // Record baseline entity list length
    const before = await page.evaluate(() => {
      const raw = window.cadEntities ? window.cadEntities() : '[]';
      return JSON.parse(raw || '[]').length;
    });

    // Activate line tool and draw
    await page.locator('#tool-line').click();
    const canvas = page.locator('#canvas');
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();

    await page.mouse.click(box.x + box.width * 0.3, box.y + box.height * 0.5);
    await page.mouse.click(box.x + box.width * 0.7, box.y + box.height * 0.5);
    await page.keyboard.press('Escape');

    // Verify the entity list now has more entries than before
    const entities = await page.evaluate(() => {
      const raw = window.cadEntities ? window.cadEntities() : '[]';
      return JSON.parse(raw || '[]');
    });
    expect(entities.length).toBeGreaterThan(before);

    // Verify the new entry is a 'line' with numeric coordinates
    const lineEntity = entities.find(e => e.type === 'line');
    expect(lineEntity).toBeTruthy();
    expect(typeof lineEntity.x1).toBe('number');
    expect(typeof lineEntity.y1).toBe('number');
    expect(typeof lineEntity.x2).toBe('number');
    expect(typeof lineEntity.y2).toBe('number');
  });

  test('drawn line appears in the inspector when selected on canvas', async ({ page }) => {
    // Draw a horizontal line across the canvas centre
    await page.locator('#tool-line').click();
    const canvas = page.locator('#canvas');
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();

    await page.mouse.click(box.x + box.width * 0.3, box.y + box.height * 0.5);
    await page.mouse.click(box.x + box.width * 0.7, box.y + box.height * 0.5);
    await page.keyboard.press('Escape');

    // Ensure select tool is active after Escape
    await expect(page.locator('#tool-select')).toHaveClass(/active/);

    // Click near the midpoint of the drawn line to select it
    await page.mouse.click(box.x + box.width * 0.5, box.y + box.height * 0.5);

    // Inspector must show the entity type row with 'LINE'
    const typeLabel = page.locator('#inspector-content .insp-type');
    await expect(typeLabel).toBeVisible({ timeout: 5_000 });
    await expect(typeLabel).toHaveText(/line/i);
  });

  test('drawing a line increments the entity count sidebar', async ({ page }) => {
    const entCount = page.locator('#ent-count');
    const before = parseInt((await entCount.textContent()) || '0', 10);

    await page.locator('#tool-line').click();
    const canvas = page.locator('#canvas');
    const box = await canvas.boundingBox();

    await page.mouse.click(box.x + box.width * 0.3, box.y + box.height * 0.5);
    await page.mouse.click(box.x + box.width * 0.7, box.y + box.height * 0.5);
    await page.keyboard.press('Escape');

    // Entity count sidebar refreshes via setInterval every 2 s — allow up to 5 s
    await expect(entCount).not.toHaveText(String(before), { timeout: 5_000 });
    const after = parseInt((await entCount.textContent()) || '0', 10);
    expect(after).toBeGreaterThan(before);
  });

  test('drawing a circle increments entity count', async ({ page }) => {
    const entCount = page.locator('#ent-count');
    const before = parseInt((await entCount.textContent()) || '0', 10);

    await page.locator('#tool-circle').click();
    const canvas = page.locator('#canvas');
    const box = await canvas.boundingBox();

    await page.mouse.click(box.x + box.width * 0.5, box.y + box.height * 0.5);
    await page.mouse.click(box.x + box.width * 0.5 + 50, box.y + box.height * 0.5);
    await page.keyboard.press('Escape');

    await expect(entCount).not.toHaveText(String(before), { timeout: 5_000 });
    const after = parseInt((await entCount.textContent()) || '0', 10);
    expect(after).toBeGreaterThan(before);
  });

  test('Clear button resets entity list to empty', async ({ page }) => {
    // Draw a line first
    await page.locator('#tool-line').click();
    const canvas = page.locator('#canvas');
    const box = await canvas.boundingBox();
    await page.mouse.click(box.x + box.width * 0.3, box.y + box.height * 0.5);
    await page.mouse.click(box.x + box.width * 0.7, box.y + box.height * 0.5);
    await page.keyboard.press('Escape');

    // Wait for entity to appear in cadEntities()
    await page.waitForFunction(() => {
      const raw = window.cadEntities ? window.cadEntities() : '[]';
      return JSON.parse(raw || '[]').length > 0;
    }, { timeout: 5_000 });

    await page.locator('#btn-clear').click();

    // cadEntities() must return an empty list immediately after Clear
    const entities = await page.evaluate(() => {
      const raw = window.cadEntities ? window.cadEntities() : '[]';
      return JSON.parse(raw || '[]');
    });
    expect(entities.length).toBe(0);

    // And the sidebar count should also reflect 0 within 5 s
    await expect(page.locator('#ent-count')).toHaveText('0', { timeout: 5_000 });
  });

  test('status bar shows ready message after init', async ({ page }) => {
    const status = page.locator('#status');
    const text = await status.textContent();
    expect(text).toMatch(/ready|demo mode/i);
  });
});
