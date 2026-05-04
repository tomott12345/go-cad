package geometry

import "math"

// Polyline is a sequence of connected line segments.
type Polyline struct {
	Points []Point
	Closed bool
}

// NewPolyline constructs a Polyline.
func NewPolyline(pts []Point, closed bool) Polyline {
	return Polyline{Points: pts, Closed: closed}
}

// NumSegments returns the number of segments in the polyline.
func (p Polyline) NumSegments() int {
	n := len(p.Points)
	if n < 2 {
		return 0
	}
	if p.Closed {
		return n
	}
	return n - 1
}

// Segment returns the i-th segment of the polyline.
func (p Polyline) Segment(i int) Segment {
	n := len(p.Points)
	return Segment{p.Points[i % n], p.Points[(i+1) % n]}
}

// Length returns the total length of the polyline.
func (p Polyline) Length() float64 {
	total := 0.0
	for i := 0; i < p.NumSegments(); i++ {
		total += p.Segment(i).Length()
	}
	return total
}

// BoundingBox returns the bounding box of all vertices.
func (p Polyline) BoundingBox() BBox {
	b := EmptyBBox()
	for _, pt := range p.Points {
		b = b.Extend(pt)
	}
	return b
}

// ClosestPoint returns the nearest point on the polyline to q.
func (p Polyline) ClosestPoint(q Point) Point {
	best := math.Inf(1)
	var bestPt Point
	for i := 0; i < p.NumSegments(); i++ {
		cp, _ := p.Segment(i).ClosestPoint(q)
		if d := q.Dist(cp); d < best {
			best = d
			bestPt = cp
		}
	}
	return bestPt
}

// DistToPoint returns the minimum distance from q to the polyline.
func (p Polyline) DistToPoint(q Point) float64 {
	return q.Dist(p.ClosestPoint(q))
}

// Offset returns a new polyline offset by dist (positive = left/outward).
// Uses simple vertex offsetting; for complex polylines use a proper clipper.
func (p Polyline) Offset(dist float64) Polyline {
	n := len(p.Points)
	if n < 2 {
		return p
	}
	out := make([]Point, n)
	for i := 0; i < n; i++ {
		// Compute bisector direction at vertex i
		prev := (i - 1 + n) % n
		next := (i + 1) % n
		if !p.Closed {
			if i == 0 {
				seg := Segment{p.Points[0], p.Points[1]}
				out[0] = p.Points[0].Add(seg.Dir().Normalize().Perp().Scale(dist))
				continue
			}
			if i == n-1 {
				seg := Segment{p.Points[n-2], p.Points[n-1]}
				out[n-1] = p.Points[n-1].Add(seg.Dir().Normalize().Perp().Scale(dist))
				continue
			}
		}
		d1 := p.Points[i].Sub(p.Points[prev]).Normalize()
		d2 := p.Points[next].Sub(p.Points[i]).Normalize()
		n1 := d1.Perp()
		n2 := d2.Perp()
		bisect := n1.Add(n2).Normalize()
		// Scale by 1/cos(half-angle) to keep uniform offset
		dot := n1.Dot(bisect)
		if math.Abs(dot) < Epsilon {
			dot = Epsilon
		}
		scale := dist / dot
		out[i] = p.Points[i].Add(bisect.Scale(scale))
	}
	return Polyline{Points: out, Closed: p.Closed}
}

// TrimAt splits the polyline at a parametric t (relative to total length).
// Returns two polylines.
func (p Polyline) TrimAt(t float64) (Polyline, Polyline) {
	totalLen := p.Length()
	if totalLen < Epsilon {
		return p, Polyline{}
	}
	targetLen := t * totalLen
	accumulated := 0.0
	for i := 0; i < p.NumSegments(); i++ {
		seg := p.Segment(i)
		segLen := seg.Length()
		if accumulated+segLen >= targetLen {
			localT := (targetLen - accumulated) / segLen
			mid := seg.PointAt(localT)
			first := make([]Point, i+2)
			copy(first, p.Points[:i+1])
			first[i+1] = mid
			second := make([]Point, len(p.Points)-i)
			second[0] = mid
			copy(second[1:], p.Points[i+1:])
			return Polyline{Points: first}, Polyline{Points: second}
		}
		accumulated += segLen
	}
	return p, Polyline{}
}
