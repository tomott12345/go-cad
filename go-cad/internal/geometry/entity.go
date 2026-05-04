package geometry

import (
	"encoding/json"
	"fmt"
)

// Kind is the discriminator for entity types.
type Kind string

const (
	KindSegment      Kind = "segment"
	KindLine         Kind = "line"
	KindCircle       Kind = "circle"
	KindArc          Kind = "arc"
	KindEllipse      Kind = "ellipse"
	KindPolyline     Kind = "polyline"
	KindBezierSpline Kind = "bezier"
	KindNURBSSpline  Kind = "nurbs"
)

// Entity is the common interface for all 2D geometric primitives.
type Entity interface {
	// Kind returns the entity's type discriminator.
	Kind() Kind
	// BoundingBox returns the axis-aligned bounding box.
	BoundingBox() BBox
	// ClosestPoint returns the nearest point on this entity to p.
	ClosestPoint(p Point) Point
	// Length returns the arc length of the entity.
	Length() float64
	// Offset returns a new entity offset by dist (positive = left/outward).
	// The returned type may differ (e.g. Spline offset returns Polyline).
	Offset(dist float64) Entity
}

// ─── Entity wrappers ─────────────────────────────────────────────────────────

// SegmentEntity wraps Segment to implement Entity.
type SegmentEntity struct{ Segment }

func (e SegmentEntity) Kind() Kind           { return KindSegment }
func (e SegmentEntity) BoundingBox() BBox    { return e.Segment.BoundingBox() }
func (e SegmentEntity) ClosestPoint(p Point) Point { cp, _ := e.Segment.ClosestPoint(p); return cp }
func (e SegmentEntity) Length() float64      { return e.Segment.Length() }
func (e SegmentEntity) Offset(dist float64) Entity {
	return SegmentEntity{e.Segment.Offset(dist)}
}

// CircleEntity wraps Circle to implement Entity.
type CircleEntity struct{ Circle }

func (e CircleEntity) Kind() Kind           { return KindCircle }
func (e CircleEntity) BoundingBox() BBox    { return e.Circle.BoundingBox() }
func (e CircleEntity) ClosestPoint(p Point) Point { return e.Circle.ClosestPoint(p) }
func (e CircleEntity) Length() float64      { return e.Circle.Circumference() }
func (e CircleEntity) Offset(dist float64) Entity {
	return CircleEntity{e.Circle.Offset(dist)}
}

// ArcEntity wraps Arc to implement Entity.
type ArcEntity struct{ Arc }

func (e ArcEntity) Kind() Kind           { return KindArc }
func (e ArcEntity) BoundingBox() BBox    { return e.Arc.BoundingBox() }
func (e ArcEntity) ClosestPoint(p Point) Point { return e.Arc.ClosestPoint(p) }
func (e ArcEntity) Length() float64      { return e.Arc.Length() }
func (e ArcEntity) Offset(dist float64) Entity {
	return ArcEntity{e.Arc.Offset(dist)}
}

// EllipseEntity wraps Ellipse to implement Entity.
type EllipseEntity struct{ Ellipse }

func (e EllipseEntity) Kind() Kind           { return KindEllipse }
func (e EllipseEntity) BoundingBox() BBox    { return e.Ellipse.BoundingBox() }
func (e EllipseEntity) ClosestPoint(p Point) Point { return e.Ellipse.ClosestPoint(p) }
func (e EllipseEntity) Length() float64      { return e.Ellipse.Circumference() }
func (e EllipseEntity) Offset(dist float64) Entity {
	return EllipseEntity{e.Ellipse.Offset(dist)}
}

// PolylineEntity wraps Polyline to implement Entity.
type PolylineEntity struct{ Polyline }

func (e PolylineEntity) Kind() Kind           { return KindPolyline }
func (e PolylineEntity) BoundingBox() BBox    { return e.Polyline.BoundingBox() }
func (e PolylineEntity) ClosestPoint(p Point) Point { return e.Polyline.ClosestPoint(p) }
func (e PolylineEntity) Length() float64      { return e.Polyline.Length() }
func (e PolylineEntity) Offset(dist float64) Entity {
	return PolylineEntity{e.Polyline.Offset(dist)}
}

// BezierEntity wraps BezierSpline to implement Entity.
type BezierEntity struct{ BezierSpline }

func (e BezierEntity) Kind() Kind           { return KindBezierSpline }
func (e BezierEntity) BoundingBox() BBox    { return e.BezierSpline.BoundingBox() }
func (e BezierEntity) ClosestPoint(p Point) Point { return e.BezierSpline.ClosestPoint(p) }
func (e BezierEntity) Length() float64      { return e.BezierSpline.Length() }
func (e BezierEntity) Offset(dist float64) Entity {
	return PolylineEntity{e.BezierSpline.Offset(dist)}
}

// NURBSEntity wraps NURBSSpline to implement Entity.
type NURBSEntity struct{ NURBSSpline }

func (e NURBSEntity) Kind() Kind           { return KindNURBSSpline }
func (e NURBSEntity) BoundingBox() BBox    { return e.NURBSSpline.BoundingBox() }
func (e NURBSEntity) ClosestPoint(p Point) Point { return e.NURBSSpline.ClosestPoint(p) }
func (e NURBSEntity) Length() float64 {
	pts := e.NURBSSpline.ApproxPolyline(100)
	total := 0.0
	for i := 1; i < len(pts); i++ {
		total += pts[i].Dist(pts[i-1])
	}
	return total
}
func (e NURBSEntity) Offset(dist float64) Entity {
	pts := e.NURBSSpline.ApproxPolyline(100)
	return PolylineEntity{Polyline{Points: pts}.Offset(dist)}
}

// ─── JSON envelope ───────────────────────────────────────────────────────────

// RawEntity is a JSON-serialisable envelope for any Entity type.
type RawEntity struct {
	EntityKind Kind            `json:"kind"`
	Data       json.RawMessage `json:"data"`
}

// MarshalEntity serialises any Entity to a RawEntity JSON envelope.
func MarshalEntity(e Entity) ([]byte, error) {
	var data interface{}
	switch v := e.(type) {
	case SegmentEntity:
		data = v.Segment
	case CircleEntity:
		data = v.Circle
	case ArcEntity:
		data = v.Arc
	case EllipseEntity:
		data = v.Ellipse
	case PolylineEntity:
		data = v.Polyline
	case BezierEntity:
		data = v.BezierSpline
	case NURBSEntity:
		data = v.NURBSSpline
	default:
		return nil, fmt.Errorf("unknown entity type: %T", e)
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(RawEntity{EntityKind: e.Kind(), Data: raw})
}

// UnmarshalEntity deserialises a RawEntity JSON envelope to an Entity.
func UnmarshalEntity(b []byte) (Entity, error) {
	var re RawEntity
	if err := json.Unmarshal(b, &re); err != nil {
		return nil, err
	}
	switch re.EntityKind {
	case KindSegment:
		var s Segment
		if err := json.Unmarshal(re.Data, &s); err != nil {
			return nil, err
		}
		return SegmentEntity{s}, nil
	case KindCircle:
		var c Circle
		if err := json.Unmarshal(re.Data, &c); err != nil {
			return nil, err
		}
		return CircleEntity{c}, nil
	case KindArc:
		var a Arc
		if err := json.Unmarshal(re.Data, &a); err != nil {
			return nil, err
		}
		return ArcEntity{a}, nil
	case KindEllipse:
		var e Ellipse
		if err := json.Unmarshal(re.Data, &e); err != nil {
			return nil, err
		}
		return EllipseEntity{e}, nil
	case KindPolyline:
		var p Polyline
		if err := json.Unmarshal(re.Data, &p); err != nil {
			return nil, err
		}
		return PolylineEntity{p}, nil
	case KindBezierSpline:
		var sp BezierSpline
		if err := json.Unmarshal(re.Data, &sp); err != nil {
			return nil, err
		}
		return BezierEntity{sp}, nil
	case KindNURBSSpline:
		var sp NURBSSpline
		if err := json.Unmarshal(re.Data, &sp); err != nil {
			return nil, err
		}
		return NURBSEntity{sp}, nil
	default:
		return nil, fmt.Errorf("unknown kind: %s", re.EntityKind)
	}
}
