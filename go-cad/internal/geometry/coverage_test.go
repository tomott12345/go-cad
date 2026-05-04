// coverage_test.go provides targeted tests for every function path not yet
// exercised by the primary test files, pushing geometry package coverage to ~100%.
package geometry

import (
        "math"
        "testing"
)

// ─── Arc constructors / helpers ───────────────────────────────────────────────

func TestNewArc(t *testing.T) {
        a := NewArc(1, 2, 5, 30, 150)
        if a.Center.X != 1 || a.Radius != 5 || a.StartDeg != 30 {
                t.Errorf("NewArc: %+v", a)
        }
}

func TestArcPointAt(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}
        p := a.PointAt(0)
        if math.Abs(p.X-5) > 1e-9 || math.Abs(p.Y) > 1e-9 {
                t.Errorf("PointAt(0): %v", p)
        }
        p2 := a.PointAt(1)
        if math.Abs(p2.Y-5) > 1e-9 || math.Abs(p2.X) > 1e-9 {
                t.Errorf("PointAt(1): %v", p2)
        }
}

func TestArcDistToPoint(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}
        // Point directly on the arc at 45°
        theta := 45 * math.Pi / 180
        onArc := Point{5 * math.Cos(theta), 5 * math.Sin(theta)}
        if d := a.DistToPoint(onArc); d > 1e-6 {
                t.Errorf("DistToPoint on arc: %v", d)
        }
}

func TestArcMidpoint(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}
        m := a.Midpoint()
        if math.Abs(m.Y-5) > 1e-9 || math.Abs(m.X) > 1e-9 {
                t.Errorf("Midpoint of semicircle: %v", m)
        }
}

func TestArcNormAngle(t *testing.T) {
        // Test normAngle wrapping
        if n := normAngle(-10); n < 0 || n >= 360 {
                t.Errorf("normAngle(-10): %v", n)
        }
        if n := normAngle(370); n < 0 || n >= 360 {
                t.Errorf("normAngle(370): %v", n)
        }
}

func TestArcClosestPoint_EndPoints(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}
        // Point behind start — should snap to start point (5,0)
        p := a.ClosestPoint(Point{-1, -1})
        if math.Abs(p.Dist(a.StartPoint())) > 0.5 {
                t.Errorf("arc closest behind start: %v want near %v", p, a.StartPoint())
        }
}

// ─── BBox helpers ─────────────────────────────────────────────────────────────

func TestBBoxWidth(t *testing.T) {
        b := BBox{Min: Point{1, 2}, Max: Point{4, 6}}
        if b.Width() != 3 {
                t.Errorf("Width: %v", b.Width())
        }
}

func TestBBoxHeight(t *testing.T) {
        b := BBox{Min: Point{1, 2}, Max: Point{4, 6}}
        if b.Height() != 4 {
                t.Errorf("Height: %v", b.Height())
        }
}

func TestBBoxUnion_Empty(t *testing.T) {
        a := EmptyBBox()
        b := BBox{Min: Point{1, 1}, Max: Point{3, 3}}
        u := a.Union(b)
        if u.Min.X != 1 || u.Max.X != 3 {
                t.Errorf("Union with empty: %+v", u)
        }
}

func TestBBoxUnion_BothEmpty(t *testing.T) {
        u := EmptyBBox().Union(EmptyBBox())
        if !u.IsEmpty() {
                t.Errorf("Union of two empties should be empty")
        }
}

// ─── Circle constructors / helpers ────────────────────────────────────────────

func TestNewCircle(t *testing.T) {
        c := NewCircle(1, 2, 3)
        if c.Center.X != 1 || c.Radius != 3 {
                t.Errorf("NewCircle: %+v", c)
        }
}

func TestCircleArea(t *testing.T) {
        c := Circle{Center: Point{0, 0}, Radius: 3}
        if math.Abs(c.Area()-math.Pi*9) > 1e-9 {
                t.Errorf("Area: %v", c.Area())
        }
}

func TestCircleContainsInterior(t *testing.T) {
        c := Circle{Center: Point{0, 0}, Radius: 5}
        if !c.ContainsInterior(Point{0, 0}) {
                t.Error("center should be interior")
        }
        if c.ContainsInterior(Point{5, 0}) {
                t.Error("boundary should not be interior")
        }
        if c.ContainsInterior(Point{10, 0}) {
                t.Error("outside should not be interior")
        }
}

func TestCirclePointAt(t *testing.T) {
        c := Circle{Center: Point{0, 0}, Radius: 5}
        p := c.PointAt(0)
        if math.Abs(p.X-5) > 1e-9 || math.Abs(p.Y) > 1e-9 {
                t.Errorf("PointAt(0): %v", p)
        }
}

func TestCircleTangentPoints(t *testing.T) {
        c := Circle{Center: Point{0, 0}, Radius: 3}
        // Point outside
        pts := c.TangentPoints(Point{5, 0})
        if len(pts) != 2 {
                t.Errorf("expected 2 tangent points, got %d", len(pts))
        }
        // Point on circle — returns 1 point
        pts2 := c.TangentPoints(Point{3, 0})
        if len(pts2) != 1 {
                t.Errorf("on-circle tangent: expected 1, got %d", len(pts2))
        }
        // Point inside — returns nil
        pts3 := c.TangentPoints(Point{0, 0})
        if pts3 != nil {
                t.Errorf("inside tangent: expected nil, got %d", len(pts3))
        }
}

func TestCircleClosestPoint_AtCenter(t *testing.T) {
        c := Circle{Center: Point{0, 0}, Radius: 3}
        // When p == center, return rightmost point
        p := c.ClosestPoint(Point{0, 0})
        if math.Abs(p.X-3) > 1e-9 || math.Abs(p.Y) > 1e-9 {
                t.Errorf("center closest: %v", p)
        }
}

// ─── Ellipse constructors / helpers ──────────────────────────────────────────

func TestNewEllipse(t *testing.T) {
        e := NewEllipse(1, 2, 5, 3, 45)
        if e.Center.X != 1 || e.A != 5 || e.Rotation != 45 {
                t.Errorf("NewEllipse: %+v", e)
        }
}

func TestEllipseDistToPoint(t *testing.T) {
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        // Point on the major axis end
        onEllipse := Point{5, 0}
        d := e.DistToPoint(onEllipse)
        if d > 1e-3 {
                t.Errorf("DistToPoint on ellipse boundary: %v", d)
        }
}

// ─── Entity: BoundingBox / ClosestPoint coverage for all entity types ─────────

func TestLineEntity_BoundingBox(t *testing.T) {
        l := LineEntity{Line{P: Point{0, 0}, Q: Point{5, 5}}}
        // An infinite line has no finite bounding box; the implementation returns
        // the empty sentinel {Min:+Inf, Max:-Inf}.
        bb := l.BoundingBox()
        if bb.Min.X <= bb.Max.X {
                t.Errorf("infinite line BBox should be empty sentinel, got %+v", bb)
        }
}

func TestRayEntity_BoundingBox(t *testing.T) {
        r := RayEntity{Ray{Origin: Point{1, 2}, Dir: Point{1, 0}}}
        bb := r.BoundingBox()
        if bb.Min.X != 1 || bb.Min.Y != 2 {
                t.Errorf("Ray BBox min: %v", bb.Min)
        }
}

func TestEllipseEntity_ClosestPoint(t *testing.T) {
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        p := e.ClosestPoint(Point{10, 0})
        if math.Abs(p.X-5) > 0.1 {
                t.Errorf("closest to rightmost: %v", p)
        }
}

func TestNURBSEntity_Length(t *testing.T) {
        sp := NURBSEntity{NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{0, 0}, {5, 5}, {10, 0}}, nil)}
        l := sp.Length()
        if l <= 0 {
                t.Errorf("NURBS length should be > 0, got %v", l)
        }
}

func TestNURBSEntity_Offset(t *testing.T) {
        sp := NURBSEntity{NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{0, 0}, {5, 5}, {10, 0}}, nil)}
        off := sp.Offset(1)
        if off == nil {
                t.Error("NURBS Offset returned nil")
        }
}

func TestBezierEntity_TrimAt(t *testing.T) {
        b := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 3}, {7, 3}, {10, 0}})}
        a, c := b.TrimAt(0.5)
        if a == nil || c == nil {
                t.Error("BezierEntity TrimAt returned nil")
        }
}

func TestNURBSEntity_TrimAt(t *testing.T) {
        sp := NURBSEntity{NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{0, 0}, {5, 5}, {10, 0}}, nil)}
        a, b := sp.TrimAt(0.5)
        if a == nil || b == nil {
                t.Error("NURBSEntity TrimAt returned nil")
        }
}

func TestEllipseEntity_Offset(t *testing.T) {
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        off := e.Offset(1)
        oe, ok := off.(EllipseEntity)
        if !ok {
                t.Fatalf("EllipseEntity Offset: got %T", off)
        }
        if oe.A != 6 || oe.B != 4 {
                t.Errorf("Ellipse offset: A=%v B=%v", oe.A, oe.B)
        }
}

func TestCircleEntity_Offset(t *testing.T) {
        c := CircleEntity{Circle{Center: Point{0, 0}, Radius: 5}}
        off := c.Offset(2)
        ce := off.(CircleEntity)
        if ce.Radius != 7 {
                t.Errorf("Circle offset radius: %v", ce.Radius)
        }
}

func TestArcEntity_Offset(t *testing.T) {
        a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}}
        off := a.Offset(1)
        ae := off.(ArcEntity)
        if ae.Radius != 6 {
                t.Errorf("Arc offset radius: %v", ae.Radius)
        }
}

func TestLineEntity_Offset(t *testing.T) {
        l := LineEntity{Line{P: Point{0, 0}, Q: Point{10, 0}}}
        off := l.Offset(3)
        le := off.(LineEntity)
        if math.Abs(le.Line.P.Y-3) > 1e-9 {
                t.Errorf("Line offset Y: %v", le.Line.P.Y)
        }
}

func TestRayEntity_Offset(t *testing.T) {
        r := RayEntity{Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}}
        off := r.Offset(3)
        re := off.(RayEntity)
        if math.Abs(re.Ray.Origin.Y-3) > 1e-9 {
                t.Errorf("Ray offset origin Y: %v", re.Ray.Origin.Y)
        }
}

// ─── Intersect dispatcher: remaining uncovered arms ───────────────────────────

func TestIntersect_EllipseEllipse(t *testing.T) {
        e1 := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        e2 := EllipseEntity{Ellipse{Center: Point{2, 0}, A: 5, B: 3, Rotation: 0}}
        pts := Intersect(e1, e2)
        if len(pts) == 0 {
                t.Error("overlapping ellipses: expected intersections")
        }
}

func TestIntersect_ArcEllipse(t *testing.T) {
        a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 5, Rotation: 0}}
        assertNoNaN(t, Intersect(a, e), "arc×ellipse")
}

func TestIntersect_NURBSSegment(t *testing.T) {
        sp := NURBSEntity{NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{0, 0}, {5, 5}, {10, 0}}, nil)}
        s := SegmentEntity{Segment{Start: Point{3, -2}, End: Point{3, 8}}}
        assertNoNaN(t, Intersect(sp, s), "nurbs×segment")
}

func TestIntersect_EllipsePolyline(t *testing.T) {
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        p := PolylineEntity{Polyline{Points: []Point{{-10, 0}, {10, 0}}}}
        pts := Intersect(e, p)
        if len(pts) < 2 {
                t.Errorf("ellipse-polyline: expected ≥2, got %d", len(pts))
        }
}

func TestIntersect_RayRay(t *testing.T) {
        r1 := RayEntity{Ray{Origin: Point{0, 0}, Dir: Point{1, 1}}}
        r2 := RayEntity{Ray{Origin: Point{5, 0}, Dir: Point{-1, 1}}}
        pts := Intersect(r1, r2)
        if len(pts) != 1 {
                t.Fatalf("ray-ray: expected 1, got %d", len(pts))
        }
        if math.Abs(pts[0].X-2.5) > 0.01 {
                t.Errorf("ray-ray intersection: %v", pts[0])
        }
}

func TestIntersect_RaySegment(t *testing.T) {
        r := RayEntity{Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}}
        s := SegmentEntity{Segment{Start: Point{5, -3}, End: Point{5, 3}}}
        pts := Intersect(r, s)
        if len(pts) != 1 || math.Abs(pts[0].X-5) > 1e-9 {
                t.Errorf("ray-segment: %v", pts)
        }
}

func TestIntersect_RayLine(t *testing.T) {
        r := RayEntity{Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}}
        l := LineEntity{Line{P: Point{5, -3}, Q: Point{5, 3}}}
        pts := Intersect(r, l)
        if len(pts) != 1 || math.Abs(pts[0].X-5) > 1e-9 {
                t.Errorf("ray-line: %v", pts)
        }
}

func TestIntersect_RayBehind(t *testing.T) {
        r := RayEntity{Ray{Origin: Point{10, 0}, Dir: Point{1, 0}}} // pointing right
        s := SegmentEntity{Segment{Start: Point{5, -3}, End: Point{5, 3}}}
        pts := Intersect(r, s)
        if len(pts) != 0 {
                t.Errorf("ray should miss segment behind origin, got %d", len(pts))
        }
}

func TestIntersect_LineCircle(t *testing.T) {
        l := LineEntity{Line{P: Point{-10, 0}, Q: Point{10, 0}}}
        c := CircleEntity{Circle{Center: Point{0, 0}, Radius: 5}}
        pts := Intersect(l, c)
        if len(pts) != 2 {
                t.Fatalf("line-circle: expected 2, got %d", len(pts))
        }
}

func TestIntersect_LinePolyline(t *testing.T) {
        l := LineEntity{Line{P: Point{-10, 5}, Q: Point{10, 5}}}
        p := PolylineEntity{Polyline{Points: []Point{{0, 0}, {0, 10}}}}
        pts := Intersect(l, p)
        if len(pts) != 1 {
                t.Fatalf("line-polyline: expected 1, got %d", len(pts))
        }
}

func TestIntersect_LineLine(t *testing.T) {
        l1 := LineEntity{Line{P: Point{0, 0}, Q: Point{10, 0}}}
        l2 := LineEntity{Line{P: Point{5, -5}, Q: Point{5, 5}}}
        pts := Intersect(l1, l2)
        if len(pts) != 1 || math.Abs(pts[0].X-5) > 1e-9 {
                t.Errorf("line-line: %v", pts)
        }
}

func TestIntersect_LineEllipse(t *testing.T) {
        l := LineEntity{Line{P: Point{-10, 0}, Q: Point{10, 0}}}
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        pts := Intersect(l, e)
        if len(pts) < 2 {
                t.Errorf("line-ellipse: expected ≥2, got %d", len(pts))
        }
}

func TestIntersect_PolylineArc(t *testing.T) {
        p := PolylineEntity{Polyline{Points: []Point{{-10, 0}, {10, 0}}}}
        a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
        pts := Intersect(p, a)
        if len(pts) != 2 {
                t.Fatalf("polyline-arc: expected 2, got %d", len(pts))
        }
}

func TestIntersect_PolylineCircle(t *testing.T) {
        p := PolylineEntity{Polyline{Points: []Point{{-10, 0}, {10, 0}}}}
        c := CircleEntity{Circle{Center: Point{0, 0}, Radius: 5}}
        pts := Intersect(p, c)
        if len(pts) != 2 {
                t.Fatalf("polyline-circle: expected 2, got %d", len(pts))
        }
}

func TestIntersect_PolylineEllipse(t *testing.T) {
        p := PolylineEntity{Polyline{Points: []Point{{-10, 0}, {10, 0}}}}
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        pts := Intersect(p, e)
        if len(pts) < 2 {
                t.Errorf("polyline-ellipse: expected ≥2, got %d", len(pts))
        }
}

func TestIntersect_NURBSPolyline(t *testing.T) {
        sp := NURBSEntity{NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{-5, 0}, {0, 5}, {5, 0}}, nil)}
        p := PolylineEntity{Polyline{Points: []Point{{-10, 2}, {10, 2}}}}
        assertNoNaN(t, Intersect(sp, p), "nurbs×polyline")
}

func TestIntersect_SegmentNURBS(t *testing.T) {
        sp := NURBSEntity{NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{0, 0}, {5, 5}, {10, 0}}, nil)}
        s := SegmentEntity{Segment{Start: Point{0, 0}, End: Point{10, 5}}}
        assertNoNaN(t, Intersect(s, sp), "segment×nurbs")
}

func TestIntersect_LineNURBS(t *testing.T) {
        sp := NURBSEntity{NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{0, 0}, {5, 5}, {10, 0}}, nil)}
        l := LineEntity{Line{P: Point{-10, 2}, Q: Point{20, 2}}}
        assertNoNaN(t, Intersect(l, sp), "line×nurbs")
}

func TestIntersect_LineBezier(t *testing.T) {
        b := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})}
        l := LineEntity{Line{P: Point{-5, 3}, Q: Point{15, 3}}}
        pts := Intersect(l, b)
        if len(pts) == 0 {
                t.Error("line-bezier: expected intersections")
        }
}

func TestIntersect_RayEllipse(t *testing.T) {
        r := RayEntity{Ray{Origin: Point{-10, 0}, Dir: Point{1, 0}}}
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        pts := Intersect(r, e)
        if len(pts) < 2 {
                t.Errorf("ray-ellipse: expected ≥2, got %d", len(pts))
        }
}

func TestIntersect_RayPolyline(t *testing.T) {
        r := RayEntity{Ray{Origin: Point{0, 5}, Dir: Point{1, 0}}}
        p := PolylineEntity{Polyline{Points: []Point{{3, 0}, {3, 10}}}}
        pts := Intersect(r, p)
        if len(pts) != 1 {
                t.Fatalf("ray-polyline: expected 1, got %d", len(pts))
        }
}

func TestIntersect_RayBezier(t *testing.T) {
        r := RayEntity{Ray{Origin: Point{-5, 0}, Dir: Point{1, 0}}}
        b := BezierEntity{NewBezierSpline([]Point{{0, -3}, {3, 3}, {7, -3}, {10, 3}})}
        assertNoNaN(t, Intersect(r, b), "ray×bezier")
}

func TestIntersect_RayNURBS(t *testing.T) {
        r := RayEntity{Ray{Origin: Point{-5, 2}, Dir: Point{1, 0}}}
        sp := NURBSEntity{NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{0, 0}, {5, 5}, {10, 0}}, nil)}
        assertNoNaN(t, Intersect(r, sp), "ray×nurbs")
}

func TestIntersect_RayArc(t *testing.T) {
        r := RayEntity{Ray{Origin: Point{-10, 0}, Dir: Point{1, 0}}}
        a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
        pts := Intersect(r, a)
        if len(pts) == 0 {
                t.Error("ray-arc: expected intersections")
        }
}

func TestIntersect_EllipseBezier(t *testing.T) {
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        b := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})}
        assertNoNaN(t, Intersect(e, b), "ellipse×bezier")
}

func TestIntersect_EllipseNURBS(t *testing.T) {
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        sp := NURBSEntity{NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{0, 0}, {5, 5}, {10, 0}}, nil)}
        assertNoNaN(t, Intersect(e, sp), "ellipse×nurbs")
}

// ─── Collinear segment edge cases ────────────────────────────────────────────

func TestIntersectCollinear_NoOverlap(t *testing.T) {
        a := Segment{Start: Point{0, 0}, End: Point{3, 0}}
        b := Segment{Start: Point{5, 0}, End: Point{8, 0}}
        pts := IntersectSegments(a, b)
        if len(pts) != 0 {
                t.Errorf("non-overlapping collinear: expected 0, got %d", len(pts))
        }
}

func TestIntersectCollinear_TouchingEnd(t *testing.T) {
        a := Segment{Start: Point{0, 0}, End: Point{5, 0}}
        b := Segment{Start: Point{5, 0}, End: Point{10, 0}}
        pts := IntersectSegments(a, b)
        if len(pts) != 1 {
                t.Errorf("touching collinear: expected 1, got %d", len(pts))
        }
}

// ─── IntersectLines parallel ──────────────────────────────────────────────────

func TestIntersectLines_Parallel(t *testing.T) {
        l1 := Line{P: Point{0, 0}, Q: Point{10, 0}}
        l2 := Line{P: Point{0, 1}, Q: Point{10, 1}}
        pts := IntersectLines(l1, l2)
        if len(pts) != 0 {
                t.Errorf("parallel lines: expected 0, got %d", len(pts))
        }
}

// ─── IntersectLineCircle tangent / miss ──────────────────────────────────────

func TestIntersectLineCircle_Miss(t *testing.T) {
        l := Line{P: Point{0, 10}, Q: Point{10, 10}}
        c := Circle{Center: Point{5, 0}, Radius: 3}
        pts := IntersectLineCircle(l, c)
        if len(pts) != 0 {
                t.Errorf("miss: expected 0, got %d", len(pts))
        }
}

func TestIntersectLineCircle_Tangent(t *testing.T) {
        l := Line{P: Point{0, 3}, Q: Point{10, 3}}
        c := Circle{Center: Point{5, 0}, Radius: 3}
        pts := IntersectLineCircle(l, c)
        if len(pts) != 1 {
                t.Fatalf("tangent: expected 1, got %d", len(pts))
        }
}

// ─── IntersectCircles edge cases ─────────────────────────────────────────────

func TestIntersectCircles_OneInside(t *testing.T) {
        c1 := Circle{Center: Point{0, 0}, Radius: 10}
        c2 := Circle{Center: Point{0, 0}, Radius: 3}
        pts := IntersectCircles(c1, c2)
        if len(pts) != 0 {
                t.Errorf("concentric: expected 0, got %d", len(pts))
        }
}

func TestIntersectCircles_TooFarCoverage(t *testing.T) {
        c1 := Circle{Center: Point{0, 0}, Radius: 1}
        c2 := Circle{Center: Point{100, 0}, Radius: 1}
        pts := IntersectCircles(c1, c2)
        if len(pts) != 0 {
                t.Errorf("too far: expected 0, got %d", len(pts))
        }
}

// ─── Segment ─────────────────────────────────────────────────────────────────

func TestSegmentContains(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        if !s.Contains(Point{5, 0}) {
                t.Error("midpoint should be contained")
        }
        if s.Contains(Point{5, 1}) {
                t.Error("off-segment point should not be contained")
        }
}

// ─── Ray helpers ─────────────────────────────────────────────────────────────

func TestRay_PointAt(t *testing.T) {
        r := Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}
        p := r.PointAt(7)
        if math.Abs(p.X-7) > 1e-9 {
                t.Errorf("PointAt(7): %v", p)
        }
}

func TestRay_DistToPoint(t *testing.T) {
        r := Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}
        d := r.DistToPoint(Point{5, 3})
        if math.Abs(d-3) > 1e-9 {
                t.Errorf("DistToPoint: %v", d)
        }
}

func TestRay_IntersectWithCircle_Miss(t *testing.T) {
        r := Ray{Origin: Point{0, 0}, Dir: Point{0, 1}} // pointing up
        c := Circle{Center: Point{100, 0}, Radius: 1}
        pts := r.IntersectWithCircle(c)
        if len(pts) != 0 {
                t.Errorf("ray-circle miss: expected 0, got %d", len(pts))
        }
}

func TestRay_IntersectWithLine_Parallel(t *testing.T) {
        r := Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}
        l := Line{P: Point{0, 1}, Q: Point{10, 1}}
        pts := r.IntersectWithLine(l)
        if len(pts) != 0 {
                t.Errorf("parallel ray-line: expected 0, got %d", len(pts))
        }
}

// ─── Spline intersections ─────────────────────────────────────────────────────

func TestIntersectNURBSNURBS(t *testing.T) {
        a := NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{0, 0}, {5, 5}, {10, 0}}, nil)
        b := NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{0, 4}, {5, -1}, {10, 4}}, nil)
        assertNoNaN(t, IntersectNURBSNURBS(a, b), "nurbs×nurbs")
}

func TestIntersectBezierBezier_Standalone(t *testing.T) {
        a := NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})
        b := NewBezierSpline([]Point{{0, 3}, {3, -2}, {7, -2}, {10, 3}})
        assertNoNaN(t, IntersectBezierBezier(a, b), "bezier×bezier")
}

// ─── IntersectArcPolyline / IntersectCirclePolyline ───────────────────────────

func TestIntersectArcPolyline_NoHit(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}
        p := Polyline{Points: []Point{{-10, -5}, {-5, -5}}}
        pts := IntersectArcPolyline(a, p)
        if len(pts) != 0 {
                t.Errorf("arc-polyline no-hit: expected 0, got %d", len(pts))
        }
}

func TestIntersectCirclePolyline_TwoHits(t *testing.T) {
        c := Circle{Center: Point{0, 0}, Radius: 5}
        p := Polyline{Points: []Point{{-10, 0}, {10, 0}}}
        pts := IntersectCirclePolyline(c, p)
        if len(pts) != 2 {
                t.Errorf("circle-polyline: expected 2, got %d", len(pts))
        }
}

// ─── MarshalEntity for all types ──────────────────────────────────────────────

func TestMarshalEntity_AllTypes(t *testing.T) {
        entities := []Entity{
                LineEntity{Line{Point{0, 0}, Point{1, 1}}},
                RayEntity{Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}},
                EllipseEntity{Ellipse{Point{0, 0}, 5, 3, 0}},
                NURBSEntity{NewNURBSSpline(2, []float64{0, 0, 0, 1, 1, 1}, []Point{{0, 0}, {5, 5}, {10, 0}}, nil)},
                BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 3}, {7, 3}, {10, 0}})},
        }
        for _, e := range entities {
                b, err := MarshalEntity(e)
                if err != nil {
                        t.Errorf("%T MarshalEntity: %v", e, err)
                        continue
                }
                got, err := UnmarshalEntity(b)
                if err != nil {
                        t.Errorf("%T UnmarshalEntity: %v", e, err)
                        continue
                }
                if got.Kind() != e.Kind() {
                        t.Errorf("%T roundtrip kind: got %v", e, got.Kind())
                }
        }
}
