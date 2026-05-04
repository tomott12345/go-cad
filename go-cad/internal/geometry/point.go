// Package geometry provides analytical 2D geometric primitives and operations
// for the go-cad engine. All types are JSON-serialisable.
package geometry

import (
	"math"
)

const Epsilon = 1e-10

// Point is a 2D point (or vector).
type Point struct {
	X, Y float64
}

// Add returns p + q.
func (p Point) Add(q Point) Point { return Point{p.X + q.X, p.Y + q.Y} }

// Sub returns p - q.
func (p Point) Sub(q Point) Point { return Point{p.X - q.X, p.Y - q.Y} }

// Scale returns p * s.
func (p Point) Scale(s float64) Point { return Point{p.X * s, p.Y * s} }

// Dot returns the dot product p · q.
func (p Point) Dot(q Point) float64 { return p.X*q.X + p.Y*q.Y }

// Cross returns the 2D cross product (scalar) p × q.
func (p Point) Cross(q Point) float64 { return p.X*q.Y - p.Y*q.X }

// Len returns the Euclidean length of the vector.
func (p Point) Len() float64 { return math.Hypot(p.X, p.Y) }

// Len2 returns the squared length.
func (p Point) Len2() float64 { return p.X*p.X + p.Y*p.Y }

// Dist returns the distance between p and q.
func (p Point) Dist(q Point) float64 { return p.Sub(q).Len() }

// Normalize returns the unit vector in the direction of p.
// Returns the zero vector if p has zero length.
func (p Point) Normalize() Point {
	l := p.Len()
	if l < Epsilon {
		return Point{}
	}
	return Point{p.X / l, p.Y / l}
}

// Perp returns the vector perpendicular to p (rotated 90° CCW).
func (p Point) Perp() Point { return Point{-p.Y, p.X} }

// Lerp returns the linear interpolation between p and q at t.
func (p Point) Lerp(q Point, t float64) Point {
	return Point{p.X + (q.X-p.X)*t, p.Y + (q.Y-p.Y)*t}
}

// Near reports whether p and q are within Epsilon of each other.
func (p Point) Near(q Point) bool { return p.Dist(q) < Epsilon }

// AngleTo returns the angle in radians from p to q, measured from the positive X axis.
func (p Point) AngleTo(q Point) float64 {
	return math.Atan2(q.Y-p.Y, q.X-p.X)
}

// Rotate returns p rotated by angle radians around the origin.
func (p Point) Rotate(angle float64) Point {
	cos, sin := math.Cos(angle), math.Sin(angle)
	return Point{p.X*cos - p.Y*sin, p.X*sin + p.Y*cos}
}

// RotateAround returns p rotated by angle radians around the pivot.
func (p Point) RotateAround(pivot Point, angle float64) Point {
	return p.Sub(pivot).Rotate(angle).Add(pivot)
}
