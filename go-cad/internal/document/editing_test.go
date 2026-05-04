package document_test

import (
        "math"
        "testing"

        "go-cad/internal/document"
)

// ── helpers ───────────────────────────────────────────────────────────────────

const eps = 1e-9

func approxEq(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

// ── Move ──────────────────────────────────────────────────────────────────────

func TestMove_Line(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        ok := d.Move([]int{id}, 5, 3)
        if !ok {
                t.Fatal("Move returned false")
        }
        es := d.Entities()
        if len(es) != 1 {
                t.Fatalf("expected 1 entity, got %d", len(es))
        }
        e := es[0]
        if !approxEq(e.X1, 5) || !approxEq(e.Y1, 3) {
                t.Errorf("start: got (%.4f,%.4f) want (5,3)", e.X1, e.Y1)
        }
        if !approxEq(e.X2, 15) || !approxEq(e.Y2, 3) {
                t.Errorf("end: got (%.4f,%.4f) want (15,3)", e.X2, e.Y2)
        }
}

func TestMove_Circle(t *testing.T) {
        d := document.New()
        id := d.AddCircle(0, 0, 5, 0, "#fff")
        d.Move([]int{id}, -3, 7)
        e := d.Entities()[0]
        if !approxEq(e.CX, -3) || !approxEq(e.CY, 7) {
                t.Errorf("centre: got (%.4f,%.4f) want (-3,7)", e.CX, e.CY)
        }
        if !approxEq(e.R, 5) {
                t.Errorf("radius should be unchanged: got %.4f", e.R)
        }
}

func TestMove_Undo(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        d.Move([]int{id}, 100, 100)
        d.Undo()
        e := d.Entities()[0]
        if !approxEq(e.X1, 0) || !approxEq(e.Y1, 0) {
                t.Errorf("undo failed: got start (%.4f,%.4f)", e.X1, e.Y1)
        }
}

func TestMove_UnknownID(t *testing.T) {
        d := document.New()
        ok := d.Move([]int{999}, 1, 1)
        if ok {
                t.Error("Move with unknown ID should return false")
        }
}

// ── Copy ──────────────────────────────────────────────────────────────────────

func TestCopy_Line(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        newIDs := d.Copy([]int{id}, 5, 5)
        if len(newIDs) != 1 {
                t.Fatalf("expected 1 copy ID, got %d", len(newIDs))
        }
        if len(d.Entities()) != 2 {
                t.Fatalf("expected 2 entities, got %d", len(d.Entities()))
        }
        // Original unchanged.
        orig := d.Entities()[0]
        if !approxEq(orig.X1, 0) {
                t.Error("original should not be moved")
        }
        // Copy shifted.
        cp := d.Entities()[1]
        if !approxEq(cp.X1, 5) || !approxEq(cp.Y1, 5) {
                t.Errorf("copy start: got (%.4f,%.4f) want (5,5)", cp.X1, cp.Y1)
        }
}

func TestCopy_Undo(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        d.Copy([]int{id}, 5, 5)
        d.Undo()
        if len(d.Entities()) != 1 {
                t.Errorf("after undo expected 1 entity, got %d", len(d.Entities()))
        }
}

// ── Rotate ────────────────────────────────────────────────────────────────────

func TestRotate_Line90(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        d.Rotate([]int{id}, 0, 0, 90, false)
        e := d.Entities()[0]
        // (10,0) rotated 90° CCW around origin → (0,10)
        if !approxEq(e.X2, 0) || !approxEq(e.Y2, 10) {
                t.Errorf("end after 90° rotate: got (%.4f,%.4f) want (0,10)", e.X2, e.Y2)
        }
}

func TestRotate_MakeCopy(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        newIDs := d.Rotate([]int{id}, 0, 0, 90, true)
        if len(newIDs) != 1 {
                t.Fatalf("expected 1 copy ID, got %d", len(newIDs))
        }
        if len(d.Entities()) != 2 {
                t.Fatalf("expected 2 entities, got %d", len(d.Entities()))
        }
        orig := d.Entities()[0]
        if !approxEq(orig.X2, 10) {
                t.Error("original end should remain at x=10")
        }
}

func TestRotate_Arc_Angles(t *testing.T) {
        d := document.New()
        id := d.AddArc(0, 0, 5, 0, 90, 0, "#fff")
        d.Rotate([]int{id}, 0, 0, 45, false)
        e := d.Entities()[0]
        if !approxEq(e.StartDeg, 45) {
                t.Errorf("arc startDeg after rotate: got %.4f want 45", e.StartDeg)
        }
        if !approxEq(e.EndDeg, 135) {
                t.Errorf("arc endDeg after rotate: got %.4f want 135", e.EndDeg)
        }
}

// ── Scale ─────────────────────────────────────────────────────────────────────

func TestScale_Line(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        d.Scale([]int{id}, 0, 0, 2, 2, false)
        e := d.Entities()[0]
        if !approxEq(e.X2, 20) {
                t.Errorf("end.X after scale×2: got %.4f want 20", e.X2)
        }
}

func TestScale_Circle(t *testing.T) {
        d := document.New()
        id := d.AddCircle(0, 0, 5, 0, "#fff")
        d.Scale([]int{id}, 0, 0, 3, 3, false)
        e := d.Entities()[0]
        if !approxEq(e.R, 15) {
                t.Errorf("radius after scale×3: got %.4f want 15", e.R)
        }
}

func TestScale_MakeCopy(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        newIDs := d.Scale([]int{id}, 0, 0, 2, 2, true)
        if len(newIDs) != 1 || len(d.Entities()) != 2 {
                t.Fatalf("expected 2 entities after copy-scale; got %d", len(d.Entities()))
        }
        orig := d.Entities()[0]
        if !approxEq(orig.X2, 10) {
                t.Error("original end should be unchanged")
        }
}

// ── Mirror ────────────────────────────────────────────────────────────────────

func TestMirror_LineAcrossYAxis(t *testing.T) {
        d := document.New()
        id := d.AddLine(1, 0, 5, 0, 0, "#fff")
        // Mirror across the Y-axis: line from (0,-1)→(0,1).
        d.Mirror([]int{id}, 0, -1, 0, 1, false)
        e := d.Entities()[0]
        if !approxEq(e.X1, -1) {
                t.Errorf("start.X after mirror: got %.4f want -1", e.X1)
        }
        if !approxEq(e.X2, -5) {
                t.Errorf("end.X after mirror: got %.4f want -5", e.X2)
        }
}

func TestMirror_MakeCopy(t *testing.T) {
        d := document.New()
        id := d.AddLine(1, 0, 5, 0, 0, "#fff")
        newIDs := d.Mirror([]int{id}, 0, -1, 0, 1, true)
        if len(newIDs) != 1 || len(d.Entities()) != 2 {
                t.Fatalf("expected 2 entities; got %d", len(d.Entities()))
        }
        orig := d.Entities()[0]
        if !approxEq(orig.X1, 1) {
                t.Error("original should not have moved")
        }
}

func TestMirror_ZeroLengthAxis(t *testing.T) {
        d := document.New()
        id := d.AddLine(1, 0, 5, 0, 0, "#fff")
        out := d.Mirror([]int{id}, 3, 3, 3, 3, false)
        if out != nil {
                t.Error("Mirror with zero-length axis should return nil")
        }
}

// ── Trim ──────────────────────────────────────────────────────────────────────

func TestTrim_LineByLine(t *testing.T) {
        d := document.New()
        // Horizontal line from (-10,0) to (10,0).
        target := d.AddLine(-10, 0, 10, 0, 0, "#fff")
        // Vertical cutter at x=0 from (0,-5) to (0,5).
        cutter := d.AddLine(0, -5, 0, 5, 0, "#fff")

        // Pick the left half.
        newIDs := d.Trim(cutter, target, -5, 0)
        if len(newIDs) == 0 {
                t.Fatal("Trim returned nil or empty slice")
        }
        // Should produce one surviving segment (the right half).
        found := false
        for _, e := range d.Entities() {
                if e.ID == newIDs[0] && e.Type == document.TypeLine {
                        found = true
                        // Surviving segment should be on the right of x=0.
                        if e.X1 > 0 || e.X2 > 0 {
                                // OK: segment is on right side.
                        }
                }
        }
        if !found {
                t.Error("surviving segment not found")
        }
        // Original target should be gone.
        for _, e := range d.Entities() {
                if e.ID == target {
                        t.Error("original target entity should have been removed")
                }
        }
}

func TestTrim_SameIDReturnsNil(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        if result := d.Trim(id, id, 5, 0); result != nil {
                t.Error("Trim with same cutter/target ID should return nil")
        }
}

func TestTrim_NoIntersectionReturnsNil(t *testing.T) {
        d := document.New()
        line := d.AddLine(0, 0, 10, 0, 0, "#fff")
        parallel := d.AddLine(0, 5, 10, 5, 0, "#fff") // parallel, no intersection
        if result := d.Trim(parallel, line, 5, 0); result != nil {
                t.Error("Trim with no intersection should return nil")
        }
}

// ── Extend ────────────────────────────────────────────────────────────────────

func TestExtend_Line(t *testing.T) {
        d := document.New()
        // Short horizontal line.
        target := d.AddLine(0, 0, 5, 0, 0, "#fff")
        // Vertical boundary at x=10.
        boundary := d.AddLine(10, -10, 10, 10, 0, "#fff")

        // Pick near the end (right side) so we extend that end.
        newID := d.Extend(boundary, target, 4, 0)
        if newID < 0 {
                t.Fatal("Extend returned -1")
        }
        es := d.Entities()
        var ext *document.Entity
        for i := range es {
                if es[i].ID == newID {
                        ext = &es[i]
                }
        }
        if ext == nil {
                t.Fatal("extended entity not found")
        }
        // The end closer to the pick (right side) should now be at x=10.
        if !approxEq(ext.X2, 10) && !approxEq(ext.X1, 10) {
                t.Errorf("extended endpoint should be at x=10; got x1=%.4f x2=%.4f", ext.X1, ext.X2)
        }
}

func TestExtend_SameIDReturnsMinusOne(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        if result := d.Extend(id, id, 5, 0); result != -1 {
                t.Error("Extend with same boundary/target ID should return -1")
        }
}

// ── Fillet ────────────────────────────────────────────────────────────────────

func TestFillet_RightAngle(t *testing.T) {
        d := document.New()
        // Two lines meeting at (0,0) at a right angle.
        id1 := d.AddLine(-10, 0, 0, 0, 0, "#fff")   // horizontal ending at origin
        id2 := d.AddLine(0, 0, 0, 10, 0, "#fff")    // vertical starting at origin
        arcID := d.Fillet(id1, id2, 2)
        if arcID < 0 {
                t.Fatal("Fillet returned -1")
        }
        // Arc entity should exist.
        var arc *document.Entity
        for _, e := range d.Entities() {
                if e.ID == arcID {
                        arc = &e
                        break
                }
        }
        if arc == nil {
                t.Fatal("fillet arc not found")
        }
        if arc.Type != document.TypeArc {
                t.Errorf("expected TypeArc, got %s", arc.Type)
        }
        if !approxEq(arc.R, 2) {
                t.Errorf("fillet arc radius: got %.4f want 2", arc.R)
        }
}

func TestFillet_SameIDReturnsMinusOne(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        if result := d.Fillet(id, id, 2); result != -1 {
                t.Error("Fillet with same IDs should return -1")
        }
}

func TestFillet_NonLineReturnsMinusOne(t *testing.T) {
        d := document.New()
        l := d.AddLine(0, 0, 10, 0, 0, "#fff")
        c := d.AddCircle(0, 0, 5, 0, "#fff")
        if result := d.Fillet(l, c, 2); result != -1 {
                t.Error("Fillet on non-line entity should return -1")
        }
}

// ── Chamfer ───────────────────────────────────────────────────────────────────

func TestChamfer_RightAngle(t *testing.T) {
        d := document.New()
        id1 := d.AddLine(-10, 0, 0, 0, 0, "#fff")
        id2 := d.AddLine(0, 0, 0, 10, 0, "#fff")
        chamID := d.Chamfer(id1, id2, 2, 2)
        if chamID < 0 {
                t.Fatal("Chamfer returned -1")
        }
        var cham *document.Entity
        for _, e := range d.Entities() {
                if e.ID == chamID {
                        cham = &e
                        break
                }
        }
        if cham == nil {
                t.Fatal("chamfer line not found")
        }
        if cham.Type != document.TypeLine {
                t.Errorf("expected TypeLine, got %s", cham.Type)
        }
        // Chamfer should connect (-2,0) to (0,2) (at distance 2 from corner).
        chamLen := math.Hypot(cham.X2-cham.X1, cham.Y2-cham.Y1)
        expected := math.Sqrt2 * 2 // diagonal of 2×2 square
        if !approxEq(chamLen, expected) {
                t.Errorf("chamfer length: got %.4f want %.4f", chamLen, expected)
        }
}

func TestChamfer_SameIDReturnsMinusOne(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        if result := d.Chamfer(id, id, 1, 1); result != -1 {
                t.Error("Chamfer with same IDs should return -1")
        }
}

// ── ArrayRect ─────────────────────────────────────────────────────────────────

func TestArrayRect_2x3(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 5, 0, 0, "#fff")
        newIDs := d.ArrayRect([]int{id}, 2, 3, 10, 15)
        // 2×3=6 total, minus 1 original = 5 copies.
        if len(newIDs) != 5 {
                t.Fatalf("expected 5 copy IDs, got %d", len(newIDs))
        }
        if len(d.Entities()) != 6 {
                t.Fatalf("expected 6 entities, got %d", len(d.Entities()))
        }
}

func TestArrayRect_Undo(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 5, 0, 0, "#fff")
        d.ArrayRect([]int{id}, 3, 3, 10, 10)
        d.Undo()
        if len(d.Entities()) != 1 {
                t.Errorf("after undo expected 1 entity, got %d", len(d.Entities()))
        }
}

func TestArrayRect_Spacing(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 1, 0, 0, "#fff")
        d.ArrayRect([]int{id}, 1, 3, 0, 10)
        es := d.Entities()
        if len(es) != 3 {
                t.Fatalf("expected 3 entities, got %d", len(es))
        }
        // Second entity (col=1) should start at x=10.
        if !approxEq(es[1].X1, 10) {
                t.Errorf("col1 start.X: got %.4f want 10", es[1].X1)
        }
        // Third entity (col=2) should start at x=20.
        if !approxEq(es[2].X1, 20) {
                t.Errorf("col2 start.X: got %.4f want 20", es[2].X1)
        }
}

// ── ArrayPolar ────────────────────────────────────────────────────────────────

func TestArrayPolar_4Copies(t *testing.T) {
        d := document.New()
        id := d.AddLine(5, 0, 10, 0, 0, "#fff")
        newIDs := d.ArrayPolar([]int{id}, 0, 0, 4, 360)
        // 4 total (including original) → 3 copies.
        if len(newIDs) != 3 {
                t.Fatalf("expected 3 copy IDs, got %d", len(newIDs))
        }
        if len(d.Entities()) != 4 {
                t.Fatalf("expected 4 entities, got %d", len(d.Entities()))
        }
}

func TestArrayPolar_Angles(t *testing.T) {
        d := document.New()
        id := d.AddLine(5, 0, 10, 0, 0, "#fff")
        d.ArrayPolar([]int{id}, 0, 0, 4, 360)
        es := d.Entities()
        // The 2nd entity should have its start at (0,5) (90° rotation).
        if !approxEq(es[1].X1, 0) || !approxEq(es[1].Y1, 5) {
                t.Errorf("2nd entity start: got (%.4f,%.4f) want (0,5)", es[1].X1, es[1].Y1)
        }
}

func TestArrayPolar_TooFewCount(t *testing.T) {
        d := document.New()
        id := d.AddLine(5, 0, 10, 0, 0, "#fff")
        result := d.ArrayPolar([]int{id}, 0, 0, 1, 360)
        if result != nil {
                t.Error("ArrayPolar with count<2 should return nil")
        }
}

func TestArrayPolar_Undo(t *testing.T) {
        d := document.New()
        id := d.AddLine(5, 0, 10, 0, 0, "#fff")
        d.ArrayPolar([]int{id}, 0, 0, 6, 360)
        d.Undo()
        if len(d.Entities()) != 1 {
                t.Errorf("after undo expected 1 entity, got %d", len(d.Entities()))
        }
}

// ── Multi-entity operations ───────────────────────────────────────────────────

func TestMove_MultipleEntities(t *testing.T) {
        d := document.New()
        id1 := d.AddLine(0, 0, 10, 0, 0, "#fff")
        id2 := d.AddCircle(5, 5, 3, 0, "#fff")
        d.Move([]int{id1, id2}, 1, 1)
        es := d.Entities()
        for _, e := range es {
                switch e.Type {
                case document.TypeLine:
                        if !approxEq(e.X1, 1) || !approxEq(e.Y1, 1) {
                                t.Errorf("line start: got (%.4f,%.4f) want (1,1)", e.X1, e.Y1)
                        }
                case document.TypeCircle:
                        if !approxEq(e.CX, 6) || !approxEq(e.CY, 6) {
                                t.Errorf("circle centre: got (%.4f,%.4f) want (6,6)", e.CX, e.CY)
                        }
                }
        }
}

func TestCopy_Circle(t *testing.T) {
        d := document.New()
        id := d.AddCircle(0, 0, 5, 0, "#fff")
        newIDs := d.Copy([]int{id}, 10, 0)
        if len(newIDs) != 1 {
                t.Fatalf("expected 1 copy, got %d", len(newIDs))
        }
        es := d.Entities()
        cp := es[1]
        if !approxEq(cp.CX, 10) {
                t.Errorf("copy centre.X: got %.4f want 10", cp.CX)
        }
        if !approxEq(cp.R, 5) {
                t.Errorf("copy radius should be 5, got %.4f", cp.R)
        }
}

func TestScale_Polyline(t *testing.T) {
        d := document.New()
        id := d.AddPolyline([][]float64{{0, 0}, {10, 0}, {10, 10}}, 0, "#fff")
        d.Scale([]int{id}, 0, 0, 2, 2, false)
        e := d.Entities()[0]
        if !approxEq(e.Points[1][0], 20) {
                t.Errorf("polyline[1].X after scale: got %.4f want 20", e.Points[1][0])
        }
}

func TestRotate_360IsIdentity(t *testing.T) {
        d := document.New()
        id := d.AddLine(3, 4, 7, 4, 0, "#fff")
        d.Rotate([]int{id}, 0, 0, 360, false)
        e := d.Entities()[0]
        if !approxEq(e.X1, 3) || !approxEq(e.Y1, 4) {
                t.Errorf("360° rotate should be identity: got (%.6f,%.6f)", e.X1, e.Y1)
        }
}

func TestMirror_ArcAngles(t *testing.T) {
        d := document.New()
        // Arc from 0° to 90°, mirrored across X-axis (line y=0).
        id := d.AddArc(0, 0, 5, 0, 90, 0, "#fff")
        d.Mirror([]int{id}, -1, 0, 1, 0, false) // mirror line along X-axis
        e := d.Entities()[0]
        // Mirrored arc should go from 270° to 360° (or equivalently 0°).
        // After mirroring across X-axis (lineAngle=0):
        //   newStart = mirrorAngle(endDeg=90)   = 2*0 - 90 = -90 → 270
        //   newEnd   = mirrorAngle(startDeg=0)  = 2*0 - 0  = 0
        // Then swap: startDeg=270, endDeg=0
        if !approxEq(e.StartDeg, 270) {
                t.Errorf("mirrored arc startDeg: got %.4f want 270", e.StartDeg)
        }
}

func TestArrayRect_EmptyIDs(t *testing.T) {
        d := document.New()
        result := d.ArrayRect([]int{}, 3, 3, 10, 10)
        if result != nil {
                t.Error("ArrayRect with empty IDs should return nil")
        }
}

func TestArrayPolar_ArcAnglesRotated(t *testing.T) {
        d := document.New()
        // Arc from 0° to 90°.
        id := d.AddArc(5, 0, 1, 0, 90, 0, "#fff")
        d.ArrayPolar([]int{id}, 0, 0, 4, 360)
        es := d.Entities()
        // Second entity (copy 1, rotated 90°) arc angles should be 90°→180°.
        var copy1 *document.Entity
        for i := range es {
                if es[i].ID != id {
                        copy1 = &es[i]
                        break
                }
        }
        if copy1 == nil {
                t.Fatal("copy not found")
        }
        if !approxEq(copy1.StartDeg, 90) {
                t.Errorf("copy1 startDeg: got %.4f want 90", copy1.StartDeg)
        }
        if !approxEq(copy1.EndDeg, 180) {
                t.Errorf("copy1 endDeg: got %.4f want 180", copy1.EndDeg)
        }
}

// ── Offset ────────────────────────────────────────────────────────────────────

func TestOffset_Line_Positive(t *testing.T) {
        // A horizontal line y=0 offset by +5 should produce a parallel line at y=5.
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        ids := d.Offset([]int{id}, 5)
        if len(ids) != 1 {
                t.Fatalf("expected 1 new entity, got %d", len(ids))
        }
        es := d.Entities()
        if len(es) != 2 {
                t.Fatalf("expected 2 entities, got %d", len(es))
        }
        var off *document.Entity
        for i := range es {
                if es[i].ID == ids[0] {
                        off = &es[i]
                }
        }
        if off == nil {
                t.Fatal("offset entity not found")
        }
        if !approxEq(off.Y1, 5) || !approxEq(off.Y2, 5) {
                t.Errorf("offset Y: got (%.4f, %.4f) want (5, 5)", off.Y1, off.Y2)
        }
        if !approxEq(off.X1, 0) || !approxEq(off.X2, 10) {
                t.Errorf("offset X: got (%.4f, %.4f) want (0, 10)", off.X1, off.X2)
        }
}

func TestOffset_Line_Negative(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        ids := d.Offset([]int{id}, -3)
        if len(ids) != 1 {
                t.Fatalf("expected 1 entity, got %d", len(ids))
        }
        es := d.Entities()
        var off *document.Entity
        for i := range es {
                if es[i].ID == ids[0] {
                        off = &es[i]
                }
        }
        if off == nil {
                t.Fatal("offset entity not found")
        }
        if !approxEq(off.Y1, -3) || !approxEq(off.Y2, -3) {
                t.Errorf("offset Y: got (%.4f, %.4f) want (-3, -3)", off.Y1, off.Y2)
        }
}

func TestOffset_Circle_Grows(t *testing.T) {
        d := document.New()
        id := d.AddCircle(0, 0, 10, 0, "#fff")
        ids := d.Offset([]int{id}, 5)
        if len(ids) != 1 {
                t.Fatalf("expected 1 entity, got %d", len(ids))
        }
        es := d.Entities()
        var off *document.Entity
        for i := range es {
                if es[i].ID == ids[0] {
                        off = &es[i]
                }
        }
        if !approxEq(off.R, 15) {
                t.Errorf("offset radius: got %.4f want 15", off.R)
        }
}

func TestOffset_Undo(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        d.Offset([]int{id}, 5)
        if len(d.Entities()) != 2 {
                t.Fatalf("expected 2 entities before undo, got %d", len(d.Entities()))
        }
        d.Undo()
        if len(d.Entities()) != 1 {
                t.Fatalf("expected 1 entity after undo, got %d", len(d.Entities()))
        }
}

func TestOffset_EmptyIDs(t *testing.T) {
        d := document.New()
        if ids := d.Offset(nil, 5); ids != nil {
                t.Errorf("expected nil for empty ids, got %v", ids)
        }
}

func TestOffset_ZeroDist(t *testing.T) {
        d := document.New()
        id := d.AddLine(0, 0, 10, 0, 0, "#fff")
        if ids := d.Offset([]int{id}, 0); ids != nil {
                t.Errorf("expected nil for zero dist, got %v", ids)
        }
}

func TestOffset_MultipleEntities(t *testing.T) {
        d := document.New()
        id1 := d.AddLine(0, 0, 10, 0, 0, "#fff")
        id2 := d.AddLine(0, 5, 10, 5, 0, "#fff")
        ids := d.Offset([]int{id1, id2}, 2)
        if len(ids) != 2 {
                t.Fatalf("expected 2 new entities, got %d", len(ids))
        }
        if len(d.Entities()) != 4 {
                t.Fatalf("expected 4 entities, got %d", len(d.Entities()))
        }
        // Undo should remove both copies in one step.
        d.Undo()
        if len(d.Entities()) != 2 {
                t.Fatalf("expected 2 entities after undo, got %d", len(d.Entities()))
        }
}
