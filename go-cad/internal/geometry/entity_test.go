package geometry

import (
	"encoding/json"
	"testing"
)

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
	// Ensure RawEntity is valid JSON
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

func TestEntityInterface(t *testing.T) {
	entities := []Entity{
		SegmentEntity{Segment{Point{0, 0}, Point{10, 0}}},
		CircleEntity{Circle{Point{5, 5}, 3}},
		ArcEntity{Arc{Point{0, 0}, 5, 0, 90}},
		EllipseEntity{Ellipse{Point{0, 0}, 6, 3, 0}},
		PolylineEntity{Polyline{Points: []Point{{0, 0}, {5, 0}, {5, 5}}}},
		BezierEntity{BezierSpline{Controls: []Point{{0, 0}, {1, 2}, {3, 2}, {4, 0}}}},
	}
	for _, e := range entities {
		bb := e.BoundingBox()
		if bb.IsEmpty() {
			t.Errorf("%T BoundingBox is empty", e)
		}
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
