package document_test

import (
	"math"
	"strings"
	"testing"

	"go-cad/internal/document"
	"go-cad/internal/geometry"
)

// ── Spline ───────────────────────────────────────────────────────────────────

func TestAddSpline_Basic(t *testing.T) {
	d := document.New()
	pts := [][]float64{{0, 0}, {10, 20}, {20, -10}, {30, 0}}
	id := d.AddSpline(pts, 0, "#ff0000")
	if id <= 0 {
		t.Fatalf("AddSpline returned invalid id %d", id)
	}
	es := d.Entities()
	if len(es) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(es))
	}
	e := es[0]
	if e.Type != document.TypeSpline {
		t.Errorf("type: got %q, want %q", e.Type, document.TypeSpline)
	}
	if len(e.Points) != 4 {
		t.Errorf("expected 4 control points, got %d", len(e.Points))
	}
}

func TestSpline_Length(t *testing.T) {
	d := document.New()
	pts := [][]float64{{0, 0}, {10, 20}, {20, -10}, {30, 0}}
	id := d.AddSpline(pts, 0, "")
	l := d.EntityLength(id)
	if l <= 0 {
		t.Errorf("spline length should be positive, got %f", l)
	}
}

func TestSpline_ToGeometryEntity(t *testing.T) {
	pts := [][]float64{{0, 0}, {10, 20}, {20, -10}, {30, 0}}
	e := document.Entity{Type: document.TypeSpline, Points: pts}
	ge := e.ToGeometryEntity()
	if ge == nil {
		t.Fatal("ToGeometryEntity returned nil for spline")
	}
	_, ok := ge.(geometry.BezierEntity)
	if !ok {
		t.Errorf("expected BezierEntity, got %T", ge)
	}
}

func TestSpline_ToGeometryEntity_TooFewPoints(t *testing.T) {
	// Fewer than 4 control points should return nil (no valid cubic segment).
	pts := [][]float64{{0, 0}, {10, 10}, {20, 0}}
	e := document.Entity{Type: document.TypeSpline, Points: pts}
	if ge := e.ToGeometryEntity(); ge != nil {
		t.Errorf("expected nil for 3-point spline, got %T", ge)
	}
}

func TestSpline_BoundingBox(t *testing.T) {
	pts := [][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 0}}
	e := document.Entity{Type: document.TypeSpline, Points: pts}
	bb := e.BoundingBox()
	if bb.IsEmpty() {
		t.Error("spline bounding box should not be empty")
	}
	if bb.Min.X > 0 || bb.Max.X < 15 {
		t.Errorf("spline bbox X range unexpected: [%v,%v]", bb.Min.X, bb.Max.X)
	}
}

func TestSpline_DXFExport(t *testing.T) {
	d := document.New()
	d.AddSpline([][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 0}}, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "LWPOLYLINE") {
		t.Error("DXF for spline should contain LWPOLYLINE")
	}
}

func TestSpline_JSONRoundtrip(t *testing.T) {
	d := document.New()
	pts := [][]float64{{0, 0}, {1, 2}, {3, 4}, {5, 0}}
	d.AddSpline(pts, 1, "#123456")
	json := d.ToJSON()
	if !strings.Contains(json, "spline") {
		t.Errorf("JSON should contain type 'spline', got: %s", json)
	}
}

// ── Ellipse ──────────────────────────────────────────────────────────────────

func TestAddEllipse_Basic(t *testing.T) {
	d := document.New()
	id := d.AddEllipse(10, 20, 15, 8, 45, 0, "#00ffff")
	if id <= 0 {
		t.Fatalf("AddEllipse returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if e.Type != document.TypeEllipse {
		t.Errorf("type: got %q, want %q", e.Type, document.TypeEllipse)
	}
	if e.CX != 10 || e.CY != 20 {
		t.Errorf("centre: got (%v,%v), want (10,20)", e.CX, e.CY)
	}
	if e.R != 15 || e.R2 != 8 {
		t.Errorf("axes: got R=%v R2=%v, want 15,8", e.R, e.R2)
	}
	if e.RotDeg != 45 {
		t.Errorf("rotation: got %v, want 45", e.RotDeg)
	}
}

func TestEllipse_ToGeometryEntity(t *testing.T) {
	e := document.Entity{Type: document.TypeEllipse, CX: 5, CY: 5, R: 10, R2: 6, RotDeg: 30}
	ge := e.ToGeometryEntity()
	if ge == nil {
		t.Fatal("ToGeometryEntity returned nil for ellipse")
	}
	ee, ok := ge.(geometry.EllipseEntity)
	if !ok {
		t.Fatalf("expected EllipseEntity, got %T", ge)
	}
	if ee.Center.X != 5 || ee.Center.Y != 5 {
		t.Errorf("ellipse centre wrong: %v", ee.Center)
	}
	if ee.A != 10 || ee.B != 6 {
		t.Errorf("ellipse axes wrong: A=%v B=%v", ee.A, ee.B)
	}
}

func TestEllipse_Length(t *testing.T) {
	d := document.New()
	// For a=10, b=10 it should approximate 2π×10 ≈ 62.83.
	id := d.AddEllipse(0, 0, 10, 10, 0, 0, "")
	l := d.EntityLength(id)
	if math.Abs(l-2*math.Pi*10) > 2 {
		t.Errorf("circle-like ellipse length: got %v, want ~%v", l, 2*math.Pi*10)
	}
}

func TestEllipse_BoundingBox(t *testing.T) {
	e := document.Entity{Type: document.TypeEllipse, CX: 0, CY: 0, R: 10, R2: 5, RotDeg: 0}
	bb := e.BoundingBox()
	if bb.IsEmpty() {
		t.Error("ellipse bounding box should not be empty")
	}
	// For un-rotated ellipse at origin with a=10: X range should be ≈ [-10,10].
	if bb.Max.X < 9.9 || bb.Min.X > -9.9 {
		t.Errorf("ellipse bbox X: [%v,%v]", bb.Min.X, bb.Max.X)
	}
}

func TestEllipse_DXFExport(t *testing.T) {
	d := document.New()
	d.AddEllipse(0, 0, 10, 5, 0, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "ELLIPSE") {
		t.Error("DXF should contain ELLIPSE entity")
	}
}

func TestEllipse_GeometryRoundtrip(t *testing.T) {
	orig := document.Entity{Type: document.TypeEllipse, CX: 3, CY: 7, R: 12, R2: 5, RotDeg: 0}
	ge := orig.ToGeometryEntity()
	if ge == nil {
		t.Fatal("ToGeometryEntity nil for ellipse")
	}
	back := document.GeometryEntityToDocument(ge, 0, "")
	if back == nil {
		t.Fatal("GeometryEntityToDocument returned nil")
	}
	if back.Type != document.TypeEllipse {
		t.Errorf("roundtrip type: got %q, want ellipse", back.Type)
	}
	if back.R != 12 || back.R2 != 5 {
		t.Errorf("roundtrip axes: R=%v R2=%v", back.R, back.R2)
	}
}

// ── Text ─────────────────────────────────────────────────────────────────────

func TestAddText_Basic(t *testing.T) {
	d := document.New()
	id := d.AddText(5, 10, "hello", 2.5, 0, 0, "#ffffff")
	if id <= 0 {
		t.Fatalf("AddText returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if e.Type != document.TypeText {
		t.Errorf("type: got %q, want %q", e.Type, document.TypeText)
	}
	if e.Text != "hello" {
		t.Errorf("text content: got %q, want %q", e.Text, "hello")
	}
	if e.TextHeight != 2.5 {
		t.Errorf("text height: got %v, want 2.5", e.TextHeight)
	}
	if e.X1 != 5 || e.Y1 != 10 {
		t.Errorf("text position: got (%v,%v), want (5,10)", e.X1, e.Y1)
	}
}

func TestText_Length(t *testing.T) {
	e := document.Entity{Type: document.TypeText, Text: "hello"}
	if l := e.Length(); l != 0 {
		t.Errorf("text length should be 0, got %v", l)
	}
}

func TestText_ToGeometryEntity(t *testing.T) {
	// Text entities have no geometric representation.
	e := document.Entity{Type: document.TypeText, X1: 0, Y1: 0, Text: "A"}
	if ge := e.ToGeometryEntity(); ge != nil {
		t.Errorf("text ToGeometryEntity should return nil, got %T", ge)
	}
}

func TestText_DXFExport(t *testing.T) {
	d := document.New()
	d.AddText(0, 0, "go-cad", 3.0, 0, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "TEXT") {
		t.Error("DXF should contain TEXT entity")
	}
	if !strings.Contains(dxf, "go-cad") {
		t.Error("DXF TEXT entity should contain the text string")
	}
}

func TestText_JSONRoundtrip(t *testing.T) {
	d := document.New()
	d.AddText(1, 2, "CAD", 4.0, 15, 0, "")
	json := d.ToJSON()
	if !strings.Contains(json, `"text"`) {
		t.Errorf("JSON should contain text field: %s", json)
	}
	if !strings.Contains(json, "CAD") {
		t.Errorf("JSON should contain text content 'CAD': %s", json)
	}
}

// ── Linear dimension ──────────────────────────────────────────────────────────

func TestAddLinearDim(t *testing.T) {
	d := document.New()
	id := d.AddLinearDim(0, 0, 100, 0, 10, 0, "#ffffff")
	if id <= 0 {
		t.Fatalf("AddLinearDim returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if e.Type != document.TypeDimLinear {
		t.Errorf("type: got %q, want %q", e.Type, document.TypeDimLinear)
	}
}

func TestLinearDim_Length(t *testing.T) {
	e := document.Entity{Type: document.TypeDimLinear, X1: 0, Y1: 0, X2: 30, Y2: 40}
	// Length should be the straight-line distance between the two definition points.
	if l := e.Length(); math.Abs(l-50) > 1e-9 {
		t.Errorf("dimlin length: got %v, want 50", l)
	}
}

func TestLinearDim_DXFExport(t *testing.T) {
	d := document.New()
	d.AddLinearDim(0, 0, 50, 0, 20, 0, "#fff")
	dxf := d.ExportDXF()
	// The dim is approximated as LINE + TEXT entities.
	if !strings.Contains(dxf, "LINE") {
		t.Error("linear dim DXF should contain LINE entities")
	}
}

// ── Aligned dimension ─────────────────────────────────────────────────────────

func TestAddAlignedDim(t *testing.T) {
	d := document.New()
	id := d.AddAlignedDim(0, 0, 30, 40, 8, 0, "#ffffff")
	if id <= 0 {
		t.Fatalf("AddAlignedDim returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if e.Type != document.TypeDimAligned {
		t.Errorf("type: got %q, want %q", e.Type, document.TypeDimAligned)
	}
}

func TestAlignedDim_DXFExport(t *testing.T) {
	d := document.New()
	d.AddAlignedDim(0, 0, 30, 40, 5, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "LINE") {
		t.Error("aligned dim DXF should contain LINE entities")
	}
}

// ── Angular dimension ─────────────────────────────────────────────────────────

func TestAddAngularDim(t *testing.T) {
	d := document.New()
	id := d.AddAngularDim(0, 0, 10, 0, 0, 10, 5, 0, "#ffffff")
	if id <= 0 {
		t.Fatalf("AddAngularDim returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if e.Type != document.TypeDimAngular {
		t.Errorf("type: got %q, want %q", e.Type, document.TypeDimAngular)
	}
}

func TestAngularDim_Length(t *testing.T) {
	// 90° arc at radius 10 → length = π/2 × 10 ≈ 15.71.
	e := document.Entity{
		Type: document.TypeDimAngular,
		CX: 0, CY: 0,
		X1: 10, Y1: 0,    // 0° ray
		X2: 0, Y2: 10,    // 90° ray
		R: 10,
	}
	l := e.Length()
	if math.Abs(l-math.Pi/2*10) > 0.5 {
		t.Errorf("angular dim arc length: got %v, want ~%v", l, math.Pi/2*10)
	}
}

func TestAngularDim_DXFExport(t *testing.T) {
	d := document.New()
	d.AddAngularDim(0, 0, 10, 0, 0, 10, 5, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "LINE") {
		t.Error("angular dim DXF should contain LINE entities for arc approx")
	}
}

// ── Radial dimension ──────────────────────────────────────────────────────────

func TestAddRadialDim(t *testing.T) {
	d := document.New()
	id := d.AddRadialDim(0, 0, 15, 45, 0, "#ffffff")
	if id <= 0 {
		t.Fatalf("AddRadialDim returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if e.Type != document.TypeDimRadial {
		t.Errorf("type: got %q, want %q", e.Type, document.TypeDimRadial)
	}
	if e.R != 15 {
		t.Errorf("radius: got %v, want 15", e.R)
	}
}

func TestRadialDim_Length(t *testing.T) {
	e := document.Entity{Type: document.TypeDimRadial, R: 7.5}
	if l := e.Length(); l != 7.5 {
		t.Errorf("radial dim length: got %v, want 7.5", l)
	}
}

func TestRadialDim_DXFExport(t *testing.T) {
	d := document.New()
	d.AddRadialDim(0, 0, 10, 0, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "LINE") {
		t.Error("radial dim DXF should contain LINE entity for leader")
	}
	if !strings.Contains(dxf, "R10.000") {
		t.Errorf("radial dim DXF should contain radius label, got: %s", dxf)
	}
}

// ── Diameter dimension ────────────────────────────────────────────────────────

func TestAddDiameterDim(t *testing.T) {
	d := document.New()
	id := d.AddDiameterDim(5, 5, 12, 0, 0, "#00ff00")
	if id <= 0 {
		t.Fatalf("AddDiameterDim returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if e.Type != document.TypeDimDiameter {
		t.Errorf("type: got %q, want %q", e.Type, document.TypeDimDiameter)
	}
}

func TestDiameterDim_Length(t *testing.T) {
	e := document.Entity{Type: document.TypeDimDiameter, R: 5}
	if l := e.Length(); l != 10 {
		t.Errorf("diameter dim length: got %v, want 10", l)
	}
}

// ── AddEntity dispatch ────────────────────────────────────────────────────────

func TestAddEntity_AllNewTypes(t *testing.T) {
	cases := []document.Entity{
		{Type: document.TypeSpline, Points: [][]float64{{0,0},{1,2},{3,4},{5,0}}},
		{Type: document.TypeEllipse, CX: 0, CY: 0, R: 5, R2: 3},
		{Type: document.TypeText, X1: 0, Y1: 0, Text: "hi", TextHeight: 2},
		{Type: document.TypeDimLinear, X1: 0, Y1: 0, X2: 10, Y2: 0},
		{Type: document.TypeDimAligned, X1: 0, Y1: 0, X2: 10, Y2: 5},
		{Type: document.TypeDimAngular, CX: 0, CY: 0, X1: 10, Y1: 0, X2: 0, Y2: 10, R: 5},
		{Type: document.TypeDimRadial, CX: 0, CY: 0, R: 8},
		{Type: document.TypeDimDiameter, CX: 0, CY: 0, R: 6},
	}
	for _, tc := range cases {
		d := document.New()
		id := d.AddEntity(tc)
		if id <= 0 {
			t.Errorf("AddEntity(%q) returned %d, want >0", tc.Type, id)
		}
		if d.EntityCount() != 1 {
			t.Errorf("AddEntity(%q): expected 1 entity in doc, got %d", tc.Type, d.EntityCount())
		}
	}
}

func TestAddEntity_Unknown(t *testing.T) {
	d := document.New()
	id := d.AddEntity(document.Entity{Type: "xyzzy"})
	if id != -1 {
		t.Errorf("AddEntity(unknown) should return -1, got %d", id)
	}
}

// ── Undo/redo with new types ──────────────────────────────────────────────────

func TestUndoRedo_Spline(t *testing.T) {
	d := document.New()
	d.AddSpline([][]float64{{0,0},{1,2},{3,4},{5,0}}, 0, "")
	if d.EntityCount() != 1 {
		t.Fatalf("expected 1 entity before undo")
	}
	d.Undo()
	if d.EntityCount() != 0 {
		t.Errorf("expected 0 entities after undo, got %d", d.EntityCount())
	}
	d.Redo()
	if d.EntityCount() != 1 {
		t.Errorf("expected 1 entity after redo, got %d", d.EntityCount())
	}
}

// ── DXF completeness ─────────────────────────────────────────────────────────

func TestExportDXF_AllNewTypes(t *testing.T) {
	d := document.New()
	d.AddSpline([][]float64{{0,0},{5,10},{10,10},{15,0}}, 0, "#fff")
	d.AddEllipse(0, 0, 10, 5, 0, 0, "#fff")
	d.AddText(0, 0, "label", 3, 0, 0, "#fff")
	d.AddLinearDim(0, 0, 50, 0, 10, 0, "#fff")
	d.AddAlignedDim(0, 0, 30, 40, 5, 0, "#fff")
	d.AddAngularDim(0, 0, 10, 0, 0, 10, 5, 0, "#fff")
	d.AddRadialDim(0, 0, 8, 0, 0, "#fff")
	d.AddDiameterDim(0, 0, 6, 0, 0, "#fff")

	dxf := d.ExportDXF()
	for _, want := range []string{"LWPOLYLINE", "ELLIPSE", "TEXT", "LINE", "SECTION", "ENTITIES", "EOF"} {
		if !strings.Contains(dxf, want) {
			t.Errorf("DXF output missing %q", want)
		}
	}
}

// ── Spline geometry engine integration ───────────────────────────────────────

func TestSpline_GeometryBoundingBox_ViaEngine(t *testing.T) {
	d := document.New()
	// Control points that form a curve over [0..30] in X.
	id := d.AddSpline([][]float64{{0,0},{0,20},{30,20},{30,0}}, 0, "")
	bb := d.EntityBoundingBox(id)
	if bb.IsEmpty() {
		t.Error("EntityBoundingBox empty for spline")
	}
	if bb.Max.X < 25 {
		t.Errorf("spline bbox MaxX: got %v, want ≥25", bb.Max.X)
	}
}

func TestEllipse_GeometryBoundingBox_ViaEngine(t *testing.T) {
	d := document.New()
	id := d.AddEllipse(0, 0, 10, 5, 0, 0, "")
	bb := d.EntityBoundingBox(id)
	if bb.IsEmpty() {
		t.Error("EntityBoundingBox empty for ellipse")
	}
	// Un-rotated ellipse: X in [-10,10], Y in [-5,5].
	if bb.Max.X < 9.9 {
		t.Errorf("ellipse bbox MaxX: got %v, want ≥9.9", bb.Max.X)
	}
}
