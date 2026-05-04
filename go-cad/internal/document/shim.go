// Package document bridges the legacy flat Entity model with the new
// internal/geometry typed primitives. Both representations are kept in sync:
// callers working with the document API use the existing Entity struct,
// while the geometry engine works with the richer typed forms.
package document

import (
        "go-cad/internal/geometry"
)

// ToGeometryEntity converts a document Entity to the typed geometry.Entity interface.
// Returns nil if the entity type is not recognised.
func (e Entity) ToGeometryEntity() geometry.Entity {
        switch e.Type {
        case TypeLine:
                return geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{X: e.X1, Y: e.Y1},
                        End:   geometry.Point{X: e.X2, Y: e.Y2},
                }}

        case TypeCircle:
                return geometry.CircleEntity{Circle: geometry.Circle{
                        Center: geometry.Point{X: e.CX, Y: e.CY},
                        Radius: e.R,
                }}

        case TypeArc:
                return geometry.ArcEntity{Arc: geometry.Arc{
                        Center:   geometry.Point{X: e.CX, Y: e.CY},
                        Radius:   e.R,
                        StartDeg: e.StartDeg,
                        EndDeg:   e.EndDeg,
                }}

        case TypeRectangle:
                // Rectangle represented as a closed polyline of 4 corners
                x1, y1, x2, y2 := e.X1, e.Y1, e.X2, e.Y2
                return geometry.PolylineEntity{Polyline: geometry.Polyline{
                        Points: []geometry.Point{
                                {X: x1, Y: y1},
                                {X: x2, Y: y1},
                                {X: x2, Y: y2},
                                {X: x1, Y: y2},
                        },
                        Closed: true,
                }}

        case TypePolyline:
                pts := make([]geometry.Point, len(e.Points))
                for i, p := range e.Points {
                        if len(p) >= 2 {
                                pts[i] = geometry.Point{X: p[0], Y: p[1]}
                        }
                }
                return geometry.PolylineEntity{Polyline: geometry.Polyline{Points: pts}}

        case TypeSpline:
                if len(e.Points) < 4 {
                        return nil // need at least one cubic segment (4 control points)
                }
                pts := make([]geometry.Point, len(e.Points))
                for i, p := range e.Points {
                        if len(p) >= 2 {
                                pts[i] = geometry.Point{X: p[0], Y: p[1]}
                        }
                }
                return geometry.BezierEntity{BezierSpline: geometry.NewBezierSpline(pts)}

        case TypeEllipse:
                return geometry.EllipseEntity{Ellipse: geometry.NewEllipse(e.CX, e.CY, e.R, e.R2, e.RotDeg)}

        case TypeText:
                // Text has no geometric boundary for snap/intersect; return nil.
                return nil

        case TypeDimLinear, TypeDimAligned:
                // Treat the measured segment as the geometry for nearest-entity snap.
                return geometry.SegmentEntity{Segment: geometry.Segment{
                        Start: geometry.Point{X: e.X1, Y: e.Y1},
                        End:   geometry.Point{X: e.X2, Y: e.Y2},
                }}

        case TypeDimAngular, TypeDimRadial, TypeDimDiameter:
                // Angular/radial/diameter dims snap to their centre point only — no
                // geometry representative is returned; they fall back to the point test.
                return nil

        default:
                return nil
        }
}

// BoundingBox returns the axis-aligned bounding box of the entity using the geometry engine.
func (e Entity) BoundingBox() geometry.BBox {
        ge := e.ToGeometryEntity()
        if ge == nil {
                return geometry.EmptyBBox()
        }
        return ge.BoundingBox()
}

// ClosestPoint returns the nearest point on the entity to p using the geometry engine.
func (e Entity) ClosestPoint(p geometry.Point) geometry.Point {
        ge := e.ToGeometryEntity()
        if ge == nil {
                return p
        }
        return ge.ClosestPoint(p)
}

// Offset returns a new Entity offset by dist (via the geometry engine).
// For types that produce a Polyline on offset, the entity is returned as TypePolyline.
func (e Entity) Offset(dist float64) *Entity {
        ge := e.ToGeometryEntity()
        if ge == nil {
                return nil
        }
        off := ge.Offset(dist)
        return GeometryEntityToDocument(off, e.Layer, e.Color)
}

// GeometryEntityToDocument converts a geometry.Entity back to a document Entity.
// The resulting entity has ID=0 (caller should assign).
func GeometryEntityToDocument(ge geometry.Entity, layer int, color string) *Entity {
        if ge == nil {
                return nil
        }
        switch v := ge.(type) {
        case geometry.SegmentEntity:
                return &Entity{
                        Type:  TypeLine,
                        Layer: layer, Color: color,
                        X1: v.Start.X, Y1: v.Start.Y,
                        X2: v.End.X, Y2: v.End.Y,
                }
        case geometry.CircleEntity:
                return &Entity{
                        Type:  TypeCircle,
                        Layer: layer, Color: color,
                        CX: v.Center.X, CY: v.Center.Y, R: v.Radius,
                }
        case geometry.ArcEntity:
                return &Entity{
                        Type:     TypeArc,
                        Layer:    layer, Color: color,
                        CX: v.Center.X, CY: v.Center.Y, R: v.Radius,
                        StartDeg: v.StartDeg, EndDeg: v.EndDeg,
                }
        case geometry.PolylineEntity:
                pts := make([][]float64, len(v.Points))
                for i, p := range v.Points {
                        pts[i] = []float64{p.X, p.Y}
                }
                return &Entity{
                        Type:   TypePolyline,
                        Layer:  layer, Color: color,
                        Points: pts,
                }
        case geometry.EllipseEntity:
                // Preserve ellipse parameters if the geometry values are accessible.
                return &Entity{
                        Type:   TypeEllipse,
                        Layer:  layer, Color: color,
                        CX: v.Center.X, CY: v.Center.Y,
                        R: v.A, R2: v.B, RotDeg: v.Rotation,
                }

        case geometry.BezierEntity:
                pts := make([][]float64, len(v.Controls))
                for i, p := range v.Controls {
                        pts[i] = []float64{p.X, p.Y}
                }
                return &Entity{
                        Type:   TypeSpline,
                        Layer:  layer, Color: color,
                        Points: pts,
                }

        default:
                return nil
        }
}

// IntersectWith returns the intersection points between this entity and another,
// using the geometry engine.
func (e Entity) IntersectWith(other Entity) []geometry.Point {
        ga := e.ToGeometryEntity()
        gb := other.ToGeometryEntity()
        if ga == nil || gb == nil {
                return nil
        }
        return intersectEntities(ga, gb)
}

// intersectEntities dispatches intersection between two geometry.Entity values.
func intersectEntities(a, b geometry.Entity) []geometry.Point {
        switch av := a.(type) {
        case geometry.SegmentEntity:
                return intersectSegmentWith(av.Segment, b)
        case geometry.CircleEntity:
                return intersectCircleWith(av.Circle, b)
        case geometry.ArcEntity:
                return intersectArcWith(av.Arc, b)
        case geometry.PolylineEntity:
                return intersectPolylineWith(av.Polyline, b)
        }
        // Fallback: swap and try the other order
        switch bv := b.(type) {
        case geometry.SegmentEntity:
                return intersectSegmentWith(bv.Segment, a)
        }
        return nil
}

func intersectSegmentWith(s geometry.Segment, b geometry.Entity) []geometry.Point {
        switch bv := b.(type) {
        case geometry.SegmentEntity:
                return geometry.IntersectSegments(s, bv.Segment)
        case geometry.CircleEntity:
                return geometry.IntersectSegmentCircle(s, bv.Circle)
        case geometry.ArcEntity:
                return geometry.IntersectSegmentArc(s, bv.Arc)
        case geometry.PolylineEntity:
                return geometry.IntersectSegmentPolyline(s, bv.Polyline)
        case geometry.EllipseEntity:
                return geometry.IntersectSegmentEllipse(s, bv.Ellipse)
        }
        return nil
}

func intersectCircleWith(c geometry.Circle, b geometry.Entity) []geometry.Point {
        switch bv := b.(type) {
        case geometry.SegmentEntity:
                return geometry.IntersectSegmentCircle(bv.Segment, c)
        case geometry.CircleEntity:
                return geometry.IntersectCircles(c, bv.Circle)
        case geometry.ArcEntity:
                return geometry.IntersectCircleArc(c, bv.Arc)
        }
        return nil
}

func intersectArcWith(a geometry.Arc, b geometry.Entity) []geometry.Point {
        switch bv := b.(type) {
        case geometry.SegmentEntity:
                return geometry.IntersectSegmentArc(bv.Segment, a)
        case geometry.CircleEntity:
                return geometry.IntersectCircleArc(bv.Circle, a)
        case geometry.ArcEntity:
                return geometry.IntersectArcs(a, bv.Arc)
        }
        return nil
}

func intersectPolylineWith(p geometry.Polyline, b geometry.Entity) []geometry.Point {
        switch bv := b.(type) {
        case geometry.SegmentEntity:
                return geometry.IntersectSegmentPolyline(bv.Segment, p)
        case geometry.PolylineEntity:
                return geometry.IntersectPolylines(p, bv.Polyline)
        }
        return nil
}
