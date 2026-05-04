// entity_solver.go provides the entity-level constraint solving API.
// It bridges the geometry.Entity world (rich typed primitives) with the
// point-handle Gauss-Seidel solver, allowing constraints to be expressed
// directly in terms of entities and returned as updated entities.
package constraints

import (
	"go-cad/internal/geometry"
)

// EntityConstraint describes a constraint referencing named entities by index
// into the entity slice rather than raw point handles.
// Use this API when working with document-level geometry, not bare points.
type EntityConstraint struct {
	// Kind is the constraint type.
	Kind ConstraintKind
	// Indices holds the entity indices this constraint acts on.
	// Semantics depend on Kind:
	//   Coincident         [A, B]   — endpoints coincide
	//   Horizontal/Vertical[A, B]   — segment endpoints horizontal/vertical
	//   Parallel/Perp      [A1,A2,B1,B2] — two segment pairs
	//   EqualLength        [A1,A2,B1,B2] — two segment pairs
	//   Fixed              [A]      — entity fixed in place
	//   Midpoint           [A,B,M]  — M is midpoint of segment A→B
	//   Symmetric          [A,B,Ax1,Ax2] — A,B symmetric about axis Ax
	Indices []int
	// FixedPosition is used by Fixed constraints to pin a specific location.
	FixedPosition *geometry.Point
}

// SolveEntities solves parametric constraints over a slice of geometry entities.
// It extracts representative "anchor" points from each entity (e.g. endpoints,
// center), solves the constraints with the Gauss-Seidel solver, and then
// reconstructs updated entities from the moved points.
//
// The function returns the updated entity slice. The original entities are not
// modified.
//
// Limitations: Only Segment, Line, Arc, Circle, and Polyline entities expose
// movable anchor points. Spline entities are treated as immovable.
func SolveEntities(
	entities []geometry.Entity,
	constraints []EntityConstraint,
	maxIter int,
	tol float64,
) ([]geometry.Entity, SolveResult) {
	// Step 1: Extract anchor points from each entity.
	// Each entity maps to one or two anchor points; we keep a lookup table.
	pts, anchors := extractAnchors(entities)

	// Step 2: Map EntityConstraint → Constraint (point-handle form).
	var solverConstraints []Constraint
	for _, ec := range constraints {
		if c := mapEntityConstraint(ec, anchors, pts); c != nil {
			solverConstraints = append(solverConstraints, c)
		}
	}

	// Step 3: Run the Gauss-Seidel solver.
	result := Solve(pts, solverConstraints, maxIter, tol)

	// Step 4: Reconstruct entities from the (possibly moved) anchor points.
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

// ─── anchor registry ─────────────────────────────────────────────────────────

// entityAnchor records the point indices for a single entity.
type entityAnchor struct {
	// p0, p1 are the primary anchor point indices (e.g. segment start/end,
	// or circle/arc center). -1 if unused.
	p0, p1 int
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
			anchors[i] = entityAnchor{
				p0: add(v.Center),
				p1: -1,
			}
		case geometry.ArcEntity:
			anchors[i] = entityAnchor{
				p0: add(v.Center),
				p1: -1,
			}
		case geometry.PolylineEntity:
			// Use first and last vertex as anchors.
			if len(v.Points) > 0 {
				anchors[i] = entityAnchor{
					p0: add(v.Points[0]),
					p1: add(v.Points[len(v.Points)-1]),
				}
			} else {
				anchors[i] = entityAnchor{p0: -1, p1: -1}
			}
		default:
			// Spline / unknown: no movable anchors.
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
				Dir:    v.Dir, // direction preserved
			}}
		case geometry.CircleEntity:
			out[i] = geometry.CircleEntity{Circle: geometry.Circle{
				Center: *pts[a.p0],
				Radius: v.Radius,
			}}
		case geometry.ArcEntity:
			out[i] = geometry.ArcEntity{Arc: geometry.Arc{
				Center:   *pts[a.p0],
				Radius:   v.Radius,
				StartDeg: v.StartDeg,
				EndDeg:   v.EndDeg,
			}}
		case geometry.PolylineEntity:
			if a.p0 >= 0 && a.p1 >= 0 && len(v.Points) >= 2 {
				// Translate all polyline points by the delta applied to p0.
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
			out[i] = e // splines preserved as-is
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
) Constraint {
	idx := ec.Indices

	// Helper: safely retrieve p0 or p1 of anchor for entity i.
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
		a, b := p1(idx[0]), p0(idx[1]) // end of A coincides with start of B
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
	}
	return nil
}
