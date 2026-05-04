// Package constraints_test demonstrates how to build and solve entity-level
// constraints using the EntityConstraint API.
//
// Index semantics quick-reference
//
//	Coincident         Indices: [A, B]             — endpoint of A meets start of B
//	Horizontal         Indices: [A]                — segment A is horizontal
//	Vertical           Indices: [A]                — segment A is vertical
//	Parallel           Indices: [A, B]             — segments A and B are parallel
//	Perpendicular      Indices: [A, B]             — segments A and B are perpendicular
//	EqualLength        Indices: [A, B]             — two segment entities equal length
//	EqualRadius        Indices: [A, B]             — two circle/arc entities equal radius
//	Tangent            Indices: [lineEnt, circEnt] — line entity tangent to circle/arc
//	Fixed              Indices: [A]                — entity A is pinned in place
//	Midpoint           Indices: [A, B, M]          — M is midpoint of segment A→B
//	Symmetric          Indices: [A, B, Ax1, Ax2]  — A,B symmetric about axis entity Ax1→Ax2
package constraints_test

import (
	"math"
	"testing"

	"go-cad/internal/constraints"
	"go-cad/internal/geometry"
)

// TestSolveUsage_Horizontal shows how to enforce a horizontal constraint on a
// tilted line segment.  After solving, both endpoints share the same Y value.
func TestSolveUsage_Horizontal(t *testing.T) {
	seg := geometry.SegmentEntity{Segment: geometry.Segment{
		Start: geometry.Point{X: 0, Y: 0},
		End:   geometry.Point{X: 10, Y: 4},
	}}

	cs := []constraints.EntityConstraint{
		{Kind: constraints.Horizontal, Indices: []int{0}},
	}

	solved, _ := constraints.SolveEntitiesDefault([]geometry.Entity{seg}, cs)
	s := solved[0].(geometry.SegmentEntity)
	if math.Abs(s.Start.Y-s.End.Y) > 1e-6 {
		t.Errorf("horizontal: start.Y=%v end.Y=%v should be equal", s.Start.Y, s.End.Y)
	}
}

// TestSolveUsage_EqualLength shows how to make two segments the same length.
func TestSolveUsage_EqualLength(t *testing.T) {
	seg1 := geometry.SegmentEntity{Segment: geometry.Segment{
		Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 6, Y: 0},
	}}
	seg2 := geometry.SegmentEntity{Segment: geometry.Segment{
		Start: geometry.Point{X: 0, Y: 2}, End: geometry.Point{X: 10, Y: 2},
	}}

	cs := []constraints.EntityConstraint{
		{Kind: constraints.EqualLength, Indices: []int{0, 1}},
	}

	solved, _ := constraints.SolveEntitiesDefault([]geometry.Entity{seg1, seg2}, cs)
	l1 := solved[0].(geometry.SegmentEntity).Segment.Length()
	l2 := solved[1].(geometry.SegmentEntity).Segment.Length()
	if math.Abs(l1-l2) > 1e-4 {
		t.Errorf("EqualLength: l1=%v l2=%v should be equal", l1, l2)
	}
}

// TestSolveUsage_Coincident shows how to snap the end of one segment to the
// start of another (chain two segments into a polyline vertex).
func TestSolveUsage_Coincident(t *testing.T) {
	seg1 := geometry.SegmentEntity{Segment: geometry.Segment{
		Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 5, Y: 0},
	}}
	seg2 := geometry.SegmentEntity{Segment: geometry.Segment{
		Start: geometry.Point{X: 5.5, Y: 0.5}, End: geometry.Point{X: 10, Y: 0},
	}}

	cs := []constraints.EntityConstraint{
		{Kind: constraints.Coincident, Indices: []int{0, 1}},
	}

	solved, _ := constraints.SolveEntitiesDefault([]geometry.Entity{seg1, seg2}, cs)
	e1 := solved[0].(geometry.SegmentEntity).End
	s2 := solved[1].(geometry.SegmentEntity).Start
	if e1.Dist(s2) > 1e-4 {
		t.Errorf("Coincident: end %v and start %v should coincide", e1, s2)
	}
}
