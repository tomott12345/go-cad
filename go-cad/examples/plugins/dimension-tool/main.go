// Package main is the dimension-tool example go-cad plugin.
//
// It registers a linear dimension tool and a DIM command.  When the DIM command
// is executed it adds a dimension annotation entity (represented as a polyline)
// between two points supplied as arguments.
//
// Build as a subprocess plugin (works on all platforms):
//
//	go build -o dimension-tool ./examples/plugins/dimension-tool/
//	mkdir -p ~/.go-cad/plugins
//	mv dimension-tool ~/.go-cad/plugins/
//
// Build as a Go .so plugin (Linux/macOS only, requires CGO):
//
//	go build -buildmode=plugin -o dimension-tool.so ./examples/plugins/dimension-tool/
//	mv dimension-tool.so ~/.go-cad/plugins/
//
// See PLUGIN_SDK.md for full details.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"go-cad/pkg/plugin"
)

// ─── Plugin implementation ───────────────────────────────────────────────────

// DimensionPlugin implements the linear dimension tool.
type DimensionPlugin struct {
	api plugin.HostAPI
}

// NewPlugin is the symbol the go-cad loader looks up when loading a .so file.
func NewPlugin() plugin.Plugin {
	return &DimensionPlugin{}
}

func (d *DimensionPlugin) Name() string    { return "dimension-tool" }
func (d *DimensionPlugin) Version() string { return "1.0.0" }

func (d *DimensionPlugin) Register(api plugin.HostAPI) error {
	d.api = api

	// Register the drawing tool.
	if err := api.RegisterTool(plugin.ToolDescriptor{
		Name:       "Linear Dimension",
		IconPath:   "icons/dimension.png",
		Shortcut:   "D",
		CursorType: "crosshair",
	}); err != nil {
		return err
	}

	// Register the DIM command.
	if err := api.RegisterCommand(plugin.CommandDescriptor{
		Name:    "DIM",
		Aliases: []string{"DIMENSION", "LINEARDIM"},
		Handler: d.handleDIM,
	}); err != nil {
		return err
	}

	return nil
}

func (d *DimensionPlugin) Unregister() error {
	d.api = nil
	return nil
}

// handleDIM creates a linear dimension annotation between two points.
// args format: ["x1,y1", "x2,y2"]  e.g. ["0,0", "100,0"]
func (d *DimensionPlugin) handleDIM(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("DIM requires two point arguments: x1,y1 x2,y2")
	}

	x1, y1, err := parsePoint(args[0])
	if err != nil {
		return fmt.Errorf("DIM: invalid first point %q: %w", args[0], err)
	}
	x2, y2, err := parsePoint(args[1])
	if err != nil {
		return fmt.Errorf("DIM: invalid second point %q: %w", args[1], err)
	}

	// Dimension line is drawn 20 units above the measured points.
	offset := 20.0

	// Witness lines (from measured points up to the dimension line).
	_, err = d.api.AddEntity(plugin.Entity{
		Type: "line",
		X1: x1, Y1: y1,
		X2: x1, Y2: y1 - offset,
	})
	if err != nil {
		return err
	}
	_, err = d.api.AddEntity(plugin.Entity{
		Type: "line",
		X1: x2, Y1: y2,
		X2: x2, Y2: y2 - offset,
	})
	if err != nil {
		return err
	}

	// Main dimension line with arrow ticks.
	dimY := math.Min(y1, y2) - offset
	tickSize := 5.0
	_, err = d.api.AddEntity(plugin.Entity{
		Type: "polyline",
		Points: [][]float64{
			{x1 + tickSize, dimY - tickSize / 2},
			{x1, dimY},
			{x1 + tickSize, dimY + tickSize / 2},
			{x1, dimY},
			{x2, dimY},
			{x2 - tickSize, dimY - tickSize / 2},
			{x2, dimY},
			{x2 - tickSize, dimY + tickSize / 2},
		},
	})
	return err
}

// parsePoint parses a "x,y" string into two float64 values.
func parsePoint(s string) (float64, float64, error) {
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected x,y format")
	}
	x, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, err
	}
	y, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, err
	}
	return x, y, nil
}

// ─── Subprocess JSON-RPC entrypoint ──────────────────────────────────────────
//
// When loaded as a subprocess (not a .so), this main() function drives the
// JSON-RPC protocol described in PLUGIN_SDK.md.

type rpcMsg struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   string          `json:"error,omitempty"`
	ID      int64           `json:"id,omitempty"`
}

func main() {
	p := NewPlugin()
	enc := json.NewEncoder(os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)

	// hostAPI proxies calls back to the host via JSON-RPC.
	hostProxy := &subprocessHostAPI{enc: enc, scanner: scanner}

	for scanner.Scan() {
		var msg rpcMsg
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		var result json.RawMessage
		var errStr string

		switch msg.Method {
		case "plugin.name":
			result, _ = json.Marshal(p.Name())

		case "plugin.version":
			result, _ = json.Marshal(p.Version())

		case "plugin.register":
			if err := p.Register(hostProxy); err != nil {
				errStr = err.Error()
			} else {
				result = json.RawMessage(`null`)
			}

		case "plugin.unregister":
			if err := p.Unregister(); err != nil {
				errStr = err.Error()
			} else {
				result = json.RawMessage(`null`)
			}

		default:
			errStr = fmt.Sprintf("unknown method: %s", msg.Method)
		}

		_ = enc.Encode(rpcMsg{
			JSONRPC: "2.0",
			Result:  result,
			Error:   errStr,
			ID:      msg.ID,
		})
	}
}

// ─── Subprocess HostAPI proxy ─────────────────────────────────────────────────

type subprocessHostAPI struct {
	enc     *json.Encoder
	scanner *bufio.Scanner
	idSeq   int64
}

func (h *subprocessHostAPI) nextID() int64 {
	h.idSeq++
	return h.idSeq
}

// call makes a reverse JSON-RPC call to the host and returns the raw result.
func (h *subprocessHostAPI) call(method string, params any) (json.RawMessage, error) {
	raw, _ := json.Marshal(params)
	id := h.nextID()
	if err := h.enc.Encode(rpcMsg{
		JSONRPC: "2.0",
		Method:  method,
		Params:  raw,
		ID:      id,
	}); err != nil {
		return nil, err
	}
	if !h.scanner.Scan() {
		return nil, fmt.Errorf("subprocess: host closed connection")
	}
	var resp rpcMsg
	if err := json.Unmarshal(h.scanner.Bytes(), &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("host: %s", resp.Error)
	}
	return resp.Result, nil
}

func (h *subprocessHostAPI) AddEntity(e plugin.Entity) (int, error) {
	result, err := h.call("host.addEntity", e)
	if err != nil {
		return 0, err
	}
	var id int
	_ = json.Unmarshal(result, &id)
	return id, nil
}

func (h *subprocessHostAPI) DeleteEntity(id int) bool {
	result, err := h.call("host.deleteEntity", id)
	if err != nil {
		return false
	}
	var ok bool
	_ = json.Unmarshal(result, &ok)
	return ok
}

func (h *subprocessHostAPI) GetEntities() []plugin.Entity { return nil }

func (h *subprocessHostAPI) GetDocument() plugin.DocumentInfo { return plugin.DocumentInfo{} }

func (h *subprocessHostAPI) RegisterTool(td plugin.ToolDescriptor) error {
	_, err := h.call("host.registerTool", td)
	return err
}

func (h *subprocessHostAPI) RegisterCommand(cd plugin.CommandDescriptor) error {
	type cmdPayload struct {
		Name    string   `json:"name"`
		Aliases []string `json:"aliases"`
	}
	_, err := h.call("host.registerCommand", cmdPayload{Name: cd.Name, Aliases: cd.Aliases})
	return err
}

func (h *subprocessHostAPI) Subscribe(_ plugin.EventKind, _ plugin.EventHandler) string {
	return ""
}

func (h *subprocessHostAPI) Unsubscribe(_ string) {}
