// web/js/snap.js — Object Snap engine
import { state, setStatus, w2s, s2w } from './state.js';

// Snap mode bit definitions (matching Task #5 spec)
export const SNAP_BITS = {
  'snp-endpoint':      0x01,
  'snp-midpoint':      0x02,
  'snp-center':        0x04,
  'snp-quadrant':      0x08,
  'snp-intersection':  0x10,
  'snp-perpendicular': 0x20,
  'snp-tangent':       0x40,
  'snp-nearest':       0x80,
};

// Snap marker colors per mode
export const snapColors = {
  endpoint:      '#ff4444',
  midpoint:      '#44ff44',
  center:        '#4488ff',
  quadrant:      '#ffaa00',
  intersection:  '#ff44ff',
  perpendicular: '#00ffff',
  tangent:       '#ffff44',
  nearest:       '#aaaaaa',
};

export function snapMask() {
  let m = 0;
  Object.entries(SNAP_BITS).forEach(([id, bit]) => {
    const el = document.getElementById(id);
    if (el && el.checked) m |= bit;
  });
  return m;
}

// Main snap function: returns world-space snapped point {x, y, kind}
// or null if no snap within threshold
export function snapWorldPt(sx, sy) {
  if (!state.snapEnabled || !state.wasmReady) return null;
  const [wx, wy] = s2w(sx, sy);
  const mask = snapMask();
  const THRESH_PX = 14;
  const threshWorld = THRESH_PX / state.zoom;

  try {
    if (!window.cadFindSnap) return null;
    const raw = window.cadFindSnap(wx, wy, threshWorld, mask);
    if (!raw) return null;
    const result = JSON.parse(raw);
    if (!result || typeof result.x !== 'number') return null;
    // Normalise .type → .kind (lower-case) for consistent use in drawSnapMarker
    return { x: result.x, y: result.y, kind: (result.type || 'nearest').toLowerCase(), entityID: result.entityID };
  } catch (_) {
    return null;
  }
}

export function drawSnapMarker() {
  const marker = document.getElementById('snap-marker');
  if (!marker) return;
  if (!state.snapResult || !state.snapEnabled) {
    marker.style.display = 'none';
    return;
  }
  const [sx, sy] = w2s(state.snapResult.x, state.snapResult.y);
  const color = snapColors[state.snapResult.kind] || '#ffffff';
  const svg   = document.getElementById('snap-svg');
  if (!svg) return;

  marker.style.display = 'block';
  marker.style.left = (sx - 10) + 'px';
  marker.style.top  = (sy - 10) + 'px';

  const kind = state.snapResult.kind;
  let shape = '';
  if (kind === 'endpoint')     shape = `<rect x="-6" y="-6" width="12" height="12" fill="none" stroke="${color}" stroke-width="1.5"/>`;
  else if (kind === 'midpoint') shape = `<path d="M-6 6 L0 -6 L6 6 Z" fill="none" stroke="${color}" stroke-width="1.5"/>`;
  else if (kind === 'center')   shape = `<circle r="6" fill="none" stroke="${color}" stroke-width="1.5"/><circle r="2" fill="${color}"/>`;
  else if (kind === 'quadrant') shape = `<path d="M0 -7 L7 0 L0 7 L-7 0 Z" fill="none" stroke="${color}" stroke-width="1.5"/>`;
  else if (kind === 'intersection') shape = `<line x1="-7" y1="-7" x2="7" y2="7" stroke="${color}" stroke-width="1.5"/><line x1="7" y1="-7" x2="-7" y2="7" stroke="${color}" stroke-width="1.5"/>`;
  else if (kind === 'perpendicular') shape = `<rect x="-6" y="-6" width="12" height="12" fill="none" stroke="${color}" stroke-width="1.5"/><line x1="-4" y1="0" x2="4" y2="0" stroke="${color}" stroke-width="1.5"/>`;
  else if (kind === 'tangent')  shape = `<circle r="6" fill="none" stroke="${color}" stroke-width="1.5"/><line x1="0" y1="-6" x2="0" y2="6" stroke="${color}" stroke-width="1"/>`;
  else                           shape = `<circle r="4" fill="none" stroke="${color}" stroke-width="1.5"/>`;

  svg.innerHTML = shape;
}

export function updateSnapBtn() {
  const btn = document.getElementById('btn-snap');
  if (!btn) return;
  if (state.snapEnabled) {
    btn.classList.add('snap-on');
    btn.textContent = 'Snap ✓';
  } else {
    btn.classList.remove('snap-on');
    btn.textContent = 'Snap';
  }
  // Also update the individual snap-bar buttons
  updateSnapBarButtons();
}

function updateSnapBarButtons() {
  Object.entries(SNAP_BITS).forEach(([id, bit]) => {
    const cb = document.getElementById(id);
    const barBtn = document.getElementById('snapbtn-' + id.replace('snp-', ''));
    if (!barBtn) return;
    const on = cb && cb.checked && state.snapEnabled;
    barBtn.classList.toggle('snap-bar-on', !!on);
  });
}

export function setAllSnapModes(v) {
  Object.keys(SNAP_BITS).forEach(id => {
    const el = document.getElementById(id);
    if (el) el.checked = v;
  });
  saveSnapSettings();
  updateSnapBarButtons();
}

export function closeSnapSettings() {
  document.getElementById('snap-settings')?.classList.remove('open');
}

export function loadSnapSettings() {
  const mask = parseInt(localStorage.getItem('snapMask') ?? '0xFF');
  const en   = localStorage.getItem('snapEnabled') !== 'false';
  state.snapEnabled = en;
  Object.entries(SNAP_BITS).forEach(([id, bit]) => {
    const el = document.getElementById(id);
    if (el) el.checked = !!(mask & bit);
  });
  updateSnapBtn();
}

export function saveSnapSettings() {
  localStorage.setItem('snapMask', String(snapMask()));
  localStorage.setItem('snapEnabled', String(state.snapEnabled));
}

// Build the snap toolbar buttons row (called from app.js after DOM ready)
export function initSnapToolbar() {
  const bar = document.getElementById('snap-btn-bar');
  if (!bar) return;

  const modes = [
    { id: 'endpoint',     label: 'End',  title: 'Endpoint snap' },
    { id: 'midpoint',     label: 'Mid',  title: 'Midpoint snap' },
    { id: 'center',       label: 'Cen',  title: 'Center snap' },
    { id: 'quadrant',     label: 'Quad', title: 'Quadrant snap' },
    { id: 'intersection', label: 'Int',  title: 'Intersection snap' },
    { id: 'perpendicular',label: 'Per',  title: 'Perpendicular snap' },
    { id: 'tangent',      label: 'Tan',  title: 'Tangent snap' },
    { id: 'nearest',      label: 'Nea',  title: 'Nearest snap' },
  ];

  bar.innerHTML = modes.map(m => {
    const color = snapColors[m.id] || '#888';
    return `<button id="snapbtn-${m.id}" class="snap-bar-btn" title="${m.title}"
      style="--snap-color:${color}">${m.label}</button>`;
  }).join('');

  bar.querySelectorAll('.snap-bar-btn').forEach(btn => {
    const modeId = btn.id.replace('snapbtn-', '');
    const cbId   = 'snp-' + modeId;
    btn.addEventListener('click', () => {
      const cb = document.getElementById(cbId);
      if (cb) {
        cb.checked = !cb.checked;
        saveSnapSettings();
      }
      btn.classList.toggle('snap-bar-on', !!(document.getElementById(cbId)?.checked));
    });
  });

  // Toggle all on global snap button
  document.getElementById('btn-snap')?.addEventListener('click', () => {
    state.snapEnabled = !state.snapEnabled;
    state.snapResult  = null;
    updateSnapBtn();
    saveSnapSettings();
    setStatus(state.snapEnabled ? 'Object Snap ON (F3 to toggle)' : 'Object Snap OFF (F3 to toggle)');
  });

  // Settings popover
  document.getElementById('btn-snap-settings')?.addEventListener('click', e => {
    e.stopPropagation();
    document.getElementById('snap-settings')?.classList.toggle('open');
  });
  document.addEventListener('click', e => {
    const panel = document.getElementById('snap-settings');
    if (panel?.classList.contains('open') &&
        !panel.contains(e.target) &&
        e.target.id !== 'btn-snap-settings') {
      closeSnapSettings();
    }
  });

  // Auto-save on checkbox change
  Object.keys(SNAP_BITS).forEach(id => {
    document.getElementById(id)?.addEventListener('change', () => {
      saveSnapSettings();
      updateSnapBarButtons();
    });
  });

  // All / None buttons in popover footer
  document.getElementById('btn-snap-all')?.addEventListener('click',  () => setAllSnapModes(true));
  document.getElementById('btn-snap-none')?.addEventListener('click', () => setAllSnapModes(false));
  document.getElementById('btn-snap-close')?.addEventListener('click', closeSnapSettings);
}
