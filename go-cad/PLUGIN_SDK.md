# go-cad Plugin SDK

This guide explains how to write, build, and install a plugin for the go-cad
CAD engine. Plugins can add new drawing tools, register commands (invoked from
the command palette or via the REST API), listen to document events, and
interact with the full CAD document through a stable, versioned API.

---

## Prerequisites

- Go 1.22 or later
- A working go-cad development server (`go run ./cmd/serve`)

---

## Quick-start: scaffold a plugin

Create a new directory for your plugin (outside the go-cad repo):

```
mkdir my-plugin && cd my-plugin
go mod init my-plugin
go get github.com/tomott12345/go-cad/pkg/plugin@latest
```

Create `main.go`:

```go
package main

import "github.com/tomott12345/go-cad/pkg/plugin"

// MyPlugin is a minimal go-cad plugin.
type MyPlugin struct{ api plugin.HostAPI }

func (p *MyPlugin) Name() string    { return "my-plugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }

func (p *MyPlugin) Register(api plugin.HostAPI) error {
    p.api = api
    return api.RegisterCommand(plugin.CommandDescriptor{
        Name:    "HELLO",
        Aliases: []string{"HI"},
        Handler: func(args []string) error {
            _, err := api.AddEntity(plugin.Entity{
                Type: "line",
                X1: 0, Y1: 0, X2: 100, Y2: 0,
            })
            return err
        },
    })
}

func (p *MyPlugin) Unregister() error { return nil }

// NewPlugin is the symbol the loader looks up in .so files.
func NewPlugin() plugin.Plugin { return &MyPlugin{} }
```

---

## Transport options

go-cad supports two plugin transports. Choose based on your target platform:

### 1. Subprocess JSON-RPC (all platforms — recommended)

Build your plugin as a standalone executable:

```bash
go build -o my-plugin ./
```

Install it:

```bash
mkdir -p ~/.go-cad/plugins
cp my-plugin ~/.go-cad/plugins/
```

The plugin must expose a `main()` function that drives the JSON-RPC protocol
(see the [dimension-tool example](examples/plugins/dimension-tool/main.go) for
a complete implementation of the subprocess entrypoint).

### 2. Go plugin .so (Linux / macOS — requires CGO)

Build your plugin as a shared library:

```bash
go build -buildmode=plugin -o my-plugin.so ./
```

Install it:

```bash
cp my-plugin.so ~/.go-cad/plugins/
```

The loader calls the exported `NewPlugin() plugin.Plugin` symbol. The symbol
must be exported at package level.

**Important:** .so plugins must be compiled with the same Go version and module
graph as go-cad itself.

---

## Subprocess JSON-RPC protocol

When a plugin runs as a subprocess it communicates over stdin/stdout using
newline-delimited JSON-RPC 2.0.

### Host → Plugin calls

The host sends these requests to the plugin process via **stdin**:

| Method               | Params | Description                        |
|----------------------|--------|------------------------------------|
| `plugin.name`        | —      | Query plugin name                  |
| `plugin.version`     | —      | Query plugin version               |
| `plugin.register`    | —      | Tell plugin to register itself     |
| `plugin.unregister`  | —      | Tell plugin to clean up and exit   |

### Plugin → Host reverse calls

While handling `plugin.register` (or any host call), the plugin may send
**reverse calls** to the host via **stdout** before sending its own response:

| Method                  | Params                         | Description                   |
|-------------------------|--------------------------------|-------------------------------|
| `host.addEntity`        | `plugin.Entity` (JSON)         | Add a CAD entity              |
| `host.deleteEntity`     | entity ID (int)                | Delete a CAD entity           |
| `host.registerTool`     | `plugin.ToolDescriptor` (JSON) | Register a drawing tool       |
| `host.registerCommand`  | `{name, aliases}` (JSON)       | Register a command            |

### Example exchange

```
Host → stdin:  {"jsonrpc":"2.0","method":"plugin.name","id":1}
Plugin → stdout: {"jsonrpc":"2.0","result":"dimension-tool","id":1}

Host → stdin:  {"jsonrpc":"2.0","method":"plugin.register","id":2}
Plugin → stdout: {"jsonrpc":"2.0","method":"host.registerTool",
                   "params":{"name":"Linear Dimension","shortcut":"D"},"id":10}
Host → stdin:  {"jsonrpc":"2.0","result":null,"id":10}
Plugin → stdout: {"jsonrpc":"2.0","result":null,"id":2}   ← register response
```

---

## HostAPI reference

All methods are safe to call from any goroutine.

### Entity operations

```go
// AddEntity adds a primitive and returns its assigned ID.
// Returns an error for unknown entity types.
AddEntity(e plugin.Entity) (int, error)

// DeleteEntity removes the entity with the given ID.
// Returns false if the entity does not exist.
DeleteEntity(id int) bool

// GetEntities returns a snapshot of all entities.
GetEntities() []plugin.Entity

// GetDocument returns metadata: entity count, layers, bounding box.
GetDocument() plugin.DocumentInfo
```

### Tool & command registration

```go
// RegisterTool contributes a drawing tool to the host application.
RegisterTool(td plugin.ToolDescriptor) error

// RegisterCommand contributes a named command and zero or more aliases.
// The Handler is called with []string arguments when the command fires.
RegisterCommand(cd plugin.CommandDescriptor) error
```

### Event pub/sub

```go
// Subscribe registers handler for events of the given kind.
// Returns an opaque subscription ID for later cancellation.
Subscribe(kind plugin.EventKind, handler plugin.EventHandler) string

// Unsubscribe removes the subscription.
Unsubscribe(subscriptionID string)
```

**Event kinds:**

| Constant                  | String value          | Payload         |
|---------------------------|-----------------------|-----------------|
| `plugin.EntityAdded`      | `entity.added`        | entity ID (int) |
| `plugin.EntityDeleted`    | `entity.deleted`      | entity ID (int) |
| `plugin.SelectionChanged` | `selection.changed`   | —               |
| `plugin.DocumentSaved`    | `document.saved`      | —               |
| `plugin.DocumentLoaded`   | `document.loaded`     | —               |
| `plugin.ToolChanged`      | `tool.changed`        | tool name       |

---

## Entity types

| `type`       | Meaningful fields                        |
|--------------|------------------------------------------|
| `line`       | `x1,y1` (start) · `x2,y2` (end)         |
| `circle`     | `cx,cy` (centre) · `r` (radius)         |
| `arc`        | `cx,cy,r` · `startDeg,endDeg`           |
| `rectangle`  | `x1,y1` (top-left) · `x2,y2` (bottom-right) |
| `polyline`   | `points` `[][]float64` — list of [x,y]  |

All entities also carry optional `layer int` (default 0) and `color string`
(default `"#ffffff"`).

---

## Testing your plugin against the dev server

Start the server:

```bash
cd /path/to/go-cad
go run ./cmd/serve -plugins ./my-plugin-dir
```

Add an entity via the REST API and confirm your plugin's event handler fires:

```bash
# Add a line
curl -s -X POST http://localhost:8080/api/v1/entities \
  -H 'Content-Type: application/json' \
  -d '{"type":"line","x1":0,"y1":0,"x2":100,"y2":0}' | jq

# Execute your plugin's command
curl -s -X POST http://localhost:8080/api/v1/command \
  -H 'Content-Type: application/json' \
  -d '{"command":"HELLO"}' | jq

# List loaded plugins
curl -s http://localhost:8080/api/v1/plugins | jq
```

---

## Full example: dimension-tool

See [`examples/plugins/dimension-tool/main.go`](examples/plugins/dimension-tool/main.go)
for a complete plugin that:

- Registers a **Linear Dimension** tool with keyboard shortcut `D`
- Registers a `DIM` command (aliases: `DIMENSION`, `LINEARDIM`)
- Adds witness lines and a dimension annotation polyline when invoked

Build and test it:

```bash
go build -o dim-tool ./examples/plugins/dimension-tool/
mkdir -p ~/.go-cad/plugins
cp dim-tool ~/.go-cad/plugins/

go run ./cmd/serve
curl -s -X POST http://localhost:8080/api/v1/command \
  -H 'Content-Type: application/json' \
  -d '{"command":"DIM","args":["0,0","100,0"]}' | jq
```

---

## Plugin API versioning

The current SDK version is **1.0.0** (`plugin.PluginAPIVersion`). The major
version is incremented only on breaking interface changes. Plugins may check
this constant at registration time:

```go
func (p *MyPlugin) Register(api plugin.HostAPI) error {
    if plugin.PluginAPIVersion < "1.0.0" {
        return fmt.Errorf("requires go-cad plugin API >= 1.0.0")
    }
    // ...
}
```

---

## Directory structure reference

```
~/.go-cad/
└── plugins/
    ├── dimension-tool        # subprocess plugin (executable)
    └── my-feature.so         # .so plugin (Linux/macOS + CGO)

./plugins/                    # project-local plugin directory
    └── ...
```

The server scans both `~/.go-cad/plugins/` and `./plugins/` at startup, plus
any additional directory specified with the `-plugins` flag.
