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
| `internal/document` | Core CAD document model (Entity, undo/redo, DXF export) + geometry bridge shim (BoundingBox, ClosestPoint, Offset, IntersectWith) |
| `cmd/serve` | Dev HTTP server for WASM builds; binds to 127.0.0.1:8080 by default (security fix) |

### Key commands
- `cd go-cad && go test ./...` — run all tests (geometry, constraints, document)
- `cd go-cad && go build ./cmd/serve` — build the dev server
- WASM build: `GOOS=js GOARCH=wasm go build -o web/main.wasm ./cmd/wasm`
