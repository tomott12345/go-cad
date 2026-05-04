// final_coverage_test.go covers the remaining private function branches that
// are unreachable through the public Intersect() API due to lexicographic
// Kind ordering, plus remaining edge cases in other functions.
package geometry

import (
        "encoding/json"
        "math"
        "testing"
)

// ─── intersectSegmentWith: private function, all arms ────────────────────────
// Due to lexicographic ordering ("segment" is largest), intersectSegmentWith
// is only called via the public API for Segment×Segment. All other arms must
// be exercised by calling the private function directly.

func TestIntersectSegmentWith_Line(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        l := LineEntity{Line{P: Point{5, -5}, Q: Point{5, 5}}}
        pts := intersectSegmentWith(s, l)
        if len(pts) != 1 || math.Abs(pts[0].X-5) > 1e-9 {
                t.Errorf("seg×line: %v", pts)
        }
}

func TestIntersectSegmentWith_Ray(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        r := RayEntity{Ray{Origin: Point{5, -5}, Dir: Point{0, 1}}}
        pts := intersectSegmentWith(s, r)
        if len(pts) != 1 || math.Abs(pts[0].X-5) > 1e-9 {
                t.Errorf("seg×ray: %v", pts)
        }
}

func TestIntersectSegmentWith_Circle(t *testing.T) {
        s := Segment{Start: Point{-10, 0}, End: Point{10, 0}}
        c := CircleEntity{Circle{Center: Point{0, 0}, Radius: 3}}
        pts := intersectSegmentWith(s, c)
        if len(pts) != 2 {
                t.Errorf("seg×circle: expected 2, got %d", len(pts))
        }
}

func TestIntersectSegmentWith_Arc(t *testing.T) {
        s := Segment{Start: Point{-10, 0}, End: Point{10, 0}}
        a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 3, StartDeg: 0, EndDeg: 180}}
        pts := intersectSegmentWith(s, a)
        if len(pts) != 2 {
                t.Errorf("seg×arc: expected 2, got %d", len(pts))
        }
}

func TestIntersectSegmentWith_Ellipse(t *testing.T) {
        s := Segment{Start: Point{-10, 0}, End: Point{10, 0}}
        e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
        pts := intersectSegmentWith(s, e)
        if len(pts) != 2 {
                t.Errorf("seg×ellipse: expected 2, got %d", len(pts))
        }
}

func TestIntersectSegmentWith_Polyline(t *testing.T) {
        s := Segment{Start: Point{5, -5}, End: Point{5, 5}}
        p := PolylineEntity{Polyline{Points: []Point{{0, 0}, {10, 0}}}}
        pts := intersectSegmentWith(s, p)
        if len(pts) != 1 || math.Abs(pts[0].X-5) > 1e-9 {
                t.Errorf("seg×polyline: %v", pts)
        }
}

func TestIntersectSegmentWith_Bezier(t *testing.T) {
        s := Segment{Start: Point{5, -5}, End: Point{5, 5}}
        b := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 3}, {7, 3}, {10, 0}})}
        pts := intersectSegmentWith(s, b)
        _ = pts // just ensure no panic and arm is covered
}

func TestIntersectSegmentWith_NURBS(t *testing.T) {
        s := Segment{Start: Point{5, -5}, End: Point{5, 5}}
        sp := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )}
        pts := intersectSegmentWith(s, sp)
        _ = pts
}

func TestIntersectSegmentWith_Unknown(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        pts := intersectSegmentWith(s, nil)
        if pts != nil {
                t.Errorf("unknown entity: expected nil, got %v", pts)
        }
}

// ─── intersectPolylineWith: Line and Ray arms ─────────────────────────────────

func TestIntersectPolylineWith_Line(t *testing.T) {
        p := Polyline{Points: []Point{{0, 0}, {10, 0}}}
        l := LineEntity{Line{P: Point{5, -5}, Q: Point{5, 5}}}
        pts := intersectPolylineWith(p, l)
        if len(pts) != 1 || math.Abs(pts[0].X-5) > 1e-9 {
                t.Errorf("poly×line: %v", pts)
        }
}

func TestIntersectPolylineWith_Ray(t *testing.T) {
        p := Polyline{Points: []Point{{0, 0}, {10, 0}}}
        r := RayEntity{Ray{Origin: Point{5, -5}, Dir: Point{0, 1}}}
        pts := intersectPolylineWith(p, r)
        if len(pts) != 1 || math.Abs(pts[0].X-5) > 1e-9 {
                t.Errorf("poly×ray: %v", pts)
        }
}

// ─── intersectEllipseWith: Line and Ray arms ──────────────────────────────────

func TestIntersectEllipseWith_LineNoArm(t *testing.T) {
        // intersectEllipseWith has no Line arm — falls to nil (line × ellipse
        // is handled by intersectLineWith in the public API due to ordering).
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        l := LineEntity{Line{P: Point{-10, 0}, Q: Point{10, 0}}}
        pts := intersectEllipseWith(e, l)
        _ = pts // nil is expected — just ensure no panic and the default arm is hit
}

func TestIntersectEllipseWith_RayNoArm(t *testing.T) {
        // Same: no RayEntity arm in intersectEllipseWith.
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        r := RayEntity{Ray{Origin: Point{-10, 0}, Dir: Point{1, 0}}}
        pts := intersectEllipseWith(e, r)
        _ = pts
}

// ─── intersectCircleWith: remaining unreachable arms ─────────────────────────

func TestIntersectCircleWith_Segment(t *testing.T) {
        c := Circle{Center: Point{0, 0}, Radius: 5}
        s := SegmentEntity{Segment{Start: Point{-10, 0}, End: Point{10, 0}}}
        pts := intersectCircleWith(c, s)
        if len(pts) != 2 {
                t.Errorf("circle×seg: expected 2, got %d", len(pts))
        }
}

func TestIntersectCircleWith_Line(t *testing.T) {
        c := Circle{Center: Point{0, 0}, Radius: 5}
        l := LineEntity{Line{P: Point{-10, 0}, Q: Point{10, 0}}}
        pts := intersectCircleWith(c, l)
        if len(pts) != 2 {
                t.Errorf("circle×line: expected 2, got %d", len(pts))
        }
}

func TestIntersectCircleWith_Ray(t *testing.T) {
        c := Circle{Center: Point{0, 0}, Radius: 5}
        r := RayEntity{Ray{Origin: Point{-10, 0}, Dir: Point{1, 0}}}
        pts := intersectCircleWith(c, r)
        if len(pts) != 2 {
                t.Errorf("circle×ray: expected 2, got %d", len(pts))
        }
}

// ─── intersectArcWith: all remaining arms ─────────────────────────────────────

func TestIntersectArcWith_Segment(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}
        s := SegmentEntity{Segment{Start: Point{-10, 0}, End: Point{10, 0}}}
        pts := intersectArcWith(a, s)
        if len(pts) != 2 {
                t.Errorf("arc×seg: expected 2, got %d", len(pts))
        }
}

// ─── intersectLineWith: NURBS arm (already tested? check) ─────────────────────

func TestIntersectLineWith_NURBS(t *testing.T) {
        l := Line{P: Point{-10, 2}, Q: Point{20, 2}}
        sp := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )}
        pts := intersectLineWith(l, sp)
        _ = pts
}

// ─── MarshalEntity: default/unknown branch ────────────────────────────────────

type customEntity struct{}

func (c customEntity) Kind() Kind               { return Kind("custom") }
func (c customEntity) BoundingBox() BBox        { return EmptyBBox() }
func (c customEntity) ClosestPoint(p Point) Point { return p }
func (c customEntity) Length() float64          { return 0 }
func (c customEntity) Offset(d float64) Entity  { return c }
func (c customEntity) TrimAt(t float64) (Entity, Entity) { return c, c }

func TestMarshalEntity_Unknown(t *testing.T) {
        _, err := MarshalEntity(customEntity{})
        if err == nil {
                t.Error("expected error for unknown entity type")
        }
}

// ─── UnmarshalEntity: bad inner data for each type ────────────────────────────

func TestUnmarshalEntity_BadSegment(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{"kind":"segment","data":"bad"}`))
        if err == nil {
                t.Error("expected error for bad segment data")
        }
}

func TestUnmarshalEntity_BadLine(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{"kind":"line","data":"bad"}`))
        if err == nil {
                t.Error("expected error for bad line data")
        }
}

func TestUnmarshalEntity_BadRay(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{"kind":"ray","data":"bad"}`))
        if err == nil {
                t.Error("expected error for bad ray data")
        }
}

func TestUnmarshalEntity_BadCircle(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{"kind":"circle","data":"bad"}`))
        if err == nil {
                t.Error("expected error for bad circle data")
        }
}

func TestUnmarshalEntity_BadPolyline(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{"kind":"polyline","data":"bad"}`))
        if err == nil {
                t.Error("expected error for bad polyline data")
        }
}

func TestUnmarshalEntity_BadBezier(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{"kind":"bezier","data":"bad"}`))
        if err == nil {
                t.Error("expected error for bad bezier data")
        }
}

func TestUnmarshalEntity_BadNURBS(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{"kind":"nurbs","data":"bad"}`))
        if err == nil {
                t.Error("expected error for bad nurbs data")
        }
}

// ─── intersect.go:intersectOrdered remaining arm ─────────────────────────────

// The default (unknown) arm of intersectOrdered
func TestIntersectOrdered_Unknown(t *testing.T) {
        pts := intersectOrdered(customEntity{}, customEntity{})
        if pts != nil {
                t.Errorf("unknown pair: expected nil, got %v", pts)
        }
}

// ─── polyline.go:Offset with closed polyline ─────────────────────────────────

func TestPolyline_Offset_Closed(t *testing.T) {
        p := Polyline{
                Points: []Point{{0, 0}, {5, 0}, {5, 5}, {0, 5}},
                Closed: true,
        }
        off := p.Offset(1)
        if len(off.Points) != 4 {
                t.Errorf("closed offset: expected 4 pts, got %d", len(off.Points))
        }
}

func TestPolyline_Offset_SinglePoint(t *testing.T) {
        p := Polyline{Points: []Point{{3, 3}}}
        off := p.Offset(5)
        // < 2 points → returns self
        if len(off.Points) != 1 {
                t.Errorf("single-point offset: expected 1 pt, got %d", len(off.Points))
        }
}

// ─── polyline.go:TrimAt with zero-length polyline ────────────────────────────

func TestPolyline_TrimAt_ZeroLength(t *testing.T) {
        p := Polyline{Points: []Point{{5, 5}, {5, 5}}} // zero-length
        a, b := p.TrimAt(0.5)
        _ = a
        _ = b
}

// ─── filterBySegment: exactly one inside one outside ─────────────────────────

func TestFilterBySegment_MixedInsideOutside(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        pts := filterBySegment(s, []Point{{5, 0}, {15, 0}})
        if len(pts) != 1 || math.Abs(pts[0].X-5) > 1e-9 {
                t.Errorf("mixed: expected [5,0], got %v", pts)
        }
}

// ─── IntersectLineCircle: discriminant == 0 exactly ──────────────────────────

func TestIntersectLineCircle_TangentBottom(t *testing.T) {
        l := Line{P: Point{-10, -5}, Q: Point{10, -5}}
        c := Circle{Center: Point{0, 0}, Radius: 5}
        pts := IntersectLineCircle(l, c)
        if len(pts) != 1 {
                t.Fatalf("tangent bottom: expected 1, got %d", len(pts))
        }
        if math.Abs(pts[0].Y+5) > 1e-6 {
                t.Errorf("tangent Y: %v", pts[0].Y)
        }
}

// ─── spline.go:PointAt clamping ──────────────────────────────────────────────

func TestBezierSpline_PointAt_AtOne(t *testing.T) {
        sp := NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})
        p := sp.PointAt(1.0)
        if math.Abs(p.X-10) > 1e-9 || math.Abs(p.Y) > 1e-9 {
                t.Errorf("PointAt(1.0): %v", p)
        }
}

func TestBezierSpline_PointAt_SegBoundary(t *testing.T) {
        // Multi-segment spline: 7 control points = 2 segments
        sp := NewBezierSpline([]Point{
                {0, 0}, {2, 4}, {4, 4}, {6, 0}, {8, -4}, {10, -4}, {12, 0},
        })
        // t at exact segment boundary (0.5 for 2-segment spline)
        p := sp.PointAt(0.5)
        if math.IsNaN(p.X) || math.IsNaN(p.Y) {
                t.Errorf("PointAt boundary: NaN %v", p)
        }
}

// ─── NURBS PointAt boundary ───────────────────────────────────────────────────

func TestNURBSSpline_PointAt_BelowLo(t *testing.T) {
        sp := NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )
        // t below domain minimum should be clamped to lo
        p := sp.PointAt(-1)
        _ = p
}

// ─── intersectCollinearSegments: parallel (not collinear) branch ──────────────

func TestIntersectCollinear_Parallel(t *testing.T) {
        a := Segment{Start: Point{0, 0}, End: Point{5, 0}}
        b := Segment{Start: Point{0, 1}, End: Point{5, 1}}
        pts := intersectCollinearSegments(a, b)
        if pts != nil {
                t.Errorf("parallel non-collinear: expected nil, got %v", pts)
        }
}

// ─── intersectPolylineWith nil default ───────────────────────────────────────

func TestIntersectPolylineWith_Unknown(t *testing.T) {
        p := Polyline{Points: []Point{{0, 0}, {10, 0}}}
        pts := intersectPolylineWith(p, nil)
        if pts != nil {
                t.Errorf("unknown: expected nil, got %v", pts)
        }
}

func TestIntersectEllipseWith_Unknown(t *testing.T) {
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        pts := intersectEllipseWith(e, nil)
        if pts != nil {
                t.Errorf("unknown: expected nil, got %v", pts)
        }
}

func TestIntersectCircleWith_Unknown(t *testing.T) {
        c := Circle{Center: Point{0, 0}, Radius: 5}
        pts := intersectCircleWith(c, nil)
        if pts != nil {
                t.Errorf("unknown: expected nil, got %v", pts)
        }
}

func TestIntersectArcWith_Unknown(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}
        pts := intersectArcWith(a, nil)
        if pts != nil {
                t.Errorf("unknown: expected nil, got %v", pts)
        }
}

// ─── arc.go:ClosestPoint: zero-radius degenerate ────────────────────────────

func TestArc_ClosestPoint_AtCenter(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 45, EndDeg: 135}
        // Point at center — should return the closest arc point without panic
        p := a.ClosestPoint(Point{0, 0})
        _ = p // just no panic
}

// ─── RawEntity JSON roundtrip ─────────────────────────────────────────────────

func TestRawEntity_MarshalRoundtrip(t *testing.T) {
        re := RawEntity{EntityKind: KindCircle, Data: json.RawMessage(`{"center":{"x":0,"y":0},"radius":5}`)}
        b, err := json.Marshal(re)
        if err != nil {
                t.Fatalf("marshal: %v", err)
        }
        var re2 RawEntity
        if err := json.Unmarshal(b, &re2); err != nil {
                t.Fatalf("unmarshal: %v", err)
        }
        if re2.EntityKind != KindCircle {
                t.Errorf("kind: %v", re2.EntityKind)
        }
}
