// web/js/welcome.js — startup welcome screen
import { state } from './state.js';

const RECENT_KEY = 'go-cad-recent-files';
const MAX_RECENT = 5;

export function addRecentFile(name) {
  const list = getRecentFiles().filter(f => f !== name);
  list.unshift(name);
  localStorage.setItem(RECENT_KEY, JSON.stringify(list.slice(0, MAX_RECENT)));
}

export function getRecentFiles() {
  try { return JSON.parse(localStorage.getItem(RECENT_KEY) || '[]'); } catch (_) { return []; }
}

function renderRecentFiles(container) {
  const files = getRecentFiles();
  if (!files.length) {
    container.innerHTML = '<div style="color:#555;font-size:11px;padding:4px 0">(no recent files)</div>';
    return;
  }
  container.innerHTML = files.map(f =>
    `<div style="padding:2px 0;font-size:11px;color:#9ec8e8;cursor:pointer;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="${f}" class="recent-file-item">${f}</div>`
  ).join('');
  container.querySelectorAll('.recent-file-item').forEach((el, i) => {
    el.addEventListener('click', () => {
      hideWelcome();
      document.getElementById('file-input')?.click();
    });
  });
}

export function initWelcome() {
  const overlay = document.getElementById('welcome-overlay');
  if (!overlay) return;

  const shown = localStorage.getItem('go-cad-welcome-shown');
  if (shown) {
    overlay.style.display = 'none';
    return;
  }
  overlay.style.display = 'flex';

  // Populate recent files
  const recentEl = overlay.querySelector('#welcome-recent-list');
  if (recentEl) renderRecentFiles(recentEl);

  overlay.querySelector('#btn-welcome-new')?.addEventListener('click', () => {
    hideWelcome();
    if (state.wasmReady && window.cadClear) window.cadClear();
  });
  overlay.querySelector('#btn-welcome-open')?.addEventListener('click', () => {
    hideWelcome();
    document.getElementById('file-input')?.click();
  });
  overlay.querySelector('#btn-welcome-close')?.addEventListener('click', hideWelcome);
  overlay.addEventListener('click', e => { if (e.target === overlay) hideWelcome(); });

  // Dismiss on Escape key
  const onKey = e => { if (e.key === 'Escape') { hideWelcome(); document.removeEventListener('keydown', onKey, true); } };
  document.addEventListener('keydown', onKey, true);
}

export function hideWelcome() {
  const overlay = document.getElementById('welcome-overlay');
  if (overlay) overlay.style.display = 'none';
  localStorage.setItem('go-cad-welcome-shown', '1');
}
