// cmd/cad is a terminal-mode desktop interface for go-cad.
//
// It exposes the full document model (entities, layers) and object-snap engine
// through an interactive Read-Eval-Print loop.  Users can draw entities, manage
// layers, enable/disable individual snap modes, and export DXF — all without a
// browser.
//
// Usage:
//
//      go run ./cmd/cad
//
// Available commands (case-insensitive):
//
//      LINE x1 y1 x2 y2             – add a line
//      CIRCLE cx cy r               – add a circle
//      ARC cx cy r startDeg endDeg  – add an arc
//      RECT x1 y1 x2 y2             – add a rectangle
//      LIST                         – list all entities
//      UNDO / REDO                  – undo/redo last change
//      LAYERS                       – list all layers
//      ADDLAYER name color          – add a layer (color: #rrggbb)
//      RMLAYER id                   – remove layer by ID
//      SETLAYER id                  – set current drawing layer
//      LAYERCOLOR id color          – set layer color
//      LAYERLT id lineType          – set layer line type (Solid/Dashed/Dotted/DashDot/Center/Hidden)
//      LAYERLW id weight            – set layer line weight in mm
//      LAYERVIS id true|false       – toggle layer visibility
//      LAYERLCK id true|false       – toggle layer lock
//      LAYERFRZ id true|false       – toggle layer frozen state
//      LAYERPRT id true|false       – toggle layer print flag
//      SNAP                         – show snap engine status and active modes
//      SNAPON / SNAPOFF             – enable / disable object snap
//      SNAPMODE mode true|false     – enable/disable individual snap mode
//                                     modes: endpoint midpoint center quadrant
//                                            intersection perpendicular tangent nearest
//      FINDSNAP x y radius          – run snap query at world point (x,y) within radius
//      SAVE file.json               – save document to JSON
//      LOAD file.json               – load document from JSON
//      EXPORT file.dxf [r12]        – export DXF (default R2000; pass r12 for AutoCAD R12)
//      HELP                         – show this help
//      QUIT / EXIT                  – exit
package main

import (
        "bufio"
        "fmt"
        "os"
        "strconv"
        "strings"

        "go-cad/internal/document"
        "go-cad/internal/snap"
)

// ─── Snap settings ────────────────────────────────────────────────────────────

type snapConfig struct {
        enabled bool
        mask    snap.SnapType
}

func defaultSnapConfig() snapConfig {
        return snapConfig{enabled: true, mask: snap.SnapAll}
}

var snapModeNames = map[string]snap.SnapType{
        "endpoint":      snap.SnapEndpoint,
        "midpoint":      snap.SnapMidpoint,
        "center":        snap.SnapCenter,
        "quadrant":      snap.SnapQuadrant,
        "intersection":  snap.SnapIntersection,
        "perpendicular": snap.SnapPerpendicular,
        "tangent":       snap.SnapTangent,
        "nearest":       snap.SnapNearest,
}

// ─── Main REPL ────────────────────────────────────────────────────────────────

func main() {
        doc := document.New()
        cfg := defaultSnapConfig()

        sc := bufio.NewScanner(os.Stdin)
        fmt.Println("go-cad desktop terminal  (type HELP for commands, QUIT to exit)")
        fmt.Printf("Layer: %d  Snap: ON  Modes: ALL\n\n", doc.CurrentLayer())

        for {
                fmt.Print("> ")
                if !sc.Scan() {
                        break
                }
                line := strings.TrimSpace(sc.Text())
                if line == "" {
                        continue
                }
                parts := strings.Fields(line)
                cmd := strings.ToUpper(parts[0])
                args := parts[1:]

                switch cmd {
                case "HELP":
                        printHelp()

                case "QUIT", "EXIT":
                        fmt.Println("bye")
                        return

                // ── Entity creation ──────────────────────────────────────────────────
                case "LINE":
                        if len(args) < 4 {
                                fmt.Println("usage: LINE x1 y1 x2 y2")
                                continue
                        }
                        coords, err := parseFloats(args[:4])
                        if err != nil {
                                fmt.Println("error:", err)
                                continue
                        }
                        id := doc.AddLine(coords[0], coords[1], coords[2], coords[3], doc.CurrentLayer(), "BYLAYER")
                        fmt.Printf("added line id=%d layer=%d\n", id, doc.CurrentLayer())

                case "CIRCLE":
                        if len(args) < 3 {
                                fmt.Println("usage: CIRCLE cx cy r")
                                continue
                        }
                        coords, err := parseFloats(args[:3])
                        if err != nil {
                                fmt.Println("error:", err)
                                continue
                        }
                        id := doc.AddCircle(coords[0], coords[1], coords[2], doc.CurrentLayer(), "BYLAYER")
                        fmt.Printf("added circle id=%d layer=%d\n", id, doc.CurrentLayer())

                case "ARC":
                        if len(args) < 5 {
                                fmt.Println("usage: ARC cx cy r startDeg endDeg")
                                continue
                        }
                        coords, err := parseFloats(args[:5])
                        if err != nil {
                                fmt.Println("error:", err)
                                continue
                        }
                        id := doc.AddArc(coords[0], coords[1], coords[2], coords[3], coords[4], doc.CurrentLayer(), "BYLAYER")
                        fmt.Printf("added arc id=%d layer=%d\n", id, doc.CurrentLayer())

                case "RECT":
                        if len(args) < 4 {
                                fmt.Println("usage: RECT x1 y1 x2 y2")
                                continue
                        }
                        coords, err := parseFloats(args[:4])
                        if err != nil {
                                fmt.Println("error:", err)
                                continue
                        }
                        id := doc.AddRectangle(coords[0], coords[1], coords[2], coords[3], doc.CurrentLayer(), "BYLAYER")
                        fmt.Printf("added rect id=%d layer=%d\n", id, doc.CurrentLayer())

                // ── List entities ────────────────────────────────────────────────────
                case "LIST":
                        entities := doc.Entities()
                        if len(entities) == 0 {
                                fmt.Println("(no entities)")
                                continue
                        }
                        for _, e := range entities {
                                layerName := doc.LayerByID(e.Layer)
                                ln := "?"
                                if layerName != nil {
                                        ln = layerName.Name
                                }
                                fmt.Printf("  id=%-4d type=%-12s layer=%d(%s)\n", e.ID, e.Type, e.Layer, ln)
                        }
                        fmt.Printf("%d entities total\n", len(entities))

                // ── Undo / Redo ──────────────────────────────────────────────────────
                case "UNDO":
                        doc.Undo()
                        fmt.Println("undo applied")
                case "REDO":
                        doc.Redo()
                        fmt.Println("redo applied")

                // ── Layer management ─────────────────────────────────────────────────
                case "LAYERS":
                        printLayers(doc)

                case "ADDLAYER":
                        if len(args) < 2 {
                                fmt.Println("usage: ADDLAYER name color")
                                continue
                        }
                        color := args[1]
                        if !strings.HasPrefix(color, "#") {
                                color = "#" + color
                        }
                        id := doc.AddLayer(args[0], color, document.LineTypeSolid, 0.25)
                        fmt.Printf("added layer id=%d name=%q color=%s\n", id, args[0], color)

                case "RMLAYER":
                        if len(args) < 1 {
                                fmt.Println("usage: RMLAYER id")
                                continue
                        }
                        id, err := strconv.Atoi(args[0])
                        if err != nil {
                                fmt.Println("invalid id:", args[0])
                                continue
                        }
                        if doc.RemoveLayer(id) {
                                fmt.Printf("removed layer %d\n", id)
                        } else {
                                fmt.Printf("cannot remove layer %d (not found or protected)\n", id)
                        }

                case "SETLAYER":
                        if len(args) < 1 {
                                fmt.Println("usage: SETLAYER id")
                                continue
                        }
                        id, err := strconv.Atoi(args[0])
                        if err != nil {
                                fmt.Println("invalid id:", args[0])
                                continue
                        }
                        if doc.SetCurrentLayer(id) {
                                fmt.Printf("current layer → %d\n", id)
                        } else {
                                fmt.Printf("layer %d not found\n", id)
                        }

                case "LAYERCOLOR":
                        if len(args) < 2 {
                                fmt.Println("usage: LAYERCOLOR id color")
                                continue
                        }
                        id, err := strconv.Atoi(args[0])
                        if err != nil {
                                fmt.Println("invalid id:", args[0])
                                continue
                        }
                        color := args[1]
                        if !strings.HasPrefix(color, "#") {
                                color = "#" + color
                        }
                        if doc.SetLayerColor(id, color) {
                                fmt.Printf("layer %d color → %s\n", id, color)
                        } else {
                                fmt.Printf("layer %d not found\n", id)
                        }

                case "LAYERLT":
                        if len(args) < 2 {
                                fmt.Println("usage: LAYERLT id lineType")
                                continue
                        }
                        id, err := strconv.Atoi(args[0])
                        if err != nil {
                                fmt.Println("invalid id:", args[0])
                                continue
                        }
                        lt := document.LineType(args[1])
                        if doc.SetLayerLineType(id, lt) {
                                fmt.Printf("layer %d lineType → %s\n", id, lt)
                        } else {
                                fmt.Printf("layer %d not found\n", id)
                        }

                case "LAYERLW":
                        if len(args) < 2 {
                                fmt.Println("usage: LAYERLW id weight")
                                continue
                        }
                        id, err := strconv.Atoi(args[0])
                        if err != nil {
                                fmt.Println("invalid id:", args[0])
                                continue
                        }
                        lw, err2 := strconv.ParseFloat(args[1], 64)
                        if err2 != nil {
                                fmt.Println("invalid weight:", args[1])
                                continue
                        }
                        if doc.SetLayerLineWeight(id, lw) {
                                fmt.Printf("layer %d lineWeight → %.2fmm\n", id, lw)
                        } else {
                                fmt.Printf("layer %d not found\n", id)
                        }

                case "LAYERVIS":
                        if len(args) < 2 {
                                fmt.Println("usage: LAYERVIS id true|false")
                                continue
                        }
                        id, err := strconv.Atoi(args[0])
                        if err != nil {
                                fmt.Println("invalid id:", args[0])
                                continue
                        }
                        v := strings.ToLower(args[1]) != "false"
                        if doc.SetLayerVisible(id, v) {
                                fmt.Printf("layer %d visible → %v\n", id, v)
                        } else {
                                fmt.Printf("layer %d not found\n", id)
                        }

                case "LAYERLCK":
                        if len(args) < 2 {
                                fmt.Println("usage: LAYERLCK id true|false")
                                continue
                        }
                        id, err := strconv.Atoi(args[0])
                        if err != nil {
                                fmt.Println("invalid id:", args[0])
                                continue
                        }
                        v := strings.ToLower(args[1]) != "false"
                        if doc.SetLayerLocked(id, v) {
                                fmt.Printf("layer %d locked → %v\n", id, v)
                        } else {
                                fmt.Printf("layer %d not found\n", id)
                        }

                case "LAYERFRZ":
                        if len(args) < 2 {
                                fmt.Println("usage: LAYERFRZ id true|false")
                                continue
                        }
                        id, err := strconv.Atoi(args[0])
                        if err != nil {
                                fmt.Println("invalid id:", args[0])
                                continue
                        }
                        v := strings.ToLower(args[1]) != "false"
                        if doc.SetLayerFrozen(id, v) {
                                fmt.Printf("layer %d frozen → %v\n", id, v)
                        } else {
                                fmt.Printf("layer %d not found\n", id)
                        }

                case "LAYERPRT":
                        if len(args) < 2 {
                                fmt.Println("usage: LAYERPRT id true|false")
                                continue
                        }
                        id, err := strconv.Atoi(args[0])
                        if err != nil {
                                fmt.Println("invalid id:", args[0])
                                continue
                        }
                        v := strings.ToLower(args[1]) != "false"
                        if doc.SetLayerPrint(id, v) {
                                fmt.Printf("layer %d print → %v\n", id, v)
                        } else {
                                fmt.Printf("layer %d not found\n", id)
                        }

                // ── Snap management ──────────────────────────────────────────────────
                case "SNAP":
                        printSnapStatus(cfg)

                case "SNAPON":
                        cfg.enabled = true
                        fmt.Println("object snap: ON")

                case "SNAPOFF":
                        cfg.enabled = false
                        fmt.Println("object snap: OFF")

                case "SNAPMODE":
                        if len(args) < 2 {
                                fmt.Println("usage: SNAPMODE modeName true|false")
                                fmt.Println("modes:", strings.Join(snapModeNamesList(), ", "))
                                continue
                        }
                        bit, ok := snapModeNames[strings.ToLower(args[0])]
                        if !ok {
                                fmt.Println("unknown snap mode:", args[0])
                                fmt.Println("modes:", strings.Join(snapModeNamesList(), ", "))
                                continue
                        }
                        on := strings.ToLower(args[1]) != "false"
                        if on {
                                cfg.mask |= bit
                        } else {
                                cfg.mask &^= bit
                        }
                        fmt.Printf("snap mode %q → %v  (mask=0x%02x)\n", args[0], on, cfg.mask)

                case "FINDSNAP":
                        if len(args) < 3 {
                                fmt.Println("usage: FINDSNAP x y radius")
                                continue
                        }
                        coords, err := parseFloats(args[:3])
                        if err != nil {
                                fmt.Println("error:", err)
                                continue
                        }
                        if !cfg.enabled {
                                fmt.Println("snap is disabled (SNAPON to enable)")
                                continue
                        }
                        mask := cfg.mask
                        // mask==0 means all modes disabled — respect it exactly.
                        if mask == 0 {
                                fmt.Println("all snap modes are disabled (use SNAPMODE … true or set mask > 0)")
                                continue
                        }
                        result := snap.FindSnap(coords[0], coords[1], doc.Entities(), coords[2], mask)
                        if result == nil {
                                fmt.Printf("no snap found within radius %.2f of (%.2f, %.2f)\n",
                                        coords[2], coords[0], coords[1])
                        } else {
                                typeName := snap.SnapNames[result.Type]
                                fmt.Printf("snap: %s at (%.4f, %.4f)  entityID=%d\n",
                                        typeName, result.X, result.Y, result.EntityID)
                        }

                // ── File I/O ─────────────────────────────────────────────────────────
                case "SAVE":
                        if len(args) < 1 {
                                fmt.Println("usage: SAVE file.json")
                                continue
                        }
                        if err := doc.Save(args[0]); err != nil {
                                fmt.Println("save error:", err)
                                continue
                        }
                        fmt.Printf("saved to %s\n", args[0])

                case "LOAD":
                        if len(args) < 1 {
                                fmt.Println("usage: LOAD file.json")
                                continue
                        }
                        if err := doc.Load(args[0]); err != nil {
                                fmt.Println("load error:", err)
                                continue
                        }
                        fmt.Printf("loaded %d entities, %d layers from %s\n",
                                len(doc.Entities()), len(doc.Layers()), args[0])

                case "EXPORT":
                        if len(args) < 1 {
                                fmt.Println("usage: EXPORT file.dxf [r12]")
                                continue
                        }
                        r12 := len(args) >= 2 && strings.ToLower(args[1]) == "r12"
                        var dxf string
                        if r12 {
                                dxf = doc.ExportDXFR12()
                        } else {
                                dxf = doc.ExportDXF()
                        }
                        if err := os.WriteFile(args[0], []byte(dxf), 0644); err != nil {
                                fmt.Println("write error:", err)
                                continue
                        }
                        ver := "R2000"
                        if r12 {
                                ver = "R12"
                        }
                        fmt.Printf("exported %d bytes (%s) to %s\n", len(dxf), ver, args[0])

                default:
                        fmt.Printf("unknown command %q — type HELP\n", cmd)
                }
        }
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func parseFloats(ss []string) ([]float64, error) {
        out := make([]float64, len(ss))
        for i, s := range ss {
                v, err := strconv.ParseFloat(s, 64)
                if err != nil {
                        return nil, fmt.Errorf("bad number %q: %w", s, err)
                }
                out[i] = v
        }
        return out, nil
}

func snapModeNamesList() []string {
        names := make([]string, 0, len(snapModeNames))
        for n := range snapModeNames {
                names = append(names, n)
        }
        return names
}

func printLayers(doc *document.Document) {
        layers := doc.Layers()
        cur := doc.CurrentLayer()
        fmt.Printf("  %-4s %-3s %-18s %-9s %-8s %-6s %-6s %-6s %-5s %-5s\n",
                "ID", "Cur", "Name", "Color", "LineType", "LW mm", "Vis", "Locked", "Frozen", "Print")
        fmt.Println("  " + strings.Repeat("-", 85))
        for _, l := range layers {
                curMark := " "
                if l.ID == cur {
                        curMark = "●"
                }
                fmt.Printf("  %-4d %-3s %-18s %-9s %-8s %-6.2f %-6v %-6v %-6v %-5v\n",
                        l.ID, curMark, l.Name, l.Color, l.LineTyp, l.LineWeight,
                        l.Visible, l.Locked, l.Frozen, l.PrintEnabled)
        }
        fmt.Printf("%d layers\n", len(layers))
}

func printSnapStatus(cfg snapConfig) {
        onOff := "ON"
        if !cfg.enabled {
                onOff = "OFF"
        }
        fmt.Printf("snap: %s  mask=0x%02x\n", onOff, cfg.mask)
        for name, bit := range snapModeNames {
                state := "✗"
                if cfg.mask&bit != 0 {
                        state = "✓"
                }
                fmt.Printf("  %s %-14s\n", state, name)
        }
}

func printHelp() {
        fmt.Print(`
Entity creation:
  LINE x1 y1 x2 y2               add a line
  CIRCLE cx cy r                  add a circle
  ARC cx cy r startDeg endDeg     add an arc
  RECT x1 y1 x2 y2               add a rectangle
  LIST                            list all entities
  UNDO / REDO                     undo / redo

Layer management:
  LAYERS                          list layers
  ADDLAYER name color             add layer (color: #rrggbb)
  RMLAYER id                      remove layer
  SETLAYER id                     set current drawing layer
  LAYERCOLOR id #rrggbb           set layer color
  LAYERLT id lineType             set line type (Solid/Dashed/Dotted/DashDot/Center/Hidden)
  LAYERLW id weight               set line weight (mm)
  LAYERVIS id true|false          visibility
  LAYERLCK id true|false          locked
  LAYERFRZ id true|false          frozen
  LAYERPRT id true|false          print enabled

Object snap:
  SNAP                            show snap status and active modes
  SNAPON / SNAPOFF                enable / disable snap globally
  SNAPMODE modeName true|false    toggle individual snap mode
  FINDSNAP x y radius             run snap query and print result

File I/O:
  SAVE file.json                  save document
  LOAD file.json                  load document
  EXPORT file.dxf [r12]          export DXF

  HELP   QUIT
`)
}
