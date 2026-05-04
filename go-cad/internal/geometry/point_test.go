package geometry

import (
	"math"
	"testing"
)

func TestPointAdd(t *testing.T) {
	p := Point{1, 2}
	q := Point{3, 4}
	got := p.Add(q)
	if got.X != 4 || got.Y != 6 {
		t.Errorf("Add: got %v, want {4 6}", got)
	}
}

func TestPointSub(t *testing.T) {
	p := Point{5, 7}
	q := Point{2, 3}
	got := p.Sub(q)
	if got.X != 3 || got.Y != 4 {
		t.Errorf("Sub: got %v", got)
	}
}

func TestPointDot(t *testing.T) {
	p := Point{1, 0}
	q := Point{0, 1}
	if p.Dot(q) != 0 {
		t.Errorf("Dot of perpendicular unit vectors should be 0")
	}
	r := Point{3, 4}
	s := Point{4, 3}
	if r.Dot(s) != 24 {
		t.Errorf("Dot: got %v, want 24", r.Dot(s))
	}
}

func TestPointCross(t *testing.T) {
	p := Point{1, 0}
	q := Point{0, 1}
	if p.Cross(q) != 1 {
		t.Errorf("Cross of (1,0)×(0,1) should be 1, got %v", p.Cross(q))
	}
}

func TestPointLen(t *testing.T) {
	p := Point{3, 4}
	if math.Abs(p.Len()-5) > Epsilon {
		t.Errorf("Len: got %v, want 5", p.Len())
	}
}

func TestPointNormalize(t *testing.T) {
	p := Point{3, 4}
	n := p.Normalize()
	if math.Abs(n.Len()-1) > Epsilon {
		t.Errorf("Normalize: length should be 1, got %v", n.Len())
	}
	// zero vector
	z := Point{}.Normalize()
	if z.X != 0 || z.Y != 0 {
		t.Errorf("Normalize of zero: got %v", z)
	}
}

func TestPointPerp(t *testing.T) {
	p := Point{1, 0}
	perp := p.Perp()
	if p.Dot(perp) != 0 {
		t.Errorf("Perp should be orthogonal")
	}
}

func TestPointDist(t *testing.T) {
	p := Point{0, 0}
	q := Point{3, 4}
	if math.Abs(p.Dist(q)-5) > Epsilon {
		t.Errorf("Dist: got %v, want 5", p.Dist(q))
	}
}

func TestPointLerp(t *testing.T) {
	p := Point{0, 0}
	q := Point{10, 10}
	mid := p.Lerp(q, 0.5)
	if !mid.Near(Point{5, 5}) {
		t.Errorf("Lerp: got %v, want {5,5}", mid)
	}
}

func TestPointRotate(t *testing.T) {
	p := Point{1, 0}
	rotated := p.Rotate(math.Pi / 2)
	if !rotated.Near(Point{0, 1}) {
		t.Errorf("Rotate 90°: got %v, want {0,1}", rotated)
	}
}

func TestPointRotateAround(t *testing.T) {
	p := Point{2, 0}
	pivot := Point{1, 0}
	rotated := p.RotateAround(pivot, math.Pi)
	if !rotated.Near(Point{0, 0}) {
		t.Errorf("RotateAround: got %v, want {0,0}", rotated)
	}
}

func TestPointNear(t *testing.T) {
	p := Point{1, 1}
	q := Point{1 + Epsilon/2, 1}
	if !p.Near(q) {
		t.Errorf("Near: expected true for points within Epsilon")
	}
	r := Point{2, 1}
	if p.Near(r) {
		t.Errorf("Near: expected false for distant points")
	}
}
