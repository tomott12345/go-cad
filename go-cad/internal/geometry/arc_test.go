package geometry

import (
	"math"
	"testing"
)

func TestArcLength(t *testing.T) {
	// Quarter circle, radius 1 → length = π/2
	a := Arc{Center: Point{0, 0}, Radius: 1, StartDeg: 0, EndDeg: 90}
	want := math.Pi / 2
	if math.Abs(a.Length()-want) > 1e-9 {
		t.Errorf("ArcLength: got %v, want %v", a.Length(), want)
	}
}

func TestArcStartEnd(t *testing.T) {
	a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}
	if !a.StartPoint().Near(Point{5, 0}) {
		t.Errorf("StartPoint: got %v", a.StartPoint())
	}
	if !a.EndPoint().Near(Point{0, 5}) {
		t.Errorf("EndPoint: got %v", a.EndPoint())
	}
}

func TestArcBoundingBox(t *testing.T) {
	// Full circle arc 0→360
	a := Arc{Center: Point{0, 0}, Radius: 3, StartDeg: 0, EndDeg: 360}
	bb := a.BoundingBox()
	if bb.Min.X > -3+Epsilon || bb.Max.X < 3-Epsilon {
		t.Errorf("BoundingBox full circle: got min=%v max=%v", bb.Min, bb.Max)
	}
}

func TestArcBoundingBoxQuarter(t *testing.T) {
	// Quarter arc 0→90, radius 4
	a := Arc{Center: Point{0, 0}, Radius: 4, StartDeg: 0, EndDeg: 90}
	bb := a.BoundingBox()
	// Should include (4,0), (0,4) — min close to (0,0)
	if bb.Max.X < 4-Epsilon || bb.Max.Y < 4-Epsilon {
		t.Errorf("BoundingBox quarter: got min=%v max=%v", bb.Min, bb.Max)
	}
}

func TestArcClosestPoint(t *testing.T) {
	a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}
	// Point at (10, 0) — closest is (5, 0)
	cp := a.ClosestPoint(Point{10, 0})
	if !cp.Near(Point{5, 0}) {
		t.Errorf("ClosestPoint: got %v, want {5,0}", cp)
	}
}

func TestArcOffset(t *testing.T) {
	a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}
	off := a.Offset(2)
	if math.Abs(off.Radius-7) > Epsilon {
		t.Errorf("Offset: got %v, want 7", off.Radius)
	}
	if off.StartDeg != a.StartDeg || off.EndDeg != a.EndDeg {
		t.Errorf("Offset: angles changed")
	}
}

func TestArcTrimAt(t *testing.T) {
	a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}
	first, second := a.TrimAt(0.5)
	if math.Abs(first.EndDeg-45) > Epsilon {
		t.Errorf("TrimAt first.EndDeg: got %v, want 45", first.EndDeg)
	}
	if math.Abs(second.StartDeg-45) > Epsilon {
		t.Errorf("TrimAt second.StartDeg: got %v, want 45", second.StartDeg)
	}
}

func TestArcContainsAngle(t *testing.T) {
	a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 30, EndDeg: 120}
	if !a.containsAngleDeg(60) {
		t.Errorf("containsAngleDeg: 60 should be in [30,120]")
	}
	if a.containsAngleDeg(200) {
		t.Errorf("containsAngleDeg: 200 should NOT be in [30,120]")
	}
}
