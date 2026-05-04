package geometry

import (
	"encoding/json"
	"math"
	"testing"
)

// ─── Existing JSON roundtrip tests ────────────────────────────────────────────

func TestMarshalUnmarshalSegment(t *testing.T) {
	e := SegmentEntity{Segment{Point{1, 2}, Point{3, 4}}}
	b, err := MarshalEntity(e)
	if err != nil {
		t.Fatalf("MarshalEntity: %v", err)
	}
	got, err := UnmarshalEntity(b)
	if err != nil {
		t.Fatalf("UnmarshalEntity: %v", err)
	}
	se, ok := got.(SegmentEntity)
	if !ok {
		t.Fatalf("Expected SegmentEntity, got %T", got)
	}
	if !se.Start.Near(e.Start) || !se.End.Near(e.End) {
		t.Errorf("Round-trip mismatch: got %v, want %v", se.Segment, e.Segment)
	}
}

func TestMarshalUnmarshalCircle(t *testing.T) {
	e := CircleEntity{Circle{Point{5, 6}, 7}}
	b, err := MarshalEntity(e)
	if err != nil {
		t.Fatalf("MarshalEntity: %v", err)
	}
	got, err := UnmarshalEntity(b)
	if err != nil {
		t.Fatalf("UnmarshalEntity: %v", err)
	}
	ce, ok := got.(CircleEntity)
	if !ok {
		t.Fatalf("Expected CircleEntity, got %T", got)
	}
	if !ce.Center.Near(e.Center) || ce.Radius != e.Radius {
		t.Errorf("Round-trip: got %v, want %v", ce.Circle, e.Circle)
	}
}

func TestMarshalUnmarshalArc(t *testing.T) {
	e := ArcEntity{Arc{Point{0, 0}, 3, 30, 150}}
	b, err := MarshalEntity(e)
	if err != nil {
		t.Fatalf("MarshalEntity: %v", err)
	}
	got, err := UnmarshalEntity(b)
	if err != nil {
		t.Fatalf("UnmarshalEntity: %v", err)
	}
	ae, ok := got.(ArcEntity)
	if !ok {
		t.Fatalf("Expected ArcEntity, got %T", got)
	}
	if ae.Radius != e.Radius || ae.StartDeg != e.StartDeg {
		t.Errorf("Round-trip: got %v, want %v", ae.Arc, e.Arc)
	}
}

func TestMarshalUnmarshalPolyline(t *testing.T) {
	e := PolylineEntity{Polyline{Points: []Point{{0, 0}, {5, 0}, {5, 5}}, Closed: true}}
	b, err := MarshalEntity(e)
	if err != nil {
		t.Fatalf("MarshalEntity: %v", err)
	}
	got, err := UnmarshalEntity(b)
	if err != nil {
		t.Fatalf("UnmarshalEntity: %v", err)
	}
	pe, ok := got.(PolylineEntity)
	if !ok {
		t.Fatalf("Expected PolylineEntity, got %T", got)
	}
	if len(pe.Points) != 3 {
		t.Errorf("Polyline point count: got %d, want 3", len(pe.Points))
	}
	if !pe.Closed {
		t.Error("Polyline Closed flag lost in round-trip")
	}
}

func TestUnmarshalInvalidKind(t *testing.T) {
	raw := `{"kind":"unknown","data":{}}`
	_, err := UnmarshalEntity([]byte(raw))
	if err == nil {
		t.Error("Expected error for unknown kind")
	}
}

func TestRawEntityJSON(t *testing.T) {
	e := SegmentEntity{Segment{Point{0, 0}, Point{10, 10}}}
	b, _ := MarshalEntity(e)
	var raw RawEntity
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("RawEntity JSON parse: %v", err)
	}
	if raw.EntityKind != KindSegment {
		t.Errorf("Kind: got %v, want %v", raw.EntityKind, KindSegment)
	}
}

// ─── Entity interface (BoundingBox, Length, Offset) ───────────────────────────

func TestEntityInterface(t *testing.T) {
	entities := []Entity{
		SegmentEntity{Segment{Point{0, 0}, Point{10, 0}}},
		LineEntity{Line{Point{0, 0}, Point{10, 0}}},
		RayEntity{Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}},
		CircleEntity{Circle{Point{5, 5}, 3}},
		ArcEntity{Arc{Point{0, 0}, 5, 0, 90}},
		EllipseEntity{Ellipse{Point{0, 0}, 6, 3, 0}},
		PolylineEntity{Polyline{Points: []Point{{0, 0}, {5, 0}, {5, 5}}}},
		BezierEntity{BezierSpline{Controls: []Point{{0, 0}, {1, 2}, {3, 2}, {4, 0}}}},
	}
	for _, e := range entities {
		l := e.Length()
		if l < 0 {
			t.Errorf("%T Length is negative: %v", e, l)
		}
		off := e.Offset(1)
		if off == nil {
			t.Errorf("%T Offset returned nil", e)
		}
	}
}

// ─── TrimAt interface ─────────────────────────────────────────────────────────

func TestSegmentEntity_TrimAt(t *testing.T) {
	e := SegmentEntity{Segment{Start: Point{0, 0}, End: Point{10, 0}}}
	a, b := e.TrimAt(0.4)
	sa := a.(SegmentEntity)
	sb := b.(SegmentEntity)
	if math.Abs(sa.End.X-4) > 1e-9 {
		t.Errorf("first half end: want X=4, got %v", sa.End.X)
	}
	if math.Abs(sb.Start.X-4) > 1e-9 {
		t.Errorf("second half start: want X=4, got %v", sb.Start.X)
	}
}

func TestArcEntity_TrimAt(t *testing.T) {
	e := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
	a, b := e.TrimAt(0.5)
	ae := a.(ArcEntity)
	be := b.(ArcEntity)
	if math.Abs(ae.Arc.EndDeg-90) > 1e-6 {
		t.Errorf("first half end angle: want 90°, got %v", ae.Arc.EndDeg)
	}
	if math.Abs(be.Arc.StartDeg-90) > 1e-6 {
		t.Errorf("second half start angle: want 90°, got %v", be.Arc.StartDeg)
	}
}

func TestCircleEntity_TrimAt(t *testing.T) {
	e := CircleEntity{Circle{Center: Point{0, 0}, Radius: 3}}
	a, b := e.TrimAt(0.5)
	_, ok1 := a.(ArcEntity)
	_, ok2 := b.(ArcEntity)
	if !ok1 || !ok2 {
		t.Errorf("expected two ArcEntitys, got %T and %T", a, b)
	}
}

func TestPolylineEntity_TrimAt(t *testing.T) {
	e := PolylineEntity{Polyline{Points: []Point{{0, 0}, {5, 0}, {10, 0}}}}
	a, b := e.TrimAt(0.5)
	pa := a.(PolylineEntity)
	pb := b.(PolylineEntity)
	if len(pa.Points) < 2 || len(pb.Points) < 2 {
		t.Errorf("expected both halves ≥2 pts; got %d and %d", len(pa.Points), len(pb.Points))
	}
}

func TestEllipseEntity_TrimAt(t *testing.T) {
	e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
	a, b := e.TrimAt(0.5)
	_, ok1 := a.(PolylineEntity)
	_, ok2 := b.(PolylineEntity)
	if !ok1 || !ok2 {
		t.Errorf("expected PolylineEntitys, got %T and %T", a, b)
	}
}

// ─── Ray type ─────────────────────────────────────────────────────────────────

func TestRayEntity_Kind(t *testing.T) {
	r := RayEntity{Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}}
	if r.Kind() != KindRay {
		t.Errorf("kind: got %v, want %v", r.Kind(), KindRay)
	}
}

func TestRayEntity_ClosestPoint_OnRay(t *testing.T) {
	r := RayEntity{Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}}
	p := r.ClosestPoint(Point{5, 3})
	if math.Abs(p.X-5) > 1e-9 || math.Abs(p.Y) > 1e-9 {
		t.Errorf("closest point: got %v, want {5 0}", p)
	}
}

func TestRayEntity_ClosestPoint_BehindOrigin(t *testing.T) {
	r := RayEntity{Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}}
	p := r.ClosestPoint(Point{-5, 0})
	if math.Abs(p.X) > 1e-9 {
		t.Errorf("behind origin: expected origin {0,0}, got %v", p)
	}
}

func TestRayEntity_TrimAt(t *testing.T) {
	r := RayEntity{Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}}
	seg, ray := r.TrimAt(7)
	s := seg.(SegmentEntity)
	if math.Abs(s.End.X-7) > 1e-9 {
		t.Errorf("trim segment end: want X=7, got %v", s.End.X)
	}
	ra := ray.(RayEntity)
	if math.Abs(ra.Ray.Origin.X-7) > 1e-9 {
		t.Errorf("trim ray origin: want X=7, got %v", ra.Ray.Origin.X)
	}
}

func TestRayEntity_IntersectWithSegment(t *testing.T) {
	r := Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}
	s := Segment{Start: Point{5, -3}, End: Point{5, 3}}
	pts := r.IntersectWithSegment(s)
	if len(pts) != 1 {
		t.Fatalf("ray-segment: expected 1, got %d", len(pts))
	}
	if math.Abs(pts[0].X-5) > 1e-9 {
		t.Errorf("intersection: want X=5, got %v", pts[0].X)
	}
}

func TestRayEntity_BehindOriginNoIntersect(t *testing.T) {
	r := Ray{Origin: Point{10, 0}, Dir: Point{1, 0}} // pointing right
	s := Segment{Start: Point{5, -3}, End: Point{5, 3}} // segment at X=5 (behind)
	pts := r.IntersectWithSegment(s)
	if len(pts) != 0 {
		t.Errorf("ray pointing away: expected 0 intersections, got %d", len(pts))
	}
}

// ─── LineEntity ───────────────────────────────────────────────────────────────

func TestLineEntity_Kind(t *testing.T) {
	l := LineEntity{Line{P: Point{0, 0}, Q: Point{1, 0}}}
	if l.Kind() != KindLine {
		t.Errorf("kind: got %v, want %v", l.Kind(), KindLine)
	}
}

func TestLineEntity_ClosestPoint(t *testing.T) {
	l := LineEntity{Line{P: Point{0, 0}, Q: Point{10, 0}}}
	p := l.ClosestPoint(Point{5, 7})
	if math.Abs(p.X-5) > 1e-9 || math.Abs(p.Y) > 1e-9 {
		t.Errorf("closest: got %v, want {5 0}", p)
	}
}

func TestLineEntity_TrimAt(t *testing.T) {
	l := LineEntity{Line{P: Point{0, 0}, Q: Point{10, 0}}}
	seg, ray := l.TrimAt(3)
	s, ok := seg.(SegmentEntity)
	if !ok {
		t.Fatalf("expected SegmentEntity, got %T", seg)
	}
	if math.Abs(s.End.X-3) > 1e-9 {
		t.Errorf("line trim segment end: want X=3, got %v", s.End.X)
	}
	_, ok2 := ray.(RayEntity)
	if !ok2 {
		t.Errorf("expected RayEntity for second part, got %T", ray)
	}
}

// ─── JSON roundtrip for Line and Ray ─────────────────────────────────────────

func TestMarshalUnmarshalLine(t *testing.T) {
	e := LineEntity{Line{P: Point{0, 0}, Q: Point{5, 5}}}
	b, err := MarshalEntity(e)
	if err != nil {
		t.Fatalf("MarshalEntity: %v", err)
	}
	got, err := UnmarshalEntity(b)
	if err != nil {
		t.Fatalf("UnmarshalEntity: %v", err)
	}
	le, ok := got.(LineEntity)
	if !ok {
		t.Fatalf("expected LineEntity, got %T", got)
	}
	if !le.P.Near(e.P) || !le.Q.Near(e.Q) {
		t.Errorf("roundtrip mismatch: got %v, want %v", le.Line, e.Line)
	}
}

func TestMarshalUnmarshalRay(t *testing.T) {
	e := RayEntity{Ray{Origin: Point{1, 2}, Dir: Point{3, 4}}}
	b, err := MarshalEntity(e)
	if err != nil {
		t.Fatalf("MarshalEntity: %v", err)
	}
	got, err := UnmarshalEntity(b)
	if err != nil {
		t.Fatalf("UnmarshalEntity: %v", err)
	}
	re, ok := got.(RayEntity)
	if !ok {
		t.Fatalf("expected RayEntity, got %T", got)
	}
	if !re.Ray.Origin.Near(e.Ray.Origin) || !re.Ray.Dir.Near(e.Ray.Dir) {
		t.Errorf("roundtrip mismatch: got %v, want %v", re.Ray, e.Ray)
	}
}

// ─── Centralized Intersect dispatcher ────────────────────────────────────────

func TestIntersect_SegmentSegment(t *testing.T) {
	h := SegmentEntity{Segment{Start: Point{0, 5}, End: Point{10, 5}}}
	v := SegmentEntity{Segment{Start: Point{5, 0}, End: Point{5, 10}}}
	pts := Intersect(h, v)
	if len(pts) != 1 {
		t.Fatalf("expected 1 intersection, got %d", len(pts))
	}
	if math.Abs(pts[0].X-5) > 1e-9 || math.Abs(pts[0].Y-5) > 1e-9 {
		t.Errorf("wrong intersection: %v", pts[0])
	}
}

func TestIntersect_Symmetric(t *testing.T) {
	a := CircleEntity{Circle{Center: Point{0, 0}, Radius: 5}}
	b := SegmentEntity{Segment{Start: Point{-10, 0}, End: Point{10, 0}}}
	pts1 := Intersect(a, b)
	pts2 := Intersect(b, a)
	if len(pts1) != len(pts2) {
		t.Errorf("Intersect not symmetric: %d vs %d", len(pts1), len(pts2))
	}
}

func TestIntersect_RayCircle(t *testing.T) {
	r := RayEntity{Ray{Origin: Point{-10, 0}, Dir: Point{1, 0}}}
	c := CircleEntity{Circle{Center: Point{0, 0}, Radius: 3}}
	pts := Intersect(r, c)
	if len(pts) != 2 {
		t.Fatalf("ray-circle: expected 2, got %d", len(pts))
	}
}

func TestIntersect_LineArc(t *testing.T) {
	l := LineEntity{Line{P: Point{-10, 0}, Q: Point{10, 0}}}
	a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
	pts := Intersect(l, a)
	if len(pts) != 2 {
		t.Fatalf("line-arc: expected 2, got %d", len(pts))
	}
}

func TestIntersect_CirclePolyline(t *testing.T) {
	c := CircleEntity{Circle{Center: Point{0, 0}, Radius: 5}}
	p := PolylineEntity{Polyline{Points: []Point{{-10, 0}, {10, 0}}}}
	pts := Intersect(c, p)
	if len(pts) != 2 {
		t.Fatalf("circle-polyline: expected 2, got %d", len(pts))
	}
}

func TestIntersect_ArcPolyline(t *testing.T) {
	a := ArcEntity{Arc{Center: Point{0, 0}, Radius: 5, StartDeg: 0, EndDeg: 180}}
	p := PolylineEntity{Polyline{Points: []Point{{-10, 0}, {10, 0}}}}
	pts := Intersect(a, p)
	if len(pts) != 2 {
		t.Fatalf("arc-polyline: expected 2, got %d", len(pts))
	}
}

func TestIntersect_CircleEllipse(t *testing.T) {
	c := CircleEntity{Circle{Center: Point{0, 0}, Radius: 3}}
	e := EllipseEntity{Ellipse{Center: Point{0, 0}, A: 5, B: 3, Rotation: 0}}
	pts := Intersect(c, e)
	// Circle r=3 == ellipse minor B=3, tangent on Y-axis: 2 points
	if len(pts) < 2 {
		t.Errorf("circle-ellipse: expected ≥2, got %d", len(pts))
	}
}

func TestIntersect_BezierBezier(t *testing.T) {
	h := BezierEntity{NewBezierSpline([]Point{{0, 0}, {3, 0}, {7, 0}, {10, 0}})}
	v := BezierEntity{NewBezierSpline([]Point{{5, -5}, {5, -1}, {5, 1}, {5, 5}})}
	pts := Intersect(h, v)
	if len(pts) == 0 {
		t.Error("bezier-bezier: expected at least one intersection")
	}
}

func TestIntersect_LineSegment(t *testing.T) {
	l := LineEntity{Line{P: Point{0, 0}, Q: Point{10, 0}}}
	s := SegmentEntity{Segment{Start: Point{5, -3}, End: Point{5, 3}}}
	pts := Intersect(l, s)
	if len(pts) != 1 {
		t.Fatalf("line-segment: expected 1, got %d", len(pts))
	}
	if math.Abs(pts[0].X-5) > 1e-9 {
		t.Errorf("intersection: want X=5, got %v", pts[0])
	}
}

// ─── JSON tag verification for Line and Ray ───────────────────────────────────

func TestLineJSON_Tags(t *testing.T) {
	l := Line{P: Point{1, 2}, Q: Point{3, 4}}
	b, _ := json.Marshal(l)
	var l2 Line
	if err := json.Unmarshal(b, &l2); err != nil {
		t.Errorf("Line unmarshal: %v", err)
	}
	if l2.P.X != 1 || l2.Q.Y != 4 {
		t.Errorf("Line roundtrip failed: %+v", l2)
	}
}

func TestRayJSON_Tags(t *testing.T) {
	r := Ray{Origin: Point{0, 0}, Dir: Point{1, 0}}
	b, _ := json.Marshal(r)
	var r2 Ray
	if err := json.Unmarshal(b, &r2); err != nil {
		t.Errorf("Ray unmarshal: %v", err)
	}
	if r2.Origin.X != 0 || r2.Dir.X != 1 {
		t.Errorf("Ray roundtrip failed: %+v", r2)
	}
}
