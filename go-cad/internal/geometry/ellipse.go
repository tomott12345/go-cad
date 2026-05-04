package geometry

import "math"

// Ellipse is defined by a center, semi-major axis (A), semi-minor axis (B),
// and a rotation angle (degrees, CCW from positive X).
type Ellipse struct {
	Center   Point
	A, B     float64 // semi-major and semi-minor axes
	Rotation float64 // degrees
}

// NewEllipse constructs an Ellipse.
func NewEllipse(cx, cy, a, b, rotDeg float64) Ellipse {
	return Ellipse{Center: Point{cx, cy}, A: a, B: b, Rotation: rotDeg}
}

// PointAt returns the point at angle θ (radians) on the ellipse (before rotation).
func (e Ellipse) PointAt(theta float64) Point {
	local := Point{e.A * math.Cos(theta), e.B * math.Sin(theta)}
	rot := e.Rotation * math.Pi / 180
	cos, sin := math.Cos(rot), math.Sin(rot)
	return Point{
		e.Center.X + local.X*cos - local.Y*sin,
		e.Center.Y + local.X*sin + local.Y*cos,
	}
}

// Circumference approximates the ellipse perimeter using Ramanujan's formula.
func (e Ellipse) Circumference() float64 {
	h := (e.A - e.B) * (e.A - e.B) / ((e.A + e.B) * (e.A + e.B))
	return math.Pi * (e.A + e.B) * (1 + 3*h/(10+math.Sqrt(4-3*h)))
}

// BoundingBox returns an axis-aligned bounding box of the ellipse.
func (e Ellipse) BoundingBox() BBox {
	rot := e.Rotation * math.Pi / 180
	cos, sin := math.Cos(rot), math.Sin(rot)
	// Extremes of parametric ellipse after rotation
	tx := math.Atan2(-e.B*sin, e.A*cos)
	ty := math.Atan2(e.B*cos, e.A*sin)
	pts := make([]Point, 0, 8)
	for _, t := range []float64{tx, tx + math.Pi, ty, ty + math.Pi} {
		pts = append(pts, e.PointAt(t))
	}
	b := EmptyBBox()
	for _, p := range pts {
		b = b.Extend(p)
	}
	return b
}

// ApproxPolyline returns a polyline approximation of the ellipse with n segments.
func (e Ellipse) ApproxPolyline(n int) []Point {
	pts := make([]Point, n+1)
	for i := 0; i <= n; i++ {
		theta := 2 * math.Pi * float64(i) / float64(n)
		pts[i] = e.PointAt(theta)
	}
	return pts
}

// ClosestPoint returns the nearest point on the ellipse boundary to p (numerical approximation).
func (e Ellipse) ClosestPoint(p Point) Point {
	// Transform p to ellipse-local frame
	rot := -e.Rotation * math.Pi / 180
	local := p.Sub(e.Center).Rotate(rot)
	// Binary search angle that minimises distance
	bestAngle, bestDist := 0.0, math.Inf(1)
	const steps = 360
	for i := 0; i < steps; i++ {
		theta := 2 * math.Pi * float64(i) / steps
		candidate := Point{e.A * math.Cos(theta), e.B * math.Sin(theta)}
		if d := local.Dist(candidate); d < bestDist {
			bestDist = d
			bestAngle = theta
		}
	}
	// Refine with golden-section search around bestAngle
	lo, hi := bestAngle-2*math.Pi/steps, bestAngle+2*math.Pi/steps
	for range [50]struct{}{} {
		mid1 := lo + (hi-lo)/3
		mid2 := hi - (hi-lo)/3
		c1 := Point{e.A * math.Cos(mid1), e.B * math.Sin(mid1)}
		c2 := Point{e.A * math.Cos(mid2), e.B * math.Sin(mid2)}
		if local.Dist(c1) < local.Dist(c2) {
			hi = mid2
		} else {
			lo = mid1
		}
	}
	theta := (lo + hi) / 2
	localPt := Point{e.A * math.Cos(theta), e.B * math.Sin(theta)}
	// Rotate back
	rot2 := e.Rotation * math.Pi / 180
	return e.Center.Add(localPt.Rotate(rot2))
}

// DistToPoint returns the distance from p to the nearest point on the ellipse.
func (e Ellipse) DistToPoint(p Point) float64 {
	return p.Dist(e.ClosestPoint(p))
}

// Offset returns an approximated offset ellipse (scales both axes).
// For a true parallel curve use the approximation only.
func (e Ellipse) Offset(dist float64) Ellipse {
	return Ellipse{e.Center, e.A + dist, e.B + dist, e.Rotation}
}
