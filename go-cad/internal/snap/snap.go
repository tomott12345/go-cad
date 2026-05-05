// Package snap implements the object-snap engine for go-cad.
//
// FindSnap evaluates all document entities and returns the highest-priority
// snap candidate within a given cursor radius. Snap types match QCAD/AutoCAD
// drafting conventions.
//
// Priority order (highest first):
//
//	Endpoint > Midpoint > Center > Quadrant > Intersection >
//	Perpendicular > Tangent > Nearest
package snap

import (
	"math"

	"github.com/tomott12345/go-cad/internal/document"
	"github.com/tomott12345/go-cad/internal/geometry"
)

// ─── Snap type mask ───────────────────────────────────────────────────────────

// SnapType is a bitmask of enabled snap modes.
type SnapType int

const (
	SnapEndpoint      SnapType = 1 << iota // 1
	SnapMidpoint                            // 2
	SnapCenter                              // 4
	SnapQuadrant                            // 8
	SnapIntersection                        // 16
	SnapPerpendicular                       // 32
	SnapTangent                             // 64
	SnapNearest                             // 128

	SnapAll = SnapEndpoint | SnapMidpoint | SnapCenter | SnapQuadrant |
		SnapIntersection | SnapPerpendicular | SnapTangent | SnapNearest
)

// SnapNames maps SnapType to a display name (only single-bit values).
var SnapNames = map[SnapType]string{
	SnapEndpoint:      "Endpoint",
	SnapMidpoint:      "Midpoint",
	SnapCenter:        "Center",
	SnapQuadrant:      "Quadrant",
	SnapIntersection:  "Intersection",
	SnapPerpendicular: "Perpendicular",
	SnapTangent:       "Tangent",
	SnapNearest:       "Nearest",
}

// ─── Snap candidate ───────────────────────────────────────────────────────────

// SnapCandidate holds a single snap point result.
type SnapCandidate struct {
	X, Y     float64
	Type     SnapType
	EntityID int
}

// ─── Public API ───────────────────────────────────────────────────────────────

// FindSnap evaluates all entities and returns the highest-priority snap
// candidate within radius of the cursor at (cx,cy).
// Returns nil if no snap candidate is found within radius.
// mask controls which snap types are active (use SnapAll for all types).
func FindSnap(cx, cy float64, entities []document.Entity, radius float64, mask SnapType) *SnapCandidate {
	if radius <= 0 {
		radius = 10
	}
	priority := []SnapType{
		SnapEndpoint, SnapMidpoint, SnapCenter, SnapQuadrant,
		SnapIntersection, SnapPerpendicular, SnapTangent, SnapNearest,
	}

	for _, stype := range priority {
		if mask&stype == 0 {
			continue
		}

		var best *SnapCandidate
		bestD := radius

		if stype == SnapIntersection {
			// O(n²) pairwise — acceptable for typical CAD document sizes.
			for i := 0; i < len(entities); i++ {
				for j := i + 1; j < len(entities); j++ {
					gpts := entities[i].IntersectWith(entities[j])
					for _, p := range gpts {
						d := math.Hypot(p.X-cx, p.Y-cy)
						if d < bestD {
							bestD = d
							best = &SnapCandidate{X: p.X, Y: p.Y, Type: stype, EntityID: entities[i].ID}
						}
					}
				}
			}
		} else {
			for _, e := range entities {
				cands := candidatesFor(e, stype, cx, cy)
				for k := range cands {
					d := math.Hypot(cands[k].X-cx, cands[k].Y-cy)
					if d < bestD {
						bestD = d
						c := cands[k]
						c.EntityID = e.ID
						best = &c
					}
				}
			}
		}

		if best != nil {
			return best
		}
	}
	return nil
}

// ─── Per-entity candidate generators ─────────────────────────────────────────

func candidatesFor(e document.Entity, stype SnapType, cx, cy float64) []SnapCandidate {
	switch stype {
	case SnapEndpoint:
		return endpointCandidates(e)
	case SnapMidpoint:
		return midpointCandidates(e)
	case SnapCenter:
		return centerCandidates(e)
	case SnapQuadrant:
		return quadrantCandidates(e)
	case SnapPerpendicular:
		return perpendicularCandidates(e, cx, cy)
	case SnapTangent:
		return tangentCandidates(e, cx, cy)
	case SnapNearest:
		return nearestCandidates(e, cx, cy)
	}
	return nil
}

// ── Endpoint ──────────────────────────────────────────────────────────────────

func endpointCandidates(e document.Entity) []SnapCandidate {
	switch e.Type {
	case document.TypeLine:
		return mkPts(SnapEndpoint, [2]float64{e.X1, e.Y1}, [2]float64{e.X2, e.Y2})

	case document.TypeArc:
		s := polarPt(e.CX, e.CY, e.R, e.StartDeg)
		en := polarPt(e.CX, e.CY, e.R, e.EndDeg)
		return mkPts(SnapEndpoint, s, en)

	case document.TypeRectangle:
		return mkPts(SnapEndpoint,
			[2]float64{e.X1, e.Y1}, [2]float64{e.X2, e.Y1},
			[2]float64{e.X2, e.Y2}, [2]float64{e.X1, e.Y2})

	case document.TypePolyline, document.TypeSpline:
		out := make([]SnapCandidate, 0, len(e.Points))
		for _, p := range e.Points {
			if len(p) >= 2 {
				out = append(out, SnapCandidate{X: p[0], Y: p[1], Type: SnapEndpoint})
			}
		}
		return out
	}
	return nil
}

// ── Midpoint ──────────────────────────────────────────────────────────────────

func midpointCandidates(e document.Entity) []SnapCandidate {
	switch e.Type {
	case document.TypeLine:
		return mkPts(SnapMidpoint, [2]float64{(e.X1 + e.X2) / 2, (e.Y1 + e.Y2) / 2})

	case document.TypeArc:
		span := e.EndDeg - e.StartDeg
		if span < 0 {
			span += 360
		}
		mid := polarPt(e.CX, e.CY, e.R, e.StartDeg+span/2)
		return mkPts(SnapMidpoint, mid)

	case document.TypeRectangle:
		return mkPts(SnapMidpoint,
			[2]float64{(e.X1 + e.X2) / 2, e.Y1},
			[2]float64{e.X2, (e.Y1 + e.Y2) / 2},
			[2]float64{(e.X1 + e.X2) / 2, e.Y2},
			[2]float64{e.X1, (e.Y1 + e.Y2) / 2})

	case document.TypePolyline, document.TypeSpline:
		out := make([]SnapCandidate, 0, len(e.Points))
		for i := 1; i < len(e.Points); i++ {
			if len(e.Points[i]) < 2 || len(e.Points[i-1]) < 2 {
				continue
			}
			out = append(out, SnapCandidate{
				X:    (e.Points[i][0] + e.Points[i-1][0]) / 2,
				Y:    (e.Points[i][1] + e.Points[i-1][1]) / 2,
				Type: SnapMidpoint,
			})
		}
		return out
	}
	return nil
}

// ── Center ────────────────────────────────────────────────────────────────────

func centerCandidates(e document.Entity) []SnapCandidate {
	switch e.Type {
	case document.TypeCircle, document.TypeArc:
		return mkPts(SnapCenter, [2]float64{e.CX, e.CY})
	case document.TypeEllipse:
		return mkPts(SnapCenter, [2]float64{e.CX, e.CY})
	case document.TypeRectangle:
		return mkPts(SnapCenter, [2]float64{(e.X1 + e.X2) / 2, (e.Y1 + e.Y2) / 2})
	}
	return nil
}

// ── Quadrant ──────────────────────────────────────────────────────────────────

func quadrantCandidates(e document.Entity) []SnapCandidate {
	switch e.Type {
	case document.TypeCircle:
		return mkPts(SnapQuadrant,
			[2]float64{e.CX + e.R, e.CY},
			[2]float64{e.CX, e.CY + e.R},
			[2]float64{e.CX - e.R, e.CY},
			[2]float64{e.CX, e.CY - e.R})

	case document.TypeArc:
		var out []SnapCandidate
		for _, deg := range []float64{0, 90, 180, 270} {
			if angleInArc(deg, e.StartDeg, e.EndDeg) {
				out = append(out, SnapCandidate{
					X:    e.CX + e.R*math.Cos(deg*math.Pi/180),
					Y:    e.CY + e.R*math.Sin(deg*math.Pi/180),
					Type: SnapQuadrant,
				})
			}
		}
		return out
	}
	return nil
}

// ── Perpendicular ─────────────────────────────────────────────────────────────

func perpendicularCandidates(e document.Entity, cx, cy float64) []SnapCandidate {
	switch e.Type {
	case document.TypeLine:
		fx, fy, ok := perpFoot(cx, cy, e.X1, e.Y1, e.X2, e.Y2)
		if !ok {
			return nil
		}
		return mkPts(SnapPerpendicular, [2]float64{fx, fy})

	case document.TypeRectangle:
		sides := [][4]float64{
			{e.X1, e.Y1, e.X2, e.Y1},
			{e.X2, e.Y1, e.X2, e.Y2},
			{e.X2, e.Y2, e.X1, e.Y2},
			{e.X1, e.Y2, e.X1, e.Y1},
		}
		var out []SnapCandidate
		for _, s := range sides {
			if fx, fy, ok := perpFoot(cx, cy, s[0], s[1], s[2], s[3]); ok {
				out = append(out, SnapCandidate{X: fx, Y: fy, Type: SnapPerpendicular})
			}
		}
		return out

	case document.TypePolyline:
		var out []SnapCandidate
		for i := 1; i < len(e.Points); i++ {
			if len(e.Points[i]) < 2 || len(e.Points[i-1]) < 2 {
				continue
			}
			if fx, fy, ok := perpFoot(cx, cy,
				e.Points[i-1][0], e.Points[i-1][1],
				e.Points[i][0], e.Points[i][1]); ok {
				out = append(out, SnapCandidate{X: fx, Y: fy, Type: SnapPerpendicular})
			}
		}
		return out
	}
	return nil
}

// ── Tangent ───────────────────────────────────────────────────────────────────

func tangentCandidates(e document.Entity, cx, cy float64) []SnapCandidate {
	switch e.Type {
	case document.TypeCircle:
		return tangentToCircle(cx, cy, e.CX, e.CY, e.R, 0, 360, false)
	case document.TypeArc:
		return tangentToCircle(cx, cy, e.CX, e.CY, e.R, e.StartDeg, e.EndDeg, true)
	}
	return nil
}

// tangentToCircle computes the two tangent points from cursor (px,py) to a
// circle of (cx,cy,r). checkArc restricts results to the arc angular range.
func tangentToCircle(px, py, cx, cy, r, startDeg, endDeg float64, checkArc bool) []SnapCandidate {
	dx, dy := px-cx, py-cy
	d := math.Hypot(dx, dy)
	if d <= r+1e-10 {
		return nil
	}
	ex, ey := dx/d, dy/d
	fx, fy := -ey, ex
	a := r * r / d
	b := r * math.Sqrt(d*d-r*r) / d

	var out []SnapCandidate
	for _, sign := range []float64{1, -1} {
		tx := cx + a*ex + sign*b*fx
		ty := cy + a*ey + sign*b*fy
		if checkArc {
			deg := math.Atan2(ty-cy, tx-cx) * 180 / math.Pi
			if !angleInArc(deg, startDeg, endDeg) {
				continue
			}
		}
		out = append(out, SnapCandidate{X: tx, Y: ty, Type: SnapTangent})
	}
	return out
}

// ── Nearest ───────────────────────────────────────────────────────────────────

func nearestCandidates(e document.Entity, cx, cy float64) []SnapCandidate {
	ge := e.ToGeometryEntity()
	if ge == nil {
		return nil
	}
	cp := ge.ClosestPoint(geometry.Point{X: cx, Y: cy})
	return mkPts(SnapNearest, [2]float64{cp.X, cp.Y})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// mkPts builds a []SnapCandidate from variadic [2]float64 points.
func mkPts(stype SnapType, ps ...[2]float64) []SnapCandidate {
	out := make([]SnapCandidate, len(ps))
	for i, p := range ps {
		out[i] = SnapCandidate{X: p[0], Y: p[1], Type: stype}
	}
	return out
}

// polarPt returns the point at angleDeg on a circle of given centre and radius.
func polarPt(cx, cy, r, deg float64) [2]float64 {
	rad := deg * math.Pi / 180
	return [2]float64{cx + r*math.Cos(rad), cy + r*math.Sin(rad)}
}

// perpFoot returns the foot of the perpendicular from P=(px,py) onto segment
// A=(ax,ay)–B=(bx,by). ok is false when the segment is degenerate or the foot
// would lie outside the segment (parameter t ∉ [0,1]).
func perpFoot(px, py, ax, ay, bx, by float64) (fx, fy float64, ok bool) {
	dx, dy := bx-ax, by-ay
	lenSq := dx*dx + dy*dy
	if lenSq < 1e-14 {
		return 0, 0, false
	}
	t := ((px-ax)*dx + (py-ay)*dy) / lenSq
	if t < 0 || t > 1 {
		return 0, 0, false
	}
	return ax + t*dx, ay + t*dy, true
}

// angleInArc returns true if angleDeg lies within the arc from startDeg to
// endDeg (CCW convention; handles wrapping through 0°).
func angleInArc(angleDeg, startDeg, endDeg float64) bool {
	a := normDeg(angleDeg)
	s := normDeg(startDeg)
	e := normDeg(endDeg)
	if math.Abs(s-e) < 1e-9 {
		return true // full circle
	}
	if s <= e {
		return a >= s && a <= e
	}
	return a >= s || a <= e // wraps through 0°
}

// normDeg normalises degrees to [0, 360).
func normDeg(d float64) float64 {
	d = math.Mod(d, 360)
	if d < 0 {
		d += 360
	}
	return d
}
