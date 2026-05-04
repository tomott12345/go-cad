// web/js/state.js — shared mutable application state (imported by all modules)
export const state = {
  wasmReady: false,
  currentTool: 'line',
  clicks: [],         // world-space click points for current tool
  tempPt: null,       // current cursor world-space position
  selectedId: 0,      // selected entity ID (0 = none)
  currentLayer: 0,
  currentColor: '#00ff00',
  pendingText: null,  // captured text for text/mtext/leader
  editPickIds: [],    // IDs being collected for multi-step edit ops
  panX: 0, panY: 0, zoom: 1,
  snapEnabled: true,
  snapResult: null,   // {x, y, kind} or null
  _blockEntCache: {}, // blockName → Entity[]
};

export function setStatus(msg) {
  const el = document.getElementById('status');
  if (el) el.textContent = msg;
}

export function escH(s) {
  return String(s)
    .replace(/&/g, '&amp;').replace(/</g, '&lt;')
    .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

export function invalidateBlockCache(name) {
  if (name) delete state._blockEntCache[name];
  else Object.keys(state._blockEntCache).forEach(k => delete state._blockEntCache[k]);
}

// ── Coordinate transforms (depend only on state.panX/panY/zoom) ───────────────
// Exported here to avoid circular deps between canvas.js and snap.js
export function w2s(wx, wy) {
  return [
    state.panX + wx * state.zoom,
    state.panY - wy * state.zoom,
  ];
}
export function s2w(sx, sy) {
  return [
    (sx - state.panX) / state.zoom,
    (state.panY - sy) / state.zoom,
  ];
}

export function getBlockEntities(name) {
  if (!state.wasmReady || !window.cadGetBlockEntities) return [];
  if (!state._blockEntCache[name]) {
    try {
      state._blockEntCache[name] = JSON.parse(window.cadGetBlockEntities(name) || '[]');
    } catch (_) {
      state._blockEntCache[name] = [];
    }
  }
  return state._blockEntCache[name];
}
