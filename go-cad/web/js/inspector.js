// web/js/inspector.js — Properties Inspector (right panel)
import { state, setStatus, escH } from './state.js';

export function initInspector() {
  clearInspector();
}

export function clearInspector() {
  const el = document.getElementById('inspector-content');
  if (!el) return;
  el.innerHTML = '<div class="insp-empty">No entity selected.<br><small>Click an entity to inspect it.</small></div>';
}

export function showEntityProperties(entity) {
  const el = document.getElementById('inspector-content');
  if (!el || !entity) { clearInspector(); return; }

  const rows = buildPropertyRows(entity);
  el.innerHTML = `
    <div class="insp-type-row">
      <span class="insp-type">${escH(entity.type.toUpperCase())}</span>
      <span class="insp-id">#${entity.id}</span>
    </div>
    <div class="insp-table">${rows}</div>`;

  // Wire editable fields
  el.querySelectorAll('.insp-field[data-field]').forEach(input => {
    input.addEventListener('change', () => {
      const field = input.dataset.field;
      let val = input.value;
      if (input.type === 'color') val = input.value;
      if (!state.wasmReady) return;
      if (window.cadSetEntityProp) {
        const ok = window.cadSetEntityProp(entity.id, field, val);
        if (ok) {
          import('./canvas.js').then(m => m.render());
          setStatus(`Updated ${field} = ${val}`);
        } else {
          setStatus(`Could not update ${field}`);
        }
      }
    });
    // Live update for color picker
    if (input.type === 'color') {
      input.addEventListener('input', () => {
        if (!state.wasmReady || !window.cadSetEntityProp) return;
        window.cadSetEntityProp(entity.id, 'color', input.value);
        import('./canvas.js').then(m => m.render());
      });
    }
  });
}

function prop(label, value, field = null, type = 'text') {
  const display = value === null || value === undefined ? '—' : String(value);
  if (field) {
    const inputVal = type === 'color'
      ? (value && /^#/.test(value) ? value : '#ffffff')
      : escH(display);
    return `<div class="insp-row">
      <span class="insp-label">${escH(label)}</span>
      <input class="insp-field" data-field="${escH(field)}" type="${type}" value="${inputVal}">
    </div>`;
  }
  return `<div class="insp-row">
    <span class="insp-label">${escH(label)}</span>
    <span class="insp-value">${escH(display)}</span>
  </div>`;
}

function num(v, dec = 4) {
  return typeof v === 'number' ? v.toFixed(dec) : '—';
}

function buildPropertyRows(e) {
  const rows = [];
  // Common props
  rows.push(prop('Layer', e.layer ?? 0, 'layer', 'number'));
  const colorVal = e.color && /^#/.test(e.color) ? e.color : null;
  rows.push(prop('Color', e.color || 'BYLAYER', colorVal ? 'color' : null, 'color'));

  switch (e.type) {
    case 'line':
      rows.push(prop('X1', num(e.x1)), prop('Y1', num(e.y1)));
      rows.push(prop('X2', num(e.x2)), prop('Y2', num(e.y2)));
      rows.push(prop('Length', num(Math.hypot(e.x2 - e.x1, e.y2 - e.y1))));
      rows.push(prop('Angle°', num(Math.atan2(e.y2 - e.y1, e.x2 - e.x1) * 180 / Math.PI, 2)));
      break;
    case 'circle':
      rows.push(prop('CX', num(e.cx)), prop('CY', num(e.cy)));
      rows.push(prop('Radius', num(e.r)));
      rows.push(prop('Diameter', num(e.r * 2)));
      rows.push(prop('Area', num(Math.PI * e.r * e.r)));
      break;
    case 'arc':
      rows.push(prop('CX', num(e.cx)), prop('CY', num(e.cy)));
      rows.push(prop('Radius', num(e.r)));
      rows.push(prop('Start°', num(e.startDeg, 2)), prop('End°', num(e.endDeg, 2)));
      break;
    case 'rectangle':
      rows.push(prop('X1', num(e.x1)), prop('Y1', num(e.y1)));
      rows.push(prop('X2', num(e.x2)), prop('Y2', num(e.y2)));
      rows.push(prop('Width', num(Math.abs(e.x2 - e.x1))));
      rows.push(prop('Height', num(Math.abs(e.y2 - e.y1))));
      rows.push(prop('Area', num(Math.abs((e.x2 - e.x1) * (e.y2 - e.y1)))));
      break;
    case 'text':
    case 'mtext':
      rows.push(prop('Text', e.text || '', 'text'));
      rows.push(prop('Height', num(e.textHeight, 3), 'textHeight', 'number'));
      rows.push(prop('Rotation°', num(e.rotDeg, 2), 'rotDeg', 'number'));
      rows.push(prop('X', num(e.x1)), prop('Y', num(e.y1)));
      if (e.font) rows.push(prop('Font', e.font));
      break;
    case 'ellipse':
      rows.push(prop('CX', num(e.cx)), prop('CY', num(e.cy)));
      rows.push(prop('Semi-major', num(e.r)));
      rows.push(prop('Semi-minor', num(e.r2)));
      rows.push(prop('Rotation°', num(e.rotDeg, 2), 'rotDeg', 'number'));
      break;
    case 'polyline':
    case 'spline':
    case 'nurbs':
      rows.push(prop('Points', (e.points || []).length));
      if (e.nurbsDeg) rows.push(prop('Degree', e.nurbsDeg));
      break;
    case 'blockref':
      rows.push(prop('Block Name', e.text || ''));
      rows.push(prop('X', num(e.x1)), prop('Y', num(e.y1)));
      rows.push(prop('ScaleX', num(e.r, 3)), prop('ScaleY', num(e.r2, 3)));
      rows.push(prop('Rotation°', num(e.rotDeg, 2), 'rotDeg', 'number'));
      break;
    case 'hatch':
      rows.push(prop('Pattern', e.text || 'ANSI31'));
      rows.push(prop('Points', (e.points || []).length));
      break;
    case 'leader':
      rows.push(prop('Label', e.text || '', 'text'));
      rows.push(prop('Points', (e.points || []).length));
      break;
    case 'dimlin':
    case 'dimali':
      rows.push(prop('X1', num(e.x1)), prop('Y1', num(e.y1)));
      rows.push(prop('X2', num(e.x2)), prop('Y2', num(e.y2)));
      rows.push(prop('Offset', num(e.cx, 2)));
      rows.push(prop('Measure', num(Math.hypot(e.x2 - e.x1, e.y2 - e.y1))));
      break;
    case 'dimang':
      rows.push(prop('Vertex CX', num(e.cx)), prop('Vertex CY', num(e.cy)));
      rows.push(prop('Radius', num(e.r)));
      break;
    case 'dimrad':
    case 'dimdia':
      rows.push(prop('CX', num(e.cx)), prop('CY', num(e.cy)));
      rows.push(prop('Radius', num(e.r)));
      break;
    default:
      if (e.points) rows.push(prop('Points', (e.points || []).length));
  }
  return rows.join('');
}
