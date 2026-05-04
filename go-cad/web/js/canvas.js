// web/js/canvas.js — Canvas rendering, transforms, and pan/zoom
import { state, getBlockEntities, w2s, s2w } from './state.js';
import { drawSnapMarker } from './snap.js';
import { showEntityProperties, clearInspector } from './inspector.js';

// Re-export so other modules can still `import { w2s, s2w } from './canvas.js'`
export { w2s, s2w };

let canvas, ctx;

export function setupCanvas() {
  canvas = document.getElementById('canvas');
  ctx    = canvas.getContext('2d');

  // Fit canvas to viewport using ResizeObserver
  const viewport = document.getElementById('viewport');
  const ro = new ResizeObserver(() => resize());
  ro.observe(viewport);
  resize();
}

export function getCanvas() { return canvas; }
export function getCtx()    { return ctx; }

export function resize() {
  if (!canvas) return;
  const vp = document.getElementById('viewport');
  if (!vp) return;
  canvas.width  = vp.clientWidth;
  canvas.height = vp.clientHeight;
  render();
}

// ── Main render ────────────────────────────────────────────────────────────────
export function render() {
  if (!canvas || !ctx) return;
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  drawGrid();

  if (!state.wasmReady) {
    ctx.fillStyle = '#666';
    ctx.font      = '14px sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('Loading WASM…', canvas.width / 2, canvas.height / 2);
    return;
  }

  // Build layer cache
  const hiddenLayers = new Set();
  window._layerCache   = {};
  window._lockedLayers = new Set();
  try {
    const layers = JSON.parse(window.cadGetLayers() || '[]');
    layers.forEach(l => {
      window._layerCache[l.id] = l;
      if (!l.visible || l.frozen) hiddenLayers.add(l.id);
      if (l.locked) window._lockedLayers.add(l.id);
    });
  } catch (_) {}

  const entities = JSON.parse(window.cadEntities() || '[]');
  entities.forEach(e => {
    if (hiddenLayers.has(e.layer)) return;
    drawEntity(e, e.id === state.selectedId);
  });

  drawPreview();

  if (state.tempPt) {
    const [sx, sy] = w2s(state.tempPt[0], state.tempPt[1]);
    ctx.save();
    ctx.strokeStyle = 'rgba(255,255,255,0.15)';
    ctx.lineWidth   = 1;
    ctx.setLineDash([4, 4]);
    ctx.beginPath(); ctx.moveTo(sx, 0);       ctx.lineTo(sx, canvas.height); ctx.stroke();
    ctx.beginPath(); ctx.moveTo(0,  sy);      ctx.lineTo(canvas.width, sy);  ctx.stroke();
    ctx.restore();
  }

  drawSnapMarker();
}

function drawGrid() {
  const step = Math.max(10, Math.round(50 / state.zoom) * 10);
  const [ox, oy] = w2s(0, 0);
  ctx.save();
  ctx.lineWidth = 1;

  ctx.strokeStyle = 'rgba(255,255,255,0.05)';
  const sX = Math.floor((-state.panX) / (state.zoom * step)) * step;
  for (let wx = sX; wx < sX + canvas.width / state.zoom + step; wx += step) {
    const [sx] = w2s(wx, 0);
    ctx.beginPath(); ctx.moveTo(sx, 0); ctx.lineTo(sx, canvas.height); ctx.stroke();
  }
  const sY = Math.floor((state.panY - canvas.height) / (state.zoom * step)) * step;
  for (let wy = sY; wy < sY + canvas.height / state.zoom + step; wy += step) {
    const [, sy] = w2s(0, wy);
    ctx.beginPath(); ctx.moveTo(0, sy); ctx.lineTo(canvas.width, sy); ctx.stroke();
  }

  ctx.strokeStyle = 'rgba(255,255,255,0.12)';
  ctx.beginPath(); ctx.moveTo(0, oy); ctx.lineTo(canvas.width, oy);  ctx.stroke();
  ctx.beginPath(); ctx.moveTo(ox, 0); ctx.lineTo(ox, canvas.height); ctx.stroke();
  ctx.restore();
}

// ── Layer linetype dash patterns ───────────────────────────────────────────────
const LAYER_DASH = {
  Solid:   [],
  Dashed:  [8, 4],
  Dotted:  [1, 4],
  DashDot: [8, 4, 1, 4],
  Center:  [16, 4, 4, 4],
  Hidden:  [4, 4],
};

function resolveEntityStyle(e, sel) {
  if (sel) return { color: '#ff9800', dash: [] };
  let color = e.color || '';
  let dash  = [];
  const lyr = window._layerCache ? (window._layerCache[e.layer || 0] || null) : null;
  if (lyr) dash = LAYER_DASH[lyr.lineType] || [];
  if (!color || color.toUpperCase() === 'BYLAYER') color = lyr ? lyr.color : '#ffffff';
  if (!color) color = '#ffffff';
  return { color, dash };
}

function drawEntity(e, sel = false) {
  const { color: col, dash } = resolveEntityStyle(e, sel);
  ctx.save();
  ctx.strokeStyle = col;
  ctx.fillStyle   = col;
  ctx.lineWidth   = sel ? 2 : 1;
  if (dash && dash.length) ctx.setLineDash(dash);
  else ctx.setLineDash([]);

  switch (e.type) {
    case 'line': {
      const [x1,y1] = w2s(e.x1,e.y1), [x2,y2] = w2s(e.x2,e.y2);
      ctx.beginPath(); ctx.moveTo(x1,y1); ctx.lineTo(x2,y2); ctx.stroke(); break;
    }
    case 'circle': {
      const [cx,cy] = w2s(e.cx,e.cy);
      ctx.beginPath(); ctx.arc(cx,cy,e.r*state.zoom,0,2*Math.PI); ctx.stroke(); break;
    }
    case 'arc': {
      const [cx,cy] = w2s(e.cx,e.cy);
      const a1 = -e.startDeg*Math.PI/180, a2 = -e.endDeg*Math.PI/180;
      ctx.beginPath(); ctx.arc(cx,cy,e.r*state.zoom,a1,a2,a1>a2); ctx.stroke(); break;
    }
    case 'rectangle': {
      const [x1,y1] = w2s(e.x1,e.y1), [x2,y2] = w2s(e.x2,e.y2);
      ctx.beginPath();
      ctx.rect(Math.min(x1,x2),Math.min(y1,y2),Math.abs(x2-x1),Math.abs(y2-y1));
      ctx.stroke(); break;
    }
    case 'polyline':
      drawPolylineScrPts(ctx, (e.points||[]).map(p => w2s(p[0],p[1]))); break;
    case 'spline':
      drawBezierSpline(ctx, (e.points||[]).map(p => w2s(p[0],p[1]))); break;
    case 'nurbs':
      drawNURBSEntity(ctx, e); break;
    case 'ellipse': {
      const [cx,cy] = w2s(e.cx,e.cy);
      ctx.beginPath();
      ctx.ellipse(cx,cy,(e.r||1)*state.zoom,(e.r2||e.r||1)*state.zoom,-(e.rotDeg||0)*Math.PI/180,0,2*Math.PI);
      ctx.stroke(); break;
    }
    case 'text': {
      const [sx,sy] = w2s(e.x1,e.y1);
      const h = Math.max(8,(e.textHeight||2.5)*state.zoom);
      ctx.save(); ctx.translate(sx,sy); ctx.rotate(-(e.rotDeg||0)*Math.PI/180);
      ctx.font = `${h}px ${e.font||'monospace'}`; ctx.fillText(e.text||'',0,0);
      ctx.restore(); break;
    }
    case 'mtext': {
      const [sx,sy] = w2s(e.x1,e.y1);
      const h = Math.max(8,(e.textHeight||3)*state.zoom);
      const lines = (e.text||'').split('\n');
      ctx.save(); ctx.translate(sx,sy); ctx.rotate(-(e.rotDeg||0)*Math.PI/180);
      ctx.font = `${h}px ${e.font||'monospace'}`;
      lines.forEach((ln,i) => ctx.fillText(ln,0,i*h*1.5));
      ctx.restore(); break;
    }
    case 'dimlin':  case 'dimali': drawLinearDim(ctx,e,col); break;
    case 'dimang':                  drawAngularDim(ctx,e,col); break;
    case 'dimrad':                  drawRadialDim(ctx,e,col,false); break;
    case 'dimdia':                  drawRadialDim(ctx,e,col,true);  break;
    case 'blockref': drawBlockRef(ctx,e,col,sel); break;
    case 'hatch':    drawHatch(ctx,e,col);   break;
    case 'leader':   drawLeader(ctx,e,col);  break;
    case 'revcloud': drawRevCloud(ctx,e,col); break;
    case 'wipeout':  drawWipeout(ctx,e,col); break;
  }
  ctx.restore();
}

// ── Block reference renderer ───────────────────────────────────────────────────
function drawBlockRef(c2, e, col, sel) {
  const name = e.text || '';
  const ix = e.x1, iy = e.y1;
  const sx = e.r || 1, sy2 = e.r2 || 1;
  const rot = (e.rotDeg || 0) * Math.PI / 180;
  const cosR = Math.cos(rot), sinR = Math.sin(rot);

  function localToWorld(lx, ly) {
    return [ix + lx*sx*cosR - ly*sy2*sinR, iy + lx*sx*sinR + ly*sy2*cosR];
  }

  const ents = getBlockEntities(name);
  const strokeCol = sel ? '#ff9800' : col;

  if (!ents || ents.length === 0) {
    const [scx, scy] = w2s(ix, iy);
    const sz = 10;
    c2.save(); c2.strokeStyle = strokeCol; c2.lineWidth = sel ? 2 : 1;
    c2.beginPath(); c2.moveTo(scx-sz,scy); c2.lineTo(scx+sz,scy); c2.stroke();
    c2.beginPath(); c2.moveTo(scx,scy-sz); c2.lineTo(scx,scy+sz); c2.stroke();
    c2.fillStyle = strokeCol; c2.font='9px monospace'; c2.textAlign='left';
    c2.fillText(name||'?', scx+sz+2, scy-2);
    c2.restore(); return;
  }

  c2.save(); c2.strokeStyle = strokeCol; c2.fillStyle = strokeCol;
  c2.lineWidth = sel ? 2 : 1; c2.setLineDash([]);

  ents.forEach(be => {
    switch(be.type) {
      case 'line': {
        const [ax,ay] = w2s(...localToWorld(be.x1,be.y1));
        const [bx,by] = w2s(...localToWorld(be.x2,be.y2));
        c2.beginPath(); c2.moveTo(ax,ay); c2.lineTo(bx,by); c2.stroke(); break;
      }
      case 'circle': {
        const [cx2,cy2] = w2s(...localToWorld(be.cx,be.cy));
        c2.beginPath(); c2.arc(cx2,cy2,Math.abs(be.r*sx*state.zoom),0,2*Math.PI); c2.stroke(); break;
      }
      case 'arc': {
        const [cx2,cy2] = w2s(...localToWorld(be.cx,be.cy));
        const sa = -(be.startDeg+(e.rotDeg||0))*Math.PI/180;
        const ea = -(be.endDeg+(e.rotDeg||0))*Math.PI/180;
        c2.beginPath(); c2.arc(cx2,cy2,Math.abs(be.r*sx*state.zoom),sa,ea); c2.stroke(); break;
      }
      case 'text': {
        const [tx,ty] = w2s(...localToWorld(be.x1,be.y1));
        c2.font=`${Math.max(7,(be.textHeight||2.5)*state.zoom)}px sans-serif`;
        c2.textAlign='left'; c2.fillText(be.text||'',tx,ty); break;
      }
      case 'polyline': case 'polygon': {
        const pts2 = be.points||[];
        if (pts2.length<2) break;
        c2.beginPath();
        const [fx,fy]=w2s(...localToWorld(pts2[0][0],pts2[0][1])); c2.moveTo(fx,fy);
        for (let j=1;j<pts2.length;j++) {
          const [px,py]=w2s(...localToWorld(pts2[j][0],pts2[j][1])); c2.lineTo(px,py);
        }
        if (be.type==='polygon') c2.closePath();
        c2.stroke(); break;
      }
    }
  });

  const [mx,my] = w2s(ix,iy);
  c2.globalAlpha=0.4; c2.lineWidth=0.7;
  c2.beginPath(); c2.moveTo(mx-5,my); c2.lineTo(mx+5,my); c2.stroke();
  c2.beginPath(); c2.moveTo(mx,my-5); c2.lineTo(mx,my+5); c2.stroke();
  c2.globalAlpha=1;
  c2.restore();
}

function drawHatch(c2, e, col) {
  const pts = e.points || [];
  if (pts.length < 3) return;
  const solid = (e.text||'').toUpperCase() === 'SOLID';
  if (solid) {
    c2.save(); c2.fillStyle=col; c2.globalAlpha=0.35;
    c2.beginPath();
    const [fx,fy]=w2s(pts[0][0],pts[0][1]); c2.moveTo(fx,fy);
    for (let i=1;i<pts.length;i++){const[px,py]=w2s(pts[i][0],pts[i][1]);c2.lineTo(px,py);}
    c2.closePath(); c2.fill(); c2.restore();
  } else {
    c2.save(); c2.strokeStyle=col; c2.lineWidth=0.8;
    c2.beginPath();
    const [fx,fy]=w2s(pts[0][0],pts[0][1]); c2.moveTo(fx,fy);
    for (let i=1;i<pts.length;i++){const[px,py]=w2s(pts[i][0],pts[i][1]);c2.lineTo(px,py);}
    c2.closePath(); c2.stroke();
    if (state.wasmReady && window.cadRenderHatch) {
      try {
        const segs = JSON.parse(window.cadRenderHatch(e.id)||'[]');
        segs.forEach(s=>{
          const [sx1,sy1]=w2s(s[0],s[1]),[sx2,sy2]=w2s(s[2],s[3]);
          c2.beginPath();c2.moveTo(sx1,sy1);c2.lineTo(sx2,sy2);c2.stroke();
        });
      } catch(_){}
    } else {
      drawHatchFallback(c2, pts, col);
    }
    c2.restore();
  }
}

function drawHatchFallback(c2, pts, col) {
  let minX=Infinity,minY=Infinity,maxX=-Infinity,maxY=-Infinity;
  pts.forEach(p=>{const[sx,sy]=w2s(p[0],p[1]);
    if(sx<minX)minX=sx;if(sy<minY)minY=sy;if(sx>maxX)maxX=sx;if(sy>maxY)maxY=sy;});
  c2.save(); c2.strokeStyle=col; c2.lineWidth=0.7;
  c2.beginPath();
  const[fx,fy]=w2s(pts[0][0],pts[0][1]);c2.moveTo(fx,fy);
  for(let i=1;i<pts.length;i++){const[px,py]=w2s(pts[i][0],pts[i][1]);c2.lineTo(px,py);}
  c2.closePath(); c2.clip();
  const spacing=10;
  for(let t=minX-maxY+minY;t<maxX+maxY-minY;t+=spacing){
    c2.beginPath();c2.moveTo(t,minY-spacing);c2.lineTo(t+maxY-minY+spacing,maxY+spacing);c2.stroke();
  }
  c2.restore();
}

function drawLeader(c2, e, col) {
  const pts = e.points || [];
  if (pts.length < 2) return;
  c2.save(); c2.strokeStyle=col; c2.fillStyle=col; c2.lineWidth=1; c2.setLineDash([]);
  c2.beginPath();
  const [fx,fy]=w2s(pts[0][0],pts[0][1]); c2.moveTo(fx,fy);
  for (let i=1;i<pts.length;i++){const[px,py]=w2s(pts[i][0],pts[i][1]);c2.lineTo(px,py);}
  c2.stroke();
  const [p0x,p0y]=w2s(pts[0][0],pts[0][1]);
  const [p1x,p1y]=w2s(pts[1][0],pts[1][1]);
  screenArrow(c2,p0x,p0y,p0x-p1x,p0y-p1y);
  if (e.text) {
    const last=pts[pts.length-1];
    const [lx,ly]=w2s(last[0],last[1]);
    c2.font='11px sans-serif'; c2.textAlign='left';
    c2.fillText(e.text,lx+4,ly-4);
  }
  c2.restore();
}

function drawRevCloud(c2, e, col) {
  const pts = e.points || [];
  if (pts.length < 3) return;
  c2.save(); c2.strokeStyle=col; c2.lineWidth=1; c2.setLineDash([]);
  const all = [...pts, pts[0]];
  for (let i=0;i<all.length-1;i++){
    const [ax,ay]=w2s(all[i][0],all[i][1]);
    const [bx,by]=w2s(all[i+1][0],all[i+1][1]);
    const mx=(ax+bx)/2, my=(ay+by)/2;
    const len=Math.hypot(bx-ax,by-ay)||1;
    const nx=-(by-ay)/len, ny=(bx-ax)/len;
    const bump=6;
    c2.beginPath(); c2.moveTo(ax,ay);
    c2.quadraticCurveTo(mx+nx*bump,my+ny*bump,bx,by); c2.stroke();
  }
  c2.restore();
}

function drawWipeout(c2, e, col) {
  const pts = e.points || [];
  if (pts.length < 3) return;
  c2.save();
  c2.fillStyle='#1e1e1e'; c2.strokeStyle=col||'#555';
  c2.lineWidth=1; c2.setLineDash([4,3]);
  c2.beginPath();
  const [fx,fy]=w2s(pts[0][0],pts[0][1]); c2.moveTo(fx,fy);
  for (let i=1;i<pts.length;i++){const[px,py]=w2s(pts[i][0],pts[i][1]);c2.lineTo(px,py);}
  c2.closePath(); c2.fill(); c2.stroke();
  c2.restore();
}

// ── Spline renderers ───────────────────────────────────────────────────────────
function drawPolylineScrPts(c2, pts) {
  if (pts.length < 2) return;
  c2.beginPath(); c2.moveTo(pts[0][0],pts[0][1]);
  for (let i=1;i<pts.length;i++) c2.lineTo(pts[i][0],pts[i][1]);
  c2.stroke();
}
export { drawPolylineScrPts };

function drawBezierSpline(c2, pts) {
  if (pts.length < 4) { drawPolylineScrPts(c2, pts); return; }
  const nSegs = Math.floor((pts.length-1)/3);
  c2.beginPath(); c2.moveTo(pts[0][0],pts[0][1]);
  for (let i=0;i<nSegs;i++){
    const b=i*3;
    c2.bezierCurveTo(pts[b+1][0],pts[b+1][1],pts[b+2][0],pts[b+2][1],pts[b+3][0],pts[b+3][1]);
  }
  c2.stroke();
}
export { drawBezierSpline };

// ── NURBS evaluator ────────────────────────────────────────────────────────────
function nurbsBasis(i,k,knots,t){
  if(k===0) return knots[i]<=t&&t<knots[i+1]?1:0;
  const d1=knots[i+k]-knots[i],d2=knots[i+k+1]-knots[i+1];
  let l=0,r=0;
  if(d1>1e-10) l=(t-knots[i])/d1*nurbsBasis(i,k-1,knots,t);
  if(d2>1e-10) r=(knots[i+k+1]-t)/d2*nurbsBasis(i+1,k-1,knots,t);
  return l+r;
}
function nurbsEval(deg,controls,knots,weights,t){
  let wx=0,wy=0,w=0;
  for(let i=0;i<controls.length;i++){
    const b=nurbsBasis(i,deg,knots,t);
    const wi=weights?weights[i]:1;
    const bw=b*wi;
    wx+=bw*controls[i][0];wy+=bw*controls[i][1];w+=bw;
  }
  if(w<1e-10) return controls[0]||[0,0];
  return [wx/w,wy/w];
}
export { nurbsEval };

export function clampedUniformKnots(n, deg) {
  const m=n+deg+1, knots=new Array(m).fill(0);
  const inner=n-deg;
  for(let i=0;i<=deg;i++) knots[i]=0;
  for(let i=1;i<inner;i++) knots[deg+i]=i/inner;
  for(let i=n;i<m;i++) knots[i]=1;
  return knots;
}

function drawNURBSEntity(c2, e) {
  const controls=e.points||[];
  const n=controls.length;
  if(n<2) return;
  const deg=e.nurbsDeg||3;
  const knots=e.knots&&e.knots.length>=n+deg+1?e.knots:clampedUniformKnots(n,deg);
  const weights=e.weights;
  const lo=knots[deg],hi=knots[n];
  const samples=Math.max(50,n*15);
  c2.beginPath();
  for(let k=0;k<=samples;k++){
    let t=lo+(k/samples)*(hi-lo);
    if(t>=hi) t=hi-1e-10;
    const[wx,wy]=nurbsEval(deg,controls,knots,weights,t);
    const[sx,sy]=w2s(wx,wy);
    if(k===0) c2.moveTo(sx,sy); else c2.lineTo(sx,sy);
  }
  c2.stroke();
}

// ── Dimension renderers ────────────────────────────────────────────────────────
const ARROW = 8;

function screenArrow(c2, px, py, dx, dy) {
  const len=Math.hypot(dx,dy); if(len<0.01) return;
  const ux=dx/len,uy=dy/len,sz=ARROW;
  c2.beginPath();
  c2.moveTo(px,py);
  c2.lineTo(px-ux*sz+uy*sz*0.4,py-uy*sz-ux*sz*0.4);
  c2.lineTo(px-ux*sz-uy*sz*0.4,py-uy*sz+ux*sz*0.4);
  c2.closePath(); c2.fill();
}

function drawLinearDim(c2, e, col) {
  const dx=e.x2-e.x1,dy=e.y2-e.y1;
  const isH=Math.abs(dx)>=Math.abs(dy);
  const off=e.cx||30;
  let d1x,d1y,d2x,d2y;
  if(e.type==='dimali'){
    const len=Math.hypot(dx,dy)||1;
    const ux=dx/len,uy=dy/len,px=-uy,py=ux;
    d1x=e.x1+px*off;d1y=e.y1+py*off;d2x=e.x2+px*off;d2y=e.y2+py*off;
  } else {
    if(isH){const dY=(e.y1+e.y2)/2-off;d1x=e.x1;d1y=dY;d2x=e.x2;d2y=dY;}
    else   {const dX=(e.x1+e.x2)/2-off;d1x=dX;d1y=e.y1;d2x=dX;d2y=e.y2;}
  }
  const[sx1,sy1]=w2s(e.x1,e.y1),[sx2,sy2]=w2s(e.x2,e.y2);
  const[sd1x,sd1y]=w2s(d1x,d1y),[sd2x,sd2y]=w2s(d2x,d2y);
  c2.save();c2.strokeStyle=col;c2.fillStyle=col;c2.lineWidth=1;
  c2.beginPath();c2.moveTo(sx1,sy1);c2.lineTo(sd1x,sd1y);c2.stroke();
  c2.beginPath();c2.moveTo(sx2,sy2);c2.lineTo(sd2x,sd2y);c2.stroke();
  c2.beginPath();c2.moveTo(sd1x,sd1y);c2.lineTo(sd2x,sd2y);c2.stroke();
  screenArrow(c2,sd1x,sd1y,sd2x-sd1x,sd2y-sd1y);
  screenArrow(c2,sd2x,sd2y,sd1x-sd2x,sd1y-sd2y);
  const val=Math.hypot(e.x2-e.x1,e.y2-e.y1).toFixed(2);
  c2.font='11px sans-serif';c2.textAlign='center';
  c2.fillText(val,(sd1x+sd2x)/2,(sd1y+sd2y)/2-5);
  c2.restore();
}

function drawAngularDim(c2, e, col) {
  const ang1=Math.atan2(e.y1-e.cy,e.x1-e.cx);
  const ang2=Math.atan2(e.y2-e.cy,e.x2-e.cx);
  let span=ang2-ang1; if(span<0)span+=2*Math.PI;
  const r=(e.r||30)*state.zoom;
  const[csx,csy]=w2s(e.cx,e.cy);
  c2.save();c2.strokeStyle=col;c2.fillStyle=col;c2.lineWidth=1;
  c2.beginPath();c2.arc(csx,csy,r,-ang1,-(ang1+span),span>Math.PI);c2.stroke();
  c2.beginPath();c2.moveTo(csx,csy);c2.lineTo(csx+Math.cos(ang1)*r,csy-Math.sin(ang1)*r);c2.stroke();
  c2.beginPath();c2.moveTo(csx,csy);c2.lineTo(csx+Math.cos(ang2)*r,csy-Math.sin(ang2)*r);c2.stroke();
  const midAng=ang1+span/2;
  c2.font='11px sans-serif';c2.textAlign='center';
  c2.fillText(`${(span*180/Math.PI).toFixed(1)}°`,csx+Math.cos(midAng)*r*1.3,csy-Math.sin(midAng)*r*1.3);
  c2.restore();
}

function drawRadialDim(c2, e, col, diameter) {
  const ang=(e.rotDeg||0)*Math.PI/180;
  const[csx,csy]=w2s(e.cx,e.cy);
  const rx=csx+Math.cos(ang)*e.r*state.zoom,ry=csy-Math.sin(ang)*e.r*state.zoom;
  c2.save();c2.strokeStyle=col;c2.fillStyle=col;c2.lineWidth=1;
  if(diameter){
    const rx2=csx-Math.cos(ang)*e.r*state.zoom,ry2=csy+Math.sin(ang)*e.r*state.zoom;
    c2.beginPath();c2.moveTo(rx,ry);c2.lineTo(rx2,ry2);c2.stroke();
    screenArrow(c2,rx,ry,rx-rx2,ry-ry2);screenArrow(c2,rx2,ry2,rx2-rx,ry2-ry);
    c2.font='11px sans-serif';c2.textAlign='center';
    c2.fillText(`⌀${(2*e.r).toFixed(2)}`,csx+Math.cos(ang)*e.r*state.zoom*1.5,csy-Math.sin(ang)*e.r*state.zoom*1.5);
  } else {
    c2.beginPath();c2.moveTo(csx,csy);c2.lineTo(rx,ry);c2.stroke();
    screenArrow(c2,rx,ry,rx-csx,ry-csy);
    c2.font='11px sans-serif';c2.textAlign='center';
    c2.fillText(`R${e.r.toFixed(2)}`,csx+Math.cos(ang)*e.r*state.zoom*1.5,csy-Math.sin(ang)*e.r*state.zoom*1.5);
  }
  c2.restore();
}

// ── Rubber-band preview ────────────────────────────────────────────────────────
function drawPreview() {
  if (!state.tempPt || state.clicks.length === 0) return;
  ctx.save();
  ctx.strokeStyle = 'rgba(0,120,212,0.7)';
  ctx.fillStyle   = 'rgba(0,120,212,0.7)';
  ctx.lineWidth   = 1;
  ctx.setLineDash([4, 4]);
  const pt = w2s(state.tempPt[0], state.tempPt[1]);
  const c0 = state.clicks.length > 0 ? w2s(state.clicks[0][0], state.clicks[0][1]) : null;
  const c1 = state.clicks.length > 1 ? w2s(state.clicks[1][0], state.clicks[1][1]) : null;

  switch (state.currentTool) {
    case 'line':
      if (c0) { ctx.beginPath(); ctx.moveTo(c0[0],c0[1]); ctx.lineTo(pt[0],pt[1]); ctx.stroke(); } break;
    case 'circle':
      if (c0) { const r=Math.hypot(pt[0]-c0[0],pt[1]-c0[1]); ctx.beginPath(); ctx.arc(c0[0],c0[1],r,0,2*Math.PI); ctx.stroke(); } break;
    case 'arc':
      if (c0 && !c1) { const r=Math.hypot(pt[0]-c0[0],pt[1]-c0[1]); ctx.beginPath(); ctx.arc(c0[0],c0[1],r,0,2*Math.PI); ctx.stroke(); }
      else if (c0 && c1) { const r=Math.hypot(c1[0]-c0[0],c1[1]-c0[1]); ctx.beginPath(); ctx.arc(c0[0],c0[1],r,Math.atan2(c1[1]-c0[1],c1[0]-c0[0]),Math.atan2(pt[1]-c0[1],pt[0]-c0[0])); ctx.stroke(); } break;
    case 'rect':
      if (c0) { ctx.beginPath(); ctx.rect(Math.min(c0[0],pt[0]),Math.min(c0[1],pt[1]),Math.abs(pt[0]-c0[0]),Math.abs(pt[1]-c0[1])); ctx.stroke(); } break;
    case 'poly': case 'hatch': case 'revcloud': case 'wipeout': {
      const all=[...state.clicks.map(p=>w2s(p[0],p[1])),pt];
      drawPolylineScrPts(ctx,all);
      if (all.length>=3) {
        ctx.setLineDash([2,4]);
        ctx.beginPath(); ctx.moveTo(all[all.length-1][0],all[all.length-1][1]); ctx.lineTo(all[0][0],all[0][1]); ctx.stroke();
      }
      break;
    }
    case 'spline': {
      const all=[...state.clicks.map(p=>w2s(p[0],p[1])),pt];
      if(all.length>=4) drawBezierSpline(ctx,all); else drawPolylineScrPts(ctx,all); break;
    }
    case 'nurbs': {
      const all=[...state.clicks.map(p=>w2s(p[0],p[1])),pt];
      if(all.length>=4){
        const wPts=[...state.clicks,state.tempPt];
        const n=wPts.length,deg=Math.min(3,n-1);
        const knots=clampedUniformKnots(n,deg);
        const lo=knots[deg],hi=knots[n];
        const samples=Math.max(50,n*15);
        ctx.beginPath();
        for(let k=0;k<=samples;k++){
          let t=lo+(k/samples)*(hi-lo);if(t>=hi)t=hi-1e-10;
          const[wx,wy]=nurbsEval(deg,wPts,knots,null,t);
          const[sx,sy]=w2s(wx,wy);
          if(k===0) ctx.moveTo(sx,sy); else ctx.lineTo(sx,sy);
        }
        ctx.stroke();
      } else drawPolylineScrPts(ctx,all);
      break;
    }
    case 'ellipse':
      if (c0 && !c1) { const r=Math.hypot(pt[0]-c0[0],pt[1]-c0[1]); ctx.beginPath(); ctx.arc(c0[0],c0[1],r,0,2*Math.PI); ctx.stroke(); }
      else if (c0 && c1) { const a=Math.hypot(c1[0]-c0[0],c1[1]-c0[1]),b=Math.hypot(pt[0]-c0[0],pt[1]-c0[1]),rot=Math.atan2(c1[1]-c0[1],c1[0]-c0[0]); ctx.beginPath(); ctx.ellipse(c0[0],c0[1],a,b,rot,0,2*Math.PI); ctx.stroke(); } break;
    case 'dimlin': case 'dimali':
      if (c0) { ctx.beginPath(); ctx.moveTo(c0[0],c0[1]); ctx.lineTo(pt[0],pt[1]); ctx.stroke(); } break;
    case 'dimang':
      if (c0) { ctx.beginPath(); ctx.moveTo(c0[0],c0[1]); ctx.lineTo(pt[0],pt[1]); ctx.stroke(); } break;
    case 'dimrad': case 'dimdia':
      if (c0) { const r=Math.hypot(pt[0]-c0[0],pt[1]-c0[1]); ctx.beginPath(); ctx.arc(c0[0],c0[1],r,0,2*Math.PI); ctx.stroke(); ctx.beginPath(); ctx.moveTo(c0[0],c0[1]); ctx.lineTo(pt[0],pt[1]); ctx.stroke(); } break;
    case 'leader': {
      const all=[...state.clicks.map(p=>w2s(p[0],p[1])),pt];
      drawPolylineScrPts(ctx,all); break;
    }
  }

  // Click markers
  state.clicks.forEach(([wx, wy]) => {
    const [sx, sy] = w2s(wx, wy);
    ctx.setLineDash([]); ctx.beginPath(); ctx.arc(sx,sy,3,0,2*Math.PI); ctx.fill();
  });
  ctx.restore();
}

// ── Zoom to fit ────────────────────────────────────────────────────────────────
export function zoomFit() {
  if (!state.wasmReady || !canvas) return;
  const es = JSON.parse(window.cadEntities() || '[]');
  if (!es.length) { state.panX=0; state.panY=0; state.zoom=1; render(); return; }
  let minX=Infinity,minY=Infinity,maxX=-Infinity,maxY=-Infinity;
  es.forEach(e => entitySamplePoints(e).forEach(([x,y]) => {
    if(x<minX)minX=x;if(y<minY)minY=y;if(x>maxX)maxX=x;if(y>maxY)maxY=y;
  }));
  if (!isFinite(minX)) { state.panX=0; state.panY=0; state.zoom=1; render(); return; }
  const padX=(maxX-minX)*0.1+10, padY=(maxY-minY)*0.1+10;
  minX-=padX; maxX+=padX; minY-=padY; maxY+=padY;
  state.zoom = Math.min(canvas.width/(maxX-minX), canvas.height/(maxY-minY));
  state.panX  = canvas.width/2  - ((minX+maxX)/2)*state.zoom;
  state.panY  = canvas.height/2 + ((minY+maxY)/2)*state.zoom;
  render();
}

export function entitySamplePoints(e) {
  switch (e.type) {
    case 'line': {
      // Include midpoint and quarter-points so clicking anywhere on the line works
      const mx = (e.x1+e.x2)/2, my = (e.y1+e.y2)/2;
      return [[e.x1,e.y1],[e.x2,e.y2],[mx,my],
              [(e.x1+mx)/2,(e.y1+my)/2],[(e.x2+mx)/2,(e.y2+my)/2]];
    }
    case 'circle':
      return [[e.cx+e.r,e.cy],[e.cx-e.r,e.cy],[e.cx,e.cy+e.r],[e.cx,e.cy-e.r],[e.cx,e.cy]];
    case 'arc':
      return [[e.cx+e.r,e.cy],[e.cx-e.r,e.cy],[e.cx,e.cy+e.r],[e.cx,e.cy-e.r],[e.cx,e.cy]];
    case 'rectangle': {
      const mx2=(e.x1+e.x2)/2, my2=(e.y1+e.y2)/2;
      return [[e.x1,e.y1],[e.x2,e.y2],[e.x2,e.y1],[e.x1,e.y2],[mx2,my2]];
    }
    case 'polyline': case 'spline': case 'nurbs': return e.points||[];
    case 'ellipse':  return [[e.cx+(e.r||0),e.cy],[e.cx-(e.r||0),e.cy],[e.cx,e.cy+(e.r2||0)],[e.cx,e.cy-(e.r2||0)],[e.cx,e.cy]];
    case 'text': case 'mtext': return [[e.x1,e.y1]];
    case 'blockref': return [[e.x1,e.y1]];
    case 'hatch': case 'revcloud': case 'wipeout': return (e.points||[]).map(p=>[p[0],p[1]]);
    case 'leader': return (e.points||[]).map(p=>[p[0],p[1]]);
    default: return e.cx!=null?[[e.cx,e.cy]]:[[e.x1||0,e.y1||0],[e.x2||0,e.y2||0]];
  }
}
