// entity_solver.go provides the entity-level constraint solving API.
// It bridges the geometry.Entity world (rich typed primitives) with the
// point-handle Gauss-Seidel solver.
package constraints

import (
	"github.com/tomott12345/go-cad/internal/geometry"
)

// EntityConstraint describes a constraint referencing named entities by index.
// Use this API when working with document-level geometry, not bare points.
type EntityConstraint struct {
	// Kind is the constraint type.
	Kind ConstraintKind
	// Indices holds the entity indices this constraint acts on.
	// Semantics depend on Kind:
	//   Coincident         [A, B]        — endpoint of A coincides with start of B
	//   Horizontal/Vertical[A]           — segment endpoints horizontal/vertical
	//   Parallel/Perp      [A, B]        — two segment entities
	//   EqualLength        [A, B]        — two segment entities equal length
	//   EqualRadius        [A, B]        — two circle/arc entities equal radius
	//   Tangent            [lineEnt, circleEnt] — line entity tangent to circle/arc
	//   Fixed              [A]           — entity fixed in place
	//   Midpoint           [A, B, M]     — M is midpoint of segment A→B
	//   Symmetric          [A, B, Ax1, Ax2] — A,B symmetric about axis entity
	Indices []int
	// FixedPosition is used by Fixed constraints to pin a specific location.
	FixedPosition *geometry.Point
}

// SolveEntities solves parametric constraints over a slice of geometry entities.
// It extracts representative anchor points from each entity (e.g. endpoints,
// center), solves the constraints with the Gauss-Seidel solver, and then
// reconstructs updated entities from the moved points.
//
// Supported entity kinds: Segment, Line, Ray, Arc, Circle, Polyline.
// Spline entities are treated as immovable.
//
// Supported constraints: Coincident, Horizontal, Vertical, Parallel,
// Perpendicular, EqualLength, EqualRadius, Tangent, Fixed, Midpoint, Symmetric.
func SolveEntities(
	entities []geometry.Entity,
	constraints []EntityConstraint,
	maxIter int,
	tol float64,
) ([]geometry.Entity, SolveResult) {
	pts, anchors := extractAnchors(entities)

	var solverConstraints []Constraint
	for _, ec := range constraints {
		if c := mapEntityConstraint(ec, anchors, pts, entities); c != nil {
			solverConstraints = append(solverConstraints, c)
		}
	}

	result := Solve(pts, solverConstraints, maxIter, tol)
	updated := reconstructEntities(entities, anchors, pts)
	return updated, result
}

// SolveEntitiesDefault calls SolveEntities with default iteration and tolerance.
func SolveEntitiesDefault(
	entities []geometry.Entity,
	constraints []EntityConstraint,
) ([]geometry.Entity, SolveResult) {
	return SolveEntities(entities, constraints, defaultMaxIter, defaultTol)
}

// entityAnchor records the point indices for a single entity.
type entityAnchor struct {
	p0, p1 int      // primary anchor point indices (-1 if unused)
	radius *float64 // pointer to the entity's radius field (circles/arcs only)
}

// extractAnchors builds the flat point slice and per-entity anchor index map.
func extractAnchors(entities []geometry.Entity) ([]*geometry.Point, []entityAnchor) {
	var pts []*geometry.Point
	anchors := make([]entityAnchor, len(entities))

	add := func(p geometry.Point) int {
		pp := new(geometry.Point)
		*pp = p
		pts = append(pts, pp)
		return len(pts) - 1
	}

	for i, e := range entities {
		switch v := e.(type) {
		case geometry.SegmentEntity:
			anchors[i] = entityAnchor{
				p0: add(v.Start),
				p1: add(v.End),
			}
		case geometry.LineEntity:
			anchors[i] = entityAnchor{
				p0: add(v.P),
				p1: add(v.Q),
			}
		case geometry.RayEntity:
			anchors[i] = entityAnchor{
				p0: add(v.Origin),
				p1: -1,
			}
		case geometry.CircleEntity:
			r := v.Radius
			anchors[i] = entityAnchor{
				p0:     add(v.Center),
				p1:     -1,
				radius: &r,
			}
		case geometry.ArcEntity:
			r := v.Radius
			anchors[i] = entityAnchor{
				p0:     add(v.Center),
				p1:     -1,
				radius: &r,
			}
		case geometry.PolylineEntity:
			if len(v.Points) > 0 {
				anchors[i] = entityAnchor{
					p0: add(v.Points[0]),
					p1: add(v.Points[len(v.Points)-1]),
				}
			} else {
				anchors[i] = entityAnchor{p0: -1, p1: -1}
			}
		default:
			anchors[i] = entityAnchor{p0: -1, p1: -1}
		}
	}
	return pts, anchors
}

// reconstructEntities builds a new entity slice reflecting moved anchor points.
func reconstructEntities(
	original []geometry.Entity,
	anchors []entityAnchor,
	pts []*geometry.Point,
) []geometry.Entity {
	out := make([]geometry.Entity, len(original))
	for i, e := range original {
		a := anchors[i]
		switch v := e.(type) {
		case geometry.SegmentEntity:
			out[i] = geometry.SegmentEntity{Segment: geometry.Segment{
				Start: *pts[a.p0],
				End:   *pts[a.p1],
			}}
		case geometry.LineEntity:
			out[i] = geometry.LineEntity{Line: geometry.Line{
				P: *pts[a.p0],
				Q: *pts[a.p1],
			}}
		case geometry.RayEntity:
			out[i] = geometry.RayEntity{Ray: geometry.Ray{
				Origin: *pts[a.p0],
				Dir:    v.Dir,
			}}
		case geometry.CircleEntity:
			out[i] = geometry.CircleEntity{Circle: geometry.Circle{
				Center: *pts[a.p0],
				Radius: *a.radius,
			}}
		case geometry.ArcEntity:
			out[i] = geometry.ArcEntity{Arc: geometry.Arc{
				Center:   *pts[a.p0],
				Radius:   *a.radius,
				StartDeg: v.StartDeg,
				EndDeg:   v.EndDeg,
			}}
		case geometry.PolylineEntity:
			if a.p0 >= 0 && a.p1 >= 0 && len(v.Points) >= 2 {
				delta := pts[a.p0].Sub(v.Points[0])
				newPts := make([]geometry.Point, len(v.Points))
				for j, p := range v.Points {
					newPts[j] = p.Add(delta)
				}
				out[i] = geometry.PolylineEntity{Polyline: geometry.Polyline{
					Points: newPts,
					Closed: v.Closed,
				}}
			} else {
				out[i] = e
			}
		default:
			out[i] = e
		}
	}
	return out
}

// mapEntityConstraint converts an EntityConstraint to a point-handle Constraint.
// Returns nil if the entity indices are invalid or the kind is unsupported.
func mapEntityConstraint(
	ec EntityConstraint,
	anchors []entityAnchor,
	pts []*geometry.Point,
	entities []geometry.Entity,
) Constraint {
	idx := ec.Indices

	p0 := func(i int) int {
		if i < 0 || i >= len(anchors) {
			return -1
		}
		return anchors[i].p0
	}
	p1 := func(i int) int {
		if i < 0 || i >= len(anchors) {
			return -1
		}
		return anchors[i].p1
	}
	valid := func(indices ...int) bool {
		for _, i := range indices {
			if i < 0 || i >= len(pts) {
				return false
			}
		}
		return true
	}

	switch ec.Kind {
	case Coincident:
		if len(idx) < 2 {
			return nil
		}
		a, b := p1(idx[0]), p0(idx[1])
		if !valid(a, b) {
			return nil
		}
		return CoincidentConstraint{A: a, B: b}

	case Horizontal:
		if len(idx) < 1 {
			return nil
		}
		a, b := p0(idx[0]), p1(idx[0])
		if !valid(a, b) {
			return nil
		}
		return HorizontalConstraint{A: a, B: b}

	case Vertical:
		if len(idx) < 1 {
			return nil
		}
		a, b := p0(idx[0]), p1(idx[0])
		if !valid(a, b) {
			return nil
		}
		return VerticalConstraint{A: a, B: b}

	case Parallel:
		if len(idx) < 2 {
			return nil
		}
		a1, a2 := p0(idx[0]), p1(idx[0])
		b1, b2 := p0(idx[1]), p1(idx[1])
		if !valid(a1, a2, b1, b2) {
			return nil
		}
		return ParallelConstraint{A1: a1, A2: a2, B1: b1, B2: b2}

	case Perpendicular:
		if len(idx) < 2 {
			return nil
		}
		a1, a2 := p0(idx[0]), p1(idx[0])
		b1, b2 := p0(idx[1]), p1(idx[1])
		if !valid(a1, a2, b1, b2) {
			return nil
		}
		return PerpendicularConstraint{A1: a1, A2: a2, B1: b1, B2: b2}

	case EqualLength:
		if len(idx) < 2 {
			return nil
		}
		a1, a2 := p0(idx[0]), p1(idx[0])
		b1, b2 := p0(idx[1]), p1(idx[1])
		if !valid(a1, a2, b1, b2) {
			return nil
		}
		return EqualLengthConstraint{A1: a1, A2: a2, B1: b1, B2: b2}

	case EqualRadius:
		// Requires two entities that each have a radius pointer (Circle or Arc).
		if len(idx) < 2 {
			return nil
		}
		ia, ib := idx[0], idx[1]
		if ia < 0 || ia >= len(anchors) || ib < 0 || ib >= len(anchors) {
			return nil
		}
		ra, rb := anchors[ia].radius, anchors[ib].radius
		if ra == nil || rb == nil {
			return nil
		}
		return EqualRadiusConstraint{RA: ra, RB: rb}

	case Tangent:
		// Indices[0] = line/segment entity, Indices[1] = circle/arc entity.
		if len(idx) < 2 {
			return nil
		}
		lineIdx, circIdx := idx[0], idx[1]
		if lineIdx < 0 || lineIdx >= len(anchors) || circIdx < 0 || circIdx >= len(anchors) {
			return nil
		}
		la, lb := p0(lineIdx), p1(lineIdx)
		cc := p0(circIdx)
		if !valid(la, lb, cc) {
			return nil
		}
		aCirc := anchors[circIdx]
		if aCirc.radius == nil {
			return nil
		}
		return TangentArcConstraint{
			LineA:        la,
			LineB:        lb,
			CircleCenter: cc,
			Radius:       aCirc.radius,
		}

	case Fixed:
		if len(idx) < 1 || ec.FixedPosition == nil {
			return nil
		}
		a := p0(idx[0])
		if !valid(a) {
			return nil
		}
		return FixedConstraint{Index: a, Position: *ec.FixedPosition}

	case Midpoint:
		if len(idx) < 3 {
			return nil
		}
		a, b, m := p0(idx[0]), p1(idx[1]), p0(idx[2])
		if !valid(a, b, m) {
			return nil
		}
		return MidpointConstraint{A: a, B: b, M: m}

	case Symmetric:
		if len(idx) < 4 {
			return nil
		}
		a, b := p0(idx[0]), p0(idx[1])
		ax1, ax2 := p0(idx[2]), p1(idx[3])
		if !valid(a, b, ax1, ax2) {
			return nil
		}
		return SymmetricConstraint{A: a, B: b, Axis1: ax1, Axis2: ax2}
	}
	return nil
}
