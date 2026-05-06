// web/tests/layers.spec.js
// Tests: Layer Manager dialog — open, close, add layer, set color, canvas update.

const { test, expect } = require('@playwright/test');
const { waitForAppReady, clearAppStorage } = require('./helpers');

test.describe('Layer Manager', () => {
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

  test('Layer Manager opens via toolbar button', async ({ page }) => {
    await page.locator('#btn-layers').click();
    await expect(page.locator('#layer-modal')).toHaveClass(/open/);
  });

  test('Layer Manager shows at least one layer (layer 0)', async ({ page }) => {
    await page.locator('#btn-layers').click();
    const rows = page.locator('#layer-tbody tr');
    await expect(rows).toHaveCount(1);
    const nameInput = rows.first().locator('.layer-name-inp');
    await expect(nameInput).toHaveValue('0');
  });

  test('Layer Manager closes via Cancel button', async ({ page }) => {
    await page.locator('#btn-layers').click();
    await expect(page.locator('#layer-modal')).toHaveClass(/open/);
    await page.locator('#btn-close-layers').click();
    await expect(page.locator('#layer-modal')).not.toHaveClass(/open/);
  });

  test('Layer Manager closes via Escape key', async ({ page }) => {
    await page.locator('#btn-layers').click();
    await expect(page.locator('#layer-modal')).toHaveClass(/open/);
    await page.keyboard.press('Escape');
    await expect(page.locator('#layer-modal')).not.toHaveClass(/open/);
  });

  test('layer count shown in Drawing Info panel', async ({ page }) => {
    const layerCount = page.locator('#layer-count');
    const count = await layerCount.textContent();
    expect(parseInt(count || '0', 10)).toBeGreaterThanOrEqual(1);
  });

  test('New Layer button adds a row to the table and the toolbar selector', async ({ page }) => {
    // Patch cadAddLayer + cadGetLayers to be stateful so the table actually grows.
    // (The default demo stub always returns a fixed single-layer list.)
    await page.evaluate(() => {
      const layers = JSON.parse(window.cadGetLayers());
      let nextId = layers.reduce((m, l) => Math.max(m, l.id), 0) + 1;
      window.cadGetLayers = () => JSON.stringify(layers);
      window.cadAddLayer = (name, color) => {
        const id = nextId++;
        layers.push({
          id, name, color: color || '#ffffff',
          lineType: 'Solid', lineWeight: 0.25,
          visible: true, locked: false, frozen: false, printEnabled: true,
        });
        return id;
      };
    });

    await page.locator('#btn-layers').click();
    const rowsBefore = await page.locator('#layer-tbody tr').count();
    const optsBefore = await page.locator('#layer-sel option').count();

    // btn-add-layer calls prompt() for the new layer name
    page.once('dialog', dialog => dialog.accept('TestLayer'));
    await page.locator('#btn-add-layer').click();

    // Row count must increase by exactly 1
    await expect(page.locator('#layer-tbody tr')).toHaveCount(rowsBefore + 1, { timeout: 5_000 });

    // The new layer name must appear in the table
    const names = await page.locator('#layer-tbody .layer-name-inp').allInputValues();
    expect(names).toContain('TestLayer');

    // The toolbar layer selector must also list the new layer
    const optsAfter = await page.locator('#layer-sel option').count();
    expect(optsAfter).toBeGreaterThan(optsBefore);
    const selectorText = await page.locator('#layer-sel').innerHTML();
    expect(selectorText).toContain('TestLayer');
  });

  test('layer selector in toolbar lists available layers', async ({ page }) => {
    const sel = page.locator('#layer-sel');
    await expect(sel).toBeVisible();
    const options = await sel.locator('option').count();
    expect(options).toBeGreaterThanOrEqual(1);
  });

  test('setting layer color calls cadSetLayerColor and triggers canvas re-render', async ({ page }) => {
    // Intercept clearRect as the render() sentinel — render() calls clearRect first.
    await page.evaluate(() => {
      window.__clearRectCallsBeforeColor = 0;
      window.__clearRectCallsAfterColor  = 0;
      window.__trackingPhase = 'before';
      const orig = CanvasRenderingContext2D.prototype.clearRect;
      CanvasRenderingContext2D.prototype.clearRect = function (...args) {
        if (window.__trackingPhase === 'after') window.__clearRectCallsAfterColor++;
        else window.__clearRectCallsBeforeColor++;
        return orig.apply(this, args);
      };
    });

    // Spy on cadSetLayerColor
    await page.evaluate(() => {
      window.__layerColorCalls = [];
      const orig = window.cadSetLayerColor;
      window.cadSetLayerColor = (...args) => {
        window.__layerColorCalls.push([...args]);
        return orig ? orig(...args) : undefined;
      };
    });

    // Switch tracking phase to 'after' and open Layer Manager
    await page.locator('#btn-layers').click();
    await expect(page.locator('#layer-modal')).toHaveClass(/open/);

    await page.evaluate(() => { window.__trackingPhase = 'after'; });

    // Change the color input on layer 0 and fire the 'input' event
    // (dialogs.js wires cadSetLayerColor + render() on 'input' events of .layer-color-inp)
    const newColor = '#ff3300';
    const colorInput = page.locator('#layer-tbody .layer-color-inp').first();
    await colorInput.evaluate((el, color) => {
      el.value = color;
      el.dispatchEvent(new Event('input', { bubbles: true }));
    }, newColor);

    // cadSetLayerColor must have been called with the correct color
    const calls = await page.evaluate(() => window.__layerColorCalls || []);
    expect(calls.length).toBeGreaterThan(0);
    expect(calls[0][1]).toBe(newColor);

    // render() must have been triggered — clearRect should have fired at least once
    const clearRectAfter = await page.evaluate(() => window.__clearRectCallsAfterColor);
    expect(clearRectAfter).toBeGreaterThan(0);

    // Close the dialog
    await page.locator('#btn-close-layers').click();
    await expect(page.locator('#layer-modal')).not.toHaveClass(/open/);
  });

  test('drawing an entity on a layer records the current layer id on the entity', async ({ page }) => {
    await page.locator('#tool-line').click();
    const canvas = page.locator('#canvas');
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();

    await page.mouse.click(box.x + box.width * 0.3, box.y + box.height * 0.5);
    await page.mouse.click(box.x + box.width * 0.7, box.y + box.height * 0.5);
    await page.keyboard.press('Escape');

    const entities = await page.evaluate(() => {
      const raw = window.cadEntities ? window.cadEntities() : '[]';
      return JSON.parse(raw || '[]');
    });

    expect(entities.length).toBeGreaterThan(0);
    const line = entities.find(e => e.type === 'line');
    expect(line).toBeTruthy();
    expect(typeof line.layer).toBe('number');
  });
});
