package geometry

import "math"

// Package-level intersection API.
//
// Analytical pairs (exact, within floating-point precision):
//   Segment × Segment, Segment × Line, Segment × Ray
//   Line    × Line, Line × Ray
//   Circle  × Circle, Circle × Segment, Circle × Line, Circle × Ray
//   Arc     × Arc, Arc × Circle, Arc × Segment, Arc × Line, Arc × Ray
//   Segment × Ellipse (Newton root-finder, error < 1e-9)
//
// Approximated pairs (polyline tessellation at 100 samples, error < 0.5 % of
// entity extent for smooth curves; results are not guaranteed exact):
//   Any pair involving BezierEntity or NURBSEntity
//   EllipseEntity × Circle/Arc/Polyline/Bezier/NURBS

// ─── Top-level pairwise dispatcher ───────────────────────────────────────────

// Intersect returns the intersection points of any two entities.
// Analytical methods are used where available; spline and ellipse combinations
// fall back to polyline approximation (100 samples).  See package-level comment
// for the full analytical-vs-approximated matrix and tolerance contracts.
func Intersect(a, b Entity) []Point {
        // Normalise order so dispatch is symmetric.
        pts := intersectOrdered(a, b)
        if pts != nil {
                return pts
        }
        // Try swapped order for commutative pairs.
        return intersectOrdered(b, a)
}

func intersectOrdered(a, b Entity) []Point {
        switch av := a.(type) {
        case SegmentEntity:
                return intersectSegmentWith(av.Segment, b)
        case LineEntity:
                return intersectLineWith(av.Line, b)
        case RayEntity:
                return intersectRayWith(av.Ray, b)
        case CircleEntity:
                return intersectCircleWith(av.Circle, b)
        case ArcEntity:
                return intersectArcWith(av.Arc, b)
        case EllipseEntity:
                return intersectEllipseWith(av.Ellipse, b)
        case PolylineEntity:
                return intersectPolylineWith(av.Polyline, b)
        case BezierEntity:
                return intersectSplineApproxWith(av.BezierSpline.ApproxPolyline(100), b)
        case NURBSEntity:
                return intersectSplineApproxWith(av.NURBSSpline.ApproxPolyline(100), b)
        }
        return nil
}

// ─── Per-type dispatchers ─────────────────────────────────────────────────────

func intersectSegmentWith(s Segment, b Entity) []Point {
        switch bv := b.(type) {
        case SegmentEntity:
                return IntersectSegments(s, bv.Segment)
        case LineEntity:
                l := bv.Line
                pts := IntersectLines(Line{s.Start, s.End}, l)
                return filterBySegment(s, pts)
        case RayEntity:
                return bv.Ray.IntersectWithSegment(s)
        case CircleEntity:
                return IntersectSegmentCircle(s, bv.Circle)
        case ArcEntity:
                return IntersectSegmentArc(s, bv.Arc)
        case EllipseEntity:
                return IntersectSegmentEllipse(s, bv.Ellipse)
        case PolylineEntity:
                return IntersectSegmentPolyline(s, bv.Polyline)
        case BezierEntity:
                return IntersectSegmentBezier(s, bv.BezierSpline)
        case NURBSEntity:
                return IntersectSegmentPolyline(s, Polyline{Points: bv.NURBSSpline.ApproxPolyline(100)})
        }
        return nil
}

func intersectLineWith(l Line, b Entity) []Point {
        switch bv := b.(type) {
        case SegmentEntity:
                pts := IntersectLines(l, Line{bv.Start, bv.End})
                return filterBySegment(bv.Segment, pts)
        case LineEntity:
                return IntersectLines(l, bv.Line)
        case RayEntity:
                return bv.Ray.IntersectWithLine(l)
        case CircleEntity:
                return IntersectLineCircle(l, bv.Circle)
        case ArcEntity:
                pts := IntersectLineCircle(l, Circle{bv.Center, bv.Radius})
                return filterByArc(bv.Arc, pts)
        case EllipseEntity:
                approx := Polyline{Points: bv.Ellipse.ApproxPolyline(200)}
                var result []Point
                for i := 0; i < approx.NumSegments(); i++ {
                        seg := approx.Segment(i)
                        pts := IntersectLines(l, Line{seg.Start, seg.End})
                        result = appendUnique(result, filterBySegment(seg, pts)...)
                }
                return result
        case PolylineEntity:
                var result []Point
                for i := 0; i < bv.Polyline.NumSegments(); i++ {
                        seg := bv.Polyline.Segment(i)
                        pts := IntersectLines(l, Line{seg.Start, seg.End})
                        result = appendUnique(result, filterBySegment(seg, pts)...)
                }
                return result
        case BezierEntity:
                return intersectLineWith(l, PolylineEntity{Polyline{Points: bv.BezierSpline.ApproxPolyline(100)}})
        case NURBSEntity:
                return intersectLineWith(l, PolylineEntity{Polyline{Points: bv.NURBSSpline.ApproxPolyline(100)}})
        }
        return nil
}

func intersectRayWith(r Ray, b Entity) []Point {
        switch bv := b.(type) {
        case SegmentEntity:
                return r.IntersectWithSegment(bv.Segment)
        case LineEntity:
                return r.IntersectWithLine(bv.Line)
        case RayEntity:
                // Ray vs ray: treat as two half-lines
                pts := r.IntersectWithLine(Line{bv.Origin, bv.Origin.Add(bv.Dir)})
                var out []Point
                d := bv.Direction()
                for _, p := range pts {
                        if p.Sub(bv.Origin).Dot(d) >= -Epsilon {
                                out = append(out, p)
                        }
                }
                return out
        case CircleEntity:
                return r.IntersectWithCircle(bv.Circle)
        case ArcEntity:
                pts := r.IntersectWithCircle(Circle{bv.Center, bv.Radius})
                return filterByArc(bv.Arc, pts)
        case EllipseEntity:
                return intersectRayWith(r, PolylineEntity{Polyline{Points: bv.Ellipse.ApproxPolyline(200)}})
        case PolylineEntity:
                var result []Point
                for i := 0; i < bv.Polyline.NumSegments(); i++ {
                        pts := r.IntersectWithSegment(bv.Polyline.Segment(i))
                        result = appendUnique(result, pts...)
                }
                return result
        case BezierEntity:
                return intersectRayWith(r, PolylineEntity{Polyline{Points: bv.BezierSpline.ApproxPolyline(100)}})
        case NURBSEntity:
                return intersectRayWith(r, PolylineEntity{Polyline{Points: bv.NURBSSpline.ApproxPolyline(100)}})
        }
        return nil
}

func intersectCircleWith(c Circle, b Entity) []Point {
        switch bv := b.(type) {
        case SegmentEntity:
                return IntersectSegmentCircle(bv.Segment, c)
        case LineEntity:
                return IntersectLineCircle(bv.Line, c)
        case RayEntity:
                return bv.Ray.IntersectWithCircle(c)
        case CircleEntity:
                return IntersectCircles(c, bv.Circle)
        case ArcEntity:
                return IntersectCircleArc(c, bv.Arc)
        case EllipseEntity:
                return IntersectCircleEllipse(c, bv.Ellipse)
        case PolylineEntity:
                return IntersectCirclePolyline(c, bv.Polyline)
        case BezierEntity:
                return IntersectCirclePolyline(c, Polyline{Points: bv.BezierSpline.ApproxPolyline(100)})
        case NURBSEntity:
                return IntersectCirclePolyline(c, Polyline{Points: bv.NURBSSpline.ApproxPolyline(100)})
        }
        return nil
}

func intersectArcWith(a Arc, b Entity) []Point {
        switch bv := b.(type) {
        case SegmentEntity:
                return IntersectSegmentArc(bv.Segment, a)
        case LineEntity:
                pts := IntersectLineCircle(bv.Line, Circle{a.Center, a.Radius})
                return filterByArc(a, pts)
        case RayEntity:
                pts := bv.Ray.IntersectWithCircle(Circle{a.Center, a.Radius})
                return filterByArc(a, pts)
        case CircleEntity:
                return IntersectCircleArc(bv.Circle, a)
        case ArcEntity:
                return IntersectArcs(a, bv.Arc)
        case EllipseEntity:
                return IntersectArcEllipse(a, bv.Ellipse)
        case PolylineEntity:
                return IntersectArcPolyline(a, bv.Polyline)
        case BezierEntity:
                return IntersectArcPolyline(a, Polyline{Points: bv.BezierSpline.ApproxPolyline(100)})
        case NURBSEntity:
                return IntersectArcPolyline(a, Polyline{Points: bv.NURBSSpline.ApproxPolyline(100)})
        }
        return nil
}

func intersectEllipseWith(e Ellipse, b Entity) []Point {
        switch bv := b.(type) {
        case SegmentEntity:
                return IntersectSegmentEllipse(bv.Segment, e)
        case CircleEntity:
                return IntersectCircleEllipse(bv.Circle, e)
        case ArcEntity:
                return IntersectArcEllipse(bv.Arc, e)
        case EllipseEntity:
                return IntersectEllipses(e, bv.Ellipse)
        case PolylineEntity:
                approx := Polyline{Points: e.ApproxPolyline(200)}
                return IntersectPolylines(approx, bv.Polyline)
        case BezierEntity:
                approx := Polyline{Points: e.ApproxPolyline(200)}
                return IntersectPolylines(approx, Polyline{Points: bv.BezierSpline.ApproxPolyline(100)})
        case NURBSEntity:
                approx := Polyline{Points: e.ApproxPolyline(200)}
                return IntersectPolylines(approx, Polyline{Points: bv.NURBSSpline.ApproxPolyline(100)})
        }
        return nil
}

func intersectPolylineWith(p Polyline, b Entity) []Point {
        switch bv := b.(type) {
        case SegmentEntity:
                return IntersectSegmentPolyline(bv.Segment, p)
        case LineEntity:
                return intersectLineWith(bv.Line, PolylineEntity{p})
        case RayEntity:
                return intersectRayWith(bv.Ray, PolylineEntity{p})
        case CircleEntity:
                return IntersectCirclePolyline(bv.Circle, p)
        case ArcEntity:
                return IntersectArcPolyline(bv.Arc, p)
        case EllipseEntity:
                approx := Polyline{Points: bv.Ellipse.ApproxPolyline(200)}
                return IntersectPolylines(p, approx)
        case PolylineEntity:
                return IntersectPolylines(p, bv.Polyline)
        case BezierEntity:
                return IntersectPolylines(p, Polyline{Points: bv.BezierSpline.ApproxPolyline(100)})
        case NURBSEntity:
                return IntersectPolylines(p, Polyline{Points: bv.NURBSSpline.ApproxPolyline(100)})
        }
        return nil
}

func intersectSplineApproxWith(approxPts []Point, b Entity) []Point {
        p := Polyline{Points: approxPts}
        return intersectPolylineWith(p, b)
}

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
        // Project b endpoints onto a's parametric axis using UNCLAMPED projection
        // so that points beyond the segment ends are not incorrectly mapped to t=0 or t=1.
        da := a.End.Sub(a.Start)
        len2 := da.Len2()
        if len2 < Epsilon*Epsilon {
                return nil
        }
        ta := da.Dot(b.Start.Sub(a.Start)) / len2
        tb := da.Dot(b.End.Sub(a.Start)) / len2
        // Intersect the parametric intervals [0,1] and [min(ta,tb), max(ta,tb)]
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

// ─── Circle × Ellipse ───────────────────────────────────────────────────────

// IntersectCircleEllipse returns approximate intersection points of a circle and ellipse.
func IntersectCircleEllipse(c Circle, e Ellipse) []Point {
        return IntersectCirclePolyline(c, Polyline{Points: e.ApproxPolyline(200)})
}

// ─── Arc × Ellipse ──────────────────────────────────────────────────────────

// IntersectArcEllipse returns approximate intersection points of an arc and ellipse.
func IntersectArcEllipse(a Arc, e Ellipse) []Point {
        return IntersectArcPolyline(a, Polyline{Points: e.ApproxPolyline(200)})
}

// ─── Ellipse × Ellipse ──────────────────────────────────────────────────────

// IntersectEllipses returns approximate intersection points of two ellipses.
func IntersectEllipses(e1, e2 Ellipse) []Point {
        p1 := Polyline{Points: e1.ApproxPolyline(200)}
        p2 := Polyline{Points: e2.ApproxPolyline(200)}
        return IntersectPolylines(p1, p2)
}

// ─── Circle × Polyline ──────────────────────────────────────────────────────

// IntersectCirclePolyline returns intersection points of a circle and polyline.
func IntersectCirclePolyline(c Circle, p Polyline) []Point {
        var result []Point
        for i := 0; i < p.NumSegments(); i++ {
                result = appendUnique(result, IntersectSegmentCircle(p.Segment(i), c)...)
        }
        return result
}

// ─── Arc × Polyline ─────────────────────────────────────────────────────────

// IntersectArcPolyline returns intersection points of an arc and polyline.
func IntersectArcPolyline(a Arc, p Polyline) []Point {
        var result []Point
        for i := 0; i < p.NumSegments(); i++ {
                result = appendUnique(result, IntersectSegmentArc(p.Segment(i), a)...)
        }
        return result
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

// IntersectBezierBezier approximates intersection of two Bezier splines.
func IntersectBezierBezier(a, b BezierSpline) []Point {
        pa := Polyline{Points: a.ApproxPolyline(100)}
        pb := Polyline{Points: b.ApproxPolyline(100)}
        return IntersectPolylines(pa, pb)
}

// IntersectNURBSNURBS approximates intersection of two NURBS splines.
func IntersectNURBSNURBS(a, b NURBSSpline) []Point {
        pa := Polyline{Points: a.ApproxPolyline(100)}
        pb := Polyline{Points: b.ApproxPolyline(100)}
        return IntersectPolylines(pa, pb)
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
