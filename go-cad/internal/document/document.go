// Package document provides the core CAD document model shared by both
// the native desktop (Fyne) and the browser WASM targets.
package document

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
)

// ─── Entity types ────────────────────────────────────────────────────────────

const (
	TypeLine        = "line"
	TypeCircle      = "circle"
	TypeArc         = "arc"
	TypeRectangle   = "rectangle"
	TypePolyline    = "polyline"
	TypeSpline      = "spline"
	TypeEllipse     = "ellipse"
	TypeText        = "text"
	TypeDimLinear   = "dimlin"
	TypeDimAligned  = "dimali"
	TypeDimAngular  = "dimang"
	TypeDimRadial   = "dimrad"
	TypeDimDiameter = "dimdia"
)

// Entity represents any CAD primitive stored in a document.
// All numeric fields that are always present are serialised without omitempty
// so that zero-valued coordinates are preserved correctly in JSON.
// New optional fields use omitempty.
type Entity struct {
	ID       int         `json:"id"`
	Type     string      `json:"type"`
	Layer    int         `json:"layer"`
	Color    string      `json:"color"`
	X1       float64     `json:"x1"`
	Y1       float64     `json:"y1"`
	X2       float64     `json:"x2"`
	Y2       float64     `json:"y2"`
	CX       float64     `json:"cx"`
	CY       float64     `json:"cy"`
	R        float64     `json:"r"`
	StartDeg float64     `json:"startDeg"`
	EndDeg   float64     `json:"endDeg"`
	Points   [][]float64 `json:"points,omitempty"`
	// New fields (Task #3 — Advanced Drawing Tools)
	R2         float64 `json:"r2,omitempty"`         // ellipse semi-minor axis
	RotDeg     float64 `json:"rotDeg,omitempty"`     // ellipse/text/dim leader rotation (degrees)
	Text       string  `json:"text,omitempty"`       // text content
	TextHeight float64 `json:"textHeight,omitempty"` // font height in document units
}

// Length returns the geometric length / circumference of the entity
// (useful for the properties panel).
func (e Entity) Length() float64 {
	switch e.Type {
	case TypeLine:
		return math.Hypot(e.X2-e.X1, e.Y2-e.Y1)
	case TypeCircle:
		return 2 * math.Pi * e.R
	case TypeArc:
		span := e.EndDeg - e.StartDeg
		if span < 0 {
			span += 360
		}
		return (span / 360) * 2 * math.Pi * e.R
	case TypeRectangle:
		return 2 * (math.Abs(e.X2-e.X1) + math.Abs(e.Y2-e.Y1))
	case TypePolyline:
		total := 0.0
		for i := 1; i < len(e.Points); i++ {
			total += math.Hypot(e.Points[i][0]-e.Points[i-1][0],
				e.Points[i][1]-e.Points[i-1][1])
		}
		return total
	case TypeSpline:
		// Approximate arc length via polyline of control points (fast estimate).
		// The geometry engine gives a more accurate value via BezierSpline.
		total := 0.0
		for i := 1; i < len(e.Points); i++ {
			total += math.Hypot(e.Points[i][0]-e.Points[i-1][0],
				e.Points[i][1]-e.Points[i-1][1])
		}
		return total
	case TypeEllipse:
		// Ramanujan's formula for ellipse circumference.
		a, b := e.R, e.R2
		if b < 0 {
			b = -b
		}
		h := (a-b)*(a-b)/((a+b)*(a+b)+1e-12)
		return math.Pi * (a + b) * (1 + 3*h/(10+math.Sqrt(4-3*h)))
	case TypeText:
		return 0
	case TypeDimLinear, TypeDimAligned:
		return math.Hypot(e.X2-e.X1, e.Y2-e.Y1)
	case TypeDimAngular:
		// Arc length of the angular dim arc (radius × angle in radians).
		dx1, dy1 := e.X1-e.CX, e.Y1-e.CY
		dx2, dy2 := e.X2-e.CX, e.Y2-e.CY
		ang1 := math.Atan2(dy1, dx1)
		ang2 := math.Atan2(dy2, dx2)
		span := ang2 - ang1
		if span < 0 {
			span += 2 * math.Pi
		}
		return span * e.R
	case TypeDimRadial:
		return e.R
	case TypeDimDiameter:
		return 2 * e.R
	}
	return 0
}

// maxUndoDepth caps the undo stack to prevent unbounded memory growth when
// many operations are performed in a single session.
const maxUndoDepth = 100

// ─── Document ────────────────────────────────────────────────────────────────

// Document is the in-memory CAD document with undo/redo support.
type Document struct {
	entities  []Entity
	nextID    int
	undoStack [][]Entity
	redoStack [][]Entity
}

// New returns an empty Document ready for use.
func New() *Document {
	return &Document{nextID: 1}
}

// Entities returns a copy of the current entity slice.
func (d *Document) Entities() []Entity {
	out := make([]Entity, len(d.entities))
	copy(out, d.entities)
	return out
}

// EntityCount returns the number of entities in the document.
func (d *Document) EntityCount() int { return len(d.entities) }

// ─── Internal helpers ─────────────────────────────────────────────────────────

func (d *Document) snapshot() []Entity {
	cp := make([]Entity, len(d.entities))
	copy(cp, d.entities)
	return cp
}

func (d *Document) pushUndo() {
	d.undoStack = append(d.undoStack, d.snapshot())
	// Trim oldest entries if we exceed the cap.
	if len(d.undoStack) > maxUndoDepth {
		d.undoStack = d.undoStack[len(d.undoStack)-maxUndoDepth:]
	}
	d.redoStack = nil // adding a new action clears the redo stack
}

func (d *Document) add(e Entity) int {
	d.pushUndo()
	e.ID = d.nextID
	d.nextID++
	if e.Color == "" {
		e.Color = "#ffffff"
	}
	d.entities = append(d.entities, e)
	return e.ID
}

// ─── Add operations ───────────────────────────────────────────────────────────

func (d *Document) AddLine(x1, y1, x2, y2 float64, layer int, color string) int {
	return d.add(Entity{Type: TypeLine, X1: x1, Y1: y1, X2: x2, Y2: y2, Layer: layer, Color: color})
}

func (d *Document) AddCircle(cx, cy, r float64, layer int, color string) int {
	return d.add(Entity{Type: TypeCircle, CX: cx, CY: cy, R: r, Layer: layer, Color: color})
}

func (d *Document) AddArc(cx, cy, r, startDeg, endDeg float64, layer int, color string) int {
	return d.add(Entity{Type: TypeArc, CX: cx, CY: cy, R: r, StartDeg: startDeg, EndDeg: endDeg, Layer: layer, Color: color})
}

func (d *Document) AddRectangle(x1, y1, x2, y2 float64, layer int, color string) int {
	return d.add(Entity{Type: TypeRectangle, X1: x1, Y1: y1, X2: x2, Y2: y2, Layer: layer, Color: color})
}

func (d *Document) AddPolyline(points [][]float64, layer int, color string) int {
	return d.add(Entity{Type: TypePolyline, Points: points, Layer: layer, Color: color})
}

// AddSpline adds a cubic Bezier spline entity defined by control points.
// For a single cubic segment, supply exactly 4 points; for N segments supply
// 3N+1 points with shared endpoints (the standard cubic Bezier chain layout).
// Keyboard shortcut: S / command: SPLINE
func (d *Document) AddSpline(points [][]float64, layer int, color string) int {
	return d.add(Entity{Type: TypeSpline, Points: points, Layer: layer, Color: color})
}

// AddEllipse adds an ellipse entity.
// cx, cy: centre; a: semi-major axis; b: semi-minor axis; rotDeg: rotation
// angle in degrees CCW from the positive X-axis.
// Keyboard shortcut: E / command: ELLIPSE
func (d *Document) AddEllipse(cx, cy, a, b, rotDeg float64, layer int, color string) int {
	return d.add(Entity{Type: TypeEllipse, CX: cx, CY: cy, R: a, R2: b, RotDeg: rotDeg, Layer: layer, Color: color})
}

// AddText adds a single-line text entity anchored at (x, y).
// height is the cap-height in document units; rotDeg is the baseline rotation
// in degrees CCW.  font is stored in the Color field for DXF SHX compatibility
// when a non-empty string is passed; otherwise the entity color is used.
// Keyboard shortcut: T / command: TEXT
func (d *Document) AddText(x, y float64, text string, height, rotDeg float64, layer int, color string) int {
	return d.add(Entity{Type: TypeText, X1: x, Y1: y, Text: text, TextHeight: height, RotDeg: rotDeg, Layer: layer, Color: color})
}

// AddLinearDim adds a linear dimension entity between two definition points.
// offset is the signed perpendicular distance from the measurement line to the
// dimension line (positive = above/left depending on orientation).
// Command: DIMLIN
func (d *Document) AddLinearDim(x1, y1, x2, y2, offset float64, layer int, color string) int {
	return d.add(Entity{Type: TypeDimLinear, X1: x1, Y1: y1, X2: x2, Y2: y2, CX: offset, Layer: layer, Color: color})
}

// AddAlignedDim adds an aligned dimension entity (along the entity direction).
// offset is the perpendicular offset of the dimension line from the measured segment.
// Command: DIMALI
func (d *Document) AddAlignedDim(x1, y1, x2, y2, offset float64, layer int, color string) int {
	return d.add(Entity{Type: TypeDimAligned, X1: x1, Y1: y1, X2: x2, Y2: y2, CX: offset, Layer: layer, Color: color})
}

// AddAngularDim adds an angular dimension between two rays from a common vertex.
// cx, cy: vertex; x1, y1 and x2, y2: points on the two rays; radius: arc radius.
// Command: DIMANG
func (d *Document) AddAngularDim(cx, cy, x1, y1, x2, y2, radius float64, layer int, color string) int {
	return d.add(Entity{Type: TypeDimAngular, CX: cx, CY: cy, X1: x1, Y1: y1, X2: x2, Y2: y2, R: radius, Layer: layer, Color: color})
}

// AddRadialDim adds a radial dimension with a leader line at angle rotDeg.
// Command: DIMRAD
func (d *Document) AddRadialDim(cx, cy, r, angle float64, layer int, color string) int {
	return d.add(Entity{Type: TypeDimRadial, CX: cx, CY: cy, R: r, RotDeg: angle, Layer: layer, Color: color})
}

// AddDiameterDim adds a diameter dimension with a leader line at angle rotDeg.
// Command: DIMDIA
func (d *Document) AddDiameterDim(cx, cy, r, angle float64, layer int, color string) int {
	return d.add(Entity{Type: TypeDimDiameter, CX: cx, CY: cy, R: r, RotDeg: angle, Layer: layer, Color: color})
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (d *Document) DeleteEntity(id int) bool {
	for i, e := range d.entities {
		if e.ID == id {
			d.pushUndo()
			d.entities = append(d.entities[:i], d.entities[i+1:]...)
			return true
		}
	}
	return false
}

// ─── Undo / Redo ──────────────────────────────────────────────────────────────

func (d *Document) Undo() bool {
	if len(d.undoStack) == 0 {
		return false
	}
	d.redoStack = append(d.redoStack, d.snapshot())
	last := len(d.undoStack) - 1
	d.entities = d.undoStack[last]
	d.undoStack = d.undoStack[:last]
	return true
}

func (d *Document) Redo() bool {
	if len(d.redoStack) == 0 {
		return false
	}
	d.undoStack = append(d.undoStack, d.snapshot())
	last := len(d.redoStack) - 1
	d.entities = d.redoStack[last]
	d.redoStack = d.redoStack[:last]
	return true
}

func (d *Document) Clear() {
	d.pushUndo()
	d.entities = nil
}

// ─── Serialisation ────────────────────────────────────────────────────────────

// ToJSON returns all entities as a JSON array string.
func (d *Document) ToJSON() string {
	b, _ := json.Marshal(d.entities)
	return string(b)
}

// ─── Generic add ──────────────────────────────────────────────────────────────

// AddEntity adds a generic entity to the document.
// The entity's ID field is ignored; a new ID is assigned.
// Returns the new entity ID, or -1 for an unknown type.
func (d *Document) AddEntity(e Entity) int {
	switch e.Type {
	case TypeLine:
		return d.AddLine(e.X1, e.Y1, e.X2, e.Y2, e.Layer, e.Color)
	case TypeCircle:
		return d.AddCircle(e.CX, e.CY, e.R, e.Layer, e.Color)
	case TypeArc:
		return d.AddArc(e.CX, e.CY, e.R, e.StartDeg, e.EndDeg, e.Layer, e.Color)
	case TypeRectangle:
		return d.AddRectangle(e.X1, e.Y1, e.X2, e.Y2, e.Layer, e.Color)
	case TypePolyline:
		return d.AddPolyline(e.Points, e.Layer, e.Color)
	case TypeSpline:
		return d.AddSpline(e.Points, e.Layer, e.Color)
	case TypeEllipse:
		return d.AddEllipse(e.CX, e.CY, e.R, e.R2, e.RotDeg, e.Layer, e.Color)
	case TypeText:
		return d.AddText(e.X1, e.Y1, e.Text, e.TextHeight, e.RotDeg, e.Layer, e.Color)
	case TypeDimLinear:
		return d.AddLinearDim(e.X1, e.Y1, e.X2, e.Y2, e.CX, e.Layer, e.Color)
	case TypeDimAligned:
		return d.AddAlignedDim(e.X1, e.Y1, e.X2, e.Y2, e.CX, e.Layer, e.Color)
	case TypeDimAngular:
		return d.AddAngularDim(e.CX, e.CY, e.X1, e.Y1, e.X2, e.Y2, e.R, e.Layer, e.Color)
	case TypeDimRadial:
		return d.AddRadialDim(e.CX, e.CY, e.R, e.RotDeg, e.Layer, e.Color)
	case TypeDimDiameter:
		return d.AddDiameterDim(e.CX, e.CY, e.R, e.RotDeg, e.Layer, e.Color)
	default:
		return -1
	}
}

// ─── Persistence ──────────────────────────────────────────────────────────────

// Save serialises all entities to a JSON file at path.
func (d *Document) Save(path string) error {
	data, err := json.Marshal(d.entities)
	if err != nil {
		return fmt.Errorf("document.Save: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("document.Save: %w", err)
	}
	return nil
}

// Load replaces the document contents from a JSON file at path.
// A snapshot is pushed onto the undo stack so the load can be undone.
func (d *Document) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("document.Load: %w", err)
	}
	var entities []Entity
	if err := json.Unmarshal(data, &entities); err != nil {
		return fmt.Errorf("document.Load: %w", err)
	}
	d.pushUndo()
	d.entities = entities
	for _, e := range entities {
		if e.ID >= d.nextID {
			d.nextID = e.ID + 1
		}
	}
	return nil
}

// ─── DXF export ───────────────────────────────────────────────────────────────

// ExportDXF returns a DXF R12/R2000-compatible string for all entities.
// Y-axis is flipped (DXF uses Cartesian, canvas uses screen coordinates).
func (d *Document) ExportDXF() string {
	var sb strings.Builder
	sb.WriteString("  0\nSECTION\n  2\nHEADER\n  9\n$ACADVER\n  1\nAC1015\n  0\nENDSEC\n")
	sb.WriteString("  0\nSECTION\n  2\nENTITIES\n")
	for _, e := range d.entities {
		switch e.Type {
		case TypeLine:
			fmt.Fprintf(&sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
				e.Layer, e.X1, -e.Y1, e.X2, -e.Y2)

		case TypeCircle:
			fmt.Fprintf(&sb, "  0\nCIRCLE\n  8\n%d\n 10\n%f\n 20\n%f\n 40\n%f\n",
				e.Layer, e.CX, -e.CY, e.R)

		case TypeArc:
			fmt.Fprintf(&sb, "  0\nARC\n  8\n%d\n 10\n%f\n 20\n%f\n 40\n%f\n 50\n%f\n 51\n%f\n",
				e.Layer, e.CX, -e.CY, e.R, e.StartDeg, e.EndDeg)

		case TypeRectangle:
			x1, y1, x2, y2 := e.X1, e.Y1, e.X2, e.Y2
			fmt.Fprintf(&sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n", e.Layer, x1, -y1, x2, -y1)
			fmt.Fprintf(&sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n", e.Layer, x2, -y1, x2, -y2)
			fmt.Fprintf(&sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n", e.Layer, x2, -y2, x1, -y2)
			fmt.Fprintf(&sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n", e.Layer, x1, -y2, x1, -y1)

		case TypePolyline:
			if len(e.Points) < 2 {
				continue
			}
			for i := 0; i < len(e.Points)-1; i++ {
				p1, p2 := e.Points[i], e.Points[i+1]
				fmt.Fprintf(&sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
					e.Layer, p1[0], -p1[1], p2[0], -p2[1])
			}

		case TypeSpline:
			// Export as a LWPOLYLINE approximation (DXF R2000).
			pts := approxBezierPoints(e.Points, 20)
			if len(pts) < 2 {
				continue
			}
			fmt.Fprintf(&sb, "  0\nLWPOLYLINE\n  8\n%d\n 90\n%d\n 70\n0\n",
				e.Layer, len(pts))
			for _, p := range pts {
				fmt.Fprintf(&sb, " 10\n%f\n 20\n%f\n", p[0], -p[1])
			}

		case TypeEllipse:
			// DXF R2000 ELLIPSE entity.
			// Group 11,21: endpoint of major axis relative to centre (before rotation).
			// Group 40: ratio of minor to major axis (b/a).
			rot := e.RotDeg * math.Pi / 180
			mx := e.R * math.Cos(rot)
			my := e.R * math.Sin(rot)
			ratio := 1.0
			if e.R > 1e-12 {
				ratio = e.R2 / e.R
			}
			fmt.Fprintf(&sb,
				"  0\nELLIPSE\n  8\n%d\n 10\n%f\n 20\n%f\n 30\n0.0\n 11\n%f\n 21\n%f\n 31\n0.0\n 40\n%f\n 41\n0.0\n 42\n%f\n",
				e.Layer, e.CX, -e.CY, mx, -my, ratio, 2*math.Pi)

		case TypeText:
			h := e.TextHeight
			if h <= 0 {
				h = 2.5
			}
			fmt.Fprintf(&sb,
				"  0\nTEXT\n  8\n%d\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n%f\n  1\n%s\n 50\n%f\n",
				e.Layer, e.X1, -e.Y1, h, e.Text, e.RotDeg)

		case TypeDimLinear:
			dxfLinearDim(&sb, e)

		case TypeDimAligned:
			dxfAlignedDim(&sb, e)

		case TypeDimAngular:
			dxfAngularDim(&sb, e)

		case TypeDimRadial:
			dxfRadialDim(&sb, e)

		case TypeDimDiameter:
			dxfDiameterDim(&sb, e)
		}
	}
	sb.WriteString("  0\nENDSEC\n  0\nEOF\n")
	return sb.String()
}

// ─── DXF helpers ──────────────────────────────────────────────────────────────

// approxBezierPoints returns a polyline approximation of a cubic Bezier spline
// defined by control points pts. n = subdivisions per segment.
func approxBezierPoints(pts [][]float64, n int) [][2]float64 {
	nCtrl := len(pts)
	if nCtrl < 4 {
		// Not enough for even one cubic segment; just return as-is.
		out := make([][2]float64, nCtrl)
		for i, p := range pts {
			if len(p) >= 2 {
				out[i] = [2]float64{p[0], p[1]}
			}
		}
		return out
	}
	nSegs := (nCtrl - 1) / 3
	var out [][2]float64
	for seg := 0; seg < nSegs; seg++ {
		i := seg * 3
		p0 := [2]float64{pts[i][0], pts[i][1]}
		p1 := [2]float64{pts[i+1][0], pts[i+1][1]}
		p2 := [2]float64{pts[i+2][0], pts[i+2][1]}
		p3 := [2]float64{pts[i+3][0], pts[i+3][1]}
		for k := 0; k < n; k++ {
			t := float64(k) / float64(n)
			u := 1 - t
			x := u*u*u*p0[0] + 3*u*u*t*p1[0] + 3*u*t*t*p2[0] + t*t*t*p3[0]
			y := u*u*u*p0[1] + 3*u*u*t*p1[1] + 3*u*t*t*p2[1] + t*t*t*p3[1]
			out = append(out, [2]float64{x, y})
		}
	}
	// Add final point.
	last := pts[len(pts)-1]
	out = append(out, [2]float64{last[0], last[1]})
	return out
}

// arrowLine emits a DXF LINE arrowhead at point p in direction dir of length len.
func arrowLine(sb *strings.Builder, layer int, px, py, dx, dy float64) {
	alen := 3.0 // arrowhead size in document units
	// Normalize direction
	mag := math.Hypot(dx, dy)
	if mag < 1e-12 {
		return
	}
	dx /= mag
	dy /= mag
	// Two barb lines at ±20° from the backward direction
	for _, sign := range []float64{1, -1} {
		ang := math.Atan2(dy, dx) + math.Pi + sign*math.Pi/9
		fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
			layer, px, -py,
			px+math.Cos(ang)*alen, -(py+math.Sin(ang)*alen))
	}
}

// dxfLinearDim emits a linear dimension as LINE + TEXT entities.
// e.X1,Y1 and e.X2,Y2 are the definition points; e.CX is the offset of the
// dimension line from the Y midpoint (horizontal dim) or X midpoint (vertical).
func dxfLinearDim(sb *strings.Builder, e Entity) {
	// Determine if horizontal or vertical based on coordinates.
	dx := e.X2 - e.X1
	dy := e.Y2 - e.Y1
	offset := e.CX
	if math.Abs(dx) >= math.Abs(dy) {
		// Horizontal dimension: dim line runs parallel to X axis at Y = midY - offset.
		dimY := (e.Y1+e.Y2)/2 - offset
		// Extension lines from definition points to dim line.
		fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
			e.Layer, e.X1, -e.Y1, e.X1, -dimY)
		fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
			e.Layer, e.X2, -e.Y2, e.X2, -dimY)
		// Dimension line.
		fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
			e.Layer, e.X1, -dimY, e.X2, -dimY)
		// Arrowheads.
		arrowLine(sb, e.Layer, e.X1, dimY, e.X2-e.X1, 0)
		arrowLine(sb, e.Layer, e.X2, dimY, e.X1-e.X2, 0)
		// Text label.
		midX := (e.X1 + e.X2) / 2
		val := math.Abs(dx)
		fmt.Fprintf(sb, "  0\nTEXT\n  8\n%d\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n2.5\n  1\n%.3f\n 50\n0.0\n",
			e.Layer, midX, -dimY+2, val)
	} else {
		// Vertical dimension: dim line runs parallel to Y axis at X = midX - offset.
		dimX := (e.X1+e.X2)/2 - offset
		fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
			e.Layer, e.X1, -e.Y1, dimX, -e.Y1)
		fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
			e.Layer, e.X2, -e.Y2, dimX, -e.Y2)
		fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
			e.Layer, dimX, -e.Y1, dimX, -e.Y2)
		arrowLine(sb, e.Layer, dimX, e.Y1, 0, e.Y2-e.Y1)
		arrowLine(sb, e.Layer, dimX, e.Y2, 0, e.Y1-e.Y2)
		midY := (e.Y1 + e.Y2) / 2
		val := math.Abs(dy)
		fmt.Fprintf(sb, "  0\nTEXT\n  8\n%d\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n2.5\n  1\n%.3f\n 50\n90.0\n",
			e.Layer, dimX-3, -midY, val)
	}
}

// dxfAlignedDim emits an aligned dimension (along the direction of the segment).
func dxfAlignedDim(sb *strings.Builder, e Entity) {
	dx := e.X2 - e.X1
	dy := e.Y2 - e.Y1
	dist := math.Hypot(dx, dy)
	if dist < 1e-12 {
		return
	}
	// Unit direction and perpendicular.
	ux, uy := dx/dist, dy/dist
	px, py := -uy, ux // perpendicular (left normal)

	offset := e.CX
	// Dim line endpoints: offset from definition points along perpendicular.
	d1x, d1y := e.X1+px*offset, e.Y1+py*offset
	d2x, d2y := e.X2+px*offset, e.Y2+py*offset

	// Extension lines.
	fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
		e.Layer, e.X1, -e.Y1, d1x, -d1y)
	fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
		e.Layer, e.X2, -e.Y2, d2x, -d2y)
	// Dimension line.
	fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
		e.Layer, d1x, -d1y, d2x, -d2y)
	arrowLine(sb, e.Layer, d1x, d1y, d2x-d1x, d2y-d1y)
	arrowLine(sb, e.Layer, d2x, d2y, d1x-d2x, d1y-d2y)
	// Text.
	midX := (d1x + d2x) / 2
	midY := (d1y + d2y) / 2
	angDeg := math.Atan2(dy, dx) * 180 / math.Pi
	fmt.Fprintf(sb, "  0\nTEXT\n  8\n%d\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n2.5\n  1\n%.3f\n 50\n%f\n",
		e.Layer, midX, -midY+2, dist, angDeg)
}

// dxfAngularDim emits an angular dimension arc between two rays from a vertex.
func dxfAngularDim(sb *strings.Builder, e Entity) {
	// Compute angles of the two rays.
	ang1 := math.Atan2(e.Y1-e.CY, e.X1-e.CX)
	ang2 := math.Atan2(e.Y2-e.CY, e.X2-e.CX)

	// Draw arc approximation with 16 segments.
	r := e.R
	if r <= 0 {
		r = math.Min(math.Hypot(e.X1-e.CX, e.Y1-e.CY),
			math.Hypot(e.X2-e.CX, e.Y2-e.CY)) * 0.5
	}
	span := ang2 - ang1
	if span < 0 {
		span += 2 * math.Pi
	}
	const nArc = 16
	prev := [2]float64{e.CX + r*math.Cos(ang1), e.CY + r*math.Sin(ang1)}
	for k := 1; k <= nArc; k++ {
		a := ang1 + float64(k)/float64(nArc)*span
		cur := [2]float64{e.CX + r*math.Cos(a), e.CY + r*math.Sin(a)}
		fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
			e.Layer, prev[0], -prev[1], cur[0], -cur[1])
		prev = cur
	}
	// Leader lines from vertex.
	fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
		e.Layer, e.CX, -e.CY, e.CX+r*math.Cos(ang1), -(e.CY+r*math.Sin(ang1)))
	fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
		e.Layer, e.CX, -e.CY, e.CX+r*math.Cos(ang2), -(e.CY+r*math.Sin(ang2)))
	// Angle text at mid-arc.
	midAng := ang1 + span/2
	angDeg := span * 180 / math.Pi
	fmt.Fprintf(sb, "  0\nTEXT\n  8\n%d\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n2.5\n  1\n%.1f°\n 50\n0.0\n",
		e.Layer, e.CX+r*1.3*math.Cos(midAng), -(e.CY+r*1.3*math.Sin(midAng)), angDeg)
}

// dxfRadialDim emits a radial dimension leader line.
func dxfRadialDim(sb *strings.Builder, e Entity) {
	ang := e.RotDeg * math.Pi / 180
	px := e.CX + e.R*math.Cos(ang)
	py := e.CY + e.R*math.Sin(ang)
	// Leader from center to circumference.
	fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
		e.Layer, e.CX, -e.CY, px, -py)
	arrowLine(sb, e.Layer, px, py, px-e.CX, py-e.CY)
	// Text just outside.
	tx := e.CX + (e.R+4)*math.Cos(ang)
	ty := e.CY + (e.R+4)*math.Sin(ang)
	fmt.Fprintf(sb, "  0\nTEXT\n  8\n%d\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n2.5\n  1\nR%.3f\n 50\n0.0\n",
		e.Layer, tx, -ty, e.R)
}

// dxfDiameterDim emits a diameter dimension with two-sided leader.
func dxfDiameterDim(sb *strings.Builder, e Entity) {
	ang := e.RotDeg * math.Pi / 180
	p1x := e.CX + e.R*math.Cos(ang)
	p1y := e.CY + e.R*math.Sin(ang)
	p2x := e.CX - e.R*math.Cos(ang)
	p2y := e.CY - e.R*math.Sin(ang)
	// Diameter line through center.
	fmt.Fprintf(sb, "  0\nLINE\n  8\n%d\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
		e.Layer, p1x, -p1y, p2x, -p2y)
	arrowLine(sb, e.Layer, p1x, p1y, p1x-p2x, p1y-p2y)
	arrowLine(sb, e.Layer, p2x, p2y, p2x-p1x, p2y-p1y)
	// Text at center+offset.
	tx := e.CX + (e.R+4)*math.Cos(ang)
	ty := e.CY + (e.R+4)*math.Sin(ang)
	fmt.Fprintf(sb, "  0\nTEXT\n  8\n%d\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n2.5\n  1\n⌀%.3f\n 50\n0.0\n",
		e.Layer, tx, -ty, 2*e.R)
}
