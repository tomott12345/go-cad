// web/js/tools.js — Tool management, mouse event handlers, entity commit
import { state, setStatus, escH, invalidateBlockCache } from './state.js';
import { render, w2s, s2w, zoomFit, entitySamplePoints } from './canvas.js';
import { snapWorldPt, drawSnapMarker } from './snap.js';
import { showEntityProperties, clearInspector } from './inspector.js';

// ── Tool hints ─────────────────────────────────────────────────────────────────
const toolHints = {
  line:       'Line: click start point',
  circle:     'Circle: click centre',
  arc:        'Arc: click centre',
  rect:       'Rectangle: click first corner',
  poly:       'Polyline: click points, Enter to finish',
  spline:     'Spline: click ≥4 control pts, Enter to finish',
  nurbs:      'NURBS: click ≥4 control pts, Enter to finish',
  ellipse:    'Ellipse: click centre',
  text:       'Text: click placement point',
  mtext:      'MText: click placement point',
  dimlin:     'Linear Dim: click first point',
  dimali:     'Aligned Dim: click first point',
  dimang:     'Angular Dim: click vertex',
  dimrad:     'Radial Dim: click centre',
  dimdia:     'Diameter Dim: click centre',
  move:       'Move: click entity to move',
  copy:       'Copy: click entity to copy',
  rotate:     'Rotate: click entity to rotate',
  scale:      'Scale: click entity to scale',
  mirror:     'Mirror: click first point on mirror axis',
  trim:       'Trim: click the cutting entity',
  extend:     'Extend: click the boundary entity',
  fillet:     'Fillet: click first line',
  chamfer:    'Chamfer: click first line',
  arrayrect:  'ArrayRect: click entity',
  arraypolar: 'ArrayPolar: click entity',
  offset:     'Offset: click entity to offset',
  hatch:      'Hatch: click polygon vertices, Enter to finish',
  leader:     'Leader: click arrowhead point, then jog points, Enter to finish',
  revcloud:   'RevCloud: click polygon vertices, Enter to finish',
  wipeout:    'Wipeout: click polygon vertices, Enter to finish',
};

// Tool click counts (how many clicks before commit)
const toolClicks = {
  line: 2, circle: 2, arc: 3, rect: 2, ellipse: 3,
  dimlin: 2, dimali: 2, dimang: 3, dimrad: 2, dimdia: 2,
  text: 1, mtext: 1,
};

export function setTool(name) {
  state.currentTool = name;
  state.clicks      = [];
  state.tempPt      = null;
  state.editPickIds = [];

  // Update toolbar active state
  document.querySelectorAll('#main-toolbar button[id^="tool-"]').forEach(b => b.classList.remove('active'));
  const btn = document.getElementById('tool-' + name);
  if (btn) btn.classList.add('active');

  // Hatch options bar
  const hb = document.getElementById('hatch-bar');
  if (hb) hb.classList.toggle('open', name === 'hatch');

  const hint = toolHints[name] || 'Click to draw';
  setStatus(hint);
  render();
}

export function tryDelete(id) {
  if (!id || !state.wasmReady) return false;
  const ok = window.cadDeleteEntity(id);
  if (ok) invalidateBlockCache();
  return ok;
}

// ── Entity commit ──────────────────────────────────────────────────────────────
export function commitEntity() {
  if (!state.wasmReady) return;
  const clicks = state.clicks;
  const tool   = state.currentTool;

  let id = -1;
  const lay = state.currentLayer;
  const col = state.currentColor;

  switch (tool) {
    case 'line':
      if (clicks.length >= 2)
        id = window.cadAddLine(clicks[0][0],clicks[0][1],clicks[1][0],clicks[1][1],lay,col);
      break;
    case 'circle':
      if (clicks.length >= 2)
        id = window.cadAddCircle(clicks[0][0],clicks[0][1],Math.hypot(clicks[1][0]-clicks[0][0],clicks[1][1]-clicks[0][1]),lay,col);
      break;
    case 'arc':
      if (clicks.length >= 3) {
        const r    = Math.hypot(clicks[1][0]-clicks[0][0],clicks[1][1]-clicks[0][1]);
        const sAng = Math.atan2(clicks[1][1]-clicks[0][1],clicks[1][0]-clicks[0][0])*180/Math.PI;
        const eAng = Math.atan2(clicks[2][1]-clicks[0][1],clicks[2][0]-clicks[0][0])*180/Math.PI;
        id = window.cadAddArc(clicks[0][0],clicks[0][1],r,sAng,eAng,lay,col);
      }
      break;
    case 'rect':
      if (clicks.length >= 2)
        id = window.cadAddRectangle(clicks[0][0],clicks[0][1],clicks[1][0],clicks[1][1],lay,col);
      break;
    case 'poly':
      if (clicks.length >= 2)
        id = window.cadAddPolyline(clicks,lay,col);
      break;
    case 'spline':
      if (clicks.length >= 4)
        id = window.cadAddSpline(clicks,lay,col);
      else setStatus('Spline needs ≥4 points');
      break;
    case 'nurbs': {
      if (clicks.length < 4) { setStatus('NURBS needs ≥4 points'); break; }
      const deg = Math.min(3, clicks.length - 1);
      id = window.cadAddNURBS(deg,clicks,[],[],lay,col);
      break;
    }
    case 'ellipse':
      if (clicks.length >= 3) {
        const a   = Math.hypot(clicks[1][0]-clicks[0][0],clicks[1][1]-clicks[0][1]);
        const b   = Math.hypot(clicks[2][0]-clicks[0][0],clicks[2][1]-clicks[0][1]);
        const rot = Math.atan2(clicks[1][1]-clicks[0][1],clicks[1][0]-clicks[0][0])*180/Math.PI;
        id = window.cadAddEllipse(clicks[0][0],clicks[0][1],a,b,rot,lay,col);
      }
      break;
    case 'text': case 'mtext': {
      if (clicks.length < 1) break;
      const label = state.pendingText || prompt('Enter text:','');
      if (!label) break;
      if (tool === 'text')
        id = window.cadAddText(clicks[0][0],clicks[0][1],label,2.5,0,'',lay,col);
      else
        id = window.cadAddMText(clicks[0][0],clicks[0][1],label,3,0,0,'',lay,col);
      state.pendingText = null;
      break;
    }
    case 'dimlin': case 'dimali':
      if (clicks.length >= 2) {
        const fn = tool === 'dimali' ? window.cadAddAlignedDim : window.cadAddLinearDim;
        id = fn(clicks[0][0],clicks[0][1],clicks[1][0],clicks[1][1],30,lay,col);
      }
      break;
    case 'dimang':
      if (clicks.length >= 3) {
        const r = Math.hypot(clicks[1][0]-clicks[0][0],clicks[1][1]-clicks[0][1]);
        id = window.cadAddAngularDim(clicks[0][0],clicks[0][1],clicks[1][0],clicks[1][1],clicks[2][0],clicks[2][1],r,lay,col);
      }
      break;
    case 'dimrad':
      if (clicks.length >= 2) {
        const r   = Math.hypot(clicks[1][0]-clicks[0][0],clicks[1][1]-clicks[0][1]);
        const ang = Math.atan2(clicks[1][1]-clicks[0][1],clicks[1][0]-clicks[0][0])*180/Math.PI;
        id = window.cadAddRadialDim(clicks[0][0],clicks[0][1],r,ang,lay,col);
      }
      break;
    case 'dimdia':
      if (clicks.length >= 2) {
        const r   = Math.hypot(clicks[1][0]-clicks[0][0],clicks[1][1]-clicks[0][1]);
        const ang = Math.atan2(clicks[1][1]-clicks[0][1],clicks[1][0]-clicks[0][0])*180/Math.PI;
        id = window.cadAddDiameterDim(clicks[0][0],clicks[0][1],r,ang,lay,col);
      }
      break;
    case 'hatch': {
      if (clicks.length < 3) { setStatus('Hatch needs ≥3 points'); break; }
      const pat   = document.getElementById('hatch-pattern')?.value || 'ANSI31';
      const angle = parseFloat(document.getElementById('hatch-angle')?.value || '0');
      const scale = parseFloat(document.getElementById('hatch-scale')?.value || '5');
      id = window.cadAddHatch(clicks,pat,angle,scale,lay,col);
      break;
    }
    case 'leader': {
      if (clicks.length < 2) { setStatus('Leader needs ≥2 points'); break; }
      const label2 = state.pendingText || prompt('Leader label (blank = none):','') || '';
      id = window.cadAddLeader(clicks,label2,lay,col);
      state.pendingText = null;
      break;
    }
    case 'revcloud':
      if (clicks.length >= 3)
        id = window.cadAddRevisionCloud(clicks, 1, lay, col);
      else setStatus('RevCloud needs ≥3 points');
      break;
    case 'wipeout':
      if (clicks.length >= 3)
        id = window.cadAddWipeout(clicks,lay,col);
      else setStatus('Wipeout needs ≥3 points');
      break;
  }

  state.clicks  = [];
  state.tempPt  = null;
  if (id >= 0) {
    state.selectedId = id;
    setStatus(`${tool} id=${id} created`);
  }
  render();
}

// ── Mouse event handlers ───────────────────────────────────────────────────────
let panStartX = 0, panStartY = 0, panStartPanX = 0, panStartPanY = 0;
let isPanning = false;

export function initTools() {
  const canvas = document.getElementById('canvas');
  if (!canvas) return;

  canvas.addEventListener('mousedown', onMouseDown);
  canvas.addEventListener('mousemove', onMouseMove);
  canvas.addEventListener('mouseup',   onMouseUp);
  canvas.addEventListener('wheel',     onWheel, { passive: false });
  canvas.addEventListener('dblclick',  onDblClick);
  canvas.addEventListener('contextmenu', e => { e.preventDefault(); state.clicks=[]; state.tempPt=null; drawSnapMarker(); render(); });
}

function onMouseDown(e) {
  const canvas = document.getElementById('canvas');
  // Middle mouse or Ctrl+left drag → pan
  if (e.button === 1 || (e.button === 0 && e.ctrlKey)) {
    isPanning  = true;
    panStartX  = e.clientX;
    panStartY  = e.clientY;
    panStartPanX = state.panX;
    panStartPanY = state.panY;
    canvas.style.cursor = 'grab';
    return;
  }
  if (e.button !== 0) return;

  const rect = canvas.getBoundingClientRect();
  const sx   = e.clientX - rect.left;
  const sy   = e.clientY - rect.top;

  const snap = snapWorldPt(sx, sy);
  const [wx, wy] = snap ? [snap.x, snap.y] : s2w(sx, sy);

  // ── Edit-op tools: pick entity on first click ────────────────────────────
  const editTools = ['move','copy','rotate','scale','mirror','trim','extend','fillet','chamfer','arrayrect','arraypolar','offset'];
  if (editTools.includes(state.currentTool)) {
    if (state.editPickIds.length === 0 && state.wasmReady) {
      // Try to pick entity
      const thresh = 12 / state.zoom;
      const ents   = JSON.parse(window.cadEntities() || '[]');
      let   best   = 0, bestD = thresh;
      ents.forEach(ent => {
        const pts = entitySamplePoints(ent);
        pts.forEach(([ex, ey]) => {
          const d = Math.hypot(ex - wx, ey - wy);
          if (d < bestD) { bestD = d; best = ent.id; }
        });
      });
      if (best) {
        state.selectedId    = best;
        state.editPickIds   = [best];
        const found = ents.find(e => e.id === best);
        if (found) showEntityProperties(found);
        setStatus(`${state.currentTool}: entity ${best} selected. Now pick second point or type command.`);
        render();
      } else {
        setStatus(`${state.currentTool}: no entity near cursor. Click on an entity.`);
      }
      return;
    }

    // Mirror: pick two axis points
    if (state.currentTool === 'mirror') {
      state.clicks.push([wx, wy]);
      if (state.clicks.length >= 2) {
        const ids   = state.editPickIds.filter(x => typeof x === 'number');
        const isCopy = state.editPickIds.includes('copy');
        const [ax,ay] = state.clicks[0], [bx,by] = state.clicks[1];
        const newIds  = JSON.parse(window.cadMirror(ids,ax,ay,bx,by,isCopy)||'[]');
        if (isCopy && newIds.length > 0) state.selectedId = newIds[0];
        setStatus(`Mirror${isCopy?' copy':''} done`);
        state.editPickIds=[]; state.clicks=[]; state.tempPt=null; render();
      } else {
        setStatus('Mirror: click second point on mirror axis.');
      }
      return;
    }
    // Trim/extend: pick second entity
    if (state.currentTool === 'trim' || state.currentTool === 'extend') {
      const thresh2 = 12 / state.zoom;
      const ents2   = JSON.parse(window.cadEntities() || '[]');
      let   best2   = 0, bestD2 = thresh2;
      ents2.forEach(ent => {
        if (ent.id === state.editPickIds[0]) return;
        const pts = entitySamplePoints(ent);
        pts.forEach(([ex, ey]) => {
          const d = Math.hypot(ex - wx, ey - wy);
          if (d < bestD2) { bestD2 = d; best2 = ent.id; }
        });
      });
      if (best2) {
        const fn  = state.currentTool === 'trim' ? window.cadTrim : window.cadExtend;
        const res = fn ? fn(state.editPickIds[0], best2) : 'null';
        setStatus(`${state.currentTool} done: ${res}`);
        state.editPickIds=[]; state.clicks=[]; state.tempPt=null; render();
      }
      return;
    }
    // Fillet/Chamfer: pick second line
    if (state.currentTool === 'fillet' || state.currentTool === 'chamfer') {
      if (state.editPickIds.length >= 2) {
        const thresh3 = 12 / state.zoom;
        const ents3   = JSON.parse(window.cadEntities() || '[]');
        let   best3   = 0, bestD3 = thresh3;
        ents3.forEach(ent => {
          if (ent.id === state.editPickIds[0]) return;
          entitySamplePoints(ent).forEach(([ex, ey]) => {
            const d = Math.hypot(ex - wx, ey - wy);
            if (d < bestD3) { bestD3 = d; best3 = ent.id; }
          });
        });
        if (best3) {
          const r  = typeof state.editPickIds[1] === 'number' ? state.editPickIds[1] : 1;
          const fn = state.currentTool === 'fillet'
            ? (id1, id2, rv) => window.cadFillet && window.cadFillet(id1, id2, rv)
            : (id1, id2, rv) => window.cadChamfer && window.cadChamfer(id1, id2, rv, rv);
          fn(state.editPickIds[0], best3, r);
          setStatus(`${state.currentTool} done`);
          state.editPickIds=[]; state.clicks=[]; state.tempPt=null; render();
        }
      }
      return;
    }
    return;
  }

  // ── Drawing tools: accumulate clicks ────────────────────────────────────────
  state.clicks.push([wx, wy]);
  const needed = toolClicks[state.currentTool];

  if (state.currentTool === 'text' || state.currentTool === 'mtext') {
    commitEntity(); return;
  }
  if (state.currentTool === 'leader') {
    setStatus(`Leader: ${state.clicks.length} pt(s). Click more, Enter to finish.`); render(); return;
  }

  if (needed && state.clicks.length >= needed) {
    commitEntity();
    return;
  }

  // Multi-click tools with Enter to finish
  const multiClick = ['poly','spline','nurbs','hatch','revcloud','wipeout'];
  if (multiClick.includes(state.currentTool)) {
    setStatus(`${state.currentTool}: ${state.clicks.length} pt(s). Click more, Enter/Esc to finish.`);
    render();
    return;
  }

  // Select / hover
  if (!toolClicks[state.currentTool] && !multiClick.includes(state.currentTool)) {
    // Pick entity
    if (state.wasmReady) {
      const thresh = 12 / state.zoom;
      const ents   = JSON.parse(window.cadEntities() || '[]');
      let   best   = 0, bestD = thresh;
      ents.forEach(ent => {
        entitySamplePoints(ent).forEach(([ex, ey]) => {
          const d = Math.hypot(ex - wx, ey - wy);
          if (d < bestD) { bestD = d; best = ent.id; }
        });
      });
      if (state.selectedId !== best) {
        state.selectedId = best;
        if (best) {
          const found = ents.find(e => e.id === best);
          if (found) showEntityProperties(found);
          setStatus(`Selected entity id=${best} (${ents.find(e=>e.id===best)?.type||'?'})`);
        } else {
          clearInspector();
          setStatus('Ready');
        }
        render();
      }
    }
  } else {
    // Show step hints
    const stepHints = {
      arc:     ['Arc: click radius point','Arc: click end angle point'],
      ellipse: ['Ellipse: click semi-major endpoint','Ellipse: click semi-minor endpoint'],
      dimang:  ['Angular Dim: click point on first ray','Angular Dim: click point on second ray'],
    };
    const hints = stepHints[state.currentTool];
    if (hints && state.clicks.length - 1 < hints.length) {
      setStatus(hints[state.clicks.length - 1]);
    }
    render();
  }
}

function onMouseMove(e) {
  const canvas = document.getElementById('canvas');
  const rect   = canvas.getBoundingClientRect();
  const sx     = e.clientX - rect.left;
  const sy     = e.clientY - rect.top;

  if (isPanning) {
    state.panX = panStartPanX + (e.clientX - panStartX);
    state.panY = panStartPanY + (e.clientY - panStartY);
    render();
    return;
  }

  const snap = snapWorldPt(sx, sy);
  if (snap) {
    state.snapResult = snap;
    state.tempPt     = [snap.x, snap.y];
  } else {
    state.snapResult = null;
    state.tempPt     = s2w(sx, sy);
  }

  // Update coordinate display
  const [wx, wy] = state.tempPt;
  const coordEl  = document.getElementById('coords-display');
  if (coordEl) coordEl.textContent = `${wx.toFixed(3)}, ${wy.toFixed(3)}`;

  // Update coord-bar display
  const coordBar = document.getElementById('coord-readout');
  if (coordBar) coordBar.textContent = `X: ${wx.toFixed(3)}   Y: ${wy.toFixed(3)}`;

  if (state.clicks.length > 0 || state.editPickIds.length > 0) {
    render();
  } else {
    drawSnapMarker();
  }
}

function onMouseUp(e) {
  if (isPanning) {
    isPanning = false;
    const canvas = document.getElementById('canvas');
    if (canvas) canvas.style.cursor = 'crosshair';
  }
}

function onDblClick(e) {
  if (['poly','spline','nurbs','hatch','revcloud','wipeout','leader'].includes(state.currentTool)) {
    commitEntity();
  }
}

function onWheel(e) {
  e.preventDefault();
  const canvas = document.getElementById('canvas');
  const rect   = canvas.getBoundingClientRect();
  const sx     = e.clientX - rect.left;
  const sy     = e.clientY - rect.top;
  const [wx, wy] = s2w(sx, sy);
  const factor = e.deltaY < 0 ? 1.15 : 1/1.15;
  state.zoom  *= factor;
  state.panX   = sx - wx * state.zoom;
  state.panY   = sy + wy * state.zoom;
  render();
}

// ── Coordinate input bar: parse and apply ──────────────────────────────────────
export function applyCoordInput(raw) {
  if (!raw.trim()) return false;
  let wx, wy;
  const last = state.clicks.length > 0
    ? state.clicks[state.clicks.length - 1]
    : (state.tempPt || [0, 0]);

  // Polar: @dist<angle
  const polarMatch = raw.match(/^@([\d.+-]+)<([\d.+-]+)$/);
  if (polarMatch) {
    const dist = parseFloat(polarMatch[1]);
    const ang  = parseFloat(polarMatch[2]) * Math.PI / 180;
    wx = last[0] + dist * Math.cos(ang);
    wy = last[1] + dist * Math.sin(ang);
    // Simulate a click at this world point
    state.clicks.push([wx, wy]);
    const needed = toolClicks[state.currentTool];
    if (needed && state.clicks.length >= needed) commitEntity();
    else render();
    return true;
  }

  // Relative: @dx,dy
  const relMatch = raw.match(/^@([\d.+-]+)[, ]([\d.+-]+)$/);
  if (relMatch) {
    wx = last[0] + parseFloat(relMatch[1]);
    wy = last[1] + parseFloat(relMatch[2]);
    state.clicks.push([wx, wy]);
    const needed = toolClicks[state.currentTool];
    if (needed && state.clicks.length >= needed) commitEntity();
    else render();
    return true;
  }

  // Absolute: x,y
  const absMatch = raw.match(/^([\d.+-]+)[, ]([\d.+-]+)$/);
  if (absMatch) {
    wx = parseFloat(absMatch[1]);
    wy = parseFloat(absMatch[2]);
    state.clicks.push([wx, wy]);
    const needed = toolClicks[state.currentTool];
    if (needed && state.clicks.length >= needed) commitEntity();
    else render();
    return true;
  }

  return false;
}
