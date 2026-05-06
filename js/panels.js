// web/js/panels.js — resizable panel system with localStorage persistence

const DEFAULTS = { leftW: 200, rightW: 240, bottomH: 160 };

function clamp(val, min, max) { return Math.max(min, Math.min(max, val)); }

function save() {
  const lp = document.getElementById('left-panel');
  const rp = document.getElementById('right-panel');
  const bp = document.getElementById('bottom-panel');
  if (lp) localStorage.setItem('go-cad-left-w', lp.offsetWidth);
  if (rp) localStorage.setItem('go-cad-right-w', rp.offsetWidth);
  if (bp) localStorage.setItem('go-cad-bottom-h', bp.offsetHeight);
}

function load() {
  const lp = document.getElementById('left-panel');
  const rp = document.getElementById('right-panel');
  const bp = document.getElementById('bottom-panel');
  const lw = parseInt(localStorage.getItem('go-cad-left-w')) || DEFAULTS.leftW;
  const rw = parseInt(localStorage.getItem('go-cad-right-w')) || DEFAULTS.rightW;
  const bh = parseInt(localStorage.getItem('go-cad-bottom-h')) || DEFAULTS.bottomH;
  if (lp) lp.style.width = lw + 'px';
  if (rp) rp.style.width = rw + 'px';
  if (bp) bp.style.height = bh + 'px';
}

function makeSplitter(splitterId, targetId, direction) {
  const splitter = document.getElementById(splitterId);
  const target   = document.getElementById(targetId);
  if (!splitter || !target) return;

  let dragging = false, startPos = 0, startSize = 0;

  splitter.addEventListener('mousedown', e => {
    e.preventDefault();
    dragging  = true;
    startPos  = direction === 'h' ? e.clientY : e.clientX;
    startSize = direction === 'h' ? target.offsetHeight : target.offsetWidth;
    document.body.style.cursor     = direction === 'h' ? 'row-resize' : 'col-resize';
    document.body.style.userSelect = 'none';
  });

  document.addEventListener('mousemove', e => {
    if (!dragging) return;
    const delta   = (direction === 'h' ? e.clientY : e.clientX) - startPos;
    let newSize;
    if (direction === 'h') {
      newSize = clamp(startSize - delta, 60, window.innerHeight * 0.6);
    } else if (targetId === 'right-panel') {
      newSize = clamp(startSize - delta, 100, 500);
    } else {
      newSize = clamp(startSize + delta, 100, 500);
    }
    target.style[direction === 'h' ? 'height' : 'width'] = newSize + 'px';
    // trigger canvas resize
    window.dispatchEvent(new Event('resize'));
  });

  document.addEventListener('mouseup', () => {
    if (!dragging) return;
    dragging = false;
    document.body.style.cursor     = '';
    document.body.style.userSelect = '';
    save();
  });
}

export function toggleSection(sectionEl) {
  if (!sectionEl) return;
  const content = sectionEl.querySelector('.panel-content');
  const btn     = sectionEl.querySelector('.collapse-btn');
  if (!content) return;
  const collapsed = content.style.display === 'none';
  content.style.display = collapsed ? '' : 'none';
  if (btn) btn.textContent = collapsed ? '▾' : '▸';
}

export function initPanels() {
  load();
  makeSplitter('vsplitter-left',   'left-panel',    'v');
  makeSplitter('vsplitter-right',  'right-panel',   'v');
  makeSplitter('hsplitter-bottom', 'bottom-panel',  'h');

  // Collapse/expand on panel-header click
  document.querySelectorAll('.panel-section > .panel-header').forEach(header => {
    header.style.cursor = 'pointer';
    header.addEventListener('click', e => {
      // Don't collapse if clicking a child button inside the header
      if (e.target.tagName === 'BUTTON' && !e.target.classList.contains('collapse-btn')) return;
      const section = header.closest('.panel-section');
      if (section) toggleSection(section);
    });
  });

  // Bottom panel toggle
  const bottomToggle = document.getElementById('btn-toggle-bottom');
  if (bottomToggle) {
    bottomToggle.addEventListener('click', () => {
      const bp = document.getElementById('bottom-panel');
      if (!bp) return;
      const hidden = bp.style.display === 'none';
      bp.style.display = hidden ? '' : 'none';
      document.getElementById('hsplitter-bottom').style.display = hidden ? '' : 'none';
      bottomToggle.textContent = hidden ? '▾' : '▸';
    });
  }
}
