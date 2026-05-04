// web/js/dialogs.js — Layer Manager, Block Manager, Symbols, Drafting Settings, Print/Plot
import { state, setStatus, escH, invalidateBlockCache } from './state.js';
import { render, zoomFit, w2s, s2w, entitySamplePoints } from './canvas.js';
import { setTool } from './tools.js';

// ── Helpers ────────────────────────────────────────────────────────────────────
function openModal(id)  {
  const el = document.getElementById(id);
  if (!el) return;
  el.classList.add('open');
  trapFocus(el);
}
function closeModal(id) {
  document.getElementById(id)?.classList.remove('open');
}

// ── Focus trap ─────────────────────────────────────────────────────────────────
const FOCUSABLE = 'a[href],button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),[tabindex]:not([tabindex="-1"])';

function trapFocus(modal) {
  // Focus the first focusable element
  const focusable = Array.from(modal.querySelectorAll(FOCUSABLE));
  if (focusable.length) focusable[0].focus();

  function onKeyDown(e) {
    if (!modal.classList.contains('open')) {
      document.removeEventListener('keydown', onKeyDown, true);
      return;
    }
    if (e.key === 'Escape') {
      e.preventDefault();
      modal.classList.remove('open');
      document.removeEventListener('keydown', onKeyDown, true);
      return;
    }
    if (e.key !== 'Tab') return;
    const focusableNow = Array.from(modal.querySelectorAll(FOCUSABLE));
    if (!focusableNow.length) return;
    const first = focusableNow[0];
    const last  = focusableNow[focusableNow.length - 1];
    if (e.shiftKey) {
      if (document.activeElement === first) { e.preventDefault(); last.focus(); }
    } else {
      if (document.activeElement === last)  { e.preventDefault(); first.focus(); }
    }
  }
  document.addEventListener('keydown', onKeyDown, true);
}

// ── Layer Manager ──────────────────────────────────────────────────────────────
export function openLayerManager() {
  refreshLayerTable();
  openModal('layer-modal');
}
export function closeLayerManager() { closeModal('layer-modal'); }

export function refreshLayers() {
  refreshLayerTable();
  refreshLayerSel();
}

export function refreshLayerSel() {
  const sel = document.getElementById('layer-sel');
  if (!sel || !state.wasmReady) return;
  const layers = JSON.parse(window.cadGetLayers() || '[]');
  sel.innerHTML = layers.map(l => `<option value="${l.id}" ${l.id===state.currentLayer?'selected':''}>${escH(l.name)}</option>`).join('');
}

export function refreshLayerTable() {
  if (!state.wasmReady) return;
  const layers = JSON.parse(window.cadGetLayers() || '[]');
  const tbody  = document.getElementById('layer-tbody');
  if (!tbody) return;

  tbody.innerHTML = layers.map(l => `
    <tr>
      <td><input type="radio" name="cur-layer" ${l.id===state.currentLayer?'checked':''} data-lid="${l.id}" class="layer-check layer-cur" aria-label="Set as current layer"></td>
      <td><input type="text" value="${escH(l.name)}" data-lid="${l.id}" class="layer-name-inp" aria-label="Layer name"></td>
      <td><input type="color" value="${l.color||'#ffffff'}" data-lid="${l.id}" class="layer-color-inp" aria-label="Layer colour"></td>
      <td>
        <select data-lid="${l.id}" class="layer-lt-sel" aria-label="Linetype">
          ${['Solid','Dashed','Dotted','DashDot','Center','Hidden'].map(lt =>
            `<option ${l.lineType===lt?'selected':''}>${lt}</option>`).join('')}
        </select>
      </td>
      <td><input type="number" value="${(l.lineWeight||0.25).toFixed(2)}" step="0.05" min="0.05" max="2" data-lid="${l.id}" class="layer-lw-inp" style="width:50px" aria-label="Lineweight"></td>
      <td><input type="checkbox" ${l.visible?'checked':''} data-lid="${l.id}" class="layer-check layer-vis" aria-label="Visible"></td>
      <td><input type="checkbox" ${l.locked?'checked':''} data-lid="${l.id}" class="layer-check layer-lck" aria-label="Locked"></td>
      <td><input type="checkbox" ${l.frozen?'checked':''} data-lid="${l.id}" class="layer-check layer-frz" aria-label="Frozen"></td>
      <td><input type="checkbox" ${l.print!==false?'checked':''} data-lid="${l.id}" class="layer-check layer-prt" aria-label="Print"></td>
      <td><button class="layer-btn layer-del" data-lid="${l.id}" ${l.id===0?'disabled':''} aria-label="Delete layer">Del</button></td>
    </tr>`).join('');

  // Wire events
  tbody.querySelectorAll('.layer-cur').forEach(el => el.addEventListener('change', () => {
    state.currentLayer = parseInt(el.dataset.lid);
    if (state.wasmReady) window.cadSetCurrentLayer(state.currentLayer);
    refreshLayerSel();
  }));
  tbody.querySelectorAll('.layer-color-inp').forEach(el => el.addEventListener('input', () => {
    if (state.wasmReady) window.cadSetLayerColor(parseInt(el.dataset.lid), el.value);
    render();
  }));
  tbody.querySelectorAll('.layer-lt-sel').forEach(el => el.addEventListener('change', () => {
    if (state.wasmReady) window.cadSetLayerLineType(parseInt(el.dataset.lid), el.value);
    render();
  }));
  tbody.querySelectorAll('.layer-vis').forEach(el => el.addEventListener('change', () => {
    if (state.wasmReady) window.cadSetLayerVisible(parseInt(el.dataset.lid), el.checked);
    render();
  }));
  tbody.querySelectorAll('.layer-lck').forEach(el => el.addEventListener('change', () => {
    if (state.wasmReady) window.cadSetLayerLocked(parseInt(el.dataset.lid), el.checked);
  }));
  tbody.querySelectorAll('.layer-frz').forEach(el => el.addEventListener('change', () => {
    if (state.wasmReady) window.cadSetLayerFrozen(parseInt(el.dataset.lid), el.checked);
    render();
  }));
  tbody.querySelectorAll('.layer-name-inp').forEach(el => el.addEventListener('change', () => {
    if (state.wasmReady) window.cadSetLayerName(parseInt(el.dataset.lid), el.value);
  }));
  tbody.querySelectorAll('.layer-lw-inp').forEach(el => el.addEventListener('change', () => {
    if (state.wasmReady) window.cadSetLayerLineWeight(parseInt(el.dataset.lid), parseFloat(el.value));
  }));
  tbody.querySelectorAll('.layer-prt').forEach(el => el.addEventListener('change', () => {
    if (state.wasmReady) window.cadSetLayerPrint(parseInt(el.dataset.lid), el.checked);
  }));
  tbody.querySelectorAll('.layer-del').forEach(el => el.addEventListener('click', () => {
    const id = parseInt(el.dataset.lid);
    if (id === 0) return;
    if (!confirm(`Delete layer ${id}? Entities on this layer will be moved to layer 0.`)) return;
    if (state.wasmReady && window.cadRemoveLayer) {
      const ok = window.cadRemoveLayer(id);
      if (ok) {
        if (state.currentLayer === id) state.currentLayer = 0;
        refreshLayerTable();
        refreshLayerSel();
        render();
        setStatus(`Layer ${id} deleted.`);
      } else {
        setStatus(`Could not delete layer ${id}.`);
      }
    }
  }));
}

// ── Block Manager ──────────────────────────────────────────────────────────────
export function openBlockManager() {
  if (!state.wasmReady) { setStatus('WASM not ready'); return; }
  invalidateBlockCache();
  const blocks = JSON.parse(window.cadGetBlocks() || '[]');
  const list   = document.getElementById('block-list');
  if (!list) return;

  if (!blocks.length) {
    list.innerHTML = '<div style="padding:12px;color:#666;font-size:12px">(no blocks defined — select entities then type: DEFINEBLOCK name)</div>';
  } else {
    list.innerHTML = blocks.map(b => `
      <div class="block-item">
        <canvas class="block-thumb" data-block="${escH(b.name)}" width="72" height="54" title="${escH(b.name)}"></canvas>
        <div class="block-info">
          <span class="block-name">${escH(b.name)}</span>
          <span class="block-meta">${b.count} ent. &nbsp;base(${b.baseX.toFixed(1)},${b.baseY.toFixed(1)})</span>
        </div>
        <button class="block-insert-btn" data-block="${escH(b.name)}">Insert</button>
      </div>`).join('');

    list.querySelectorAll('.block-thumb').forEach(cv => drawBlockPreview(cv, cv.dataset.block));
    list.querySelectorAll('.block-insert-btn').forEach(btn => {
      btn.addEventListener('click', () => promptInsertBlock(btn.dataset.block));
    });
  }
  openModal('block-modal');
}
export function closeBlockManager() { closeModal('block-modal'); }

export function drawBlockPreview(canvas, blockName) {
  const ents = (() => {
    if (!state.wasmReady || !window.cadGetBlockEntities) return [];
    try { return JSON.parse(window.cadGetBlockEntities(blockName) || '[]'); } catch (_) { return []; }
  })();

  const ctx2 = canvas.getContext('2d');
  const W = canvas.width, H = canvas.height;
  ctx2.clearRect(0, 0, W, H);

  if (!ents || ents.length === 0) {
    ctx2.fillStyle = '#555'; ctx2.font = '9px sans-serif'; ctx2.textAlign = 'center';
    ctx2.fillText('(empty)', W/2, H/2+3); return;
  }

  let minX=Infinity, minY=Infinity, maxX=-Infinity, maxY=-Infinity;
  function bboxPt(x,y) { if(x<minX)minX=x; if(y<minY)minY=y; if(x>maxX)maxX=x; if(y>maxY)maxY=y; }
  ents.forEach(be => {
    switch(be.type) {
      case 'line': bboxPt(be.x1,be.y1); bboxPt(be.x2,be.y2); break;
      case 'circle': case 'arc':
        bboxPt(be.cx-be.r,be.cy-be.r); bboxPt(be.cx+be.r,be.cy+be.r); break;
      case 'text': bboxPt(be.x1,be.y1); break;
      default: (be.points||[]).forEach(p => bboxPt(p[0],p[1]));
    }
  });
  if (!isFinite(minX)) { ctx2.fillStyle='#555'; ctx2.font='9px sans-serif'; ctx2.textAlign='center'; ctx2.fillText('?',W/2,H/2); return; }

  const pad=4, rangeX=maxX-minX||1, rangeY=maxY-minY||1;
  const sc = Math.min((W-pad*2)/rangeX, (H-pad*2)/rangeY);
  function toThumb(x,y) { return [pad+(x-minX)*sc, H-pad-(y-minY)*sc]; }

  ctx2.strokeStyle='#7ec8e3'; ctx2.fillStyle='#7ec8e3'; ctx2.lineWidth=1;
  ents.forEach(be => {
    switch(be.type) {
      case 'line': {
        const [ax,ay]=toThumb(be.x1,be.y1),[bx,by]=toThumb(be.x2,be.y2);
        ctx2.beginPath(); ctx2.moveTo(ax,ay); ctx2.lineTo(bx,by); ctx2.stroke(); break;
      }
      case 'circle': {
        const [cx2,cy2]=toThumb(be.cx,be.cy);
        ctx2.beginPath(); ctx2.arc(cx2,cy2,be.r*sc,0,2*Math.PI); ctx2.stroke(); break;
      }
      case 'arc': {
        const [cx2,cy2]=toThumb(be.cx,be.cy);
        ctx2.beginPath();
        ctx2.arc(cx2,cy2,be.r*sc,-be.startDeg*Math.PI/180,-be.endDeg*Math.PI/180,false);
        ctx2.stroke(); break;
      }
      default: {
        const pts2=be.points||[];
        if(pts2.length<2) break;
        const [fx,fy]=toThumb(pts2[0][0],pts2[0][1]);
        ctx2.beginPath(); ctx2.moveTo(fx,fy);
        for(let j=1;j<pts2.length;j++){const[px,py]=toThumb(pts2[j][0],pts2[j][1]);ctx2.lineTo(px,py);}
        ctx2.stroke();
      }
    }
  });
}

function promptInsertBlock(name) {
  const xStr = prompt(`Insert block "${name}" — X:`, '0');
  if (xStr === null) return;
  const yStr = prompt('Y:', '0');
  if (yStr === null) return;
  const scStr = prompt('Scale (1=full size):', '1');
  const x=parseFloat(xStr)||0, y=parseFloat(yStr)||0, sc=parseFloat(scStr)||1;
  const id = window.cadInsertBlock(name,x,y,sc,sc,0,state.currentLayer,state.currentColor);
  setStatus(id >= 0 ? `Inserted block "${name}" id=${id}` : `Failed to insert block "${name}"`);
  closeBlockManager();
  render();
}

// ── Symbols Panel ──────────────────────────────────────────────────────────────
export function openSymbolsPanel() {
  const names = state.wasmReady
    ? JSON.parse(window.cadGetSymbols() || '[]')
    : ['CENTER_MARK','NORTH_ARROW','REVISION_TRIANGLE','DATUM_TRIANGLE','SURFACE_FINISH'];
  const list = document.getElementById('sym-list');
  if (!list) return;

  list.innerHTML = names.map(n => `
    <div class="sym-item">
      <span>${escH(n)}</span>
      <button class="sym-insert-btn" data-sym="${escH(n)}">Insert</button>
    </div>`).join('');

  list.querySelectorAll('.sym-insert-btn').forEach(btn => {
    btn.addEventListener('click', () => promptInsertSymbol(btn.dataset.sym));
  });
  openModal('sym-modal');
}
export function closeSymbolsPanel() { closeModal('sym-modal'); }

function promptInsertSymbol(name) {
  const xStr = prompt(`Insert symbol "${name}" — X:`, '0');
  if (xStr === null) return;
  const yStr = prompt('Y:', '0');
  if (yStr === null) return;
  const scStr = prompt('Scale:', '1');
  const x=parseFloat(xStr)||0, y=parseFloat(yStr)||0, sc=parseFloat(scStr)||1;
  if (!state.wasmReady) { setStatus('WASM not ready'); return; }
  const id = window.cadInsertSymbol(name,x,y,sc,0,state.currentLayer,state.currentColor);
  setStatus(id >= 0 ? `Inserted symbol "${name}" id=${id}` : `Failed to insert symbol "${name}"`);
  closeSymbolsPanel();
  render();
}

// ── Drafting Settings Dialog ───────────────────────────────────────────────────
export function openDraftingSettings() {
  const modal = document.getElementById('drafting-modal');
  if (!modal) return;
  // Populate fields from localStorage / state
  document.getElementById('ds-grid-x').value    = localStorage.getItem('grid-x') || '10';
  document.getElementById('ds-grid-y').value    = localStorage.getItem('grid-y') || '10';
  document.getElementById('ds-grid-on').checked  = localStorage.getItem('grid-on') !== 'false';
  document.getElementById('ds-unit-sel').value   = localStorage.getItem('cad-units') || 'mm';
  document.getElementById('ds-precision').value  = localStorage.getItem('cad-precision') || '4';
  document.getElementById('ds-def-color').value  = state.currentColor || '#00ff00';
  const ltEl = document.getElementById('ds-def-linetype');
  if (ltEl) ltEl.value = localStorage.getItem('cad-def-linetype') || 'Solid';
  const lwEl = document.getElementById('ds-def-lw');
  if (lwEl) lwEl.value = localStorage.getItem('cad-def-lw') || '0.25';
  openModal('drafting-modal');
}
export function closeDraftingSettings() { closeModal('drafting-modal'); }

export function applyDraftingSettings() {
  const gx  = document.getElementById('ds-grid-x')?.value;
  const gy  = document.getElementById('ds-grid-y')?.value;
  const gon = document.getElementById('ds-grid-on')?.checked;
  const uni = document.getElementById('ds-unit-sel')?.value;
  const prec= document.getElementById('ds-precision')?.value;
  const col = document.getElementById('ds-def-color')?.value;
  const lt  = document.getElementById('ds-def-linetype')?.value;
  const lw  = document.getElementById('ds-def-lw')?.value;
  if (gx)  localStorage.setItem('grid-x', gx);
  if (gy)  localStorage.setItem('grid-y', gy);
  if (gon !== undefined) localStorage.setItem('grid-on', String(gon));
  if (uni) localStorage.setItem('cad-units', uni);
  if (prec) localStorage.setItem('cad-precision', prec);
  if (lt)  localStorage.setItem('cad-def-linetype', lt);
  if (lw)  localStorage.setItem('cad-def-lw', lw);
  if (col) {
    state.currentColor = col;
    const inp = document.getElementById('color-inp');
    if (inp) inp.value = col;
  }
  closeDraftingSettings();
  render();
  setStatus('Drafting settings applied.');
}

// ── Print / Plot Dialog ────────────────────────────────────────────────────────
export function openPrintPlot() {
  // Show/hide DPI row based on format
  const fmtEl  = document.getElementById('print-fmt');
  const dpiRow = document.getElementById('print-dpi-row');
  function updateDpiRow() {
    if (dpiRow) dpiRow.style.display = (fmtEl?.value === 'png') ? 'flex' : 'none';
  }
  updateDpiRow();
  fmtEl?.addEventListener('change', updateDpiRow);

  // Plot-area preview overlay — show on canvas behind dialog
  createPlotPreviewEl();
  updatePlotPreview();
  document.getElementById('print-area')?.addEventListener('change', updatePlotPreview);

  openModal('print-modal');
}
export function closePrintPlot() {
  removePlotPreview();
  closeModal('print-modal');
}

// ── Plot-area bounding box ──────────────────────────────────────────────────────
// Returns world-coordinate bounds {minX,minY,maxX,maxY} for the chosen area,
// or null if nothing to export.
function getPlotBounds(area) {
  const srcCanvas = document.getElementById('canvas');
  if (!srcCanvas) return null;
  if (area === 'view') {
    // Current viewport: convert screen corners to world coords via static import
    const [wx0, wy0] = s2w(0, 0);
    const [wx1, wy1] = s2w(srcCanvas.width, srcCanvas.height);
    return { minX: Math.min(wx0,wx1), maxX: Math.max(wx0,wx1),
             minY: Math.min(wy0,wy1), maxY: Math.max(wy0,wy1) };
  }
  // 'all' — entity extents
  if (!state.wasmReady || !window.cadEntities) return null;
  const ents = JSON.parse(window.cadEntities() || '[]');
  if (!ents.length) return null;
  let minX=Infinity, minY=Infinity, maxX=-Infinity, maxY=-Infinity;
  ents.forEach(e => entitySamplePoints(e).forEach(([x,y]) => {
    if(x<minX)minX=x; if(y<minY)minY=y; if(x>maxX)maxX=x; if(y>maxY)maxY=y;
  }));
  if (!isFinite(minX)) return null;
  const pad = (maxX - minX + maxY - minY) * 0.02 + 1;
  return { minX: minX-pad, maxX: maxX+pad, minY: minY-pad, maxY: maxY+pad };
}

// ── Plot-area preview overlay ───────────────────────────────────────────────────
function updatePlotPreview() {
  const overlay = document.getElementById('_plot-preview');
  if (!overlay) return;
  const area = document.getElementById('print-area')?.value || 'all';
  const srcCanvas = document.getElementById('canvas');
  if (!srcCanvas) { overlay.style.display = 'none'; return; }
  const bounds = getPlotBounds(area);
  if (!bounds) { overlay.style.display = 'none'; return; }
  const [sx0, sy0] = w2s(bounds.minX, bounds.maxY); // top-left in screen
  const [sx1, sy1] = w2s(bounds.maxX, bounds.minY); // bottom-right in screen
  const rect = srcCanvas.getBoundingClientRect();
  overlay.style.display = 'block';
  overlay.style.left    = (rect.left + Math.min(sx0,sx1)) + 'px';
  overlay.style.top     = (rect.top  + Math.min(sy0,sy1)) + 'px';
  overlay.style.width   = Math.abs(sx1-sx0) + 'px';
  overlay.style.height  = Math.abs(sy1-sy0) + 'px';
}

function createPlotPreviewEl() {
  let el = document.getElementById('_plot-preview');
  if (!el) {
    el = document.createElement('div');
    el.id = '_plot-preview';
    el.style.cssText = 'position:fixed;pointer-events:none;border:2px dashed #f90;background:rgba(255,153,0,0.06);z-index:50;display:none';
    document.body.appendChild(el);
  }
  return el;
}
function removePlotPreview() {
  document.getElementById('_plot-preview')?.remove();
}

export function executePrint() {
  const fmt     = document.getElementById('print-fmt')?.value || 'pdf';
  const margins = parseInt(document.getElementById('print-margins')?.value || '10');
  const dpi     = parseInt(document.getElementById('print-dpi')?.value || '300');
  const area    = document.getElementById('print-area')?.value || 'all';
  const scale   = document.getElementById('print-scale')?.value || 'fit';
  const paper   = document.getElementById('print-paper')?.value || 'a4';
  const orient  = document.getElementById('print-orient')?.value || 'landscape';

  // scaleRatio: mm of paper per world unit (0 = fit-to-page).
  // e.g. 1:100 → 1 world unit = 0.01 mm on paper → scaleRatio = 0.01
  const scaleRatios = { 'fit': 0, '1:1': 1, '1:2': 0.5, '1:5': 0.2,
                        '1:10': 0.1, '1:50': 0.02, '1:100': 0.01 };
  const scaleRatio = scaleRatios[scale] ?? 0;

  if (fmt === 'pdf') {
    // Apply @page CSS for chosen paper/orientation/margins, then invoke browser print
    const pageCSS = `@page { size: ${paper} ${orient}; margin: ${margins}mm; }`;
    let styleEl = document.getElementById('_print-page-style');
    if (!styleEl) {
      styleEl = document.createElement('style');
      styleEl.id = '_print-page-style';
      document.head.appendChild(styleEl);
    }
    styleEl.textContent = pageCSS;
    window.print();
    setStatus(`PDF print dialog opened (${paper} ${orient}, ${margins}mm margins, area: ${area}).`);

  } else if (fmt === 'svg') {
    if (!state.wasmReady || !window.cadExportSVG) { setStatus('SVG export requires WASM.'); return; }
    let svgStr = window.cadExportSVG();
    const bounds = getPlotBounds(area);
    if (bounds) {
      // Crop SVG to the desired area by replacing viewBox and dimensions.
      // The SVG writer uses CAD world coordinates directly (Y increases upward in CAD,
      // SVG Y increases downward — the writer uses entity coords verbatim so we match).
      const vbW = bounds.maxX - bounds.minX;
      const vbH = bounds.maxY - bounds.minY;
      const wMM = scaleRatio > 0 ? vbW * scaleRatio : vbW;
      const hMM = scaleRatio > 0 ? vbH * scaleRatio : vbH;
      // Replace both viewBox and width/height in the <svg> opening tag
      svgStr = svgStr.replace(
        /(<svg[^>]*)\sviewBox="[^"]*"[^>]*width="[^"]*"[^>]*height="[^"]*"/,
        `$1 viewBox="${bounds.minX.toFixed(4)} ${bounds.minY.toFixed(4)} ${vbW.toFixed(4)} ${vbH.toFixed(4)}" width="${wMM.toFixed(2)}mm" height="${hMM.toFixed(2)}mm"`
      );
    }
    const a = document.createElement('a');
    a.href = URL.createObjectURL(new Blob([svgStr], { type: 'image/svg+xml' }));
    a.download = 'drawing.svg';
    a.click();
    URL.revokeObjectURL(a.href);
    setStatus(`Exported drawing.svg (area: ${area}, scale: ${scale}).`);

  } else {
    // PNG export — crop to area bounds, apply named scale or DPI-only scaling
    const srcCanvas = document.getElementById('canvas');
    if (!srcCanvas) return;

    // Crop source region from current canvas based on plot bounds
    let sx0 = 0, sy0 = 0, sw = srcCanvas.width, sh = srcCanvas.height;
    const bounds = getPlotBounds(area);
    if (bounds) {
      const [bsx0, bsy0] = w2s(bounds.minX, bounds.maxY); // top-left in screen
      const [bsx1, bsy1] = w2s(bounds.maxX, bounds.minY); // bottom-right in screen
      sx0 = Math.max(0, Math.floor(Math.min(bsx0, bsx1)));
      sy0 = Math.max(0, Math.floor(Math.min(bsy0, bsy1)));
      sw  = Math.min(srcCanvas.width  - sx0, Math.ceil(Math.abs(bsx1 - bsx0)));
      sh  = Math.min(srcCanvas.height - sy0, Math.ceil(Math.abs(bsy1 - bsy0)));
    }

    // Apply scale to output resolution.
    // For named scale: desiredPixelsPerWorldUnit = scaleRatio * dpi / 25.4
    // Current canvas has state.zoom pixels per world unit in the source crop.
    // Scale factor from source crop to output = desiredPPWU / state.zoom
    // For "fit": use DPI / 96 (screen-to-print resolution upgrade only)
    let scaleFactor;
    if (scaleRatio > 0 && bounds && state.zoom > 0) {
      const desiredPPWU = scaleRatio * dpi / 25.4; // pixels-per-world-unit at print scale
      scaleFactor = desiredPPWU / state.zoom;
    } else {
      scaleFactor = dpi / 96;
    }

    const outW = Math.max(1, Math.round(sw * scaleFactor));
    const outH = Math.max(1, Math.round(sh * scaleFactor));
    const offscreen = document.createElement('canvas');
    offscreen.width  = outW;
    offscreen.height = outH;
    const ctx2 = offscreen.getContext('2d');
    ctx2.scale(scaleFactor, scaleFactor);
    ctx2.drawImage(srcCanvas, sx0, sy0, sw, sh, 0, 0, sw, sh);
    const a = document.createElement('a');
    a.href     = offscreen.toDataURL('image/png');
    a.download = `drawing_${dpi}dpi.png`;
    a.click();
    setStatus(`Exported PNG ${outW}×${outH}px (area: ${area}, scale: ${scale}, ${dpi}dpi).`);
  }
  removePlotPreview();
  closePrintPlot();
}

// ── Wire all dialog button events ──────────────────────────────────────────────
export function initDialogs() {
  // Layer modal
  document.getElementById('btn-layers')?.addEventListener('click', openLayerManager);
  document.getElementById('btn-close-layers')?.addEventListener('click', closeLayerManager);
  document.getElementById('btn-add-layer')?.addEventListener('click', () => {
    const name = prompt('New layer name:', 'Layer ' + Date.now());
    if (!name) return;
    if (state.wasmReady) window.cadAddLayer(name, '#ffffff');
    refreshLayers();
  });
  document.getElementById('layer-modal')?.addEventListener('click', e => {
    if (e.target === document.getElementById('layer-modal')) closeLayerManager();
  });
  document.getElementById('layer-sel')?.addEventListener('change', e => {
    state.currentLayer = +e.target.value;
    if (state.wasmReady) window.cadSetCurrentLayer(state.currentLayer);
  });

  // Block modal
  document.getElementById('btn-blocks')?.addEventListener('click', openBlockManager);
  document.getElementById('btn-close-blocks')?.addEventListener('click', closeBlockManager);
  document.getElementById('btn-close-blocks2')?.addEventListener('click', closeBlockManager);
  document.getElementById('block-modal')?.addEventListener('click', e => {
    if (e.target === document.getElementById('block-modal')) closeBlockManager();
  });

  // Symbols modal
  document.getElementById('btn-symbols')?.addEventListener('click', openSymbolsPanel);
  document.getElementById('btn-close-sym')?.addEventListener('click', closeSymbolsPanel);
  document.getElementById('sym-modal')?.addEventListener('click', e => {
    if (e.target === document.getElementById('sym-modal')) closeSymbolsPanel();
  });

  // Drafting settings — wire both close buttons (✕ header + Cancel footer)
  document.getElementById('btn-drafting')?.addEventListener('click', openDraftingSettings);
  document.getElementById('btn-close-drafting-x')?.addEventListener('click', closeDraftingSettings);
  document.getElementById('btn-close-drafting')?.addEventListener('click', closeDraftingSettings);
  document.getElementById('btn-apply-drafting')?.addEventListener('click', applyDraftingSettings);
  document.getElementById('drafting-modal')?.addEventListener('click', e => {
    if (e.target === document.getElementById('drafting-modal')) closeDraftingSettings();
  });

  // Print/Plot — wire both close buttons (✕ header + Cancel footer)
  document.getElementById('btn-print')?.addEventListener('click', openPrintPlot);
  document.getElementById('btn-close-print-x')?.addEventListener('click', closePrintPlot);
  document.getElementById('btn-close-print')?.addEventListener('click', closePrintPlot);
  document.getElementById('btn-execute-print')?.addEventListener('click', executePrint);
  document.getElementById('print-modal')?.addEventListener('click', e => {
    if (e.target === document.getElementById('print-modal')) closePrintPlot();
  });

  // Color input
  document.getElementById('color-inp')?.addEventListener('input', e => {
    state.currentColor = e.target.value;
  });
}
