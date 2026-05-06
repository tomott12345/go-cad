// web/js/wasm.js — WebAssembly loading and in-browser demo stubs
import { state } from './state.js';

export async function initWasm() {
  await new Promise((res, rej) => {
    const s = document.createElement('script');
    s.src = 'wasm_exec.js';
    s.onload = res;
    s.onerror = rej;
    document.head.appendChild(s);
  });
  const go = new Go(); // eslint-disable-line no-undef
  const result = await WebAssembly.instantiateStreaming(fetch('main.wasm'), go.importObject);
  go.run(result.instance);
  state.wasmReady = true;
}

export function installDemoStubs() {
  let nextId = 1;
  const ents = [];
  const mk = obj => { const e = { id: nextId++, ...obj }; ents.push(e); return e.id; };

  window.cadAddLine        = (x1,y1,x2,y2,l,c)      => mk({type:'line',x1,y1,x2,y2,layer:l,color:c});
  window.cadAddCircle      = (cx,cy,r,l,c)            => mk({type:'circle',cx,cy,r,layer:l,color:c});
  window.cadAddArc         = (cx,cy,r,s,e,l,c)        => mk({type:'arc',cx,cy,r,startDeg:s,endDeg:e,layer:l,color:c});
  window.cadAddRectangle   = (x1,y1,x2,y2,l,c)       => mk({type:'rectangle',x1,y1,x2,y2,layer:l,color:c});
  window.cadAddPolyline    = (pts,l,c)                => mk({type:'polyline',points:pts,layer:l,color:c});
  window.cadAddSpline      = (pts,l,c)                => mk({type:'spline',points:pts,layer:l,color:c});
  window.cadAddNURBS       = (deg,pts,kn,wt,l,c)      => mk({type:'nurbs',nurbsDeg:deg,points:pts,knots:kn,weights:wt,layer:l,color:c});
  window.cadAddEllipse     = (cx,cy,a,b,rot,l,c)      => mk({type:'ellipse',cx,cy,r:a,r2:b,rotDeg:rot,layer:l,color:c});
  window.cadAddText        = (x,y,t,h,rot,fn,l,c)     => mk({type:'text',x1:x,y1:y,text:t,textHeight:h,rotDeg:rot,font:fn,layer:l,color:c});
  window.cadAddMText       = (x,y,t,h,w,rot,fn,l,c)  => mk({type:'mtext',x1:x,y1:y,text:t,textHeight:h,r2:w,rotDeg:rot,font:fn,layer:l,color:c});
  window.cadAddLinearDim   = (x1,y1,x2,y2,off,l,c)   => mk({type:'dimlin',x1,y1,x2,y2,cx:off,layer:l,color:c});
  window.cadAddAlignedDim  = (x1,y1,x2,y2,off,l,c)   => mk({type:'dimali',x1,y1,x2,y2,cx:off,layer:l,color:c});
  window.cadAddAngularDim  = (cx,cy,x1,y1,x2,y2,r,l,c) => mk({type:'dimang',cx,cy,x1,y1,x2,y2,r,layer:l,color:c});
  window.cadAddRadialDim   = (cx,cy,r,ang,l,c)        => mk({type:'dimrad',cx,cy,r,rotDeg:ang,layer:l,color:c});
  window.cadAddDiameterDim = (cx,cy,r,ang,l,c)        => mk({type:'dimdia',cx,cy,r,rotDeg:ang,layer:l,color:c});
  window.cadAddHatch       = (pts,pat,ang,sc,l,c)     => mk({type:'hatch',points:pts,text:pat,layer:l,color:c});
  window.cadAddLeader      = (pts,txt,l,c)            => mk({type:'leader',points:pts,text:txt,layer:l,color:c});
  window.cadAddRevisionCloud = (pts,arcLen,l,c)       => mk({type:'revcloud',points:pts,layer:l,color:c});
  window.cadAddWipeout     = (pts,l,c)                => mk({type:'wipeout',points:pts,layer:l,color:c});
  window.cadDeleteEntity   = id => { const i=ents.findIndex(e=>e.id===id); if(i>=0){ents.splice(i,1);return true;} return false; };
  window.cadUndo           = () => { if(ents.length>0){ents.pop();return true;} return false; };
  window.cadRedo           = () => false;
  window.cadClear          = () => { ents.length=0; };
  window.cadEntities       = () => JSON.stringify(ents);
  window.cadExportDXF      = () => '  0\nSECTION\n  2\nENTITIES\n  0\nENDSEC\n  0\nEOF\n';
  window.cadExportDXFR12   = () => window.cadExportDXF();
  window.cadExportSVG      = () => '<svg xmlns="http://www.w3.org/2000/svg"></svg>';
  window.cadNearestEntity  = (x,y,r) => {
    let best=0,bestD=r||Infinity;
    ents.forEach(e=>{const d=Math.hypot((e.cx||e.x1||0)-x,(e.cy||e.y1||0)-y);if(d<bestD){bestD=d;best=e.id;}});
    return best;
  };
  window.cadMove = (ids,dx,dy) => {
    ids.forEach(id=>{const e=ents.find(e=>e.id===id);if(!e)return;
      if(e.x1!=null){e.x1+=dx;e.y1+=dy;}if(e.x2!=null){e.x2+=dx;e.y2+=dy;}
      if(e.cx!=null){e.cx+=dx;e.cy+=dy;}if(e.points)e.points=e.points.map(p=>[p[0]+dx,p[1]+dy]);});
    return true;
  };
  window.cadCopy = (ids,dx,dy) => ids.map(id=>{
    const e=ents.find(e=>e.id===id);if(!e)return -1;
    const ne=JSON.parse(JSON.stringify(e));ne.id=nextId++;
    if(ne.x1!=null){ne.x1+=dx;ne.y1+=dy;}if(ne.x2!=null){ne.x2+=dx;ne.y2+=dy;}
    if(ne.cx!=null){ne.cx+=dx;ne.cy+=dy;}if(ne.points)ne.points=ne.points.map(p=>[p[0]+dx,p[1]+dy]);
    ents.push(ne);return ne.id;
  });
  window.cadRotate = (ids,cx,cy,ang,mc) => {
    const r=ang*Math.PI/180,cs=Math.cos(r),sn=Math.sin(r);
    const rot=(x,y)=>[cx+(x-cx)*cs-(y-cy)*sn,cy+(x-cx)*sn+(y-cy)*cs];
    if(mc){return ids.map(id=>{const e=ents.find(e=>e.id===id);if(!e)return -1;
      const ne=JSON.parse(JSON.stringify(e));ne.id=nextId++;
      if(ne.x1!=null){[ne.x1,ne.y1]=rot(ne.x1,ne.y1);}if(ne.x2!=null){[ne.x2,ne.y2]=rot(ne.x2,ne.y2);}
      if(ne.cx!=null){[ne.cx,ne.cy]=rot(ne.cx,ne.cy);}ents.push(ne);return ne.id;});}
    ids.forEach(id=>{const e=ents.find(e=>e.id===id);if(!e)return;
      if(e.x1!=null){[e.x1,e.y1]=rot(e.x1,e.y1);}if(e.x2!=null){[e.x2,e.y2]=rot(e.x2,e.y2);}
      if(e.cx!=null){[e.cx,e.cy]=rot(e.cx,e.cy);}});return ids;
  };
  window.cadScale = (ids,cx,cy,sx,sy,mc) => {
    const sc=(x,y)=>[cx+(x-cx)*sx,cy+(y-cy)*sy];
    if(mc){return ids.map(id=>{const e=ents.find(e=>e.id===id);if(!e)return -1;
      const ne=JSON.parse(JSON.stringify(e));ne.id=nextId++;
      if(ne.x1!=null){[ne.x1,ne.y1]=sc(ne.x1,ne.y1);}if(ne.x2!=null){[ne.x2,ne.y2]=sc(ne.x2,ne.y2);}
      if(ne.cx!=null){[ne.cx,ne.cy]=sc(ne.cx,ne.cy);ne.r*=Math.sqrt(Math.abs(sx*sy));}ents.push(ne);return ne.id;});}
    ids.forEach(id=>{const e=ents.find(e=>e.id===id);if(!e)return;
      if(e.x1!=null){[e.x1,e.y1]=sc(e.x1,e.y1);}if(e.x2!=null){[e.x2,e.y2]=sc(e.x2,e.y2);}
      if(e.cx!=null){[e.cx,e.cy]=sc(e.cx,e.cy);e.r*=Math.sqrt(Math.abs(sx*sy));}});return ids;
  };
  window.cadMirror = (ids,ax,ay,bx,by,mc) => {
    const dx=bx-ax,dy=by-ay,l2=dx*dx+dy*dy;
    const mir=(x,y)=>{const t=((x-ax)*dx+(y-ay)*dy)/l2;return[2*(ax+t*dx)-x,2*(ay+t*dy)-y];};
    if(mc){return ids.map(id=>{const e=ents.find(e=>e.id===id);if(!e)return -1;
      const ne=JSON.parse(JSON.stringify(e));ne.id=nextId++;
      if(ne.x1!=null){[ne.x1,ne.y1]=mir(ne.x1,ne.y1);}if(ne.x2!=null){[ne.x2,ne.y2]=mir(ne.x2,ne.y2);}
      if(ne.cx!=null){[ne.cx,ne.cy]=mir(ne.cx,ne.cy);}ents.push(ne);return ne.id;});}
    ids.forEach(id=>{const e=ents.find(e=>e.id===id);if(!e)return;
      if(e.x1!=null){[e.x1,e.y1]=mir(e.x1,e.y1);}if(e.x2!=null){[e.x2,e.y2]=mir(e.x2,e.y2);}
      if(e.cx!=null){[e.cx,e.cy]=mir(e.cx,e.cy);}});return ids;
  };
  window.cadTrim        = () => JSON.stringify([]);
  window.cadExtend      = () => -1;
  window.cadFillet      = () => -1;
  window.cadChamfer     = () => -1;
  window.cadArrayRect   = (ids,r,c,rs,cs) => JSON.stringify(ids);
  window.cadArrayPolar  = (ids,cx,cy,n,a) => JSON.stringify(ids);
  window.cadOffset      = (ids,d)          => JSON.stringify([]);
  window.cadGetLayers   = () => JSON.stringify([{id:0,name:'0',color:'#ffffff',lineType:'Solid',lineWeight:0.25,visible:true,locked:false,frozen:false,print:true}]);
  window.cadAddLayer    = (name,color)     => 1;
  window.cadSetCurrentLayer = () => {};
  window.cadSetLayerVisible = () => {};
  window.cadSetLayerLocked  = () => {};
  window.cadSetLayerFrozen  = () => {};
  window.cadSetLayerColor   = () => {};
  window.cadSetLayerLineType   = () => {};
  window.cadSetLayerLineWeight = () => {};
  window.cadSetLayerPrint      = () => {};
  window.cadSetLayerName       = () => {};
  window.cadRemoveLayer        = () => false;
  window.cadDefineBlock   = () => true;
  window.cadInsertBlock   = () => -1;
  window.cadGetBlocks     = () => '[]';
  window.cadGetBlockEntities = () => '[]';
  window.cadExplodeBlock  = () => 'null';
  window.cadInsertSymbol  = () => -1;
  window.cadGetSymbols    = () => '[]';
  window.cadBoundingBox   = () => '';
  window.cadSnapToEntity  = () => '';
  window.cadFindSnap      = () => '';
  window.cadIntersect     = () => '';
  window.cadRenderHatch   = () => '[]';
  window.cadLoadDXF       = () => JSON.stringify({ok:false,error:'Demo mode — WASM not loaded'});
  window.cadSetEntityProp = (id, field, value) => {
    const e = ents.find(e => e.id === id);
    if (!e) return false;
    if (field === 'color') { e.color = value; return true; }
    if (field === 'layer') { e.layer = parseInt(value); return true; }
    if (field === 'text')  { e.text  = value; return true; }
    if (field === 'rotDeg')     { e.rotDeg = parseFloat(value); return true; }
    if (field === 'textHeight') { e.textHeight = parseFloat(value); return true; }
    if (field === 'x1')        { e.x1 = parseFloat(value); return true; }
    if (field === 'y1')        { e.y1 = parseFloat(value); return true; }
    if (field === 'x2')        { e.x2 = parseFloat(value); return true; }
    if (field === 'y2')        { e.y2 = parseFloat(value); return true; }
    if (field === 'cx')        { e.cx = parseFloat(value); return true; }
    if (field === 'cy')        { e.cy = parseFloat(value); return true; }
    if (field === 'r')         { e.r  = parseFloat(value); return true; }
    if (field === 'startDeg')  { e.startDeg = parseFloat(value); return true; }
    if (field === 'endDeg')    { e.endDeg   = parseFloat(value); return true; }
    return false;
  };
}
