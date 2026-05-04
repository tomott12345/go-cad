package hatch

import (
	"math"
	"testing"
)

func poly(pts ...[2]float64) [][]float64 {
	out := make([][]float64, len(pts))
	for i, p := range pts {
		out[i] = []float64{p[0], p[1]}
	}
	return out
}

func TestGenerateLinesANSI31(t *testing.T) {
	boundary := poly([2]float64{0, 0}, [2]float64{100, 0}, [2]float64{50, 100})
	segs := GenerateLines(boundary, "ANSI31", 0, 10)
	if len(segs) == 0 {
		t.Fatal("expected hatch lines for ANSI31 pattern in triangle, got none")
	}
	for _, s := range segs {
		for _, y := range []float64{s[1], s[3]} {
			if y < -0.1 || y > 100.1 {
				t.Errorf("segment Y=%v out of boundary [0,100]", y)
			}
		}
	}
}

func TestGenerateLinesAngle90(t *testing.T) {
	boundary := poly([2]float64{0, 0}, [2]float64{100, 0}, [2]float64{100, 100}, [2]float64{0, 100})
	segs := GenerateLines(boundary, "ANSI31", 90, 10)
	if len(segs) == 0 {
		t.Fatal("expected hatch lines for 90° rotation")
	}
}

func TestGenerateLinesANSI32Cross(t *testing.T) {
	boundary := poly([2]float64{0, 0}, [2]float64{60, 0}, [2]float64{60, 60}, [2]float64{0, 60})
	segs := GenerateLines(boundary, "ANSI32", 0, 8)
	if len(segs) == 0 {
		t.Fatal("expected cross-hatch lines")
	}
	segs0 := GenerateLines(boundary, "ANSI31", 0, 8)
	if len(segs) < len(segs0) {
		t.Errorf("ANSI32 should have >= segments vs ANSI31, got %d vs %d", len(segs), len(segs0))
	}
}

func TestGenerateLinesSOLID(t *testing.T) {
	boundary := poly([2]float64{0, 0}, [2]float64{50, 0}, [2]float64{50, 50}, [2]float64{0, 50})
	// SOLID just calls scanFill with 0.5× scale — should not panic.
	segs := GenerateLines(boundary, "SOLID", 0, 1)
	if len(segs) == 0 {
		t.Fatal("SOLID should generate fill segments")
	}
}

func TestGenerateLinesDOTS(t *testing.T) {
	boundary := poly([2]float64{0, 0}, [2]float64{80, 0}, [2]float64{80, 80}, [2]float64{0, 80})
	segs := GenerateLines(boundary, "DOTS", 0, 10)
	if len(segs) == 0 {
		t.Fatal("expected dot segments for DOTS pattern")
	}
}

func TestGenerateLinesScaleEffect(t *testing.T) {
	boundary := poly([2]float64{0, 0}, [2]float64{200, 0}, [2]float64{200, 200}, [2]float64{0, 200})
	s5 := GenerateLines(boundary, "ANSI31", 0, 5)
	s20 := GenerateLines(boundary, "ANSI31", 0, 20)
	if len(s5) <= len(s20) {
		t.Errorf("scale=5 should produce more lines than scale=20, got %d vs %d", len(s5), len(s20))
	}
}

func TestGenerateLinesDegenerate(t *testing.T) {
	segs := GenerateLines([][]float64{}, "ANSI31", 0, 10)
	if len(segs) != 0 {
		t.Errorf("empty boundary should yield no segments, got %d", len(segs))
	}
	segs2 := GenerateLines(poly([2]float64{0, 0}, [2]float64{10, 10}), "ANSI31", 0, 10)
	_ = segs2
}

func TestGenerateLinesNonZeroLength(t *testing.T) {
	boundary := poly([2]float64{0, 0}, [2]float64{100, 0}, [2]float64{100, 80}, [2]float64{0, 80})
	segs := GenerateLines(boundary, "ANSI31", 0, 7)
	for i, s := range segs {
		l := math.Hypot(s[2]-s[0], s[3]-s[1])
		if l < 1e-9 {
			t.Errorf("segment %d has zero length", i)
		}
	}
}

func TestGenerateLinesUnknownFallsBackToANSI31(t *testing.T) {
	boundary := poly([2]float64{0, 0}, [2]float64{50, 0}, [2]float64{50, 50}, [2]float64{0, 50})
	segsUnknown := GenerateLines(boundary, "FOOBAR", 0, 10)
	segsANSI31 := GenerateLines(boundary, "ANSI31", 0, 10)
	if len(segsUnknown) != len(segsANSI31) {
		t.Errorf("unknown pattern should fall back to ANSI31: got %d vs %d segs", len(segsUnknown), len(segsANSI31))
	}
}
