// Package symbols provides a library of built-in CAD symbols that can be
// inserted into a document as named block definitions.
//
// Entities inside each symbol are in block-local coordinates (base = origin).
// All entity IDs are left at 0; real IDs are assigned when the block is
// registered with a Document via document.DefineBlockRaw.
package symbols

import "go-cad/internal/document"

// Names returns the list of built-in symbol names.
func Names() []string {
	return []string{
		"CENTER_MARK",
		"NORTH_ARROW",
		"REVISION_TRIANGLE",
		"DATUM_TRIANGLE",
		"SURFACE_FINISH",
	}
}

// Entities returns the raw block-local entities for the named symbol, or nil
// if the symbol is unknown.  Units are normalised so the symbol fits inside a
// 10×10 box centred on the origin.
func Entities(name string) []document.Entity {
	switch name {
	case "CENTER_MARK":
		return centerMark()
	case "NORTH_ARROW":
		return northArrow()
	case "REVISION_TRIANGLE":
		return revisionTriangle()
	case "DATUM_TRIANGLE":
		return datumTriangle()
	case "SURFACE_FINISH":
		return surfaceFinish()
	}
	return nil
}

// Register installs all built-in symbols into the given document as block
// definitions.  Call once at startup.  Safe to call multiple times (idempotent).
func Register(doc *document.Document) {
	for _, name := range Names() {
		ents := Entities(name)
		if len(ents) > 0 {
			doc.DefineBlockRaw(name, 0, 0, ents)
		}
	}
}

// ─── Symbol definitions ───────────────────────────────────────────────────────

// centerMark: two intersecting lines forming a + symbol.
//   Fits in ±5 × ±5 box.
func centerMark() []document.Entity {
	col := "#ffffff"
	return []document.Entity{
		{Type: document.TypeLine, X1: -5, Y1: 0, X2: 5, Y2: 0, Color: col},
		{Type: document.TypeLine, X1: 0, Y1: -5, X2: 0, Y2: 5, Color: col},
		{Type: document.TypeCircle, CX: 0, CY: 0, R: 1.5, Color: col},
	}
}

// northArrow: an upward-pointing arrow with the letter N.
//   Fits in ±4 × 0..10 box.
func northArrow() []document.Entity {
	col := "#ffffff"
	return []document.Entity{
		// Arrow shaft
		{Type: document.TypeLine, X1: 0, Y1: 0, X2: 0, Y2: 8, Color: col},
		// Arrow head (two lines)
		{Type: document.TypeLine, X1: 0, Y1: 8, X2: -2.5, Y2: 4, Color: col},
		{Type: document.TypeLine, X1: 0, Y1: 8, X2: 2.5, Y2: 4, Color: col},
		// "N" label above
		{Type: document.TypeText, X1: -1.5, Y1: 10, Text: "N", TextHeight: 3, Color: col},
	}
}

// revisionTriangle: an upward-pointing filled triangle with an "R" inside.
func revisionTriangle() []document.Entity {
	col := "#ffffff"
	pts := [][]float64{{0, 6}, {-5, -3}, {5, -3}, {0, 6}}
	return []document.Entity{
		{Type: document.TypePolyline, Points: pts, Color: col},
		{Type: document.TypeText, X1: -1, Y1: 1, Text: "R", TextHeight: 3, Color: col},
	}
}

// datumTriangle: an equilateral triangle pointing down, used as a datum feature symbol.
func datumTriangle() []document.Entity {
	col := "#ffffff"
	pts := [][]float64{{0, -5}, {-4, 3}, {4, 3}, {0, -5}}
	return []document.Entity{
		{Type: document.TypePolyline, Points: pts, Color: col},
		{Type: document.TypeLine, X1: 0, Y1: 3, X2: 0, Y2: 7, Color: col},
	}
}

// surfaceFinish: a basic surface texture symbol (check-mark with a bar).
func surfaceFinish() []document.Entity {
	col := "#ffffff"
	return []document.Entity{
		{Type: document.TypeLine, X1: -3, Y1: 0, X2: 0, Y2: -4, Color: col},
		{Type: document.TypeLine, X1: 0, Y1: -4, X2: 5, Y2: 4, Color: col},
		{Type: document.TypeLine, X1: 0, Y1: -4, X2: 8, Y2: -4, Color: col},
	}
}
