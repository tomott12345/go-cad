// Package plugin is the public SDK for go-cad plugins.
// Third-party developers import this package to implement a Plugin and interact
// with the host application via the HostAPI.
//
// Minimal plugin skeleton:
//
//	type MyPlugin struct{ api plugin.HostAPI }
//
//	func (p *MyPlugin) Name() string    { return "my-plugin" }
//	func (p *MyPlugin) Version() string { return "1.0.0" }
//	func (p *MyPlugin) Register(api plugin.HostAPI) error {
//	    p.api = api
//	    return api.RegisterCommand(plugin.CommandDescriptor{
//	        Name: "HELLO", Handler: func(args []string) error {
//	            _, err := api.AddEntity(plugin.Entity{Type: "line", X1: 0, Y1: 0, X2: 100, Y2: 0})
//	            return err
//	        },
//	    })
//	}
//	func (p *MyPlugin) Unregister() error { return nil }
//
//	// NewPlugin is the required symbol looked up by the loader.
//	func NewPlugin() plugin.Plugin { return &MyPlugin{} }
package plugin

// PluginAPIVersion is the semantic version of this SDK.
// Plugins should check this at registration time if they require a minimum API.
const PluginAPIVersion = "1.0.0"

// ─── Event system ─────────────────────────────────────────────────────────────

// EventKind identifies a host-application state change.
type EventKind string

const (
	EntityAdded      EventKind = "entity.added"
	EntityDeleted    EventKind = "entity.deleted"
	SelectionChanged EventKind = "selection.changed"
	DocumentSaved    EventKind = "document.saved"
	DocumentLoaded   EventKind = "document.loaded"
	ToolChanged      EventKind = "tool.changed"
)

// Event carries information about a state change in the host application.
type Event struct {
	Kind    EventKind `json:"kind"`
	Payload any       `json:"payload,omitempty"`
}

// EventHandler is invoked synchronously when a subscribed event fires.
// Handlers must not block the caller for extended periods.
type EventHandler func(Event)

// ─── Data types ───────────────────────────────────────────────────────────────

// Entity is the plugin-facing representation of a CAD primitive.
// It mirrors document.Entity; use the Type field to determine which coordinate
// fields are meaningful.
//
// Types and their fields:
//
//	"line"      — X1,Y1 (start), X2,Y2 (end)
//	"circle"    — CX,CY (centre), R (radius)
//	"arc"       — CX,CY,R, StartDeg,EndDeg
//	"rectangle" — X1,Y1 (top-left), X2,Y2 (bottom-right)
//	"polyline"  — Points [][]float64
type Entity struct {
	ID       int             `json:"id"`
	Type     string          `json:"type"`
	Layer    int             `json:"layer"`
	Color    string          `json:"color"`
	X1       float64         `json:"x1"`
	Y1       float64         `json:"y1"`
	X2       float64         `json:"x2"`
	Y2       float64         `json:"y2"`
	CX       float64         `json:"cx"`
	CY       float64         `json:"cy"`
	R        float64         `json:"r"`
	StartDeg float64         `json:"startDeg"`
	EndDeg   float64         `json:"endDeg"`
	Points   [][]float64     `json:"points,omitempty"`
	Extra    map[string]any  `json:"extra,omitempty"`
}

// DocumentInfo carries metadata about the current document.
type DocumentInfo struct {
	EntityCount int     `json:"entityCount"`
	Layers      []int   `json:"layers"`
	BBoxMinX    float64 `json:"bboxMinX"`
	BBoxMinY    float64 `json:"bboxMinY"`
	BBoxMaxX    float64 `json:"bboxMaxX"`
	BBoxMaxY    float64 `json:"bboxMaxY"`
}

// ToolDescriptor describes a drawing tool that a plugin registers with the host.
type ToolDescriptor struct {
	// Name is the unique display name of the tool (e.g. "Linear Dimension").
	Name string `json:"name"`
	// IconPath is an optional path to a PNG/SVG icon relative to the plugin directory.
	IconPath string `json:"iconPath"`
	// Shortcut is an optional keyboard shortcut string (e.g. "D", "Ctrl+D").
	Shortcut string `json:"shortcut"`
	// CursorType controls the canvas cursor when the tool is active.
	// Accepted values: "crosshair" (default), "pointer", "default".
	CursorType string `json:"cursorType"`
}

// CommandDescriptor describes a named command that a plugin registers.
// Commands are invoked by name from the command palette or REST API.
type CommandDescriptor struct {
	// Name is the canonical command name (e.g. "DIM", "HATCH").
	Name string
	// Aliases are alternative names that resolve to this command.
	Aliases []string
	// Handler is called when the command is executed with optional string arguments.
	Handler func(args []string) error
}

// PluginInfo is a lightweight summary returned by HostAPI.GetPlugins and the
// REST /api/v1/plugins endpoint.
type PluginInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ─── Interfaces ───────────────────────────────────────────────────────────────

// HostAPI is the interface the go-cad host exposes to loaded plugins.
// All methods are safe to call from any goroutine.
type HostAPI interface {
	// AddEntity adds a new entity and returns its assigned ID, or an error.
	AddEntity(e Entity) (int, error)
	// DeleteEntity removes the entity with the given ID. Returns false if not found.
	DeleteEntity(id int) bool
	// GetEntities returns a snapshot of all entities in the document.
	GetEntities() []Entity
	// GetDocument returns metadata about the current document.
	GetDocument() DocumentInfo

	// RegisterTool registers a drawing tool contributed by the plugin.
	RegisterTool(td ToolDescriptor) error
	// RegisterCommand registers a named command contributed by the plugin.
	RegisterCommand(cd CommandDescriptor) error

	// Subscribe registers an event handler and returns an opaque subscription ID.
	Subscribe(kind EventKind, handler EventHandler) string
	// Unsubscribe removes a previously registered subscription.
	Unsubscribe(subscriptionID string)
}

// Plugin is the interface every go-cad plugin must implement.
// The loader looks up a `NewPlugin` symbol in the .so or subprocess binary.
type Plugin interface {
	// Name returns the unique plugin name.
	Name() string
	// Version returns the plugin's semantic version string.
	Version() string
	// Register is called once when the plugin is loaded.
	// The plugin should save api for later use and register its tools/commands.
	Register(api HostAPI) error
	// Unregister is called when the plugin is unloaded.
	// The plugin should release any resources it holds.
	Unregister() error
}
