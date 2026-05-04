package document

import (
	"encoding/json"
	"fmt"

	"go-cad/internal/geometry"
)

// Kind returns the geometry-layer kind discriminator for this entity,
// mapping legacy Type constants to the geometry engine's Kind values.
// For types with no geometry representation the legacy Type string is returned.
func (e Entity) Kind() string {
	ge := e.ToGeometryEntity()
	if ge == nil {
		return e.Type
	}
	return string(ge.Kind())
}

// MarshalGeometryJSON serialises this entity using the geometry-layer
// kind-discriminated format: {"kind":"...", "data":{...}}.
// This is the canonical wire format for exchanging entities with the geometry engine.
func (e Entity) MarshalGeometryJSON() ([]byte, error) {
	ge := e.ToGeometryEntity()
	if ge == nil {
		return nil, fmt.Errorf("document: no geometry representation for type %q", e.Type)
	}
	return geometry.MarshalEntity(ge)
}

// UnmarshalGeometryJSON parses a geometry-layer kind-discriminated JSON payload
// and returns the equivalent document Entity (ID=0; caller must assign).
// It accepts the {"kind":"...", "data":{...}} format emitted by MarshalGeometryJSON.
func UnmarshalGeometryJSON(data []byte) (*Entity, error) {
	ge, err := geometry.UnmarshalEntity(data)
	if err != nil {
		return nil, fmt.Errorf("document: %w", err)
	}
	e := GeometryEntityToDocument(ge, 0, "")
	if e == nil {
		return nil, fmt.Errorf("document: unsupported geometry kind %T", ge)
	}
	return e, nil
}

// ToGeometryJSONArray returns all entities serialised as a JSON array using
// the geometry-layer kind-discriminated format.  Entities with no geometry
// representation (e.g. unknown types) are silently skipped.
func (d *Document) ToGeometryJSONArray() ([]byte, error) {
	entities := d.Entities()
	arr := make([]json.RawMessage, 0, len(entities))
	for _, e := range entities {
		b, err := e.MarshalGeometryJSON()
		if err != nil {
			continue
		}
		arr = append(arr, b)
	}
	return json.Marshal(arr)
}
