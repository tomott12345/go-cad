package geometry

import "math"

// Intersect computes intersection points between two geometric entities.
// Returns a (possibly empty) slice of Points.

// ─── Segment × Segment ──────────────────────────────────────────────────────

// IntersectSegments returns the intersection point(s) of two line segments.
// Returns nil if they do not intersect; one point if they cross; two if collinear overlap.
func IntersectSegments(a, b Segment) []Point {
        da := a.End.Sub(a.Start)
        db := b.End.Sub(b.Start)
        denom := da.Cross(db)

        if math.Abs(denom) < Epsilon {
                // Parallel or collinear
                return intersectCollinearSegments(a, b)
        }
        diff := b.Start.Sub(a.Start)
        t := diff.Cross(db) / denom
        u := diff.Cross(da) / denom
        if t >= -Epsilon && t <= 1+Epsilon && u >= -Epsilon && u <= 1+Epsilon {
                return []Point{a.PointAt(clamp01(t))}
        }
        return nil
}

func intersectCollinearSegments(a, b Segment) []Point {
        // First verify they are truly collinear (not just parallel)
        l := Line{a.Start, a.End}
        if math.Abs(l.DistToPoint(b.Start)) > Epsilon || math.Abs(l.DistToPoint(b.End)) > Epsilon {
                return nil // parallel but not collinear
        }
        // Project b endpoints onto a's parametric axis
        _, ta := a.ClosestPoint(b.Start)
        _, tb := a.ClosestPoint(b.End)
        // Overlap test
        lo := math.Max(math.Min(ta, tb), 0)
        hi := math.Min(math.Max(ta, tb), 1)
        if hi < lo-Epsilon {
                return nil
        }
        if math.Abs(hi-lo) < Epsilon {
                return []Point{a.PointAt((lo + hi) / 2)}
        }
        return []Point{a.PointAt(lo), a.PointAt(hi)}
}

// IntersectLines returns the intersection of two infinite lines.
// Returns nil if parallel.
func IntersectLines(a, b Line) []Point {
        da := a.Dir()
        db := b.Dir()
        denom := da.Cross(db)
        if math.Abs(denom) < Epsilon {
                return nil
        }
        diff := b.P.Sub(a.P)
        t := diff.Cross(db) / denom
        return []Point{a.P.Add(da.Scale(t))}
}

// ─── Segment × Circle ───────────────────────────────────────────────────────

// IntersectSegmentCircle returns the intersection points of a segment and circle.
func IntersectSegmentCircle(s Segment, c Circle) []Point {
        return filterBySegment(s, IntersectLineCircle(Line{s.Start, s.End}, c))
}

// IntersectLineCircle returns intersection points of an infinite line and circle.
func IntersectLineCircle(l Line, c Circle) []Point {
        d := l.Dir().Normalize()
        if d.Len() < Epsilon {
                return nil
        }
        fc := l.ClosestPoint(c.Center)
        dist := c.Center.Dist(fc)
        if dist > c.Radius+Epsilon {
                return nil
        }
        if math.Abs(dist-c.Radius) < Epsilon {
                return []Point{fc} // tangent
        }
        offset := math.Sqrt(c.Radius*c.Radius - dist*dist)
        return []Point{fc.Add(d.Scale(-offset)), fc.Add(d.Scale(offset))}
}

// ─── Segment × Arc ──────────────────────────────────────────────────────────

// IntersectSegmentArc returns intersection points of a segment and arc.
func IntersectSegmentArc(s Segment, a Arc) []Point {
        circle := Circle{a.Center, a.Radius}
        pts := IntersectSegmentCircle(s, circle)
        return filterByArc(a, pts)
}

// ─── Circle × Circle ────────────────────────────────────────────────────────

// IntersectCircles returns intersection points of two circles.
func IntersectCircles(c1, c2 Circle) []Point {
        d := c1.Center.Dist(c2.Center)
        if d < Epsilon {
                return nil // concentric
        }
        r1, r2 := c1.Radius, c2.Radius
        if d > r1+r2+Epsilon || d < math.Abs(r1-r2)-Epsilon {
                return nil // too far or one inside the other
        }
        if math.Abs(d-r1-r2) < Epsilon || math.Abs(d-math.Abs(r1-r2)) < Epsilon {
                // Tangent — one point
                dir := c2.Center.Sub(c1.Center).Normalize()
                return []Point{c1.Center.Add(dir.Scale(r1))}
        }
        a := (r1*r1 - r2*r2 + d*d) / (2 * d)
        h := math.Sqrt(r1*r1 - a*a)
        dir := c2.Center.Sub(c1.Center).Normalize()
        mid := c1.Center.Add(dir.Scale(a))
        perp := dir.Perp().Scale(h)
        return []Point{mid.Add(perp), mid.Sub(perp)}
}

// ─── Circle × Arc ───────────────────────────────────────────────────────────

// IntersectCircleArc returns intersection points of a circle and arc.
func IntersectCircleArc(c Circle, a Arc) []Point {
        circle2 := Circle{a.Center, a.Radius}
        pts := IntersectCircles(c, circle2)
        return filterByArc(a, pts)
}

// ─── Arc × Arc ──────────────────────────────────────────────────────────────

// IntersectArcs returns intersection points of two arcs.
func IntersectArcs(a1, a2 Arc) []Point {
        c1 := Circle{a1.Center, a1.Radius}
        pts := IntersectCircleArc(c1, a2)
        return filterByArc(a1, pts)
}

// ─── Segment × Ellipse (numerical) ──────────────────────────────────────────

// IntersectSegmentEllipse returns approximate intersection points of a segment and ellipse.
func IntersectSegmentEllipse(s Segment, e Ellipse) []Point {
        const steps = 200
        pts := e.ApproxPolyline(steps)
        var result []Point
        for i := 0; i < len(pts)-1; i++ {
                sub := Segment{pts[i], pts[i+1]}
                if xs := IntersectSegments(s, sub); len(xs) > 0 {
                        result = appendUnique(result, xs...)
                }
        }
        return result
}

// ─── Segment × Polyline ─────────────────────────────────────────────────────

// IntersectSegmentPolyline returns intersection points of a segment and polyline.
func IntersectSegmentPolyline(s Segment, p Polyline) []Point {
        var result []Point
        for i := 0; i < p.NumSegments(); i++ {
                if xs := IntersectSegments(s, p.Segment(i)); len(xs) > 0 {
                        result = appendUnique(result, xs...)
                }
        }
        return result
}

// ─── Polyline × Polyline ────────────────────────────────────────────────────

// IntersectPolylines returns all intersection points between two polylines.
func IntersectPolylines(a, b Polyline) []Point {
        var result []Point
        for i := 0; i < a.NumSegments(); i++ {
                for j := 0; j < b.NumSegments(); j++ {
                        if xs := IntersectSegments(a.Segment(i), b.Segment(j)); len(xs) > 0 {
                                result = appendUnique(result, xs...)
                        }
                }
        }
        return result
}

// ─── Spline intersections (numerical via polyline approximation) ─────────────

// IntersectSegmentBezier approximates intersection of a segment and Bezier spline.
func IntersectSegmentBezier(s Segment, b BezierSpline) []Point {
        poly := Polyline{Points: b.ApproxPolyline(100)}
        return IntersectSegmentPolyline(s, poly)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func clamp01(t float64) float64 {
        if t < 0 {
                return 0
        }
        if t > 1 {
                return 1
        }
        return t
}

// filterBySegment removes points that don't lie within the segment bounds.
func filterBySegment(s Segment, pts []Point) []Point {
        var out []Point
        da := s.End.Sub(s.Start)
        l2 := da.Len2()
        for _, p := range pts {
                if l2 < Epsilon*Epsilon {
                        if p.Near(s.Start) {
                                out = append(out, p)
                        }
                        continue
                }
                t := p.Sub(s.Start).Dot(da) / l2
                if t >= -Epsilon && t <= 1+Epsilon {
                        out = append(out, p)
                }
        }
        return out
}

// filterByArc removes points that don't lie on the arc.
func filterByArc(a Arc, pts []Point) []Point {
        var out []Point
        for _, p := range pts {
                angle := math.Atan2(p.Y-a.Center.Y, p.X-a.Center.X) * 180 / math.Pi
                if a.containsAngleDeg(angle) {
                        out = append(out, p)
                }
        }
        return out
}

// appendUnique appends points not already in dst (within Epsilon).
func appendUnique(dst []Point, pts ...Point) []Point {
outer:
        for _, p := range pts {
                for _, q := range dst {
                        if p.Near(q) {
                                continue outer
                        }
                }
                dst = append(dst, p)
        }
        return dst
}
