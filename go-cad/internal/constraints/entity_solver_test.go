package constraints_test

import (
        "math"
        "testing"

        "github.com/tomott12345/go-cad/internal/constraints"
        "github.com/tomott12345/go-cad/internal/geometry"
)

func TestSolveEntities_HorizontalSegment(t *testing.T) {
        entities := []geometry.Entity{
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 10, Y: 2},
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
                        Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 3, Y: 10},
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
                        Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 5, Y: 0},
                }},
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{X: 6, Y: 0}, End: geometry.Point{X: 10, Y: 0},
                }},
        }
        cs := []constraints.EntityConstraint{
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
                        Start: geometry.Point{X: 1, Y: 1}, End: geometry.Point{X: 10, Y: 0},
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
                        Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 10, Y: 0},
                }},
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{X: 0, Y: 5}, End: geometry.Point{X: 10, Y: 8},
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
                        Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 10, Y: 0},
                }},
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{X: 0, Y: 5}, End: geometry.Point{X: 4, Y: 5},
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
                        Center: geometry.Point{X: 5, Y: 5}, Radius: 3,
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
                        Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 10, Y: 0},
                }},
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{X: 5, Y: 0}, End: geometry.Point{X: 8, Y: 5},
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
                        Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 10, Y: 0},
                }},
        }
        cs := []constraints.EntityConstraint{
                {Kind: constraints.Horizontal, Indices: []int{99}},
        }
        updated, _ := constraints.SolveEntitiesDefault(entities, cs)
        if len(updated) == 0 {
                t.Error("SolveEntitiesDefault out-of-bounds index: expected entities returned unchanged")
        }
}

func TestSolveEntities_Midpoint(t *testing.T) {
        entities := []geometry.Entity{
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 10, Y: 0},
                }},
                geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{X: 0, Y: 5}, End: geometry.Point{X: 10, Y: 5},
                }},
                geometry.CircleEntity{Circle: geometry.Circle{
                        Center: geometry.Point{X: 0, Y: 0}, Radius: 1,
                }},
        }
        cs := []constraints.EntityConstraint{
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
                        Points: []geometry.Point{{X: 0, Y: 0}, {X: 5, Y: 1}, {X: 10, Y: 0}},
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

func TestTangentArcConstraint(t *testing.T) {
        r := 3.0
        p0 := geometry.Point{X: 0, Y: 0}
        p1 := geometry.Point{X: 10, Y: 0}
        center := geometry.Point{X: 5, Y: 3}
        pts := []*geometry.Point{&p0, &p1, &center}
        c := constraints.TangentArcConstraint{
                LineA: 0, LineB: 1, CircleCenter: 2, Radius: &r,
        }
        err := c.Error(pts)
        if math.Abs(err) > 1e-6 {
                t.Errorf("initial error: %v, want 0 (already tangent)", err)
        }
        c.Apply(pts)
        afterErr := c.Error(pts)
        if math.Abs(afterErr) > 1e-6 {
                t.Errorf("error after Apply on tangent config: %v, want ~0", afterErr)
        }
}

func TestSolveEntities_EqualRadius(t *testing.T) {
        c1 := geometry.CircleEntity{Circle: geometry.Circle{
                Center: geometry.Point{X: 0, Y: 0}, Radius: 3,
        }}
        c2 := geometry.CircleEntity{Circle: geometry.Circle{
                Center: geometry.Point{X: 10, Y: 0}, Radius: 7,
        }}
        entities := []geometry.Entity{c1, c2}
        cs := []constraints.EntityConstraint{{Kind: constraints.EqualRadius, Indices: []int{0, 1}}}
        updated, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Error("equal_radius: did not converge")
        }
        r1 := updated[0].(geometry.CircleEntity).Radius
        r2 := updated[1].(geometry.CircleEntity).Radius
        if math.Abs(r1-r2) > 1e-4 {
                t.Errorf("equal_radius: radii differ: %.4f vs %.4f", r1, r2)
        }
        if math.Abs(r1-5.0) > 1e-4 {
                t.Errorf("equal_radius: expected avg 5.0, got %.4f", r1)
        }
}

func TestSolveEntities_EqualRadius_Arc(t *testing.T) {
        a1 := geometry.ArcEntity{Arc: geometry.Arc{
                Center: geometry.Point{X: 0, Y: 0}, Radius: 4, StartDeg: 0, EndDeg: 90,
        }}
        a2 := geometry.ArcEntity{Arc: geometry.Arc{
                Center: geometry.Point{X: 5, Y: 0}, Radius: 8, StartDeg: 0, EndDeg: 90,
        }}
        entities := []geometry.Entity{a1, a2}
        cs := []constraints.EntityConstraint{{Kind: constraints.EqualRadius, Indices: []int{0, 1}}}
        updated, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Error("equal_radius arc: did not converge")
        }
        r1 := updated[0].(geometry.ArcEntity).Radius
        r2 := updated[1].(geometry.ArcEntity).Radius
        if math.Abs(r1-r2) > 1e-4 {
                t.Errorf("equal_radius arc: radii differ: %.4f vs %.4f", r1, r2)
        }
}

func TestSolveEntities_Tangent(t *testing.T) {
        line := geometry.SegmentEntity{Segment: geometry.Segment{
                Start: geometry.Point{X: -10, Y: 3}, End: geometry.Point{X: 10, Y: 3},
        }}
        circle := geometry.CircleEntity{Circle: geometry.Circle{
                Center: geometry.Point{X: 0, Y: 0}, Radius: 3,
        }}
        entities := []geometry.Entity{line, circle}
        cs := []constraints.EntityConstraint{{Kind: constraints.Tangent, Indices: []int{0, 1}}}
        _, res := constraints.SolveEntitiesDefault(entities, cs)
        if !res.Converged {
                t.Errorf("tangent: did not converge (error=%.8f)", res.FinalError)
        }
}

func TestSolveEntities_EqualRadius_NilRadius(t *testing.T) {
        seg := geometry.SegmentEntity{Segment: geometry.Segment{
                Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 5, Y: 0},
        }}
        circle := geometry.CircleEntity{Circle: geometry.Circle{
                Center: geometry.Point{X: 10, Y: 0}, Radius: 3,
        }}
        entities := []geometry.Entity{seg, circle}
        cs := []constraints.EntityConstraint{{Kind: constraints.EqualRadius, Indices: []int{0, 1}}}
        result, _ := constraints.SolveEntitiesDefault(entities, cs)
        if len(result) == 0 {
                t.Error("EqualRadius mixed types: expected entities returned")
        }
}

func TestSolveEntities_Tangent_NoRadius(t *testing.T) {
        seg1 := geometry.SegmentEntity{Segment: geometry.Segment{
                Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 5, Y: 0},
        }}
        seg2 := geometry.SegmentEntity{Segment: geometry.Segment{
                Start: geometry.Point{X: 0, Y: 1}, End: geometry.Point{X: 5, Y: 1},
        }}
        entities := []geometry.Entity{seg1, seg2}
        cs := []constraints.EntityConstraint{{Kind: constraints.Tangent, Indices: []int{0, 1}}}
        result, _ := constraints.SolveEntitiesDefault(entities, cs)
        if len(result) == 0 {
                t.Error("Tangent no-radius: expected entities returned")
        }
}

func TestSolveEntities_Symmetric(t *testing.T) {
        seg1 := geometry.SegmentEntity{Segment: geometry.Segment{
                Start: geometry.Point{X: -3, Y: 5}, End: geometry.Point{X: -3, Y: 0},
        }}
        axis := geometry.SegmentEntity{Segment: geometry.Segment{
                Start: geometry.Point{X: 0, Y: -10}, End: geometry.Point{X: 0, Y: 10},
        }}
        seg2 := geometry.SegmentEntity{Segment: geometry.Segment{
                Start: geometry.Point{X: 4, Y: 5}, End: geometry.Point{X: 4, Y: 0},
        }}
        entities := []geometry.Entity{seg1, seg2, axis}
        cs := []constraints.EntityConstraint{
                {Kind: constraints.Symmetric, Indices: []int{0, 1, 2, 2}},
        }
        updated, _ := constraints.SolveEntitiesDefault(entities, cs)
        s1 := updated[0].(geometry.SegmentEntity)
        s2 := updated[1].(geometry.SegmentEntity)
        if math.Abs(s1.Start.X+s2.Start.X) > 0.5 {
                t.Errorf("symmetric: expected X mirror, got %v %v", s1.Start.X, s2.Start.X)
        }
}

func TestSolveEntities_Spline(t *testing.T) {
        sp := geometry.NURBSEntity{NURBSSpline: geometry.NewNURBSSpline(2,
                []float64{0, 0, 0, 1, 1, 1},
                []geometry.Point{{X: 0, Y: 0}, {X: 5, Y: 5}, {X: 10, Y: 0}},
                nil,
        )}
        seg := geometry.SegmentEntity{Segment: geometry.Segment{
                Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 10, Y: 0},
        }}
        entities := []geometry.Entity{sp, seg}
        cs := []constraints.EntityConstraint{{Kind: constraints.Horizontal, Indices: []int{1}}}
        updated, _ := constraints.SolveEntitiesDefault(entities, cs)
        if len(updated) != 2 {
                t.Errorf("spline: expected 2 entities, got %d", len(updated))
        }
}

func TestSolveEntities_EmptyPolyline(t *testing.T) {
        poly := geometry.PolylineEntity{Polyline: geometry.Polyline{Points: nil}}
        entities := []geometry.Entity{poly}
        cs := []constraints.EntityConstraint{{Kind: constraints.Horizontal, Indices: []int{0}}}
        updated, _ := constraints.SolveEntitiesDefault(entities, cs)
        if len(updated) != 1 {
                t.Errorf("empty polyline: expected 1 entity, got %d", len(updated))
        }
}

func TestSolveEntities_NilConstraint(t *testing.T) {
        seg := geometry.SegmentEntity{Segment: geometry.Segment{
                Start: geometry.Point{X: 0, Y: 0}, End: geometry.Point{X: 5, Y: 0},
        }}
        entities := []geometry.Entity{seg}
        cs := []constraints.EntityConstraint{
                {Kind: constraints.Coincident, Indices: []int{99, 99}},
                {Kind: constraints.EqualRadius, Indices: []int{0, 99}},
                {Kind: constraints.EqualRadius, Indices: []int{}},
                {Kind: constraints.Tangent, Indices: []int{99, 99}},
                {Kind: constraints.Tangent, Indices: []int{}},
                {Kind: constraints.Horizontal, Indices: []int{}},
                {Kind: constraints.Vertical, Indices: []int{}},
                {Kind: constraints.Parallel, Indices: []int{0}},
                {Kind: constraints.Perpendicular, Indices: []int{0}},
                {Kind: constraints.EqualLength, Indices: []int{0}},
                {Kind: constraints.Fixed, Indices: []int{0}},
                {Kind: constraints.Midpoint, Indices: []int{0, 0}},
                {Kind: constraints.Symmetric, Indices: []int{0, 0, 0}},
        }
        updated, _ := constraints.SolveEntitiesDefault(entities, cs)
        if len(updated) != 1 {
                t.Errorf("nil constraints: expected 1 entity, got %d", len(updated))
        }
}
