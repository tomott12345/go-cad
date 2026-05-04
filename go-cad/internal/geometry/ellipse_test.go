package geometry

import (
	"math"
	"testing"
)

func TestEllipsePointAt_Axes(t *testing.T) {
	e := Ellipse{Center: Point{0, 0}, A: 6, B: 3, Rotation: 0}
	// θ=0 should be (6,0), θ=π/2 should be (0,3)
	p0 := e.PointAt(0)
	if !p0.Near(Point{6, 0}) {
		t.Errorf("PointAt(0): got %v, want {6,0}", p0)
	}
	p90 := e.PointAt(math.Pi / 2)
	if !p90.Near(Point{0, 3}) {
		t.Errorf("PointAt(π/2): got %v, want {0,3}", p90)
	}
}

func TestEllipseBoundingBox(t *testing.T) {
	e := Ellipse{Center: Point{0, 0}, A: 6, B: 3, Rotation: 0}
	bb := e.BoundingBox()
	if bb.Max.X < 6-0.1 || bb.Min.X > -6+0.1 {
		t.Errorf("BoundingBox X: min=%v max=%v", bb.Min.X, bb.Max.X)
	}
	if bb.Max.Y < 3-0.1 || bb.Min.Y > -3+0.1 {
		t.Errorf("BoundingBox Y: min=%v max=%v", bb.Min.Y, bb.Max.Y)
	}
}

func TestEllipseCircumference(t *testing.T) {
	// Circle as degenerate ellipse: A=B=r, circumference ≈ 2πr
	e := Ellipse{Center: Point{0, 0}, A: 5, B: 5, Rotation: 0}
	want := 2 * math.Pi * 5
	if math.Abs(e.Circumference()-want) > 0.01 {
		t.Errorf("Circumference: got %v, want %v", e.Circumference(), want)
	}
}

func TestEllipseClosestPoint(t *testing.T) {
	e := Ellipse{Center: Point{0, 0}, A: 6, B: 3, Rotation: 0}
	// Point on right — closest should be at (6,0)
	cp := e.ClosestPoint(Point{100, 0})
	if math.Abs(cp.X-6) > 0.5 || math.Abs(cp.Y) > 0.5 {
		t.Errorf("ClosestPoint right: got %v, want ~{6,0}", cp)
	}
}

func TestEllipseOffset(t *testing.T) {
	e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
	off := e.Offset(2)
	if math.Abs(off.A-7) > Epsilon || math.Abs(off.B-5) > Epsilon {
		t.Errorf("Offset: got A=%v B=%v, want A=7 B=5", off.A, off.B)
	}
}

func TestEllipseApproxPolyline(t *testing.T) {
	e := Ellipse{Center: Point{0, 0}, A: 4, B: 2, Rotation: 0}
	pts := e.ApproxPolyline(36)
	if len(pts) != 37 {
		t.Errorf("ApproxPolyline: got %d points, want 37", len(pts))
	}
	// First and last points should be the same (closed)
	if !pts[0].Near(pts[len(pts)-1]) {
		t.Errorf("ApproxPolyline: first and last should be equal for closed ellipse")
	}
}
