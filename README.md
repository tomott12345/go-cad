# go-cad

[![CI](https://github.com/tomott12345/go-cad/actions/workflows/ci.yml/badge.svg)](https://github.com/tomott12345/go-cad/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/tomott12345/go-cad/branch/main/graph/badge.svg)](https://codecov.io/gh/tomott12345/go-cad)
[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org/dl/)
[![Release](https://img.shields.io/github/v/release/tomott12345/go-cad)](https://github.com/tomott12345/go-cad/releases/latest)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Live Demo](https://img.shields.io/badge/demo-GitHub%20Pages-orange.svg)](https://tomott12345.github.io/go-cad)

A robust, open-source 2D CAD application written in Go — zero external dependencies, runs natively on your desktop and in any browser via WebAssembly.

---

## Try it now

**[Live browser demo →](https://tomott12345.github.io/go-cad)**

Or run locally:

```sh
git clone https://github.com/tomott12345/go-cad
cd go-cad
make serve        # builds WASM + starts dev server on :8080
```

Open `http://localhost:8080`, click a tool in the toolbar (or press `L` for Line), then click on the canvas to draw.

---

## Features

### Drawing tools

| Category | Tools |
|----------|-------|
| Primitives | Line, Circle, Arc, Rectangle, Ellipse |
| Advanced curves | Polyline, Cubic Bézier Spline, NURBS |
| Annotations | Single-line Text, Multi-line Text (MText) |
| Dimensions | Linear, Aligned, Angular, Radial, Diameter |
| Drafting | Hatch fill, Leader, Revision Cloud, Wipeout |
| Blocks | Define blocks, insert references, explode |

### CAD engine

- **Layer system** — color, linetype (Solid/Dashed/Dotted/DashDot/Center/Hidden), lineweight, visibility, lock, freeze, print flags
- **DXF I/O** — import and export R12 (AC1009) and R2000 (AC1015); compatible with AutoCAD, QCAD, LibreCAD
- **SVG export** — layer-aware, full dash-pattern linetypes, text rendering
- **Print/Plot** — PNG (DPI-scaled), SVG (cropped), or DXF; 1:1 through 1:100 scale presets
- **Object snap** — 8 modes with configurable toggles (F3 to toggle on/off)
- **Parametric constraints** — coincident, horizontal, vertical, parallel, perpendicular, equal length (Gauss-Seidel solver)
- **Undo/Redo** — full document history via Ctrl+Z / Ctrl+Y

### Platform targets

- **Browser** — WebAssembly frontend, zero npm, native ES modules, full keyboard shortcuts
- **REST API** — entity CRUD, layer management, DXF/SVG export; spec at [`api/openapi.yaml`](api/openapi.yaml)
- **Terminal REPL** — full document model access from the command line
- **Plugin API** — extend go-cad with subprocess or `.so` plugins using a stable versioned SDK

---

## Quick start

### Prerequisites

- **Go 1.25** or later — verify with `go version`
- Nothing else. Zero external dependencies.

### Build and run

```sh
# Clone
git clone https://github.com/tomott12345/go-cad && cd go-cad

# Build WASM + start browser dev server (default PORT=8080)
make serve

# Run all tests
make test

# Build desktop terminal REPL
make build && ./cad
```

### Makefile reference

| Command | Description |
|---------|-------------|
| `make build` | Desktop binary + WASM for the current platform |
| `make wasm` | WASM frontend only → `web/main.wasm` |
| `make serve` | Rebuild WASM, then start dev server on `$PORT` (default `8080`) |
| `make test` | All Go tests with race detector |
| `make vet` | `go vet ./...` |
| `make lint` | `golangci-lint ./...` |
| `make release` | Cross-compile for Linux/macOS/Windows (amd64 + arm64) |
| `make clean` | Remove build artifacts |

> **Important:** `make serve` always rebuilds `web/main.wasm` before starting the server. If you change Go source files and restart only the server binary, the browser will run stale WASM. Always use `make serve` or `make wasm` after any Go changes.

---

## Browser usage

### Keyboard shortcuts

| Key | Tool | Key | Tool |
|-----|------|-----|------|
| `L` | Line | `S` | Spline |
| `C` | Circle | `N` | NURBS |
| `A` | Arc | `E` | Ellipse |
| `R` | Rectangle | `T` | Text |
| `P` | Polyline | `Esc` | Select |

| Shortcut | Action |
|----------|--------|
| `Ctrl+Z` / `Ctrl+Y` | Undo / Redo |
| `Delete` | Delete selected entity |
| `Enter` | Finish multi-point tool (Polyline, Spline, Hatch …) |
| `F3` | Toggle Object Snap on/off |
| `Ctrl+O` | Open DXF file |
| `Ctrl+P` | Print / Plot dialog |
| Middle mouse or Ctrl+drag | Pan canvas |
| Scroll wheel | Zoom in/out |

### Command bar

Start typing anywhere — the command bar at the bottom auto-focuses. Press Enter to execute.

**Drawing**

| Command | Alias | | Command | Alias |
|---------|-------|-|---------|-------|
| `LINE` | `L` | | `HATCH` | `H` |
| `CIRCLE` | `C` | | `LEADER` | `LD` |
| `ARC` | `A` | | `REVCLOUD` | `RC` |
| `RECT` | `R` | | `WIPEOUT` | `WP` |
| `POLY` | `P` | | `TEXT` | `T` |
| `SPLINE` | `S` | | `MTEXT` | `MT` |
| `NURBS` | `N` | | `ELLIPSE` | `E` |

**Dimensions:** `DIMLIN` (`DL`), `DIMALI` (`DA`), `DIMANG` (`DANG`), `DIMRAD` (`DR`), `DIMDIA` (`DD`)

**Edit:** `MOVE` (`M`), `COPY` (`CP`), `ROTATE` (`RO`), `SCALE` (`SC`), `MIRROR` (`MI`), `TRIM` (`TR`), `EXTEND` (`EX`), `FILLET` (`F`), `CHAMFER` (`CHA`), `ARRAYRECT` (`AR`), `ARRAYPOLAR` (`AP`), `OFFSET` (`O`)

**Document:** `LAYERS` / `LA`, `BLOCKS`, `SYMBOLS`, `INSERT name x y`, `DEFINEBLOCK name`, `EXPLODE id`, `EXPORT`, `PRINT`, `UNDO`, `REDO`, `CLEAR`, `ZOOMFIT` / `ZF`

**Coordinate input** (while a drawing tool is active)

| Format | Example | Meaning |
|--------|---------|---------|
| `x,y` | `100,50` | Absolute world coordinates |
| `@dx,dy` | `@50,-25` | Relative to last point |
| `@dist<angle` | `@100<45` | Polar relative to last point |

### Object snap modes

| Mode | Color | Snaps to |
|------|-------|----------|
| Endpoint | Red | Start/end of lines and arcs |
| Midpoint | Green | Midpoint of any segment |
| Center | Blue | Circle/arc center point |
| Quadrant | Orange | 0°/90°/180°/270° on circles |
| Intersection | Magenta | Where two entities cross |
| Perpendicular | Cyan | Foot of perpendicular from cursor |
| Tangent | Yellow | Tangent point on circle/arc |
| Nearest | Gray | Closest point on any entity |

---

## Architecture

go-cad compiles three targets from a single Go module:

```
go-cad/
├── cmd/
│   ├── cad/           Terminal REPL (native binary)
│   ├── serve/         HTTP dev server (REST API + static files)
│   └── wasm/          Browser WebAssembly entry point
├── internal/
│   ├── document/      Core document model — entities, layers, undo, DXF I/O
│   ├── geometry/      2D geometry primitives and intersection engine
│   ├── snap/          Object-snap engine (8 snap modes)
│   ├── constraints/   Parametric constraint solver (Gauss-Seidel)
│   ├── hatch/         Scanline polygon fill
│   ├── pluginhost/    Plugin loader and event dispatch
│   └── symbols/       Built-in symbol library
├── pkg/
│   ├── dxf/           DXF R12/R2000 reader and writer
│   ├── svg/           SVG exporter
│   └── plugin/        Public plugin SDK (stable versioned API)
├── web/               Browser frontend — HTML + ES modules + WASM
├── api/
│   └── openapi.yaml   REST API specification
└── examples/
    └── plugins/       Example plugin: dimension-tool
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed component diagrams and data-flow descriptions.

---

## REST API

The dev server exposes a REST API at `/api/v1/`. Full spec: [`api/openapi.yaml`](api/openapi.yaml).

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/entities` | List all entities |
| `POST` | `/api/v1/entities` | Add an entity |
| `PATCH` | `/api/v1/entities/{id}` | Update entity properties |
| `DELETE` | `/api/v1/entities/{id}` | Delete an entity |
| `GET` | `/api/v1/layers` | List layers |
| `POST` | `/api/v1/layers` | Add a layer |
| `POST` | `/api/v1/export/dxf` | Export DXF R2000 |
| `POST` | `/api/v1/export/svg` | Export SVG |
| `POST` | `/api/v1/undo` | Undo last operation |
| `POST` | `/api/v1/redo` | Redo last undone operation |

---

## Plugin development

go-cad supports two plugin deployment modes:

| Mode | Platform | How to build |
|------|----------|-------------|
| Subprocess (recommended) | All platforms | `go build -o myplugin ./` |
| Shared library | Linux/macOS only | `go build -buildmode=plugin -o myplugin.so ./` |

Drop the compiled binary or `.so` in `~/.go-cad/plugins/` — it loads automatically on startup.

### Minimal plugin

```go
package main

import "github.com/tomott12345/go-cad/pkg/plugin"

type HelloPlugin struct{ api plugin.HostAPI }

func (p *HelloPlugin) Name() string    { return "hello" }
func (p *HelloPlugin) Version() string { return "1.0.0" }

func (p *HelloPlugin) Register(api plugin.HostAPI) error {
    p.api = api
    return api.RegisterCommand(plugin.CommandDescriptor{
        Name: "HELLO",
        Handler: func(args []string) error {
            _, err := api.AddEntity(plugin.Entity{
                Type: "line", X1: 0, Y1: 0, X2: 100, Y2: 0,
            })
            return err
        },
    })
}

func (p *HelloPlugin) Unregister() error { return nil }

// NewPlugin is the required symbol looked up by the loader.
func NewPlugin() plugin.Plugin { return &HelloPlugin{} }
```

```sh
go build -o hello ./
mkdir -p ~/.go-cad/plugins && mv hello ~/.go-cad/plugins/
```

A fully-worked example lives in [`examples/plugins/dimension-tool/`](examples/plugins/dimension-tool/).
See [PLUGIN_SDK.md](PLUGIN_SDK.md) for the complete guide including events, entity inspection, and the subprocess protocol.

---

## CI / CD

Every push and pull request runs:

| Job | What it checks |
|-----|---------------|
| **Lint** | `go vet` + `golangci-lint` |
| **Test** | All packages with `-race -covermode=atomic` |
| **Build check** | Every package compiles cleanly |
| **Cross-compile** | Linux/macOS/Windows × amd64/arm64 |
| **WASM** | Browser WASM compiles without errors |
| **E2E** | Playwright tests against a live dev server |

Pull requests automatically receive a sticky coverage comment showing the total percentage and the change versus the base branch.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Short version:

1. `make test` must pass
2. `make vet` must pass
3. No external Go dependencies — stdlib only
4. New packages need tests

---

## License

MIT — see [LICENSE](LICENSE).
