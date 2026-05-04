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
