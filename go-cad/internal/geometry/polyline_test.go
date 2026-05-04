package geometry

import (
	"math"
	"testing"
)

func TestPolylineLength(t *testing.T) {
	pts := []Point{{0, 0}, {3, 0}, {3, 4}}
	p := Polyline{Points: pts}
	want := 3.0 + 4.0
	if math.Abs(p.Length()-want) > Epsilon {
		t.Errorf("Length: got %v, want %v", p.Length(), want)
	}
}

func TestPolylineNumSegments_Open(t *testing.T) {
	p := Polyline{Points: []Point{{0, 0}, {1, 0}, {2, 0}}}
	if p.NumSegments() != 2 {
		t.Errorf("NumSegments open: got %d, want 2", p.NumSegments())
	}
}

func TestPolylineNumSegments_Closed(t *testing.T) {
	p := Polyline{Points: []Point{{0, 0}, {1, 0}, {2, 0}}, Closed: true}
	if p.NumSegments() != 3 {
		t.Errorf("NumSegments closed: got %d, want 3", p.NumSegments())
	}
}

func TestPolylineBoundingBox(t *testing.T) {
	p := Polyline{Points: []Point{{-1, -2}, {3, 0}, {1, 4}}}
	bb := p.BoundingBox()
	if !bb.Min.Near(Point{-1, -2}) || !bb.Max.Near(Point{3, 4}) {
		t.Errorf("BoundingBox: min=%v max=%v", bb.Min, bb.Max)
	}
}

func TestPolylineClosestPoint(t *testing.T) {
	p := Polyline{Points: []Point{{0, 0}, {10, 0}}}
	cp := p.ClosestPoint(Point{5, 5})
	if !cp.Near(Point{5, 0}) {
		t.Errorf("ClosestPoint: got %v, want {5,0}", cp)
	}
}

func TestPolylineTrimAt(t *testing.T) {
	p := Polyline{Points: []Point{{0, 0}, {10, 0}, {10, 10}}}
	first, second := p.TrimAt(0.5)
	// total length = 10+10=20, midpoint at t=0.5 → length 10 → at {10,0}
	if len(first.Points) < 2 || len(second.Points) < 2 {
		t.Errorf("TrimAt: empty result")
	}
}

func TestPolylineOffset(t *testing.T) {
	p := Polyline{Points: []Point{{0, 0}, {10, 0}}}
	off := p.Offset(3)
	if len(off.Points) != 2 {
		t.Errorf("Offset: expected 2 points, got %d", len(off.Points))
	}
	// All Y coords should be 3
	for _, pt := range off.Points {
		if math.Abs(pt.Y-3) > Epsilon {
			t.Errorf("Offset: got Y=%v, want 3", pt.Y)
		}
	}
}
