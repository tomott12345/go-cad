// web/js/history.js — command history panel
import { escH } from './state.js';

const entries = []; // { cmd, result, time }

export function addHistoryEntry(cmd, result = '') {
  const time = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  entries.push({ cmd: String(cmd), result: String(result || ''), time });
  if (entries.length > 300) entries.shift();
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
      const inp = document.getElementById('cmd-input');
      if (inp) { inp.value = row.dataset.cmd; inp.focus(); }
    });
  });
}

export function initHistory() {
  renderHistory();
}
