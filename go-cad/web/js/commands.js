// web/js/commands.js — Command bar processing and history
import { state, setStatus, escH, invalidateBlockCache } from './state.js';
import { render, zoomFit } from './canvas.js';
import { setTool, tryDelete, commitEntity, applyCoordInput } from './tools.js';
import { addHistoryEntry, updateLastResult } from './history.js';

const toolMap = {
  LINE:'line', L:'line',
  CIRCLE:'circle', C:'circle',
  ARC:'arc', A:'arc',
  RECT:'rect', RECTANGLE:'rect', R:'rect',
  POLY:'poly', POLYLINE:'poly', P:'poly',
  SPLINE:'spline', SP:'spline', S:'spline',
  NURBS:'nurbs', N:'nurbs',
  ELLIPSE:'ellipse', EL:'ellipse', E:'ellipse',
  TEXT:'text', T:'text',
  MTEXT:'mtext', MT:'mtext',
  DIMLIN:'dimlin', DIMALI:'dimali', DIMANG:'dimang', DIMRAD:'dimrad', DIMDIA:'dimdia',
  UNDO:'undo', U:'undo', REDO:'redo',
  DELETE:'delete', DEL:'delete', ERASE:'delete',
  CLEAR:'clear',
  ZOOMFIT:'zoomfit', ZF:'zoomfit', Z:'zoomfit',
  EXPORT:'export', DXF:'export',
  MOVE:'move', M:'move',
  COPY:'copy', CP:'copy', CO:'copy',
  ROTATE:'rotate', RO:'rotate',
  SCALE:'scale', SC:'scale',
  MIRROR:'mirror', MI:'mirror',
  TRIM:'trim', TR:'trim',
  EXTEND:'extend', EX:'extend',
  FILLET:'fillet', F:'fillet',
  CHAMFER:'chamfer', CHA:'chamfer',
  ARRAYRECT:'arrayrect', AR:'arrayrect',
  ARRAYPOLAR:'arraypolar', AP:'arraypolar',
  OFFSET:'offset', O:'offset',
  HATCH:'hatch', H:'hatch',
  LEADER:'leader', LD:'leader',
  REVCLOUD:'revcloud', RC:'revcloud',
  WIPEOUT:'wipeout', WP:'wipeout',
  DRAFTING:'drafting', DS:'drafting',
  PRINT:'print', PLOT:'print',
};

const verbAliases = {
  M:'MOVE', CP:'COPY', CO:'COPY', RO:'ROTATE', SC:'SCALE',
  MI:'MIRROR', TR:'TRIM', EX:'EXTEND', F:'FILLET', CHA:'CHAMFER',
  AR:'ARRAYRECT', AP:'ARRAYPOLAR', O:'OFFSET',
};

export function processCommand(raw) {
  if (!raw.trim()) return;
  const cmd  = raw.trim().toUpperCase();
  const parts = raw.trim().split(/\s+/);
  const rawVerb = parts[0].toUpperCase();

  addHistoryEntry(raw);

  // Wrap setStatus so every call also updates the history entry result
  const _st = (msg) => { setStatus(msg); updateLastResult(msg); };

  // ── Coordinate input (point entry from command bar) ──────────────────────
  if (applyCoordInput(raw)) { return; }

  // ── DEFINEBLOCK name [bx by] [id …] ───────────────────────────────────────
  if (rawVerb === 'DEFINEBLOCK' && parts.length >= 2 && state.wasmReady) {
    const bname = parts[1];
    if (!bname) { setStatus('Usage: DEFINEBLOCK name [id id …]'); return; }
    let ids = [];
    if (parts.length >= 3) {
      ids = parts.slice(2).map(s => parseInt(s)).filter(n => !isNaN(n));
    } else if (state.selectedId) {
      ids = [state.selectedId];
    }
    if (ids.length === 0) { setStatus('DEFINEBLOCK: select at least one entity or list IDs'); return; }
    const entities = JSON.parse(window.cadEntities() || '[]');
    const firstEnt = entities.find(e => e.id === ids[0]);
    let bx = 0, by = 0;
    if (firstEnt) {
      if (firstEnt.type === 'circle' || firstEnt.type === 'arc') { bx = firstEnt.cx||0; by = firstEnt.cy||0; }
      else { bx = firstEnt.x1||0; by = firstEnt.y1||0; }
    }
    const ok = window.cadDefineBlock(bname, bx, by, JSON.stringify(ids));
    if (ok) { invalidateBlockCache(bname); _st(`Block "${bname}" defined with ${ids.length} entity(s).`); }
    else _st('DEFINEBLOCK failed — ensure entity IDs exist.');
    render(); return;
  }

  // ── EXPLODE id ─────────────────────────────────────────────────────────────
  if (rawVerb === 'EXPLODE' && parts.length >= 2 && state.wasmReady) {
    const eid = parseInt(parts[1]);
    if (!isNaN(eid)) {
      const r = window.cadExplodeBlock(eid);
      _st(r && r !== 'null' ? `Exploded block ${eid} → ${r}` : `Entity ${eid} is not a block reference.`);
      render();
    }
    return;
  }

  // ── INSERT name x y [sx [sy [rot]]] ───────────────────────────────────────
  if (rawVerb === 'INSERT' && parts.length >= 4 && state.wasmReady) {
    const name = parts[1];
    const x = parseFloat(parts[2]), y = parseFloat(parts[3]);
    const sx = parts[4] ? parseFloat(parts[4]) : 1;
    const sy2 = parts[5] ? parseFloat(parts[5]) : sx;
    const rot = parts[6] ? parseFloat(parts[6]) : 0;
    const id  = window.cadInsertBlock(name, x, y, sx, sy2, rot, state.currentLayer, state.currentColor);
    _st(id >= 0 ? `Inserted block "${name}" id=${id}` : `Block "${name}" not found.`);
    render(); return;
  }

  // ── BLOCKS / SYMBOLS ───────────────────────────────────────────────────────
  if (rawVerb === 'BLOCKS' && state.wasmReady) {
    import('./dialogs.js').then(m => m.openBlockManager()); return;
  }
  if (rawVerb === 'SYMBOLS' && state.wasmReady) {
    import('./dialogs.js').then(m => m.openSymbolsPanel()); return;
  }
  if (rawVerb === 'LAYERS' || rawVerb === 'LA') {
    import('./dialogs.js').then(m => m.openLayerManager()); return;
  }
  if (rawVerb === 'DRAFTING' || rawVerb === 'DS') {
    import('./dialogs.js').then(m => m.openDraftingSettings()); return;
  }
  if (rawVerb === 'PRINT' || rawVerb === 'PLOT') {
    import('./dialogs.js').then(m => m.openPrintPlot()); return;
  }

  // ── Leader label from command bar ─────────────────────────────────────────
  if (state.currentTool === 'leader' && state.clicks.length >= 1 && rawVerb !== '' && !(rawVerb in toolMap)) {
    state.pendingText = raw.trim();
    commitEntity(); return;
  }

  // ── Edit op commands ────────────────────────────────────────────────────────
  const verb = verbAliases[rawVerb] || rawVerb;

  if (state.wasmReady && state.selectedId && verb === 'MOVE' && parts.length >= 3) {
    const dx=parseFloat(parts[1]),dy=parseFloat(parts[2]);
    window.cadMove([state.selectedId],dx,dy); render(); return;
  }
  if (state.wasmReady && state.selectedId && verb === 'COPY' && parts.length >= 3) {
    const dx=parseFloat(parts[1]),dy=parseFloat(parts[2]);
    const newIDs=JSON.parse(window.cadCopy([state.selectedId],dx,dy)||'[]');
    if(newIDs.length>0) state.selectedId=newIDs[0];
    render(); return;
  }
  if (state.wasmReady && state.selectedId && (verb === 'ROTATE' || verb === 'ROTATECOPY') && parts.length >= 2) {
    const makeCopy = verb === 'ROTATECOPY';
    const ang=parseFloat(parts[1]);
    const cx=parts.length>=4?parseFloat(parts[2]):0;
    const cy=parts.length>=4?parseFloat(parts[3]):0;
    const newIDs=JSON.parse(window.cadRotate([state.selectedId],cx,cy,ang,makeCopy)||'[]');
    if(makeCopy && newIDs.length>0) state.selectedId=newIDs[0];
    render(); return;
  }
  if (state.wasmReady && state.selectedId && (verb === 'SCALE' || verb === 'SCALECOPY') && parts.length >= 2) {
    const makeCopy = verb === 'SCALECOPY';
    const sx=parseFloat(parts[1]);
    let sy=sx,cx=0,cy=0;
    if(parts.length===4){cx=parseFloat(parts[2]);cy=parseFloat(parts[3]);}
    else if(parts.length===3){sy=parseFloat(parts[2]);}
    else if(parts.length>=5){sy=parseFloat(parts[2]);cx=parseFloat(parts[3]);cy=parseFloat(parts[4]);}
    const newIDs=JSON.parse(window.cadScale([state.selectedId],cx,cy,sx,sy,makeCopy)||'[]');
    if(makeCopy && newIDs.length>0) state.selectedId=newIDs[0];
    render(); return;
  }
  if (state.wasmReady && state.selectedId && verb === 'MIRRORCOPY') {
    state.editPickIds = ['copy'];
    setTool('mirror');
    _st('Mirror-Copy: click first point on mirror axis.'); return;
  }
  if (state.wasmReady && state.selectedId && verb === 'FILLET' && parts.length >= 2) {
    const r=parseFloat(parts[1]);
    state.editPickIds=[state.selectedId,r];
    setTool('fillet');
    _st(`Fillet r=${r}: click the second line.`); return;
  }
  if (state.wasmReady && (state.currentTool === 'fillet') && state.editPickIds.length === 1 && verb === 'FILLET' && parts.length >= 2) {
    const r=parseFloat(parts[1]);
    _st(`Fillet r=${r}: now click the second line.`);
    state.editPickIds.push(r); return;
  }
  if (state.wasmReady && state.selectedId && verb === 'CHAMFER' && parts.length >= 3) {
    const d1=parseFloat(parts[1]),d2=parseFloat(parts[2]);
    state.editPickIds=[state.selectedId,d1,d2];
    setTool('chamfer');
    _st(`Chamfer d1=${d1} d2=${d2}: click the second line.`); return;
  }
  if (state.wasmReady && state.selectedId && verb === 'ARRAYRECT' && parts.length >= 5) {
    const rows=parseInt(parts[1]),cols=parseInt(parts[2]);
    const rs=parseFloat(parts[3]),cs=parseFloat(parts[4]);
    JSON.parse(window.cadArrayRect([state.selectedId],rows,cols,rs,cs)||'[]');
    render(); return;
  }
  if (state.wasmReady && state.selectedId && verb === 'ARRAYPOLAR' && parts.length >= 3) {
    const count=parseInt(parts[1]),ang=parseFloat(parts[2]);
    const cx=parts.length>=5?parseFloat(parts[3]):0;
    const cy=parts.length>=5?parseFloat(parts[4]):0;
    JSON.parse(window.cadArrayPolar([state.selectedId],cx,cy,count,ang)||'[]');
    render(); return;
  }
  if (state.wasmReady && state.selectedId && verb === 'OFFSET' && parts.length >= 2) {
    const dist=parseFloat(parts[1]);
    const newIDs=JSON.parse(window.cadOffset([state.selectedId],dist)||'[]');
    _st(newIDs.length>0 ? `Offset OK: ${newIDs.length} entity(ies).` : 'Offset failed.');
    render(); return;
  }

  // ── Tool map ───────────────────────────────────────────────────────────────
  const action = toolMap[cmd];
  if (!action) { _st(`Unknown command: ${raw}`); return; }

  switch (action) {
    case 'undo':
      if (state.wasmReady) window.cadUndo();
      _st('Undo'); render(); break;
    case 'redo':
      if (state.wasmReady) window.cadRedo();
      _st('Redo'); render(); break;
    case 'delete':
      if (state.wasmReady && state.selectedId) {
        tryDelete(state.selectedId); state.selectedId = 0;
        _st('Entity deleted'); render();
      }
      break;
    case 'clear':
      if (state.wasmReady) window.cadClear();
      invalidateBlockCache();
      _st('Drawing cleared'); render(); break;
    case 'zoomfit':
      _st('Zoom to fit'); zoomFit(); break;
    case 'export':
      exportDXF(); break;
    case 'drafting':
      import('./dialogs.js').then(m => m.openDraftingSettings()); break;
    case 'print':
      import('./dialogs.js').then(m => m.openPrintPlot()); break;
    default:
      _st(`Tool: ${action}`); setTool(action);
  }
}

function exportDXF() {
  if (!state.wasmReady) return;
  const fmt = document.getElementById('export-fmt')?.value || 'r2000';
  let content, mime, ext;
  if (fmt === 'svg') {
    content = window.cadExportSVG();
    mime = 'image/svg+xml'; ext = 'svg';
  } else if (fmt === 'r12') {
    content = window.cadExportDXFR12();
    mime = 'text/plain'; ext = 'dxf';
  } else {
    content = window.cadExportDXF();
    mime = 'text/plain'; ext = 'dxf';
  }
  const a = document.createElement('a');
  a.href = URL.createObjectURL(new Blob([content], { type: mime }));
  a.download = `drawing_${fmt}.${ext}`;
  a.click();
  URL.revokeObjectURL(a.href);
}

export function initCommandBar() {
  const input = document.getElementById('cmd-input');
  if (!input) return;

  input.addEventListener('keydown', e => {
    if (e.key === 'Enter') {
      e.preventDefault();
      const raw = input.value;
      input.value = '';
      processCommand(raw);
    }
    if (e.key === 'Escape') {
      input.value = '';
      state.clicks  = [];
      state.tempPt  = null;
      import('./canvas.js').then(m => { import('./snap.js').then(s => s.drawSnapMarker()); m.render(); });
    }
  });

  // Coord input bar
  const coordInput = document.getElementById('coord-input');
  if (coordInput) {
    coordInput.addEventListener('keydown', e => {
      if (e.key === 'Enter') {
        const val = coordInput.value.trim();
        coordInput.value = '';
        if (!applyCoordInput(val)) {
          processCommand(val);
        }
      }
    });
  }
}
