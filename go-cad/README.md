# go-cad

[![CI](https://github.com/tomott12345/go-cad/actions/workflows/ci.yml/badge.svg)](https://github.com/tomott12345/go-cad/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/tomott12345/go-cad/branch/main/graph/badge.svg)](https://codecov.io/gh/tomott12345/go-cad)
[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org/dl/)
[![Release](https://img.shields.io/github/v/release/tomott12345/go-cad)](https://github.com/tomott12345/go-cad/releases/latest)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Live Demo](https://img.shields.io/badge/demo-GitHub%20Pages-orange.svg)](https://tomott12345.github.io/go-cad)

A robust, open-source CAD application written in Go — zero external dependencies, zero npm, runs natively and in the browser via WebAssembly.

---

## Try it

**[Live browser demo →](https://tomott12345.github.io/go-cad)**

Or run locally in 30 seconds:

```sh
git clone https://github.com/tomott12345/go-cad
cd go-cad
make serve        # builds WASM + starts dev server on :8080
```

---

## Features

- **Full 2D entity set** — line, circle, arc, ellipse, rectangle, polyline, spline (cubic Bézier), NURBS, text, mtext, dimensions (linear/aligned/angular/radial/diameter), hatch, leader, revision cloud, wipeout, block inserts
- **Layer system** — color, line type, line weight, visibility, lock, freeze, print flags
- **DXF I/O** — import and export R12 and R2000 files; compatible with AutoCAD, QCAD, LibreCAD
- **SVG export** — layer-aware, dash-pattern line types, text rendering
- **Print/Plot** — PNG (DPI-scaled), SVG (cropped), DXF output; 1:1 through 1:100 scale presets
- **Object snap** — endpoint, midpoint, center, quadrant, intersection, perpendicular, tangent, nearest
- **Parametric constraints** — coincident, horizontal, vertical, parallel, perpendicular, equal length
- **Plugin API** — extend go-cad with `.so` plugins using a stable versioned SDK
- **Browser WASM frontend** — zero npm, native ES modules, full keyboard shortcut set
- **REST API** — entity CRUD, layer management, DXF/SVG export, plugin management
- **Terminal REPL** — full document model access without a browser

## Architecture

go-cad compiles to two targets from the same source:

```
go-cad/
├── cmd/cad/        Terminal REPL (native binary)
├── cmd/serve/      HTTP dev server (REST API + static files)
├── cmd/wasm/       Browser WebAssembly entry point
├── internal/
│   ├── document/   Core document model (shared by all targets)
│   ├── geometry/   2D geometry primitives and intersection engine
│   ├── snap/       Object-snap engine
│   ├── constraints/ Parametric constraint solver
│   ├── hatch/      Scanline polygon fill
│   ├── pluginhost/ Plugin loader and event dispatch
│   └── symbols/    Built-in symbol library
├── pkg/
│   ├── dxf/        DXF R12/R2000 reader and writer
│   ├── svg/        SVG exporter
│   └── plugin/     Public plugin SDK
└── web/            Browser frontend (HTML + ES modules + WASM)
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed diagrams.

## Quick start

### Prerequisites

- Go 1.25 or later (`go version`)
- No other dependencies required

### Build and run

```sh
# Clone
git clone https://github.com/tomott12345/go-cad && cd go-cad

# Run all tests
make test

# Build desktop terminal REPL
make build
./cad

# Build + run browser dev server (PORT=8080)
make serve

# Build WASM only
make wasm
```

### Using the Makefile

| Command | Description |
|---------|-------------|
| `make build` | Desktop binary for the current platform |
| `make wasm` | WASM frontend → `web/main.wasm` |
| `make serve` | Dev server on `$PORT` (default 8080) |
| `make test` | All Go tests |
| `make vet` | `go vet ./...` |
| `make lint` | `golangci-lint ./...` |
| `make release` | Cross-compile for Linux/macOS/Windows |
| `make clean` | Remove build artifacts |

## Plugin development

See [PLUGIN_SDK.md](PLUGIN_SDK.md) for the full plugin guide.

Quick scaffold:

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

func NewPlugin() plugin.Plugin { return &HelloPlugin{} }
```

Build as a shared library and drop it in the `plugins/` directory:

```sh
go build -buildmode=plugin -o plugins/hello.so .
```

## REST API

The dev server exposes a REST API at `/api/v1/`. See [`api/openapi.yaml`](api/openapi.yaml) for the full spec.

Key endpoints:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/entities` | List all entities |
| `POST` | `/api/v1/entities` | Add an entity |
| `PATCH` | `/api/v1/entities/{id}` | Update entity properties |
| `DELETE` | `/api/v1/entities/{id}` | Delete an entity |
| `GET` | `/api/v1/layers` | List layers |
| `POST` | `/api/v1/export/dxf` | Export DXF R2000 |
| `POST` | `/api/v1/export/svg` | Export SVG |
| `POST` | `/api/v1/undo` | Undo last operation |
| `POST` | `/api/v1/redo` | Redo last undone operation |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide.

## License

MIT — see [LICENSE](LICENSE).
