package svg_test

import (
	"strings"
	"testing"

	"github.com/tomott12345/go-cad/internal/document"
	"github.com/tomott12345/go-cad/pkg/svg"
)

func TestGenerateEmpty(t *testing.T) {
	d := document.New()
	s, err := svg.Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "<svg") {
		t.Error("expected <svg in output")
	}
	if !strings.Contains(s, "</svg>") {
		t.Error("expected </svg> in output")
	}
}

func TestGenerateLine(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 100, 100, 0, "#ff0000")
	s, err := svg.Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "<line") {
		t.Error("expected <line in SVG output")
	}
}

func TestGenerateCircle(t *testing.T) {
	d := document.New()
	d.AddCircle(50, 50, 25, 0, "#00ff00")
	s, err := svg.Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "<circle") {
		t.Error("expected <circle in SVG output")
	}
}

func TestGenerateArc(t *testing.T) {
	d := document.New()
	d.AddArc(0, 0, 10, 0, 90, 0, "#ffffff")
	s, err := svg.Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "<path") {
		t.Error("expected <path for arc in SVG output")
	}
}

func TestGenerateRect(t *testing.T) {
	d := document.New()
	d.AddRectangle(0, 0, 100, 100, 0, "#ffffff")
	s, err := svg.Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "<rect") {
		t.Error("expected <rect in SVG output")
	}
}

func TestGenerateText(t *testing.T) {
	d := document.New()
	d.AddText(10, 10, "Hello SVG", 3, 0, "", 0, "#ffffff")
	s, err := svg.Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "Hello SVG") {
		t.Error("expected text content in SVG output")
	}
}

func TestGenerateLayers(t *testing.T) {
	d := document.New()
	layID := d.AddLayer("Walls", "#ff8800", document.LineTypeDashed, 0.5)
	d.AddLine(0, 0, 10, 0, layID, "#ff8800")
	s, err := svg.Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "Walls") {
		t.Error("expected layer name in SVG output")
	}
	if !strings.Contains(s, "stroke-dasharray") {
		t.Error("expected stroke-dasharray for dashed layer")
	}
}

func TestGeneratePolyline(t *testing.T) {
	d := document.New()
	d.AddPolyline([][]float64{{0, 0}, {10, 0}, {10, 10}}, 0, "#ffffff")
	s, err := svg.Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "<polyline") {
		t.Error("expected <polyline in SVG output")
	}
}

func TestXMLEscaping(t *testing.T) {
	d := document.New()
	d.AddText(0, 0, "a & b < c > d", 3, 0, "", 0, "#ffffff")
	s, err := svg.Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(s, " & ") {
		t.Error("raw & should be escaped as &amp;")
	}
	if !strings.Contains(s, "&amp;") {
		t.Error("expected &amp; in SVG output")
	}
}

func TestViewBoxContainsEntities(t *testing.T) {
	d := document.New()
	d.AddLine(0, 0, 200, 150, 0, "#ffffff")
	s, err := svg.Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	// viewBox should encompass the line extents
	if !strings.Contains(s, "viewBox") {
		t.Error("expected viewBox attribute")
	}
}
