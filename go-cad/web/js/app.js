// web/js/app.js — Main entry point (ES module)
import { state, setStatus, invalidateBlockCache } from './state.js';
import { initWasm, installDemoStubs }             from './wasm.js';
import { setupCanvas, render, zoomFit }           from './canvas.js';
import { initSnapToolbar, loadSnapSettings, updateSnapBtn } from './snap.js';
import { setTool, initTools, tryDelete, commitEntity } from './tools.js';
import { initCommandBar, processCommand }          from './commands.js';
import { initPanels }                              from './panels.js';
import { initInspector }                           from './inspector.js';
import { initHistory, setReplayCallback }           from './history.js';
import { initWelcome }                             from './welcome.js';
import { initDialogs, refreshLayers, openLayerManager,
         openBlockManager, openSymbolsPanel,
         openDraftingSettings, openPrintPlot }    from './dialogs.js';

// ── Bootstrap ──────────────────────────────────────────────────────────────────
(async () => {
  // 1. Panel layout
  initPanels();

  // 2. Canvas + mouse handlers
  setupCanvas();
  initTools();

  // 3. Snap toolbar
  initSnapToolbar();

  // 4. Inspector, history, welcome
  initInspector();
  initHistory();
  initWelcome();

  // 5. Dialogs
  initDialogs();

  // 6. Command bar
  initCommandBar();
  setReplayCallback(processCommand);

  // 7. WASM
  try {
    await initWasm();
    setStatus('Ready — select a tool and click to draw');
  } catch (_) {
    installDemoStubs();
    state.wasmReady = true;
    setStatus('Demo mode — serve wasm_exec.js + main.wasm for full WASM mode');
  }

  // 8. Post-WASM init
  loadSnapSettings();
  updateSnapBtn();
  refreshLayers();
  render();
  setTool('line');

  // 9. Toolbar buttons
  ['line','circle','arc','rect','poly','spline','nurbs','ellipse','text','mtext',
   'dimlin','dimali','dimang','dimrad','dimdia',
   'move','copy','rotate','scale','mirror','trim','extend','fillet','chamfer','arrayrect','arraypolar','offset',
   'hatch','leader','revcloud','wipeout',
  ].forEach(t => {
    document.getElementById('tool-' + t)?.addEventListener('click', () => setTool(t));
  });

  document.getElementById('btn-undo')?.addEventListener('click', () => {
    if (state.wasmReady) window.cadUndo(); render();
  });
  document.getElementById('btn-redo')?.addEventListener('click', () => {
    if (state.wasmReady) window.cadRedo(); render();
  });
  document.getElementById('btn-del')?.addEventListener('click', () => {
    if (state.wasmReady && state.selectedId) {
      tryDelete(state.selectedId); state.selectedId = 0; render();
    }
  });
  document.getElementById('btn-clear')?.addEventListener('click', () => {
    if (state.wasmReady) window.cadClear();
    invalidateBlockCache();
    render();
  });
  document.getElementById('btn-zoomfit')?.addEventListener('click', zoomFit);
  document.getElementById('btn-export')?.addEventListener('click', () => {
    processCommand('EXPORT');
  });
  document.getElementById('btn-open')?.addEventListener('click', () => {
    document.getElementById('file-input')?.click();
  });
  document.getElementById('btn-print')?.addEventListener('click', openPrintPlot);
  document.getElementById('btn-drafting')?.addEventListener('click', openDraftingSettings);

  // DXF file import
  document.getElementById('file-input')?.addEventListener('change', function (ev) {
    const file = ev.target.files[0];
    if (!file) return;
    if (!state.wasmReady) { setStatus('WASM not ready yet — try again.'); return; }
    const reader = new FileReader();
    reader.onload = e => {
      let result;
      try { result = JSON.parse(window.cadLoadDXF(e.target.result)); }
      catch (err) { setStatus('DXF import failed: ' + err.message); return; }
      if (!result.ok) { setStatus('DXF import error: ' + result.error); return; }
      refreshLayers();
      invalidateBlockCache();
      render();
      zoomFit();
      const warnTxt = result.warnings?.length
        ? ` (${result.warnings.length} warning${result.warnings.length > 1 ? 's' : ''})` : '';
      setStatus(`Loaded ${file.name}: ${result.count} entities${warnTxt}`);
    };
    reader.onerror = () => setStatus('Failed to read file.');
    reader.readAsText(file);
    ev.target.value = '';
  });

  // ── Keyboard shortcuts ───────────────────────────────────────────────────────
  document.addEventListener('keydown', e => {
    const cmdInput = document.getElementById('cmd-input');
    const coordInput = document.getElementById('coord-input');
    if (document.activeElement === cmdInput || document.activeElement === coordInput) return;

    // Ctrl / Meta shortcuts
    if (e.ctrlKey || e.metaKey) {
      if (e.key === 'z') { e.preventDefault(); if (state.wasmReady) window.cadUndo(); render(); }
      if (e.key === 'y' || e.key === 'Y') { e.preventDefault(); if (state.wasmReady) window.cadRedo(); render(); }
      if (e.key === 'o' || e.key === 'O') { e.preventDefault(); document.getElementById('file-input')?.click(); }
      if (e.key === 'p' || e.key === 'P') { e.preventDefault(); openPrintPlot(); }
      return;
    }

    // F3: toggle Object Snap
    if (e.key === 'F3') {
      e.preventDefault();
      state.snapEnabled = !state.snapEnabled;
      state.snapResult  = null;
      updateSnapBtn();
      setStatus(state.snapEnabled ? 'Object Snap ON' : 'Object Snap OFF');
      return;
    }

    // Single-key tool shortcuts
    const map = { l:'line', c:'circle', a:'arc', r:'rect', p:'poly',
                  s:'spline', n:'nurbs', e:'ellipse', t:'text' };
    if (map[e.key?.toLowerCase()]) { setTool(map[e.key.toLowerCase()]); return; }

    if (e.key === 'Delete' && state.wasmReady && state.selectedId) {
      tryDelete(state.selectedId); state.selectedId = 0; render();
    }
    if (e.key === 'Escape') {
      state.clicks = []; state.snapResult = null;
      import('./snap.js').then(m => m.drawSnapMarker());
      render();
    }
    if (e.key === 'Enter' &&
        ['poly','spline','nurbs','hatch','revcloud','wipeout'].includes(state.currentTool)) {
      commitEntity();
    }
    if (e.key === 'Enter' && state.currentTool === 'leader' && state.clicks.length >= 2) {
      commitEntity();
    }
    // Focus command bar on any printable key
    if (e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
      if (cmdInput) {
        cmdInput.focus();
        // Don't preventDefault — let the key land in the input
      }
    }
  });

  // ── Snap popover close-on-outside-click ──────────────────────────────────────
  document.addEventListener('click', e => {
    const panel = document.getElementById('snap-settings');
    if (panel?.classList.contains('open') &&
        !panel.contains(e.target) &&
        e.target.id !== 'btn-snap-settings') {
      panel.classList.remove('open');
    }
  });

})();
