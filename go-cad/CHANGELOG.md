# Changelog

All notable changes to go-cad are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
go-cad adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- CI/CD pipeline: GitHub Actions for lint, test, cross-compile, release, and GitHub Pages
- Makefile with `build`, `wasm`, `serve`, `test`, `vet`, `lint`, `release`, and `clean` targets
- CONTRIBUTING.md and ARCHITECTURE.md for new contributors
- golangci-lint configuration (`.golangci.yml`)
- README badges for CI status, coverage, Go version, and license

---

## [0.1.0] — 2024-05-04

### Added

#### Core geometry engine (`internal/geometry`)
- 2D primitives: segment, infinite line, ray, circle, arc, ellipse, polyline, cubic Bézier spline, NURBS
- Axis-aligned bounding box (`BBox`) with union, contains, and expand operations
- Intersection engine: segment×segment, segment×circle, segment×arc, segment×ellipse, circle×circle
- Closest-point and distance queries on all entity types

#### CAD document model (`internal/document`)
- Entity types: line, circle, arc, rectangle, polyline, spline, NURBS, ellipse, text, mtext
- Dimension types: linear, aligned, angular, radial, diameter
- Block types: block reference (insert), hatch, leader, revision cloud, wipeout
- Layer system with color, line type, line weight, visibility, lock, freeze, and print flags
- Undo/redo history (unlimited steps)
- JSON serialization for full document round-trip
- DXF R2000 and DXF R12 export
- DXF R12/R2000 import
- SVG export (layer-aware, line-type dash patterns, text rendering)
- Entity property setters wired to WASM bridge (`cadSetEntityProp`)

#### Snap engine (`internal/snap`)
- Eight snap modes: endpoint, midpoint, center, quadrant, intersection, perpendicular, tangent, nearest
- Per-mode enable/disable toggle
- Configurable snap radius (world units)
- Batch snap-candidates query for canvas rendering

#### Constraint solver (`internal/constraints`)
- Geometric constraints: coincident, horizontal, vertical, parallel, perpendicular, equal length, fixed point
- Iterative Gauss-Seidel solver with configurable tolerance and iteration limit
- Entity-level solver shim (apply constraints to Document entities)

#### Plugin API (`pkg/plugin`, `internal/pluginhost`)
- Stable versioned plugin SDK (`PluginAPIVersion = "1.0.0"`)
- Plugin interface: `Name`, `Version`, `Register`, `Unregister`
- `HostAPI`: `AddEntity`, `GetEntity`, `UpdateEntity`, `DeleteEntity`, `RegisterCommand`, `Subscribe`
- Event system: `entity.added`, `entity.deleted`, `selection.changed`, `document.saved`, `document.loaded`, `tool.changed`
- Dynamic plugin loader (`pkg/plugin/loader`) — load `.so` shared libraries at runtime
- Plugin management REST endpoints: list, load, unload, execute command

#### DXF I/O (`pkg/dxf`)
- Reader: SECTIONS parser for HEADER, TABLES (layer, ltype), BLOCKS, ENTITIES
- Supports POINT, LINE, CIRCLE, ARC, ELLIPSE, LWPOLYLINE, SPLINE, MTEXT, TEXT, INSERT, HATCH, LEADER, REVCLOUD, WIPEOUT
- Writer: DXF R2000 (AC1015) and DXF R12 (AC1009) with full layer and line-type tables

#### SVG export (`pkg/svg`)
- Layer-to-`<g>` mapping with `id="layer-N"` and `data-name`
- Line type to `stroke-dasharray` mapping (solid, dashed, dotted, dash-dot, center, hidden)
- Text and mtext rendering as `<text>` elements
- Auto-computed `viewBox` from entity bounding boxes

#### Browser WASM frontend (`cmd/wasm`)
- Zero-npm, native ES module architecture
- JavaScript API surface: 60+ `cad*` functions exported to `globalThis`
- Panelized canvas layout: left tool panel, right inspector, bottom command history, center viewport
- Panel splitters with localStorage persistence
- Properties inspector with live entity edits
- Command history panel with click-to-replay
- Snap toolbar (F3 toggle, per-mode checkboxes, popover UI)
- Absolute/relative/polar coordinate input parser
- Drafting settings dialog (snap distance, grid, angle increment)
- Print/Plot dialog: PNG (DPI-scaled), SVG (cropped viewBox), DXF export; scale presets 1:1 through 1:100
- Welcome overlay with recent-files list and documentation link
- SELECT tool as default; Escape returns to select
- Full keyboard shortcut set (L, C, A, R, P, S, N, E, T, MT, M, CP, RO, SC, MI, TR, EX, F, CHA, AR, AP, O, H, LD, RC, WP, ZF, Ctrl+Z, Ctrl+Y)

#### REST API server (`cmd/serve`)
- Endpoints: entity CRUD, layer CRUD, undo/redo, export (DXF/SVG), command execution, plugin management
- OpenAPI 3.0 spec (`api/openapi.yaml`)
- Static file server for `web/` directory

#### Desktop terminal interface (`cmd/cad`)
- Interactive REPL with 40+ commands
- Full document model access, snap control, DXF import/export

#### Built-in symbol library (`internal/symbols`)
- CENTER_MARK, NORTH_ARROW, REVISION_TRIANGLE, ELECTRICAL_GROUND, WELD_BUTT, WELD_FILLET, SURFACE_ROUGHNESS, DATUM_TRIANGLE

#### Hatch engine (`internal/hatch`)
- Scanline polygon fill for ANSI31, ANSI32, ANSI33, ANSI34, ANSI35, ANSI36, ANSI37, SOLID patterns

---

[Unreleased]: https://github.com/tomott12345/go-cad/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/tomott12345/go-cad/releases/tag/v0.1.0
