package geometry

import "math"

// BezierSpline is a cubic Bezier spline with N control points (N >= 4, N = 3k+1).
// It is composed of multiple cubic Bezier segments sharing endpoints.
type BezierSpline struct {
        Controls []Point
}

// NewBezierSpline constructs a BezierSpline from control points.
func NewBezierSpline(pts []Point) BezierSpline { return BezierSpline{Controls: pts} }

// NumSegments returns the number of cubic segments (each needs 4 control points,
// sharing endpoints).
func (b BezierSpline) NumSegments() int {
        n := len(b.Controls)
        if n < 4 {
                return 0
        }
        return (n - 1) / 3
}

// cubicBezier evaluates the cubic Bezier defined by p0..p3 at t ∈ [0,1].
func cubicBezier(p0, p1, p2, p3 Point, t float64) Point {
        u := 1 - t
        return p0.Scale(u*u*u).Add(p1.Scale(3*u*u*t)).Add(p2.Scale(3*u*t*t)).Add(p3.Scale(t*t*t))
}

// PointAt returns the point at global parametric t ∈ [0,1] along the spline.
func (b BezierSpline) PointAt(t float64) Point {
        ns := b.NumSegments()
        if ns == 0 {
                if len(b.Controls) > 0 {
                        return b.Controls[0]
                }
                return Point{}
        }
        if t >= 1.0 {
                p0 := b.Controls[3*(ns-1)]
                p1 := b.Controls[3*(ns-1)+1]
                p2 := b.Controls[3*(ns-1)+2]
                p3 := b.Controls[3*(ns-1)+3]
                return cubicBezier(p0, p1, p2, p3, 1.0)
        }
        seg := min(int(t*float64(ns)), ns-1)
        localT := t*float64(ns) - float64(seg)
        i := seg * 3
        return cubicBezier(b.Controls[i], b.Controls[i+1], b.Controls[i+2], b.Controls[i+3], localT)
}

// ApproxPolyline returns a polyline approximation with n points per segment.
func (b BezierSpline) ApproxPolyline(n int) []Point {
        ns := b.NumSegments()
        if ns == 0 {
                return b.Controls
        }
        pts := make([]Point, 0, ns*n+1)
        for seg := 0; seg < ns; seg++ {
                i := seg * 3
                p0, p1, p2, p3 := b.Controls[i], b.Controls[i+1], b.Controls[i+2], b.Controls[i+3]
                for k := 0; k < n; k++ {
                        t := float64(k) / float64(n)
                        pts = append(pts, cubicBezier(p0, p1, p2, p3, t))
                }
        }
        // Add final point
        last := b.Controls[len(b.Controls)-1]
        pts = append(pts, last)
        return pts
}

// BoundingBox returns the bounding box via polyline approximation.
func (b BezierSpline) BoundingBox() BBox {
        pts := b.ApproxPolyline(20)
        bb := EmptyBBox()
        for _, p := range pts {
                bb = bb.Extend(p)
        }
        return bb
}

// Length approximates the arc length using the polyline approximation.
func (b BezierSpline) Length() float64 {
        pts := b.ApproxPolyline(50)
        total := 0.0
        for i := 1; i < len(pts); i++ {
                total += pts[i].Dist(pts[i-1])
        }
        return total
}

// ClosestPoint returns the nearest point on the spline to p (numerical).
func (b BezierSpline) ClosestPoint(p Point) Point {
        pts := b.ApproxPolyline(100)
        best := math.Inf(1)
        var bestPt Point
        for i := 1; i < len(pts); i++ {
                seg := Segment{pts[i-1], pts[i]}
                cp, _ := seg.ClosestPoint(p)
                if d := p.Dist(cp); d < best {
                        best = d
                        bestPt = cp
                }
        }
        return bestPt
}

// Offset returns an offset spline approximated as a polyline.
func (b BezierSpline) Offset(dist float64) Polyline {
        pts := b.ApproxPolyline(100)
        return Polyline{Points: pts}.Offset(dist)
}

// ────────────────────────────────────────────────────────────────────────────
// NURBS (Non-Uniform Rational B-Spline)
// ────────────────────────────────────────────────────────────────────────────

// NURBSSpline represents a NURBS curve (degree 3 by default).
type NURBSSpline struct {
        Degree   int
        Knots    []float64
        Controls []Point
        Weights  []float64
}

// NewNURBSSpline constructs a NURBS spline. Weights default to 1 if nil.
func NewNURBSSpline(degree int, knots []float64, controls []Point, weights []float64) NURBSSpline {
        if weights == nil {
                weights = make([]float64, len(controls))
                for i := range weights {
                        weights[i] = 1.0
                }
        }
        return NURBSSpline{Degree: degree, Knots: knots, Controls: controls, Weights: weights}
}

// basisFunc evaluates the Cox-de Boor recursion N_{i,k}(t).
func basisFunc(i, k int, knots []float64, t float64) float64 {
        if k == 0 {
                if knots[i] <= t && t < knots[i+1] {
                        return 1
                }
                return 0
        }
        d1, d2 := knots[i+k]-knots[i], knots[i+k+1]-knots[i+1]
        var left, right float64
        if d1 > Epsilon {
                left = (t - knots[i]) / d1 * basisFunc(i, k-1, knots, t)
        }
        if d2 > Epsilon {
                right = (knots[i+k+1]-t) / d2 * basisFunc(i+1, k-1, knots, t)
        }
        return left + right
}

// PointAt evaluates the NURBS at parameter t ∈ [knots[deg], knots[n]].
func (n NURBSSpline) PointAt(t float64) Point {
        numCtrl := len(n.Controls)
        // Clamp t to domain
        lo, hi := n.Knots[n.Degree], n.Knots[numCtrl]
        if t >= hi {
                t = hi - 1e-12
        }
        if t < lo {
                t = lo
        }
        var wx, wy, w float64
        for i, cp := range n.Controls {
                b := basisFunc(i, n.Degree, n.Knots, t)
                bw := b * n.Weights[i]
                wx += bw * cp.X
                wy += bw * cp.Y
                w += bw
        }
        if math.Abs(w) < Epsilon {
                return Point{}
        }
        return Point{wx / w, wy / w}
}

// ApproxPolyline returns a polyline approximation.
func (n NURBSSpline) ApproxPolyline(samples int) []Point {
        numCtrl := len(n.Controls)
        if numCtrl == 0 || len(n.Knots) == 0 {
                return nil
        }
        lo, hi := n.Knots[n.Degree], n.Knots[numCtrl]
        pts := make([]Point, samples+1)
        for i := 0; i <= samples; i++ {
                t := lo + float64(i)/float64(samples)*(hi-lo)
                pts[i] = n.PointAt(t)
        }
        return pts
}

// BoundingBox returns the bounding box via approximation.
func (n NURBSSpline) BoundingBox() BBox {
        pts := n.ApproxPolyline(50)
        bb := EmptyBBox()
        for _, p := range pts {
                bb = bb.Extend(p)
        }
        return bb
}

// ClosestPoint returns the nearest point on the NURBS to p.
func (n NURBSSpline) ClosestPoint(p Point) Point {
        pts := n.ApproxPolyline(200)
        best := math.Inf(1)
        var bestPt Point
        for i := 1; i < len(pts); i++ {
                seg := Segment{pts[i-1], pts[i]}
                cp, _ := seg.ClosestPoint(p)
                if d := p.Dist(cp); d < best {
                        best = d
                        bestPt = cp
                }
        }
        return bestPt
}
