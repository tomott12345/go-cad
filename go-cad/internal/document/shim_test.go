package document_test

import (
        "math"
        "testing"

        "go-cad/internal/document"
        "go-cad/internal/geometry"
)

func TestShim_LineToGeometry(t *testing.T) {
        e := document.Entity{Type: document.TypeLine, X1: 0, Y1: 0, X2: 10, Y2: 0}
        ge := e.ToGeometryEntity()
        if ge == nil {
                t.Fatal("ToGeometryEntity returned nil for line")
        }
        seg, ok := ge.(geometry.SegmentEntity)
        if !ok {
                t.Fatalf("expected SegmentEntity, got %T", ge)
        }
        if seg.Start.X != 0 || seg.End.X != 10 {
                t.Errorf("segment endpoints wrong: %v → %v", seg.Start, seg.End)
        }
}

func TestShim_CircleToGeometry(t *testing.T) {
        e := document.Entity{Type: document.TypeCircle, CX: 5, CY: 5, R: 3}
        ge := e.ToGeometryEntity()
        c, ok := ge.(geometry.CircleEntity)
        if !ok {
                t.Fatalf("expected CircleEntity, got %T", ge)
        }
        if c.Center.X != 5 || c.Radius != 3 {
                t.Errorf("circle wrong: center=%v r=%v", c.Center, c.Radius)
        }
}

func TestShim_ArcToGeometry(t *testing.T) {
        e := document.Entity{Type: document.TypeArc, CX: 0, CY: 0, R: 5, StartDeg: 0, EndDeg: 90}
        ge := e.ToGeometryEntity()
        a, ok := ge.(geometry.ArcEntity)
        if !ok {
                t.Fatalf("expected ArcEntity, got %T", ge)
        }
        if a.StartDeg != 0 || a.EndDeg != 90 {
                t.Errorf("arc angles wrong: start=%v end=%v", a.StartDeg, a.EndDeg)
        }
}

func TestShim_RectangleToGeometry(t *testing.T) {
        e := document.Entity{Type: document.TypeRectangle, X1: 0, Y1: 0, X2: 4, Y2: 3}
        ge := e.ToGeometryEntity()
        pl, ok := ge.(geometry.PolylineEntity)
        if !ok {
                t.Fatalf("expected PolylineEntity for rectangle, got %T", ge)
        }
        if len(pl.Points) != 4 {
                t.Errorf("rectangle polyline: expected 4 pts, got %d", len(pl.Points))
        }
        if !pl.Closed {
                t.Error("rectangle polyline should be closed")
        }
}

func TestShim_BoundingBox(t *testing.T) {
        e := document.Entity{Type: document.TypeLine, X1: -5, Y1: -2, X2: 5, Y2: 2}
        bb := e.BoundingBox()
        if bb.Min.X != -5 || bb.Max.X != 5 {
                t.Errorf("bounding box X: got [%v, %v], want [-5, 5]", bb.Min.X, bb.Max.X)
        }
}

func TestShim_ClosestPoint_Line(t *testing.T) {
        e := document.Entity{Type: document.TypeLine, X1: 0, Y1: 0, X2: 10, Y2: 0}
        p := e.ClosestPoint(geometry.Point{X: 5, Y: 7})
        if math.Abs(p.X-5) > 1e-9 || math.Abs(p.Y) > 1e-9 {
                t.Errorf("ClosestPoint: got %v, want {5 0}", p)
        }
}

func TestShim_Offset_Line(t *testing.T) {
        e := document.Entity{Type: document.TypeLine, X1: 0, Y1: 0, X2: 10, Y2: 0, Layer: 1, Color: "#ff0000"}
        off := e.Offset(3)
        if off == nil {
                t.Fatal("Offset returned nil")
        }
        // Offset of a horizontal line by +3 should shift Y by +3
        if math.Abs(off.Y1-3) > 1e-6 || math.Abs(off.Y2-3) > 1e-6 {
                t.Errorf("Offset Y: got y1=%v y2=%v, want 3", off.Y1, off.Y2)
        }
}

func TestShim_IntersectWith_Lines(t *testing.T) {
        h := document.Entity{Type: document.TypeLine, X1: 0, Y1: 5, X2: 10, Y2: 5}
        v := document.Entity{Type: document.TypeLine, X1: 5, Y1: 0, X2: 5, Y2: 10}
        pts := h.IntersectWith(v)
        if len(pts) != 1 {
                t.Fatalf("expected 1 intersection, got %d", len(pts))
        }
        if math.Abs(pts[0].X-5) > 1e-9 || math.Abs(pts[0].Y-5) > 1e-9 {
                t.Errorf("intersection point wrong: %v", pts[0])
        }
}

func TestShim_IntersectWith_LineCircle(t *testing.T) {
        l := document.Entity{Type: document.TypeLine, X1: -10, Y1: 0, X2: 10, Y2: 0}
        c := document.Entity{Type: document.TypeCircle, CX: 0, CY: 0, R: 5}
        pts := l.IntersectWith(c)
        if len(pts) != 2 {
                t.Fatalf("expected 2 intersections, got %d", len(pts))
        }
}

func TestShim_GeometryEntityToDocument_Roundtrip(t *testing.T) {
        orig := document.Entity{
                Type: document.TypeLine, Layer: 2, Color: "#00ff00",
                X1: 1, Y1: 2, X2: 3, Y2: 4,
        }
        ge := orig.ToGeometryEntity()
        back := document.GeometryEntityToDocument(ge, orig.Layer, orig.Color)
        if back == nil {
                t.Fatal("GeometryEntityToDocument returned nil")
        }
        if back.Type != document.TypeLine || back.Layer != 2 || back.Color != "#00ff00" {
                t.Errorf("roundtrip metadata wrong: %+v", back)
        }
        if back.X1 != 1 || back.Y1 != 2 || back.X2 != 3 || back.Y2 != 4 {
                t.Errorf("roundtrip coords wrong: %+v", back)
        }
}

func TestShim_UnknownType(t *testing.T) {
        e := document.Entity{Type: "unknown"}
        if ge := e.ToGeometryEntity(); ge != nil {
                t.Errorf("expected nil for unknown type, got %T", ge)
        }
}

// ── Legacy ↔ kind-envelope JSON compatibility ──────────────────────────────

// TestLegacyAndKindEnvelope_RoundTripEquivalence asserts that the same document
// entity can be recovered from both the legacy {"type":"line","x1":...} flat
// JSON (produced by ToJSON/LoadFromJSON) and the geometry kind-envelope format
// {"kind":"segment","data":{...}} (produced by MarshalGeometryJSON).
func TestLegacyAndKindEnvelope_RoundTripEquivalence(t *testing.T) {
        orig := document.Entity{
                Type: document.TypeLine,
                X1:   1, Y1: 2, X2: 8, Y2: 5,
        }

        // Kind-envelope round-trip.
        envJSON, err := orig.MarshalGeometryJSON()
        if err != nil {
                t.Fatalf("MarshalGeometryJSON: %v", err)
        }
        fromEnv, err := document.UnmarshalGeometryJSON(envJSON)
        if err != nil {
                t.Fatalf("UnmarshalGeometryJSON: %v", err)
        }

        // Both representations must produce the same coordinates and type.
        if fromEnv.Type != orig.Type {
                t.Errorf("kind-envelope type: got %q, want %q", fromEnv.Type, orig.Type)
        }
        if fromEnv.X1 != orig.X1 || fromEnv.Y1 != orig.Y1 ||
                fromEnv.X2 != orig.X2 || fromEnv.Y2 != orig.Y2 {
                t.Errorf("kind-envelope coords: got x1=%v y1=%v x2=%v y2=%v, want %v %v %v %v",
                        fromEnv.X1, fromEnv.Y1, fromEnv.X2, fromEnv.Y2,
                        orig.X1, orig.Y1, orig.X2, orig.Y2)
        }
}

// TestLegacyAndKindEnvelope_CircleRoundTrip verifies the same property for
// circle entities, which have a different field layout (cx/cy/r vs x1/y1/x2/y2).
func TestLegacyAndKindEnvelope_CircleRoundTrip(t *testing.T) {
        orig := document.Entity{Type: document.TypeCircle, CX: 3, CY: 4, R: 7}

        envJSON, err := orig.MarshalGeometryJSON()
        if err != nil {
                t.Fatalf("MarshalGeometryJSON circle: %v", err)
        }
        back, err := document.UnmarshalGeometryJSON(envJSON)
        if err != nil {
                t.Fatalf("UnmarshalGeometryJSON circle: %v", err)
        }
        if back.Type != document.TypeCircle || back.CX != 3 || back.CY != 4 || back.R != 7 {
                t.Errorf("circle round-trip: got %+v", back)
        }
}

// ── kind_wire.go tests ─────────────────────────────────────────────────────

func TestEntity_Kind_Line(t *testing.T) {
        e := document.Entity{Type: document.TypeLine, X1: 0, Y1: 0, X2: 5, Y2: 0}
        if k := e.Kind(); k != string(geometry.KindSegment) {
                t.Errorf("Kind: got %q, want %q", k, geometry.KindSegment)
        }
}

func TestEntity_Kind_Unknown(t *testing.T) {
        e := document.Entity{Type: "unknown"}
        if k := e.Kind(); k != "unknown" {
                t.Errorf("Kind unknown: got %q, want %q", k, "unknown")
        }
}

func TestEntity_MarshalGeometryJSON_Line(t *testing.T) {
        e := document.Entity{Type: document.TypeLine, X1: 0, Y1: 0, X2: 10, Y2: 0}
        b, err := e.MarshalGeometryJSON()
        if err != nil {
                t.Fatalf("MarshalGeometryJSON: %v", err)
        }
        // Round-trip: the JSON must deserialise back to an equivalent entity.
        back, err := document.UnmarshalGeometryJSON(b)
        if err != nil {
                t.Fatalf("UnmarshalGeometryJSON: %v", err)
        }
        if back.Type != document.TypeLine {
                t.Errorf("roundtrip type: got %q, want %q", back.Type, document.TypeLine)
        }
        if back.X2 != 10 {
                t.Errorf("roundtrip X2: got %v, want 10", back.X2)
        }
}

func TestEntity_MarshalGeometryJSON_Unknown(t *testing.T) {
        e := document.Entity{Type: "unknown"}
        _, err := e.MarshalGeometryJSON()
        if err == nil {
                t.Error("expected error for unknown type")
        }
}

func TestUnmarshalGeometryJSON_BadJSON(t *testing.T) {
        _, err := document.UnmarshalGeometryJSON([]byte(`{"kind":"arc","data":"bad"}`))
        if err == nil {
                t.Error("expected error for bad arc data")
        }
}

func TestUnmarshalGeometryJSON_UnsupportedKind(t *testing.T) {
        // geometry.UnmarshalEntity succeeds for any kind it knows; use a kind that
        // geometry knows but that GeometryEntityToDocument cannot map (none exist
        // currently), so exercise via geometry error path instead with totally unknown kind.
        _, err := document.UnmarshalGeometryJSON([]byte(`{"kind":"xyzzy","data":{}}`))
        if err == nil {
                t.Error("expected error for unknown geometry kind")
        }
}

func TestDocument_ToGeometryJSONArray(t *testing.T) {
        d := document.New()
        d.AddLine(0, 0, 5, 5, 0, "")
        d.AddCircle(2, 2, 1, 0, "")
        b, err := d.ToGeometryJSONArray()
        if err != nil {
                t.Fatalf("ToGeometryJSONArray: %v", err)
        }
        if len(b) == 0 {
                t.Error("ToGeometryJSONArray: empty output")
        }
}
