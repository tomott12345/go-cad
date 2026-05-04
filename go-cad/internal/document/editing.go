// editing.go — Task #4: CAD editing operations.
//
// Operations: Move, Copy, Rotate, Scale, Mirror, Trim, Extend,
// Fillet, Chamfer, ArrayRect, ArrayPolar.
//
// All mutating operations call d.pushUndo() once before modifying d.entities,
// so the entire operation is a single atomic undo step.
package document

import (
        "math"
        "sort"

        "go-cad/internal/geometry"
)

// ── Internal helpers ──────────────────────────────────────────────────────────

type xfmFn func(x, y float64) (float64, float64)

// normAngleDeg normalises an angle (degrees) to [0, 360).
func normAngleDeg(deg float64) float64 {
        deg = math.Mod(deg, 360)
        if deg < 0 {
                deg += 360
        }
        return deg
}

// deepCopyEntity returns a deep copy of e (copies all slice fields so that
// the returned entity does not share backing arrays with the original).
func deepCopyEntity(e Entity) Entity {
        if len(e.Points) > 0 {
                pts := make([][]float64, len(e.Points))
                for i, p := range e.Points {
                        cp := make([]float64, len(p))
                        copy(cp, p)
                        pts[i] = cp
                }
                e.Points = pts
        }
        if len(e.Knots) > 0 {
                k := make([]float64, len(e.Knots))
                copy(k, e.Knots)
                e.Knots = k
        }
        if len(e.Weights) > 0 {
                w := make([]float64, len(e.Weights))
                copy(w, e.Weights)
                e.Weights = w
        }
        return e
}

// applyXfm transforms all geometric coordinates of e by xfm and scales all
// radii by radScale (1.0 = unchanged; use |sx| for uniform scale).
// The entity is deep-copied first, so the original is never modified.
func applyXfm(e Entity, xfm xfmFn, radScale float64) Entity {
        e = deepCopyEntity(e)

        // Transform Points slice (polyline, spline, NURBS control points).
        for i := range e.Points {
                if len(e.Points[i]) >= 2 {
                        e.Points[i][0], e.Points[i][1] = xfm(e.Points[i][0], e.Points[i][1])
                }
        }

        switch e.Type {
        case TypeLine, TypeRectangle:
                e.X1, e.Y1 = xfm(e.X1, e.Y1)
                e.X2, e.Y2 = xfm(e.X2, e.Y2)
        case TypeCircle:
                e.CX, e.CY = xfm(e.CX, e.CY)
                e.R *= radScale
        case TypeArc:
                e.CX, e.CY = xfm(e.CX, e.CY)
                e.R *= radScale
        case TypeEllipse:
                e.CX, e.CY = xfm(e.CX, e.CY)
                e.R *= radScale
                e.R2 *= radScale
        case TypeText, TypeMText:
                e.X1, e.Y1 = xfm(e.X1, e.Y1)
        case TypeDimLinear, TypeDimAligned:
                e.X1, e.Y1 = xfm(e.X1, e.Y1)
                e.X2, e.Y2 = xfm(e.X2, e.Y2)
        case TypeDimAngular:
                e.CX, e.CY = xfm(e.CX, e.CY)
                e.X1, e.Y1 = xfm(e.X1, e.Y1)
                e.X2, e.Y2 = xfm(e.X2, e.Y2)
                e.R *= radScale
        case TypeDimRadial, TypeDimDiameter:
                e.CX, e.CY = xfm(e.CX, e.CY)
                e.R *= radScale
        }
        return e
}

// entitiesByIDs returns deep copies of entities whose IDs appear in ids,
// preserving document order.
func (d *Document) entitiesByIDs(ids []int) []Entity {
        set := make(map[int]bool, len(ids))
        for _, id := range ids {
                set[id] = true
        }
        var out []Entity
        for _, e := range d.entities {
                if set[e.ID] {
                        out = append(out, deepCopyEntity(e))
                }
        }
        return out
}

// appendNew assigns a new ID to e, appends it to d.entities, and returns the
// ID. Must only be called after d.pushUndo() has already been invoked.
func (d *Document) appendNew(e Entity) int {
        e.ID = d.nextID
        d.nextID++
        if e.Color == "" {
                e.Color = "#ffffff"
        }
        d.entities = append(d.entities, e)
        return e.ID
}

// removeByID removes the first entity matching id (no undo push — the caller
// must have already called d.pushUndo()).
func (d *Document) removeByID(id int) {
        for i, e := range d.entities {
                if e.ID == id {
                        d.entities = append(d.entities[:i], d.entities[i+1:]...)
                        return
                }
        }
}

// applyTransformOp applies xfm + radScale + optional postFn to the entities
// identified by ids. When makeCopy is true the originals are left intact and
// the transformed copies are returned; otherwise the entities are modified in
// place. Returns the IDs of the affected (or copied) entities.
func (d *Document) applyTransformOp(ids []int, xfm xfmFn, radScale float64, makeCopy bool, postFn func(*Entity)) []int {
        if len(ids) == 0 {
                return nil
        }
        src := d.entitiesByIDs(ids)
        if len(src) == 0 {
                return nil
        }
        d.pushUndo()
        newIDs := make([]int, 0, len(src))
        if makeCopy {
                for _, e := range src {
                        ne := applyXfm(e, xfm, radScale)
                        if postFn != nil {
                                postFn(&ne)
                        }
                        newIDs = append(newIDs, d.appendNew(ne))
                }
        } else {
                idSet := make(map[int]bool, len(ids))
                for _, id := range ids {
                        idSet[id] = true
                }
                for i := range d.entities {
                        if idSet[d.entities[i].ID] {
                                ne := applyXfm(d.entities[i], xfm, radScale)
                                if postFn != nil {
                                        postFn(&ne)
                                }
                                d.entities[i] = ne
                                newIDs = append(newIDs, ne.ID)
                        }
                }
        }
        return newIDs
}

// ── Move ──────────────────────────────────────────────────────────────────────

// Move translates entities by (dx, dy). Returns true if at least one entity
// was moved.
// Command: M
func (d *Document) Move(ids []int, dx, dy float64) bool {
        xfm := func(x, y float64) (float64, float64) { return x + dx, y + dy }
        out := d.applyTransformOp(ids, xfm, 1, false, nil)
        return len(out) > 0
}

// ── Copy ──────────────────────────────────────────────────────────────────────

// Copy duplicates entities and translates the copies by (dx, dy).
// Returns the IDs of the new copies.
// Command: CP / CO
func (d *Document) Copy(ids []int, dx, dy float64) []int {
        xfm := func(x, y float64) (float64, float64) { return x + dx, y + dy }
        return d.applyTransformOp(ids, xfm, 1, true, nil)
}

// ── Rotate ────────────────────────────────────────────────────────────────────

// Rotate rotates entities around pivot (cx, cy) by angleDeg degrees CCW.
// If makeCopy is true, rotates copies and leaves the originals unchanged.
// Returns the IDs of the rotated entities (or new copies).
// Command: RO
func (d *Document) Rotate(ids []int, cx, cy, angleDeg float64, makeCopy bool) []int {
        rad := angleDeg * math.Pi / 180
        cos, sin := math.Cos(rad), math.Sin(rad)
        xfm := func(x, y float64) (float64, float64) {
                dx, dy := x-cx, y-cy
                return cx + dx*cos - dy*sin, cy + dx*sin + dy*cos
        }
        return d.applyTransformOp(ids, xfm, 1, makeCopy, func(e *Entity) {
                switch e.Type {
                case TypeArc:
                        e.StartDeg = normAngleDeg(e.StartDeg + angleDeg)
                        e.EndDeg = normAngleDeg(e.EndDeg + angleDeg)
                case TypeEllipse, TypeText, TypeMText:
                        e.RotDeg += angleDeg
                }
        })
}

// ── Scale ─────────────────────────────────────────────────────────────────────

// Scale scales entities from base point (cx, cy) by factors (sx, sy).
// For non-uniform scale the radius of circles/arcs is averaged (geometric
// mean of |sx| and |sy|). If makeCopy is true, scales copies.
// Returns the IDs of the scaled entities (or new copies).
// Command: SC
func (d *Document) Scale(ids []int, cx, cy, sx, sy float64, makeCopy bool) []int {
        radScale := math.Sqrt(math.Abs(sx) * math.Abs(sy))
        xfm := func(x, y float64) (float64, float64) {
                return cx + (x-cx)*sx, cy + (y-cy)*sy
        }
        return d.applyTransformOp(ids, xfm, radScale, makeCopy, nil)
}

// ── Mirror ────────────────────────────────────────────────────────────────────

// Mirror reflects entities across the line defined by (ax, ay)→(bx, by).
// If makeCopy is true, mirrors copies and leaves the originals unchanged.
// Returns the IDs of the mirrored entities (or new copies).
// Command: MI
func (d *Document) Mirror(ids []int, ax, ay, bx, by float64, makeCopy bool) []int {
        dx, dy := bx-ax, by-ay
        len2 := dx*dx + dy*dy
        if len2 < 1e-20 {
                return nil
        }
        lineAngle := math.Atan2(dy, dx) * 180 / math.Pi
        xfm := func(x, y float64) (float64, float64) {
                t := ((x-ax)*dx + (y-ay)*dy) / len2
                fx := ax + t*dx
                fy := ay + t*dy
                return 2*fx - x, 2*fy - y
        }
        return d.applyTransformOp(ids, xfm, 1, makeCopy, func(e *Entity) {
                switch e.Type {
                case TypeArc:
                        // Mirror the start and end angles across the mirror line, then swap
                        // them to maintain the CCW arc convention (mirroring reverses chirality).
                        newS := normAngleDeg(2*lineAngle - e.StartDeg)
                        newE := normAngleDeg(2*lineAngle - e.EndDeg)
                        e.StartDeg, e.EndDeg = newE, newS
                case TypeEllipse, TypeText, TypeMText:
                        e.RotDeg = 2*lineAngle - e.RotDeg
                }
        })
}

// ── Trim (at intersection) ────────────────────────────────────────────────────

// Trim cuts entity targetID at its intersections with cutterID, removing the
// portion nearest to the pick point (pickX, pickY). The original entity is
// replaced by up to two surviving sub-entities; their IDs are returned.
// Returns nil if no trimming is possible (no intersections, or unsupported type).
// Command: TR
func (d *Document) Trim(cutterID, targetID int, pickX, pickY float64) []int {
        if cutterID == targetID {
                return nil
        }
        pick := geometry.Point{X: pickX, Y: pickY}

        var cutter, target Entity
        cutterFound, targetFound := false, false
        for _, e := range d.entities {
                switch e.ID {
                case cutterID:
                        cutter = e
                        cutterFound = true
                case targetID:
                        target = e
                        targetFound = true
                }
        }
        if !cutterFound || !targetFound {
                return nil
        }

        ge := target.ToGeometryEntity()
        if ge == nil {
                return nil
        }

        pts := target.IntersectWith(cutter)
        if len(pts) == 0 {
                return nil
        }

        // Compute parametric t values for each intersection point along the target.
        tVals := make([]float64, 0, len(pts))
        for _, ip := range pts {
                t := paramAtGeomEntity(ge, ip)
                if t > 1e-6 && t < 1-1e-6 {
                        tVals = append(tVals, t)
                }
        }
        if len(tVals) == 0 {
                return nil
        }
        sort.Float64s(tVals)

        // Build interval boundaries and find which one contains the pick point.
        boundaries := make([]float64, 0, len(tVals)+2)
        boundaries = append(boundaries, 0)
        boundaries = append(boundaries, tVals...)
        boundaries = append(boundaries, 1)

        tPick := paramAtGeomEntity(ge, pick)
        pickInterval := len(boundaries) - 2
        for j := 0; j < len(boundaries)-1; j++ {
                if tPick >= boundaries[j] && tPick <= boundaries[j+1] {
                        pickInterval = j
                        break
                }
        }

        // Split the geometry entity at all tVals and keep the non-trimmed segments.
        subEntities := splitGeomAt(ge, tVals)

        d.pushUndo()
        layer, color := target.Layer, target.Color
        d.removeByID(targetID)

        var newIDs []int
        for j, sub := range subEntities {
                if j == pickInterval {
                        continue
                }
                if docE := GeometryEntityToDocument(sub, layer, color); docE != nil {
                        newIDs = append(newIDs, d.appendNew(*docE))
                }
        }
        return newIDs
}

// paramAtGeomEntity returns the parametric position t ∈ [0,1] of point p
// along the geometry entity ge.
func paramAtGeomEntity(ge geometry.Entity, p geometry.Point) float64 {
        switch v := ge.(type) {
        case geometry.SegmentEntity:
                _, t := v.Segment.ClosestPoint(p)
                return t

        case geometry.ArcEntity:
                angle := math.Atan2(p.Y-v.Center.Y, p.X-v.Center.X) * 180 / math.Pi
                span := v.EndDeg - v.StartDeg
                if span <= 0 {
                        span += 360
                }
                d := angle - v.StartDeg
                for d < 0 {
                        d += 360
                }
                for d > 360 {
                        d -= 360
                }
                t := d / span
                if t < 0 {
                        t = 0
                }
                if t > 1 {
                        t = 1
                }
                return t

        case geometry.CircleEntity:
                angle := math.Atan2(p.Y-v.Center.Y, p.X-v.Center.X) * 180 / math.Pi
                if angle < 0 {
                        angle += 360
                }
                return angle / 360

        case geometry.PolylineEntity:
                totalLen := v.Polyline.Length()
                if totalLen < 1e-12 {
                        return 0
                }
                accum := 0.0
                for i := 0; i < v.Polyline.NumSegments(); i++ {
                        seg := v.Polyline.Segment(i)
                        segLen := seg.Length()
                        cp, t := seg.ClosestPoint(p)
                        if p.Dist(cp) < 1e-4 {
                                return (accum + t*segLen) / totalLen
                        }
                        accum += segLen
                }
                // Fallback: pick the parameter nearest to an endpoint.
                if len(v.Polyline.Points) > 0 {
                        start := v.Polyline.Points[0]
                        end := v.Polyline.Points[len(v.Polyline.Points)-1]
                        if p.Dist(start) <= p.Dist(end) {
                                return 0
                        }
                }
                return 1
        }
        return 0.5
}

// splitGeomAt splits a geometry entity at a sorted list of parametric t
// values, returning len(tVals)+1 sub-entities.
func splitGeomAt(ge geometry.Entity, tVals []float64) []geometry.Entity {
        if len(tVals) == 0 {
                return []geometry.Entity{ge}
        }
        var result []geometry.Entity
        current := ge
        cumT := 0.0
        for _, t := range tVals {
                if t <= cumT+1e-9 || t >= 1-1e-9 {
                        continue
                }
                localT := (t - cumT) / (1.0 - cumT)
                if localT <= 1e-6 || localT >= 1-1e-6 {
                        continue
                }
                left, right := current.TrimAt(localT)
                result = append(result, left)
                current = right
                cumT = t
        }
        result = append(result, current)
        return result
}

// ── Extend ────────────────────────────────────────────────────────────────────

// Extend extends entity targetID to meet boundaryID. The pick point (pickX,
// pickY) indicates which end of the target to extend. Returns the ID of the
// updated entity, or -1 on failure.
// Currently supports TypeLine and TypeArc targets.
// Command: EX
func (d *Document) Extend(boundaryID, targetID int, pickX, pickY float64) int {
        if boundaryID == targetID {
                return -1
        }
        pick := geometry.Point{X: pickX, Y: pickY}

        var boundary, target Entity
        bFound, tFound := false, false
        for _, e := range d.entities {
                switch e.ID {
                case boundaryID:
                        boundary = e
                        bFound = true
                case targetID:
                        target = e
                        tFound = true
                }
        }
        if !bFound || !tFound {
                return -1
        }

        switch target.Type {
        case TypeLine:
                return d.extendLine(target, boundary, pick)
        case TypeArc:
                return d.extendArc(target, boundary, pick)
        default:
                return -1
        }
}

func (d *Document) extendLine(target, boundary Entity, pick geometry.Point) int {
        start := geometry.Point{X: target.X1, Y: target.Y1}
        end := geometry.Point{X: target.X2, Y: target.Y2}

        bge := boundary.ToGeometryEntity()
        if bge == nil {
                return -1
        }

        // Intersect the infinite extension of the target with the boundary.
        infLine := geometry.LineEntity{Line: geometry.Line{P: start, Q: end}}
        pts := geometry.Intersect(infLine, bge)
        if len(pts) == 0 {
                return -1
        }

        // Use the intersection closest to the pick point.
        best := pts[0]
        for _, p := range pts[1:] {
                if pick.Dist(p) < pick.Dist(best) {
                        best = p
                }
        }

        newLine := Entity{
                Type:  TypeLine,
                Layer: target.Layer, Color: target.Color,
        }
        if pick.Dist(start) <= pick.Dist(end) {
                newLine.X1, newLine.Y1 = best.X, best.Y
                newLine.X2, newLine.Y2 = target.X2, target.Y2
        } else {
                newLine.X1, newLine.Y1 = target.X1, target.Y1
                newLine.X2, newLine.Y2 = best.X, best.Y
        }

        d.pushUndo()
        d.removeByID(target.ID)
        return d.appendNew(newLine)
}

func (d *Document) extendArc(target, boundary Entity, pick geometry.Point) int {
        center := geometry.Point{X: target.CX, Y: target.CY}
        startPt := geometry.Point{
                X: target.CX + target.R*math.Cos(target.StartDeg*math.Pi/180),
                Y: target.CY + target.R*math.Sin(target.StartDeg*math.Pi/180),
        }
        endPt := geometry.Point{
                X: target.CX + target.R*math.Cos(target.EndDeg*math.Pi/180),
                Y: target.CY + target.R*math.Sin(target.EndDeg*math.Pi/180),
        }
        extendStart := pick.Dist(startPt) <= pick.Dist(endPt)

        bge := boundary.ToGeometryEntity()
        if bge == nil {
                return -1
        }

        // Intersect the full circle with the boundary.
        fullCircle := geometry.CircleEntity{Circle: geometry.Circle{Center: center, Radius: target.R}}
        pts := geometry.Intersect(fullCircle, bge)
        if len(pts) == 0 {
                return -1
        }

        // Closest intersection to pick point.
        best := pts[0]
        for _, p := range pts[1:] {
                if pick.Dist(p) < pick.Dist(best) {
                        best = p
                }
        }
        newAngle := normAngleDeg(math.Atan2(best.Y-center.Y, best.X-center.X) * 180 / math.Pi)

        newArc := Entity{
                Type: TypeArc, CX: target.CX, CY: target.CY, R: target.R,
                StartDeg: target.StartDeg, EndDeg: target.EndDeg,
                Layer: target.Layer, Color: target.Color,
        }
        if extendStart {
                newArc.StartDeg = newAngle
        } else {
                newArc.EndDeg = newAngle
        }

        d.pushUndo()
        d.removeByID(target.ID)
        return d.appendNew(newArc)
}

// ── Fillet ────────────────────────────────────────────────────────────────────

// Fillet creates a tangent arc of the given radius connecting two line
// entities (id1 and id2), trimming them to the tangent points. Returns the ID
// of the new arc, or -1 on failure. Both entities must be TypeLine.
// Command: F
func (d *Document) Fillet(id1, id2 int, radius float64) int {
        if id1 == id2 || radius <= 0 {
                return -1
        }
        var e1, e2 Entity
        e1Found, e2Found := false, false
        for _, e := range d.entities {
                switch e.ID {
                case id1:
                        e1 = e
                        e1Found = true
                case id2:
                        e2 = e
                        e2Found = true
                }
        }
        if !e1Found || !e2Found || e1.Type != TypeLine || e2.Type != TypeLine {
                return -1
        }

        s1 := geometry.Segment{
                Start: geometry.Point{X: e1.X1, Y: e1.Y1},
                End:   geometry.Point{X: e1.X2, Y: e1.Y2},
        }
        s2 := geometry.Segment{
                Start: geometry.Point{X: e2.X1, Y: e2.Y1},
                End:   geometry.Point{X: e2.X2, Y: e2.Y2},
        }

        // Find intersection of the infinite lines.
        corners := geometry.IntersectLines(
                geometry.Line{P: s1.Start, Q: s1.End},
                geometry.Line{P: s2.Start, Q: s2.End},
        )
        if len(corners) == 0 {
                return -1 // parallel
        }
        corner := corners[0]

        // Angle between the two lines.
        d1 := s1.End.Sub(s1.Start).Normalize()
        d2 := s2.End.Sub(s2.Start).Normalize()
        dot := math.Max(-1, math.Min(1, d1.Dot(d2)))
        angle := math.Acos(math.Abs(dot))
        if angle < 1e-6 {
                return -1
        }

        // Tangent distance from the corner to each tangent point.
        tangentLen := radius / math.Tan(angle/2)

        // Unit vectors from the corner toward the body of each segment.
        c1 := segDirFromCorner(s1, corner)
        c2 := segDirFromCorner(s2, corner)
        t1 := corner.Add(c1.Scale(tangentLen))
        t2 := corner.Add(c2.Scale(tangentLen))

        // Arc centre: on the bisector, at distance r/sin(angle/2) from corner.
        bisector := c1.Add(c2).Normalize()
        hyp := radius / math.Sin(angle/2)
        arcCenter := corner.Add(bisector.Scale(hyp))

        // Arc start/end angles (CCW convention).
        startAngle := normAngleDeg(math.Atan2(t1.Y-arcCenter.Y, t1.X-arcCenter.X) * 180 / math.Pi)
        endAngle := normAngleDeg(math.Atan2(t2.Y-arcCenter.Y, t2.X-arcCenter.X) * 180 / math.Pi)
        // Ensure we take the shorter arc.
        span := endAngle - startAngle
        if span < 0 {
                span += 360
        }
        if span > 180 {
                startAngle, endAngle = endAngle, startAngle
        }

        d.pushUndo()
        trimSegToPoint(d, &e1, t1, corner)
        trimSegToPoint(d, &e2, t2, corner)
        return d.appendNew(Entity{
                Type:     TypeArc,
                CX:       arcCenter.X, CY: arcCenter.Y, R: radius,
                StartDeg: normAngleDeg(startAngle),
                EndDeg:   normAngleDeg(endAngle),
                Layer:    e1.Layer, Color: e1.Color,
        })
}

// ── Chamfer ───────────────────────────────────────────────────────────────────

// Chamfer creates a chamfer line between two line entities, trimming each by
// dist1 and dist2 respectively from their common corner. Returns the ID of the
// chamfer line, or -1 on failure. Both entities must be TypeLine.
// Command: CHA
func (d *Document) Chamfer(id1, id2 int, dist1, dist2 float64) int {
        if id1 == id2 || dist1 < 0 || dist2 < 0 {
                return -1
        }
        var e1, e2 Entity
        e1Found, e2Found := false, false
        for _, e := range d.entities {
                switch e.ID {
                case id1:
                        e1 = e
                        e1Found = true
                case id2:
                        e2 = e
                        e2Found = true
                }
        }
        if !e1Found || !e2Found || e1.Type != TypeLine || e2.Type != TypeLine {
                return -1
        }

        s1 := geometry.Segment{
                Start: geometry.Point{X: e1.X1, Y: e1.Y1},
                End:   geometry.Point{X: e1.X2, Y: e1.Y2},
        }
        s2 := geometry.Segment{
                Start: geometry.Point{X: e2.X1, Y: e2.Y1},
                End:   geometry.Point{X: e2.X2, Y: e2.Y2},
        }

        corners := geometry.IntersectLines(
                geometry.Line{P: s1.Start, Q: s1.End},
                geometry.Line{P: s2.Start, Q: s2.End},
        )
        if len(corners) == 0 {
                return -1
        }
        corner := corners[0]

        c1 := segDirFromCorner(s1, corner)
        c2 := segDirFromCorner(s2, corner)
        t1 := corner.Add(c1.Scale(dist1))
        t2 := corner.Add(c2.Scale(dist2))

        d.pushUndo()
        trimSegToPoint(d, &e1, t1, corner)
        trimSegToPoint(d, &e2, t2, corner)
        return d.appendNew(Entity{
                Type:  TypeLine,
                X1: t1.X, Y1: t1.Y,
                X2: t2.X, Y2: t2.Y,
                Layer: e1.Layer, Color: e1.Color,
        })
}

// segDirFromCorner returns the unit direction from corner toward the interior
// of seg (away from the corner endpoint).
func segDirFromCorner(seg geometry.Segment, corner geometry.Point) geometry.Point {
        toStart := seg.Start.Sub(corner)
        toEnd := seg.End.Sub(corner)
        if toStart.Len() >= toEnd.Len() {
                return toStart.Normalize()
        }
        return toEnd.Normalize()
}

// trimSegToPoint updates the entity with e.ID in d.entities so that the
// endpoint nearest to corner is moved to newEnd. Must be called after
// d.pushUndo() has already been invoked.
func trimSegToPoint(d *Document, e *Entity, newEnd, corner geometry.Point) {
        start := geometry.Point{X: e.X1, Y: e.Y1}
        end := geometry.Point{X: e.X2, Y: e.Y2}
        for i := range d.entities {
                if d.entities[i].ID == e.ID {
                        if start.Dist(corner) <= end.Dist(corner) {
                                d.entities[i].X1, d.entities[i].Y1 = newEnd.X, newEnd.Y
                        } else {
                                d.entities[i].X2, d.entities[i].Y2 = newEnd.X, newEnd.Y
                        }
                        return
                }
        }
}

// ── Offset ────────────────────────────────────────────────────────────────────

// Offset creates parallel copies of entities at signed distance dist
// (positive = left/outward; negative = right/inward). Unlike the legacy
// OffsetEntity helper, this version treats all ids as a single atomic undo step.
// Returns the IDs of the newly created offset entities.
// Command: OFFSET / O
func (d *Document) Offset(ids []int, dist float64) []int {
        if len(ids) == 0 || dist == 0 {
                return nil
        }
        src := d.entitiesByIDs(ids)
        if len(src) == 0 {
                return nil
        }

        // Pre-compute all offset entities before touching d.entities.
        type offsetResult struct{ e Entity }
        results := make([]offsetResult, 0, len(src))
        for _, e := range src {
                off := e.Offset(dist)
                if off != nil {
                        results = append(results, offsetResult{*off})
                }
        }
        if len(results) == 0 {
                return nil
        }

        d.pushUndo()
        newIDs := make([]int, 0, len(results))
        for _, r := range results {
                newIDs = append(newIDs, d.appendNew(r.e))
        }
        return newIDs
}

// ── ArrayRect ─────────────────────────────────────────────────────────────────

// ArrayRect creates a rectangular array of copies.
// rows × cols is the total grid size; the originals occupy row 0, col 0.
// rowSpacing and colSpacing are the centre-to-centre distances along Y and X.
// Returns the IDs of all new copies (originals are not included).
// Command: ARRAYRECT
func (d *Document) ArrayRect(ids []int, rows, cols int, rowSpacing, colSpacing float64) []int {
        src := d.entitiesByIDs(ids)
        if len(src) == 0 || rows < 1 || cols < 1 {
                return nil
        }
        d.pushUndo()
        var newIDs []int
        for row := 0; row < rows; row++ {
                for col := 0; col < cols; col++ {
                        if row == 0 && col == 0 {
                                continue // skip originals
                        }
                        dx := float64(col) * colSpacing
                        dy := float64(row) * rowSpacing
                        xfm := func(x, y float64) (float64, float64) { return x + dx, y + dy }
                        for _, e := range src {
                                ne := applyXfm(e, xfm, 1)
                                newIDs = append(newIDs, d.appendNew(ne))
                        }
                }
        }
        return newIDs
}

// ── ArrayPolar ────────────────────────────────────────────────────────────────

// ArrayPolar creates a polar (circular) array of count copies (including the
// originals) distributed over totalAngleDeg around centre (cx, cy).
// Returns the IDs of all new copies (originals are not included).
// Command: ARRAYPOLAR
func (d *Document) ArrayPolar(ids []int, cx, cy float64, count int, totalAngleDeg float64) []int {
        src := d.entitiesByIDs(ids)
        if len(src) == 0 || count < 2 {
                return nil
        }
        d.pushUndo()
        step := totalAngleDeg / float64(count)
        var newIDs []int
        for i := 1; i < count; i++ {
                angleDeg := float64(i) * step
                rad := angleDeg * math.Pi / 180
                cos, sin := math.Cos(rad), math.Sin(rad)
                xfm := func(x, y float64) (float64, float64) {
                        ddx, ddy := x-cx, y-cy
                        return cx + ddx*cos - ddy*sin, cy + ddx*sin + ddy*cos
                }
                for _, e := range src {
                        ne := applyXfm(e, xfm, 1)
                        if ne.Type == TypeArc {
                                ne.StartDeg = normAngleDeg(ne.StartDeg + angleDeg)
                                ne.EndDeg = normAngleDeg(ne.EndDeg + angleDeg)
                        }
                        if ne.Type == TypeEllipse || ne.Type == TypeText || ne.Type == TypeMText {
                                ne.RotDeg += angleDeg
                        }
                        newIDs = append(newIDs, d.appendNew(ne))
                }
        }
        return newIDs
}
