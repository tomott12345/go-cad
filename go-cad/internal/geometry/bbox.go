package geometry

import "math"

// BBox is an axis-aligned bounding box.
type BBox struct {
	Min, Max Point
}

// EmptyBBox returns an inverted bounding box (suitable for expansion via Extend).
func EmptyBBox() BBox {
	return BBox{
		Min: Point{math.Inf(1), math.Inf(1)},
		Max: Point{math.Inf(-1), math.Inf(-1)},
	}
}

// IsEmpty reports whether the bounding box contains no area.
func (b BBox) IsEmpty() bool {
	return math.IsInf(b.Min.X, 1) || b.Min.X > b.Max.X
}

// Extend expands the bounding box to include point p.
func (b BBox) Extend(p Point) BBox {
	return BBox{
		Min: Point{math.Min(b.Min.X, p.X), math.Min(b.Min.Y, p.Y)},
		Max: Point{math.Max(b.Max.X, p.X), math.Max(b.Max.Y, p.Y)},
	}
}

// Union returns the bounding box that contains both b and c.
func (b BBox) Union(c BBox) BBox {
	if b.IsEmpty() {
		return c
	}
	if c.IsEmpty() {
		return b
	}
	return BBox{
		Min: Point{math.Min(b.Min.X, c.Min.X), math.Min(b.Min.Y, c.Min.Y)},
		Max: Point{math.Max(b.Max.X, c.Max.X), math.Max(b.Max.Y, c.Max.Y)},
	}
}

// Contains reports whether p is inside (or on the edge of) the bounding box.
func (b BBox) Contains(p Point) bool {
	return p.X >= b.Min.X && p.X <= b.Max.X && p.Y >= b.Min.Y && p.Y <= b.Max.Y
}

// Overlaps reports whether two bounding boxes overlap.
func (b BBox) Overlaps(c BBox) bool {
	return b.Min.X <= c.Max.X && b.Max.X >= c.Min.X &&
		b.Min.Y <= c.Max.Y && b.Max.Y >= c.Min.Y
}

// Center returns the center of the bounding box.
func (b BBox) Center() Point {
	return Point{(b.Min.X + b.Max.X) / 2, (b.Min.Y + b.Max.Y) / 2}
}

// Width returns the width of the bounding box.
func (b BBox) Width() float64 { return b.Max.X - b.Min.X }

// Height returns the height of the bounding box.
func (b BBox) Height() float64 { return b.Max.Y - b.Min.Y }

// Expand returns a bounding box expanded by margin on all sides.
func (b BBox) Expand(margin float64) BBox {
	return BBox{
		Min: Point{b.Min.X - margin, b.Min.Y - margin},
		Max: Point{b.Max.X + margin, b.Max.Y + margin},
	}
}
