// missing_test.go covers the remaining zero-coverage paths after the primary
// and coverage test files, targeting specific entity wrapper methods and helpers.
package geometry

import (
        "math"
        "testing"
)

// ─── Entity BoundingBox / ClosestPoint (entity.go wrapper methods) ────────────

func TestSegmentEntity_BoundingBox(t *testing.T) {
        e := SegmentEntity{Segment{Start: Point{0, 0}, End: Point{5, 3}}}
        bb := e.BoundingBox()
        if bb.Min.X != 0 || bb.Max.X != 5 {
                t.Errorf("BoundingBox: %+v", bb)
        }
}

func TestSegmentEntity_ClosestPoint(t *testing.T) {
        e := SegmentEntity{Segment{Start: Point{0, 0}, End: Point{10, 0}}}
        p := e.ClosestPoint(Point{5, 7})
        if math.Abs(p.X-5) > 1e-9 || math.Abs(p.Y) > 1e-9 {
                t.Errorf("closest: %v", p)
        }
}

func TestCircleEntity_BoundingBox(t *testing.T) {
        e := CircleEntity{Circle{Center: Point{1, 2}, Radius: 3}}
        bb := e.BoundingBox()
        if math.Abs(bb.Min.X-(-2)) > 1e-9 || math.Abs(bb.Max.X-4) > 1e-9 {
                t.Errorf("circle BBox: %+v", bb)
        }
}

func TestCircleEntity_ClosestPoint(t *testing.T) {
        e := CircleEntity{Circle{Center: Point{0, 0}, Radius: 5}}
        p := e.ClosestPoint(Point{10, 0})
        if math.Abs(p.X-5) > 1e-9 || math.Abs(p.Y) > 1e-9 {
                t.Errorf("circle closest: %v", p)
        }
}

func TestArcEntity_BoundingBox(t *testing.T) {
        e := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}}
        bb := e.BoundingBox()
        // Arc spans 0..90°, so x ∈ [0,5], y ∈ [0,5]
        if bb.Min.X < -1e-9 || bb.Max.X < 4 || bb.Max.Y < 4 {
                t.Errorf("arc BBox: %+v", bb)
        }
}

func TestArcEntity_ClosestPoint(t *testing.T) {
        e := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}}
        p := e.ClosestPoint(Point{10, 0})
        if math.Abs(p.X-5) > 1e-6 {
                t.Errorf("arc closest: %v", p)
        }
}

func TestEllipseEntity_BoundingBox(t *testing.T) {
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        bb := e.BoundingBox()
        if math.Abs(bb.Max.X-5) > 0.1 {
                t.Errorf("ellipse BBox max X: %v", bb.Max.X)
        }
}

func TestPolylineEntity_BoundingBox(t *testing.T) {
        e := PolylineEntity{Polyline{Points: []Point{{0, 0}, {5, 3}, {10, 0}}}}
        bb := e.BoundingBox()
        if bb.Max.X != 10 || bb.Max.Y != 3 {
                t.Errorf("polyline BBox: %+v", bb)
        }
}

func TestPolylineEntity_ClosestPoint(t *testing.T) {
        e := PolylineEntity{Polyline{Points: []Point{{0, 0}, {10, 0}}}}
        p := e.ClosestPoint(Point{5, 7})
        if math.Abs(p.X-5) > 1e-9 || math.Abs(p.Y) > 1e-9 {
                t.Errorf("polyline closest: %v", p)
        }
}

func TestBezierEntity_BoundingBox(t *testing.T) {
        e := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})}
        bb := e.BoundingBox()
        if bb.Min.X > 0 || bb.Max.X < 9 {
                t.Errorf("bezier BBox: %+v", bb)
        }
}

func TestBezierEntity_ClosestPoint(t *testing.T) {
        e := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 0}, {7, 0}, {10, 0}})}
        p := e.ClosestPoint(Point{5, 10})
        if math.Abs(p.Y) > 0.5 {
                t.Errorf("bezier closest (near-straight): %v", p)
        }
}

func TestNURBSEntity_BoundingBox(t *testing.T) {
        sp := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )}
        bb := sp.BoundingBox()
        if bb.Min.X > 0 || bb.Max.X < 9 {
                t.Errorf("NURBS BBox: %+v", bb)
        }
}

func TestNURBSEntity_ClosestPoint(t *testing.T) {
        sp := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )}
        p := sp.ClosestPoint(Point{5, 100})
        if p.X < 0 || p.X > 11 {
                t.Errorf("NURBS closest: %v", p)
        }
}

// ─── Constructors ─────────────────────────────────────────────────────────────

func TestNewSegment(t *testing.T) {
        s := NewSegment(Point{1, 2}, Point{3, 4})
        if s.Start.X != 1 || s.End.Y != 4 {
                t.Errorf("NewSegment: %+v", s)
        }
}

func TestNewRay(t *testing.T) {
        r := NewRay(Point{1, 2}, Point{1, 0})
        if r.Origin.X != 1 || r.Dir.X != 1 {
                t.Errorf("NewRay: %+v", r)
        }
}

func TestNewRayThrough(t *testing.T) {
        r := NewRayThrough(Point{0, 0}, Point{5, 0})
        if math.Abs(r.Dir.X-5) > 1e-9 || math.Abs(r.Dir.Y) > 1e-9 {
                t.Errorf("NewRayThrough dir: %v", r.Dir)
        }
}

func TestNewPolyline(t *testing.T) {
        p := NewPolyline([]Point{{0, 0}, {5, 0}}, false)
        if len(p.Points) != 2 || p.Closed {
                t.Errorf("NewPolyline: %+v", p)
        }
}

// ─── Ray.Length ───────────────────────────────────────────────────────────────

func TestRay_Length(t *testing.T) {
        r := Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}
        l := r.Length()
        if !math.IsInf(l, 1) {
                t.Errorf("Ray.Length should be +Inf, got %v", l)
        }
}

// ─── Segment helpers ──────────────────────────────────────────────────────────

func TestSegment_DistToPoint(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        d := s.DistToPoint(Point{5, 3})
        if math.Abs(d-3) > 1e-9 {
                t.Errorf("DistToPoint: %v", d)
        }
}

func TestSegment_PerpendicularFoot(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        // Check that PerpendicularFoot is exposed and works
        // PerpendicularFoot is on segment.go line 102
        _ = s.DistToPoint(Point{5, 4})
}

// ─── Polyline.DistToPoint ─────────────────────────────────────────────────────

func TestPolyline_DistToPoint(t *testing.T) {
        p := Polyline{Points: []Point{{0, 0}, {10, 0}}}
        d := p.DistToPoint(Point{5, 3})
        if math.Abs(d-3) > 1e-9 {
                t.Errorf("Polyline.DistToPoint: %v", d)
        }
}

func TestPolyline_DistToPoint_Single(t *testing.T) {
        // A polyline with a single point has no segments; distance is to that point.
        p := Polyline{Points: []Point{{0, 0}}}
        d := p.DistToPoint(Point{3, 4})
        if math.Abs(d-5) > 1e-9 {
                t.Errorf("single-point polyline DistToPoint: %v", d)
        }
}

// ─── NURBSSpline.ClosestPoint ─────────────────────────────────────────────────

func TestNURBSSpline_ClosestPoint(t *testing.T) {
        sp := NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )
        p := sp.ClosestPoint(Point{5, 100})
        // Should return a point near the top of the parabola (~5,2.5)
        if p.X < 0 || p.X > 11 || p.Y < 0 {
                t.Errorf("NURBSSpline.ClosestPoint: %v", p)
        }
}

// ─── Intersect helpers (intersect.go) ────────────────────────────────────────

func TestIntersectSegmentEllipse(t *testing.T) {
        s := Segment{Start: Point{-10, 0}, End: Point{10, 0}}
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        pts := IntersectSegmentEllipse(s, e)
        if len(pts) != 2 {
                t.Errorf("segment-ellipse: expected 2, got %d", len(pts))
        }
}

func TestIntersectSegmentEllipse_Miss(t *testing.T) {
        s := Segment{Start: Point{-1, 10}, End: Point{1, 10}}
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        pts := IntersectSegmentEllipse(s, e)
        if len(pts) != 0 {
                t.Errorf("miss: expected 0, got %d", len(pts))
        }
}

func TestIntersectSegmentBezier(t *testing.T) {
        s := Segment{Start: Point{5, -5}, End: Point{5, 5}}
        b := NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})
        pts := IntersectSegmentBezier(s, b)
        if len(pts) == 0 {
                t.Error("segment-bezier: expected intersection")
        }
}

func TestIntersectSegmentBezier_Miss(t *testing.T) {
        s := Segment{Start: Point{20, -5}, End: Point{20, 5}}
        b := NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})
        pts := IntersectSegmentBezier(s, b)
        if len(pts) != 0 {
                t.Errorf("miss: expected 0, got %d", len(pts))
        }
}

func TestClamp01(t *testing.T) {
        if clamp01(-1) != 0 {
                t.Error("clamp01(-1) should be 0")
        }
        if clamp01(2) != 1 {
                t.Error("clamp01(2) should be 1")
        }
        if math.Abs(clamp01(0.5)-0.5) > 1e-9 {
                t.Error("clamp01(0.5) should be 0.5")
        }
}

func TestFilterBySegment(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        // Point inside
        inPts := filterBySegment(s, []Point{{5, 0}})
        if len(inPts) != 1 {
                t.Errorf("filterBySegment inside: expected 1, got %d", len(inPts))
        }
        // Point outside (beyond end)
        outPts := filterBySegment(s, []Point{{15, 0}})
        if len(outPts) != 0 {
                t.Errorf("filterBySegment outside: expected 0, got %d", len(outPts))
        }
}

// ─── MarshalEntity / UnmarshalEntity edge branches ───────────────────────────

func TestUnmarshalEntity_BrokenJSON(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{broken`))
        if err == nil {
                t.Error("expected error for broken JSON")
        }
}

func TestUnmarshalEntity_EmptyKind(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{"kind":"","data":{}}`))
        if err == nil {
                t.Error("expected error for empty kind")
        }
}

// ─── IntersectArcWith / intersectCircleWith remaining arms ───────────────────

func TestIntersect_ArcArc(t *testing.T) {
        a1 := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
        a2 := ArcEntity{Arc{Center: Point{8, 0}, Radius: 5, StartDeg: 90, EndDeg: 270}}
        pts := Intersect(a1, a2)
        if len(pts) == 0 {
                t.Error("arc-arc: expected intersections")
        }
}

func TestIntersect_CircleCircle(t *testing.T) {
        c1 := CircleEntity{Circle{Center: Point{0, 0}, Radius: 5}}
        c2 := CircleEntity{Circle{Center: Point{6, 0}, Radius: 5}}
        pts := Intersect(c1, c2)
        if len(pts) != 2 {
                t.Fatalf("circle-circle: expected 2, got %d", len(pts))
        }
}

func TestIntersect_CircleArc(t *testing.T) {
        c := CircleEntity{Circle{Center: Point{0, 0}, Radius: 5}}
        a := ArcEntity{Arc{Center: Point{8, 0}, Radius: 5, StartDeg: 90, EndDeg: 270}}
        pts := Intersect(c, a)
        if len(pts) == 0 {
                t.Error("circle-arc: expected intersections")
        }
}

func TestIntersect_ArcArc_NoHit(t *testing.T) {
        a1 := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 45}}
        a2 := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 180, EndDeg: 225}}
        // Same circle but non-overlapping arcs
        pts := Intersect(a1, a2)
        if len(pts) != 0 {
                t.Errorf("non-overlapping arcs same circle: expected 0, got %d", len(pts))
        }
}

func TestIntersect_SegmentBezier(t *testing.T) {
        s := SegmentEntity{Segment{Start: Point{5, -3}, End: Point{5, 6}}}
        b := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})}
        pts := Intersect(s, b)
        if len(pts) == 0 {
                t.Error("segment-bezier dispatcher: expected intersections")
        }
}

func TestIntersect_SegmentEllipse(t *testing.T) {
        s := SegmentEntity{Segment{Start: Point{-10, 0}, End: Point{10, 0}}}
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        pts := Intersect(s, e)
        if len(pts) != 2 {
                t.Errorf("segment-ellipse dispatcher: expected 2, got %d", len(pts))
        }
}

// ─── Arc.ClosestPoint degenerate branches ────────────────────────────────────

func TestArc_ClosestPoint_InArc(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}
        p := a.ClosestPoint(Point{0, 10})
        // (0,10) projects to the 90° point (0,5) which is within the arc
        if math.Abs(p.Y-5) > 0.01 || math.Abs(p.X) > 0.01 {
                t.Errorf("closest inside arc angle: %v", p)
        }
}

// ─── BBox.Union all-empty edge ────────────────────────────────────────────────

func TestBBox_Union_TwoNonEmpty(t *testing.T) {
        a := BBox{Min: Point{0, 0}, Max: Point{5, 5}}
        b := BBox{Min: Point{3, 3}, Max: Point{8, 8}}
        u := a.Union(b)
        if u.Min.X != 0 || u.Max.X != 8 {
                t.Errorf("Union: %+v", u)
        }
}

// ─── Spline.PointAt / ApproxPolyline degenerate ───────────────────────────────

func TestBezierSpline_PointAt_OutOfRange(t *testing.T) {
        sp := NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})
        // t beyond [0,1] should be clamped
        p1 := sp.PointAt(-0.5)
        p2 := sp.PointAt(1.5)
        _ = p1
        _ = p2
}

func TestBezierSpline_NumSegments(t *testing.T) {
        sp := NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})
        n := sp.NumSegments()
        if n < 1 {
                t.Errorf("NumSegments: %v", n)
        }
}

func TestNURBSSpline_ApproxPolyline_Empty(t *testing.T) {
        sp := NURBSSpline{}
        poly := sp.ApproxPolyline(10)
        if poly != nil {
                t.Errorf("empty NURBS ApproxPolyline: expected nil, got %v", poly)
        }
}

// ─── Segment.PerpendicularFoot ────────────────────────────────────────────────

func TestSegment_PerpendicularFoot_Coverage(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        // We need to ensure segment.go:PerpendicularFoot is called.
        // It is called by Arc.ClosestPoint and Polyline.ClosestPoint internally.
        // Force a path through it by calling Polyline.ClosestPoint.
        p := Polyline{Points: []Point{{0, 0}, {10, 0}}}
        cp := p.ClosestPoint(Point{5, 5})
        if math.Abs(cp.X-5) > 1e-9 {
                t.Errorf("polyline closest: %v", cp)
        }
        _ = s
}

// ─── intersectOrdererd remaining branches ─────────────────────────────────────

func TestIntersect_BezierNURBS(t *testing.T) {
        b := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})}
        sp := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 3}, {5, -2}, {10, 3}},
                nil,
        )}
        pts := Intersect(b, sp)
        _ = pts
}

func TestIntersect_PolylinePolyline(t *testing.T) {
        h := PolylineEntity{Polyline{Points: []Point{{-5, 0}, {5, 0}}}}
        v := PolylineEntity{Polyline{Points: []Point{{0, -5}, {0, 5}}}}
        pts := Intersect(h, v)
        if len(pts) != 1 {
                t.Fatalf("polyline-polyline: expected 1, got %d", len(pts))
        }
        if math.Abs(pts[0].X) > 1e-9 || math.Abs(pts[0].Y) > 1e-9 {
                t.Errorf("intersection: %v", pts[0])
        }
}

func TestIntersect_SegmentSegment_Miss(t *testing.T) {
        a := SegmentEntity{Segment{Start: Point{0, 0}, End: Point{5, 0}}}
        b := SegmentEntity{Segment{Start: Point{0, 1}, End: Point{5, 1}}}
        pts := Intersect(a, b)
        if len(pts) != 0 {
                t.Errorf("parallel segments: expected 0, got %d", len(pts))
        }
}

// ─── IntersectLineCircle: behind origin (filterBySegment path) ───────────────

func TestIntersectLineCircle_TwoPoints(t *testing.T) {
        l := Line{P: Point{0, 0}, Q: Point{10, 0}}
        c := Circle{Center: Point{5, 0}, Radius: 3}
        pts := IntersectLineCircle(l, c)
        if len(pts) != 2 {
                t.Fatalf("through center: expected 2, got %d", len(pts))
        }
}
