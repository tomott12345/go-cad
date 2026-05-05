package document

import (
	"github.com/tomott12345/go-cad/internal/geometry"
)

// EntityBoundingBox returns the axis-aligned bounding box of entity id,
// computed by the geometry engine. Returns an empty BBox if not found.
func (d *Document) EntityBoundingBox(id int) geometry.BBox {
	for _, e := range d.entities {
		if e.ID == id {
			return e.BoundingBox()
		}
	}
	return geometry.EmptyBBox()
}

// SnapToEntity returns the nearest point on entity id to query point (x, y).
// Used by drawing tools for object-snap operations.
// Returns (x, y) unchanged if the entity is not found.
func (d *Document) SnapToEntity(id int, x, y float64) (float64, float64) {
	for _, e := range d.entities {
		if e.ID == id {
			p := e.ClosestPoint(geometry.Point{X: x, Y: y})
			return p.X, p.Y
		}
	}
	return x, y
}

// NearestEntity returns the ID of the entity whose boundary is closest to
// query point (x, y), or 0 if the document is empty. snapRadius sets the
// maximum search distance (0 = unlimited).
func (d *Document) NearestEntity(x, y float64, snapRadius float64) int {
	q := geometry.Point{X: x, Y: y}
	bestID := 0
	bestDist := snapRadius
	unlimited := snapRadius <= 0

	for _, e := range d.entities {
		ge := e.ToGeometryEntity()
		if ge == nil {
			continue
		}
		cp := ge.ClosestPoint(q)
		dist := q.Dist(cp)
		if unlimited || dist <= bestDist {
			if bestID == 0 || dist < bestDist {
				bestDist = dist
				bestID = e.ID
			}
		}
	}
	return bestID
}

// IntersectEntities returns the intersection points between two entities by ID.
// Returns nil if either ID is not found or the entities do not intersect.
func (d *Document) IntersectEntities(idA, idB int) [][2]float64 {
	var ea, eb *Entity
	for i := range d.entities {
		switch d.entities[i].ID {
		case idA:
			ea = &d.entities[i]
		case idB:
			eb = &d.entities[i]
		}
	}
	if ea == nil || eb == nil {
		return nil
	}
	pts := ea.IntersectWith(*eb)
	if len(pts) == 0 {
		return nil
	}
	out := make([][2]float64, len(pts))
	for i, p := range pts {
		out[i] = [2]float64{p.X, p.Y}
	}
	return out
}

// OffsetEntity adds a new entity that is a geometric offset of entity id by
// dist (positive = left/outward). Returns the new entity's ID, or -1 if the
// source entity is not found or the offset cannot be computed.
func (d *Document) OffsetEntity(id int, dist float64) int {
	for _, e := range d.entities {
		if e.ID == id {
			off := e.Offset(dist)
			if off == nil {
				return -1
			}
			return d.add(*off)
		}
	}
	return -1
}

// TrimEntity splits entity id at parametric position t ∈ [0,1].
// The original entity is replaced by the two resulting sub-entities as a
// single atomic undo operation.
// Returns the IDs of the two new entities, or (-1, -1) on failure.
func (d *Document) TrimEntity(id int, t float64) (int, int) {
	for i, e := range d.entities {
		if e.ID != id {
			continue
		}
		ge := e.ToGeometryEntity()
		if ge == nil {
			return -1, -1
		}
		left, right := ge.TrimAt(t)
		leftDoc := GeometryEntityToDocument(left, e.Layer, e.Color)
		rightDoc := GeometryEntityToDocument(right, e.Layer, e.Color)
		if leftDoc == nil || rightDoc == nil {
			return -1, -1
		}
		// Take a single snapshot before any mutation so undo restores the
		// pre-trim state atomically (one undo step, not three).
		d.pushUndo()
		// Remove original entity directly (without a redundant undo push).
		d.entities = append(d.entities[:i], d.entities[i+1:]...)
		// Assign IDs and append without triggering pushUndo again.
		leftDoc.ID = d.nextID
		d.nextID++
		if leftDoc.Color == "" {
			leftDoc.Color = "#ffffff"
		}
		rightDoc.ID = d.nextID
		d.nextID++
		if rightDoc.Color == "" {
			rightDoc.Color = "#ffffff"
		}
		d.entities = append(d.entities, *leftDoc, *rightDoc)
		return leftDoc.ID, rightDoc.ID
	}
	return -1, -1
}

// EntityLength returns the arc length of entity id via the geometry engine.
func (d *Document) EntityLength(id int) float64 {
	for _, e := range d.entities {
		if e.ID == id {
			ge := e.ToGeometryEntity()
			if ge != nil {
				return ge.Length()
			}
			return e.Length()
		}
	}
	return 0
}
