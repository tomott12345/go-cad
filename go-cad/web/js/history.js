// web/js/history.js — command history panel
import { escH } from './state.js';

const entries = []; // { cmd, result, time }

// Set by app.js after both history and commands are initialised
// to avoid a circular import (commands → history → commands).
let _replayCb = null;
export function setReplayCallback(fn) { _replayCb = fn; }

export function addHistoryEntry(cmd, result = '') {
  const time = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  entries.push({ cmd: String(cmd), result: String(result || ''), time });
  if (entries.length > 300) entries.shift();
  renderHistory();
}

export function updateLastResult(result) {
  if (entries.length === 0) return;
  entries[entries.length - 1].result = String(result || '');
  renderHistory();
}

function renderHistory() {
  const el = document.getElementById('history-list');
  if (!el) return;
  el.innerHTML = entries.slice(-80).reverse().map(e => `
    <div class="hist-entry" data-cmd="${escH(e.cmd)}" title="${escH(e.result)}">
      <span class="hist-time">${e.time}</span>
      <span class="hist-cmd">${escH(e.cmd)}</span>
      ${e.result ? `<span class="hist-result">${escH(e.result)}</span>` : ''}
    </div>`).join('');
  el.querySelectorAll('.hist-entry').forEach(row => {
    row.addEventListener('click', () => {
      const cmd = row.dataset.cmd;
      // Re-execute the command via the registered callback
      if (_replayCb) {
        _replayCb(cmd);
      } else {
        // Fallback: just fill the input
        const inp = document.getElementById('cmd-input');
        if (inp) { inp.value = cmd; inp.focus(); }
      }
    });
  });
}

export function initHistory() {
  renderHistory();
}
