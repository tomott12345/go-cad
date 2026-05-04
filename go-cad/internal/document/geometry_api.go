// geometry_api.go adds geometry-engine-powered operations to the Document.
// These functions use the internal/geometry package for precision operations
// (snapping, intersection queries, bounding boxes) and are accessible from
// both the WASM and Fyne targets via the shared document API.
package document

import (
	"go-cad/internal/geometry"
)

// ─── Bounding box ─────────────────────────────────────────────────────────────

// EntityBoundingBox returns the axis-aligned bounding box of entity id,
// computed by the geometry engine. Returns the zero BBox if the id is not
// found or the type is unrecognised.
func (d *Document) EntityBoundingBox(id int) geometry.BBox {
	for _, e := range d.entities {
		if e.ID == id {
			return e.BoundingBox()
		}
	}
	return geometry.EmptyBBox()
}

// ─── Nearest point (snap) ─────────────────────────────────────────────────────

// SnapToEntity returns the nearest point on entity id to the query point (x, y).
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

// ─── Intersection ─────────────────────────────────────────────────────────────

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

// ─── Offset ───────────────────────────────────────────────────────────────────

// OffsetEntity adds a new entity to the document that is a geometric offset
// of entity id by dist (positive = left/outward).
// Returns the new entity's ID, or -1 if the source entity is not found.
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

// ─── Trim (split) ─────────────────────────────────────────────────────────────

// TrimEntity splits entity id at parametric position t ∈ [0,1].
// The original entity is replaced by the two resulting sub-entities.
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
		// Remove original (no undo push — the two adds will each push)
		d.entities = append(d.entities[:i], d.entities[i+1:]...)
		idL := d.add(*leftDoc)
		idR := d.add(*rightDoc)
		return idL, idR
	}
	return -1, -1
}

// ─── Length ───────────────────────────────────────────────────────────────────

// EntityLength returns the arc length of entity id via the geometry engine.
// Falls back to the document entity's own Length() if the type is not
// recognised by the geometry engine.
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
