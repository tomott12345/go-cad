package geometry

import "math"

// Circle is defined by a center and radius.
type Circle struct {
	Center Point
	Radius float64
}

// NewCircle constructs a Circle.
func NewCircle(cx, cy, r float64) Circle {
	return Circle{Center: Point{cx, cy}, Radius: r}
}

// Area returns π r².
func (c Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }

// Circumference returns 2π r.
func (c Circle) Circumference() float64 { return 2 * math.Pi * c.Radius }

// BoundingBox returns the axis-aligned bounding box of the circle.
func (c Circle) BoundingBox() BBox {
	r := c.Radius
	return BBox{
		Min: Point{c.Center.X - r, c.Center.Y - r},
		Max: Point{c.Center.X + r, c.Center.Y + r},
	}
}

// ClosestPoint returns the nearest point on the circle boundary to p.
func (c Circle) ClosestPoint(p Point) Point {
	d := p.Sub(c.Center)
	l := d.Len()
	if l < Epsilon {
		// p is at center — return point on right side
		return Point{c.Center.X + c.Radius, c.Center.Y}
	}
	return c.Center.Add(d.Normalize().Scale(c.Radius))
}

// DistToPoint returns the distance from p to the nearest point on the circle.
func (c Circle) DistToPoint(p Point) float64 {
	return math.Abs(p.Dist(c.Center) - c.Radius)
}

// Contains reports whether point p is on the circle boundary (within Epsilon).
func (c Circle) Contains(p Point) bool {
	return math.Abs(p.Dist(c.Center)-c.Radius) < Epsilon
}

// ContainsInterior reports whether p is strictly inside the circle.
func (c Circle) ContainsInterior(p Point) bool {
	return p.Dist(c.Center) < c.Radius-Epsilon
}

// QuadrantPoints returns the four cardinal points (0°, 90°, 180°, 270°).
func (c Circle) QuadrantPoints() [4]Point {
	return [4]Point{
		{c.Center.X + c.Radius, c.Center.Y},
		{c.Center.X, c.Center.Y + c.Radius},
		{c.Center.X - c.Radius, c.Center.Y},
		{c.Center.X, c.Center.Y - c.Radius},
	}
}

// PointAt returns the point on the circle at angle θ (radians).
func (c Circle) PointAt(theta float64) Point {
	return Point{
		c.Center.X + c.Radius*math.Cos(theta),
		c.Center.Y + c.Radius*math.Sin(theta),
	}
}

// TangentPoints returns the two tangent points from external point p to the circle.
// Returns nil if p is inside or on the circle.
func (c Circle) TangentPoints(p Point) []Point {
	d := p.Dist(c.Center)
	if d < c.Radius-Epsilon {
		return nil // inside
	}
	if math.Abs(d-c.Radius) < Epsilon {
		// On circle — only one tangent point (the point itself)
		return []Point{p}
	}
	angle := math.Acos(c.Radius / d)
	baseAngle := p.AngleTo(c.Center)
	return []Point{
		c.PointAt(baseAngle + angle),
		c.PointAt(baseAngle - angle),
	}
}

// Offset returns a concentric circle offset by dist (positive = outward).
func (c Circle) Offset(dist float64) Circle {
	return Circle{Center: c.Center, Radius: c.Radius + dist}
}
