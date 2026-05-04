package document

import (
        "fmt"
        "math"
        "sort"
        "strings"
)

// ─── LineType ────────────────────────────────────────────────────────────────

// LineType represents a CAD line type (dash pattern), matching AutoCAD/QCAD names.
type LineType string

const (
        LineTypeSolid   LineType = "Solid"
        LineTypeDashed  LineType = "Dashed"
        LineTypeDotted  LineType = "Dotted"
        LineTypeDashDot LineType = "DashDot"
        LineTypeCenter  LineType = "Center"
        LineTypeHidden  LineType = "Hidden"
)

// ─── Layer ───────────────────────────────────────────────────────────────────

// Layer represents a full CAD layer with properties matching QCAD/AutoCAD conventions.
type Layer struct {
        ID           int      `json:"id"`
        Name         string   `json:"name"`
        Color        string   `json:"color"`        // RGB hex, e.g. "#00ff00"
        LineTyp      LineType `json:"lineType"`     // Solid/Dashed/Dotted/DashDot/Center/Hidden
        LineWeight   float64  `json:"lineWeight"`   // mm; standard AutoCAD weights 0.00–2.11
        Visible      bool     `json:"visible"`
        Locked       bool     `json:"locked"`
        Frozen       bool     `json:"frozen"`
        PrintEnabled bool     `json:"printEnabled"`
}

// defaultLayer0 returns the mandatory default layer "0".
func defaultLayer0() *Layer {
        return &Layer{
                ID: 0, Name: "0", Color: "#ffffff",
                LineTyp: LineTypeSolid, LineWeight: 0.25,
                Visible: true, PrintEnabled: true,
        }
}

// ─── Document layer accessors ─────────────────────────────────────────────────

// Layers returns a copy of all layers sorted by ID.
func (d *Document) Layers() []*Layer {
        out := make([]*Layer, 0, len(d.layers))
        for _, l := range d.layers {
                cp := *l
                out = append(out, &cp)
        }
        sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
        return out
}

// LayerByID returns a copy of the layer with the given ID, or nil if not found.
func (d *Document) LayerByID(id int) *Layer {
        l := d.layers[id]
        if l == nil {
                return nil
        }
        cp := *l
        return &cp
}

// CurrentLayer returns the current (active) layer ID.
func (d *Document) CurrentLayer() int { return d.curLayer }

// SetCurrentLayer sets the active layer. Returns false if the layer doesn't exist.
func (d *Document) SetCurrentLayer(id int) bool {
        if _, ok := d.layers[id]; !ok {
                return false
        }
        d.curLayer = id
        return true
}

// AddLayer creates a new layer and returns its ID.
// If a layer with the same name already exists, the existing ID is returned.
func (d *Document) AddLayer(name string, color string, lt LineType, lw float64) int {
        for _, l := range d.layers {
                if l.Name == name {
                        return l.ID
                }
        }
        if color == "" {
                color = "#ffffff"
        }
        if lt == "" {
                lt = LineTypeSolid
        }
        id := d.nextLayerID
        d.nextLayerID++
        d.layers[id] = &Layer{
                ID: id, Name: name, Color: color, LineTyp: lt, LineWeight: lw,
                Visible: true, PrintEnabled: true,
        }
        return id
}

// RemoveLayer deletes a layer by ID. Layer 0 cannot be removed.
// Returns false if the layer is not found or protected.
func (d *Document) RemoveLayer(id int) bool {
        if id == 0 {
                return false
        }
        if _, ok := d.layers[id]; !ok {
                return false
        }
        delete(d.layers, id)
        if d.curLayer == id {
                d.curLayer = 0
        }
        return true
}

// RenameLayer renames a layer. Returns false if not found or the name is taken.
func (d *Document) RenameLayer(id int, name string) bool {
        l, ok := d.layers[id]
        if !ok {
                return false
        }
        for _, other := range d.layers {
                if other.ID != id && other.Name == name {
                        return false
                }
        }
        l.Name = name
        return true
}

// SetLayerColor sets the color of a layer (RGB hex string). Returns false if not found.
func (d *Document) SetLayerColor(id int, color string) bool {
        l, ok := d.layers[id]
        if !ok {
                return false
        }
        l.Color = color
        return true
}

// SetLayerLineType sets the line type of a layer.
func (d *Document) SetLayerLineType(id int, lt LineType) bool {
        l, ok := d.layers[id]
        if !ok {
                return false
        }
        l.LineTyp = lt
        return true
}

// SetLayerLineWeight sets the line weight of a layer in mm.
func (d *Document) SetLayerLineWeight(id int, lw float64) bool {
        l, ok := d.layers[id]
        if !ok {
                return false
        }
        l.LineWeight = lw
        return true
}

// SetLayerVisible sets whether a layer is visible.
func (d *Document) SetLayerVisible(id int, visible bool) bool {
        l, ok := d.layers[id]
        if !ok {
                return false
        }
        l.Visible = visible
        return true
}

// SetLayerLocked sets whether a layer is locked (entities on locked layers cannot be edited).
func (d *Document) SetLayerLocked(id int, locked bool) bool {
        l, ok := d.layers[id]
        if !ok {
                return false
        }
        l.Locked = locked
        return true
}

// SetLayerFrozen sets whether a layer is frozen (entities not rendered when frozen).
func (d *Document) SetLayerFrozen(id int, frozen bool) bool {
        l, ok := d.layers[id]
        if !ok {
                return false
        }
        l.Frozen = frozen
        return true
}

// SetLayerPrint sets whether a layer is included in print output.
func (d *Document) SetLayerPrint(id int, print bool) bool {
        l, ok := d.layers[id]
        if !ok {
                return false
        }
        l.PrintEnabled = print
        return true
}

// IsLayerVisible returns true if the layer is visible and not frozen.
// Unknown layer IDs return true (show by default).
func (d *Document) IsLayerVisible(id int) bool {
        l, ok := d.layers[id]
        if !ok {
                return true
        }
        return l.Visible && !l.Frozen
}

// IsLayerLocked returns true if the layer is locked.
func (d *Document) IsLayerLocked(id int) bool {
        l, ok := d.layers[id]
        if !ok {
                return false
        }
        return l.Locked
}

// EffectiveColor returns the display color for an entity.
// If entity Color is "BYLAYER", the layer color is used instead.
func (d *Document) EffectiveColor(e Entity) string {
        if !strings.EqualFold(e.Color, "BYLAYER") {
                return e.Color
        }
        if l, ok := d.layers[e.Layer]; ok {
                return l.Color
        }
        return "#ffffff"
}

// layerName returns the human-readable name for layer id, falling back to
// the integer representation for backward-compatibility with old saves.
func (d *Document) layerName(id int) string {
        if l, ok := d.layers[id]; ok {
                return l.Name
        }
        return fmt.Sprintf("%d", id)
}

// ─── DXF layer table helpers ─────────────────────────────────────────────────

// WriteDXFLayerTable emits the TABLES/LAYER DXF section into sb.
// Call this before the ENTITIES section in exportDXF.
func (d *Document) writeDXFLayerTable(sb *strings.Builder, r12 bool) {
        layers := d.Layers()
        sb.WriteString("  0\nSECTION\n  2\nTABLES\n")

        // Collect all line types actually used by layers so we only emit what is needed.
        usedLT := map[string]bool{"CONTINUOUS": true}
        for _, l := range layers {
                usedLT[dxfLTypeName(l.LineTyp)] = true
        }

        if r12 {
                // R12: LTYPE table entries use a simpler format.
                sb.WriteString("  0\nTABLE\n  2\nLTYPE\n 70\n")
                fmt.Fprintf(sb, "%d\n", len(usedLT))
                for name := range usedLT {
                        meta := dxfLTypeR12Meta(name)
                        fmt.Fprintf(sb, "  0\nLTYPE\n  2\n%s\n 70\n0\n  3\n%s\n 72\n65\n 73\n%d\n 40\n%f\n",
                                name, meta.description, meta.numDashes, meta.patternLen)
                        for _, dash := range meta.dashes {
                                fmt.Fprintf(sb, " 49\n%f\n", dash)
                        }
                }
                sb.WriteString("  0\nENDTAB\n")
        } else {
                // R2000+: LTYPE table with AcDb sub-records.
                sb.WriteString("  0\nTABLE\n  2\nLTYPE\n  5\n4\n100\nAcDbSymbolTable\n")
                fmt.Fprintf(sb, " 70\n%d\n", len(usedLT))
                handle := 14
                for name := range usedLT {
                        meta := dxfLTypeR12Meta(name)
                        fmt.Fprintf(sb, "  0\nLTYPE\n  5\n%x\n100\nAcDbSymbolTableRecord\n100\nAcDbLinetypeTableRecord\n",
                                handle)
                        handle++
                        fmt.Fprintf(sb, "  2\n%s\n 70\n0\n  3\n%s\n 72\n65\n 73\n%d\n 40\n%f\n",
                                name, meta.description, meta.numDashes, meta.patternLen)
                        for _, dash := range meta.dashes {
                                fmt.Fprintf(sb, " 49\n%f\n", dash)
                        }
                }
                sb.WriteString("  0\nENDTAB\n")
        }

        fmt.Fprintf(sb, "  0\nTABLE\n  2\nLAYER\n")
        if !r12 {
                fmt.Fprintf(sb, "  5\n2\n100\nAcDbSymbolTable\n")
        }
        fmt.Fprintf(sb, " 70\n%d\n", len(layers))

        for _, l := range layers {
                flags := dxfLayerFlags(l)
                aci := rgbToACI(l.Color)
                ltName := dxfLTypeName(l.LineTyp)
                if r12 {
                        fmt.Fprintf(sb, "  0\nLAYER\n  2\n%s\n 70\n%d\n 62\n%d\n  6\n%s\n",
                                l.Name, flags, aci, ltName)
                } else {
                        lw := dxfLineWeightCode(l.LineWeight)
                        fmt.Fprintf(sb,
                                "  0\nLAYER\n100\nAcDbSymbolTableRecord\n100\nAcDbLayerTableRecord\n  2\n%s\n 70\n%d\n 62\n%d\n  6\n%s\n370\n%d\n",
                                l.Name, flags, aci, ltName, lw)
                }
        }
        sb.WriteString("  0\nENDTAB\n  0\nENDSEC\n")
}

// dxfLayerFlags returns the DXF layer flags integer.
// Bit 0 = frozen, bit 2 = locked.
func dxfLayerFlags(l *Layer) int {
        f := 0
        if l.Frozen || !l.Visible {
                f |= 1
        }
        if l.Locked {
                f |= 4
        }
        return f
}

// rgbToACI maps a hex RGB color to the nearest AutoCAD Color Index (ACI).
// Falls back to 7 (white/black) for unknown colors.
func rgbToACI(color string) int {
        switch strings.ToUpper(strings.TrimSpace(color)) {
        case "#FF0000":
                return 1
        case "#FFFF00":
                return 2
        case "#00FF00":
                return 3
        case "#00FFFF":
                return 4
        case "#0000FF":
                return 5
        case "#FF00FF":
                return 6
        default:
                return 7
        }
}

// ltypeMeta holds the DXF descriptor data for a single line type pattern.
type ltypeMeta struct {
        description string
        numDashes   int
        patternLen  float64
        dashes      []float64
}

// dxfLTypeR12Meta returns the pattern metadata for a DXF LTYPE entry.
// These patterns match AutoCAD standard line-type definitions.
func dxfLTypeR12Meta(name string) ltypeMeta {
        switch name {
        case "DASHED":
                return ltypeMeta{"Dashed _ _ _ _ _ _ _", 2, 12.7, []float64{6.35, -6.35}}
        case "DOTTED":
                return ltypeMeta{"Dotted . . . . . . .", 2, 3.175, []float64{0.0, -3.175}}
        case "DASHDOT":
                return ltypeMeta{"Dash dot -. -. -. -.", 4, 14.0, []float64{6.35, -3.175, 0.0, -3.175}}
        case "CENTER":
                return ltypeMeta{"Center ____ _ ____ _", 4, 31.75, []float64{19.05, -3.175, 3.175, -3.175}}
        case "HIDDEN":
                return ltypeMeta{"Hidden __ __ __ __ _", 2, 9.525, []float64{6.35, -3.175}}
        default: // CONTINUOUS
                return ltypeMeta{"Solid line", 0, 0.0, nil}
        }
}

// dxfLTypeName maps a LineType to its DXF LTYPE name.
func dxfLTypeName(lt LineType) string {
        switch lt {
        case LineTypeDashed:
                return "DASHED"
        case LineTypeDotted:
                return "DOTTED"
        case LineTypeDashDot:
                return "DASHDOT"
        case LineTypeCenter:
                return "CENTER"
        case LineTypeHidden:
                return "HIDDEN"
        default:
                return "CONTINUOUS"
        }
}

// dxfLineWeightCode converts a line weight in mm to the nearest standard
// DXF lineweight integer (units: hundredths of mm).
func dxfLineWeightCode(lw float64) int {
        standard := []int{0, 5, 9, 13, 15, 18, 20, 25, 30, 35, 40, 50, 53, 60, 70, 80, 90, 100, 106, 120, 140, 158, 200, 211}
        v := int(math.Round(lw * 100))
        best := standard[0]
        for _, s := range standard {
                if intAbs(s-v) < intAbs(best-v) {
                        best = s
                }
        }
        return best
}

func intAbs(x int) int {
        if x < 0 {
                return -x
        }
        return x
}
