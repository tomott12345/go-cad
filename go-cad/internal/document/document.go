// Package document provides the core CAD document model shared by both
// the native desktop and the browser WASM targets.
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
        TypeSpline      = "spline"  // cubic Bezier spline (control-point chain)
        TypeNURBS       = "nurbs"   // Non-Uniform Rational B-Spline
        TypeEllipse     = "ellipse"
        TypeText        = "text"    // single-line text
        TypeMText       = "mtext"   // multi-line text
        TypeDimLinear   = "dimlin"
        TypeDimAligned  = "dimali"
        TypeDimAngular  = "dimang"
        TypeDimRadial   = "dimrad"
        TypeDimDiameter = "dimdia"

        // ── Task #7: Blocks, Hatching & Annotations ──────────────────────────
        TypeBlockRef      = "blockref"    // block insertion proxy
        TypeHatch         = "hatch"       // filled polygon hatch
        TypeLeader        = "leader"      // multi-segment leader with text
        TypeRevisionCloud = "revcloud"    // revision cloud (arc-chain polygon)
        TypeWipeout       = "wipeout"     // opaque masking polygon
)

// Entity represents any CAD primitive stored in a document.
// Fields shared across many entity types (x1/y1, cx/cy, r, etc.) are reused
// according to the Type discriminator. Optional fields use omitempty so that
// zero-valued entries are omitted from the JSON serialisation.
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

        // Advanced entity fields (Task #3 — Advanced Drawing Tools)
        R2         float64   `json:"r2,omitempty"`         // ellipse semi-minor axis; MTEXT reference rectangle width
        RotDeg     float64   `json:"rotDeg,omitempty"`     // ellipse/text/dim rotation in degrees (CCW)
        Text       string    `json:"text,omitempty"`       // text content (single-line or multi-line; use \n for line breaks)
        TextHeight float64   `json:"textHeight,omitempty"` // font cap-height in document units
        Font       string    `json:"font,omitempty"`       // SHX / TTF font / text-style name (DXF group 7)

        // NURBS-specific fields
        NURBSDegree int       `json:"nurbsDeg,omitempty"` // B-spline degree (typically 3)
        Knots       []float64 `json:"knots,omitempty"`    // knot vector (len = nControls + nurbsDeg + 1)
        Weights     []float64 `json:"weights,omitempty"`  // rational weights (len = nControls; nil ↦ all 1)

        // Per-entity override fields (Task #8)
        LineType   string  `json:"lineType,omitempty"`   // entity-level linetype override ("" = use layer)
        LineWeight float64 `json:"lineWeight,omitempty"` // entity-level lineweight override (0 = use layer)
}

// Length returns the geometric length / circumference of the entity.
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
                // Fast estimate via control-polygon chord length.
                total := 0.0
                for i := 1; i < len(e.Points); i++ {
                        total += math.Hypot(e.Points[i][0]-e.Points[i-1][0],
                                e.Points[i][1]-e.Points[i-1][1])
                }
                return total
        case TypeNURBS:
                // Approximate arc length via a 100-sample polyline.
                pts := nurbsApprox(e, 100)
                total := 0.0
                for i := 1; i < len(pts); i++ {
                        total += math.Hypot(pts[i][0]-pts[i-1][0], pts[i][1]-pts[i-1][1])
                }
                return total
        case TypeEllipse:
                a, b := e.R, e.R2
                if b < 0 {
                        b = -b
                }
                denom := (a+b)*(a+b) + 1e-12
                h := (a-b)*(a-b)/denom
                return math.Pi * (a + b) * (1 + 3*h/(10+math.Sqrt(4-3*h)))
        case TypeText, TypeMText:
                return 0
        case TypeDimLinear, TypeDimAligned:
                return math.Hypot(e.X2-e.X1, e.Y2-e.Y1)
        case TypeDimAngular:
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
        case TypeBlockRef:
                return 0
        case TypeHatch, TypeWipeout:
                // Perimeter of boundary polygon.
                total := 0.0
                for i := 1; i < len(e.Points); i++ {
                        total += math.Hypot(e.Points[i][0]-e.Points[i-1][0],
                                e.Points[i][1]-e.Points[i-1][1])
                }
                return total
        case TypeLeader:
                total := 0.0
                for i := 1; i < len(e.Points); i++ {
                        total += math.Hypot(e.Points[i][0]-e.Points[i-1][0],
                                e.Points[i][1]-e.Points[i-1][1])
                }
                return total
        case TypeRevisionCloud:
                // Approximate as polygon perimeter.
                total := 0.0
                n := len(e.Points)
                for i := 0; i < n; i++ {
                        j := (i + 1) % n
                        total += math.Hypot(e.Points[j][0]-e.Points[i][0],
                                e.Points[j][1]-e.Points[i][1])
                }
                return total
        }
        return 0
}

// ─── NURBS helpers (pure Go, no geometry-package import needed here) ──────────

// nurbsBasis evaluates the Cox-de Boor basis N_{i,k}(t).
func nurbsBasis(i, k int, knots []float64, t float64) float64 {
        if k == 0 {
                if knots[i] <= t && t < knots[i+1] {
                        return 1
                }
                return 0
        }
        d1 := knots[i+k] - knots[i]
        d2 := knots[i+k+1] - knots[i+1]
        var left, right float64
        if d1 > 1e-12 {
                left = (t-knots[i])/d1*nurbsBasis(i, k-1, knots, t)
        }
        if d2 > 1e-12 {
                right = (knots[i+k+1]-t)/d2*nurbsBasis(i+1, k-1, knots, t)
        }
        return left + right
}

// nurbsPoint evaluates a rational B-spline at parameter t.
func nurbsPoint(degree int, controls [][]float64, knots []float64, weights []float64, t float64) [2]float64 {
        n := len(controls)
        var wx, wy, w float64
        for i := 0; i < n; i++ {
                b := nurbsBasis(i, degree, knots, t)
                wi := 1.0
                if i < len(weights) {
                        wi = weights[i]
                }
                bw := b * wi
                wx += bw * controls[i][0]
                wy += bw * controls[i][1]
                w += bw
        }
        if w < 1e-12 {
                if len(controls) > 0 {
                        return [2]float64{controls[0][0], controls[0][1]}
                }
                return [2]float64{}
        }
        return [2]float64{wx / w, wy / w}
}

// nurbsClampedUniformKnots generates a clamped uniform knot vector.
func nurbsClampedUniformKnots(nControls, degree int) []float64 {
        m := nControls + degree + 1
        knots := make([]float64, m)
        inner := nControls - degree
        for i := 0; i <= degree; i++ {
                knots[i] = 0
        }
        for i := 1; i < inner; i++ {
                knots[degree+i] = float64(i) / float64(inner)
        }
        for i := nControls; i < m; i++ {
                knots[i] = 1
        }
        return knots
}

// nurbsApprox returns n+1 world-space sample points along the NURBS curve.
func nurbsApprox(e Entity, n int) [][2]float64 {
        controls := e.Points
        nc := len(controls)
        if nc == 0 {
                return nil
        }
        deg := e.NURBSDegree
        if deg < 1 {
                deg = 3
        }
        knots := e.Knots
        if len(knots) < nc+deg+1 {
                knots = nurbsClampedUniformKnots(nc, deg)
        }
        lo := knots[deg]
        hi := knots[nc]
        pts := make([][2]float64, n+1)
        for i := 0; i <= n; i++ {
                t := lo + float64(i)/float64(n)*(hi-lo)
                if t >= hi {
                        t = hi - 1e-12
                }
                pts[i] = nurbsPoint(deg, controls, knots, e.Weights, t)
        }
        return pts
}

// ─── maxUndoDepth ────────────────────────────────────────────────────────────

const maxUndoDepth = 100

// ─── Document ─────────────────────────────────────────────────────────────────

// docSnapshot captures a complete, self-consistent document state for undo/redo.
// Both entities and layer state are included so that import operations (which
// replace both entities and layers atomically) can be fully undone/redone.
type docSnapshot struct {
        entities    []Entity
        nextID      int
        layers      map[int]*Layer
        nextLayerID int
        curLayer    int
        blocks      map[string]*Block // Task #7
}

// Document is the in-memory CAD document with undo/redo support.
type Document struct {
        entities    []Entity
        nextID      int
        undoStack   []docSnapshot
        redoStack   []docSnapshot
        layers      map[int]*Layer  // keyed by layer ID
        nextLayerID int
        curLayer    int
        blocks      map[string]*Block // Task #7: named block definitions
}

// New returns an empty Document ready for use.
// It pre-creates the mandatory default layer "0".
func New() *Document {
        d := &Document{nextID: 1, nextLayerID: 1}
        d.layers = map[int]*Layer{0: defaultLayer0()}
        return d
}

// Entities returns a copy of the current entity slice.
func (d *Document) Entities() []Entity {
        out := make([]Entity, len(d.entities))
        copy(out, d.entities)
        return out
}

// EntityCount returns the number of entities in the document.
func (d *Document) EntityCount() int { return len(d.entities) }

// ─── Internal helpers ──────────────────────────────────────────────────────────

// copyLayers returns a deep copy of the layers map.
// Layer has only value-type fields so a struct copy is sufficient.
func copyLayers(src map[int]*Layer) map[int]*Layer {
        dst := make(map[int]*Layer, len(src))
        for id, l := range src {
                cp := *l
                dst[id] = &cp
        }
        return dst
}

func (d *Document) snapshot() docSnapshot {
        ents := make([]Entity, len(d.entities))
        copy(ents, d.entities)
        return docSnapshot{
                entities:    ents,
                nextID:      d.nextID,
                layers:      copyLayers(d.layers),
                nextLayerID: d.nextLayerID,
                curLayer:    d.curLayer,
                blocks:      copyBlocks(d.blocks),
        }
}

func (d *Document) restoreSnapshot(s docSnapshot) {
        d.entities = s.entities
        d.nextID = s.nextID
        d.layers = s.layers
        d.nextLayerID = s.nextLayerID
        d.curLayer = s.curLayer
        d.blocks = s.blocks
}

func (d *Document) pushUndo() {
        d.undoStack = append(d.undoStack, d.snapshot())
        if len(d.undoStack) > maxUndoDepth {
                d.undoStack = d.undoStack[len(d.undoStack)-maxUndoDepth:]
        }
        d.redoStack = nil
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

// ─── Primitive Add operations ─────────────────────────────────────────────────

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

// ─── Advanced Add operations (Task #3) ────────────────────────────────────────

// AddSpline adds a cubic Bezier spline defined by control points (chain layout:
// [p0, cp1, cp2, p1, cp3, cp4, p2, …]; need ≥ 4 points for one segment).
// Command: SPLINE  /  Shortcut: S
func (d *Document) AddSpline(points [][]float64, layer int, color string) int {
        return d.add(Entity{Type: TypeSpline, Points: points, Layer: layer, Color: color})
}

// AddNURBS adds a rational B-spline (NURBS) entity.
// degree: B-spline degree (typically 3 = cubic).
// controls: control point coordinates [[x,y], …].
// knots: knot vector (len must equal len(controls)+degree+1); pass nil for
//        auto-generated clamped uniform knots.
// weights: rational weights (len must equal len(controls)); pass nil for all-1
//          uniform weights (i.e. a non-rational B-spline).
// Command: NURBS
func (d *Document) AddNURBS(degree int, controls [][]float64, knots []float64, weights []float64, layer int, color string) int {
        nc := len(controls)
        if degree < 1 {
                degree = 3
        }
        if len(knots) < nc+degree+1 {
                knots = nurbsClampedUniformKnots(nc, degree)
        }
        // Normalise weights: nil or wrong length → all-1 slice.
        if len(weights) != nc {
                w := make([]float64, nc)
                for i := range w {
                        w[i] = 1
                }
                weights = w
        }
        return d.add(Entity{
                Type: TypeNURBS, NURBSDegree: degree,
                Points: controls, Knots: knots, Weights: weights,
                Layer: layer, Color: color,
        })
}

// AddEllipse adds an ellipse entity.
// cx, cy: centre; a: semi-major axis; b: semi-minor axis; rotDeg: rotation
// angle (degrees CCW from +X).
// Command: ELLIPSE  /  Shortcut: E
func (d *Document) AddEllipse(cx, cy, a, b, rotDeg float64, layer int, color string) int {
        return d.add(Entity{Type: TypeEllipse, CX: cx, CY: cy, R: a, R2: b, RotDeg: rotDeg, Layer: layer, Color: color})
}

// AddText adds a single-line text entity at insertion point (x, y).
// height: cap-height in document units; rotDeg: baseline angle (CCW degrees).
// font: SHX or TTF font / text-style name (empty = "Standard").
// Command: TEXT  /  Shortcut: T
func (d *Document) AddText(x, y float64, text string, height, rotDeg float64, font string, layer int, color string) int {
        return d.add(Entity{
                Type: TypeText, X1: x, Y1: y,
                Text: text, TextHeight: height, RotDeg: rotDeg,
                Font:  font,
                Layer: layer, Color: color,
        })
}

// AddMText adds a multi-line text entity.
// x, y: insertion point (top-left corner by default).
// text: content string; use "\n" for paragraph breaks (exported as "\\P" in DXF MTEXT).
// height: character height in document units.
// width: reference rectangle width (0 = no wrapping).
// rotDeg: rotation of the entire text block (CCW degrees).
// font: SHX or TTF font / text-style name (empty = "Standard").
// Command: MTEXT
func (d *Document) AddMText(x, y float64, text string, height, width, rotDeg float64, font string, layer int, color string) int {
        return d.add(Entity{
                Type: TypeMText, X1: x, Y1: y,
                Text: text, TextHeight: height, R2: width, RotDeg: rotDeg,
                Font:  font,
                Layer: layer, Color: color,
        })
}

// AddLinearDim adds a linear (horizontal or vertical) dimension.
// x1,y1 and x2,y2: definition points; offset: perpendicular distance from
// the measurement line to the dimension line (positive = above/left).
// Command: DIMLIN
func (d *Document) AddLinearDim(x1, y1, x2, y2, offset float64, layer int, color string) int {
        return d.add(Entity{Type: TypeDimLinear, X1: x1, Y1: y1, X2: x2, Y2: y2, CX: offset, Layer: layer, Color: color})
}

// AddAlignedDim adds an aligned dimension (along the entity direction).
// offset: perpendicular offset of the dimension line from the measured segment.
// Command: DIMALI
func (d *Document) AddAlignedDim(x1, y1, x2, y2, offset float64, layer int, color string) int {
        return d.add(Entity{Type: TypeDimAligned, X1: x1, Y1: y1, X2: x2, Y2: y2, CX: offset, Layer: layer, Color: color})
}

// AddAngularDim adds an angular dimension between two rays from a common vertex.
// cx, cy: vertex; x1,y1 and x2,y2: points on the two rays; radius: arc radius.
// Command: DIMANG
func (d *Document) AddAngularDim(cx, cy, x1, y1, x2, y2, radius float64, layer int, color string) int {
        return d.add(Entity{Type: TypeDimAngular, CX: cx, CY: cy, X1: x1, Y1: y1, X2: x2, Y2: y2, R: radius, Layer: layer, Color: color})
}

// AddRadialDim adds a radial dimension with a leader at angle rotDeg (degrees).
// Command: DIMRAD
func (d *Document) AddRadialDim(cx, cy, r, angle float64, layer int, color string) int {
        return d.add(Entity{Type: TypeDimRadial, CX: cx, CY: cy, R: r, RotDeg: angle, Layer: layer, Color: color})
}

// AddDiameterDim adds a diameter dimension with a leader at angle rotDeg (degrees).
// Command: DIMDIA
func (d *Document) AddDiameterDim(cx, cy, r, angle float64, layer int, color string) int {
        return d.add(Entity{Type: TypeDimDiameter, CX: cx, CY: cy, R: r, RotDeg: angle, Layer: layer, Color: color})
}

// ─── Task #7 Add operations ────────────────────────────────────────────────────

// AddHatch adds a polygon hatch fill entity.
//
// boundary: closed polygon [[x,y], …].
// pattern: "SOLID", "ANSI31", "ANSI32", "DOTS".
// angleDeg: additional rotation applied to the hatch lines.
// scale: spacing between hatch lines (> 0).
// Command: HATCH
func (d *Document) AddHatch(boundary [][]float64, pattern string, angleDeg, scale float64, layer int, color string) int {
        if scale <= 0 {
                scale = 5
        }
        return d.add(Entity{
                Type:   TypeHatch,
                Points: boundary,
                Text:   pattern,
                RotDeg: angleDeg,
                R:      scale,
                Layer:  layer, Color: color,
        })
}

// AddLeader adds a multi-segment leader annotation.
//
// pts: vertex list [[x,y], …] — first point is the arrowhead tip, last is
// the text attachment point.
// text: annotation label (displayed near the last point).
// Command: LEADER
func (d *Document) AddLeader(pts [][]float64, text string, layer int, color string) int {
        return d.add(Entity{
                Type:   TypeLeader,
                Points: pts,
                Text:   text,
                Layer:  layer, Color: color,
        })
}

// AddRevisionCloud adds a revision cloud entity.
//
// pts: polygon vertices [[x,y], …] (automatically closed).
// arcLength: chord length for the arc bumps (> 0).
// Command: REVCLOUD
func (d *Document) AddRevisionCloud(pts [][]float64, arcLength float64, layer int, color string) int {
        if arcLength <= 0 {
                arcLength = 5
        }
        return d.add(Entity{
                Type:   TypeRevisionCloud,
                Points: pts,
                R:      arcLength,
                Layer:  layer, Color: color,
        })
}

// AddWipeout adds an opaque masking polygon (wipeout).
//
// pts: polygon vertices [[x,y], …] (automatically closed).
// Command: WIPEOUT
func (d *Document) AddWipeout(pts [][]float64, layer int, color string) int {
        return d.add(Entity{
                Type:   TypeWipeout,
                Points: pts,
                Layer:  layer, Color: color,
        })
}

// ─── SetEntityProp ────────────────────────────────────────────────────────────

// SetEntityProp updates a single named field of an entity in-place and pushes
// an undo snapshot. Returns false if the entity or field is not found.
func (d *Document) SetEntityProp(id int, field, value string) bool {
        for i := range d.entities {
                if d.entities[i].ID != id {
                        continue
                }
                d.pushUndo()
                e := &d.entities[i]
                switch field {
                case "color":
                        e.Color = value
                case "layer":
                        var n int
                        if _, err := fmt.Sscanf(value, "%d", &n); err == nil {
                                e.Layer = n
                        }
                case "text":
                        e.Text = value
                case "rotDeg":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.RotDeg = f
                        }
                case "textHeight":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.TextHeight = f
                        }
                case "x1":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.X1 = f
                        }
                case "y1":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.Y1 = f
                        }
                case "x2":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.X2 = f
                        }
                case "y2":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.Y2 = f
                        }
                case "cx":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.CX = f
                        }
                case "cy":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.CY = f
                        }
                case "r":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.R = f
                        }
                case "startDeg":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.StartDeg = f
                        }
                case "endDeg":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.EndDeg = f
                        }
                case "lineType":
                        e.LineType = value
                case "lineWeight":
                        var f float64
                        if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
                                e.LineWeight = f
                        }
                default:
                        return false
                }
                return true
        }
        return false
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
        d.restoreSnapshot(d.undoStack[last])
        d.undoStack = d.undoStack[:last]
        return true
}

func (d *Document) Redo() bool {
        if len(d.redoStack) == 0 {
                return false
        }
        d.undoStack = append(d.undoStack, d.snapshot())
        last := len(d.redoStack) - 1
        d.restoreSnapshot(d.redoStack[last])
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

// AddEntity adds a generic entity, dispatching on its Type field.
// The entity's ID is ignored; a new one is assigned. Returns -1 for unknown types.
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
        case TypeNURBS:
                return d.AddNURBS(e.NURBSDegree, e.Points, e.Knots, e.Weights, e.Layer, e.Color)
        case TypeEllipse:
                return d.AddEllipse(e.CX, e.CY, e.R, e.R2, e.RotDeg, e.Layer, e.Color)
        case TypeText:
                return d.AddText(e.X1, e.Y1, e.Text, e.TextHeight, e.RotDeg, e.Font, e.Layer, e.Color)
        case TypeMText:
                return d.AddMText(e.X1, e.Y1, e.Text, e.TextHeight, e.R2, e.RotDeg, e.Font, e.Layer, e.Color)
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
        case TypeBlockRef:
                return d.InsertBlock(e.Text, e.X1, e.Y1, e.R, e.R2, e.RotDeg, e.Layer, e.Color)
        case TypeHatch:
                return d.AddHatch(e.Points, e.Text, e.RotDeg, e.R, e.Layer, e.Color)
        case TypeLeader:
                return d.AddLeader(e.Points, e.Text, e.Layer, e.Color)
        case TypeRevisionCloud:
                return d.AddRevisionCloud(e.Points, e.R, e.Layer, e.Color)
        case TypeWipeout:
                return d.AddWipeout(e.Points, e.Layer, e.Color)
        default:
                return -1
        }
}

// ─── Persistence ──────────────────────────────────────────────────────────────

// docState is the full on-disk JSON structure (new format).
// Old saves that contain only a JSON entity array are still supported
// via the Load fallback path.
type docState struct {
        Entities    []Entity          `json:"entities"`
        Layers      map[int]*Layer    `json:"layers,omitempty"`
        NextLayerID int               `json:"nextLayerID,omitempty"`
        CurLayer    int               `json:"curLayer,omitempty"`
        Blocks      map[string]*Block `json:"blocks,omitempty"` // Task #7
}

func (d *Document) Save(path string) error {
        state := docState{
                Entities:    d.entities,
                Layers:      d.layers,
                NextLayerID: d.nextLayerID,
                CurLayer:    d.curLayer,
                Blocks:      d.blocks, // Task #7: persist block definitions
        }
        data, err := json.Marshal(state)
        if err != nil {
                return fmt.Errorf("document.Save: %w", err)
        }
        if err := os.WriteFile(path, data, 0o644); err != nil {
                return fmt.Errorf("document.Save: %w", err)
        }
        return nil
}

func (d *Document) Load(path string) error {
        data, err := os.ReadFile(path)
        if err != nil {
                return fmt.Errorf("document.Load: %w", err)
        }

        // Try the new full-document format first (JSON object with "entities" key).
        var state docState
        if err := json.Unmarshal(data, &state); err == nil && state.Entities != nil {
                d.pushUndo()
                d.entities = state.Entities
                if state.Layers != nil {
                        d.layers = state.Layers
                        d.nextLayerID = state.NextLayerID
                        d.curLayer = state.CurLayer
                } else {
                        d.layers = map[int]*Layer{0: defaultLayer0()}
                        d.nextLayerID = 1
                }
                // Task #7: merge saved user blocks, preserving built-in symbols
                // (Builtin=true, registered at startup) that are absent from
                // older saves.  User blocks in the saved state override built-ins
                // with the same name; blocks absent from the saved state and not
                // built-in are removed (preventing stale user-block leakage).
                {
                        merged := make(map[string]*Block)
                        for name, blk := range d.blocks {
                                if blk.Builtin {
                                        merged[name] = blk // keep built-ins
                                }
                        }
                        for name, blk := range state.Blocks {
                                merged[name] = blk // loaded blocks override
                        }
                        if len(merged) > 0 {
                                d.blocks = merged
                        } else {
                                d.blocks = nil
                        }
                }
                for _, e := range d.entities {
                        if e.ID >= d.nextID {
                                d.nextID = e.ID + 1
                        }
                }
                return nil
        }

        // Legacy format: bare JSON array of entities.
        var entities []Entity
        if err := json.Unmarshal(data, &entities); err != nil {
                return fmt.Errorf("document.Load: %w", err)
        }
        d.pushUndo()
        d.entities = entities
        d.layers = map[int]*Layer{0: defaultLayer0()}
        d.nextLayerID = 1
        for _, e := range entities {
                if e.ID >= d.nextID {
                        d.nextID = e.ID + 1
                }
        }
        return nil
}

// ─── DXF import ───────────────────────────────────────────────────────────────

// LoadDXFBytes parses a DXF text payload (R12 or R2000) and replaces the
// current document's entities and layers with what was parsed. The previous
// state is pushed onto the undo stack so the import can be undone.
//
// It delegates to pkg/dxf.Read; the import is done via a function variable so
// that the internal/document package does not depend on pkg/dxf (avoiding a
// circular import). Call document.RegisterDXFReader from cmd/wasm or cmd/cad
// to wire the implementation.
func (d *Document) LoadDXFBytes(data []byte) (warnings []string, err error) {
        if dxfReader == nil {
                return nil, fmt.Errorf("document.LoadDXFBytes: no DXF reader registered (call document.RegisterDXFReader)")
        }
        imported, warns, err := dxfReader(data)
        if err != nil {
                return warns, fmt.Errorf("document.LoadDXFBytes: %w", err)
        }
        d.pushUndo()
        d.entities = imported.entities
        d.nextID = imported.nextID
        d.layers = imported.layers
        d.nextLayerID = imported.nextLayerID
        d.curLayer = 0
        return warns, nil
}

// dxfReader is set by RegisterDXFReader to break the import cycle.
var dxfReader func([]byte) (*Document, []string, error)

// RegisterDXFReader registers the DXF parsing function. Must be called once at
// startup (typically in cmd/wasm/main.go init() or cmd/cad/main.go).
func RegisterDXFReader(fn func([]byte) (*Document, []string, error)) {
        dxfReader = fn
}

// ─── DXF export ───────────────────────────────────────────────────────────────

// ExportDXF returns a DXF R2000 (AC1015) string for all entities.
// Splines are exported as LWPOLYLINE approximations; dimensions as proper
// DIMENSION entities with AcDb subclass markers; ellipses as ELLIPSE entities.
// Y-axis is flipped (DXF Cartesian vs. screen coordinates).
func (d *Document) ExportDXF() string {
        return d.exportDXF(false)
}

// ExportDXFR12 returns a DXF R12 (AC1009) string for all entities.
// All entities are reduced to R12-compatible primitives (LINE, CIRCLE, ARC,
// TEXT, POLYLINE+VERTEX+SEQEND) so the file loads in legacy applications such
// as AutoCAD R12, early QCAD, and embedded controllers.
//
// Differences from ExportDXF (R2000):
//   - Version header: AC1009 instead of AC1015
//   - Splines/NURBS: POLYLINE+VERTEX+SEQEND instead of LWPOLYLINE
//   - Ellipses: POLYLINE approximation instead of native ELLIPSE entity
//   - Dimensions: LINE + TEXT helper geometry instead of DIMENSION entities
//   - MTEXT: approximated as multiple TEXT entities (one per line)
func (d *Document) ExportDXFR12() string {
        return d.exportDXF(true)
}

func (d *Document) exportDXF(r12 bool) string {
        ver := "AC1015"
        if r12 {
                ver = "AC1009"
        }
        var sb strings.Builder
        fmt.Fprintf(&sb, "  0\nSECTION\n  2\nHEADER\n  9\n$ACADVER\n  1\n%s\n  0\nENDSEC\n", ver)

        // TABLES section: layer table
        d.writeDXFLayerTable(&sb, r12)

        // BLOCKS section: emit BLOCK/ENDBLK definitions so INSERT entities resolve.
        // The mandatory *Model_Space block is always written; user blocks follow.
        sb.WriteString("  0\nSECTION\n  2\nBLOCKS\n")
        // Mandatory model-space block required by all DXF readers.
        if r12 {
                sb.WriteString("  0\nBLOCK\n  8\n0\n  2\n*Model_Space\n 70\n0\n 10\n0.0\n 20\n0.0\n  0\nENDBLK\n")
        } else {
                sb.WriteString("  0\nBLOCK\n  8\n0\n100\nAcDbEntity\n100\nAcDbBlockBegin\n  2\n*Model_Space\n 70\n0\n 10\n0.0\n 20\n0.0\n 30\n0.0\n  3\n*Model_Space\n  1\n\n  0\nENDBLK\n  8\n0\n100\nAcDbEntity\n100\nAcDbBlockEnd\n")
        }
        for _, blk := range d.blocks {
                ln0 := "0"
                // Entities are stored in block-local coordinates (base = origin),
                // so the DXF BLOCK base point is always (0,0). Using the original
                // world BaseX/BaseY here would double-offset INSERT placement.
                if r12 {
                        fmt.Fprintf(&sb, "  0\nBLOCK\n  8\n%s\n  2\n%s\n 70\n0\n 10\n0.0\n 20\n0.0\n",
                                ln0, blk.Name)
                } else {
                        fmt.Fprintf(&sb, "  0\nBLOCK\n  8\n%s\n100\nAcDbEntity\n100\nAcDbBlockBegin\n  2\n%s\n 70\n0\n 10\n0.0\n 20\n0.0\n 30\n0.0\n  3\n%s\n  1\n\n",
                                ln0, blk.Name, blk.Name)
                }
                // Emit each block-local entity (simplified: lines, circles, arcs, polylines, text).
                for _, be := range blk.Entities {
                        switch be.Type {
                        case TypeLine:
                                fmt.Fprintf(&sb, "  0\nLINE\n  8\n%s\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
                                        ln0, be.X1, -be.Y1, be.X2, -be.Y2)
                        case TypeCircle:
                                fmt.Fprintf(&sb, "  0\nCIRCLE\n  8\n%s\n 10\n%f\n 20\n%f\n 40\n%f\n",
                                        ln0, be.CX, -be.CY, be.R)
                        case TypeArc:
                                fmt.Fprintf(&sb, "  0\nARC\n  8\n%s\n 10\n%f\n 20\n%f\n 40\n%f\n 50\n%f\n 51\n%f\n",
                                        ln0, be.CX, -be.CY, be.R, be.StartDeg, be.EndDeg)
                        case TypeText:
                                h := be.TextHeight
                                if h <= 0 {
                                        h = 2.5
                                }
                                fmt.Fprintf(&sb, "  0\nTEXT\n  8\n%s\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n%f\n  1\n%s\n",
                                        ln0, be.X1, -be.Y1, h, be.Text)
                        case TypePolyline:
                                pts := make([][2]float64, len(be.Points))
                                for i, p := range be.Points {
                                        if len(p) >= 2 {
                                                pts[i] = [2]float64{p[0], p[1]}
                                        }
                                }
                                if r12 {
                                        dxfR12Polyline(&sb, ln0, pts)
                                } else {
                                        dxfLWPolyline(&sb, ln0, pts)
                                }
                        }
                }
                if r12 {
                        sb.WriteString("  0\nENDBLK\n")
                } else {
                        fmt.Fprintf(&sb, "  0\nENDBLK\n  8\n%s\n100\nAcDbEntity\n100\nAcDbBlockEnd\n", ln0)
                }
        }
        sb.WriteString("  0\nENDSEC\n")

        sb.WriteString("  0\nSECTION\n  2\nENTITIES\n")

        for _, e := range d.entities {
                ln := d.layerName(e.Layer) // layer name string for group code 8
                switch e.Type {
                case TypeLine:
                        fmt.Fprintf(&sb, "  0\nLINE\n  8\n%s\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
                                ln, e.X1, -e.Y1, e.X2, -e.Y2)

                case TypeCircle:
                        fmt.Fprintf(&sb, "  0\nCIRCLE\n  8\n%s\n 10\n%f\n 20\n%f\n 40\n%f\n",
                                ln, e.CX, -e.CY, e.R)

                case TypeArc:
                        fmt.Fprintf(&sb, "  0\nARC\n  8\n%s\n 10\n%f\n 20\n%f\n 40\n%f\n 50\n%f\n 51\n%f\n",
                                ln, e.CX, -e.CY, e.R, e.StartDeg, e.EndDeg)

                case TypeRectangle:
                        x1, y1, x2, y2 := e.X1, e.Y1, e.X2, e.Y2
                        dxfLine(&sb, ln, x1, y1, x2, y1)
                        dxfLine(&sb, ln, x2, y1, x2, y2)
                        dxfLine(&sb, ln, x2, y2, x1, y2)
                        dxfLine(&sb, ln, x1, y2, x1, y1)

                case TypePolyline:
                        dxfPolylineLines(&sb, ln, e.Points)

                case TypeSpline:
                        pts := approxBezierPoints(e.Points, 20)
                        if r12 {
                                dxfR12Polyline(&sb, ln, pts)
                        } else {
                                dxfLWPolyline(&sb, ln, pts)
                        }

                case TypeNURBS:
                        pts := nurbsApprox(e, 50)
                        pts2 := make([][2]float64, len(pts))
                        copy(pts2, pts)
                        if r12 {
                                dxfR12Polyline(&sb, ln, pts2)
                        } else {
                                dxfLWPolyline(&sb, ln, pts2)
                        }

                case TypeEllipse:
                        if r12 {
                                el := ellipseApprox(e.CX, e.CY, e.R, e.R2, e.RotDeg, 72)
                                dxfR12Polyline(&sb, ln, el)
                        } else {
                                rot := e.RotDeg * math.Pi / 180
                                mx := e.R * math.Cos(rot)
                                my := e.R * math.Sin(rot)
                                ratio := 1.0
                                if e.R > 1e-12 {
                                        ratio = e.R2 / e.R
                                }
                                fmt.Fprintf(&sb,
                                        "  0\nELLIPSE\n  8\n%s\n 10\n%f\n 20\n%f\n 30\n0.0\n 11\n%f\n 21\n%f\n 31\n0.0\n 40\n%f\n 41\n0.0\n 42\n%f\n",
                                        ln, e.CX, -e.CY, mx, -my, ratio, 2*math.Pi)
                        }

                case TypeText:
                        h := e.TextHeight
                        if h <= 0 {
                                h = 2.5
                        }
                        style := e.Font
                        if style == "" {
                                style = "Standard"
                        }
                        fmt.Fprintf(&sb,
                                "  0\nTEXT\n  8\n%s\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n%f\n  1\n%s\n 50\n%f\n  7\n%s\n",
                                ln, e.X1, -e.Y1, h, e.Text, e.RotDeg, style)

                case TypeMText:
                        h := e.TextHeight
                        if h <= 0 {
                                h = 2.5
                        }
                        width := e.R2
                        style := e.Font
                        if style == "" {
                                style = "Standard"
                        }
                        if r12 {
                                lines := strings.Split(e.Text, "\n")
                                for i, line := range lines {
                                        fmt.Fprintf(&sb,
                                                "  0\nTEXT\n  8\n%s\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n%f\n  1\n%s\n 50\n%f\n  7\n%s\n",
                                                ln, e.X1, -(e.Y1-float64(i)*h*1.5), h,
                                                line, e.RotDeg, style)
                                }
                        } else {
                                content := strings.ReplaceAll(e.Text, "\n", "\\P")
                                fmt.Fprintf(&sb,
                                        "  0\nMTEXT\n  8\n%s\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n%f\n 41\n%f\n 71\n1\n 72\n1\n  1\n%s\n  7\n%s\n 50\n%f\n",
                                        ln, e.X1, -e.Y1, h, width, content, style, e.RotDeg)
                        }

                case TypeDimLinear:
                        if r12 {
                                dxfLinearDimLines(&sb, e, ln)
                        } else {
                                dxfDimensionLinear(&sb, e, ln)
                        }

                case TypeDimAligned:
                        if r12 {
                                dxfAlignedDimLines(&sb, e, ln)
                        } else {
                                dxfDimensionAligned(&sb, e, ln)
                        }

                case TypeDimAngular:
                        if r12 {
                                dxfAngularDimLines(&sb, e, ln)
                        } else {
                                dxfDimensionAngular(&sb, e, ln)
                        }

                case TypeDimRadial:
                        if r12 {
                                dxfRadialDimLines(&sb, e, ln)
                        } else {
                                dxfDimensionRadial(&sb, e, ln)
                        }

                case TypeDimDiameter:
                        if r12 {
                                dxfDiameterDimLines(&sb, e, ln)
                        } else {
                                dxfDimensionDiameter(&sb, e, ln)
                        }

                // ── Task #7 types ─────────────────────────────────────────────────
                case TypeBlockRef:
                        // Export as INSERT entity (R2000) or INSERT (R12).
                        sx, sy := e.R, e.R2
                        if sx == 0 {
                                sx = 1
                        }
                        if sy == 0 {
                                sy = 1
                        }
                        if r12 {
                                fmt.Fprintf(&sb, "  0\nINSERT\n  8\n%s\n  2\n%s\n 10\n%f\n 20\n%f\n 41\n%f\n 42\n%f\n 50\n%f\n",
                                        ln, e.Text, e.X1, -e.Y1, sx, sy, e.RotDeg)
                        } else {
                                fmt.Fprintf(&sb, "  0\nINSERT\n  8\n%s\n100\nAcDbEntity\n100\nAcDbBlockReference\n  2\n%s\n 10\n%f\n 20\n%f\n 30\n0.0\n 41\n%f\n 42\n%f\n 50\n%f\n",
                                        ln, e.Text, e.X1, -e.Y1, sx, sy, e.RotDeg)
                        }

                case TypeHatch:
                        if !r12 {
                                // R2000 HATCH entity.
                                n := len(e.Points)
                                if n < 3 {
                                        break
                                }
                                solid := strings.ToUpper(e.Text) == "SOLID"
                                gradFill := 0
                                if solid {
                                        gradFill = 1
                                }
                                _ = gradFill
                                patName := e.Text
                                if patName == "" {
                                        patName = "ANSI31"
                                }
                                fmt.Fprintf(&sb, "  0\nHATCH\n  8\n%s\n100\nAcDbEntity\n100\nAcDbHatch\n 10\n0.0\n 20\n0.0\n 30\n0.0\n 210\n0.0\n 220\n0.0\n 230\n1.0\n  2\n%s\n 70\n%d\n 71\n0\n 91\n1\n 92\n3\n 73\n1\n 93\n%d\n",
                                        ln, patName, btoi(solid), n)
                                for _, p := range e.Points {
                                        fmt.Fprintf(&sb, " 10\n%f\n 20\n%f\n", p[0], -p[1])
                                }
                                fmt.Fprintf(&sb, " 75\n0\n 76\n1\n 52\n%f\n 41\n%f\n 77\n0\n 78\n0\n 47\n0.0\n 98\n0\n",
                                        e.RotDeg, e.R)
                        } else {
                                // R12: export as polyline boundary + SOLID fill approximation.
                                pts := make([][2]float64, len(e.Points))
                                for i, p := range e.Points {
                                        if len(p) >= 2 {
                                                pts[i] = [2]float64{p[0], p[1]}
                                        }
                                }
                                dxfR12Polyline(&sb, ln, pts)
                        }

                case TypeLeader:
                        if len(e.Points) < 2 {
                                break
                        }
                        if r12 {
                                // R12: series of lines + text.
                                for i := 0; i < len(e.Points)-1; i++ {
                                        dxfLine(&sb, ln, e.Points[i][0], e.Points[i][1],
                                                e.Points[i+1][0], e.Points[i+1][1])
                                }
                                last := e.Points[len(e.Points)-1]
                                if e.Text != "" {
                                        dxfText(&sb, ln, last[0]+1, last[1], 2.5, e.Text, "Standard")
                                }
                        } else {
                                // R2000 LEADER entity.
                                fmt.Fprintf(&sb, "  0\nLEADER\n  8\n%s\n100\nAcDbEntity\n100\nAcDbLeader\n  3\nStandard\n 71\n1\n 72\n0\n 73\n3\n 74\n1\n 75\n0\n 76\n%d\n",
                                        ln, len(e.Points))
                                for _, p := range e.Points {
                                        fmt.Fprintf(&sb, " 10\n%f\n 20\n%f\n 30\n0.0\n", p[0], -p[1])
                                }
                                if e.Text != "" {
                                        last := e.Points[len(e.Points)-1]
                                        dxfText(&sb, ln, last[0]+1, last[1], 2.5, e.Text, "Standard")
                                }
                        }

                case TypeRevisionCloud:
                        // Revision cloud: closed poly with concave arc per edge (bulge=-0.4142 ≈ 90°).
                        pts := make([][2]float64, len(e.Points))
                        for i, p := range e.Points {
                                if len(p) >= 2 {
                                        pts[i] = [2]float64{p[0], p[1]}
                                }
                        }
                        if r12 {
                                dxfR12PolylineRevCloud(&sb, ln, pts)
                        } else {
                                dxfLWPolylineRevCloud(&sb, ln, pts)
                        }

                case TypeWipeout:
                        // Wipeout: plain closed polygon (no arc bulge needed).
                        pts := make([][2]float64, len(e.Points))
                        for i, p := range e.Points {
                                if len(p) >= 2 {
                                        pts[i] = [2]float64{p[0], p[1]}
                                }
                        }
                        if r12 {
                                dxfR12Polyline(&sb, ln, pts)
                        } else {
                                dxfLWPolyline(&sb, ln, pts)
                        }
                }
        }

        sb.WriteString("  0\nENDSEC\n  0\nEOF\n")
        return sb.String()
}

// btoi returns 1 if b is true, 0 otherwise (DXF boolean group codes).
func btoi(b bool) int {
        if b {
                return 1
        }
        return 0
}

// ─── DXF primitive helpers ────────────────────────────────────────────────────

func dxfLine(sb *strings.Builder, layer string, x1, y1, x2, y2 float64) {
        fmt.Fprintf(sb, "  0\nLINE\n  8\n%s\n 10\n%f\n 20\n%f\n 11\n%f\n 21\n%f\n",
                layer, x1, -y1, x2, -y2)
}

func dxfText(sb *strings.Builder, layer string, x, y, h float64, text string, style string) {
        if style == "" {
                style = "Standard"
        }
        fmt.Fprintf(sb, "  0\nTEXT\n  8\n%s\n 10\n%f\n 20\n%f\n 30\n0.0\n 40\n%f\n  1\n%s\n 50\n0.0\n  7\n%s\n",
                layer, x, -y, h, text, style)
}

// dxfPolylineLines emits a series of LINE entities for a polyline.
func dxfPolylineLines(sb *strings.Builder, layer string, pts [][]float64) {
        for i := 0; i < len(pts)-1; i++ {
                dxfLine(sb, layer, pts[i][0], pts[i][1], pts[i+1][0], pts[i+1][1])
        }
}

// dxfR12Polyline emits a POLYLINE + VERTEX + SEQEND block (R12 compatible).
func dxfR12Polyline(sb *strings.Builder, layer string, pts [][2]float64) {
        if len(pts) < 2 {
                return
        }
        fmt.Fprintf(sb, "  0\nPOLYLINE\n  8\n%s\n 66\n1\n 70\n0\n", layer)
        for _, p := range pts {
                fmt.Fprintf(sb, "  0\nVERTEX\n  8\n%s\n 10\n%f\n 20\n%f\n", layer, p[0], -p[1])
        }
        sb.WriteString("  0\nSEQEND\n")
}

// dxfLWPolyline emits an LWPOLYLINE entity (R2000+).
func dxfLWPolyline(sb *strings.Builder, layer string, pts [][2]float64) {
        if len(pts) < 2 {
                return
        }
        fmt.Fprintf(sb, "  0\nLWPOLYLINE\n  8\n%s\n 90\n%d\n 70\n0\n", layer, len(pts))
        for _, p := range pts {
                fmt.Fprintf(sb, " 10\n%f\n 20\n%f\n", p[0], -p[1])
        }
}

// dxfLWPolylineRevCloud emits a closed LWPOLYLINE (R2000+) with a bulge value
// on every vertex so that CAD readers draw concave arcs between vertices —
// the characteristic revision-cloud appearance.
// bulge = -tan(θ/4) where θ is the included angle; -0.4142 ≈ -tan(π/8) ≈ 90°
// clockwise (concave when the polygon is CCW).
func dxfLWPolylineRevCloud(sb *strings.Builder, layer string, pts [][2]float64) {
        if len(pts) < 3 {
                return
        }
        const bulge = -0.4142 // concave 90° arc per edge
        fmt.Fprintf(sb, "  0\nLWPOLYLINE\n  8\n%s\n 90\n%d\n 70\n1\n", layer, len(pts))
        for _, p := range pts {
                fmt.Fprintf(sb, " 10\n%f\n 20\n%f\n 42\n%f\n", p[0], -p[1], bulge)
        }
}

// dxfR12PolylineRevCloud emits a closed POLYLINE+VERTEX+SEQEND (R12 compatible)
// with bulge values so downstream readers render the revision-cloud arcs.
func dxfR12PolylineRevCloud(sb *strings.Builder, layer string, pts [][2]float64) {
        if len(pts) < 3 {
                return
        }
        const bulge = -0.4142
        fmt.Fprintf(sb, "  0\nPOLYLINE\n  8\n%s\n 66\n1\n 70\n1\n", layer)
        for _, p := range pts {
                fmt.Fprintf(sb, "  0\nVERTEX\n  8\n%s\n 10\n%f\n 20\n%f\n 42\n%f\n",
                        layer, p[0], -p[1], bulge)
        }
        sb.WriteString("  0\nSEQEND\n")
}

// ─── Bezier approximation helper ──────────────────────────────────────────────

func approxBezierPoints(pts [][]float64, n int) [][2]float64 {
        nCtrl := len(pts)
        if nCtrl < 4 {
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
        last := pts[len(pts)-1]
        out = append(out, [2]float64{last[0], last[1]})
        return out
}

// ─── Ellipse approximation helper ────────────────────────────────────────────

func ellipseApprox(cx, cy, a, b, rotDeg float64, n int) [][2]float64 {
        rot := rotDeg * math.Pi / 180
        cosR, sinR := math.Cos(rot), math.Sin(rot)
        pts := make([][2]float64, n+1)
        for i := 0; i <= n; i++ {
                theta := 2 * math.Pi * float64(i) / float64(n)
                lx := a * math.Cos(theta)
                ly := b * math.Sin(theta)
                pts[i] = [2]float64{
                        cx + lx*cosR - ly*sinR,
                        cy + lx*sinR + ly*cosR,
                }
        }
        return pts
}

// ─── R2000 DIMENSION entities ─────────────────────────────────────────────────
//
// Group-code layout used here:
//   0  DIMENSION
//   8  layer
//  100 AcDbEntity
//  100 AcDbDimension
//    3 style name ("Standard")
//   70 type flags (0=rotated/linear, 1=aligned, 2=angular3pt, 3=diameter, 4=radial)
//   10,20,30  dimension-line definition point
//   11,21,31  text insertion point
//    1 measurement text override (empty = auto)
//   42 actual measurement value
//  100 AcDb<Type>Dimension  (subclass marker)
//   13,23,33  first extension line origin
//   14,24,34  second extension line origin
//   50 rotation angle (AcDbRotatedDimension only)
// ─────────────────────────────────────────────────────────────────────────────

func dxfDimHeader(sb *strings.Builder, layer string, dimType int,
        dlX, dlY, textX, textY, measurement float64) {
        // AcDbDimension header common to all dimension types.
        fmt.Fprintf(sb, "  0\nDIMENSION\n  8\n%s\n100\nAcDbEntity\n100\nAcDbDimension\n  3\nStandard\n 70\n%d\n",
                layer, dimType)
        fmt.Fprintf(sb, " 10\n%f\n 20\n%f\n 30\n0.0\n", dlX, -dlY)
        fmt.Fprintf(sb, " 11\n%f\n 21\n%f\n 31\n0.0\n", textX, -textY)
        fmt.Fprintf(sb, "  1\n\n 42\n%f\n", measurement)
}

// dxfDimensionLinear emits a DIMENSION entity for TypeDimLinear.
func dxfDimensionLinear(sb *strings.Builder, e Entity, layer string) {
        dx := e.X2 - e.X1
        dy := e.Y2 - e.Y1
        isHoriz := math.Abs(dx) >= math.Abs(dy)
        off := e.CX
        measured := math.Abs(dx)
        rotAngle := 0.0
        var dlX, dlY float64
        if isHoriz {
                dlX = (e.X1 + e.X2) / 2
                dlY = (e.Y1+e.Y2)/2 - off
        } else {
                measured = math.Abs(dy)
                rotAngle = 90
                dlX = (e.X1+e.X2)/2 - off
                dlY = (e.Y1 + e.Y2) / 2
        }
        dxfDimHeader(sb, layer, 0, dlX, dlY, dlX, dlY, measured)
        sb.WriteString("100\nAcDbAlignedDimension\n")
        fmt.Fprintf(sb, " 13\n%f\n 23\n%f\n 33\n0.0\n", e.X1, -e.Y1)
        fmt.Fprintf(sb, " 14\n%f\n 24\n%f\n 34\n0.0\n", e.X2, -e.Y2)
        sb.WriteString("100\nAcDbRotatedDimension\n")
        fmt.Fprintf(sb, " 50\n%f\n", rotAngle)
}

// dxfDimensionAligned emits a DIMENSION entity for TypeDimAligned.
func dxfDimensionAligned(sb *strings.Builder, e Entity, layer string) {
        dist := math.Hypot(e.X2-e.X1, e.Y2-e.Y1)
        off := e.CX
        ux, uy := e.X2-e.X1, e.Y2-e.Y1
        if dist > 1e-12 {
                ux /= dist
                uy /= dist
        }
        dlX := (e.X1+e.X2)/2 + (-uy)*off
        dlY := (e.Y1+e.Y2)/2 + ux*off
        dxfDimHeader(sb, layer, 1, dlX, dlY, dlX, dlY, dist)
        sb.WriteString("100\nAcDbAlignedDimension\n")
        fmt.Fprintf(sb, " 13\n%f\n 23\n%f\n 33\n0.0\n", e.X1, -e.Y1)
        fmt.Fprintf(sb, " 14\n%f\n 24\n%f\n 34\n0.0\n", e.X2, -e.Y2)
}

// dxfDimensionAngular emits a DIMENSION entity for TypeDimAngular.
func dxfDimensionAngular(sb *strings.Builder, e Entity, layer string) {
        ang1 := math.Atan2(e.Y1-e.CY, e.X1-e.CX)
        ang2 := math.Atan2(e.Y2-e.CY, e.X2-e.CX)
        span := ang2 - ang1
        if span < 0 {
                span += 2 * math.Pi
        }
        angDeg := span * 180 / math.Pi
        midAng := ang1 + span/2
        r := e.R
        if r <= 0 {
                r = math.Min(math.Hypot(e.X1-e.CX, e.Y1-e.CY), math.Hypot(e.X2-e.CX, e.Y2-e.CY)) * 0.5
        }
        dlX := e.CX + r*math.Cos(midAng)
        dlY := e.CY + r*math.Sin(midAng)
        dxfDimHeader(sb, layer, 2, dlX, dlY, dlX, dlY, angDeg)
        sb.WriteString("100\nAcDb3PointAngularDimension\n")
        fmt.Fprintf(sb, " 13\n%f\n 23\n%f\n 33\n0.0\n", e.X1, -e.Y1)
        fmt.Fprintf(sb, " 14\n%f\n 24\n%f\n 34\n0.0\n", e.X2, -e.Y2)
        fmt.Fprintf(sb, " 15\n%f\n 25\n%f\n 35\n0.0\n", e.CX, -e.CY)
        fmt.Fprintf(sb, " 16\n%f\n 26\n%f\n 36\n0.0\n", dlX, -dlY)
}

// dxfDimensionRadial emits a DIMENSION entity for TypeDimRadial.
func dxfDimensionRadial(sb *strings.Builder, e Entity, layer string) {
        ang := e.RotDeg * math.Pi / 180
        px := e.CX + e.R*math.Cos(ang)
        py := e.CY + e.R*math.Sin(ang)
        dxfDimHeader(sb, layer, 4, px, py, px+(e.R+4)*math.Cos(ang), py+(e.R+4)*math.Sin(ang), e.R)
        sb.WriteString("100\nAcDbRadialDimension\n")
        fmt.Fprintf(sb, " 15\n%f\n 25\n%f\n 35\n0.0\n", e.CX, -e.CY)
        fmt.Fprintf(sb, " 40\n%f\n", e.R)
}

// dxfDimensionDiameter emits a DIMENSION entity for TypeDimDiameter.
func dxfDimensionDiameter(sb *strings.Builder, e Entity, layer string) {
        ang := e.RotDeg * math.Pi / 180
        px := e.CX + e.R*math.Cos(ang)
        py := e.CY + e.R*math.Sin(ang)
        px2 := e.CX - e.R*math.Cos(ang)
        py2 := e.CY - e.R*math.Sin(ang)
        dxfDimHeader(sb, layer, 3, px, py, e.CX+(e.R+4)*math.Cos(ang), e.CY+(e.R+4)*math.Sin(ang), 2*e.R)
        sb.WriteString("100\nAcDbDiametricDimension\n")
        fmt.Fprintf(sb, " 15\n%f\n 25\n%f\n 35\n0.0\n", px2, -py2)
        fmt.Fprintf(sb, " 40\n%f\n", e.R)
}

// ─── R12 dimension helpers (LINE + TEXT approximations) ───────────────────────

func dxfArrowLine(sb *strings.Builder, layer string, px, py, dx, dy float64) {
        alen := 3.0
        mag := math.Hypot(dx, dy)
        if mag < 1e-12 {
                return
        }
        dx /= mag
        dy /= mag
        for _, sign := range []float64{1, -1} {
                ang := math.Atan2(dy, dx) + math.Pi + sign*math.Pi/9
                dxfLine(sb, layer, px, py, px+math.Cos(ang)*alen, py+math.Sin(ang)*alen)
        }
}

func dxfLinearDimLines(sb *strings.Builder, e Entity, layer string) {
        dx := e.X2 - e.X1
        dy := e.Y2 - e.Y1
        off := e.CX
        if math.Abs(dx) >= math.Abs(dy) {
                dimY := (e.Y1+e.Y2)/2 - off
                dxfLine(sb, layer, e.X1, e.Y1, e.X1, dimY)
                dxfLine(sb, layer, e.X2, e.Y2, e.X2, dimY)
                dxfLine(sb, layer, e.X1, dimY, e.X2, dimY)
                dxfArrowLine(sb, layer, e.X1, dimY, e.X2-e.X1, 0)
                dxfArrowLine(sb, layer, e.X2, dimY, e.X1-e.X2, 0)
                mid := (e.X1 + e.X2) / 2
                dxfText(sb, layer, mid, dimY-2, 2.5, fmt.Sprintf("%.3f", math.Abs(dx)), "Standard")
        } else {
                dimX := (e.X1+e.X2)/2 - off
                dxfLine(sb, layer, e.X1, e.Y1, dimX, e.Y1)
                dxfLine(sb, layer, e.X2, e.Y2, dimX, e.Y2)
                dxfLine(sb, layer, dimX, e.Y1, dimX, e.Y2)
                dxfArrowLine(sb, layer, dimX, e.Y1, 0, e.Y2-e.Y1)
                dxfArrowLine(sb, layer, dimX, e.Y2, 0, e.Y1-e.Y2)
                mid := (e.Y1 + e.Y2) / 2
                dxfText(sb, layer, dimX-3, mid, 2.5, fmt.Sprintf("%.3f", math.Abs(dy)), "Standard")
        }
}

func dxfAlignedDimLines(sb *strings.Builder, e Entity, layer string) {
        dx := e.X2 - e.X1
        dy := e.Y2 - e.Y1
        dist := math.Hypot(dx, dy)
        if dist < 1e-12 {
                return
        }
        ux, uy := dx/dist, dy/dist
        px, py := -uy, ux
        off := e.CX
        d1x, d1y := e.X1+px*off, e.Y1+py*off
        d2x, d2y := e.X2+px*off, e.Y2+py*off
        dxfLine(sb, layer, e.X1, e.Y1, d1x, d1y)
        dxfLine(sb, layer, e.X2, e.Y2, d2x, d2y)
        dxfLine(sb, layer, d1x, d1y, d2x, d2y)
        dxfArrowLine(sb, layer, d1x, d1y, d2x-d1x, d2y-d1y)
        dxfArrowLine(sb, layer, d2x, d2y, d1x-d2x, d1y-d2y)
        dxfText(sb, layer, (d1x+d2x)/2, (d1y+d2y)/2+2, 2.5, fmt.Sprintf("%.3f", dist), "Standard")
}

func dxfAngularDimLines(sb *strings.Builder, e Entity, layer string) {
        ang1 := math.Atan2(e.Y1-e.CY, e.X1-e.CX)
        ang2 := math.Atan2(e.Y2-e.CY, e.X2-e.CX)
        r := e.R
        if r <= 0 {
                r = math.Min(math.Hypot(e.X1-e.CX, e.Y1-e.CY), math.Hypot(e.X2-e.CX, e.Y2-e.CY)) * 0.5
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
                dxfLine(sb, layer, prev[0], prev[1], cur[0], cur[1])
                prev = cur
        }
        dxfLine(sb, layer, e.CX, e.CY, e.CX+r*math.Cos(ang1), e.CY+r*math.Sin(ang1))
        dxfLine(sb, layer, e.CX, e.CY, e.CX+r*math.Cos(ang2), e.CY+r*math.Sin(ang2))
        midAng := ang1 + span/2
        angDeg := span * 180 / math.Pi
        tx := e.CX + r*1.3*math.Cos(midAng)
        ty := e.CY + r*1.3*math.Sin(midAng)
        dxfText(sb, layer, tx, ty, 2.5, fmt.Sprintf("%.1f°", angDeg), "Standard")
}

func dxfRadialDimLines(sb *strings.Builder, e Entity, layer string) {
        ang := e.RotDeg * math.Pi / 180
        px := e.CX + e.R*math.Cos(ang)
        py := e.CY + e.R*math.Sin(ang)
        dxfLine(sb, layer, e.CX, e.CY, px, py)
        dxfArrowLine(sb, layer, px, py, px-e.CX, py-e.CY)
        dxfText(sb, layer, e.CX+(e.R+4)*math.Cos(ang), e.CY+(e.R+4)*math.Sin(ang), 2.5,
                fmt.Sprintf("R%.3f", e.R), "Standard")
}

func dxfDiameterDimLines(sb *strings.Builder, e Entity, layer string) {
        ang := e.RotDeg * math.Pi / 180
        p1x := e.CX + e.R*math.Cos(ang)
        p1y := e.CY + e.R*math.Sin(ang)
        p2x := e.CX - e.R*math.Cos(ang)
        p2y := e.CY - e.R*math.Sin(ang)
        dxfLine(sb, layer, p1x, p1y, p2x, p2y)
        dxfArrowLine(sb, layer, p1x, p1y, p1x-p2x, p1y-p2y)
        dxfArrowLine(sb, layer, p2x, p2y, p2x-p1x, p2y-p1y)
        dxfText(sb, layer, e.CX+(e.R+4)*math.Cos(ang), e.CY+(e.R+4)*math.Sin(ang), 2.5,
                fmt.Sprintf("⌀%.3f", 2*e.R), "Standard")
}
