package geometry

import (
        "math"
        "testing"
)

func TestArc_ClosestPoint_ClampToStart(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 80, EndDeg: 100}
        p := a.ClosestPoint(Point{10, 0})
        startPt := a.StartPoint()
        if p.Dist(startPt) > 0.5 {
                t.Errorf("clamp start: %v want near %v", p, startPt)
        }
}

func TestArc_ClosestPoint_ClampToEnd(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 20}
        p := a.ClosestPoint(Point{0, 10})
        endPt := a.EndPoint()
        if p.Dist(endPt) > 0.5 {
                t.Errorf("clamp end: %v want near %v", p, endPt)
        }
}

func TestArc_ClosestPoint_CenterPoint(t *testing.T) {
        // Point exactly at center, angle 0° which is within [0,180] — hits l < Epsilon branch.
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}
        p := a.ClosestPoint(Point{0, 0})
        if p.Dist(a.StartPoint()) > 1e-6 {
                t.Errorf("center query: %v want %v", p, a.StartPoint())
        }
}

func TestBBox_Union_NonEmptyNonEmpty(t *testing.T) {
        a := BBox{Min: Point{-5, -5}, Max: Point{1, 1}}
        b := BBox{Min: Point{0, 0}, Max: Point{5, 5}}
        u := a.Union(b)
        if u.Min.X != -5 || u.Max.X != 5 {
                t.Errorf("Union: %+v", u)
        }
}

func TestBBox_Union_OtherEmpty(t *testing.T) {
        a := BBox{Min: Point{1, 2}, Max: Point{3, 4}}
        u := a.Union(EmptyBBox())
        if u.Min.X != 1 || u.Max.X != 3 {
                t.Errorf("Union with empty: %+v", u)
        }
}

func TestEllipseEntity_TrimAt_NearZero(t *testing.T) {
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        a, b := e.TrimAt(0.001)
        pa := a.(PolylineEntity)
        pb := b.(PolylineEntity)
        if len(pa.Points) < 2 || len(pb.Points) < 2 {
                t.Error("TrimAt near-zero: parts too short")
        }
}

func TestEllipseEntity_TrimAt_NearOne(t *testing.T) {
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        a, b := e.TrimAt(0.999)
        pa := a.(PolylineEntity)
        pb := b.(PolylineEntity)
        if len(pa.Points) < 2 || len(pb.Points) < 2 {
                t.Error("TrimAt near-one: parts too short")
        }
}

func TestMarshalEntity_NURBS(t *testing.T) {
        e := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )}
        b, err := MarshalEntity(e)
        if err != nil {
                t.Fatalf("marshal: %v", err)
        }
        got, err := UnmarshalEntity(b)
        if err != nil {
                t.Fatalf("unmarshal: %v", err)
        }
        if got.Kind() != KindNURBSSpline {
                t.Errorf("kind: %v", got.Kind())
        }
}

func TestUnmarshalEntity_Bezier(t *testing.T) {
        e := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})}
        b, err := MarshalEntity(e)
        if err != nil {
                t.Fatalf("marshal: %v", err)
        }
        got, err := UnmarshalEntity(b)
        if err != nil {
                t.Fatalf("unmarshal: %v", err)
        }
        if got.Kind() != KindBezierSpline {
                t.Errorf("kind: %v", got.Kind())
        }
}

func TestUnmarshalEntity_Ellipse(t *testing.T) {
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 30}}
        b, err := MarshalEntity(e)
        if err != nil {
                t.Fatalf("marshal: %v", err)
        }
        got, err := UnmarshalEntity(b)
        if err != nil {
                t.Fatalf("unmarshal: %v", err)
        }
        ee, ok := got.(EllipseEntity)
        if !ok {
                t.Fatalf("expected EllipseEntity, got %T", got)
        }
        if math.Abs(ee.Rotation-30) > 1e-9 {
                t.Errorf("rotation: %v", ee.Rotation)
        }
}

func TestUnmarshalEntity_BadData(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{"kind":"arc","data":"not-an-object"}`))
        if err == nil {
                t.Error("expected error for bad arc data")
        }
}

func TestIntersect_ArcSegment(t *testing.T) {
        a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
        s := SegmentEntity{Segment{Start: Point{-10, 0}, End: Point{10, 0}}}
        pts := Intersect(a, s)
        if len(pts) != 2 {
                t.Fatalf("arc-segment: expected 2, got %d", len(pts))
        }
}

func TestIntersect_ArcLine(t *testing.T) {
        a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
        l := LineEntity{Line{P: Point{-10, 0}, Q: Point{10, 0}}}
        pts := Intersect(a, l)
        if len(pts) != 2 {
                t.Fatalf("arc-line: expected 2, got %d", len(pts))
        }
}

func TestIntersect_ArcRay(t *testing.T) {
        a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}}
        r := RayEntity{Ray{Origin: Point{-10, 3}, Dir: Point{1, 0}}}
        pts := Intersect(a, r)
        _ = pts
}

func TestIntersect_ArcCircle(t *testing.T) {
        a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
        c := CircleEntity{Circle{Center: Point{6, 0}, Radius: 5}}
        pts := Intersect(a, c)
        _ = pts
}

func TestIntersect_ArcBezier(t *testing.T) {
        a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
        b := BezierEntity{NewBezierSpline([]Point{{-3, 3}, {0, 7}, {3, 3}, {6, -1}})}
        pts := Intersect(a, b)
        _ = pts
}

func TestIntersect_ArcNURBS(t *testing.T) {
        a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
        sp := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{-5, 3}, {0, 7}, {5, 3}},
                nil,
        )}
        pts := Intersect(a, sp)
        _ = pts
}

func TestIntersect_CircleBezier(t *testing.T) {
        c := CircleEntity{Circle{Center: Point{5, 2.5}, Radius: 3}}
        b := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})}
        pts := Intersect(c, b)
        _ = pts
}

func TestIntersect_CircleNURBS(t *testing.T) {
        c := CircleEntity{Circle{Center: Point{5, 2}, Radius: 3}}
        sp := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )}
        pts := Intersect(c, sp)
        _ = pts
}


func TestIntersect_PolylineNURBS(t *testing.T) {
        p := PolylineEntity{Polyline{Points: []Point{{-10, 2}, {10, 2}}}}
        sp := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{-5, 0}, {0, 5}, {5, 0}},
                nil,
        )}
        pts := Intersect(p, sp)
        _ = pts
}

func TestIntersect_PolylineBezier(t *testing.T) {
        p := PolylineEntity{Polyline{Points: []Point{{-10, 2}, {10, 2}}}}
        b := BezierEntity{NewBezierSpline([]Point{{-5, 0}, {-2, 5}, {2, 5}, {5, 0}})}
        pts := Intersect(p, b)
        if len(pts) == 0 {
                t.Error("polyline-bezier: expected intersection")
        }
}

func TestIntersectCollinear_Overlap(t *testing.T) {
        a := Segment{Start: Point{0, 0}, End: Point{7, 0}}
        b := Segment{Start: Point{3, 0}, End: Point{10, 0}}
        pts := IntersectSegments(a, b)
        if len(pts) != 2 {
                t.Errorf("overlap: expected 2, got %d", len(pts))
        }
}

func TestIntersectCollinear_FullContainment(t *testing.T) {
        a := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        b := Segment{Start: Point{2, 0}, End: Point{8, 0}}
        pts := IntersectSegments(a, b)
        if len(pts) != 2 {
                t.Errorf("containment: expected 2, got %d", len(pts))
        }
}


func TestFilterBySegment_AtEndpoints(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        pts := filterBySegment(s, []Point{{0, 0}, {10, 0}})
        if len(pts) != 2 {
                t.Errorf("at endpoints: expected 2, got %d", len(pts))
        }
}

func TestRay_TrimAt_Negative(t *testing.T) {
        r := Ray{Origin: Point{5, 0}, Dir: Point{1, 0}}
        seg, ray := r.TrimAt(-3)
        if math.Abs(seg.Start.Dist(seg.End)) > 1e-9 {
                t.Errorf("negative TrimAt: non-zero segment")
        }
        if ray.Origin.Dist(r.Origin) > 1e-9 {
                t.Errorf("negative TrimAt: ray origin shifted")
        }
}

func TestSegment_ClosestPoint_ZeroLength(t *testing.T) {
        s := Segment{Start: Point{5, 3}, End: Point{5, 3}}
        cp, _ := s.ClosestPoint(Point{10, 7})
        if cp.X != 5 || cp.Y != 3 {
                t.Errorf("zero-length: %v", cp)
        }
}

func TestLine_ClosestPoint_ZeroLength(t *testing.T) {
        l := Line{P: Point{3, 3}, Q: Point{3, 3}}
        cp := l.ClosestPoint(Point{10, 7})
        if cp.X != 3 || cp.Y != 3 {
                t.Errorf("zero-length line: %v", cp)
        }
}

func TestLine_PerpendicularFoot(t *testing.T) {
        l := Line{P: Point{0, 0}, Q: Point{10, 0}}
        foot := l.PerpendicularFoot(Point{5, 7})
        if math.Abs(foot.X-5) > 1e-9 || math.Abs(foot.Y) > 1e-9 {
                t.Errorf("PerpendicularFoot: %v", foot)
        }
}

func TestBezierSpline_NumSegments_TooFew(t *testing.T) {
        sp := NewBezierSpline([]Point{{0, 0}, {1, 1}})
        if sp.NumSegments() != 0 {
                t.Errorf("too few ctrl: expected 0, got %d", sp.NumSegments())
        }
}

func TestBezierSpline_PointAt_NoControls(t *testing.T) {
        sp := BezierSpline{}
        p := sp.PointAt(0.5)
        _ = p
}

func TestBezierSpline_PointAt_TooFew(t *testing.T) {
        sp := NewBezierSpline([]Point{{3, 4}})
        p := sp.PointAt(0.5)
        if p.X != 3 || p.Y != 4 {
                t.Errorf("single-point bezier PointAt: %v", p)
        }
}

func TestBezierSpline_ApproxPolyline_ZeroSeg(t *testing.T) {
        sp := NewBezierSpline([]Point{{1, 2}, {3, 4}})
        pts := sp.ApproxPolyline(10)
        if len(pts) != 2 {
                t.Errorf("degenerate ApproxPolyline: %v", pts)
        }
}

func TestNURBSSpline_PointAt_AtHi(t *testing.T) {
        sp := NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )
        p := sp.PointAt(1.0)
        _ = p
}

func TestPolyline_Offset_ZeroLenSeg(t *testing.T) {
        p := Polyline{Points: []Point{{0, 0}, {0, 0}, {5, 0}}}
        off := p.Offset(1)
        _ = off
}

func TestPolyline_TrimAt_NearEnd(t *testing.T) {
        p := Polyline{Points: []Point{{0, 0}, {5, 0}, {10, 0}}}
        a, b := p.TrimAt(0.99)
        if len(a.Points) < 2 || len(b.Points) < 2 {
                t.Error("near-end trim: parts too short")
        }
}

func TestPolyline_TrimAt_NearStart(t *testing.T) {
        p := Polyline{Points: []Point{{0, 0}, {5, 0}, {10, 0}}}
        a, b := p.TrimAt(0.01)
        _ = a
        _ = b
}


func TestIntersect_NURBSNURBS(t *testing.T) {
        sp1 := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )}
        sp2 := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 3}, {5, -2}, {10, 3}},
                nil,
        )}
        pts := Intersect(sp1, sp2)
        _ = pts
}

func TestFilterBySegment_BothOutside(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{5, 0}}
        pts := filterBySegment(s, []Point{{-2, 0}, {8, 0}})
        if len(pts) != 0 {
                t.Errorf("both outside: expected 0, got %d", len(pts))
        }
}

func TestIntersectArcs_NoOverlap(t *testing.T) {
        a1 := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 45}
        a2 := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 180, EndDeg: 225}
        pts := IntersectArcs(a1, a2)
        if len(pts) != 0 {
                t.Errorf("same-circle no overlap: expected 0, got %d", len(pts))
        }
}

func TestIntersectCircleArc_NoHit(t *testing.T) {
        c := Circle{Center: Point{100, 0}, Radius: 1}
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 90}
        pts := IntersectCircleArc(c, a)
        if len(pts) != 0 {
                t.Errorf("circle-arc no hit: expected 0, got %d", len(pts))
        }
}

func TestNURBSSpline_BasisFunc_ZeroSpan(t *testing.T) {
        sp := NewNURBSSpline(2,
                []float64{0, 0, 0, 0.5, 0.5, 1, 1, 1},
                []Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}, {5, -3}},
                nil,
        )
        p := sp.PointAt(0.5)
        if math.IsNaN(p.X) || math.IsNaN(p.Y) {
                t.Errorf("NURBS repeated knots: NaN %v", p)
        }
}

func TestIntersectSegmentPolyline_Direct(t *testing.T) {
        s := Segment{Start: Point{0, -5}, End: Point{0, 5}}
        p := Polyline{Points: []Point{{-5, 0}, {5, 0}}}
        pts := IntersectSegmentPolyline(s, p)
        if len(pts) != 1 {
                t.Fatalf("segment-polyline: expected 1, got %d", len(pts))
        }
}
