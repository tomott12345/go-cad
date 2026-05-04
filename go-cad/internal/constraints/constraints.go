// Package constraints provides a parametric constraint solver for 2D CAD geometry.
// It uses iterative propagation (analogous to a simple Gauss-Seidel relaxation)
// to resolve constraints between geometry entities represented as mutable handles.
package constraints

import (
        "math"

        "go-cad/internal/geometry"
)

// ─── Constraint types ─────────────────────────────────────────────────────────

// ConstraintKind identifies the type of constraint.
type ConstraintKind string

const (
        Coincident    ConstraintKind = "coincident"
        Horizontal    ConstraintKind = "horizontal"
        Vertical      ConstraintKind = "vertical"
        Parallel      ConstraintKind = "parallel"
        Perpendicular ConstraintKind = "perpendicular"
        EqualLength   ConstraintKind = "equal_length"
        EqualRadius   ConstraintKind = "equal_radius"
        Tangent       ConstraintKind = "tangent"
        Fixed         ConstraintKind = "fixed"
        Midpoint      ConstraintKind = "midpoint"
        Symmetric     ConstraintKind = "symmetric"
)

// Constraint is the interface all constraints implement.
type Constraint interface {
        // Kind returns the constraint type.
        Kind() ConstraintKind
        // Entities returns the indices (into the solve set) this constraint acts on.
        Entities() []int
        // Error returns a scalar error value (0 = satisfied).
        Error(pts []*geometry.Point) float64
        // Apply performs one corrective step, mutating pts toward satisfying the constraint.
        Apply(pts []*geometry.Point)
}

// ─── Point-handle system ──────────────────────────────────────────────────────

// Handle is a mutable 2D point used as an anchor for constraints.
// Multiple constraints can reference the same Handle pointer.
type Handle = geometry.Point

// ─── Constraint implementations ───────────────────────────────────────────────

// CoincidentConstraint forces two points to be at the same location.
type CoincidentConstraint struct {
        A, B int // indices into the handle slice
}

func (c CoincidentConstraint) Kind() ConstraintKind { return Coincident }
func (c CoincidentConstraint) Entities() []int      { return []int{c.A, c.B} }
func (c CoincidentConstraint) Error(pts []*geometry.Point) float64 {
        return pts[c.A].Dist(*pts[c.B])
}
func (c CoincidentConstraint) Apply(pts []*geometry.Point) {
        mid := pts[c.A].Add(*pts[c.B]).Scale(0.5)
        *pts[c.A] = mid
        *pts[c.B] = mid
}

// HorizontalConstraint forces a segment (start=A, end=B) to be horizontal.
type HorizontalConstraint struct {
        A, B int
}

func (c HorizontalConstraint) Kind() ConstraintKind { return Horizontal }
func (c HorizontalConstraint) Entities() []int      { return []int{c.A, c.B} }
func (c HorizontalConstraint) Error(pts []*geometry.Point) float64 {
        return math.Abs(pts[c.A].Y - pts[c.B].Y)
}
func (c HorizontalConstraint) Apply(pts []*geometry.Point) {
        midY := (pts[c.A].Y + pts[c.B].Y) / 2
        pts[c.A].Y = midY
        pts[c.B].Y = midY
}

// VerticalConstraint forces a segment (start=A, end=B) to be vertical.
type VerticalConstraint struct {
        A, B int
}

func (c VerticalConstraint) Kind() ConstraintKind { return Vertical }
func (c VerticalConstraint) Entities() []int      { return []int{c.A, c.B} }
func (c VerticalConstraint) Error(pts []*geometry.Point) float64 {
        return math.Abs(pts[c.A].X - pts[c.B].X)
}
func (c VerticalConstraint) Apply(pts []*geometry.Point) {
        midX := (pts[c.A].X + pts[c.B].X) / 2
        pts[c.A].X = midX
        pts[c.B].X = midX
}

// ParallelConstraint forces two segments (A1→A2) and (B1→B2) to be parallel.
type ParallelConstraint struct {
        A1, A2, B1, B2 int
}

func (c ParallelConstraint) Kind() ConstraintKind { return Parallel }
func (c ParallelConstraint) Entities() []int      { return []int{c.A1, c.A2, c.B1, c.B2} }
func (c ParallelConstraint) Error(pts []*geometry.Point) float64 {
        da := pts[c.A2].Sub(*pts[c.A1]).Normalize()
        db := pts[c.B2].Sub(*pts[c.B1]).Normalize()
        return math.Abs(da.Cross(db))
}
func (c ParallelConstraint) Apply(pts []*geometry.Point) {
        da := pts[c.A2].Sub(*pts[c.A1]).Normalize()
        // Rotate B to be parallel to A
        db := pts[c.B2].Sub(*pts[c.B1])
        length := db.Len()
        if length < geometry.Epsilon {
                return
        }
        newEnd := pts[c.B1].Add(da.Scale(length))
        *pts[c.B2] = newEnd
}

// PerpendicularConstraint forces two segments to be perpendicular.
type PerpendicularConstraint struct {
        A1, A2, B1, B2 int
}

func (c PerpendicularConstraint) Kind() ConstraintKind { return Perpendicular }
func (c PerpendicularConstraint) Entities() []int      { return []int{c.A1, c.A2, c.B1, c.B2} }
func (c PerpendicularConstraint) Error(pts []*geometry.Point) float64 {
        da := pts[c.A2].Sub(*pts[c.A1]).Normalize()
        db := pts[c.B2].Sub(*pts[c.B1]).Normalize()
        return math.Abs(da.Dot(db))
}
func (c PerpendicularConstraint) Apply(pts []*geometry.Point) {
        da := pts[c.A2].Sub(*pts[c.A1]).Normalize()
        perp := da.Perp()
        db := pts[c.B2].Sub(*pts[c.B1])
        length := db.Len()
        if length < geometry.Epsilon {
                return
        }
        *pts[c.B2] = pts[c.B1].Add(perp.Scale(length))
}

// EqualLengthConstraint forces two segments to have the same length.
type EqualLengthConstraint struct {
        A1, A2, B1, B2 int
}

func (c EqualLengthConstraint) Kind() ConstraintKind { return EqualLength }
func (c EqualLengthConstraint) Entities() []int      { return []int{c.A1, c.A2, c.B1, c.B2} }
func (c EqualLengthConstraint) Error(pts []*geometry.Point) float64 {
        la := pts[c.A1].Dist(*pts[c.A2])
        lb := pts[c.B1].Dist(*pts[c.B2])
        return math.Abs(la - lb)
}
func (c EqualLengthConstraint) Apply(pts []*geometry.Point) {
        la := pts[c.A1].Dist(*pts[c.A2])
        db := pts[c.B2].Sub(*pts[c.B1])
        lb := db.Len()
        if lb < geometry.Epsilon {
                return
        }
        target := (la + lb) / 2
        // Adjust B to have target length, keep B1 fixed
        *pts[c.B2] = pts[c.B1].Add(db.Normalize().Scale(target))
        // Adjust A to have target length, keep A1 fixed
        da := pts[c.A2].Sub(*pts[c.A1])
        *pts[c.A2] = pts[c.A1].Add(da.Normalize().Scale(target))
}

// FixedConstraint pins a point to a specific location.
type FixedConstraint struct {
        Index    int
        Position geometry.Point
}

func (c FixedConstraint) Kind() ConstraintKind { return Fixed }
func (c FixedConstraint) Entities() []int      { return []int{c.Index} }
func (c FixedConstraint) Error(pts []*geometry.Point) float64 {
        return pts[c.Index].Dist(c.Position)
}
func (c FixedConstraint) Apply(pts []*geometry.Point) {
        *pts[c.Index] = c.Position
}

// MidpointConstraint forces point M to be the midpoint of A and B.
type MidpointConstraint struct {
        A, B, M int
}

func (c MidpointConstraint) Kind() ConstraintKind { return Midpoint }
func (c MidpointConstraint) Entities() []int      { return []int{c.A, c.B, c.M} }
func (c MidpointConstraint) Error(pts []*geometry.Point) float64 {
        mid := pts[c.A].Add(*pts[c.B]).Scale(0.5)
        return pts[c.M].Dist(mid)
}
func (c MidpointConstraint) Apply(pts []*geometry.Point) {
        mid := pts[c.A].Add(*pts[c.B]).Scale(0.5)
        *pts[c.M] = mid
}

// TangentConstraint forces a line (A→B) to be tangent to a circle at point T.
// The circle center index is C, radius stored separately.
type TangentCircleConstraint struct {
        LineA, LineB int
        CircleCenter int
        Radius       float64
}

func (c TangentCircleConstraint) Kind() ConstraintKind { return Tangent }
func (c TangentCircleConstraint) Entities() []int {
        return []int{c.LineA, c.LineB, c.CircleCenter}
}
func (c TangentCircleConstraint) Error(pts []*geometry.Point) float64 {
        l := geometry.Line{P: *pts[c.LineA], Q: *pts[c.LineB]}
        dist := math.Abs(l.DistToPoint(*pts[c.CircleCenter]))
        return math.Abs(dist - c.Radius)
}
func (c TangentCircleConstraint) Apply(pts []*geometry.Point) {
        l := geometry.Line{P: *pts[c.LineA], Q: *pts[c.LineB]}
        // Move line to be tangent: translate by the difference
        dist := l.DistToPoint(*pts[c.CircleCenter])
        correction := (c.Radius - math.Abs(dist)) / 2
        normal := l.Normal()
        if dist < 0 {
                normal = normal.Scale(-1)
        }
        delta := normal.Scale(correction)
        *pts[c.LineA] = pts[c.LineA].Add(delta)
        *pts[c.LineB] = pts[c.LineB].Add(delta)
}

// ─── Solver ──────────────────────────────────────────────────────────────────

const (
        defaultMaxIter = 200
        defaultTol     = 1e-8
)

// SolveResult holds the result of a constraint solve.
type SolveResult struct {
        Converged  bool
        Iterations int
        FinalError float64
}

// totalError sums Error() over all constraints (measured AFTER applying them all).
func totalError(pts []*geometry.Point, constraints []Constraint) float64 {
        e := 0.0
        for _, c := range constraints {
                e += c.Error(pts)
        }
        return e
}

// Solve runs the iterative constraint solver on the provided point handles,
// applying all constraints until the total error falls below tol or maxIter
// iterations are reached.
//
// pts is a slice of *geometry.Point — each constraint holds indices into this slice.
// The solver mutates the points in-place.
//
// Fixed constraints are re-applied at the end of every pass so that pinned
// points cannot be moved by subsequently executed constraints.
func Solve(pts []*geometry.Point, constraints []Constraint, maxIter int, tol float64) SolveResult {
        if maxIter <= 0 {
                maxIter = defaultMaxIter
        }
        if tol <= 0 {
                tol = defaultTol
        }

        // Separate Fixed from general constraints so we can enforce pins last.
        var fixed, general []Constraint
        for _, c := range constraints {
                if c.Kind() == Fixed {
                        fixed = append(fixed, c)
                } else {
                        general = append(general, c)
                }
        }

        for iter := 0; iter < maxIter; iter++ {
                // Apply general constraints (Gauss-Seidel pass)
                for _, c := range general {
                        c.Apply(pts)
                }
                // Re-enforce fixed pins so nothing can drag them away
                for _, c := range fixed {
                        c.Apply(pts)
                }
                // Compute total error AFTER all constraints have been applied
                err := totalError(pts, constraints)
                if err < tol {
                        return SolveResult{Converged: true, Iterations: iter + 1, FinalError: err}
                }
        }
        return SolveResult{Converged: false, Iterations: maxIter, FinalError: totalError(pts, constraints)}
}

// SolveDefault calls Solve with default iteration count and tolerance.
func SolveDefault(pts []*geometry.Point, constraints []Constraint) SolveResult {
        return Solve(pts, constraints, defaultMaxIter, defaultTol)
}

// ─── Symmetric constraint ─────────────────────────────────────────────────────

// SymmetricConstraint forces points A and B to be symmetric about the axis A1→A2.
type SymmetricConstraint struct {
        A, B       int // points to symmetrize
        Axis1, Axis2 int // axis line
}

func (c SymmetricConstraint) Kind() ConstraintKind { return Symmetric }
func (c SymmetricConstraint) Entities() []int      { return []int{c.A, c.B, c.Axis1, c.Axis2} }
func (c SymmetricConstraint) Error(pts []*geometry.Point) float64 {
        l := geometry.Line{P: *pts[c.Axis1], Q: *pts[c.Axis2]}
        mirrorA := mirrorPoint(*pts[c.A], l)
        return mirrorA.Dist(*pts[c.B])
}
func (c SymmetricConstraint) Apply(pts []*geometry.Point) {
        l := geometry.Line{P: *pts[c.Axis1], Q: *pts[c.Axis2]}
        mirrorA := mirrorPoint(*pts[c.A], l)
        mid := mirrorA.Add(*pts[c.B]).Scale(0.5)
        *pts[c.B] = mid
        // Mirror back to fix A
        *pts[c.A] = mirrorPoint(mid, l)
}

// mirrorPoint reflects p across line l.
func mirrorPoint(p geometry.Point, l geometry.Line) geometry.Point {
        foot := l.ClosestPoint(p)
        return foot.Scale(2).Sub(p)
}
