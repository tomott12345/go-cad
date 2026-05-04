// Package dxf provides DXF import (R12 / R2000) and export for go-cad documents.
//
// Read parses a DXF text stream into a *document.Document. Write wraps the
// existing document export helpers to keep a single public surface area.
package dxf

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"go-cad/internal/document"
)

// ─── Group-code pair ──────────────────────────────────────────────────────────

type gc struct {
	code int
	val  string
}

func (g gc) float() float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(g.val), 64)
	return v
}
func (g gc) int() int {
	v, _ := strconv.Atoi(strings.TrimSpace(g.val))
	return v
}
func (g gc) str() string { return strings.TrimSpace(g.val) }

// readGCs reads all group-code/value pairs from r.
// Each pair is two lines: an integer code line followed by a value line.
// Malformed lines are skipped gracefully (error-recovery).
func readGCs(r io.Reader) ([]gc, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	var pairs []gc
	for scanner.Scan() {
		codeLine := strings.TrimSpace(scanner.Text())
		if !scanner.Scan() {
			break
		}
		valLine := strings.TrimRight(scanner.Text(), "\r\n \t")
		code, err := strconv.Atoi(codeLine)
		if err != nil {
			continue
		}
		pairs = append(pairs, gc{code, valLine})
	}
	return pairs, scanner.Err()
}

// ─── Layer and block state ────────────────────────────────────────────────────

type layerInfo struct {
	color   string
	lineTyp document.LineType
	visible bool
	locked  bool
}

// ─── Parser ───────────────────────────────────────────────────────────────────

type parser struct {
	gcs    []gc
	pos    int
	doc    *document.Document
	layers map[string]layerInfo  // name → properties from TABLES section
	blocks map[string][]document.Entity // block name → pre-parsed entities
	layerIDs map[string]int           // DXF layer name → document layer ID
	warnings []string
}

func newParser(gcs []gc) *parser {
	return &parser{
		gcs:      gcs,
		doc:      document.New(),
		layers:   map[string]layerInfo{"0": {color: "#ffffff", lineTyp: document.LineTypeSolid, visible: true}},
		blocks:   map[string][]document.Entity{},
		layerIDs: map[string]int{"0": 0},
	}
}

func (p *parser) warn(format string, args ...any) {
	p.warnings = append(p.warnings, fmt.Sprintf(format, args...))
}

func (p *parser) peek() (gc, bool) {
	if p.pos >= len(p.gcs) {
		return gc{}, false
	}
	return p.gcs[p.pos], true
}

func (p *parser) next() (gc, bool) {
	g, ok := p.peek()
	if ok {
		p.pos++
	}
	return g, ok
}

// collectEntityCodes gathers group codes for the current entity (reads until
// the next code=0, without consuming it). Returns the collected codes.
func (p *parser) collectEntityCodes() []gc {
	var codes []gc
	for {
		g, ok := p.peek()
		if !ok || g.code == 0 {
			return codes
		}
		p.next()
		codes = append(codes, g)
	}
}

// skipToCode0 advances until it finds a code=0 pair equal to name (case-insensitive),
// or EOF. Returns true if found.
func (p *parser) skipToCode0(name string) bool {
	for {
		g, ok := p.next()
		if !ok {
			return false
		}
		if g.code == 0 && strings.EqualFold(g.str(), name) {
			return true
		}
	}
}

// ─── Main parse entry ─────────────────────────────────────────────────────────

func (p *parser) parse() {
	for {
		g, ok := p.next()
		if !ok {
			break
		}
		if g.code != 0 || !strings.EqualFold(g.str(), "SECTION") {
			continue
		}
		nameGC, ok2 := p.next()
		if !ok2 {
			break
		}
		switch strings.ToUpper(nameGC.str()) {
		case "HEADER":
			p.parseHeader()
		case "TABLES":
			p.parseTables()
		case "BLOCKS":
			p.parseBlocks()
		case "ENTITIES":
			p.parseEntities(false)
		default:
			p.skipToCode0("ENDSEC")
		}
	}
	p.applyLayers()
}

func (p *parser) parseHeader() {
	for {
		g, ok := p.next()
		if !ok || (g.code == 0 && strings.EqualFold(g.str(), "ENDSEC")) {
			return
		}
	}
}

// ─── TABLES section ───────────────────────────────────────────────────────────

func (p *parser) parseTables() {
	for {
		g, ok := p.next()
		if !ok || (g.code == 0 && strings.EqualFold(g.str(), "ENDSEC")) {
			return
		}
		if g.code != 0 || !strings.EqualFold(g.str(), "TABLE") {
			continue
		}
		ng, ok2 := p.next()
		if !ok2 {
			return
		}
		switch strings.ToUpper(ng.str()) {
		case "LAYER":
			p.parseLayerTable()
		default:
			p.skipToCode0("ENDTAB")
		}
	}
}

func (p *parser) parseLayerTable() {
	for {
		g, ok := p.next()
		if !ok || (g.code == 0 && strings.EqualFold(g.str(), "ENDTAB")) {
			return
		}
		if g.code == 0 && strings.EqualFold(g.str(), "LAYER") {
			p.parseLayerEntry()
		}
	}
}

func (p *parser) parseLayerEntry() {
	codes := p.collectEntityCodes()
	info := layerInfo{color: "#ffffff", lineTyp: document.LineTypeSolid, visible: true}
	name := ""
	for _, g := range codes {
		switch g.code {
		case 2:
			name = g.str()
		case 62:
			aci := g.int()
			if aci < 0 {
				aci = -aci
			}
			info.color = aciToRGB(aci)
		case 6:
			info.lineTyp = dxfLTNameToLineType(g.str())
		case 70:
			flags := g.int()
			if flags&1 != 0 {
				info.visible = false
			}
			if flags&4 != 0 {
				info.locked = true
			}
		}
	}
	if name != "" {
		p.layers[name] = info
	}
}

// ─── BLOCKS section ───────────────────────────────────────────────────────────

func (p *parser) parseBlocks() {
	for {
		g, ok := p.next()
		if !ok || (g.code == 0 && strings.EqualFold(g.str(), "ENDSEC")) {
			return
		}
		if g.code == 0 && strings.EqualFold(g.str(), "BLOCK") {
			p.parseBlockDef()
		}
	}
}

func (p *parser) parseBlockDef() {
	name := ""
	codes := p.collectEntityCodes()
	for _, g := range codes {
		if g.code == 2 {
			name = g.str()
		}
	}
	var ents []document.Entity
	for {
		g, ok := p.next()
		if !ok || (g.code == 0 && strings.EqualFold(g.str(), "ENDBLK")) {
			break
		}
		if g.code == 0 {
			e, ok2 := p.parseEntityByType(g.str())
			if ok2 {
				ents = append(ents, e)
			}
		}
	}
	if name != "" && !strings.HasPrefix(name, "*") {
		p.blocks[name] = ents
	}
}

// ─── ENTITIES section ─────────────────────────────────────────────────────────

// parseEntities reads entities from the current section and adds them to the
// document. If inBlock is true, entities are returned rather than added.
func (p *parser) parseEntities(inBlock bool) {
	for {
		g, ok := p.next()
		if !ok || (g.code == 0 && strings.EqualFold(g.str(), "ENDSEC")) {
			return
		}
		if g.code != 0 {
			continue
		}
		typeName := strings.ToUpper(g.str())
		if typeName == "INSERT" {
			p.parseInsertAndExpand()
			continue
		}
		e, ok2 := p.parseEntityByType(typeName)
		if ok2 {
			p.doc.AddEntity(e)
		}
	}
}

func (p *parser) parseEntityByType(typeName string) (document.Entity, bool) {
	switch strings.ToUpper(typeName) {
	case "LINE":
		return p.parseLine()
	case "CIRCLE":
		return p.parseCircle()
	case "ARC":
		return p.parseArc()
	case "LWPOLYLINE":
		return p.parseLWPolyline()
	case "POLYLINE":
		return p.parsePolylinePV()
	case "SPLINE":
		return p.parseSpline()
	case "ELLIPSE":
		return p.parseEllipse()
	case "TEXT":
		return p.parseText()
	case "MTEXT":
		return p.parseMText()
	case "DIMENSION":
		return p.parseDimension()
	default:
		p.collectEntityCodes()
		return document.Entity{}, false
	}
}

// ─── Per-entity parsers ───────────────────────────────────────────────────────

func (p *parser) parseLine() (document.Entity, bool) {
	codes := p.collectEntityCodes()
	var e document.Entity
	e.Type = document.TypeLine
	layerName := "0"
	for _, g := range codes {
		switch g.code {
		case 8:
			layerName = g.str()
		case 10:
			e.X1 = g.float()
		case 20:
			e.Y1 = -g.float()
		case 11:
			e.X2 = g.float()
		case 21:
			e.Y2 = -g.float()
		case 62:
			aci := g.int()
			if aci != 0 {
				e.Color = aciToRGB(absInt(aci))
			}
		}
	}
	e.Layer = p.layerID(layerName)
	if e.Color == "" {
		e.Color = p.layerColor(layerName)
	}
	return e, true
}

func (p *parser) parseCircle() (document.Entity, bool) {
	codes := p.collectEntityCodes()
	var e document.Entity
	e.Type = document.TypeCircle
	layerName := "0"
	for _, g := range codes {
		switch g.code {
		case 8:
			layerName = g.str()
		case 10:
			e.CX = g.float()
		case 20:
			e.CY = -g.float()
		case 40:
			e.R = g.float()
		case 62:
			aci := g.int()
			if aci != 0 {
				e.Color = aciToRGB(absInt(aci))
			}
		}
	}
	e.Layer = p.layerID(layerName)
	if e.Color == "" {
		e.Color = p.layerColor(layerName)
	}
	return e, true
}

func (p *parser) parseArc() (document.Entity, bool) {
	codes := p.collectEntityCodes()
	var e document.Entity
	e.Type = document.TypeArc
	layerName := "0"
	var startDXF, endDXF float64
	for _, g := range codes {
		switch g.code {
		case 8:
			layerName = g.str()
		case 10:
			e.CX = g.float()
		case 20:
			e.CY = -g.float()
		case 40:
			e.R = g.float()
		case 50:
			startDXF = g.float()
		case 51:
			endDXF = g.float()
		case 62:
			aci := g.int()
			if aci != 0 {
				e.Color = aciToRGB(absInt(aci))
			}
		}
	}
	// DXF arc: CCW in Cartesian (Y-up). After Y-flip the arc reverses,
	// so negate and swap angles to preserve visual shape.
	e.StartDeg = normalizeAngle(-endDXF)
	e.EndDeg = normalizeAngle(-startDXF)
	e.Layer = p.layerID(layerName)
	if e.Color == "" {
		e.Color = p.layerColor(layerName)
	}
	return e, true
}

func (p *parser) parseLWPolyline() (document.Entity, bool) {
	codes := p.collectEntityCodes()
	var e document.Entity
	e.Type = document.TypePolyline
	layerName := "0"
	var xs, ys []float64
	for _, g := range codes {
		switch g.code {
		case 8:
			layerName = g.str()
		case 10:
			xs = append(xs, g.float())
		case 20:
			ys = append(ys, -g.float())
		case 62:
			aci := g.int()
			if aci != 0 {
				e.Color = aciToRGB(absInt(aci))
			}
		}
	}
	n := len(xs)
	if n < len(ys) {
		n = len(ys)
	}
	pts := make([][]float64, 0, n)
	for i := 0; i < n; i++ {
		x, y := 0.0, 0.0
		if i < len(xs) {
			x = xs[i]
		}
		if i < len(ys) {
			y = ys[i]
		}
		pts = append(pts, []float64{x, y})
	}
	e.Points = pts
	e.Layer = p.layerID(layerName)
	if e.Color == "" {
		e.Color = p.layerColor(layerName)
	}
	return e, len(pts) >= 2
}

// parsePolylinePV handles old-style POLYLINE/VERTEX/SEQEND sequences.
func (p *parser) parsePolylinePV() (document.Entity, bool) {
	codes := p.collectEntityCodes()
	var e document.Entity
	e.Type = document.TypePolyline
	layerName := "0"
	for _, g := range codes {
		switch g.code {
		case 8:
			layerName = g.str()
		case 62:
			aci := g.int()
			if aci != 0 {
				e.Color = aciToRGB(absInt(aci))
			}
		}
	}
	var pts [][]float64
	for {
		g, ok := p.peek()
		if !ok || g.code != 0 {
			break
		}
		typeName := strings.ToUpper(g.str())
		if typeName == "SEQEND" {
			p.next()
			p.collectEntityCodes()
			break
		}
		if typeName != "VERTEX" {
			break
		}
		p.next()
		vcodes := p.collectEntityCodes()
		vx, vy := 0.0, 0.0
		for _, vg := range vcodes {
			switch vg.code {
			case 10:
				vx = vg.float()
			case 20:
				vy = -vg.float()
			}
		}
		pts = append(pts, []float64{vx, vy})
	}
	e.Points = pts
	e.Layer = p.layerID(layerName)
	if e.Color == "" {
		e.Color = p.layerColor(layerName)
	}
	return e, len(pts) >= 2
}

func (p *parser) parseSpline() (document.Entity, bool) {
	codes := p.collectEntityCodes()
	var e document.Entity
	e.Type = document.TypeNURBS
	layerName := "0"
	var knots, weights []float64
	var ctrlX, ctrlY []float64
	degree := 3
	for _, g := range codes {
		switch g.code {
		case 8:
			layerName = g.str()
		case 71:
			degree = g.int()
		case 40:
			knots = append(knots, g.float())
		case 41:
			weights = append(weights, g.float())
		case 10:
			ctrlX = append(ctrlX, g.float())
		case 20:
			ctrlY = append(ctrlY, -g.float())
		case 62:
			aci := g.int()
			if aci != 0 {
				e.Color = aciToRGB(absInt(aci))
			}
		}
	}
	n := len(ctrlX)
	if n < len(ctrlY) {
		n = len(ctrlY)
	}
	pts := make([][]float64, n)
	for i := 0; i < n; i++ {
		x, y := 0.0, 0.0
		if i < len(ctrlX) {
			x = ctrlX[i]
		}
		if i < len(ctrlY) {
			y = ctrlY[i]
		}
		pts[i] = []float64{x, y}
	}
	e.Points = pts
	e.NURBSDegree = degree
	e.Knots = knots
	if len(weights) == n {
		e.Weights = weights
	}
	e.Layer = p.layerID(layerName)
	if e.Color == "" {
		e.Color = p.layerColor(layerName)
	}
	return e, n >= 2
}

func (p *parser) parseEllipse() (document.Entity, bool) {
	codes := p.collectEntityCodes()
	var e document.Entity
	e.Type = document.TypeEllipse
	layerName := "0"
	var cx, cy, majX, majY, ratio float64
	ratio = 1.0
	for _, g := range codes {
		switch g.code {
		case 8:
			layerName = g.str()
		case 10:
			cx = g.float()
		case 20:
			cy = -g.float()
		case 11:
			majX = g.float()
		case 21:
			majY = -g.float()
		case 40:
			ratio = g.float()
		case 62:
			aci := g.int()
			if aci != 0 {
				e.Color = aciToRGB(absInt(aci))
			}
		}
	}
	// Major axis length and rotation from the (majX, majY) endpoint vector.
	semiMajor := math.Hypot(majX, majY)
	rotDeg := math.Atan2(majY, majX) * 180 / math.Pi
	semiMinor := semiMajor * ratio
	e.CX, e.CY = cx, cy
	e.R, e.R2, e.RotDeg = semiMajor, semiMinor, rotDeg
	e.Layer = p.layerID(layerName)
	if e.Color == "" {
		e.Color = p.layerColor(layerName)
	}
	return e, semiMajor > 1e-12
}

func (p *parser) parseText() (document.Entity, bool) {
	codes := p.collectEntityCodes()
	var e document.Entity
	e.Type = document.TypeText
	layerName := "0"
	for _, g := range codes {
		switch g.code {
		case 8:
			layerName = g.str()
		case 10:
			e.X1 = g.float()
		case 20:
			e.Y1 = -g.float()
		case 40:
			e.TextHeight = g.float()
		case 1:
			e.Text = g.str()
		case 50:
			e.RotDeg = g.float()
		case 7:
			e.Font = g.str()
		case 62:
			aci := g.int()
			if aci != 0 {
				e.Color = aciToRGB(absInt(aci))
			}
		}
	}
	if e.TextHeight <= 0 {
		e.TextHeight = 2.5
	}
	e.Layer = p.layerID(layerName)
	if e.Color == "" {
		e.Color = p.layerColor(layerName)
	}
	return e, e.Text != ""
}

func (p *parser) parseMText() (document.Entity, bool) {
	codes := p.collectEntityCodes()
	var e document.Entity
	e.Type = document.TypeMText
	layerName := "0"
	for _, g := range codes {
		switch g.code {
		case 8:
			layerName = g.str()
		case 10:
			e.X1 = g.float()
		case 20:
			e.Y1 = -g.float()
		case 40:
			e.TextHeight = g.float()
		case 41:
			e.R2 = g.float()
		case 50:
			e.RotDeg = g.float()
		case 7:
			e.Font = g.str()
		case 1, 3:
			// Group 3 = additional text chunks (overflow), 1 = primary text.
			chunk := strings.ReplaceAll(g.str(), "\\P", "\n")
			if e.Text == "" {
				e.Text = chunk
			} else {
				e.Text += chunk
			}
		case 62:
			aci := g.int()
			if aci != 0 {
				e.Color = aciToRGB(absInt(aci))
			}
		}
	}
	if e.TextHeight <= 0 {
		e.TextHeight = 2.5
	}
	e.Layer = p.layerID(layerName)
	if e.Color == "" {
		e.Color = p.layerColor(layerName)
	}
	return e, true
}

func (p *parser) parseDimension() (document.Entity, bool) {
	codes := p.collectEntityCodes()
	layerName := "0"
	dimType := 0
	var x1, y1, x2, y2, cx, cy, dlX, dlY, measurement, rotAngle float64
	var color string
	for _, g := range codes {
		switch g.code {
		case 8:
			layerName = g.str()
		case 70:
			dimType = g.int() & 0xF
		case 10:
			dlX = g.float()
		case 20:
			dlY = -g.float()
		case 11:
			_ = g.float()
		case 21:
			_ = g.float()
		case 13:
			x1 = g.float()
		case 23:
			y1 = -g.float()
		case 14:
			x2 = g.float()
		case 24:
			y2 = -g.float()
		case 15:
			cx = g.float()
		case 25:
			cy = -g.float()
		case 42:
			measurement = g.float()
		case 50:
			rotAngle = g.float()
		case 62:
			aci := g.int()
			if aci != 0 {
				color = aciToRGB(absInt(aci))
			}
		}
	}
	_ = measurement
	_ = rotAngle
	layerID := p.layerID(layerName)
	if color == "" {
		color = p.layerColor(layerName)
	}
	var e document.Entity
	switch dimType {
	case 0: // rotated / linear
		e = document.Entity{Type: document.TypeDimLinear, X1: x1, Y1: y1, X2: x2, Y2: y2, CX: dlY - (y1+y2)/2, Layer: layerID, Color: color}
	case 1: // aligned
		offset := 0.0
		dist := math.Hypot(x2-x1, y2-y1)
		if dist > 1e-12 {
			ux, uy := -(y2-y1)/dist, (x2-x1)/dist
			offset = (dlX-x1)*ux + (dlY-y1)*uy
		}
		e = document.Entity{Type: document.TypeDimAligned, X1: x1, Y1: y1, X2: x2, Y2: y2, CX: offset, Layer: layerID, Color: color}
	case 2: // angular
		r := math.Hypot(dlX-cx, dlY-cy)
		e = document.Entity{Type: document.TypeDimAngular, CX: cx, CY: cy, X1: x1, Y1: y1, X2: x2, Y2: y2, R: r, Layer: layerID, Color: color}
	case 4: // radial
		ang := math.Atan2(cy-dlY, cx-dlX) * 180 / math.Pi
		r := math.Hypot(x1-cx, y1-cy)
		e = document.Entity{Type: document.TypeDimRadial, CX: cx, CY: cy, R: r, RotDeg: ang, Layer: layerID, Color: color}
	case 3: // diameter
		ang := math.Atan2(cy-dlY, cx-dlX) * 180 / math.Pi
		r := math.Hypot(x1-cx, y1-cy)
		e = document.Entity{Type: document.TypeDimDiameter, CX: cx, CY: cy, R: r, RotDeg: ang, Layer: layerID, Color: color}
	default:
		p.warn("unsupported DIMENSION type %d", dimType)
		return document.Entity{}, false
	}
	return e, true
}

// parseInsertAndExpand expands an INSERT block reference inline.
func (p *parser) parseInsertAndExpand() {
	codes := p.collectEntityCodes()
	blockName := ""
	var ix, iy, sx, sy, rot float64
	sx, sy = 1, 1
	for _, g := range codes {
		switch g.code {
		case 2:
			blockName = g.str()
		case 10:
			ix = g.float()
		case 20:
			iy = -g.float()
		case 41:
			sx = g.float()
		case 42:
			sy = g.float()
		case 50:
			rot = g.float() * math.Pi / 180
		}
	}
	ents, ok := p.blocks[blockName]
	if !ok {
		p.warn("INSERT: block %q not found", blockName)
		return
	}
	cosR, sinR := math.Cos(rot), math.Sin(rot)
	transform := func(x, y float64) (float64, float64) {
		tx := x * sx
		ty := y * sy
		rx := tx*cosR - ty*sinR + ix
		ry := tx*sinR + ty*cosR + iy
		return rx, ry
	}
	for _, e := range ents {
		ne := e
		switch e.Type {
		case document.TypeLine:
			ne.X1, ne.Y1 = transform(e.X1, e.Y1)
			ne.X2, ne.Y2 = transform(e.X2, e.Y2)
		case document.TypeCircle, document.TypeArc:
			ne.CX, ne.CY = transform(e.CX, e.CY)
			ne.R = e.R * math.Sqrt(math.Abs(sx*sy))
		case document.TypeText, document.TypeMText:
			ne.X1, ne.Y1 = transform(e.X1, e.Y1)
			ne.RotDeg = e.RotDeg + rot*180/math.Pi
		case document.TypePolyline, document.TypeSpline, document.TypeNURBS:
			pts := make([][]float64, len(e.Points))
			for i, pt := range e.Points {
				pts[i] = make([]float64, len(pt))
				copy(pts[i], pt)
				if len(pt) >= 2 {
					pts[i][0], pts[i][1] = transform(pt[0], pt[1])
				}
			}
			ne.Points = pts
		}
		p.doc.AddEntity(ne)
	}
}

// ─── Layer application ────────────────────────────────────────────────────────

// applyLayers creates document layers matching what was parsed from the TABLES section.
func (p *parser) applyLayers() {
	for name, info := range p.layers {
		if name == "0" {
			continue
		}
		id := p.doc.AddLayer(name, info.color, info.lineTyp, 0.25)
		p.layerIDs[name] = id
		if !info.visible {
			p.doc.SetLayerVisible(id, false)
		}
		if info.locked {
			p.doc.SetLayerLocked(id, true)
		}
	}
}

// layerID returns the document layer ID for a DXF layer name.
// Unknown layers are created on demand.
func (p *parser) layerID(name string) int {
	if id, ok := p.layerIDs[name]; ok {
		return id
	}
	// Layer referenced in entities but not in TABLES — create it now.
	id := p.doc.AddLayer(name, "#ffffff", document.LineTypeSolid, 0.25)
	p.layerIDs[name] = id
	return id
}

// layerColor returns the color for a DXF layer name (or white if unknown).
func (p *parser) layerColor(name string) string {
	if info, ok := p.layers[name]; ok {
		return info.color
	}
	return "#ffffff"
}

// ─── Public API ───────────────────────────────────────────────────────────────

// Read parses a DXF R12 (AC1009) or R2000 (AC1015) text stream and returns a
// new Document. Malformed or unsupported entities are skipped with warnings
// logged to Warnings. The returned document is always non-nil.
//
// Supported entity types: LINE, CIRCLE, ARC, LWPOLYLINE, POLYLINE+VERTEX+SEQEND,
// SPLINE, ELLIPSE, TEXT, MTEXT, INSERT (block expansion), DIMENSION.
// Supported table data: LAYER (color, linetype, visible, locked).
func Read(r io.Reader) (*document.Document, []string, error) {
	gcs, err := readGCs(r)
	if err != nil {
		return nil, nil, fmt.Errorf("dxf.Read: %w", err)
	}
	p := newParser(gcs)
	p.parse()
	return p.doc, p.warnings, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func normalizeAngle(deg float64) float64 {
	for deg < 0 {
		deg += 360
	}
	for deg >= 360 {
		deg -= 360
	}
	return deg
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// dxfLTNameToLineType maps a DXF LTYPE name to our LineType enum.
func dxfLTNameToLineType(name string) document.LineType {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "DASHED", "DASHED2", "DASHEDX2":
		return document.LineTypeDashed
	case "DOT", "DOTTED", "DOT2", "DOTX2":
		return document.LineTypeDotted
	case "DASHDOT", "DASHDOT2", "DASHDOTX2":
		return document.LineTypeDashDot
	case "CENTER", "CENTER2", "CENTERX2":
		return document.LineTypeCenter
	case "HIDDEN", "HIDDEN2", "HIDDENX2":
		return document.LineTypeHidden
	default:
		return document.LineTypeSolid
	}
}

// aciToRGB maps an AutoCAD Color Index (ACI) to an RGB hex string.
// The table covers the 255 standard ACI colors; unknown indices map to white.
func aciToRGB(aci int) string {
	// Standard ACI 1-9 (exact)
	table := map[int]string{
		1: "#FF0000", 2: "#FFFF00", 3: "#00FF00",
		4: "#00FFFF", 5: "#0000FF", 6: "#FF00FF",
		7: "#FFFFFF", 8: "#414141", 9: "#808080",
	}
	if c, ok := table[aci]; ok {
		return c
	}
	// ACI 10–249 follow a hue/saturation/value pattern organized in groups of 10.
	if aci >= 10 && aci <= 249 {
		group := (aci - 10) / 10
		sub := (aci - 10) % 10
		// 24 hue groups (0° to 350° in 15° steps)
		hue := float64(group%24) * 15.0
		// sub 0-4: full sat, decreasing brightness; sub 5-9: decreasing sat, full bright
		var sat, val float64
		if sub < 5 {
			sat = 1.0
			val = 1.0 - float64(sub)*0.2
		} else {
			sat = 1.0 - float64(sub-5)*0.2
			val = 1.0
		}
		r, g, b := hsvToRGB(hue, sat, val)
		return fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}
	return "#FFFFFF"
}

func hsvToRGB(h, s, v float64) (uint8, uint8, uint8) {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return uint8((r+m)*255 + 0.5), uint8((g+m)*255 + 0.5), uint8((b+m)*255 + 0.5)
}
