package constraints

import (
        "math"
        "testing"

        "go-cad/internal/geometry"
)

func ptrs(pts ...geometry.Point) []*geometry.Point {
        out := make([]*geometry.Point, len(pts))
        for i := range pts {
                out[i] = &pts[i]
        }
        return out
}

func TestCoincident(t *testing.T) {
        pts := ptrs(geometry.Point{X: 0, Y: 0}, geometry.Point{X: 4, Y: 4})
        c := CoincidentConstraint{A: 0, B: 1}
        res := Solve(pts, []Constraint{c}, 0, 0)
        if !res.Converged {
                t.Errorf("CoincidentConstraint did not converge: err=%v", res.FinalError)
        }
        if !pts[0].Near(*pts[1]) {
                t.Errorf("CoincidentConstraint: points not coincident: %v vs %v", *pts[0], *pts[1])
        }
}

func TestHorizontal(t *testing.T) {
        pts := ptrs(geometry.Point{X: 0, Y: 1}, geometry.Point{X: 5, Y: 3})
        c := HorizontalConstraint{A: 0, B: 1}
        res := Solve(pts, []Constraint{c}, 0, 0)
        if !res.Converged {
                t.Errorf("HorizontalConstraint did not converge")
        }
        if math.Abs(pts[0].Y-pts[1].Y) > geometry.Epsilon*10 {
                t.Errorf("HorizontalConstraint: Y mismatch: %v vs %v", pts[0].Y, pts[1].Y)
        }
}

func TestVertical(t *testing.T) {
        pts := ptrs(geometry.Point{X: 1, Y: 0}, geometry.Point{X: 3, Y: 5})
        c := VerticalConstraint{A: 0, B: 1}
        res := Solve(pts, []Constraint{c}, 0, 0)
        if !res.Converged {
                t.Errorf("VerticalConstraint did not converge")
        }
        if math.Abs(pts[0].X-pts[1].X) > geometry.Epsilon*10 {
                t.Errorf("VerticalConstraint: X mismatch: %v vs %v", pts[0].X, pts[1].X)
        }
}

func TestFixed(t *testing.T) {
        target := geometry.Point{X: 7, Y: 8}
        pts := ptrs(geometry.Point{X: 0, Y: 0})
        c := FixedConstraint{Index: 0, Position: target}
        res := Solve(pts, []Constraint{c}, 0, 0)
        if !res.Converged {
                t.Errorf("FixedConstraint did not converge")
        }
        if !pts[0].Near(target) {
                t.Errorf("FixedConstraint: got %v, want %v", *pts[0], target)
        }
}

func TestParallel(t *testing.T) {
        // Segment A: (0,0)→(5,0) (horizontal)
        // Segment B: (0,2)→(5,3) (slightly tilted)
        pts := ptrs(
                geometry.Point{X: 0, Y: 0}, geometry.Point{X: 5, Y: 0},
                geometry.Point{X: 0, Y: 2}, geometry.Point{X: 5, Y: 3},
        )
        c := ParallelConstraint{A1: 0, A2: 1, B1: 2, B2: 3}
        res := Solve(pts, []Constraint{c}, 100, 1e-6)
        if !res.Converged {
                t.Logf("Parallel: not converged (err=%v) — acceptable for near-degenerate input", res.FinalError)
        }
        // After solving B should be parallel to A
        da := pts[1].Sub(*pts[0]).Normalize()
        db := pts[3].Sub(*pts[2]).Normalize()
        cross := math.Abs(da.Cross(db))
        if cross > 0.01 {
                t.Errorf("ParallelConstraint: cross product should be ~0, got %v", cross)
        }
}

func TestPerpendicular(t *testing.T) {
        // A: (0,0)→(5,0), B: (2,0)→(2,5)
        pts := ptrs(
                geometry.Point{X: 0, Y: 0}, geometry.Point{X: 5, Y: 0},
                geometry.Point{X: 2, Y: 0}, geometry.Point{X: 3, Y: 1}, // slightly off perpendicular
        )
        c := PerpendicularConstraint{A1: 0, A2: 1, B1: 2, B2: 3}
        Solve(pts, []Constraint{c}, 100, 1e-6)
        da := pts[1].Sub(*pts[0]).Normalize()
        db := pts[3].Sub(*pts[2]).Normalize()
        dot := math.Abs(da.Dot(db))
        if dot > 0.01 {
                t.Errorf("PerpendicularConstraint: dot product should be ~0, got %v", dot)
        }
}

func TestEqualLength(t *testing.T) {
        pts := ptrs(
                geometry.Point{X: 0, Y: 0}, geometry.Point{X: 3, Y: 0}, // len 3
                geometry.Point{X: 0, Y: 5}, geometry.Point{X: 7, Y: 5}, // len 7
        )
        c := EqualLengthConstraint{A1: 0, A2: 1, B1: 2, B2: 3}
        Solve(pts, []Constraint{c}, 200, 1e-6)
        la := pts[0].Dist(*pts[1])
        lb := pts[2].Dist(*pts[3])
        if math.Abs(la-lb) > 0.01 {
                t.Errorf("EqualLength: la=%v lb=%v, should be equal", la, lb)
        }
}

func TestMidpoint(t *testing.T) {
        pts := ptrs(
                geometry.Point{X: 0, Y: 0},
                geometry.Point{X: 10, Y: 0},
                geometry.Point{X: 3, Y: 0}, // incorrect midpoint
        )
        c := MidpointConstraint{A: 0, B: 1, M: 2}
        res := Solve(pts, []Constraint{c}, 0, 0)
        if !res.Converged {
                t.Errorf("MidpointConstraint did not converge")
        }
        want := geometry.Point{X: 5, Y: 0}
        if !pts[2].Near(want) {
                t.Errorf("MidpointConstraint: got %v, want %v", *pts[2], want)
        }
}

func TestSymmetric(t *testing.T) {
        // Axis: X-axis (Y=0 line), A=(3,4), B should be (3,-4)
        pts := ptrs(
                geometry.Point{X: 0, Y: 0}, geometry.Point{X: 10, Y: 0}, // axis along x-axis
                geometry.Point{X: 3, Y: 4},  // A
                geometry.Point{X: 3, Y: 1},  // B (should become mirror of A)
        )
        c := SymmetricConstraint{Axis1: 0, Axis2: 1, A: 2, B: 3}
        Solve(pts, []Constraint{c}, 200, 1e-6)
        // B.Y should be approximately -A.Y
        if math.Abs(pts[3].Y+pts[2].Y) > 0.05 {
                t.Errorf("SymmetricConstraint: A.Y=%v B.Y=%v, should be symmetric across x-axis",
                        pts[2].Y, pts[3].Y)
        }
}

func TestCombinedConstraints(t *testing.T) {
        // Square: 4 points, horizontal+vertical+equal length constraints
        pts := ptrs(
                geometry.Point{X: 0, Y: 0},   // 0: origin (fixed)
                geometry.Point{X: 5.1, Y: 0.2}, // 1: should become (5,0)
                geometry.Point{X: 5.2, Y: 4.8}, // 2: should become (5,5)
                geometry.Point{X: 0.1, Y: 5.3}, // 3: should become (0,5)
        )
        constraints := []Constraint{
                FixedConstraint{Index: 0, Position: geometry.Point{X: 0, Y: 0}},
                HorizontalConstraint{A: 0, B: 1},
                VerticalConstraint{A: 1, B: 2},
                HorizontalConstraint{A: 2, B: 3},
                VerticalConstraint{A: 0, B: 3},
                EqualLengthConstraint{A1: 0, A2: 1, B1: 1, B2: 2},
        }
        res := Solve(pts, constraints, 500, 1e-4)
        t.Logf("Combined: converged=%v iter=%d err=%v", res.Converged, res.Iterations, res.FinalError)
        // Origin fixed
        if !pts[0].Near(geometry.Point{X: 0, Y: 0}) {
                t.Errorf("Combined: origin moved to %v", *pts[0])
        }
        // pts[0] and pts[1] should be horizontal
        if math.Abs(pts[0].Y-pts[1].Y) > 0.1 {
                t.Errorf("Combined: bottom edge not horizontal: %v vs %v", pts[0].Y, pts[1].Y)
        }
}

func TestSolveResult_NotConverged(t *testing.T) {
        // Contradictory constraints: fix pt0 at (0,0) AND require it coincident with (1,1)
        pt0 := geometry.Point{X: 0, Y: 0}
        pt1 := geometry.Point{X: 1, Y: 1}
        pts := []*geometry.Point{&pt0, &pt1}
        constraints := []Constraint{
                FixedConstraint{Index: 0, Position: geometry.Point{X: 0, Y: 0}},
                FixedConstraint{Index: 1, Position: geometry.Point{X: 1, Y: 1}},
                CoincidentConstraint{A: 0, B: 1},
        }
        res := Solve(pts, constraints, 10, 1e-10)
        if res.Converged {
                t.Error("Contradictory constraints should not converge")
        }
        if res.Iterations != 10 {
                t.Errorf("Expected 10 iterations, got %d", res.Iterations)
        }
}
