//go:build js

// cmd/wasm is the browser WebAssembly entry point for go-cad.
// It registers a set of JavaScript-callable functions that delegate to the
// shared internal/document package, keeping all geometry and undo/redo logic
// in one place (identical to what the desktop binary uses).
package main

import (
	"encoding/json"
	"syscall/js"

	"go-cad/internal/document"
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
		jsArr := a[0]
		n := jsArr.Length()
		pts := make([][]float64, n)
		for i := 0; i < n; i++ {
			pt := jsArr.Index(i)
			pts[i] = []float64{pt.Index(0).Float(), pt.Index(1).Float()}
		}
		return doc.AddPolyline(pts, a[1].Int(), a[2].String())
	}))

	// ── Advanced entity creation (Task #3) ────────────────────────────────

	// cadAddSpline(points_array, layer, color) → entity ID
	// points_array: JS array of [x,y] pairs (min 4 for a single cubic segment)
	// Layout: [p0, ctrl1, ctrl2, p1, ctrl3, ctrl4, p2, ...] — standard cubic chain.
	// Keyboard shortcut: S  /  Command: SPLINE
	js.Global().Set("cadAddSpline", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) < 3 {
			return -1
		}
		jsArr := a[0]
		n := jsArr.Length()
		pts := make([][]float64, n)
		for i := 0; i < n; i++ {
			pt := jsArr.Index(i)
			pts[i] = []float64{pt.Index(0).Float(), pt.Index(1).Float()}
		}
		return doc.AddSpline(pts, a[1].Int(), a[2].String())
	}))

	// cadAddEllipse(cx, cy, a, b, rotDeg, layer, color) → entity ID
	// a: semi-major axis; b: semi-minor axis; rotDeg: rotation CCW from +X
	// Keyboard shortcut: E  /  Command: ELLIPSE
	js.Global().Set("cadAddEllipse", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) < 7 {
			return -1
		}
		return doc.AddEllipse(
			a[0].Float(), a[1].Float(), // cx, cy
			a[2].Float(), a[3].Float(), // semi-major, semi-minor
			a[4].Float(),               // rotation degrees
			a[5].Int(), a[6].String())  // layer, color
	}))

	// cadAddText(x, y, text, height, rotDeg, layer, color) → entity ID
	// Keyboard shortcut: T  /  Command: TEXT
	js.Global().Set("cadAddText", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) < 7 {
			return -1
		}
		return doc.AddText(
			a[0].Float(), a[1].Float(), // x, y
			a[2].String(),              // text content
			a[3].Float(), a[4].Float(), // height, rotation
			a[5].Int(), a[6].String())  // layer, color
	}))

	// cadAddLinearDim(x1,y1,x2,y2,offset,layer,color) → entity ID
	// offset: perpendicular distance from measurement line to dim line.
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
			a[0].Float(), a[1].Float(), // cx, cy
			a[2].Float(), a[3].Float(), // x1, y1
			a[4].Float(), a[5].Float(), // x2, y2
			a[6].Float(),               // arc radius
			a[7].Int(), a[8].String())
	}))

	// cadAddRadialDim(cx,cy,r,angle,layer,color) → entity ID
	// angle: leader line angle in degrees from +X.
	// Command: DIMRAD
	js.Global().Set("cadAddRadialDim", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) < 6 {
			return -1
		}
		return doc.AddRadialDim(
			a[0].Float(), a[1].Float(), // cx, cy
			a[2].Float(), a[3].Float(), // radius, angle
			a[4].Int(), a[5].String())
	}))

	// cadAddDiameterDim(cx,cy,r,angle,layer,color) → entity ID
	// Command: DIMDIA
	js.Global().Set("cadAddDiameterDim", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) < 6 {
			return -1
		}
		return doc.AddDiameterDim(
			a[0].Float(), a[1].Float(), // cx, cy
			a[2].Float(), a[3].Float(), // radius, angle
			a[4].Int(), a[5].String())
	}))

	// ── Deletion ─────────────────────────────────────────────────────────
	js.Global().Set("cadDeleteEntity", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) < 1 {
			return false
		}
		return doc.DeleteEntity(a[0].Int())
	}))

	// ── Undo / Redo ──────────────────────────────────────────────────────
	js.Global().Set("cadUndo", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return doc.Undo()
	}))

	js.Global().Set("cadRedo", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return doc.Redo()
	}))

	js.Global().Set("cadClear", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		doc.Clear()
		return nil
	}))

	// ── Query ────────────────────────────────────────────────────────────
	js.Global().Set("cadEntities", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return doc.ToJSON()
	}))

	js.Global().Set("cadExportDXF", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return doc.ExportDXF()
	}))

	// ── Geometry engine: bounding box ─────────────────────────────────────
	// cadBoundingBox(id) → JSON string {"minX":…,"minY":…,"maxX":…,"maxY":…}
	// Returns "" if entity not found.
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

	// ── Geometry engine: snap to nearest point on an entity ───────────────
	// cadSnapToEntity(id, x, y) → JSON string {"x":…,"y":…}
	js.Global().Set("cadSnapToEntity", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) < 3 {
			return ""
		}
		x, y := doc.SnapToEntity(a[0].Int(), a[1].Float(), a[2].Float())
		b, _ := json.Marshal(map[string]float64{"x": x, "y": y})
		return string(b)
	}))

	// ── Geometry engine: find nearest entity ──────────────────────────────
	// cadNearestEntity(x, y, snapRadius) → entity ID (int), 0 if none
	js.Global().Set("cadNearestEntity", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) < 3 {
			return 0
		}
		return doc.NearestEntity(a[0].Float(), a[1].Float(), a[2].Float())
	}))

	// ── Geometry engine: intersect two entities ───────────────────────────
	// cadIntersect(idA, idB) → JSON array [[x,y],…]  or "[]" if none
	js.Global().Set("cadIntersect", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) < 2 {
			return "[]"
		}
		pts := doc.IntersectEntities(a[0].Int(), a[1].Int())
		b, _ := json.Marshal(pts)
		return string(b)
	}))

	// ── Geometry engine: offset entity ───────────────────────────────────
	// cadOffsetEntity(id, dist) → new entity ID, or -1 on failure
	js.Global().Set("cadOffsetEntity", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) < 2 {
			return -1
		}
		return doc.OffsetEntity(a[0].Int(), a[1].Float())
	}))

	// ── Geometry engine: trim / split entity ─────────────────────────────
	// cadTrimEntity(id, t) → JSON {"left":idL,"right":idR} or null on failure
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

	// ── Geometry engine: entity arc-length ───────────────────────────────
	// cadEntityLength(id) → float64
	js.Global().Set("cadEntityLength", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) < 1 {
			return 0.0
		}
		return doc.EntityLength(a[0].Int())
	}))

	// Block forever — the Go runtime must stay alive.
	select {}
}
