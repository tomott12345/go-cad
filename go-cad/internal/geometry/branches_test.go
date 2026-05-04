// branches_test.go targets every remaining uncovered branch to push geometry
// package statement coverage above 95%.
package geometry

import (
	"math"
	"testing"
)

// ─── arc.go:ClosestPoint remaining branches ───────────────────────────────────

func TestArc_ClosestPoint_OutsideArc_ClampStart(t *testing.T) {
	// Arc from 80° to 100°; query point at 0° should snap to start
	a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 80, EndDeg: 100}
	// Point at (10, 0) projects to 0° which is before StartDeg=80 → snap to start
	p := a.ClosestPoint(Point{10, 0})
	startPt := a.StartPoint()
	if p.Dist(startPt) > 0.5 {
		t.Errorf("arc clamp to start: %v, want near %v", p, startPt)
	}
}

func TestArc_ClosestPoint_OutsideArc_ClampEnd(t *testing.T) {
	// Arc from 0° to 20°; query point at 90° should snap to end
	a := Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 20}
	p := a.ClosestPoint(Point{0, 10})
	endPt := a.EndPoint()
	if p.Dist(endPt) > 0.5 {
		t.Errorf("arc clamp to end: %v, want near %v", p, endPt)
	}
}

// ─── bbox.go:Union ────────────────────────────────────────────────────────────

func TestBBox_Union_NonEmptyNonEmpty(t *testing.T) {
	a := BBox{Min: Point{-5, -5}, Max: Point{1, 1}}
	b := BBox{Min: Point{0, 0}, Max: Point{5, 5}}
	u := a.Union(b)
	if u.Min.X != -5 || u.Max.X != 5 || u.Min.Y != -5 || u.Max.Y != 5 {
		t.Errorf("Union: %+v", u)
	}
}

func TestBBox_Union_OtherEmpty(t *testing.T) {
	a := BBox{Min: Point{1, 2}, Max: Point{3, 4}}
	u := a.Union(EmptyBBox())
	if u.Min.X != 1 || u.Max.X != 3 {
		t.Errorf("Union with empty other: %+v", u)
	}
}

// ─── entity.go:EllipseEntity.TrimAt edge branches ────────────────────────────

func TestEllipseEntity_TrimAt_NearZero(t *testing.T) {
	e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
	// t close to 0 should clamp split to 1
	a, b := e.TrimAt(0.001)
	pa := a.(PolylineEntity)
	pb := b.(PolylineEntity)
	if len(pa.Points) < 2 {
		t.Error("TrimAt near-zero: first part too short")
	}
	if len(pb.Points) < 2 {
		t.Error("TrimAt near-zero: second part too short")
	}
}

func TestEllipseEntity_TrimAt_NearOne(t *testing.T) {
	e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
	// t close to 1 should clamp split to len-2
	a, b := e.TrimAt(0.999)
	pa := a.(PolylineEntity)
	pb := b.(PolylineEntity)
	if len(pa.Points) < 2 {
		t.Error("TrimAt near-one: first part too short")
	}
	if len(pb.Points) < 2 {
		t.Error("TrimAt near-one: second part too short")
	}
}

// ─── entity.go:MarshalEntity / UnmarshalEntity ───────────────────────────────

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
	// Valid JSON but bad inner data for an arc
	_, err := UnmarshalEntity([]byte(`{"kind":"arc","data":"not-an-object"}`))
	if err == nil {
		t.Error("expected error for bad arc data")
	}
}

// ─── intersect.go: remaining dispatcher arms ─────────────────────────────────

// intersectSegmentWith: NURBS arm
func TestIntersect_SegmentNURBS_Dispatcher(t *testing.T) {
	s := SegmentEntity{Segment{Start: Point{-1, 3}, End: Point{11, 3}}}
	sp := NURBSEntity{NewNURBSSpline(2,
		[]float64{0, 0, 0, 1, 1, 1},
		[]Point{{0, 0}, {5, 8}, {10, 0}},
		nil,
	)}
	pts := Intersect(s, sp)
	_ = pts // just ensure no panic and dispatcher path is exercised
}

// intersectCircleWith: Bezier and NURBS arms
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

// intersectArcWith: all arms
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
	_ = pts // either hits or doesn't, just no panic
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

// intersectEllipseWith: Bezier and NURBS arms
func TestIntersect_EllipseBezier2(t *testing.T) {
	e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
	b := BezierEntity{NewBezierSpline([]Point{{-3, 0}, {0, 5}, {3, 0}, {6, -5}})}
	pts := Intersect(e, b)
	_ = pts
}

func TestIntersect_EllipseNURBS2(t *testing.T) {
	e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
	sp := NURBSEntity{NewNURBSSpline(2,
		[]float64{0, 0, 0, 1, 1, 1},
		[]Point{{-5, 0}, {0, 5}, {5, 0}},
		nil,
	)}
	pts := Intersect(e, sp)
	_ = pts
}

// intersectPolylineWith: NURBS arm
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

// intersectPolylineWith: Bezier arm
func TestIntersect_PolylineBezier(t *testing.T) {
	p := PolylineEntity{Polyline{Points: []Point{{-10, 2}, {10, 2}}}}
	b := BezierEntity{NewBezierSpline([]Point{{-5, 0}, {-2, 5}, {2, 5}, {5, 0}})}
	pts := Intersect(p, b)
	if len(pts) == 0 {
		t.Error("polyline-bezier: expected intersection")
	}
}

// intersectCollinearSegments: the two-point overlap branch
func TestIntersectCollinear_Overlap(t *testing.T) {
	a := Segment{Start: Point{0, 0}, End: Point{7, 0}}
	b := Segment{Start: Point{3, 0}, End: Point{10, 0}}
	pts := IntersectSegments(a, b)
	if len(pts) != 2 {
		t.Errorf("overlap: expected 2 endpoints, got %d", len(pts))
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

// IntersectLineCircle: the tangent (discriminant=0) branch is exercised;
// make sure the miss branch coverage is filled.
func TestIntersectLineCircle_Tangent2(t *testing.T) {
	l := Line{P: Point{-10, 5}, Q: Point{10, 5}}
	c := Circle{Center: Point{0, 0}, Radius: 5}
	pts := IntersectLineCircle(l, c)
	if len(pts) != 1 {
		t.Fatalf("tangent at top: expected 1, got %d", len(pts))
	}
	if math.Abs(pts[0].Y-5) > 1e-6 {
		t.Errorf("tangent pt Y: %v", pts[0].Y)
	}
}

// filterBySegment: point exactly at segment endpoints
func TestFilterBySegment_AtEndpoints(t *testing.T) {
	s := Segment{Start: Point{0, 0}, End: Point{10, 0}}
	pts := filterBySegment(s, []Point{{0, 0}, {10, 0}})
	if len(pts) != 2 {
		t.Errorf("at endpoints: expected 2, got %d", len(pts))
	}
}

// ─── ray.go:TrimAt negative t branch ─────────────────────────────────────────

func TestRay_TrimAt_Negative(t *testing.T) {
	r := Ray{Origin: Point{5, 0}, Dir: Point{1, 0}}
	seg, ray := r.TrimAt(-3)
	// Negative t clamped to 0 → segment has zero length, ray origin == r.Origin
	if math.Abs(seg.Start.Dist(seg.End)) > 1e-9 {
		t.Errorf("negative TrimAt: segment should be zero-length, start=%v end=%v", seg.Start, seg.End)
	}
	if ray.Origin.Dist(r.Origin) > 1e-9 {
		t.Errorf("negative TrimAt: ray origin should be at r.Origin")
	}
}

// ─── segment.go: degenerate ClosestPoint (zero-length segment) ───────────────

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

// ─── segment.go:PerpendicularFoot (direct call) ───────────────────────────────

func TestLine_PerpendicularFoot(t *testing.T) {
	l := Line{P: Point{0, 0}, Q: Point{10, 0}}
	foot := l.PerpendicularFoot(Point{5, 7})
	if math.Abs(foot.X-5) > 1e-9 || math.Abs(foot.Y) > 1e-9 {
		t.Errorf("PerpendicularFoot: %v", foot)
	}
}

// ─── spline.go: BezierSpline degenerate paths ────────────────────────────────

func TestBezierSpline_NumSegments_TooFew(t *testing.T) {
	sp := NewBezierSpline([]Point{{0, 0}, {1, 1}})
	if sp.NumSegments() != 0 {
		t.Errorf("too few ctrl: expected 0, got %d", sp.NumSegments())
	}
}

func TestBezierSpline_PointAt_NoControls(t *testing.T) {
	sp := BezierSpline{}
	p := sp.PointAt(0.5) // should return zero Point, no panic
	_ = p
}

func TestBezierSpline_PointAt_TooFew(t *testing.T) {
	sp := NewBezierSpline([]Point{{3, 4}}) // 1 point → returns Controls[0]
	p := sp.PointAt(0.5)
	if p.X != 3 || p.Y != 4 {
		t.Errorf("single-point bezier PointAt: %v", p)
	}
}

func TestBezierSpline_ApproxPolyline_ZeroSeg(t *testing.T) {
	sp := NewBezierSpline([]Point{{1, 2}, {3, 4}}) // < 4 pts → returns Controls
	pts := sp.ApproxPolyline(10)
	if len(pts) != 2 {
		t.Errorf("degenerate ApproxPolyline: %v", pts)
	}
}

// ─── spline.go: NURBS PointAt at domain boundary ─────────────────────────────

func TestNURBSSpline_PointAt_AtHi(t *testing.T) {
	sp := NewNURBSSpline(2,
		[]float64{0, 0, 0, 1, 1, 1},
		[]Point{{0, 0}, {5, 5}, {10, 0}},
		nil,
	)
	// Clamp to just below hi; should not panic
	p := sp.PointAt(1.0)
	_ = p
}

// ─── polyline.go: Offset and TrimAt edge branches ────────────────────────────

func TestPolyline_Offset_ZeroLen(t *testing.T) {
	// A polyline with a zero-length segment should not panic
	p := Polyline{Points: []Point{{0, 0}, {0, 0}, {5, 0}}}
	off := p.Offset(1)
	_ = off
}

func TestPolyline_TrimAt_BeyondEnd(t *testing.T) {
	p := Polyline{Points: []Point{{0, 0}, {5, 0}, {10, 0}}}
	a, b := p.TrimAt(0.99)
	pa := a
	pb := b
	if len(pa.Points) < 2 {
		t.Errorf("near-end trim: first part too short")
	}
	if len(pb.Points) < 2 {
		t.Errorf("near-end trim: second part too short")
	}
}

func TestPolyline_TrimAt_Start(t *testing.T) {
	p := Polyline{Points: []Point{{0, 0}, {5, 0}, {10, 0}}}
	a, b := p.TrimAt(0.01)
	_ = a
	_ = b
}

// ─── intersect.go:intersectOrdererd remaining arms ───────────────────────────

// Make sure all ordered cases are covered (the dispatcher)
func TestIntersect_BezierNURBS2(t *testing.T) {
	b := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}})}
	sp := NURBSEntity{NewNURBSSpline(2,
		[]float64{0, 0, 0, 1, 1, 1},
		[]Point{{0, 3}, {5, -2}, {10, 3}},
		nil,
	)}
	pts := Intersect(sp, b) // reversed order
	_ = pts
}

func TestIntersect_NURBSNURBSdirect(t *testing.T) {
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

// ─── filterBySegment: outside both sides ─────────────────────────────────────

func TestFilterBySegment_BothOutside(t *testing.T) {
	s := Segment{Start: Point{0, 0}, End: Point{5, 0}}
	pts := filterBySegment(s, []Point{{-2, 0}, {8, 0}})
	if len(pts) != 0 {
		t.Errorf("both outside: expected 0, got %d", len(pts))
	}
}

// ─── IntersectArcs / IntersectCircleArc ──────────────────────────────────────

func TestIntersectArcs_NoOverlap(t *testing.T) {
	// Same circle, arcs on different sides
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

// ─── basisFunc degenerate: zero knot span ────────────────────────────────────

func TestNURBSSpline_BasisFunc_ZeroSpan(t *testing.T) {
	// Repeated knots create zero spans; ensure basisFunc doesn't panic
	sp := NewNURBSSpline(2,
		[]float64{0, 0, 0, 0.5, 0.5, 1, 1, 1},
		[]Point{{0, 0}, {3, 5}, {7, 5}, {10, 0}, {5, -3}},
		nil,
	)
	p := sp.PointAt(0.5)
	if math.IsNaN(p.X) || math.IsNaN(p.Y) {
		t.Errorf("NURBS with repeated knots NaN: %v", p)
	}
}

// ─── IntersectSegmentPolyline (not directly tested) ───────────────────────────

func TestIntersectSegmentPolyline_Direct(t *testing.T) {
	s := Segment{Start: Point{0, -5}, End: Point{0, 5}}
	p := Polyline{Points: []Point{{-5, 0}, {5, 0}}}
	pts := IntersectSegmentPolyline(s, p)
	if len(pts) != 1 {
		t.Fatalf("segment-polyline: expected 1, got %d", len(pts))
	}
}
