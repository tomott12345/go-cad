package document_test

import (
	"math"
	"strings"
	"testing"

	"github.com/tomott12345/go-cad/internal/document"
	"github.com/tomott12345/go-cad/internal/geometry"
)

// ── Spline (cubic Bezier) ─────────────────────────────────────────────────────

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
	id := d.AddSpline([][]float64{{0, 0}, {10, 20}, {20, -10}, {30, 0}}, 0, "")
	if l := d.EntityLength(id); l <= 0 {
		t.Errorf("spline length should be positive, got %f", l)
	}
}

func TestSpline_ToGeometryEntity(t *testing.T) {
	e := document.Entity{Type: document.TypeSpline,
		Points: [][]float64{{0, 0}, {10, 20}, {20, -10}, {30, 0}}}
	ge := e.ToGeometryEntity()
	if ge == nil {
		t.Fatal("ToGeometryEntity returned nil for spline")
	}
	if _, ok := ge.(geometry.BezierEntity); !ok {
		t.Errorf("expected BezierEntity, got %T", ge)
	}
}

func TestSpline_ToGeometryEntity_TooFewPoints(t *testing.T) {
	e := document.Entity{Type: document.TypeSpline,
		Points: [][]float64{{0, 0}, {10, 10}, {20, 0}}}
	if ge := e.ToGeometryEntity(); ge != nil {
		t.Errorf("expected nil for 3-point spline, got %T", ge)
	}
}

func TestSpline_BoundingBox(t *testing.T) {
	e := document.Entity{Type: document.TypeSpline,
		Points: [][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 0}}}
	bb := e.BoundingBox()
	if bb.IsEmpty() {
		t.Error("spline bounding box should not be empty")
	}
	if bb.Min.X > 0 || bb.Max.X < 15 {
		t.Errorf("spline bbox X range: [%v,%v]", bb.Min.X, bb.Max.X)
	}
}

func TestSpline_DXFExport_R2000(t *testing.T) {
	d := document.New()
	d.AddSpline([][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 0}}, 0, "#fff")
	if dxf := d.ExportDXF(); !strings.Contains(dxf, "LWPOLYLINE") {
		t.Error("R2000 DXF for spline should contain LWPOLYLINE")
	}
}

func TestSpline_DXFExport_R12(t *testing.T) {
	d := document.New()
	d.AddSpline([][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 0}}, 0, "#fff")
	dxf := d.ExportDXFR12()
	if !strings.Contains(dxf, "POLYLINE") {
		t.Error("R12 DXF for spline should contain POLYLINE+VERTEX")
	}
	if !strings.Contains(dxf, "VERTEX") {
		t.Error("R12 DXF for spline should contain VERTEX entries")
	}
	if strings.Contains(dxf, "LWPOLYLINE") {
		t.Error("R12 DXF should not contain LWPOLYLINE (R2000 only)")
	}
}

func TestSpline_JSONRoundtrip(t *testing.T) {
	d := document.New()
	d.AddSpline([][]float64{{0, 0}, {1, 2}, {3, 4}, {5, 0}}, 1, "#123456")
	if json := d.ToJSON(); !strings.Contains(json, "spline") {
		t.Errorf("JSON should contain type 'spline', got: %s", json)
	}
}

// ── NURBS ─────────────────────────────────────────────────────────────────────

func TestAddNURBS_Basic(t *testing.T) {
	d := document.New()
	controls := [][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 5}, {20, 0}}
	id := d.AddNURBS(3, controls, nil, nil, 0, "#00ff00")
	if id <= 0 {
		t.Fatalf("AddNURBS returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if e.Type != document.TypeNURBS {
		t.Errorf("type: got %q, want %q", e.Type, document.TypeNURBS)
	}
	if e.NURBSDegree != 3 {
		t.Errorf("degree: got %d, want 3", e.NURBSDegree)
	}
	if len(e.Points) != 5 {
		t.Errorf("control points: got %d, want 5", len(e.Points))
	}
	if len(e.Knots) == 0 {
		t.Error("knots should be auto-generated when nil is passed")
	}
	if len(e.Weights) != len(e.Points) {
		t.Errorf("weights: got %d, want %d", len(e.Weights), len(e.Points))
	}
}

func TestNURBS_AutoKnots(t *testing.T) {
	d := document.New()
	controls := [][]float64{{0, 0}, {5, 10}, {10, 0}, {15, 10}, {20, 0}}
	id := d.AddNURBS(3, controls, nil, nil, 0, "")
	e := d.Entities()[0]
	_ = id
	// For n=5 degree=3, knot vector length = 5+3+1 = 9.
	if len(e.Knots) != 9 {
		t.Errorf("auto-knots length: got %d, want 9", len(e.Knots))
	}
	// Clamped: first and last (degree+1)=4 values must be 0 and 1.
	for i := 0; i <= 3; i++ {
		if e.Knots[i] != 0 {
			t.Errorf("knots[%d] should be 0 for clamped vector, got %v", i, e.Knots[i])
		}
	}
	for i := 5; i < 9; i++ {
		if e.Knots[i] != 1 {
			t.Errorf("knots[%d] should be 1 for clamped vector, got %v", i, e.Knots[i])
		}
	}
}

func TestNURBS_ToGeometryEntity(t *testing.T) {
	controls := [][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 5}, {20, 0}}
	e := document.Entity{
		Type: document.TypeNURBS, NURBSDegree: 3,
		Points: controls,
	}
	ge := e.ToGeometryEntity()
	if ge == nil {
		t.Fatal("ToGeometryEntity returned nil for NURBS")
	}
	if _, ok := ge.(geometry.NURBSEntity); !ok {
		t.Errorf("expected NURBSEntity, got %T", ge)
	}
}

func TestNURBS_Length(t *testing.T) {
	d := document.New()
	controls := [][]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	id := d.AddNURBS(3, controls, nil, nil, 0, "")
	l := d.EntityLength(id)
	if l <= 0 {
		t.Errorf("NURBS length should be positive, got %f", l)
	}
}

func TestNURBS_BoundingBox_ViaEngine(t *testing.T) {
	d := document.New()
	controls := [][]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	id := d.AddNURBS(3, controls, nil, nil, 0, "")
	bb := d.EntityBoundingBox(id)
	if bb.IsEmpty() {
		t.Error("NURBS bounding box should not be empty")
	}
}

func TestNURBS_DXFExport_R2000(t *testing.T) {
	d := document.New()
	controls := [][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 5}, {20, 0}}
	d.AddNURBS(3, controls, nil, nil, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "LWPOLYLINE") {
		t.Error("R2000 DXF for NURBS should contain LWPOLYLINE approximation")
	}
}

func TestNURBS_DXFExport_R12(t *testing.T) {
	d := document.New()
	controls := [][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 5}, {20, 0}}
	d.AddNURBS(3, controls, nil, nil, 0, "#fff")
	dxf := d.ExportDXFR12()
	if !strings.Contains(dxf, "POLYLINE") {
		t.Error("R12 DXF for NURBS should contain POLYLINE+VERTEX block")
	}
	if !strings.Contains(dxf, "VERTEX") {
		t.Error("R12 DXF for NURBS should contain VERTEX entries")
	}
}

func TestNURBS_GeometryRoundtrip(t *testing.T) {
	controls := [][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 5}, {20, 0}}
	e := document.Entity{
		Type: document.TypeNURBS, NURBSDegree: 3,
		Points: controls,
	}
	ge := e.ToGeometryEntity()
	if ge == nil {
		t.Fatal("ToGeometryEntity nil for NURBS")
	}
	back := document.GeometryEntityToDocument(ge, 1, "#aabbcc")
	if back == nil {
		t.Fatal("GeometryEntityToDocument returned nil")
	}
	if back.Type != document.TypeNURBS {
		t.Errorf("roundtrip type: got %q, want %q", back.Type, document.TypeNURBS)
	}
	if back.NURBSDegree != 3 {
		t.Errorf("roundtrip degree: got %d, want 3", back.NURBSDegree)
	}
	if len(back.Points) != 5 {
		t.Errorf("roundtrip control points: got %d, want 5", len(back.Points))
	}
	if len(back.Knots) == 0 {
		t.Error("roundtrip: knots should be preserved")
	}
}

func TestNURBS_WithExplicitWeights(t *testing.T) {
	controls := [][]float64{{0, 0}, {5, 10}, {10, 0}}
	knots := []float64{0, 0, 0, 1, 1, 1} // degree-2 clamped uniform
	weights := []float64{1, 0.5, 1}      // non-uniform weights (rational)
	d := document.New()
	id := d.AddNURBS(2, controls, knots, weights, 0, "")
	if id <= 0 {
		t.Fatalf("AddNURBS with explicit weights returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if len(e.Weights) != 3 {
		t.Errorf("weights: got %d, want 3", len(e.Weights))
	}
	if e.Weights[1] != 0.5 {
		t.Errorf("weight[1]: got %v, want 0.5", e.Weights[1])
	}
}

func TestNURBS_JSONRoundtrip(t *testing.T) {
	d := document.New()
	controls := [][]float64{{0, 0}, {5, 10}, {10, 0}, {15, 10}, {20, 0}}
	d.AddNURBS(3, controls, nil, nil, 0, "")
	json := d.ToJSON()
	if !strings.Contains(json, `"nurbs"`) {
		t.Errorf("JSON should contain type 'nurbs': %s", json)
	}
	if !strings.Contains(json, `"nurbsDeg"`) {
		t.Errorf("JSON should contain nurbsDeg field: %s", json)
	}
	if !strings.Contains(json, `"knots"`) {
		t.Errorf("JSON should contain knots field: %s", json)
	}
}

// ── Ellipse ───────────────────────────────────────────────────────────────────

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
		t.Errorf("axes: R=%v R2=%v, want 15,8", e.R, e.R2)
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
		t.Errorf("ellipse centre: %v", ee.Center)
	}
	if ee.A != 10 || ee.B != 6 {
		t.Errorf("ellipse axes: A=%v B=%v", ee.A, ee.B)
	}
}

func TestEllipse_Length(t *testing.T) {
	d := document.New()
	id := d.AddEllipse(0, 0, 10, 10, 0, 0, "")
	l := d.EntityLength(id)
	if math.Abs(l-2*math.Pi*10) > 2 {
		t.Errorf("circle-like ellipse length: got %v, want ~%v", l, 2*math.Pi*10)
	}
}

func TestEllipse_BoundingBox(t *testing.T) {
	e := document.Entity{Type: document.TypeEllipse, CX: 0, CY: 0, R: 10, R2: 5}
	bb := e.BoundingBox()
	if bb.IsEmpty() {
		t.Error("ellipse bounding box should not be empty")
	}
	if bb.Max.X < 9.9 || bb.Min.X > -9.9 {
		t.Errorf("ellipse bbox X: [%v,%v]", bb.Min.X, bb.Max.X)
	}
}

func TestEllipse_DXFExport_R2000(t *testing.T) {
	d := document.New()
	d.AddEllipse(0, 0, 10, 5, 0, 0, "#fff")
	if dxf := d.ExportDXF(); !strings.Contains(dxf, "ELLIPSE") {
		t.Error("R2000 DXF should contain ELLIPSE entity")
	}
}

func TestEllipse_DXFExport_R12(t *testing.T) {
	d := document.New()
	d.AddEllipse(0, 0, 10, 5, 0, 0, "#fff")
	dxf := d.ExportDXFR12()
	if strings.Contains(dxf, "ELLIPSE") {
		t.Error("R12 DXF should not contain native ELLIPSE entity")
	}
	if !strings.Contains(dxf, "POLYLINE") {
		t.Error("R12 DXF should approximate ellipse as POLYLINE")
	}
}

func TestEllipse_GeometryRoundtrip(t *testing.T) {
	orig := document.Entity{Type: document.TypeEllipse, CX: 3, CY: 7, R: 12, R2: 5}
	ge := orig.ToGeometryEntity()
	if ge == nil {
		t.Fatal("ToGeometryEntity nil for ellipse")
	}
	back := document.GeometryEntityToDocument(ge, 0, "")
	if back == nil {
		t.Fatal("GeometryEntityToDocument returned nil")
	}
	if back.Type != document.TypeEllipse {
		t.Errorf("roundtrip type: got %q", back.Type)
	}
	if back.R != 12 || back.R2 != 5 {
		t.Errorf("roundtrip axes: R=%v R2=%v", back.R, back.R2)
	}
}

// ── Text (single-line) ────────────────────────────────────────────────────────

func TestAddText_Basic(t *testing.T) {
	d := document.New()
	id := d.AddText(5, 10, "hello", 2.5, 0, "Romans", 0, "#ffffff")
	if id <= 0 {
		t.Fatalf("AddText returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if e.Type != document.TypeText {
		t.Errorf("type: got %q, want text", e.Type)
	}
	if e.Text != "hello" {
		t.Errorf("text: got %q, want hello", e.Text)
	}
	if e.TextHeight != 2.5 {
		t.Errorf("height: got %v, want 2.5", e.TextHeight)
	}
	if e.Font != "Romans" {
		t.Errorf("font: got %q, want Romans", e.Font)
	}
	if e.X1 != 5 || e.Y1 != 10 {
		t.Errorf("position: got (%v,%v), want (5,10)", e.X1, e.Y1)
	}
}

func TestText_Length(t *testing.T) {
	e := document.Entity{Type: document.TypeText, Text: "hello"}
	if l := e.Length(); l != 0 {
		t.Errorf("text length should be 0, got %v", l)
	}
}

func TestText_ToGeometryEntity_Nil(t *testing.T) {
	e := document.Entity{Type: document.TypeText, X1: 0, Y1: 0, Text: "A"}
	if ge := e.ToGeometryEntity(); ge != nil {
		t.Errorf("text ToGeometryEntity should return nil, got %T", ge)
	}
}

func TestText_DXFExport_FontField(t *testing.T) {
	d := document.New()
	d.AddText(0, 0, "go-cad", 3.0, 15, "Romans", 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "TEXT") {
		t.Error("DXF should contain TEXT entity")
	}
	if !strings.Contains(dxf, "go-cad") {
		t.Error("DXF TEXT should contain the text string")
	}
	// Group 7 = text style name should appear.
	if !strings.Contains(dxf, "Romans") {
		t.Error("DXF TEXT should contain font/style name (group 7)")
	}
}

func TestText_DXFExport_DefaultFont(t *testing.T) {
	d := document.New()
	d.AddText(0, 0, "test", 2.5, 0, "", 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "Standard") {
		t.Error("DXF TEXT should default to 'Standard' style when font is empty")
	}
}

// ── Multi-line Text (MTEXT) ───────────────────────────────────────────────────

func TestAddMText_Basic(t *testing.T) {
	d := document.New()
	id := d.AddMText(0, 0, "line1\nline2\nline3", 3.5, 100, 0, "Arial", 0, "#ffffff")
	if id <= 0 {
		t.Fatalf("AddMText returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if e.Type != document.TypeMText {
		t.Errorf("type: got %q, want mtext", e.Type)
	}
	if e.Text != "line1\nline2\nline3" {
		t.Errorf("text: got %q", e.Text)
	}
	if e.TextHeight != 3.5 {
		t.Errorf("height: got %v, want 3.5", e.TextHeight)
	}
	if e.R2 != 100 {
		t.Errorf("width: got %v, want 100", e.R2)
	}
	if e.Font != "Arial" {
		t.Errorf("font: got %q, want Arial", e.Font)
	}
}

func TestMText_Length(t *testing.T) {
	e := document.Entity{Type: document.TypeMText, Text: "hello\nworld"}
	if l := e.Length(); l != 0 {
		t.Errorf("mtext length should be 0, got %v", l)
	}
}

func TestMText_ToGeometryEntity_Nil(t *testing.T) {
	e := document.Entity{Type: document.TypeMText, Text: "A\nB"}
	if ge := e.ToGeometryEntity(); ge != nil {
		t.Errorf("mtext ToGeometryEntity should return nil, got %T", ge)
	}
}

func TestMText_DXFExport_R2000(t *testing.T) {
	d := document.New()
	d.AddMText(0, 0, "first line\nsecond line", 3.0, 80, 0, "Standard", 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "MTEXT") {
		t.Error("R2000 DXF should contain MTEXT entity")
	}
	// Newlines must be converted to DXF paragraph break "\\P".
	if !strings.Contains(dxf, `\P`) {
		t.Error("R2000 MTEXT should use \\P for paragraph breaks")
	}
	if !strings.Contains(dxf, "Standard") {
		t.Error("R2000 MTEXT should include style name")
	}
}

func TestMText_DXFExport_R12(t *testing.T) {
	d := document.New()
	d.AddMText(0, 0, "line1\nline2", 3.0, 0, 0, "", 0, "#fff")
	dxf := d.ExportDXFR12()
	// R12 should split MTEXT into multiple TEXT entities.
	if strings.Contains(dxf, "MTEXT") {
		t.Error("R12 DXF should not contain MTEXT entity")
	}
	if !strings.Contains(dxf, "TEXT") {
		t.Error("R12 DXF should approximate MTEXT as TEXT entities")
	}
}

func TestMText_JSONRoundtrip(t *testing.T) {
	d := document.New()
	d.AddMText(1, 2, "A\nB", 4.0, 100, 15, "Arial", 0, "")
	json := d.ToJSON()
	if !strings.Contains(json, `"mtext"`) {
		t.Errorf("JSON should contain type 'mtext': %s", json)
	}
	if !strings.Contains(json, "Arial") {
		t.Errorf("JSON should contain font name: %s", json)
	}
}

// ── Dimension types ───────────────────────────────────────────────────────────

func TestAddLinearDim(t *testing.T) {
	d := document.New()
	id := d.AddLinearDim(0, 0, 100, 0, 10, 0, "#ffffff")
	if id <= 0 {
		t.Fatalf("AddLinearDim returned invalid id %d", id)
	}
	if e := d.Entities()[0]; e.Type != document.TypeDimLinear {
		t.Errorf("type: got %q, want dimlin", e.Type)
	}
}

func TestLinearDim_Length(t *testing.T) {
	e := document.Entity{Type: document.TypeDimLinear, X1: 0, Y1: 0, X2: 30, Y2: 40}
	if l := e.Length(); math.Abs(l-50) > 1e-9 {
		t.Errorf("dimlin length: got %v, want 50", l)
	}
}

func TestLinearDim_DXFExport_R2000(t *testing.T) {
	d := document.New()
	d.AddLinearDim(0, 0, 50, 0, 20, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "DIMENSION") {
		t.Error("R2000 DXF for linear dim should contain DIMENSION entity")
	}
	if !strings.Contains(dxf, "AcDbDimension") {
		t.Error("R2000 DIMENSION should have AcDbDimension subclass marker")
	}
}

func TestLinearDim_DXFExport_R12(t *testing.T) {
	d := document.New()
	d.AddLinearDim(0, 0, 50, 0, 20, 0, "#fff")
	dxf := d.ExportDXFR12()
	if strings.Contains(dxf, "DIMENSION") {
		t.Error("R12 DXF should not contain DIMENSION entity")
	}
	if !strings.Contains(dxf, "LINE") {
		t.Error("R12 linear dim should emit LINE entities")
	}
}

func TestAddAlignedDim(t *testing.T) {
	d := document.New()
	id := d.AddAlignedDim(0, 0, 30, 40, 8, 0, "#ffffff")
	if id <= 0 {
		t.Fatalf("AddAlignedDim returned invalid id %d", id)
	}
	if e := d.Entities()[0]; e.Type != document.TypeDimAligned {
		t.Errorf("type: got %q, want dimali", e.Type)
	}
}

func TestAlignedDim_DXFExport_R2000(t *testing.T) {
	d := document.New()
	d.AddAlignedDim(0, 0, 30, 40, 5, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "DIMENSION") {
		t.Error("R2000 aligned dim should contain DIMENSION entity")
	}
	if !strings.Contains(dxf, "AcDbAlignedDimension") {
		t.Error("R2000 aligned dim should have AcDbAlignedDimension subclass")
	}
}

func TestAddAngularDim(t *testing.T) {
	d := document.New()
	id := d.AddAngularDim(0, 0, 10, 0, 0, 10, 5, 0, "#ffffff")
	if id <= 0 {
		t.Fatalf("AddAngularDim returned invalid id %d", id)
	}
	if e := d.Entities()[0]; e.Type != document.TypeDimAngular {
		t.Errorf("type: got %q, want dimang", e.Type)
	}
}

func TestAngularDim_Length(t *testing.T) {
	e := document.Entity{
		Type: document.TypeDimAngular,
		CX: 0, CY: 0, X1: 10, Y1: 0, X2: 0, Y2: 10, R: 10,
	}
	if l := e.Length(); math.Abs(l-math.Pi/2*10) > 0.5 {
		t.Errorf("angular arc length: got %v, want ~%v", l, math.Pi/2*10)
	}
}

func TestAngularDim_DXFExport_R2000(t *testing.T) {
	d := document.New()
	d.AddAngularDim(0, 0, 10, 0, 0, 10, 5, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "DIMENSION") {
		t.Error("R2000 angular dim should contain DIMENSION entity")
	}
	if !strings.Contains(dxf, "AcDb3PointAngularDimension") {
		t.Error("R2000 angular dim should have AcDb3PointAngularDimension subclass")
	}
}

func TestAngularDim_DXFExport_R12(t *testing.T) {
	d := document.New()
	d.AddAngularDim(0, 0, 10, 0, 0, 10, 5, 0, "#fff")
	dxf := d.ExportDXFR12()
	if strings.Contains(dxf, "DIMENSION") {
		t.Error("R12 DXF should not contain DIMENSION entity")
	}
	if !strings.Contains(dxf, "LINE") {
		t.Error("R12 angular dim should emit LINE entities")
	}
}

func TestAddRadialDim(t *testing.T) {
	d := document.New()
	id := d.AddRadialDim(0, 0, 15, 45, 0, "#ffffff")
	if id <= 0 {
		t.Fatalf("AddRadialDim returned invalid id %d", id)
	}
	e := d.Entities()[0]
	if e.Type != document.TypeDimRadial {
		t.Errorf("type: got %q, want dimrad", e.Type)
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

func TestRadialDim_DXFExport_R2000(t *testing.T) {
	d := document.New()
	d.AddRadialDim(0, 0, 10, 0, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "DIMENSION") {
		t.Error("R2000 radial dim should contain DIMENSION entity")
	}
	if !strings.Contains(dxf, "AcDbRadialDimension") {
		t.Error("R2000 radial dim should have AcDbRadialDimension subclass")
	}
}

func TestRadialDim_DXFExport_R12(t *testing.T) {
	d := document.New()
	d.AddRadialDim(0, 0, 10, 0, 0, "#fff")
	dxf := d.ExportDXFR12()
	// R12 approximation should have LINE (leader) and TEXT (label).
	if !strings.Contains(dxf, "LINE") {
		t.Error("R12 radial dim should emit LINE leader")
	}
	if !strings.Contains(dxf, "R10.000") {
		t.Errorf("R12 radial dim should contain radius label, got: %s", dxf)
	}
}

func TestAddDiameterDim(t *testing.T) {
	d := document.New()
	id := d.AddDiameterDim(5, 5, 12, 0, 0, "#00ff00")
	if id <= 0 {
		t.Fatalf("AddDiameterDim returned invalid id %d", id)
	}
	if e := d.Entities()[0]; e.Type != document.TypeDimDiameter {
		t.Errorf("type: got %q, want dimdia", e.Type)
	}
}

func TestDiameterDim_Length(t *testing.T) {
	e := document.Entity{Type: document.TypeDimDiameter, R: 5}
	if l := e.Length(); l != 10 {
		t.Errorf("diameter dim length: got %v, want 10", l)
	}
}

func TestDiameterDim_DXFExport_R2000(t *testing.T) {
	d := document.New()
	d.AddDiameterDim(0, 0, 8, 0, 0, "#fff")
	dxf := d.ExportDXF()
	if !strings.Contains(dxf, "DIMENSION") {
		t.Error("R2000 diameter dim should contain DIMENSION entity")
	}
	if !strings.Contains(dxf, "AcDbDiametricDimension") {
		t.Error("R2000 diameter dim should have AcDbDiametricDimension subclass")
	}
}

// ── AddEntity dispatch for all new types ──────────────────────────────────────

func TestAddEntity_AllNewTypes(t *testing.T) {
	cases := []document.Entity{
		{Type: document.TypeSpline, Points: [][]float64{{0, 0}, {1, 2}, {3, 4}, {5, 0}}},
		{Type: document.TypeNURBS, NURBSDegree: 3, Points: [][]float64{{0, 0}, {5, 10}, {10, 0}, {15, 10}, {20, 0}}},
		{Type: document.TypeEllipse, CX: 0, CY: 0, R: 5, R2: 3},
		{Type: document.TypeText, X1: 0, Y1: 0, Text: "hi", TextHeight: 2},
		{Type: document.TypeMText, X1: 0, Y1: 0, Text: "a\nb", TextHeight: 2, R2: 50},
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
			t.Errorf("AddEntity(%q): expected 1 entity, got %d", tc.Type, d.EntityCount())
		}
	}
}

func TestAddEntity_Unknown(t *testing.T) {
	d := document.New()
	if id := d.AddEntity(document.Entity{Type: "xyzzy"}); id != -1 {
		t.Errorf("AddEntity(unknown) should return -1, got %d", id)
	}
}

// ── DXF version headers ───────────────────────────────────────────────────────

func TestExportDXF_R2000Header(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 10, 10, 0, "")
	if dxf := d.ExportDXF(); !strings.Contains(dxf, "AC1015") {
		t.Error("ExportDXF should use AC1015 (R2000) version header")
	}
}

func TestExportDXFR12_Header(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 10, 10, 0, "")
	if dxf := d.ExportDXFR12(); !strings.Contains(dxf, "AC1009") {
		t.Error("ExportDXFR12 should use AC1009 (R12) version header")
	}
}

// ── DXF completeness (all new types) ─────────────────────────────────────────

func TestExportDXF_AllNewTypes(t *testing.T) {
	d := document.New()
	d.AddSpline([][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 0}}, 0, "#fff")
	d.AddNURBS(3, [][]float64{{0, 0}, {5, 10}, {10, 10}, {15, 5}, {20, 0}}, nil, nil, 0, "#fff")
	d.AddEllipse(0, 0, 10, 5, 0, 0, "#fff")
	d.AddText(0, 0, "label", 3, 0, "Standard", 0, "#fff")
	d.AddMText(0, 0, "line1\nline2", 3, 80, 0, "Standard", 0, "#fff")
	d.AddLinearDim(0, 0, 50, 0, 10, 0, "#fff")
	d.AddAlignedDim(0, 0, 30, 40, 5, 0, "#fff")
	d.AddAngularDim(0, 0, 10, 0, 0, 10, 5, 0, "#fff")
	d.AddRadialDim(0, 0, 8, 0, 0, "#fff")
	d.AddDiameterDim(0, 0, 6, 0, 0, "#fff")

	dxf := d.ExportDXF()
	for _, want := range []string{
		"AC1015", "LWPOLYLINE", "ELLIPSE", "TEXT", "MTEXT",
		"DIMENSION", "AcDbDimension", "SECTION", "ENTITIES", "EOF",
	} {
		if !strings.Contains(dxf, want) {
			t.Errorf("R2000 DXF missing %q", want)
		}
	}

	dxfR12 := d.ExportDXFR12()
	for _, want := range []string{
		"AC1009", "POLYLINE", "VERTEX", "SEQEND", "TEXT", "LINE", "EOF",
	} {
		if !strings.Contains(dxfR12, want) {
			t.Errorf("R12 DXF missing %q", want)
		}
	}
	for _, notWant := range []string{"MTEXT", "LWPOLYLINE", "ELLIPSE", "DIMENSION"} {
		if strings.Contains(dxfR12, notWant) {
			t.Errorf("R12 DXF should not contain %q (R2000-only entity)", notWant)
		}
	}
}

// ── Undo/redo with new types ──────────────────────────────────────────────────

func TestUndoRedo_Spline(t *testing.T) {
	d := document.New()
	d.AddSpline([][]float64{{0, 0}, {1, 2}, {3, 4}, {5, 0}}, 0, "")
	d.Undo()
	if d.EntityCount() != 0 {
		t.Errorf("after undo: got %d entities, want 0", d.EntityCount())
	}
	d.Redo()
	if d.EntityCount() != 1 {
		t.Errorf("after redo: got %d entities, want 1", d.EntityCount())
	}
}

func TestUndoRedo_NURBS(t *testing.T) {
	d := document.New()
	d.AddNURBS(3, [][]float64{{0, 0}, {5, 10}, {10, 0}, {15, 10}, {20, 0}}, nil, nil, 0, "")
	d.Undo()
	if d.EntityCount() != 0 {
		t.Errorf("after undo: got %d, want 0", d.EntityCount())
	}
}

// ── Geometry engine integration ───────────────────────────────────────────────

func TestSpline_GeometryBoundingBox_ViaEngine(t *testing.T) {
	d := document.New()
	id := d.AddSpline([][]float64{{0, 0}, {0, 20}, {30, 20}, {30, 0}}, 0, "")
	bb := d.EntityBoundingBox(id)
	if bb.IsEmpty() {
		t.Error("EntityBoundingBox empty for spline")
	}
	if bb.Max.X < 25 {
		t.Errorf("spline bbox MaxX: got %v, want ≥25", bb.Max.X)
	}
}

func TestNURBS_GeometryBoundingBox_ViaEngine(t *testing.T) {
	d := document.New()
	id := d.AddNURBS(3, [][]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}, nil, nil, 0, "")
	bb := d.EntityBoundingBox(id)
	if bb.IsEmpty() {
		t.Error("NURBS EntityBoundingBox should not be empty")
	}
}

func TestEllipse_GeometryBoundingBox_ViaEngine(t *testing.T) {
	d := document.New()
	id := d.AddEllipse(0, 0, 10, 5, 0, 0, "")
	bb := d.EntityBoundingBox(id)
	if bb.IsEmpty() {
		t.Error("EntityBoundingBox empty for ellipse")
	}
	if bb.Max.X < 9.9 {
		t.Errorf("ellipse bbox MaxX: got %v, want ≥9.9", bb.Max.X)
	}
}
