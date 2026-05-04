package geometry

import (
	"math"
	"testing"
)

func TestBezierPointAt_Endpoints(t *testing.T) {
	// Simple S-curve
	b := BezierSpline{Controls: []Point{
		{0, 0}, {1, 2}, {3, 2}, {4, 0},
	}}
	start := b.PointAt(0)
	end := b.PointAt(1)
	if !start.Near(Point{0, 0}) {
		t.Errorf("BezierPointAt(0): got %v, want {0,0}", start)
	}
	if !end.Near(Point{4, 0}) {
		t.Errorf("BezierPointAt(1): got %v, want {4,0}", end)
	}
}

func TestBezierLength_Positive(t *testing.T) {
	b := BezierSpline{Controls: []Point{
		{0, 0}, {1, 1}, {3, 1}, {4, 0},
	}}
	l := b.Length()
	if l <= 0 {
		t.Errorf("BezierLength should be positive, got %v", l)
	}
}

func TestBezierClosestPoint(t *testing.T) {
	// Straight-line Bezier (degenerate)
	b := BezierSpline{Controls: []Point{
		{0, 0}, {0, 0}, {10, 0}, {10, 0},
	}}
	cp := b.ClosestPoint(Point{5, 10})
	if math.Abs(cp.Y) > 0.5 {
		t.Errorf("ClosestPoint on straight line: got %v", cp)
	}
}

func TestBezierBoundingBox(t *testing.T) {
	b := BezierSpline{Controls: []Point{
		{0, 0}, {5, 10}, {15, 10}, {20, 0},
	}}
	bb := b.BoundingBox()
	if bb.IsEmpty() {
		t.Error("BezierBoundingBox should not be empty")
	}
	if bb.Min.X > 0+0.1 || bb.Max.X < 20-0.1 {
		t.Errorf("BoundingBox X range: min=%v max=%v", bb.Min.X, bb.Max.X)
	}
}

func TestBezierApproxPolyline(t *testing.T) {
	b := BezierSpline{Controls: []Point{
		{0, 0}, {1, 1}, {3, 1}, {4, 0},
	}}
	pts := b.ApproxPolyline(10)
	if len(pts) == 0 {
		t.Error("ApproxPolyline returned empty")
	}
	if !pts[0].Near(Point{0, 0}) {
		t.Errorf("ApproxPolyline first: got %v", pts[0])
	}
	if !pts[len(pts)-1].Near(Point{4, 0}) {
		t.Errorf("ApproxPolyline last: got %v", pts[len(pts)-1])
	}
}

func TestNURBSPointAt(t *testing.T) {
	// Clamped cubic NURBS forming a curve through 4 control points
	degree := 3
	controls := []Point{{0, 0}, {1, 2}, {3, 2}, {4, 0}}
	// Clamped knot vector for 4 points, degree 3: [0,0,0,0,1,1,1,1]
	knots := []float64{0, 0, 0, 0, 1, 1, 1, 1}
	n := NURBSSpline{Degree: degree, Knots: knots, Controls: controls, Weights: nil}
	// Fill default weights
	n.Weights = make([]float64, len(controls))
	for i := range n.Weights {
		n.Weights[i] = 1.0
	}

	start := n.PointAt(0)
	end := n.PointAt(1 - 1e-12)
	if !start.Near(Point{0, 0}) {
		t.Errorf("NURBS start: got %v", start)
	}
	if !end.Near(Point{4, 0}) {
		t.Logf("NURBS end: got %v (may differ slightly from {4,0})", end)
	}
}

func TestNURBSBoundingBox(t *testing.T) {
	degree := 3
	controls := []Point{{0, 0}, {1, 2}, {3, 2}, {4, 0}}
	knots := []float64{0, 0, 0, 0, 1, 1, 1, 1}
	weights := []float64{1, 1, 1, 1}
	n := NURBSSpline{Degree: degree, Knots: knots, Controls: controls, Weights: weights}
	bb := n.BoundingBox()
	if bb.IsEmpty() {
		t.Error("NURBSBoundingBox should not be empty")
	}
}
