// Package hatch implements a scanline polygon fill engine for CAD hatch patterns.
//
// Supported patterns:
//   SOLID   — dense fill at 0.5-unit spacing
//   ANSI31  — standard 45° hatching
//   ANSI32  — cross-hatching (45° + 135°)
//   DOTS    — sparse dot pattern (approximated as very short segments)
//
// All patterns are generated purely from the polygon boundary supplied by the
// caller; no rasterisation or image operations are involved.
package hatch

import (
	"math"
	"sort"
)

// Pattern name constants.
const (
	PatternSolid  = "SOLID"
	PatternANSI31 = "ANSI31"
	PatternANSI32 = "ANSI32"
	PatternDots   = "DOTS"
)

// Segment is a line segment: [x1, y1, x2, y2].
type Segment [4]float64

// GenerateLines fills a closed polygon with hatch line segments.
//
// polygon: boundary vertex list [[x,y], …] — automatically closed (first==last
// is fine but not required).
// pattern: one of the Pattern* constants (case-insensitive).
// angleDeg: additional rotation angle applied to the hatch pattern (degrees).
// scale: spacing between hatch lines (must be > 0).
//
// Returns a (possibly nil) slice of line segments ready for canvas rendering
// or DXF export.
func GenerateLines(polygon [][]float64, pattern string, angleDeg, scale float64) []Segment {
	if len(polygon) < 3 || scale <= 0 {
		return nil
	}
	switch normalise(pattern) {
	case "SOLID":
		return scanFill(polygon, angleDeg, scale*0.5)
	case "ANSI31":
		return scanFill(polygon, angleDeg+45, scale)
	case "ANSI32":
		a := scanFill(polygon, angleDeg+45, scale)
		b := scanFill(polygon, angleDeg+135, scale)
		return append(a, b...)
	case "DOTS":
		// Dots: sparse 45° lines with a large gap; very short segments mimic dots.
		return dotFill(polygon, angleDeg, scale*3)
	default:
		// Unknown pattern: fall back to ANSI31.
		return scanFill(polygon, angleDeg+45, scale)
	}
}

// ─── Core scanline fill ───────────────────────────────────────────────────────

func scanFill(polygon [][]float64, angleDeg, spacing float64) []Segment {
	// Rotate polygon by -angleDeg so scan lines become horizontal.
	sinA, cosA := math.Sincos(angleDeg * math.Pi / 180)
	rotate := func(x, y float64) (float64, float64) {
		return x*cosA + y*sinA, -x*sinA + y*cosA
	}
	unrotate := func(x, y float64) (float64, float64) {
		return x*cosA - y*sinA, x*sinA + y*cosA
	}

	// Build rotated polygon.
	n := len(polygon)
	rx := make([]float64, n)
	ry := make([]float64, n)
	minY, maxY := math.Inf(1), math.Inf(-1)
	for i, p := range polygon {
		rx[i], ry[i] = rotate(p[0], p[1])
		if ry[i] < minY {
			minY = ry[i]
		}
		if ry[i] > maxY {
			maxY = ry[i]
		}
	}

	var segs []Segment
	// Snap start so the lines tile consistently from origin.
	startY := math.Ceil(minY/spacing) * spacing
	for y := startY; y <= maxY+1e-9; y += spacing {
		// Even-odd scanline: find all X intersections at this Y.
		var xs []float64
		for i := 0; i < n; i++ {
			j := (i + 1) % n
			y0, y1 := ry[i], ry[j]
			if y0 == y1 {
				continue // horizontal edge — skip
			}
			if (y < math.Min(y0, y1)) || (y > math.Max(y0, y1)) {
				continue
			}
			t := (y - y0) / (y1 - y0)
			xi := rx[i] + t*(rx[j]-rx[i])
			xs = append(xs, xi)
		}
		if len(xs) < 2 {
			continue
		}
		sort.Float64s(xs)
		// Pair up intersections (even-odd rule).
		for k := 0; k+1 < len(xs); k += 2 {
			x1, x2 := xs[k], xs[k+1]
			if x2-x1 < 1e-9 {
				continue
			}
			wx1, wy1 := unrotate(x1, y)
			wx2, wy2 := unrotate(x2, y)
			segs = append(segs, Segment{wx1, wy1, wx2, wy2})
		}
	}
	return segs
}

// dotFill generates sparse short segments to simulate a dot pattern.
func dotFill(polygon [][]float64, angleDeg, spacing float64) []Segment {
	full := scanFill(polygon, angleDeg+45, spacing)
	var out []Segment
	dotLen := spacing * 0.05
	if dotLen < 0.1 {
		dotLen = 0.1
	}
	for _, s := range full {
		mx := (s[0] + s[2]) / 2
		my := (s[1] + s[3]) / 2
		out = append(out, Segment{mx - dotLen, my, mx + dotLen, my})
	}
	return out
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func normalise(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		out = append(out, c)
	}
	return string(out)
}
