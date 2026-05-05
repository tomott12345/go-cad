/**
 * Shared test helpers for go-cad Playwright tests.
 */

/**
 * Wait for the app to finish initialising (WASM or demo stubs).
 * Resolves once the status bar no longer says "Initialising…".
 *
 * @param {import('@playwright/test').Page} page
 */
async function waitForAppReady(page) {
  await page.waitForFunction(() => {
    const status = document.getElementById('status');
    if (!status) return false;
    const text = status.textContent || '';
    return !text.includes('Initialising') && text.length > 0;
  }, { timeout: 15_000 });
}

/**
 * Clear go-cad localStorage keys so each test starts from a clean slate.
 *
 * @param {import('@playwright/test').Page} page
 */
async function clearAppStorage(page) {
  await page.evaluate(() => {
    const keys = [
      'go-cad-welcome-shown',
      'go-cad-recent-files',
      'snapMask',
      'snapEnabled',
      'grid-x',
      'grid-y',
      'grid-on',
      'cad-units',
      'cad-precision',
      'cad-def-linetype',
      'cad-def-lw',
    ];
    keys.forEach(k => localStorage.removeItem(k));
  });
}

module.exports = { waitForAppReady, clearAppStorage };
