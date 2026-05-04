package constraints_test

import (
        "math"
        "testing"

        "go-cad/internal/constraints"
        "go-cad/internal/geometry"
)

func TestSolveEntities_HorizontalSegment(t *testing.T) {
        entities := []geometry.Entity{
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{0, 0}, End: geometry.Point{10, 2},
                }},
        }
        cs := []constraints.EntityConstraint{
                {Kind: constraints.Horizontal, Indices: []int{0}},
        }
        updated, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Errorf("horizontal: did not converge (err=%v)", res.FinalError)
        }
        seg := updated[0].(geometry.SegmentEntity)
        if math.Abs(seg.Start.Y-seg.End.Y) > 1e-6 {
                t.Errorf("horizontal: Y values not equal: %v vs %v", seg.Start.Y, seg.End.Y)
        }
}

func TestSolveEntities_VerticalSegment(t *testing.T) {
        entities := []geometry.Entity{
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{0, 0}, End: geometry.Point{3, 10},
                }},
        }
        cs := []constraints.EntityConstraint{
                {Kind: constraints.Vertical, Indices: []int{0}},
        }
        updated, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Errorf("vertical: did not converge (err=%v)", res.FinalError)
        }
        seg := updated[0].(geometry.SegmentEntity)
        if math.Abs(seg.Start.X-seg.End.X) > 1e-6 {
                t.Errorf("vertical: X values not equal: %v vs %v", seg.Start.X, seg.End.X)
        }
}

func TestSolveEntities_CoincidentEndpoints(t *testing.T) {
        entities := []geometry.Entity{
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{0, 0}, End: geometry.Point{5, 0},
                }},
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{6, 0}, End: geometry.Point{10, 0},
                }},
        }
        cs := []constraints.EntityConstraint{
                // end of entity 0 coincides with start of entity 1
                {Kind: constraints.Coincident, Indices: []int{0, 1}},
        }
        updated, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Errorf("coincident: did not converge (err=%v)", res.FinalError)
        }
        s0 := updated[0].(geometry.SegmentEntity)
        s1 := updated[1].(geometry.SegmentEntity)
        if s0.End.Dist(s1.Start) > 1e-6 {
                t.Errorf("coincident: end/start mismatch: %v vs %v", s0.End, s1.Start)
        }
}

func TestSolveEntities_FixedSegment(t *testing.T) {
        fixed := geometry.Point{X: 0, Y: 0}
        entities := []geometry.Entity{
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{1, 1}, End: geometry.Point{10, 0},
                }},
        }
        cs := []constraints.EntityConstraint{
                {Kind: constraints.Fixed, Indices: []int{0}, FixedPosition: &fixed},
                {Kind: constraints.Horizontal, Indices: []int{0}},
        }
        updated, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Errorf("fixed+horizontal: did not converge (err=%v)", res.FinalError)
        }
        seg := updated[0].(geometry.SegmentEntity)
        if seg.Start.Dist(fixed) > 1e-6 {
                t.Errorf("fixed start moved: got %v", seg.Start)
        }
}

func TestSolveEntities_ParallelSegments(t *testing.T) {
        entities := []geometry.Entity{
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{0, 0}, End: geometry.Point{10, 0},
                }},
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{0, 5}, End: geometry.Point{10, 8},
                }},
        }
        cs := []constraints.EntityConstraint{
                {Kind: constraints.Parallel, Indices: []int{0, 1}},
        }
        updated, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Errorf("parallel: did not converge (err=%v)", res.FinalError)
        }
        s0 := updated[0].(geometry.SegmentEntity)
        s1 := updated[1].(geometry.SegmentEntity)
        d0 := s0.End.Sub(s0.Start).Normalize()
        d1 := s1.End.Sub(s1.Start).Normalize()
        cross := math.Abs(d0.Cross(d1))
        if cross > 1e-6 {
                t.Errorf("parallel: cross product %v, want ~0", cross)
        }
}

func TestSolveEntities_EqualLength(t *testing.T) {
        entities := []geometry.Entity{
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{0, 0}, End: geometry.Point{10, 0},
                }},
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{0, 5}, End: geometry.Point{4, 5},
                }},
        }
        cs := []constraints.EntityConstraint{
                {Kind: constraints.EqualLength, Indices: []int{0, 1}},
        }
        updated, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Errorf("equal-length: did not converge (err=%v)", res.FinalError)
        }
        s0 := updated[0].(geometry.SegmentEntity)
        s1 := updated[1].(geometry.SegmentEntity)
        l0 := s0.Start.Dist(s0.End)
        l1 := s1.Start.Dist(s1.End)
        if math.Abs(l0-l1) > 1e-4 {
                t.Errorf("equal-length: l0=%v l1=%v", l0, l1)
        }
}

func TestSolveEntities_CirclePreservesRadius(t *testing.T) {
        entities := []geometry.Entity{
                geometry.CircleEntity{Circle: geometry.Circle{
                        Center: geometry.Point{5, 5}, Radius: 3,
                }},
        }
        fixed := geometry.Point{X: 0, Y: 0}
        cs := []constraints.EntityConstraint{
                {Kind: constraints.Fixed, Indices: []int{0}, FixedPosition: &fixed},
        }
        updated, _ := constraints.SolveEntitiesDefault(entities, cs)
        c := updated[0].(geometry.CircleEntity)
        if c.Radius != 3 {
                t.Errorf("circle radius changed: %v", c.Radius)
        }
        if c.Center.Dist(fixed) > 1e-6 {
                t.Errorf("circle center not fixed: %v", c.Center)
        }
}

func TestSolveEntities_Perpendicular(t *testing.T) {
        entities := []geometry.Entity{
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{0, 0}, End: geometry.Point{10, 0},
                }},
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{5, 0}, End: geometry.Point{8, 5},
                }},
        }
        cs := []constraints.EntityConstraint{
                {Kind: constraints.Perpendicular, Indices: []int{0, 1}},
        }
        updated, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Errorf("perpendicular: did not converge (err=%v)", res.FinalError)
        }
        s0 := updated[0].(geometry.SegmentEntity)
        s1 := updated[1].(geometry.SegmentEntity)
        d0 := s0.End.Sub(s0.Start).Normalize()
        d1 := s1.End.Sub(s1.Start).Normalize()
        dot := math.Abs(d0.Dot(d1))
        if dot > 1e-6 {
                t.Errorf("perpendicular: dot product %v, want ~0", dot)
        }
}

func TestSolveEntities_InvalidIndices(t *testing.T) {
        entities := []geometry.Entity{
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{0, 0}, End: geometry.Point{10, 0},
                }},
        }
        cs := []constraints.EntityConstraint{
                // index 99 does not exist — should be silently skipped
                {Kind: constraints.Horizontal, Indices: []int{99}},
        }
        // Must not panic
        updated, _ := constraints.SolveEntitiesDefault(entities, cs)
        _ = updated
}

func TestSolveEntities_Midpoint(t *testing.T) {
        entities := []geometry.Entity{
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{0, 0}, End: geometry.Point{10, 0},
                }},
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{0, 5}, End: geometry.Point{10, 5},
                }},
                geometry.CircleEntity{Circle: geometry.Circle{
                        Center: geometry.Point{0, 0}, Radius: 1,
                }},
        }
        cs := []constraints.EntityConstraint{
                // center of circle (entity 2) at midpoint of entity 0 endpoints
                {Kind: constraints.Midpoint, Indices: []int{0, 0, 2}},
        }
        updated, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Errorf("midpoint: did not converge (err=%v)", res.FinalError)
        }
        c := updated[2].(geometry.CircleEntity)
        if math.Abs(c.Center.X-5) > 1e-6 || math.Abs(c.Center.Y) > 1e-6 {
                t.Errorf("midpoint: circle center %v, want {5,0}", c.Center)
        }
}

func TestSolveEntities_Polyline(t *testing.T) {
        entities := []geometry.Entity{
                geometry.PolylineEntity{Polyline: geometry.Polyline{
                        Points: []geometry.Point{{0, 0}, {5, 1}, {10, 0}},
                }},
        }
        target := geometry.Point{X: 0, Y: 5}
        cs := []constraints.EntityConstraint{
                {Kind: constraints.Fixed, Indices: []int{0}, FixedPosition: &target},
        }
        updated, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Errorf("polyline fixed: did not converge")
        }
        pl := updated[0].(geometry.PolylineEntity)
        if pl.Points[0].Dist(target) > 1e-6 {
                t.Errorf("polyline first point not fixed: %v", pl.Points[0])
        }
}

// TestEqualRadiusConstraint verifies the EqualRadius constraint.
func TestEqualRadiusConstraint(t *testing.T) {
        ra := 5.0
        rb := 3.0
        c := constraints.EqualRadiusConstraint{RA: &ra, RB: &rb}

        if math.Abs(c.Error(nil)-2) > 1e-9 {
                t.Errorf("initial error: %v", c.Error(nil))
        }
        c.Apply(nil)
        if math.Abs(ra-rb) > 1e-9 {
                t.Errorf("after apply: ra=%v rb=%v", ra, rb)
        }
        if math.Abs(ra-4) > 1e-9 {
                t.Errorf("expected average 4, got %v", ra)
        }
}

// TestTangentArcConstraint verifies the pointer-based tangent constraint.
// TangentArcConstraint translates the LINE to achieve dist(line,center) == r.
// This test checks Error() semantics and that Apply() does not panic, and
// that a single application reduces the error when center is already tangent.
func TestTangentArcConstraint(t *testing.T) {
        r := 3.0
        p0 := geometry.Point{0, 0}
        p1 := geometry.Point{10, 0}
        // Place center exactly r above the line → already tangent, error = 0.
        center := geometry.Point{5, 3}

        pts := []*geometry.Point{&p0, &p1, &center}
        c := constraints.TangentArcConstraint{
                LineA: 0, LineB: 1, CircleCenter: 2, Radius: &r,
        }

        err := c.Error(pts)
        // dist = 3 (center 3 above line y=0), radius = 3 → error = 0
        if math.Abs(err) > 1e-6 {
                t.Errorf("initial error: %v, want 0 (already tangent)", err)
        }

        // Apply on already-tangent config must be a no-op (correction = 0).
        c.Apply(pts)
        afterErr := c.Error(pts)
        if math.Abs(afterErr) > 1e-6 {
                t.Errorf("error after Apply on tangent config: %v, want ~0", afterErr)
        }
}
