package geometry

import (
        "encoding/json"
        "math"
        "testing"
)

// intersectSegmentWith is called in the public API only for Segment×Segment
// (because "segment" is lexicographically largest). All other arms are exercised
// via direct private-function calls here.

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
        assertNoNaN(t, intersectSegmentWith(s, b), "seg×bezier")
}

func TestIntersectSegmentWith_NURBS(t *testing.T) {
        s := Segment{Start: Point{5, -5}, End: Point{5, 5}}
        sp := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )}
        assertNoNaN(t, intersectSegmentWith(s, sp), "seg×nurbs")
}

func TestIntersectSegmentWith_Unknown(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        pts := intersectSegmentWith(s, nil)
        if pts != nil {
                t.Errorf("unknown entity: expected nil, got %v", pts)
        }
}

// intersectPolylineWith: Line and Ray arms are unreachable via public API because
// "line" and "ray" are lexicographically smaller than "polyline", so
// intersectOrdered dispatches to intersectLineWith / intersectRayWith first.

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

// intersectEllipseWith: Arc, Circle, Bezier arms are unreachable via public API
// because "arc", "bezier", "circle" are lexicographically smaller than "ellipse".

func TestIntersectEllipseWith_Segment(t *testing.T) {
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        s := SegmentEntity{Segment{Start: Point{-10, 0}, End: Point{10, 0}}}
        pts := intersectEllipseWith(e, s)
        if len(pts) != 2 {
                t.Errorf("ellipse×segment direct: expected 2, got %d", len(pts))
        }
}

func TestIntersectEllipseWith_Circle(t *testing.T) {
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        c := CircleEntity{Circle{Center: Point{4, 0}, Radius: 3}}
        assertNoNaN(t, intersectEllipseWith(e, c), "ellipse×circle")
}

func TestIntersectEllipseWith_Arc(t *testing.T) {
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        a := ArcEntity{Arc{Center: Point{4, 0}, Radius: 3, StartDeg: 0, EndDeg: 180}}
        assertNoNaN(t, intersectEllipseWith(e, a), "ellipse×arc")
}

func TestIntersectEllipseWith_Bezier(t *testing.T) {
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        b := BezierEntity{NewBezierSpline([]Point{{-6, 0}, {-2, 5}, {2, 5}, {6, 0}})}
        assertNoNaN(t, intersectEllipseWith(e, b), "ellipse×bezier")
}

func TestIntersectEllipseWith_Polyline(t *testing.T) {
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        p := PolylineEntity{Polyline{Points: []Point{{-10, 0}, {10, 0}}}}
        pts := intersectEllipseWith(e, p)
        if len(pts) != 2 {
                t.Errorf("ellipse×polyline direct: expected 2, got %d", len(pts))
        }
}

func TestIntersectEllipseWith_NURBS(t *testing.T) {
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        sp := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{-6, 0}, {0, 5}, {6, 0}},
                nil,
        )}
        assertNoNaN(t, intersectEllipseWith(e, sp), "ellipse×nurbs")
}

func TestIntersectEllipseWith_Unknown(t *testing.T) {
        e := Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}
        pts := intersectEllipseWith(e, nil)
        if pts != nil {
                t.Errorf("unknown: expected nil, got %v", pts)
        }
}

// intersectCircleWith: Segment, Line, Ray arms are exercised directly.

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

func TestIntersectCircleWith_Unknown(t *testing.T) {
        c := Circle{Center: Point{0, 0}, Radius: 5}
        pts := intersectCircleWith(c, nil)
        if pts != nil {
                t.Errorf("unknown: expected nil, got %v", pts)
        }
}

// intersectArcWith: Segment arm is unreachable via public API ("arc" < "segment").

func TestIntersectArcWith_Segment(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}
        s := SegmentEntity{Segment{Start: Point{-10, 0}, End: Point{10, 0}}}
        pts := intersectArcWith(a, s)
        if len(pts) != 2 {
                t.Errorf("arc×seg: expected 2, got %d", len(pts))
        }
}

func TestIntersectArcWith_Unknown(t *testing.T) {
        a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}
        pts := intersectArcWith(a, nil)
        if pts != nil {
                t.Errorf("unknown: expected nil, got %v", pts)
        }
}

// intersectLineWith: Bezier arm is unreachable via public API ("bezier" < "line").

func TestIntersectLineWith_Ray(t *testing.T) {
        l := Line{P: Point{-10, 2}, Q: Point{10, 2}}
        r := RayEntity{Ray{Origin: Point{0, -5}, Dir: Point{0, 1}}}
        pts := intersectLineWith(l, r)
        if len(pts) != 1 || math.Abs(pts[0].Y-2) > 1e-9 {
                t.Errorf("line×ray direct: %v", pts)
        }
}

func TestIntersectLineWith_Bezier(t *testing.T) {
        l := Line{P: Point{-10, 2}, Q: Point{10, 2}}
        b := BezierEntity{NewBezierSpline([]Point{{-5, 0}, {-2, 5}, {2, 5}, {5, 0}})}
        assertNoNaN(t, intersectLineWith(l, b), "line×bezier")
}

func TestIntersectLineWith_NURBS(t *testing.T) {
        l := Line{P: Point{-10, 2}, Q: Point{20, 2}}
        sp := NURBSEntity{NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )}
        assertNoNaN(t, intersectLineWith(l, sp), "line×nurbs")
}

// intersectRayWith: Bezier arm is unreachable via public API ("bezier" < "ray").

func TestIntersectRayWith_Bezier(t *testing.T) {
        r := Ray{Origin: Point{-10, 2}, Dir: Point{1, 0}}
        b := BezierEntity{NewBezierSpline([]Point{{-5, 0}, {-2, 5}, {2, 5}, {5, 0}})}
        assertNoNaN(t, intersectRayWith(r, b), "ray×bezier")
}

// intersectPolylineWith: nil default arm.

func TestIntersectPolylineWith_Unknown(t *testing.T) {
        p := Polyline{Points: []Point{{0, 0}, {10, 0}}}
        pts := intersectPolylineWith(p, nil)
        if pts != nil {
                t.Errorf("unknown: expected nil, got %v", pts)
        }
}

// custom entity type for testing default branches.

type customEntity struct{}

func (c customEntity) Kind() Kind                          { return Kind("custom") }
func (c customEntity) BoundingBox() BBox                  { return EmptyBBox() }
func (c customEntity) ClosestPoint(p Point) Point         { return p }
func (c customEntity) Length() float64                    { return 0 }
func (c customEntity) Offset(d float64) Entity            { return c }
func (c customEntity) TrimAt(t float64) (Entity, Entity)  { return c, c }

func TestMarshalEntity_Unknown(t *testing.T) {
        _, err := MarshalEntity(customEntity{})
        if err == nil {
                t.Error("expected error for unknown entity type")
        }
}

func TestIntersectOrdered_Unknown(t *testing.T) {
        pts := intersectOrdered(customEntity{}, customEntity{})
        if pts != nil {
                t.Errorf("unknown pair: expected nil, got %v", pts)
        }
}

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
        if len(off.Points) != 1 {
                t.Errorf("single-point offset: expected 1 pt, got %d", len(off.Points))
        }
}

func TestPolyline_TrimAt_ZeroLength(t *testing.T) {
        p := Polyline{Points: []Point{{5, 5}, {5, 5}}}
        a, b := p.TrimAt(0.5)
        assertNoNaN(t, a.Points, "trim-zero-a")
        assertNoNaN(t, b.Points, "trim-zero-b")
}

func TestFilterBySegment_MixedInsideOutside(t *testing.T) {
        s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
        pts := filterBySegment(s, []Point{{5, 0}, {15, 0}})
        if len(pts) != 1 || math.Abs(pts[0].X-5) > 1e-9 {
                t.Errorf("mixed: expected [5,0], got %v", pts)
        }
}

func TestFilterBySegment_ZeroLengthSeg(t *testing.T) {
        s := Segment{Start: Point{3, 0}, End: Point{3, 0}}
        pts := filterBySegment(s, []Point{{3, 0}, {5, 0}})
        if len(pts) != 1 {
                t.Errorf("zero-length seg: expected 1 (near start), got %d", len(pts))
        }
}

func TestFilterBySegment_ZeroLengthSegMiss(t *testing.T) {
        s := Segment{Start: Point{3, 0}, End: Point{3, 0}}
        pts := filterBySegment(s, []Point{{9, 0}})
        if len(pts) != 0 {
                t.Errorf("zero-length miss: expected 0, got %d", len(pts))
        }
}


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

func TestBezierSpline_PointAt_AtOne(t *testing.T) {
        sp := NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})
        p := sp.PointAt(1.0)
        if math.Abs(p.X-10) > 1e-9 || math.Abs(p.Y) > 1e-9 {
                t.Errorf("PointAt(1.0): %v", p)
        }
}

func TestBezierSpline_PointAt_SegBoundaryClamped(t *testing.T) {
        // 7 control points = 2 segments; t=0.5 lands exactly on the segment boundary.
        sp := NewBezierSpline([]Point{
                {0, 0}, {2, 4}, {4, 4}, {6, 0}, {8, -4}, {10, -4}, {12, 0},
        })
        assertPointValid(t, sp.PointAt(0.5), "bezier-boundary")
}

func TestNURBSSpline_PointAt_BelowLo(t *testing.T) {
        sp := NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []Point{{0, 0}, {5, 5}, {10, 0}},
                nil,
        )
        assertPointValid(t, sp.PointAt(-1), "nurbs-below-lo")
}

func TestNURBSSpline_PointAt_ZeroWeight(t *testing.T) {
        // Construct a NURBS where all weights sum to 0 at the evaluation point
        // to hit the w < Epsilon branch. Use zero weights.
        sp := NURBSSpline{
                Degree:   1,
                Knots:    []float64{0, 0, 1, 1},
                Controls: []Point{{0, 0}, {10, 0}},
                Weights:  []float64{0, 0},
        }
        p := sp.PointAt(0.5)
        if p.X != 0 || p.Y != 0 {
                t.Errorf("zero weight: expected {0,0}, got %v", p)
        }
}

func TestIntersectCollinear_Parallel(t *testing.T) {
        a := Segment{Start: Point{0, 0}, End: Point{5, 0}}
        b := Segment{Start: Point{0, 1}, End: Point{5, 1}}
        pts := intersectCollinearSegments(a, b)
        if pts != nil {
                t.Errorf("parallel non-collinear: expected nil, got %v", pts)
        }
}

func TestIntersectCollinear_TouchPoint(t *testing.T) {
        // Two segments that just touch at a single point — hi == lo → one point.
        a := Segment{Start: Point{0, 0}, End: Point{5, 0}}
        b := Segment{Start: Point{5, 0}, End: Point{10, 0}}
        pts := IntersectSegments(a, b)
        // They share exactly the point (5,0).
        if len(pts) != 1 {
                t.Errorf("touch: expected 1, got %d", len(pts))
        }
}

func TestIntersectLineWith_DefaultNil(t *testing.T) {
        l := Line{P: Point{0, 0}, Q: Point{10, 0}}
        pts := intersectLineWith(l, nil)
        if pts != nil {
                t.Errorf("nil entity: expected nil, got %v", pts)
        }
}

func TestIntersectRayWith_DefaultNil(t *testing.T) {
        r := Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}
        pts := intersectRayWith(r, nil)
        if pts != nil {
                t.Errorf("nil entity: expected nil, got %v", pts)
        }
}

func TestIntersectLineCircle_ZeroLengthLine(t *testing.T) {
        l := Line{P: Point{3, 0}, Q: Point{3, 0}}
        c := Circle{Center: Point{0, 0}, Radius: 5}
        pts := IntersectLineCircle(l, c)
        if pts != nil {
                t.Errorf("zero-length line: expected nil, got %v", pts)
        }
}

func TestIntersectCollinear_ZeroLengthA(t *testing.T) {
        // Both segments must be at the same point to pass the collinearity check.
        // A zero-length segment at {3,0}; B spans {3,0}→{3,0} — same point.
        // DistToPoint on the zero-length line is distance from {3,0} to {3,0} = 0.
        a := Segment{Start: Point{3, 0}, End: Point{3, 0}}
        b := Segment{Start: Point{3, 0}, End: Point{3, 0}}
        pts := intersectCollinearSegments(a, b)
        if pts != nil {
                t.Errorf("zero-length A: expected nil, got %v", pts)
        }
}

func TestPolyline_TrimAt_BeyondEnd(t *testing.T) {
        p := Polyline{Points: []Point{{0, 0}, {5, 0}, {10, 0}}}
        first, second := p.TrimAt(2.0)
        if len(first.Points) < 2 {
                t.Error("TrimAt>1: first part too short")
        }
        assertNoNaN(t, second.Points, "trim-beyond-end-second")
}

func TestUnmarshalEntity_BadEllipse(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{"kind":"ellipse","data":"bad"}`))
        if err == nil {
                t.Error("expected error for bad ellipse data")
        }
}

func TestUnmarshalEntity_UnknownKind(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`{"kind":"unknown_xyz","data":{}}`))
        if err == nil {
                t.Error("expected error for unknown kind")
        }
}

func TestUnmarshalEntity_BrokenOuterJSON(t *testing.T) {
        _, err := UnmarshalEntity([]byte(`not-json`))
        if err == nil {
                t.Error("expected error for broken JSON")
        }
}

func TestBezierSpline_PointAt_NearOne(t *testing.T) {
        // t just below 1.0 on a 2-segment spline; verifies the t>=1.0 guard is not hit
        // and the result is the correct endpoint neighbourhood.
        sp := NewBezierSpline([]Point{
                {0, 0}, {2, 4}, {4, 4}, {6, 0},
                {8, -4}, {10, -4}, {12, 0},
        })
        assertPointValid(t, sp.PointAt(math.Nextafter(1.0, 0)), "bezier-near-one")
}

func TestRawEntity_MarshalRoundtrip(t *testing.T) {
        re := RawEntity{
                EntityKind: KindCircle,
                Data:       json.RawMessage(`{"center":{"x":0,"y":0},"radius":5}`),
        }
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
