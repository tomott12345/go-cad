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
	// ── Entity creation ──────────────────────────────────────────────────
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
