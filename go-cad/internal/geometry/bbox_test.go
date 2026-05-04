package geometry

import "testing"

func TestBBoxEmpty(t *testing.T) {
	b := EmptyBBox()
	if !b.IsEmpty() {
		t.Error("EmptyBBox should report IsEmpty")
	}
}

func TestBBoxExtend(t *testing.T) {
	b := EmptyBBox().Extend(Point{1, 2}).Extend(Point{5, 7})
	if !b.Min.Near(Point{1, 2}) || !b.Max.Near(Point{5, 7}) {
		t.Errorf("Extend: got min=%v max=%v", b.Min, b.Max)
	}
}

func TestBBoxContains(t *testing.T) {
	b := BBox{Point{0, 0}, Point{10, 10}}
	if !b.Contains(Point{5, 5}) {
		t.Error("Contains: interior point should be true")
	}
	if !b.Contains(Point{0, 0}) {
		t.Error("Contains: corner should be true")
	}
	if b.Contains(Point{11, 5}) {
		t.Error("Contains: exterior should be false")
	}
}

func TestBBoxOverlaps(t *testing.T) {
	a := BBox{Point{0, 0}, Point{5, 5}}
	b := BBox{Point{3, 3}, Point{8, 8}}
	if !a.Overlaps(b) {
		t.Error("Overlaps: overlapping boxes should return true")
	}
	c := BBox{Point{6, 6}, Point{10, 10}}
	if a.Overlaps(c) {
		t.Error("Overlaps: non-overlapping boxes should return false")
	}
}

func TestBBoxUnion(t *testing.T) {
	a := BBox{Point{0, 0}, Point{5, 5}}
	b := BBox{Point{3, 3}, Point{10, 10}}
	u := a.Union(b)
	if !u.Min.Near(Point{0, 0}) || !u.Max.Near(Point{10, 10}) {
		t.Errorf("Union: got min=%v max=%v", u.Min, u.Max)
	}
}

func TestBBoxCenter(t *testing.T) {
	b := BBox{Point{0, 0}, Point{10, 10}}
	if !b.Center().Near(Point{5, 5}) {
		t.Errorf("Center: got %v", b.Center())
	}
}

func TestBBoxExpand(t *testing.T) {
	b := BBox{Point{1, 1}, Point{5, 5}}
	e := b.Expand(2)
	if !e.Min.Near(Point{-1, -1}) || !e.Max.Near(Point{7, 7}) {
		t.Errorf("Expand: got min=%v max=%v", e.Min, e.Max)
	}
}
