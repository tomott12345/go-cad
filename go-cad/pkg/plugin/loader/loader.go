// Package loader discovers and loads go-cad plugins from configurable
// directories. Two plugin transports are supported:
//
//  1. Go plugin .so (Linux/macOS, requires CGO): the .so must export a symbol
//     `NewPlugin func() plugin.Plugin`.
//
//  2. Subprocess JSON-RPC (all platforms including Windows/WASM): an executable
//     in the plugin directory communicates with the host over stdin/stdout using
//     newline-delimited JSON-RPC 2.0.  See SubprocessPlugin for the protocol.
//
// Default search paths: ~/.go-cad/plugins/ and ./plugins/ (relative to the
// working directory of the host process).
package loader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"go-cad/pkg/plugin"
)

// Config controls which directories are scanned and which transports are enabled.
type Config struct {
	// Dirs lists the directories to scan for plugin files.
	// Defaults to ["~/.go-cad/plugins", "./plugins"] when nil.
	Dirs []string
	// EnableSO enables loading of Go plugin .so files.  Requires CGO.
	EnableSO bool
	// EnableSubprocess enables loading of subprocess plugins.
	EnableSubprocess bool
}

// DefaultConfig returns a Config with the standard search paths and all
// transports enabled.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		Dirs:             []string{filepath.Join(home, ".go-cad", "plugins"), "./plugins"},
		EnableSO:         true,
		EnableSubprocess: true,
	}
}

// Loader discovers and loads plugins from configured directories.
type Loader struct {
	cfg Config
}

// New creates a Loader using the provided configuration.
func New(cfg Config) *Loader {
	return &Loader{cfg: cfg}
}

// Discover returns the paths of all loadable plugin files found in the
// configured directories (.so files and executables without extension on
// Unix, or .exe on Windows).
func (l *Loader) Discover() []string {
	var found []string
	for _, dir := range l.cfg.Dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if l.cfg.EnableSO && strings.HasSuffix(name, ".so") {
				found = append(found, filepath.Join(dir, name))
			}
			if l.cfg.EnableSubprocess && isExecutable(e, dir) {
				found = append(found, filepath.Join(dir, name))
			}
		}
	}
	return found
}

// LoadAll loads every discovered plugin into host.  Errors for individual
// plugins are collected and returned together; a partial load is still
// considered a success.
func (l *Loader) LoadAll(host interface {
	LoadPlugin(plugin.Plugin) error
}) []error {
	var errs []error
	for _, path := range l.Discover() {
		var p plugin.Plugin
		var err error
		if strings.HasSuffix(path, ".so") {
			p, err = LoadSO(path)
		} else {
			p, err = LoadSubprocess(path)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("loader: %s: %w", path, err))
			continue
		}
		if err := host.LoadPlugin(p); err != nil {
			errs = append(errs, fmt.Errorf("loader: %s: %w", path, err))
		}
	}
	return errs
}

// isExecutable reports whether the directory entry is an executable file
// (non-.so, non-directory).
func isExecutable(e os.DirEntry, dir string) bool {
	if e.IsDir() {
		return false
	}
	name := e.Name()
	if strings.HasSuffix(name, ".so") {
		return false
	}
	info, err := e.Info()
	if err != nil {
		return false
	}
	// On Unix, check the execute bit.
	return info.Mode()&0o111 != 0
}

// ─── Subprocess JSON-RPC plugin ───────────────────────────────────────────────

// rpcMsg is the wire format for newline-delimited JSON-RPC 2.0.
type rpcMsg struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   string          `json:"error,omitempty"`
	ID      int64           `json:"id,omitempty"`
}

// SubprocessPlugin wraps a subprocess that communicates via newline-delimited
// JSON-RPC 2.0 over stdin/stdout.
//
// Host→Plugin protocol:
//
//	{"jsonrpc":"2.0","method":"plugin.name","id":1}
//	{"jsonrpc":"2.0","method":"plugin.version","id":2}
//	{"jsonrpc":"2.0","method":"plugin.register","id":3}
//	{"jsonrpc":"2.0","method":"plugin.unregister","id":4}
//
// Plugin→Host reverse calls (during plugin.register handling):
//
//	{"jsonrpc":"2.0","method":"host.addEntity","params":{...},"id":10}
//	{"jsonrpc":"2.0","method":"host.registerCommand","params":{...},"id":11}
//
// After all reverse calls the plugin sends its response to the pending call:
//
//	{"jsonrpc":"2.0","result":null,"id":3}
type SubprocessPlugin struct {
	path    string
	name    string
	version string
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	scanner *bufio.Scanner
	enc     *json.Encoder
	mu      sync.Mutex
	idSeq   atomic.Int64
	api     plugin.HostAPI
}

// LoadSubprocess starts the subprocess at path, queries its name and version,
// and returns a SubprocessPlugin ready for registration.
func LoadSubprocess(path string) (plugin.Plugin, error) {
	cmd := exec.Command(path) //nolint:gosec
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("subprocess: stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("subprocess: stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("subprocess: start: %w", err)
	}

	sp := &SubprocessPlugin{
		path:    path,
		cmd:     cmd,
		stdin:   stdin,
		scanner: bufio.NewScanner(stdout),
		enc:     json.NewEncoder(stdin),
	}

	// Query name and version.
	nameStr, err := sp.call("plugin.name", nil)
	if err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("subprocess: plugin.name: %w", err)
	}
	_ = json.Unmarshal(nameStr, &sp.name)

	verStr, err := sp.call("plugin.version", nil)
	if err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("subprocess: plugin.version: %w", err)
	}
	_ = json.Unmarshal(verStr, &sp.version)

	return sp, nil
}

func (sp *SubprocessPlugin) Name() string    { return sp.name }
func (sp *SubprocessPlugin) Version() string { return sp.version }

// Register sends "plugin.register" and processes any reverse HostAPI calls the
// subprocess makes before it sends its response.
func (sp *SubprocessPlugin) Register(api plugin.HostAPI) error {
	sp.api = api
	sp.mu.Lock()
	defer sp.mu.Unlock()

	id := sp.idSeq.Add(1)
	if err := sp.enc.Encode(rpcMsg{JSONRPC: "2.0", Method: "plugin.register", ID: id}); err != nil {
		return err
	}
	return sp.readUntilResponse(id)
}

// Unregister sends "plugin.unregister" and terminates the subprocess.
func (sp *SubprocessPlugin) Unregister() error {
	sp.mu.Lock()
	id := sp.idSeq.Add(1)
	_ = sp.enc.Encode(rpcMsg{JSONRPC: "2.0", Method: "plugin.unregister", ID: id})
	sp.mu.Unlock()
	_ = sp.stdin.Close()
	return sp.cmd.Wait()
}

// call sends a JSON-RPC request and returns the raw result bytes.
func (sp *SubprocessPlugin) call(method string, params any) (json.RawMessage, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	id := sp.idSeq.Add(1)
	var raw json.RawMessage
	if params != nil {
		var err error
		raw, err = json.Marshal(params)
		if err != nil {
			return nil, err
		}
	}
	if err := sp.enc.Encode(rpcMsg{JSONRPC: "2.0", Method: method, Params: raw, ID: id}); err != nil {
		return nil, err
	}
	if !sp.scanner.Scan() {
		return nil, fmt.Errorf("subprocess: read response: %w", sp.scanner.Err())
	}
	var resp rpcMsg
	if err := json.Unmarshal(sp.scanner.Bytes(), &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("subprocess: %s", resp.Error)
	}
	return resp.Result, nil
}

// readUntilResponse handles interleaved reverse calls from the subprocess until
// the response with the given id arrives.
func (sp *SubprocessPlugin) readUntilResponse(pendingID int64) error {
	for sp.scanner.Scan() {
		var msg rpcMsg
		if err := json.Unmarshal(sp.scanner.Bytes(), &msg); err != nil {
			return err
		}
		// If it's the response to our pending call, we're done.
		if msg.ID == pendingID && msg.Method == "" {
			if msg.Error != "" {
				return fmt.Errorf("subprocess: %s", msg.Error)
			}
			return nil
		}
		// Otherwise it's a reverse call from the plugin to the host.
		if err := sp.handleReverseCall(msg); err != nil {
			// Respond with error and continue.
			_ = sp.enc.Encode(rpcMsg{
				JSONRPC: "2.0", Error: err.Error(), ID: msg.ID,
			})
		}
	}
	return sp.scanner.Err()
}

// handleReverseCall executes a HostAPI call requested by the subprocess.
func (sp *SubprocessPlugin) handleReverseCall(msg rpcMsg) error {
	if sp.api == nil {
		return fmt.Errorf("no host API available")
	}
	var respResult json.RawMessage
	var respErr string

	switch msg.Method {
	case "host.addEntity":
		var e plugin.Entity
		if err := json.Unmarshal(msg.Params, &e); err != nil {
			respErr = err.Error()
		} else {
			id, err := sp.api.AddEntity(e)
			if err != nil {
				respErr = err.Error()
			} else {
				respResult, _ = json.Marshal(id)
			}
		}
	case "host.deleteEntity":
		var id int
		if err := json.Unmarshal(msg.Params, &id); err != nil {
			respErr = err.Error()
		} else {
			ok := sp.api.DeleteEntity(id)
			respResult, _ = json.Marshal(ok)
		}
	case "host.registerCommand":
		// Subprocess registers a command: only name/aliases are useful here;
		// the actual handler is the subprocess itself.
		var cd struct {
			Name    string   `json:"name"`
			Aliases []string `json:"aliases"`
		}
		if err := json.Unmarshal(msg.Params, &cd); err != nil {
			respErr = err.Error()
		} else {
			cmdRef := cd.Name // capture
			err := sp.api.RegisterCommand(plugin.CommandDescriptor{
				Name:    cd.Name,
				Aliases: cd.Aliases,
				Handler: func(args []string) error {
					params, _ := json.Marshal(args)
					_, err := sp.call("host.command."+cmdRef, params)
					return err
				},
			})
			if err != nil {
				respErr = err.Error()
			} else {
				respResult = json.RawMessage(`null`)
			}
		}
	case "host.registerTool":
		var td plugin.ToolDescriptor
		if err := json.Unmarshal(msg.Params, &td); err != nil {
			respErr = err.Error()
		} else if err := sp.api.RegisterTool(td); err != nil {
			respErr = err.Error()
		} else {
			respResult = json.RawMessage(`null`)
		}
	default:
		respErr = fmt.Sprintf("unknown host method: %s", msg.Method)
	}

	return sp.enc.Encode(rpcMsg{
		JSONRPC: "2.0",
		Result:  respResult,
		Error:   respErr,
		ID:      msg.ID,
	})
}
