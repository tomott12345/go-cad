// blocks.go — Task #7: Block definitions, insertion, and explosion.
//
// A Block is a named, reusable symbol defined by a base-point and a set of
// entities stored in block-local coordinates (shifted so the base-point is
// the origin).  InsertBlock adds a lightweight TypeBlockRef proxy that records
// the name, insertion point, scale, and rotation.  ExplodeBlock replaces the
// proxy with fully-transformed copies of the constituent entities.
package document

import "math"

// ─── Block struct ─────────────────────────────────────────────────────────────

// Block is a named reusable symbol.
type Block struct {
        Name     string   `json:"name"`
        BaseX    float64  `json:"baseX"`
        BaseY    float64  `json:"baseY"`
        Entities []Entity `json:"entities"` // block-local coords (base = origin)
}

// copyBlock returns a deep copy of a Block.
func copyBlock(b *Block) *Block {
        ents := make([]Entity, len(b.Entities))
        for i, e := range b.Entities {
                ents[i] = deepCopyEntity(e)
        }
        return &Block{Name: b.Name, BaseX: b.BaseX, BaseY: b.BaseY, Entities: ents}
}

// copyBlocks returns a deep copy of the blocks map.
func copyBlocks(src map[string]*Block) map[string]*Block {
        if src == nil {
                return nil
        }
        dst := make(map[string]*Block, len(src))
        for k, v := range src {
                dst[k] = copyBlock(v)
        }
        return dst
}

// ─── Document block accessors ─────────────────────────────────────────────────

// Blocks returns a copy of all defined block definitions.
func (d *Document) Blocks() []*Block {
        out := make([]*Block, 0, len(d.blocks))
        for _, b := range d.blocks {
                out = append(out, copyBlock(b))
        }
        return out
}

// BlockByName returns a copy of the named block, or nil if not defined.
func (d *Document) BlockByName(name string) *Block {
        b := d.blocks[name]
        if b == nil {
                return nil
        }
        return copyBlock(b)
}

// ─── DefineBlock ──────────────────────────────────────────────────────────────

// DefineBlock creates or replaces a named block definition.
//
// entityIDs: IDs of existing document entities to include (they are copied into
// the block in block-local coordinates: shifted so that baseX,baseY is the
// local origin).  The source entities are NOT removed from the document.
//
// Returns true on success, false if name is empty or no valid IDs were found.
func (d *Document) DefineBlock(name string, baseX, baseY float64, entityIDs []int) bool {
        if name == "" || len(entityIDs) == 0 {
                return false
        }
        idSet := make(map[int]bool, len(entityIDs))
        for _, id := range entityIDs {
                idSet[id] = true
        }
        var ents []Entity
        for _, e := range d.entities {
                if !idSet[e.ID] {
                        continue
                }
                local := deepCopyEntity(e)
                shiftXfm := func(x, y float64) (float64, float64) { return x - baseX, y - baseY }
                local = applyXfm(local, shiftXfm, 1.0)
                local.ID = 0
                ents = append(ents, local)
        }
        if len(ents) == 0 {
                return false
        }
        d.pushUndo()
        if d.blocks == nil {
                d.blocks = make(map[string]*Block)
        }
        d.blocks[name] = &Block{Name: name, BaseX: baseX, BaseY: baseY, Entities: ents}
        return true
}

// DefineBlockRaw stores raw block-local entities directly (used by symbol
// library and DXF import to inject pre-computed block definitions without
// touching the undo stack).
func (d *Document) DefineBlockRaw(name string, baseX, baseY float64, ents []Entity) {
        if d.blocks == nil {
                d.blocks = make(map[string]*Block)
        }
        cp := make([]Entity, len(ents))
        for i, e := range ents {
                cp[i] = deepCopyEntity(e)
        }
        d.blocks[name] = &Block{Name: name, BaseX: baseX, BaseY: baseY, Entities: cp}
}

// ─── InsertBlock ──────────────────────────────────────────────────────────────

// InsertBlock adds a block-reference entity (TypeBlockRef) to the document.
//
// x, y: insertion point (world coords); scaleX, scaleY: non-uniform scale
// (defaults to 1 if 0); rotDeg: CCW rotation in degrees.
// Returns the entity ID, or -1 if the block name is empty.
func (d *Document) InsertBlock(name string, x, y, scaleX, scaleY, rotDeg float64, layer int, color string) int {
        if name == "" {
                return -1
        }
        // Reject references to undefined blocks so INSERT entities always resolve.
        if d.blocks == nil || d.blocks[name] == nil {
                return -1
        }
        if scaleX == 0 {
                scaleX = 1
        }
        if scaleY == 0 {
                scaleY = 1
        }
        return d.add(Entity{
                Type: TypeBlockRef, Text: name,
                X1: x, Y1: y,
                R: scaleX, R2: scaleY,
                RotDeg: rotDeg,
                Layer:  layer, Color: color,
        })
}

// ─── ExplodeBlock ─────────────────────────────────────────────────────────────

// ExplodeBlock replaces the TypeBlockRef entity with ID id by the constituent
// block entities, fully transformed (scaled + rotated + translated).
//
// Returns the IDs of the newly created entities, or nil if id does not refer
// to a block reference or the referenced block is not defined.
func (d *Document) ExplodeBlock(id int) []int {
        var refIdx int = -1
        for i, e := range d.entities {
                if e.ID == id && e.Type == TypeBlockRef {
                        refIdx = i
                        break
                }
        }
        if refIdx < 0 {
                return nil
        }
        ref := d.entities[refIdx]
        blk := d.blocks[ref.Text]
        if blk == nil {
                return nil
        }

        d.pushUndo()

        sx, sy := ref.R, ref.R2
        if sx == 0 {
                sx = 1
        }
        if sy == 0 {
                sy = 1
        }
        rotRad := ref.RotDeg * math.Pi / 180
        cosR, sinR := math.Cos(rotRad), math.Sin(rotRad)
        tx, ty := ref.X1, ref.Y1
        radScale := math.Abs(sx)

        xfm := func(x, y float64) (float64, float64) {
                lx, ly := x*sx, y*sy
                return lx*cosR - ly*sinR + tx, lx*sinR + ly*cosR + ty
        }

        // Remove the reference entity.
        d.entities = append(d.entities[:refIdx], d.entities[refIdx+1:]...)

        var newIDs []int
        for _, be := range blk.Entities {
                local := deepCopyEntity(be)
                local = applyXfm(local, xfm, radScale)
                local.ID = d.nextID
                d.nextID++
                if local.Color == "" {
                        local.Color = ref.Color
                }
                local.Layer = ref.Layer
                d.entities = append(d.entities, local)
                newIDs = append(newIDs, local.ID)
        }
        return newIDs
}
