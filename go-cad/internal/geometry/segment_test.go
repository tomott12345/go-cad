package geometry

import (
	"math"
	"testing"
)

func TestSegmentLength(t *testing.T) {
	s := Segment{Point{0, 0}, Point{3, 4}}
	if math.Abs(s.Length()-5) > Epsilon {
		t.Errorf("Length: got %v, want 5", s.Length())
	}
}

func TestSegmentMidpoint(t *testing.T) {
	s := Segment{Point{0, 0}, Point{10, 0}}
	if !s.Midpoint().Near(Point{5, 0}) {
		t.Errorf("Midpoint: got %v", s.Midpoint())
	}
}

func TestSegmentPointAt(t *testing.T) {
	s := Segment{Point{0, 0}, Point{10, 0}}
	if !s.PointAt(0.25).Near(Point{2.5, 0}) {
		t.Errorf("PointAt: got %v", s.PointAt(0.25))
	}
}

func TestSegmentClosestPoint(t *testing.T) {
	s := Segment{Point{0, 0}, Point{10, 0}}
	// Point above the middle
	cp, tt := s.ClosestPoint(Point{5, 5})
	if !cp.Near(Point{5, 0}) {
		t.Errorf("ClosestPoint: got %v, want {5,0}", cp)
	}
	if math.Abs(tt-0.5) > Epsilon {
		t.Errorf("ClosestPoint t: got %v, want 0.5", tt)
	}
	// Point past end — clamp to end
	cp2, t2 := s.ClosestPoint(Point{20, 0})
	if !cp2.Near(Point{10, 0}) {
		t.Errorf("ClosestPoint clamp: got %v, want {10,0}", cp2)
	}
	if math.Abs(t2-1) > Epsilon {
		t.Errorf("ClosestPoint t clamp: got %v, want 1", t2)
	}
}

func TestSegmentOffset(t *testing.T) {
	s := Segment{Point{0, 0}, Point{10, 0}}
	off := s.Offset(5)
	// Offset left (CCW) of a rightward segment should move upward
	if !off.Start.Near(Point{0, 5}) || !off.End.Near(Point{10, 5}) {
		t.Errorf("Offset: got start=%v end=%v, want start={0,5} end={10,5}", off.Start, off.End)
	}
}

func TestSegmentTrimAt(t *testing.T) {
	s := Segment{Point{0, 0}, Point{10, 0}}
	a, b := s.TrimAt(0.3)
	if !a.End.Near(Point{3, 0}) {
		t.Errorf("TrimAt a.End: got %v, want {3,0}", a.End)
	}
	if !b.Start.Near(Point{3, 0}) {
		t.Errorf("TrimAt b.Start: got %v, want {3,0}", b.Start)
	}
}

func TestSegmentBoundingBox(t *testing.T) {
	s := Segment{Point{-1, -2}, Point{3, 4}}
	bb := s.BoundingBox()
	if !bb.Min.Near(Point{-1, -2}) || !bb.Max.Near(Point{3, 4}) {
		t.Errorf("BoundingBox: got min=%v max=%v", bb.Min, bb.Max)
	}
}

func TestLineClosestPoint(t *testing.T) {
	l := Line{Point{0, 0}, Point{10, 0}}
	// Point above line
	foot := l.ClosestPoint(Point{5, 7})
	if !foot.Near(Point{5, 0}) {
		t.Errorf("Line.ClosestPoint: got %v, want {5,0}", foot)
	}
}

func TestLineDistToPoint(t *testing.T) {
	l := Line{Point{0, 0}, Point{10, 0}}
	d := l.DistToPoint(Point{5, 3})
	if math.Abs(d-3) > Epsilon {
		t.Errorf("Line.DistToPoint: got %v, want 3", d)
	}
	// Signed: point below should be negative
	d2 := l.DistToPoint(Point{5, -3})
	if math.Abs(d2+3) > Epsilon {
		t.Errorf("Line.DistToPoint signed: got %v, want -3", d2)
	}
}
