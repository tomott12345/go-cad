package document

// DXF round-trip tests for Task #7 entity types:
//   TypeBlockRef → INSERT (with BLOCK/ENDBLK section)
//   TypeHatch    → HATCH (R2000) / POLYLINE (R12)
//   TypeLeader   → LEADER (R2000) / LINE series (R12)
//   TypeRevisionCloud / TypeWipeout → LWPOLYLINE (R2000) / POLYLINE (R12)
//
// Each test: create document → add entity → export DXF → check group codes.

import (
        "strings"
        "testing"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

// hasDXFToken returns true if consecutive lines in the DXF string, after
// trimming whitespace, equal code and value respectively.
func hasDXFToken(dxf, code, value string) bool {
        wantCode := strings.TrimSpace(code)
        wantVal := strings.TrimSpace(value)
        lines := strings.Split(dxf, "\n")
        for i := 0; i < len(lines)-1; i++ {
                if strings.TrimSpace(lines[i]) == wantCode &&
                        strings.TrimSpace(lines[i+1]) == wantVal {
                        return true
                }
        }
        return false
}

func newDocWithSymbol(t *testing.T, name string) *Document {
        t.Helper()
        doc := New()
        // Add a line to define into a block.
        id := doc.AddLine(0, 0, 10, 0, 0, "#ffffff")
        ok := doc.DefineBlock(name, 0, 0, []int{id})
        if !ok {
                t.Fatalf("DefineBlock(%q) returned false", name)
        }
        return doc
}

// ─── Block / INSERT ───────────────────────────────────────────────────────────

func TestDXF_BlockSectionEmitted(t *testing.T) {
        doc := newDocWithSymbol(t, "MYBLOCK")
        dxf := doc.ExportDXF()
        if !strings.Contains(dxf, "  0\nSECTION\n  2\nBLOCKS\n") {
                t.Error("R2000 DXF missing BLOCKS section header")
        }
        if !strings.Contains(dxf, "*Model_Space") {
                t.Error("R2000 DXF missing mandatory *Model_Space block")
        }
}

func TestDXF_BlockDefinitionEmitted(t *testing.T) {
        doc := newDocWithSymbol(t, "MYBLOCK")
        dxf := doc.ExportDXF()
        // BLOCK record with name MYBLOCK.
        if !hasDXFToken(dxf, "  2", "MYBLOCK") {
                t.Error("R2000 DXF missing BLOCK definition for MYBLOCK")
        }
        if !strings.Contains(dxf, "  0\nENDBLK\n") {
                t.Error("R2000 DXF missing ENDBLK for user block")
        }
}

func TestDXF_BlockBasepointIsZero(t *testing.T) {
        // Entities stored in local coords (base=origin) → DXF BLOCK base point
        // must be (0,0). Using the original world BaseX/BaseY would double-offset
        // INSERT placement in downstream CAD readers.
        doc := New()
        id := doc.AddLine(50, 30, 60, 30, 0, "#ffffff")
        ok := doc.DefineBlock("SHIFTED", 50, 30, []int{id})
        if !ok {
                t.Fatal("DefineBlock for SHIFTED failed")
        }
        dxf := doc.ExportDXF()
        // Find the BLOCK definition for SHIFTED.
        marker := "SHIFTED"
        idx := strings.Index(dxf, "  2\n"+marker)
        if idx < 0 {
                // Try without leading spaces (TrimSpace handles either format).
                idx = strings.Index(dxf, marker)
        }
        if idx < 0 {
                t.Fatal("SHIFTED block not found in DXF")
        }
        // Limit blockSection to just the SHIFTED BLOCK … ENDBLK region so we
        // don't accidentally read the ENTITIES section where the original (50,30)
        // line still lives.
        endblkIdx := strings.Index(dxf[idx:], "ENDBLK")
        blockSection := dxf[idx:]
        if endblkIdx >= 0 {
                blockSection = dxf[idx : idx+endblkIdx+len("ENDBLK")]
        }
        // The BLOCK base-point group-code-10 value must be 0 (not 50).
        if strings.Contains(blockSection, " 10\n50.") {
                t.Errorf("BLOCK base point is 50 (world coord); expected 0.0\n--- block section ---\n%s", blockSection)
        }
        // After shifting by (50,30): x1=0, x2=10. The world x1=50 must NOT appear.
        if strings.Contains(blockSection, " 10\n50.000000") {
                t.Errorf("block-local entity x1 is 50 (world coord); expected 0.0\n--- block section ---\n%s", blockSection)
        }
}

func TestDXF_InsertEntityEmitted_R2000(t *testing.T) {
        doc := newDocWithSymbol(t, "WIDGET")
        id := doc.InsertBlock("WIDGET", 100, 200, 1, 1, 0, 0, "#ffffff")
        if id < 0 {
                t.Fatal("InsertBlock returned -1")
        }
        dxf := doc.ExportDXF()
        if !hasDXFToken(dxf, "  0", "INSERT") {
                t.Error("R2000 DXF missing INSERT entity")
        }
        // Block name reference in INSERT.
        if !hasDXFToken(dxf, "  2", "WIDGET") {
                t.Error("R2000 INSERT missing block name reference WIDGET")
        }
}

func TestDXF_InsertEntityEmitted_R12(t *testing.T) {
        doc := newDocWithSymbol(t, "WIDGET")
        doc.InsertBlock("WIDGET", 0, 0, 2, 2, 45, 0, "#ffffff")
        dxf := doc.ExportDXFR12()
        if !hasDXFToken(dxf, "  0", "INSERT") {
                t.Error("R12 DXF missing INSERT entity")
        }
}

func TestDXF_BlockSectionBeforeEntitiesSection(t *testing.T) {
        doc := newDocWithSymbol(t, "ORDER")
        dxf := doc.ExportDXF()
        blocksIdx := strings.Index(dxf, "  2\nBLOCKS")
        entitiesIdx := strings.Index(dxf, "  2\nENTITIES")
        if blocksIdx < 0 {
                t.Fatal("BLOCKS section not found")
        }
        if entitiesIdx < 0 {
                t.Fatal("ENTITIES section not found")
        }
        if blocksIdx > entitiesIdx {
                t.Errorf("BLOCKS section appears AFTER ENTITIES section (at %d vs %d); DXF spec requires BLOCKS before ENTITIES", blocksIdx, entitiesIdx)
        }
}

func TestDXF_InsertBlockUndefinedReturnsNegOne(t *testing.T) {
        doc := New()
        id := doc.InsertBlock("NOSUCHBLOCK", 0, 0, 1, 1, 0, 0, "#ffffff")
        if id != -1 {
                t.Errorf("InsertBlock for undefined block should return -1, got %d", id)
        }
}

func TestDXF_BlocksSavedAndLoadedViaJSON(t *testing.T) {
        tmp := t.TempDir() + "/blocks.json"
        doc := newDocWithSymbol(t, "PERSIST_ME")
        if err := doc.Save(tmp); err != nil {
                t.Fatalf("Save failed: %v", err)
        }
        doc2 := New()
        if err := doc2.Load(tmp); err != nil {
                t.Fatalf("Load failed: %v", err)
        }
        blk := doc2.BlockByName("PERSIST_ME")
        if blk == nil {
                t.Fatal("block PERSIST_ME not found after Save/Load round-trip")
        }
        if len(blk.Entities) == 0 {
                t.Error("loaded block has no entities")
        }
}

// ─── Hatch ────────────────────────────────────────────────────────────────────

func TestDXF_HatchEntityR2000(t *testing.T) {
        doc := New()
        pts := [][]float64{{0, 0}, {50, 0}, {50, 50}, {0, 50}}
        doc.AddHatch(pts, "ANSI31", 0, 5, 0, "#ffffff")
        dxf := doc.ExportDXF()
        if !hasDXFToken(dxf, "  0", "HATCH") {
                t.Error("R2000 DXF missing HATCH entity")
        }
        if !hasDXFToken(dxf, "  2", "ANSI31") {
                t.Error("HATCH entity missing pattern name ANSI31")
        }
}

func TestDXF_HatchEntityR12FallsBackToPolyline(t *testing.T) {
        doc := New()
        pts := [][]float64{{0, 0}, {50, 0}, {50, 50}, {0, 50}}
        doc.AddHatch(pts, "ANSI31", 0, 5, 0, "#ffffff")
        dxf := doc.ExportDXFR12()
        // R12 has no HATCH entity; exported as POLYLINE.
        if !hasDXFToken(dxf, "  0", "POLYLINE") {
                t.Error("R12 DXF missing POLYLINE fallback for HATCH entity")
        }
}

func TestDXF_HatchSOLID(t *testing.T) {
        doc := New()
        pts := [][]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
        doc.AddHatch(pts, "SOLID", 0, 1, 0, "#ffffff")
        dxf := doc.ExportDXF()
        if !hasDXFToken(dxf, "  0", "HATCH") {
                t.Error("SOLID HATCH missing from R2000 DXF")
        }
}

// ─── Leader ───────────────────────────────────────────────────────────────────

func TestDXF_LeaderEntityR2000(t *testing.T) {
        doc := New()
        pts := [][]float64{{0, 0}, {20, 10}, {40, 10}}
        doc.AddLeader(pts, "SEE NOTE", 0, "#ffffff")
        dxf := doc.ExportDXF()
        if !hasDXFToken(dxf, "  0", "LEADER") {
                t.Error("R2000 DXF missing LEADER entity")
        }
}

func TestDXF_LeaderEntityR12FallsBackToLines(t *testing.T) {
        doc := New()
        pts := [][]float64{{0, 0}, {20, 10}, {40, 10}}
        doc.AddLeader(pts, "NOTE", 0, "#ffffff")
        dxf := doc.ExportDXFR12()
        // R12: leader exported as LINE entities.
        if !hasDXFToken(dxf, "  0", "LINE") {
                t.Error("R12 DXF missing LINE fallback for LEADER entity")
        }
}

// ─── RevisionCloud ────────────────────────────────────────────────────────────

func TestDXF_RevisionCloudR2000ExportsAsLWPOLYLINE(t *testing.T) {
        doc := New()
        pts := [][]float64{{0, 0}, {30, 0}, {30, 30}, {0, 30}}
        doc.AddRevisionCloud(pts, 10, 0, "#ffffff")
        dxf := doc.ExportDXF()
        if !hasDXFToken(dxf, "  0", "LWPOLYLINE") {
                t.Error("R2000 DXF missing LWPOLYLINE for RevisionCloud")
        }
}

func TestDXF_RevisionCloudR12ExportsAsPOLYLINE(t *testing.T) {
        doc := New()
        pts := [][]float64{{0, 0}, {30, 0}, {30, 30}, {0, 30}}
        doc.AddRevisionCloud(pts, 10, 0, "#ffffff")
        dxf := doc.ExportDXFR12()
        if !hasDXFToken(dxf, "  0", "POLYLINE") {
                t.Error("R12 DXF missing POLYLINE for RevisionCloud")
        }
}

// ─── Wipeout ──────────────────────────────────────────────────────────────────

func TestDXF_WipeoutR2000ExportsAsLWPOLYLINE(t *testing.T) {
        doc := New()
        pts := [][]float64{{0, 0}, {20, 0}, {20, 20}, {0, 20}}
        doc.AddWipeout(pts, 0, "#ffffff")
        dxf := doc.ExportDXF()
        if !hasDXFToken(dxf, "  0", "LWPOLYLINE") {
                t.Error("R2000 DXF missing LWPOLYLINE for Wipeout")
        }
}

// ─── Block R12 BLOCKS section ─────────────────────────────────────────────────

func TestDXF_BlockSectionEmitted_R12(t *testing.T) {
        doc := newDocWithSymbol(t, "R12BLK")
        dxf := doc.ExportDXFR12()
        if !strings.Contains(dxf, "  0\nSECTION\n  2\nBLOCKS\n") {
                t.Error("R12 DXF missing BLOCKS section header")
        }
        if !strings.Contains(dxf, "*Model_Space") {
                t.Error("R12 DXF missing *Model_Space block")
        }
        if !hasDXFToken(dxf, "  2", "R12BLK") {
                t.Error("R12 DXF missing BLOCK definition for R12BLK")
        }
}
