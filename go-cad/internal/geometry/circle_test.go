package geometry

import (
	"math"
	"testing"
)

func TestCircleCircumference(t *testing.T) {
	c := Circle{Center: Point{0, 0}, Radius: 1}
	want := 2 * math.Pi
	if math.Abs(c.Circumference()-want) > Epsilon {
		t.Errorf("Circumference: got %v, want %v", c.Circumference(), want)
	}
}

func TestCircleBoundingBox(t *testing.T) {
	c := Circle{Center: Point{1, 2}, Radius: 3}
	bb := c.BoundingBox()
	if !bb.Min.Near(Point{-2, -1}) || !bb.Max.Near(Point{4, 5}) {
		t.Errorf("BoundingBox: got min=%v max=%v", bb.Min, bb.Max)
	}
}

func TestCircleClosestPoint(t *testing.T) {
	c := Circle{Center: Point{0, 0}, Radius: 5}
	// Point on right side
	cp := c.ClosestPoint(Point{10, 0})
	if !cp.Near(Point{5, 0}) {
		t.Errorf("ClosestPoint: got %v, want {5,0}", cp)
	}
	// Point inside
	cp2 := c.ClosestPoint(Point{1, 0})
	if !cp2.Near(Point{5, 0}) {
		t.Errorf("ClosestPoint inside: got %v, want {5,0}", cp2)
	}
}

func TestCircleDistToPoint(t *testing.T) {
	c := Circle{Center: Point{0, 0}, Radius: 5}
	d := c.DistToPoint(Point{8, 0})
	if math.Abs(d-3) > Epsilon {
		t.Errorf("DistToPoint: got %v, want 3", d)
	}
}

func TestCircleContains(t *testing.T) {
	c := Circle{Center: Point{0, 0}, Radius: 5}
	if !c.Contains(Point{5, 0}) {
		t.Errorf("Contains: point on circle should be true")
	}
	if c.Contains(Point{3, 0}) {
		t.Errorf("Contains: interior point should be false")
	}
}

func TestCircleOffset(t *testing.T) {
	c := Circle{Center: Point{0, 0}, Radius: 5}
	off := c.Offset(2)
	if math.Abs(off.Radius-7) > Epsilon {
		t.Errorf("Offset: got %v, want 7", off.Radius)
	}
}

func TestCircleQuadrantPoints(t *testing.T) {
	c := Circle{Center: Point{0, 0}, Radius: 1}
	qp := c.QuadrantPoints()
	want := [4]Point{{1, 0}, {0, 1}, {-1, 0}, {0, -1}}
	for i, p := range qp {
		if !p.Near(want[i]) {
			t.Errorf("QuadrantPoint[%d]: got %v, want %v", i, p, want[i])
		}
	}
}
