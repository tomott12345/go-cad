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
| `internal/document` | Core CAD document model (Entity, undo/redo, DXF R2000/R12 export with layer tables) + full layer system (Layer struct, AddLayer, RemoveLayer, RenameLayer, SetLayerColor/Visible/Locked/Frozen) + geometry bridge shim |
| `internal/snap` | Object-snap engine: FindSnap evaluates all entity types for Endpoint, Midpoint, Center, Quadrant, Intersection, Perpendicular, Tangent, Nearest with priority ordering and bitmask control |
| `cmd/serve` | Dev HTTP server for WASM builds; binds to 127.0.0.1:8080 by default |
| `cmd/wasm` | WASM bridge: snap (cadFindSnap) + full layer CRUD (cadGetLayers, cadAddLayer, cadRemoveLayer, cadSetLayerName/Color/Visible/Locked/Frozen, cadGetCurrentLayer, cadSetCurrentLayer) |

### Key commands
- `cd go-cad && go test ./...` — run all tests (geometry, constraints, document, snap)
- `cd go-cad && go build ./cmd/serve` — build the dev server
- WASM build: `GOOS=js GOARCH=wasm go build -o web/main.wasm ./cmd/wasm`

### Task #5 features (Object Snap + Full Layer System)
- **Snap engine** (`internal/snap`): 8 snap types with priority ordering; `FindSnap` package-level function callable from WASM bridge
- **Layer system** (`internal/document/layers.go`): Full Layer struct (ID, Name, Color, LineTyp, LineWeight, Visible, Locked, Frozen, PrintEnabled); default layer 0 protected from deletion; Save/Load persists all layer state
- **DXF export**: All DXF helpers take `layer string`; DXF layer table written (`writeDXFLayerTable`); diameter dim lines fully fixed
- **WASM bridge**: `cadFindSnap(x,y,radius,mask)` → JSON snap result; complete layer CRUD API
- **Browser UI**: Snap marker SVG overlay with type-specific symbols and colors; F3 key toggle; Layer Manager modal with live editing (name, color, visibility, lock, freeze); frozen/invisible layers hidden from canvas; layer dropdown synced with document state
