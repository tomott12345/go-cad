//go:build js

// cmd/wasm is the browser WebAssembly entry point for go-cad.
// It exposes a stable JavaScript API covering all entity types, undo/redo,
// DXF export (both R12 and R2000), and geometry engine queries.
//
// NOTE: cmd/cad (Fyne desktop) is tracked separately and is not implemented
// here because it requires the fyne.io/fyne/v2 external dependency, which
// is deliberately excluded from this module's pure-stdlib constraint.
// See: docs/architecture.md §6 "Desktop Frontend" for the re-scoping rationale.
package main

import (
        "encoding/json"
        "syscall/js"

        "go-cad/internal/document"
        "go-cad/internal/snap"
)

var doc = document.New()

func main() {
        // ── Primitive entity creation ─────────────────────────────────────────

        js.Global().Set("cadAddLine", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 6 {
                        return -1
                }
                return doc.AddLine(a[0].Float(), a[1].Float(), a[2].Float(), a[3].Float(),
                        a[4].Int(), a[5].String())
        }))

        js.Global().Set("cadAddCircle", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 5 {
                        return -1
                }
                return doc.AddCircle(a[0].Float(), a[1].Float(), a[2].Float(),
                        a[3].Int(), a[4].String())
        }))

        js.Global().Set("cadAddArc", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 7 {
                        return -1
                }
                return doc.AddArc(a[0].Float(), a[1].Float(), a[2].Float(),
                        a[3].Float(), a[4].Float(),
                        a[5].Int(), a[6].String())
        }))

        js.Global().Set("cadAddRectangle", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 6 {
                        return -1
                }
                return doc.AddRectangle(a[0].Float(), a[1].Float(), a[2].Float(), a[3].Float(),
                        a[4].Int(), a[5].String())
        }))

        js.Global().Set("cadAddPolyline", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 3 {
                        return -1
                }
                pts := jsArrToPoints(a[0])
                return doc.AddPolyline(pts, a[1].Int(), a[2].String())
        }))

        // ── Spline entities (Task #3) ─────────────────────────────────────────

        // cadAddSpline(points, layer, color) → entity ID
        // points: JS array of [x,y] pairs. Minimum 4 for a single cubic segment.
        // Layout: [p0, ctrl1, ctrl2, p1, ctrl3, ctrl4, p2, …] — standard cubic chain.
        // Keyboard: S  /  Command: SPLINE
        js.Global().Set("cadAddSpline", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 3 {
                        return -1
                }
                return doc.AddSpline(jsArrToPoints(a[0]), a[1].Int(), a[2].String())
        }))

        // cadAddNURBS(degree, points, knots, weights, layer, color) → entity ID
        //   degree  : integer B-spline degree (typically 3 = cubic)
        //   points  : JS array of [x,y] control points
        //   knots   : JS array of floats (pass null/[] for auto clamped-uniform knots)
        //   weights : JS array of floats (pass null/[] for all-1 uniform weights)
        //   layer, color
        // Command: NURBS
        js.Global().Set("cadAddNURBS", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 6 {
                        return -1
                }
                deg := a[0].Int()
                controls := jsArrToPoints(a[1])
                knots := jsArrToFloats(a[2])
                weights := jsArrToFloats(a[3])
                return doc.AddNURBS(deg, controls, knots, weights, a[4].Int(), a[5].String())
        }))

        // ── Ellipse (Task #3) ─────────────────────────────────────────────────

        // cadAddEllipse(cx, cy, a, b, rotDeg, layer, color) → entity ID
        // a: semi-major axis; b: semi-minor axis; rotDeg: rotation CCW from +X.
        // Keyboard: E  /  Command: ELLIPSE
        js.Global().Set("cadAddEllipse", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 7 {
                        return -1
                }
                return doc.AddEllipse(
                        a[0].Float(), a[1].Float(), // cx, cy
                        a[2].Float(), a[3].Float(), // semi-major, semi-minor
                        a[4].Float(),               // rotation
                        a[5].Int(), a[6].String())
        }))

        // ── Text entities (Task #3) ───────────────────────────────────────────

        // cadAddText(x, y, text, height, rotDeg, font, layer, color) → entity ID
        // font: SHX/TTF style name (empty string = "Standard").
        // Keyboard: T  /  Command: TEXT
        js.Global().Set("cadAddText", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 8 {
                        return -1
                }
                return doc.AddText(
                        a[0].Float(), a[1].Float(), // x, y
                        a[2].String(),              // text
                        a[3].Float(), a[4].Float(), // height, rotation
                        a[5].String(),              // font / style name
                        a[6].Int(), a[7].String())  // layer, color
        }))

        // cadAddMText(x, y, text, height, width, rotDeg, font, layer, color) → entity ID
        // text: multi-line content; use "\n" for paragraph breaks.
        // width: reference rectangle width (0 = no wrapping).
        // Command: MTEXT
        js.Global().Set("cadAddMText", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 9 {
                        return -1
                }
                return doc.AddMText(
                        a[0].Float(), a[1].Float(), // x, y
                        a[2].String(),              // text (supports \n)
                        a[3].Float(), a[4].Float(), // height, width
                        a[5].Float(),               // rotation
                        a[6].String(),              // font / style name
                        a[7].Int(), a[8].String())  // layer, color
        }))

        // ── Dimension entities (Task #3) ──────────────────────────────────────

        // cadAddLinearDim(x1,y1,x2,y2,offset,layer,color) → entity ID
        // offset: signed perpendicular distance from the measurement line to the dim line.
        // Command: DIMLIN
        js.Global().Set("cadAddLinearDim", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 7 {
                        return -1
                }
                return doc.AddLinearDim(
                        a[0].Float(), a[1].Float(),
                        a[2].Float(), a[3].Float(),
                        a[4].Float(),
                        a[5].Int(), a[6].String())
        }))

        // cadAddAlignedDim(x1,y1,x2,y2,offset,layer,color) → entity ID
        // Command: DIMALI
        js.Global().Set("cadAddAlignedDim", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 7 {
                        return -1
                }
                return doc.AddAlignedDim(
                        a[0].Float(), a[1].Float(),
                        a[2].Float(), a[3].Float(),
                        a[4].Float(),
                        a[5].Int(), a[6].String())
        }))

        // cadAddAngularDim(cx,cy,x1,y1,x2,y2,radius,layer,color) → entity ID
        // cx,cy: vertex; x1,y1 and x2,y2: points on the two rays.
        // Command: DIMANG
        js.Global().Set("cadAddAngularDim", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 9 {
                        return -1
                }
                return doc.AddAngularDim(
                        a[0].Float(), a[1].Float(),
                        a[2].Float(), a[3].Float(),
                        a[4].Float(), a[5].Float(),
                        a[6].Float(),
                        a[7].Int(), a[8].String())
        }))

        // cadAddRadialDim(cx,cy,r,angle,layer,color) → entity ID
        // angle: leader angle in degrees from +X.
        // Command: DIMRAD
        js.Global().Set("cadAddRadialDim", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 6 {
                        return -1
                }
                return doc.AddRadialDim(
                        a[0].Float(), a[1].Float(),
                        a[2].Float(), a[3].Float(),
                        a[4].Int(), a[5].String())
        }))

        // cadAddDiameterDim(cx,cy,r,angle,layer,color) → entity ID
        // Command: DIMDIA
        js.Global().Set("cadAddDiameterDim", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 6 {
                        return -1
                }
                return doc.AddDiameterDim(
                        a[0].Float(), a[1].Float(),
                        a[2].Float(), a[3].Float(),
                        a[4].Int(), a[5].String())
        }))

        // ── Deletion ─────────────────────────────────────────────────────────────
        js.Global().Set("cadDeleteEntity", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 1 {
                        return false
                }
                return doc.DeleteEntity(a[0].Int())
        }))

        // ── Undo / Redo ───────────────────────────────────────────────────────────
        js.Global().Set("cadUndo", js.FuncOf(func(_ js.Value, _ []js.Value) any {
                return doc.Undo()
        }))
        js.Global().Set("cadRedo", js.FuncOf(func(_ js.Value, _ []js.Value) any {
                return doc.Redo()
        }))
        js.Global().Set("cadClear", js.FuncOf(func(_ js.Value, _ []js.Value) any {
                doc.Clear(); return nil
        }))

        // ── Query ─────────────────────────────────────────────────────────────────
        js.Global().Set("cadEntities", js.FuncOf(func(_ js.Value, _ []js.Value) any {
                return doc.ToJSON()
        }))

        // cadExportDXF() → DXF R2000 (AC1015) string with DIMENSION entities.
        js.Global().Set("cadExportDXF", js.FuncOf(func(_ js.Value, _ []js.Value) any {
                return doc.ExportDXF()
        }))

        // cadExportDXFR12() → DXF R12 (AC1009) string; all entities reduced to
        // LINE/CIRCLE/ARC/TEXT/POLYLINE+VERTEX primitives.
        js.Global().Set("cadExportDXFR12", js.FuncOf(func(_ js.Value, _ []js.Value) any {
                return doc.ExportDXFR12()
        }))

        // ── Geometry engine ───────────────────────────────────────────────────────

        // cadBoundingBox(id) → JSON {"minX":…,"minY":…,"maxX":…,"maxY":…} or ""
        js.Global().Set("cadBoundingBox", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 1 {
                        return ""
                }
                bb := doc.EntityBoundingBox(a[0].Int())
                if bb.IsEmpty() {
                        return ""
                }
                b, _ := json.Marshal(map[string]float64{
                        "minX": bb.Min.X, "minY": bb.Min.Y,
                        "maxX": bb.Max.X, "maxY": bb.Max.Y,
                })
                return string(b)
        }))

        // cadSnapToEntity(id, x, y) → JSON {"x":…,"y":…}
        js.Global().Set("cadSnapToEntity", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 3 {
                        return ""
                }
                x, y := doc.SnapToEntity(a[0].Int(), a[1].Float(), a[2].Float())
                b, _ := json.Marshal(map[string]float64{"x": x, "y": y})
                return string(b)
        }))

        // cadNearestEntity(x, y, snapRadius) → entity ID (0 if none)
        js.Global().Set("cadNearestEntity", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 3 {
                        return 0
                }
                return doc.NearestEntity(a[0].Float(), a[1].Float(), a[2].Float())
        }))

        // cadIntersect(idA, idB) → JSON [[x,y],…]
        js.Global().Set("cadIntersect", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 2 {
                        return "[]"
                }
                pts := doc.IntersectEntities(a[0].Int(), a[1].Int())
                b, _ := json.Marshal(pts)
                return string(b)
        }))

        // cadOffsetEntity(id, dist) → new entity ID or -1
        js.Global().Set("cadOffsetEntity", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 2 {
                        return -1
                }
                return doc.OffsetEntity(a[0].Int(), a[1].Float())
        }))

        // cadTrimEntity(id, t) → JSON {"left":idL,"right":idR} or "null"
        js.Global().Set("cadTrimEntity", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 2 {
                        return "null"
                }
                idL, idR := doc.TrimEntity(a[0].Int(), a[1].Float())
                if idL < 0 || idR < 0 {
                        return "null"
                }
                b, _ := json.Marshal(map[string]int{"left": idL, "right": idR})
                return string(b)
        }))

        // cadEntityLength(id) → float64
        js.Global().Set("cadEntityLength", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 1 {
                        return 0.0
                }
                return doc.EntityLength(a[0].Int())
        }))

        // ── Task #4: Editing Operations ───────────────────────────────────────

        // cadMove(ids, dx, dy) → bool
        // ids: JS array of entity IDs; dx,dy: translation delta.
        // Command: M
        js.Global().Set("cadMove", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 3 {
                        return false
                }
                return doc.Move(jsArrToInts(a[0]), a[1].Float(), a[2].Float())
        }))

        // cadCopy(ids, dx, dy) → JSON [id, …]
        // Returns IDs of the newly created copies.
        // Command: CP / CO
        js.Global().Set("cadCopy", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 3 {
                        return "[]"
                }
                b, _ := json.Marshal(doc.Copy(jsArrToInts(a[0]), a[1].Float(), a[2].Float()))
                return string(b)
        }))

        // cadRotate(ids, cx, cy, angleDeg, makeCopy) → JSON [id, …]
        // angleDeg: rotation in degrees CCW; makeCopy: bool.
        // Command: RO
        js.Global().Set("cadRotate", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 5 {
                        return "[]"
                }
                b, _ := json.Marshal(doc.Rotate(jsArrToInts(a[0]),
                        a[1].Float(), a[2].Float(), a[3].Float(), a[4].Bool()))
                return string(b)
        }))

        // cadScale(ids, cx, cy, sx, sy, makeCopy) → JSON [id, …]
        // sx,sy: scale factors (use equal values for uniform scale).
        // Command: SC
        js.Global().Set("cadScale", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 6 {
                        return "[]"
                }
                b, _ := json.Marshal(doc.Scale(jsArrToInts(a[0]),
                        a[1].Float(), a[2].Float(), a[3].Float(), a[4].Float(), a[5].Bool()))
                return string(b)
        }))

        // cadMirror(ids, ax, ay, bx, by, makeCopy) → JSON [id, …]
        // Mirror line: (ax,ay)→(bx,by).
        // Command: MI
        js.Global().Set("cadMirror", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 6 {
                        return "[]"
                }
                b, _ := json.Marshal(doc.Mirror(jsArrToInts(a[0]),
                        a[1].Float(), a[2].Float(), a[3].Float(), a[4].Float(), a[5].Bool()))
                return string(b)
        }))

        // cadTrim(cutterID, targetID, pickX, pickY) → JSON [id, …] or "null"
        // Cuts targetID at its intersection with cutterID, removing the part
        // nearest to (pickX,pickY). Returns IDs of surviving sub-entities.
        // Command: TR
        js.Global().Set("cadTrim", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 4 {
                        return "null"
                }
                ids := doc.Trim(a[0].Int(), a[1].Int(), a[2].Float(), a[3].Float())
                if ids == nil {
                        return "null"
                }
                b, _ := json.Marshal(ids)
                return string(b)
        }))

        // cadExtend(boundaryID, targetID, pickX, pickY) → id or -1
        // Extends targetID to meet boundaryID; (pickX,pickY) selects which end.
        // Command: EX
        js.Global().Set("cadExtend", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 4 {
                        return -1
                }
                return doc.Extend(a[0].Int(), a[1].Int(), a[2].Float(), a[3].Float())
        }))

        // cadFillet(id1, id2, radius) → arc id or -1
        // Creates a tangent arc of given radius between two line entities.
        // Command: F
        js.Global().Set("cadFillet", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 3 {
                        return -1
                }
                return doc.Fillet(a[0].Int(), a[1].Int(), a[2].Float())
        }))

        // cadChamfer(id1, id2, dist1, dist2) → line id or -1
        // Creates a chamfer line between two line entities.
        // Command: CHA
        js.Global().Set("cadChamfer", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 4 {
                        return -1
                }
                return doc.Chamfer(a[0].Int(), a[1].Int(), a[2].Float(), a[3].Float())
        }))

        // cadArrayRect(ids, rows, cols, rowSpacing, colSpacing) → JSON [id, …]
        // Creates a rectangular grid of copies. Originals are at row 0, col 0.
        // Command: ARRAYRECT
        js.Global().Set("cadArrayRect", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 5 {
                        return "[]"
                }
                b, _ := json.Marshal(doc.ArrayRect(jsArrToInts(a[0]),
                        a[1].Int(), a[2].Int(), a[3].Float(), a[4].Float()))
                return string(b)
        }))

        // cadArrayPolar(ids, cx, cy, count, totalAngleDeg) → JSON [id, …]
        // Creates a polar array of count items over totalAngleDeg around (cx,cy).
        // Command: ARRAYPOLAR
        js.Global().Set("cadArrayPolar", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 5 {
                        return "[]"
                }
                b, _ := json.Marshal(doc.ArrayPolar(jsArrToInts(a[0]),
                        a[1].Float(), a[2].Float(), a[3].Int(), a[4].Float()))
                return string(b)
        }))

        // cadOffset(ids, dist) → JSON [id, …]
        // Creates parallel copies of entities at signed distance dist
        // (positive = left/outward; negative = right/inward). All ids are
        // handled as a single atomic undo step.
        // Command: OFFSET / O
        js.Global().Set("cadOffset", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 2 {
                        return "[]"
                }
                b, _ := json.Marshal(doc.Offset(jsArrToInts(a[0]), a[1].Float()))
                return string(b)
        }))

        // ── Task #5: Snap engine ──────────────────────────────────────────────────

        // cadFindSnap(x, y, radius, mask) → JSON {"x":…,"y":…,"type":"Endpoint","entityID":…} or ""
        // x,y: cursor in world space; radius: snap radius in world units;
        // mask: bitmask integer of enabled snap types (0 = all enabled).
        // Returns empty string when no snap found within radius.
        js.Global().Set("cadFindSnap", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 4 {
                        return ""
                }
                mask := snap.SnapType(a[3].Int())
                if mask == 0 {
                        mask = snap.SnapAll
                }
                result := snap.FindSnap(a[0].Float(), a[1].Float(), doc.Entities(), a[2].Float(), mask)
                if result == nil {
                        return ""
                }
                typeName := snap.SnapNames[result.Type]
                if typeName == "" {
                        typeName = "Unknown"
                }
                b, _ := json.Marshal(map[string]any{
                        "x":        result.X,
                        "y":        result.Y,
                        "type":     typeName,
                        "entityID": result.EntityID,
                })
                return string(b)
        }))

        // ── Task #5: Layer system ─────────────────────────────────────────────────

        // cadGetLayers() → JSON [{id,name,color,lineType,lineWeight,visible,locked,frozen,printEnabled}, …]
        js.Global().Set("cadGetLayers", js.FuncOf(func(_ js.Value, _ []js.Value) any {
                layers := doc.Layers()
                type layerJSON struct {
                        ID           int     `json:"id"`
                        Name         string  `json:"name"`
                        Color        string  `json:"color"`
                        LineType     string  `json:"lineType"`
                        LineWeight   float64 `json:"lineWeight"`
                        Visible      bool    `json:"visible"`
                        Locked       bool    `json:"locked"`
                        Frozen       bool    `json:"frozen"`
                        PrintEnabled bool    `json:"printEnabled"`
                }
                out := make([]layerJSON, 0, len(layers))
                for _, l := range layers {
                        out = append(out, layerJSON{
                                ID: l.ID, Name: l.Name, Color: l.Color,
                                LineType: string(l.LineTyp), LineWeight: l.LineWeight,
                                Visible: l.Visible, Locked: l.Locked,
                                Frozen: l.Frozen, PrintEnabled: l.PrintEnabled,
                        })
                }
                b, _ := json.Marshal(out)
                return string(b)
        }))

        // cadAddLayer(name, color) → layer id (int)
        // Uses Solid line type and 0.25mm weight as defaults.
        js.Global().Set("cadAddLayer", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 2 {
                        return -1
                }
                return doc.AddLayer(a[0].String(), a[1].String(), document.LineTypeSolid, 0.25)
        }))

        // cadRemoveLayer(id) → bool
        js.Global().Set("cadRemoveLayer", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 1 {
                        return false
                }
                return doc.RemoveLayer(a[0].Int())
        }))

        // cadSetLayerName(id, name) → bool
        js.Global().Set("cadSetLayerName", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 2 {
                        return false
                }
                return doc.RenameLayer(a[0].Int(), a[1].String())
        }))

        // cadSetLayerColor(id, color) → bool
        js.Global().Set("cadSetLayerColor", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 2 {
                        return false
                }
                return doc.SetLayerColor(a[0].Int(), a[1].String())
        }))

        // cadSetLayerVisible(id, visible) → bool
        js.Global().Set("cadSetLayerVisible", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 2 {
                        return false
                }
                return doc.SetLayerVisible(a[0].Int(), a[1].Bool())
        }))

        // cadSetLayerLocked(id, locked) → bool
        js.Global().Set("cadSetLayerLocked", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 2 {
                        return false
                }
                return doc.SetLayerLocked(a[0].Int(), a[1].Bool())
        }))

        // cadSetLayerFrozen(id, frozen) → bool
        js.Global().Set("cadSetLayerFrozen", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 2 {
                        return false
                }
                return doc.SetLayerFrozen(a[0].Int(), a[1].Bool())
        }))

        // cadGetCurrentLayer() → int
        js.Global().Set("cadGetCurrentLayer", js.FuncOf(func(_ js.Value, _ []js.Value) any {
                return doc.CurrentLayer()
        }))

        // cadSetCurrentLayer(id) → bool
        js.Global().Set("cadSetCurrentLayer", js.FuncOf(func(_ js.Value, a []js.Value) any {
                if len(a) < 1 {
                        return false
                }
                return doc.SetCurrentLayer(a[0].Int())
        }))

        select {} // block so the Go runtime stays alive
}

// ── JS type-conversion helpers ────────────────────────────────────────────────

// jsArrToPoints converts a JS array of [x,y] pairs to [][]float64.
func jsArrToPoints(v js.Value) [][]float64 {
        n := v.Length()
        pts := make([][]float64, n)
        for i := 0; i < n; i++ {
                pt := v.Index(i)
                pts[i] = []float64{pt.Index(0).Float(), pt.Index(1).Float()}
        }
        return pts
}

// jsArrToInts converts a JS array of integers to []int.
// Returns nil if the value is null/undefined/empty.
func jsArrToInts(v js.Value) []int {
        if v.IsNull() || v.IsUndefined() {
                return nil
        }
        n := v.Length()
        if n == 0 {
                return nil
        }
        out := make([]int, n)
        for i := 0; i < n; i++ {
                out[i] = v.Index(i).Int()
        }
        return out
}

// jsArrToFloats converts a JS array of numbers to []float64.
// Returns nil if the value is null/undefined/empty.
func jsArrToFloats(v js.Value) []float64 {
        if v.IsNull() || v.IsUndefined() {
                return nil
        }
        n := v.Length()
        if n == 0 {
                return nil
        }
        out := make([]float64, n)
        for i := 0; i < n; i++ {
                out[i] = v.Index(i).Float()
        }
        return out
}
