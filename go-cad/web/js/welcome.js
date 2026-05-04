// web/js/welcome.js — startup welcome screen
import { state } from './state.js';

export function initWelcome() {
  const overlay = document.getElementById('welcome-overlay');
  if (!overlay) return;

  const shown = localStorage.getItem('go-cad-welcome-shown');
  if (shown) {
    overlay.style.display = 'none';
    return;
  }
  overlay.style.display = 'flex';

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
