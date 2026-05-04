package dxf_test

import (
        "fmt"
        "math"
        "strings"
        "testing"

        "go-cad/internal/document"
        "go-cad/pkg/dxf"
)

// ─── Round-trip tests ─────────────────────────────────────────────────────────

func TestRoundTripLine(t *testing.T) {
        d := document.New()
        d.AddLine(10, 20, 30, 40, 0, "#ff0000")
        dxfStr := dxf.String(d)
        d2, warns, err := dxf.ReadString(dxfStr)
        if err != nil {
                t.Fatalf("Read error: %v", err)
        }
        logWarns(t, warns)
        ents := d2.Entities()
        if len(ents) != 1 {
                t.Fatalf("got %d entities, want 1", len(ents))
        }
        e := ents[0]
        if e.Type != document.TypeLine {
                t.Errorf("type = %q, want %q", e.Type, document.TypeLine)
        }
        if !approxEq(e.X1, 10) || !approxEq(e.Y1, 20) {
                t.Errorf("start = (%.4f, %.4f), want (10, 20)", e.X1, e.Y1)
        }
        if !approxEq(e.X2, 30) || !approxEq(e.Y2, 40) {
                t.Errorf("end = (%.4f, %.4f), want (30, 40)", e.X2, e.Y2)
        }
}

func TestRoundTripCircle(t *testing.T) {
        d := document.New()
        d.AddCircle(5, -3, 7, 0, "#00ff00")
        d2, _, err := dxf.ReadString(dxf.String(d))
        if err != nil {
                t.Fatal(err)
        }
        ents := d2.Entities()
        if len(ents) != 1 {
                t.Fatalf("got %d entities, want 1", len(ents))
        }
        e := ents[0]
        if e.Type != document.TypeCircle {
                t.Errorf("type = %q, want %q", e.Type, document.TypeCircle)
        }
        if !approxEq(e.CX, 5) || !approxEq(e.CY, -3) || !approxEq(e.R, 7) {
                t.Errorf("circle = cx=%.3f cy=%.3f r=%.3f, want cx=5 cy=-3 r=7", e.CX, e.CY, e.R)
        }
}

func TestRoundTripArc(t *testing.T) {
        d := document.New()
        d.AddArc(0, 0, 10, 30, 150, 0, "#ffffff")
        d2, _, err := dxf.ReadString(dxf.String(d))
        if err != nil {
                t.Fatal(err)
        }
        ents := d2.Entities()
        if len(ents) != 1 {
                t.Fatalf("got %d entities, want 1", len(ents))
        }
        e := ents[0]
        if e.Type != document.TypeArc {
                t.Errorf("type = %q, want arc", e.Type)
        }
        if !approxEq(e.R, 10) {
                t.Errorf("R = %.3f, want 10", e.R)
        }
}

func TestRoundTripPolyline(t *testing.T) {
        d := document.New()
        pts := [][]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
        d.AddPolyline(pts, 0, "#ffffff")
        d2, _, err := dxf.ReadString(dxf.String(d))
        if err != nil {
                t.Fatal(err)
        }
        if d2.EntityCount() == 0 {
                t.Error("no entities after polyline round-trip")
        }
}

func TestRoundTripText(t *testing.T) {
        d := document.New()
        d.AddText(5, 10, "Hello", 3, 0, "Standard", 0, "#ffffff")
        d2, _, err := dxf.ReadString(dxf.String(d))
        if err != nil {
                t.Fatal(err)
        }
        ents := d2.Entities()
        if len(ents) != 1 {
                t.Fatalf("got %d entities, want 1", len(ents))
        }
        e := ents[0]
        if e.Type != document.TypeText {
                t.Errorf("type = %q, want text", e.Type)
        }
        if e.Text != "Hello" {
                t.Errorf("text = %q, want %q", e.Text, "Hello")
        }
}

func TestRoundTripMultiEntity(t *testing.T) {
        d := document.New()
        d.AddLine(0, 0, 10, 10, 0, "#ff0000")
        d.AddCircle(5, 5, 3, 0, "#00ff00")
        d.AddArc(0, 0, 5, 0, 90, 0, "#0000ff")
        d.AddText(1, 1, "test", 2.5, 0, "", 0, "#ffffff")
        d2, _, err := dxf.ReadString(dxf.String(d))
        if err != nil {
                t.Fatal(err)
        }
        if d2.EntityCount() != d.EntityCount() {
                t.Errorf("entity count: got %d, want %d", d2.EntityCount(), d.EntityCount())
        }
}

func TestRoundTripR12(t *testing.T) {
        d := document.New()
        d.AddLine(0, 0, 100, 100, 0, "#ffffff")
        d.AddCircle(50, 50, 25, 0, "#ff0000")
        dxfStr := dxf.StringR12(d)
        if !strings.Contains(dxfStr, "AC1009") {
                t.Error("expected AC1009 in R12 DXF")
        }
        d2, _, err := dxf.ReadString(dxfStr)
        if err != nil {
                t.Fatal(err)
        }
        if d2.EntityCount() == 0 {
                t.Error("no entities after R12 round-trip")
        }
}

func TestLayerPreservation(t *testing.T) {
        d := document.New()
        id := d.AddLayer("Construction", "#ff8800", document.LineTypeDashed, 0.35)
        d.AddLine(0, 0, 10, 0, id, "#ff8800")
        d2, _, err := dxf.ReadString(dxf.String(d))
        if err != nil {
                t.Fatal(err)
        }
        layers := d2.Layers()
        found := false
        for _, l := range layers {
                if l.Name == "Construction" {
                        found = true
                        break
                }
        }
        if !found {
                t.Errorf("layer 'Construction' not found after round-trip; layers=%v", layerNames(layers))
        }
}

// ─── Smoke tests for known DXF fragments ─────────────────────────────────────

func TestReadLineFrag(t *testing.T) {
        frag := "  0\nSECTION\n  2\nENTITIES\n" +
                "  0\nLINE\n  8\n0\n 10\n1.0\n 20\n2.0\n 11\n3.0\n 21\n4.0\n" +
                "  0\nENDSEC\n  0\nEOF\n"
        d, _, err := dxf.ReadString(frag)
        if err != nil {
                t.Fatal(err)
        }
        ents := d.Entities()
        if len(ents) != 1 {
                t.Fatalf("got %d entities, want 1", len(ents))
        }
        e := ents[0]
        // DXF Y=2.0 → imported as Y=-2.0 (Y-flip for Cartesian→screen)
        if !approxEq(e.X1, 1.0) || !approxEq(e.Y1, -2.0) {
                t.Errorf("start = (%.2f, %.2f), want (1, -2)", e.X1, e.Y1)
        }
}

func TestReadCircleFrag(t *testing.T) {
        frag := "  0\nSECTION\n  2\nENTITIES\n" +
                "  0\nCIRCLE\n  8\n0\n 10\n5.0\n 20\n5.0\n 40\n3.0\n" +
                "  0\nENDSEC\n  0\nEOF\n"
        d, _, err := dxf.ReadString(frag)
        if err != nil {
                t.Fatal(err)
        }
        ents := d.Entities()
        if len(ents) != 1 {
                t.Fatalf("got %d entities, want 1", len(ents))
        }
        if !approxEq(ents[0].R, 3.0) {
                t.Errorf("R = %.2f, want 3", ents[0].R)
        }
}

func TestReadLWPolylineFrag(t *testing.T) {
        frag := "  0\nSECTION\n  2\nENTITIES\n" +
                "  0\nLWPOLYLINE\n  8\n0\n 90\n3\n 70\n0\n" +
                " 10\n0.0\n 20\n0.0\n 10\n10.0\n 20\n0.0\n 10\n10.0\n 20\n10.0\n" +
                "  0\nENDSEC\n  0\nEOF\n"
        d, _, err := dxf.ReadString(frag)
        if err != nil {
                t.Fatal(err)
        }
        ents := d.Entities()
        if len(ents) != 1 {
                t.Fatalf("got %d entities, want 1", len(ents))
        }
        if len(ents[0].Points) != 3 {
                t.Errorf("got %d points, want 3", len(ents[0].Points))
        }
}

func TestReadLayerTableFrag(t *testing.T) {
        frag := "  0\nSECTION\n  2\nTABLES\n" +
                "  0\nTABLE\n  2\nLAYER\n 70\n1\n" +
                "  0\nLAYER\n  2\nDims\n 62\n3\n  6\nCONTINUOUS\n 70\n0\n" +
                "  0\nENDTAB\n  0\nENDSEC\n" +
                "  0\nSECTION\n  2\nENTITIES\n  0\nENDSEC\n  0\nEOF\n"
        d, _, err := dxf.ReadString(frag)
        if err != nil {
                t.Fatal(err)
        }
        layers := d.Layers()
        found := false
        for _, l := range layers {
                if l.Name == "Dims" {
                        found = true
                        if !strings.EqualFold(l.Color, "#00FF00") {
                                t.Errorf("layer color = %q, want #00FF00", l.Color)
                        }
                }
        }
        if !found {
                t.Error("layer 'Dims' not found")
        }
}

func TestReadMalformedEntity(t *testing.T) {
        // Should not panic; skip malformed entity gracefully and keep parsing rest.
        frag := "  0\nSECTION\n  2\nENTITIES\n" +
                "  0\nLINE\n  8\n0\n 10\nnot-a-number\n 20\n0.0\n 11\n5.0\n 21\n5.0\n" +
                "  0\nCIRCLE\n  8\n0\n 10\n5.0\n 20\n5.0\n 40\n2.0\n" +
                "  0\nENDSEC\n  0\nEOF\n"
        d, _, err := dxf.ReadString(frag)
        if err != nil {
                t.Fatal(err)
        }
        if d.EntityCount() == 0 {
                t.Error("expected at least the circle to be parsed")
        }
}

func TestACIColors(t *testing.T) {
        tests := []struct {
                aci  int
                want string
        }{
                {1, "#FF0000"},
                {3, "#00FF00"},
                {5, "#0000FF"},
                {7, "#FFFFFF"},
        }
        for _, tt := range tests {
                frag := fmt.Sprintf("  0\nSECTION\n  2\nENTITIES\n"+
                        "  0\nCIRCLE\n  8\n0\n 10\n0\n 20\n0\n 40\n1\n 62\n%d\n"+
                        "  0\nENDSEC\n  0\nEOF\n", tt.aci)
                d, _, err := dxf.ReadString(frag)
                if err != nil {
                        t.Fatal(err)
                }
                ents := d.Entities()
                if len(ents) == 0 {
                        t.Errorf("aci=%d: no entities", tt.aci)
                        continue
                }
                got := strings.ToUpper(ents[0].Color)
                if got != tt.want {
                        t.Errorf("aci=%d: color=%q, want %q", tt.aci, got, tt.want)
                }
        }
}

func TestElapseFrag(t *testing.T) {
        frag := "  0\nSECTION\n  2\nENTITIES\n" +
                "  0\nELLIPSE\n  8\n0\n 10\n0\n 20\n0\n 11\n10.0\n 21\n0.0\n 40\n0.5\n 41\n0\n 42\n6.283185\n" +
                "  0\nENDSEC\n  0\nEOF\n"
        d, _, err := dxf.ReadString(frag)
        if err != nil {
                t.Fatal(err)
        }
        ents := d.Entities()
        if len(ents) != 1 {
                t.Fatalf("got %d entities, want 1", len(ents))
        }
        e := ents[0]
        if e.Type != document.TypeEllipse {
                t.Errorf("type = %q, want ellipse", e.Type)
        }
        if !approxEq(e.R, 10.0) {
                t.Errorf("semi-major = %.3f, want 10", e.R)
        }
        if !approxEq(e.R2, 5.0) {
                t.Errorf("semi-minor = %.3f, want 5", e.R2)
        }
}

func TestWriterString(t *testing.T) {
        d := document.New()
        d.AddCircle(0, 0, 5, 0, "#ffffff")
        s := dxf.String(d)
        if !strings.Contains(s, "CIRCLE") {
                t.Error("expected CIRCLE in DXF output")
        }
        if !strings.Contains(s, "AC1015") {
                t.Error("expected AC1015 in DXF output")
        }
}

// ─── Undo / layer preservation tests ─────────────────────────────────────────

// TestLoadDXFUndo_RestoresLayerState verifies that undoing a DXF import via
// document.LoadDXFBytes fully restores the previous layer state, not just the
// entity list. This is the regression test for the undo snapshot bug where
// undoStack only stored entities and left imported layers behind.
func TestLoadDXFUndo_RestoresLayerState(t *testing.T) {
        document.RegisterDXFReader(func(data []byte) (*document.Document, []string, error) {
                return dxf.Read(strings.NewReader(string(data)))
        })

        // Step 1: build a base document with one entity on a named layer.
        base := document.New()
        base.AddLayer("BaseLayer", "#ff0000", document.LineTypeSolid, 0.25)
        base.AddLine(0, 0, 10, 10, 0, "#ffffff")
        baseLayerCount := len(base.Layers())
        baseEntityCount := base.EntityCount()

        // Step 2: produce a DXF that introduces a new layer "ImportedLayer".
        dxfInput := strings.Join([]string{
                "  0\nSECTION", "  2\nTABLES",
                "  0\nTABLE", "  2\nLAYER",
                "  0\nLAYER", "  2\nImportedLayer", " 62\n3", " 6\nContinuous",
                "  0\nENDTAB", "  0\nENDSEC",
                "  0\nSECTION", "  2\nENTITIES",
                "  0\nLINE", "  8\nImportedLayer",
                " 10\n5.0", " 20\n5.0", " 11\n15.0", " 21\n15.0",
                "  0\nENDSEC", "  0\nEOF",
        }, "\n")

        warns, err := base.LoadDXFBytes([]byte(dxfInput))
        if err != nil {
                t.Fatalf("LoadDXFBytes: %v", err)
        }
        logWarns(t, warns)

        // After import, the document should have ImportedLayer.
        found := false
        for _, l := range base.Layers() {
                if l.Name == "ImportedLayer" {
                        found = true
                }
        }
        if !found {
                t.Fatal("expected ImportedLayer after import")
        }

        // Step 3: undo the import — must restore original layers AND entities.
        if !base.Undo() {
                t.Fatal("Undo() returned false")
        }
        if base.EntityCount() != baseEntityCount {
                t.Errorf("after Undo: entity count = %d, want %d", base.EntityCount(), baseEntityCount)
        }
        if len(base.Layers()) != baseLayerCount {
                t.Errorf("after Undo: layer count = %d, want %d", len(base.Layers()), baseLayerCount)
        }
        for _, l := range base.Layers() {
                if l.Name == "ImportedLayer" {
                        t.Error("after Undo: ImportedLayer should not exist (was introduced by the import)")
                }
        }
}

// TestLoadDXFLayerProperties verifies that layer color, linetype, visible, and
// locked flags from the DXF LAYER table are correctly applied to imported layers.
func TestLoadDXFLayerProperties(t *testing.T) {
        dxfInput := strings.Join([]string{
                "  0\nSECTION", "  2\nTABLES",
                "  0\nTABLE", "  2\nLAYER",
                // ACI 1 = red; linetype Dashed; flags 4 = frozen+locked (bit 4=frozen, bit 2=locked)
                "  0\nLAYER", "  2\nSpecialLayer", " 62\n1", " 6\nDashed", " 70\n4",
                "  0\nENDTAB", "  0\nENDSEC",
                "  0\nSECTION", "  2\nENTITIES",
                "  0\nLINE", "  8\nSpecialLayer",
                " 10\n0.0", " 20\n0.0", " 11\n1.0", " 21\n1.0",
                "  0\nENDSEC", "  0\nEOF",
        }, "\n")

        doc, warns, err := dxf.ReadString(dxfInput)
        if err != nil {
                t.Fatalf("ReadString: %v", err)
        }
        logWarns(t, warns)

        var special *document.Layer
        for _, l := range doc.Layers() {
                if l.Name == "SpecialLayer" {
                        special = l
                }
        }
        if special == nil {
                t.Fatal("SpecialLayer not found in imported document")
        }
        // ACI 1 = red = #FF0000 (aciToRGB returns uppercase hex)
        if strings.ToUpper(special.Color) != "#FF0000" {
                t.Errorf("color = %q, want red (#FF0000 or #ff0000)", special.Color)
        }
        if special.LineTyp != document.LineTypeDashed {
                t.Errorf("lineType = %q, want Dashed", special.LineTyp)
        }
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func approxEq(a, b float64) bool { return math.Abs(a-b) < 1e-4 }

func logWarns(t *testing.T, w []string) {
        t.Helper()
        for _, msg := range w {
                t.Logf("warn: %s", msg)
        }
}

func layerNames(layers []*document.Layer) []string {
        names := make([]string, len(layers))
        for i, l := range layers {
                names[i] = l.Name
        }
        return names
}
