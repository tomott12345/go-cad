package geometry

import "math"

// Arc is a circular arc defined by center, radius, and start/end angles in degrees.
// The arc goes counter-clockwise from StartDeg to EndDeg.
type Arc struct {
	Center   Point
	Radius   float64
	StartDeg float64
	EndDeg   float64
}

// NewArc constructs an Arc.
func NewArc(cx, cy, r, startDeg, endDeg float64) Arc {
	return Arc{Center: Point{cx, cy}, Radius: r, StartDeg: startDeg, EndDeg: endDeg}
}

// spanDeg returns the angular span (always positive, going CCW).
func (a Arc) spanDeg() float64 {
	span := a.EndDeg - a.StartDeg
	for span <= 0 {
		span += 360
	}
	return span
}

// SpanRad returns the angular span in radians.
func (a Arc) SpanRad() float64 { return a.spanDeg() * math.Pi / 180 }

// Length returns the arc length.
func (a Arc) Length() float64 { return a.SpanRad() * a.Radius }

// StartPoint returns the starting point of the arc.
func (a Arc) StartPoint() Point {
	theta := a.StartDeg * math.Pi / 180
	return Point{a.Center.X + a.Radius*math.Cos(theta), a.Center.Y + a.Radius*math.Sin(theta)}
}

// EndPoint returns the ending point of the arc.
func (a Arc) EndPoint() Point {
	theta := a.EndDeg * math.Pi / 180
	return Point{a.Center.X + a.Radius*math.Cos(theta), a.Center.Y + a.Radius*math.Sin(theta)}
}

// PointAt returns the point at parametric t ∈ [0,1] along the arc.
func (a Arc) PointAt(t float64) Point {
	deg := a.StartDeg + t*a.spanDeg()
	theta := deg * math.Pi / 180
	return Point{a.Center.X + a.Radius*math.Cos(theta), a.Center.Y + a.Radius*math.Sin(theta)}
}

// containsAngleDeg reports whether angle (degrees) is within [StartDeg, EndDeg] going CCW.
func (a Arc) containsAngleDeg(deg float64) bool {
	// Normalise deg to [StartDeg, StartDeg+span]
	span := a.spanDeg()
	d := normAngle(deg - a.StartDeg)
	return d >= -Epsilon && d <= span+Epsilon
}

// normAngle normalises an angle (degrees) to [0, 360).
func normAngle(deg float64) float64 {
	for deg < 0 {
		deg += 360
	}
	for deg >= 360 {
		deg -= 360
	}
	return deg
}

// BoundingBox returns the tight bounding box of the arc.
func (a Arc) BoundingBox() BBox {
	b := EmptyBBox().Extend(a.StartPoint()).Extend(a.EndPoint())
	// Check if axis-aligned extremes are within the arc
	for _, deg := range []float64{0, 90, 180, 270} {
		if a.containsAngleDeg(deg) {
			theta := deg * math.Pi / 180
			b = b.Extend(Point{
				a.Center.X + a.Radius*math.Cos(theta),
				a.Center.Y + a.Radius*math.Sin(theta),
			})
		}
	}
	return b
}

// ClosestPoint returns the nearest point on the arc boundary to p.
func (a Arc) ClosestPoint(p Point) Point {
	// Angle from center to p
	angle := math.Atan2(p.Y-a.Center.Y, p.X-a.Center.X) * 180 / math.Pi
	if a.containsAngleDeg(angle) {
		// The closest point is directly radially
		d := p.Sub(a.Center)
		l := d.Len()
		if l < Epsilon {
			return a.StartPoint()
		}
		return a.Center.Add(d.Normalize().Scale(a.Radius))
	}
	// Closest is either start or end
	sp, ep := a.StartPoint(), a.EndPoint()
	if p.Dist(sp) <= p.Dist(ep) {
		return sp
	}
	return ep
}

// DistToPoint returns the distance from p to the nearest point on the arc.
func (a Arc) DistToPoint(p Point) float64 {
	return p.Dist(a.ClosestPoint(p))
}

// TrimAt splits the arc at parametric t, returning two arcs.
func (a Arc) TrimAt(t float64) (Arc, Arc) {
	midDeg := a.StartDeg + t*a.spanDeg()
	return Arc{a.Center, a.Radius, a.StartDeg, midDeg},
		Arc{a.Center, a.Radius, midDeg, a.EndDeg}
}

// Offset returns a concentric arc offset by dist (positive = outward).
func (a Arc) Offset(dist float64) Arc {
	return Arc{a.Center, a.Radius + dist, a.StartDeg, a.EndDeg}
}

// Midpoint returns the midpoint of the arc.
func (a Arc) Midpoint() Point { return a.PointAt(0.5) }
