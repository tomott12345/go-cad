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
        TypeLine      = "line"
        TypeCircle    = "circle"
        TypeArc       = "arc"
        TypeRectangle = "rectangle"
        TypePolyline  = "polyline"
)

// Entity represents any CAD primitive stored in a document.
// All fields are always serialised (no omitempty) so that zero-valued
// coordinates are preserved correctly in JSON.
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

// ExportDXF returns a DXF R12-compatible string for all entities.
// Y-axis is flipped (DXF uses Cartesian, canvas uses screen coordinates).
func (d *Document) ExportDXF() string {
        var sb strings.Builder
        sb.WriteString("  0\nSECTION\n  2\nHEADER\n  9\n$ACADVER\n  1\nAC1009\n  0\nENDSEC\n")
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
                }
        }
        sb.WriteString("  0\nENDSEC\n  0\nEOF\n")
        return sb.String()
}
