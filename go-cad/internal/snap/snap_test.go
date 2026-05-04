package snap_test

import (
	"math"
	"testing"

	"go-cad/internal/document"
	"go-cad/internal/snap"
)

const eps = 1e-6

func approxEq(a, b float64) bool { return math.Abs(a-b) < eps }

// ─── Endpoint ─────────────────────────────────────────────────────────────────

func TestFindSnap_Endpoint_Line(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 10, 0, 0, "#fff")
	ents := d.Entities()
	// Snap near the start endpoint (0,0)
	c := snap.FindSnap(0.5, 0.3, ents, 2, snap.SnapEndpoint)
	if c == nil {
		t.Fatal("expected endpoint snap near (0,0), got nil")
	}
	if c.Type != snap.SnapEndpoint {
		t.Errorf("type: want SnapEndpoint, got %d", c.Type)
	}
	if !approxEq(c.X, 0) || !approxEq(c.Y, 0) {
		t.Errorf("point: want (0,0), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Endpoint_LineEnd(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 10, 0, 0, "#fff")
	ents := d.Entities()
	c := snap.FindSnap(9.8, 0.2, ents, 2, snap.SnapEndpoint)
	if c == nil {
		t.Fatal("expected endpoint snap near (10,0)")
	}
	if !approxEq(c.X, 10) || !approxEq(c.Y, 0) {
		t.Errorf("point: want (10,0), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Endpoint_Arc(t *testing.T) {
	d := document.New()
	// Arc: centre (0,0), r=5, 0°→90°
	d.AddArc(0, 0, 5, 0, 90, 0, "#fff")
	ents := d.Entities()
	// Start point is (5,0)
	c := snap.FindSnap(5.3, 0.3, ents, 2, snap.SnapEndpoint)
	if c == nil {
		t.Fatal("expected arc start endpoint near (5,0)")
	}
	if !approxEq(c.X, 5) || !approxEq(c.Y, 0) {
		t.Errorf("arc start: want (5,0), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Endpoint_Rectangle(t *testing.T) {
	d := document.New()
	d.AddRectangle(0, 0, 4, 3, 0, "#fff")
	ents := d.Entities()
	c := snap.FindSnap(4.3, 2.8, ents, 2, snap.SnapEndpoint)
	if c == nil {
		t.Fatal("expected corner snap near (4,3)")
	}
	if !approxEq(c.X, 4) || !approxEq(c.Y, 3) {
		t.Errorf("corner: want (4,3), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Endpoint_Polyline(t *testing.T) {
	d := document.New()
	d.AddPolyline([][]float64{{0, 0}, {5, 0}, {5, 5}}, 0, "#fff")
	ents := d.Entities()
	c := snap.FindSnap(4.8, 4.7, ents, 2, snap.SnapEndpoint)
	if c == nil {
		t.Fatal("expected vertex snap near (5,5)")
	}
	if !approxEq(c.X, 5) || !approxEq(c.Y, 5) {
		t.Errorf("vertex: want (5,5), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Endpoint_NoMatch(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 10, 0, 0, "#fff")
	ents := d.Entities()
	// Far from any endpoint
	c := snap.FindSnap(5, 0, ents, 0.5, snap.SnapEndpoint)
	if c != nil {
		t.Errorf("expected nil, got snap at (%.4f,%.4f)", c.X, c.Y)
	}
}

// ─── Midpoint ─────────────────────────────────────────────────────────────────

func TestFindSnap_Midpoint_Line(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 10, 0, 0, "#fff")
	ents := d.Entities()
	c := snap.FindSnap(5.2, 0.3, ents, 2, snap.SnapMidpoint)
	if c == nil {
		t.Fatal("expected midpoint snap near (5,0)")
	}
	if c.Type != snap.SnapMidpoint {
		t.Errorf("type: want SnapMidpoint, got %d", c.Type)
	}
	if !approxEq(c.X, 5) || !approxEq(c.Y, 0) {
		t.Errorf("midpoint: want (5,0), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Midpoint_Rectangle(t *testing.T) {
	d := document.New()
	d.AddRectangle(0, 0, 4, 2, 0, "#fff")
	ents := d.Entities()
	// Bottom side midpoint (2, 0)
	c := snap.FindSnap(2.1, 0.2, ents, 1, snap.SnapMidpoint)
	if c == nil {
		t.Fatal("expected midpoint of bottom side")
	}
	if !approxEq(c.X, 2) || !approxEq(c.Y, 0) {
		t.Errorf("midpoint: want (2,0), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Midpoint_Arc(t *testing.T) {
	d := document.New()
	// Arc 0°→90°, midpoint at 45°
	d.AddArc(0, 0, 5, 0, 90, 0, "#fff")
	ents := d.Entities()
	// Midpoint at 45°: (5cos45, 5sin45) ≈ (3.536, 3.536)
	ex := 5 * math.Cos(45*math.Pi/180)
	ey := 5 * math.Sin(45*math.Pi/180)
	c := snap.FindSnap(ex+0.2, ey+0.2, ents, 2, snap.SnapMidpoint)
	if c == nil {
		t.Fatal("expected arc midpoint snap")
	}
	if !approxEq(c.X, ex) || !approxEq(c.Y, ey) {
		t.Errorf("arc midpoint: want (%.4f,%.4f), got (%.4f,%.4f)", ex, ey, c.X, c.Y)
	}
}

// ─── Center ───────────────────────────────────────────────────────────────────

func TestFindSnap_Center_Circle(t *testing.T) {
	d := document.New()
	d.AddCircle(3, 4, 5, 0, "#fff")
	ents := d.Entities()
	c := snap.FindSnap(3.3, 3.8, ents, 1, snap.SnapCenter)
	if c == nil {
		t.Fatal("expected center snap near (3,4)")
	}
	if c.Type != snap.SnapCenter {
		t.Errorf("type: want SnapCenter, got %d", c.Type)
	}
	if !approxEq(c.X, 3) || !approxEq(c.Y, 4) {
		t.Errorf("center: want (3,4), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Center_Arc(t *testing.T) {
	d := document.New()
	d.AddArc(7, 2, 3, 0, 180, 0, "#fff")
	ents := d.Entities()
	c := snap.FindSnap(7.1, 2.1, ents, 1, snap.SnapCenter)
	if c == nil {
		t.Fatal("expected arc center snap")
	}
	if !approxEq(c.X, 7) || !approxEq(c.Y, 2) {
		t.Errorf("center: want (7,2), got (%.4f,%.4f)", c.X, c.Y)
	}
}

// ─── Quadrant ─────────────────────────────────────────────────────────────────

func TestFindSnap_Quadrant_Circle(t *testing.T) {
	d := document.New()
	d.AddCircle(0, 0, 5, 0, "#fff")
	ents := d.Entities()
	// Right quadrant (5, 0)
	c := snap.FindSnap(5.3, 0.2, ents, 2, snap.SnapQuadrant)
	if c == nil {
		t.Fatal("expected quadrant snap near (5,0)")
	}
	if c.Type != snap.SnapQuadrant {
		t.Errorf("type: want SnapQuadrant, got %d", c.Type)
	}
	if !approxEq(c.X, 5) || !approxEq(c.Y, 0) {
		t.Errorf("quadrant: want (5,0), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Quadrant_ArcInRange(t *testing.T) {
	d := document.New()
	// Arc from 0° to 180°: 0° and 90° and 180° quadrants are valid
	d.AddArc(0, 0, 5, 0, 180, 0, "#fff")
	ents := d.Entities()
	// 90° quadrant = (0, 5)
	c := snap.FindSnap(0.2, 5.3, ents, 2, snap.SnapQuadrant)
	if c == nil {
		t.Fatal("expected 90° quadrant snap for arc 0→180")
	}
	if !approxEq(c.X, 0) || !approxEq(c.Y, 5) {
		t.Errorf("quadrant: want (0,5), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Quadrant_ArcOutOfRange(t *testing.T) {
	d := document.New()
	// Arc from 0° to 45°: 90° quadrant is NOT within range
	d.AddArc(0, 0, 5, 0, 45, 0, "#fff")
	ents := d.Entities()
	c := snap.FindSnap(0, 5, ents, 2, snap.SnapQuadrant)
	// Should not snap to 90° quadrant (out of arc range)
	if c != nil && approxEq(c.X, 0) && approxEq(c.Y, 5) {
		t.Error("quadrant snap returned 90° point outside arc range")
	}
}

// ─── Intersection ─────────────────────────────────────────────────────────────

func TestFindSnap_Intersection_TwoLines(t *testing.T) {
	d := document.New()
	d.AddLine(-5, 0, 5, 0, 0, "#fff")  // horizontal
	d.AddLine(0, -5, 0, 5, 0, "#fff")  // vertical
	ents := d.Entities()
	c := snap.FindSnap(0.3, 0.2, ents, 2, snap.SnapIntersection)
	if c == nil {
		t.Fatal("expected intersection snap near (0,0)")
	}
	if c.Type != snap.SnapIntersection {
		t.Errorf("type: want SnapIntersection, got %d", c.Type)
	}
	if !approxEq(c.X, 0) || !approxEq(c.Y, 0) {
		t.Errorf("intersection: want (0,0), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Intersection_LineCircle(t *testing.T) {
	d := document.New()
	d.AddCircle(0, 0, 5, 0, "#fff")
	d.AddLine(-10, 0, 10, 0, 0, "#fff") // horizontal through centre → 2 intersections
	ents := d.Entities()
	// Right intersection at (5, 0)
	c := snap.FindSnap(5.3, 0.2, ents, 2, snap.SnapIntersection)
	if c == nil {
		t.Fatal("expected line-circle intersection snap near (5,0)")
	}
	if !approxEq(c.X, 5) || !approxEq(c.Y, 0) {
		t.Errorf("intersection: want (5,0), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Intersection_NoMatch(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 5, 0, 0, "#fff")
	d.AddLine(0, 10, 5, 10, 0, "#fff") // parallel — no intersection
	ents := d.Entities()
	c := snap.FindSnap(2.5, 5, ents, 2, snap.SnapIntersection)
	if c != nil {
		t.Errorf("expected nil for parallel lines, got snap at (%.4f,%.4f)", c.X, c.Y)
	}
}

// ─── Perpendicular ────────────────────────────────────────────────────────────

func TestFindSnap_Perpendicular_Line(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 10, 0, 0, "#fff") // horizontal
	ents := d.Entities()
	// Cursor at (3, 4) — foot on horizontal line is (3, 0)
	c := snap.FindSnap(3, 4, ents, 6, snap.SnapPerpendicular)
	if c == nil {
		t.Fatal("expected perpendicular foot snap")
	}
	if c.Type != snap.SnapPerpendicular {
		t.Errorf("type: want SnapPerpendicular, got %d", c.Type)
	}
	if !approxEq(c.X, 3) || !approxEq(c.Y, 0) {
		t.Errorf("foot: want (3,0), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Perpendicular_OutsideSegment(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 5, 0, 0, "#fff")
	ents := d.Entities()
	// Cursor at (8, 3) — foot at (8,0) is outside segment [0,5]
	c := snap.FindSnap(8, 3, ents, 5, snap.SnapPerpendicular)
	if c != nil {
		t.Errorf("foot outside segment — expected nil, got (%.4f,%.4f)", c.X, c.Y)
	}
}

// ─── Tangent ──────────────────────────────────────────────────────────────────

func TestFindSnap_Tangent_Circle(t *testing.T) {
	d := document.New()
	d.AddCircle(0, 0, 5, 0, "#fff")
	ents := d.Entities()
	// Cursor far outside circle at (10, 0); tangent points exist
	c := snap.FindSnap(10, 0, ents, 20, snap.SnapTangent)
	if c == nil {
		t.Fatal("expected tangent snap for cursor outside circle")
	}
	if c.Type != snap.SnapTangent {
		t.Errorf("type: want SnapTangent, got %d", c.Type)
	}
	// Verify tangent length: |CT|² = r² → |CT|=r
	dist := math.Hypot(c.X, c.Y) // distance from centre (0,0)
	if !approxEq(dist, 5) {
		t.Errorf("tangent point should be on circle (dist=r=5), got dist=%.4f", dist)
	}
}

func TestFindSnap_Tangent_InsideCircle(t *testing.T) {
	d := document.New()
	d.AddCircle(0, 0, 5, 0, "#fff")
	ents := d.Entities()
	// Cursor inside circle — no tangent
	c := snap.FindSnap(1, 1, ents, 20, snap.SnapTangent)
	if c != nil {
		t.Error("cursor inside circle should return nil tangent")
	}
}

// ─── Nearest ──────────────────────────────────────────────────────────────────

func TestFindSnap_Nearest_Line(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 10, 0, 0, "#fff") // horizontal
	ents := d.Entities()
	// Cursor at (3, 4) — nearest point on horizontal line is (3,0)
	c := snap.FindSnap(3, 4, ents, 10, snap.SnapNearest)
	if c == nil {
		t.Fatal("expected nearest snap")
	}
	if c.Type != snap.SnapNearest {
		t.Errorf("type: want SnapNearest, got %d", c.Type)
	}
	if !approxEq(c.X, 3) || !approxEq(c.Y, 0) {
		t.Errorf("nearest: want (3,0), got (%.4f,%.4f)", c.X, c.Y)
	}
}

func TestFindSnap_Nearest_Circle(t *testing.T) {
	d := document.New()
	d.AddCircle(0, 0, 5, 0, "#fff")
	ents := d.Entities()
	// Cursor at (8, 0) — nearest point on circle is (5, 0)
	c := snap.FindSnap(8, 0, ents, 10, snap.SnapNearest)
	if c == nil {
		t.Fatal("expected nearest snap on circle")
	}
	if !approxEq(c.X, 5) || !approxEq(c.Y, 0) {
		t.Errorf("nearest on circle: want (5,0), got (%.4f,%.4f)", c.X, c.Y)
	}
}

// ─── Priority ─────────────────────────────────────────────────────────────────

func TestFindSnap_Priority_EndpointOverMidpoint(t *testing.T) {
	d := document.New()
	// Line from (0,0) to (10,0); midpoint at (5,0); cursor near (0,0)
	d.AddLine(0, 0, 10, 0, 0, "#fff")
	ents := d.Entities()
	// Even though midpoint is further, if cursor is near endpoint, endpoint wins.
	c := snap.FindSnap(0.3, 0, ents, 3, snap.SnapAll)
	if c == nil {
		t.Fatal("expected a snap")
	}
	if c.Type != snap.SnapEndpoint {
		t.Errorf("priority: want SnapEndpoint near (0,0), got type %d at (%.4f,%.4f)", c.Type, c.X, c.Y)
	}
}

// ─── Mask / disabled snaps ────────────────────────────────────────────────────

func TestFindSnap_MaskDisabled(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 10, 0, 0, "#fff")
	ents := d.Entities()
	// Endpoint disabled; should return nil when cursor is near endpoint.
	c := snap.FindSnap(0.3, 0.2, ents, 2, snap.SnapMidpoint) // only midpoint enabled
	if c != nil && c.Type == snap.SnapEndpoint {
		t.Error("SnapEndpoint should be disabled")
	}
}

func TestFindSnap_EmptyEntities(t *testing.T) {
	c := snap.FindSnap(0, 0, nil, 10, snap.SnapAll)
	if c != nil {
		t.Error("expected nil for empty entity list")
	}
}

// ─── Entity ID propagation ────────────────────────────────────────────────────

func TestFindSnap_EntityID(t *testing.T) {
	d := document.New()
	id := d.AddLine(0, 0, 10, 0, 0, "#fff")
	ents := d.Entities()
	c := snap.FindSnap(0.2, 0.1, ents, 2, snap.SnapEndpoint)
	if c == nil {
		t.Fatal("expected snap")
	}
	if c.EntityID != id {
		t.Errorf("entityID: want %d, got %d", id, c.EntityID)
	}
}
