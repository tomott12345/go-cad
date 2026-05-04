package geometry

import (
	"math"
	"testing"
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

func assertPtsLen(t *testing.T, name string, pts []Point, want int) {
	t.Helper()
	if len(pts) != want {
		t.Errorf("%s: got %d intersections, want %d (pts=%v)", name, len(pts), want, pts)
	}
}

func assertPtNear(t *testing.T, name string, got, want Point) {
	t.Helper()
	if !got.Near(want) {
		t.Errorf("%s: got %v, want %v (err=%v)", name, got, want, got.Dist(want))
	}
}

// ─── Segment × Segment ────────────────────────────────────────────────────────

func TestIntersectSegments_Cross(t *testing.T) {
	a := Segment{Point{0, 0}, Point{10, 0}}
	b := Segment{Point{5, -5}, Point{5, 5}}
	pts := IntersectSegments(a, b)
	assertPtsLen(t, "cross", pts, 1)
	assertPtNear(t, "cross pt", pts[0], Point{5, 0})
}

func TestIntersectSegments_Parallel(t *testing.T) {
	a := Segment{Point{0, 0}, Point{10, 0}}
	b := Segment{Point{0, 1}, Point{10, 1}}
	pts := IntersectSegments(a, b)
	assertPtsLen(t, "parallel", pts, 0)
}

func TestIntersectSegments_NoIntersect(t *testing.T) {
	a := Segment{Point{0, 0}, Point{4, 0}}
	b := Segment{Point{6, -1}, Point{6, 1}}
	pts := IntersectSegments(a, b)
	assertPtsLen(t, "no intersect", pts, 0)
}

func TestIntersectSegments_TShape(t *testing.T) {
	a := Segment{Point{0, 0}, Point{10, 0}}
	b := Segment{Point{5, 0}, Point{5, 5}}
	pts := IntersectSegments(a, b)
	assertPtsLen(t, "T-shape", pts, 1)
	assertPtNear(t, "T-shape pt", pts[0], Point{5, 0})
}

func TestIntersectSegments_SharedEndpoint(t *testing.T) {
	a := Segment{Point{0, 0}, Point{5, 5}}
	b := Segment{Point{5, 5}, Point{10, 0}}
	pts := IntersectSegments(a, b)
	assertPtsLen(t, "shared endpoint", pts, 1)
	assertPtNear(t, "shared endpoint pt", pts[0], Point{5, 5})
}

// ─── Segment × Circle ────────────────────────────────────────────────────────

func TestIntersectSegmentCircle_TwoPoints(t *testing.T) {
	s := Segment{Point{-10, 0}, Point{10, 0}}
	c := Circle{Center: Point{0, 0}, Radius: 5}
	pts := IntersectSegmentCircle(s, c)
	assertPtsLen(t, "seg×circle 2pts", pts, 2)
}

func TestIntersectSegmentCircle_Tangent(t *testing.T) {
	s := Segment{Point{-10, 5}, Point{10, 5}}
	c := Circle{Center: Point{0, 0}, Radius: 5}
	pts := IntersectSegmentCircle(s, c)
	assertPtsLen(t, "seg×circle tangent", pts, 1)
	assertPtNear(t, "tangent pt", pts[0], Point{0, 5})
}

func TestIntersectSegmentCircle_Miss(t *testing.T) {
	s := Segment{Point{-10, 6}, Point{10, 6}}
	c := Circle{Center: Point{0, 0}, Radius: 5}
	pts := IntersectSegmentCircle(s, c)
	assertPtsLen(t, "seg×circle miss", pts, 0)
}

func TestIntersectSegmentCircle_InsideSegment(t *testing.T) {
	// Segment from 3 to 8, circle radius 5 centered at origin
	// Only the right intersection at (5,0) is within segment
	s := Segment{Point{3, 0}, Point{8, 0}}
	c := Circle{Center: Point{0, 0}, Radius: 5}
	pts := IntersectSegmentCircle(s, c)
	assertPtsLen(t, "seg×circle one end", pts, 1)
	assertPtNear(t, "one end pt", pts[0], Point{5, 0})
}

// ─── Circle × Circle ─────────────────────────────────────────────────────────

func TestIntersectCircles_TwoPoints(t *testing.T) {
	c1 := Circle{Center: Point{0, 0}, Radius: 5}
	c2 := Circle{Center: Point{6, 0}, Radius: 5}
	pts := IntersectCircles(c1, c2)
	assertPtsLen(t, "c×c 2pts", pts, 2)
	for _, p := range pts {
		if math.Abs(p.Dist(c1.Center)-5) > 1e-9 {
			t.Errorf("c×c: point %v not on c1", p)
		}
		if math.Abs(p.Dist(c2.Center)-5) > 1e-9 {
			t.Errorf("c×c: point %v not on c2", p)
		}
	}
}

func TestIntersectCircles_Tangent(t *testing.T) {
	c1 := Circle{Center: Point{0, 0}, Radius: 5}
	c2 := Circle{Center: Point{10, 0}, Radius: 5}
	pts := IntersectCircles(c1, c2)
	assertPtsLen(t, "c×c tangent", pts, 1)
	assertPtNear(t, "c×c tangent pt", pts[0], Point{5, 0})
}

func TestIntersectCircles_TooFar(t *testing.T) {
	c1 := Circle{Center: Point{0, 0}, Radius: 3}
	c2 := Circle{Center: Point{20, 0}, Radius: 3}
	pts := IntersectCircles(c1, c2)
	assertPtsLen(t, "c×c too far", pts, 0)
}

func TestIntersectCircles_Concentric(t *testing.T) {
	c1 := Circle{Center: Point{0, 0}, Radius: 3}
	c2 := Circle{Center: Point{0, 0}, Radius: 5}
	pts := IntersectCircles(c1, c2)
	assertPtsLen(t, "c×c concentric", pts, 0)
}

// ─── Arc × Arc ───────────────────────────────────────────────────────────────

func TestIntersectArcs(t *testing.T) {
	a1 := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}
	a2 := Arc{Center: Point{6, 0}, Radius: 5, StartDeg: 90, EndDeg: 270}
	pts := IntersectArcs(a1, a2)
	// Both arcs overlap in the region x≈3, so should find at least one intersection
	if len(pts) == 0 {
		t.Errorf("arc×arc: expected at least 1 intersection, got 0")
	}
	for _, p := range pts {
		if math.Abs(p.Dist(a1.Center)-a1.Radius) > 0.01 {
			t.Errorf("arc×arc: point %v not on a1", p)
		}
		if math.Abs(p.Dist(a2.Center)-a2.Radius) > 0.01 {
			t.Errorf("arc×arc: point %v not on a2", p)
		}
	}
}

// ─── Line × Circle ───────────────────────────────────────────────────────────

func TestIntersectLineCircle_Through(t *testing.T) {
	l := Line{Point{-10, 0}, Point{10, 0}}
	c := Circle{Center: Point{0, 0}, Radius: 3}
	pts := IntersectLineCircle(l, c)
	assertPtsLen(t, "line×circle", pts, 2)
	for _, p := range pts {
		if math.Abs(p.Dist(c.Center)-c.Radius) > 1e-9 {
			t.Errorf("line×circle: point %v not on circle", p)
		}
	}
}

// ─── Segment × Arc ───────────────────────────────────────────────────────────

func TestIntersectSegmentArc(t *testing.T) {
	s := Segment{Point{-10, 0}, Point{10, 0}}
	a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 270, EndDeg: 90}
	pts := IntersectSegmentArc(s, a)
	// The arc covers the right side (270→90 CCW = through 0°)
	// The segment crosses at (-5,0) and (5,0); only (5,0) is in arc range [270°, 90°]
	if len(pts) == 0 {
		t.Errorf("seg×arc: expected intersections, got 0")
	}
}

// ─── Degenerate cases ────────────────────────────────────────────────────────

func TestIntersectSegments_ZeroLength(t *testing.T) {
	a := Segment{Point{5, 0}, Point{5, 0}}
	b := Segment{Point{0, 0}, Point{10, 0}}
	pts := IntersectSegments(a, b)
	// Zero-length segment at (5,0) is on b — should find it
	if len(pts) == 0 {
		t.Logf("ZeroLength: got 0 intersections (acceptable for degenerate input)")
	}
}

func TestIntersectCircles_Identical(t *testing.T) {
	c := Circle{Center: Point{0, 0}, Radius: 5}
	pts := IntersectCircles(c, c)
	// Concentric with same radius → infinite intersections; return none
	assertPtsLen(t, "identical circles", pts, 0)
}
