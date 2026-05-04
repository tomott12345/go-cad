package geometry

import "math"

// Ray represents a half-line starting at Origin and extending in direction Dir.
// Dir need not be normalized; methods normalize internally as needed.
type Ray struct {
        Origin Point `json:"origin"`
        Dir    Point `json:"dir"` // direction (unnormalised)
}

// NewRay constructs a Ray from an origin point and a direction vector.
func NewRay(origin, dir Point) Ray { return Ray{Origin: origin, Dir: dir} }

// NewRayThrough constructs a Ray from origin pointing toward through.
func NewRayThrough(origin, through Point) Ray {
        return Ray{Origin: origin, Dir: through.Sub(origin)}
}

// Direction returns the normalised direction of the ray.
func (r Ray) Direction() Point { return r.Dir.Normalize() }

// PointAt returns the point at distance t (in world units) along the ray.
func (r Ray) PointAt(t float64) Point {
        return r.Origin.Add(r.Direction().Scale(t))
}

// Length of a ray is infinite.
func (r Ray) Length() float64 { return math.Inf(1) }

// BoundingBox for a ray is undefined (infinite), so we return the bounding
// box that includes just the origin (callers should clip as needed).
func (r Ray) BoundingBox() BBox {
        return EmptyBBox().Extend(r.Origin)
}

// ClosestPoint returns the nearest point on the ray (from Origin onward) to p.
func (r Ray) ClosestPoint(p Point) (Point, float64) {
        d := r.Direction()
        t := p.Sub(r.Origin).Dot(d)
        if t < 0 {
                return r.Origin, 0
        }
        return r.Origin.Add(d.Scale(t)), t
}

// DistToPoint returns the minimum distance from p to the ray.
func (r Ray) DistToPoint(p Point) float64 {
        cp, _ := r.ClosestPoint(p)
        return p.Dist(cp)
}

// TrimAt splits the ray into a Segment [Origin, PointAt(t)] and a new Ray
// starting at PointAt(t) with the same direction.
// t is in world units (same as PointAt).
func (r Ray) TrimAt(t float64) (Segment, Ray) {
        if t < 0 {
                t = 0
        }
        cut := r.PointAt(t)
        return Segment{r.Origin, cut}, Ray{Origin: cut, Dir: r.Dir}
}

// IntersectWithLine returns the intersection of this ray with an infinite line,
// or nil if parallel or the intersection is behind the origin.
func (r Ray) IntersectWithLine(l Line) []Point {
        d := r.Direction()
        lb := l.Dir()
        denom := d.Cross(lb)
        if math.Abs(denom) < Epsilon {
                return nil
        }
        diff := l.P.Sub(r.Origin)
        t := diff.Cross(lb) / denom
        if t < -Epsilon {
                return nil // behind origin
        }
        return []Point{r.Origin.Add(d.Scale(t))}
}

// IntersectWithSegment returns the intersection of this ray with a segment.
func (r Ray) IntersectWithSegment(s Segment) []Point {
        pts := r.IntersectWithLine(Line{s.Start, s.End})
        if len(pts) == 0 {
                return nil
        }
        return filterBySegment(s, pts)
}

// IntersectWithCircle returns the intersection points of this ray with a circle.
func (r Ray) IntersectWithCircle(c Circle) []Point {
        // Extend ray as infinite line, then filter by t >= 0
        l := Line{r.Origin, r.Origin.Add(r.Direction())}
        allPts := IntersectLineCircle(l, c)
        var out []Point
        d := r.Direction()
        for _, p := range allPts {
                t := p.Sub(r.Origin).Dot(d)
                if t >= -Epsilon {
                        out = append(out, p)
                }
        }
        return out
}
