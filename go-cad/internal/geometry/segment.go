package geometry

import "math"

// Segment is a directed line segment from Start to End.
type Segment struct {
        Start, End Point
}

// NewSegment constructs a Segment.
func NewSegment(start, end Point) Segment { return Segment{start, end} }

// Dir returns the unnormalised direction vector (End - Start).
func (s Segment) Dir() Point { return s.End.Sub(s.Start) }

// Length returns the length of the segment.
func (s Segment) Length() float64 { return s.Start.Dist(s.End) }

// Midpoint returns the midpoint of the segment.
func (s Segment) Midpoint() Point { return s.Start.Lerp(s.End, 0.5) }

// PointAt returns the point at parametric position t ∈ [0,1].
func (s Segment) PointAt(t float64) Point { return s.Start.Lerp(s.End, t) }

// BoundingBox returns the axis-aligned bounding box of the segment.
func (s Segment) BoundingBox() BBox {
        return EmptyBBox().Extend(s.Start).Extend(s.End)
}

// ClosestPoint returns the nearest point on the segment to p,
// clamped to [Start, End], and the parametric t value.
func (s Segment) ClosestPoint(p Point) (Point, float64) {
        d := s.Dir()
        l2 := d.Len2()
        if l2 < Epsilon*Epsilon {
                return s.Start, 0
        }
        t := p.Sub(s.Start).Dot(d) / l2
        t = math.Max(0, math.Min(1, t))
        return s.PointAt(t), t
}

// DistToPoint returns the minimum distance from p to the segment.
func (s Segment) DistToPoint(p Point) float64 {
        cp, _ := s.ClosestPoint(p)
        return p.Dist(cp)
}

// TrimAt splits the segment at parametric t, returning the two halves.
func (s Segment) TrimAt(t float64) (Segment, Segment) {
        mid := s.PointAt(t)
        return Segment{s.Start, mid}, Segment{mid, s.End}
}

// Offset returns a segment offset perpendicular to itself by dist.
// Positive dist offsets to the left (CCW side).
func (s Segment) Offset(dist float64) Segment {
        d := s.Dir().Normalize().Perp().Scale(dist)
        return Segment{s.Start.Add(d), s.End.Add(d)}
}

// Contains reports whether point p lies on the segment (within Epsilon).
func (s Segment) Contains(p Point) bool {
        cp, _ := s.ClosestPoint(p)
        return p.Dist(cp) < Epsilon
}

// Line represents an infinite line through two points (or a point + direction).
type Line struct {
        P Point `json:"p"`
        Q Point `json:"q"`
}

// Dir returns the direction vector Q - P (unnormalised).
func (l Line) Dir() Point { return l.Q.Sub(l.P) }

// Normal returns the unit normal to the line.
func (l Line) Normal() Point { return l.Dir().Perp().Normalize() }

// DistToPoint returns the signed distance from p to the line (positive = left).
func (l Line) DistToPoint(p Point) float64 {
        d := l.Dir()
        len := d.Len()
        if len < Epsilon {
                return l.P.Dist(p)
        }
        return d.Cross(p.Sub(l.P)) / len
}

// ClosestPoint returns the foot of the perpendicular from p to the infinite line.
func (l Line) ClosestPoint(p Point) Point {
        d := l.Dir()
        l2 := d.Len2()
        if l2 < Epsilon*Epsilon {
                return l.P
        }
        t := p.Sub(l.P).Dot(d) / l2
        return l.P.Add(d.Scale(t))
}

// PerpendicularFoot returns the foot of the perpendicular from p (alias for ClosestPoint).
func (l Line) PerpendicularFoot(p Point) Point { return l.ClosestPoint(p) }
