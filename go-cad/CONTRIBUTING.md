# Contributing to go-cad

Thank you for your interest in contributing! go-cad is a zero-dependency Go project, so there is nothing to install beyond Go itself.

---

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.22+ | https://golang.org/dl/ |
| git | any | For cloning and branching |
| golangci-lint | latest | Optional — for lint checks |
| make | any | Optional — convenience targets |

go-cad has **no other build requirements**: no Node.js, no npm, no Fyne SDK, no CGO (unless you are building the optional native Fyne desktop UI). The WASM target and browser frontend are built with stock Go tooling.

---

## Clone and build

```sh
git clone https://github.com/tomott12345/go-cad
cd go-cad
```

Build the desktop terminal REPL:

```sh
make build        # or: go build -o cad ./cmd/cad/
./cad
```

Build and run the browser dev server:

```sh
make serve        # or: GOOS=js GOARCH=wasm go build -buildvcs=false -o web/main.wasm ./cmd/wasm/
                  #     PORT=8080 go run ./cmd/serve/ -host 0.0.0.0
```

Run all tests:

```sh
make test         # or: go test ./...
```

---

## Repository layout

```
go-cad/
├── cmd/
│   ├── cad/        Terminal REPL (native binary, no CGO)
│   ├── serve/      HTTP API server + static file server
│   └── wasm/       Browser WebAssembly entry point (GOOS=js)
├── internal/
│   ├── constraints/ Parametric constraint solver
│   ├── document/   Core document model (all entity types, layers, DXF I/O)
│   ├── geometry/   2D geometry primitives and intersection engine
│   ├── hatch/      Scanline polygon fill engine
│   ├── pluginhost/ Plugin loader, event bus, command router
│   ├── snap/       Object-snap engine (8 modes)
│   └── symbols/    Built-in symbol block library
├── pkg/
│   ├── dxf/        Public DXF reader and writer (R12 + R2000)
│   ├── plugin/     Public plugin SDK (stable API)
│   └── svg/        SVG exporter
├── web/            Browser frontend (HTML, ES modules, WASM binary)
├── api/            OpenAPI 3.0 specification
├── examples/       Example plugins
└── Makefile
```

---

## Running tests

```sh
go test ./...                        # all packages
go test ./internal/geometry/...      # single package
go test -run TestSnap ./internal/snap/  # single test
go test -race ./...                  # race detector (recommended before PRs)
go test -cover ./...                 # coverage summary
```

Tests live alongside the code they cover (`_test.go` suffix). go-cad does **not** use any test framework beyond the standard library.

---

## Code style

go-cad follows standard Go conventions with a few specifics:

1. **`gofmt`** — all code must be `gofmt`-clean. Run `gofmt -l .` to check.
2. **`go vet`** — must pass with no warnings (`make vet`).
3. **`golangci-lint`** — lint passes with `.golangci.yml` config (`make lint`). Key linters: `errcheck`, `govet`, `staticcheck`, `unused`, `revive`.
4. **Comments** — every exported package, type, function, and constant must have a GoDoc comment. Use full sentences starting with the symbol name.
5. **No external dependencies** — go-cad is stdlib-only. Do not add entries to `go.mod`. The plugin SDK (`pkg/plugin`) may be imported by third-party plugins, but go-cad itself imports nothing outside the standard library.
6. **Error handling** — do not ignore errors. Prefer explicit `if err != nil` checks. Do not use `_` to discard errors from significant operations.
7. **Package naming** — short, lowercase, no underscores. Internal packages live under `internal/`; stable public API packages live under `pkg/`.

### Commit messages

Use a short imperative subject line (≤72 chars) followed by a blank line and a longer body if needed:

```
Add perpendicular snap mode to snap engine

Implements the perpendicular object-snap mode as described in issue #42.
Adds snap_perp.go with a new SnapPerpendicular function and a test in
snap_test.go covering segments, circles, and arcs.
```

---

## Submitting a pull request

1. **Fork** the repository and create a feature branch:
   ```sh
   git checkout -b feature/my-feature
   ```

2. **Write tests** for new behavior. go-cad aims for >80% coverage on `internal/` packages.

3. **Ensure CI passes locally**:
   ```sh
   make vet test
   ```

4. **Open a pull request** against `main`. Include:
   - A clear description of the change and motivation
   - A link to any related issue
   - Screenshots or a short description of manual testing (for UI changes)

5. **Request a review** from a maintainer. PRs require at least one approval before merge.

---

## Architecture overview

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed component diagrams.

The key constraint for contributors is the **shared document model**: `internal/document` is compiled into both the native binary and the WASM target. It must not import anything that would break `GOOS=js` compilation (e.g. `syscall`, `os` I/O beyond `os.Getenv`). The WASM bridge in `cmd/wasm/main.go` is gated by `//go:build js` and is the only place that may import `syscall/js`.

---

## Plugin development

See [PLUGIN_SDK.md](PLUGIN_SDK.md) for the complete guide to writing plugins.

Plugin authors import `github.com/tomott12345/go-cad/pkg/plugin` which is the only stable public API surface. The `internal/` packages are not part of the public API and may change between versions.

---

## Filing issues

- **Bug reports** — include Go version, OS, reproduction steps, and the full error output.
- **Feature requests** — describe the use case and expected behavior. Check existing issues first.
- **Security vulnerabilities** — please report privately; do not open a public issue.

---

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
