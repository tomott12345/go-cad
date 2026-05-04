# Workspace

## Overview

pnpm workspace monorepo using TypeScript. Each package manages its own dependencies.

## Stack

- **Monorepo tool**: pnpm workspaces
- **Node.js version**: 24
- **Package manager**: pnpm
- **TypeScript version**: 5.9
- **API framework**: Express 5
- **Database**: PostgreSQL + Drizzle ORM
- **Validation**: Zod (`zod/v4`), `drizzle-zod`
- **API codegen**: Orval (from OpenAPI spec)
- **Build**: esbuild (CJS bundle)

## Key Commands

- `pnpm run typecheck` — full typecheck across all packages
- `pnpm run build` — typecheck + build all packages
- `pnpm --filter @workspace/api-spec run codegen` — regenerate API hooks and Zod schemas from OpenAPI spec
- `pnpm --filter @workspace/db run push` — push DB schema changes (dev only)
- `pnpm --filter @workspace/api-server run dev` — run API server locally

See the `pnpm-workspace` skill for workspace structure, TypeScript setup, and package details.

## go-cad (Go CAD application)

A standalone Go module living under `go-cad/` — a modular, open-source CAD engine targeting both browser (WASM) and desktop (Fyne). Module path: `go-cad`, requires Go 1.22.

### Packages
| Package | Purpose |
|---|---|
| `internal/geometry` | 2-D primitives: Point, BBox, Segment/Line, Circle, Arc, Ellipse, Polyline, Bezier/NURBS splines, Entity interface, full intersection engine |
| `internal/constraints` | Parametric constraint solver (Coincident, Horizontal, Vertical, Parallel, Perpendicular, EqualLength, Fixed, Midpoint, Tangent, Symmetric) using iterative Gauss-Seidel with pin enforcement |
| `internal/document` | Core CAD document model (Entity, undo/redo, DXF R2000/R12 export with layer tables) + full layer system + DXF import via `LoadDXFBytes` + `RegisterDXFReader` dependency injection to avoid import cycles |
| `internal/snap` | Object-snap engine: FindSnap evaluates all entity types for Endpoint, Midpoint, Center, Quadrant, Intersection, Perpendicular, Tangent, Nearest with priority ordering and bitmask control |
| `pkg/dxf` | DXF import/export public surface: `Read(io.Reader)` parses R12/R2000 (LINE, CIRCLE, ARC, LWPOLYLINE, POLYLINE+VERTEX, SPLINE→NURBS, ELLIPSE, TEXT, MTEXT, INSERT block expansion, DIMENSION); ACI→RGB color mapping; `Write`/`WriteR12`/`String`/`StringR12`/`ReadString` convenience wrappers |
| `pkg/svg` | SVG exporter: `Generate(doc)` produces a standards-compliant SVG with layers as `<g>` elements, stroke-dasharray for linetypes, viewBox from entity bounding box, XML-escaped text content |
| `cmd/serve` | Dev HTTP server for WASM builds; binds to 127.0.0.1:8080 by default |
| `cmd/wasm` | WASM bridge: all entity types, snap, layers, `cadLoadDXF(str)→JSON`, `cadExportSVG()→string`, `cadExportDXF/DXFRaw12` |
| `cmd/cad` | Terminal REPL: draw entities, manage layers, snap, IMPORT/EXPORT DXF, EXPORTSVG |

### Key commands
- `cd go-cad && go test ./...` — run all tests (all 10 packages pass)
- `cd go-cad && go build ./cmd/serve` — build the dev server
- WASM build: `GOOS=js GOARCH=wasm go build -buildvcs=false -o web/main.wasm ./cmd/wasm`

### Task #5 features (Object Snap + Full Layer System)
- **Snap engine** (`internal/snap`): 8 snap types with priority ordering; `FindSnap` callable from WASM bridge
- **Layer system** (`internal/document/layers.go`): Full Layer struct; default layer 0 protected; Save/Load persists state
- **DXF export**: Layer table + full DIMENSION entities in R2000; R12 approximations

### Task #6 features (DXF Import + SVG Export)
- **`pkg/dxf`**: Streaming group-code parser; imports LINE, CIRCLE, ARC, LWPOLYLINE, POLYLINE+VERTEX, SPLINE→NURBS, ELLIPSE, TEXT, MTEXT, INSERT (block expansion), DIMENSION; LAYER table (color, linetype, visible, locked); ACI→RGB; round-trip tests + smoke tests
- **`pkg/svg`**: SVG exporter with per-layer `<g>` elements, viewBox auto-computed, stroke-dasharray linetypes, XML-escaped text
- **`document.LoadDXFBytes`**: DXF import via dependency-injection (`RegisterDXFReader`) to avoid circular imports; previous state pushed to undo stack
- **WASM bridge**: `cadLoadDXF(str)→JSON{ok,count,warnings}`, `cadExportSVG()→string`
- **Browser UI**: "Open…" button + hidden `<input type="file" accept=".dxf">`, Ctrl+O shortcut, SVG option in export format selector, file import with zoom-fit and status display
- **cmd/cad REPL**: `IMPORT file.dxf` and `EXPORTSVG file.svg` commands
- **WASM bridge**: `cadFindSnap(x,y,radius,mask)` → JSON snap result; complete layer CRUD API
- **Browser UI**: Snap marker SVG overlay with type-specific symbols and colors; F3 key toggle; Layer Manager modal with live editing (name, color, visibility, lock, freeze); frozen/invisible layers hidden from canvas; layer dropdown synced with document state
