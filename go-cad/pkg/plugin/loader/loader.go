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
        "runtime"
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
// configured directories (.so files and executables).
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
                                continue
                        }
                        if l.cfg.EnableSubprocess && isExecutable(e, dir) {
                                found = append(found, filepath.Join(dir, name))
                        }
                }
        }
        return found
}

// LoadAll loads every discovered plugin into host. Errors for individual
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

// isExecutable reports whether the directory entry is an executable plugin.
// On Windows, plugins must have a .exe extension.
// On Unix/macOS, the execute bit must be set.
func isExecutable(e os.DirEntry, _ string) bool {
        if e.IsDir() {
                return false
        }
        name := e.Name()
        if strings.HasSuffix(name, ".so") {
                return false
        }
        if runtime.GOOS == "windows" {
                return strings.HasSuffix(strings.ToLower(name), ".exe")
        }
        // Unix: check execute bit.
        info, err := e.Info()
        if err != nil {
                return false
        }
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
// # Host → Plugin protocol (host sends via subprocess stdin)
//
//      {"jsonrpc":"2.0","method":"plugin.name","id":1}
//      {"jsonrpc":"2.0","method":"plugin.version","id":2}
//      {"jsonrpc":"2.0","method":"plugin.register","id":3}
//      {"jsonrpc":"2.0","method":"plugin.command","params":{"command":"DIM","args":["0,0","100,0"]},"id":4}
//      {"jsonrpc":"2.0","method":"plugin.unregister","id":5}
//
// # Plugin → Host reverse calls (interleaved within any request)
//
//      {"jsonrpc":"2.0","method":"host.addEntity","params":{...},"id":10}
//      {"jsonrpc":"2.0","method":"host.registerCommand","params":{...},"id":11}
//      {"jsonrpc":"2.0","method":"host.registerTool","params":{...},"id":12}
//
// The host processes all reverse calls before accepting the plugin's final
// response to the pending request.
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

        // Query name and version (simple call — no reverse calls expected here).
        nameRaw, err := sp.call("plugin.name", nil)
        if err != nil {
                _ = cmd.Process.Kill()
                return nil, fmt.Errorf("subprocess: plugin.name: %w", err)
        }
        _ = json.Unmarshal(nameRaw, &sp.name)

        verRaw, err := sp.call("plugin.version", nil)
        if err != nil {
                _ = cmd.Process.Kill()
                return nil, fmt.Errorf("subprocess: plugin.version: %w", err)
        }
        _ = json.Unmarshal(verRaw, &sp.version)

        return sp, nil
}

func (sp *SubprocessPlugin) Name() string    { return sp.name }
func (sp *SubprocessPlugin) Version() string { return sp.version }

// Register sends "plugin.register" and processes any reverse HostAPI calls the
// subprocess makes before it sends its response.
func (sp *SubprocessPlugin) Register(api plugin.HostAPI) error {
        sp.api = api
        _, err := sp.call("plugin.register", nil)
        return err
}

// Unregister sends "plugin.unregister" and terminates the subprocess.
// call() manages its own lock; do not acquire sp.mu here.
func (sp *SubprocessPlugin) Unregister() error {
        _, _ = sp.call("plugin.unregister", nil)
        _ = sp.stdin.Close()
        return sp.cmd.Wait()
}

// call sends a JSON-RPC request identified by a unique ID and reads the
// subprocess stdout, processing any interleaved reverse calls (host.addEntity
// etc.) until it sees the matching response.
//
// This single method handles all command dispatch correctly — reverse calls
// made by the subprocess during command execution are processed inline before
// the final response is returned to the caller.
func (sp *SubprocessPlugin) call(method string, params any) (json.RawMessage, error) {
        sp.mu.Lock()
        defer sp.mu.Unlock()

        id := sp.idSeq.Add(1)
        var rawParams json.RawMessage
        if params != nil {
                var err error
                rawParams, err = json.Marshal(params)
                if err != nil {
                        return nil, err
                }
        }
        if err := sp.enc.Encode(rpcMsg{JSONRPC: "2.0", Method: method, Params: rawParams, ID: id}); err != nil {
                return nil, err
        }

        // Read messages until we see our response; process interleaved reverse calls.
        for sp.scanner.Scan() {
                var msg rpcMsg
                if err := json.Unmarshal(sp.scanner.Bytes(), &msg); err != nil {
                        return nil, err
                }
                // Our response has a matching ID and no Method field.
                if msg.ID == id && msg.Method == "" {
                        if msg.Error != "" {
                                return nil, fmt.Errorf("subprocess: %s", msg.Error)
                        }
                        return msg.Result, nil
                }
                // Interleaved reverse call from the subprocess.
                if err := sp.handleReverseCall(msg); err != nil {
                        _ = sp.enc.Encode(rpcMsg{JSONRPC: "2.0", Error: err.Error(), ID: msg.ID})
                }
        }
        return nil, fmt.Errorf("subprocess: connection closed: %w", sp.scanner.Err())
}

// handleReverseCall executes a HostAPI call requested by the subprocess and
// writes the response.
func (sp *SubprocessPlugin) handleReverseCall(msg rpcMsg) error {
        if sp.api == nil {
                return sp.enc.Encode(rpcMsg{JSONRPC: "2.0", Error: "no host API available", ID: msg.ID})
        }
        var result json.RawMessage
        var errStr string

        switch msg.Method {
        case "host.addEntity":
                var e plugin.Entity
                if err := json.Unmarshal(msg.Params, &e); err != nil {
                        errStr = err.Error()
                } else {
                        id, err := sp.api.AddEntity(e)
                        if err != nil {
                                errStr = err.Error()
                        } else {
                                result, _ = json.Marshal(id)
                        }
                }

        case "host.deleteEntity":
                var id int
                if err := json.Unmarshal(msg.Params, &id); err != nil {
                        errStr = err.Error()
                } else {
                        ok := sp.api.DeleteEntity(id)
                        result, _ = json.Marshal(ok)
                }

        case "host.registerTool":
                var td plugin.ToolDescriptor
                if err := json.Unmarshal(msg.Params, &td); err != nil {
                        errStr = err.Error()
                } else if err := sp.api.RegisterTool(td); err != nil {
                        errStr = err.Error()
                } else {
                        result = json.RawMessage(`null`)
                }

        case "host.registerCommand":
                // The subprocess registers a command by name/aliases. When the host
                // later executes it, it sends "plugin.command" back to the subprocess.
                var cd struct {
                        Name    string   `json:"name"`
                        Aliases []string `json:"aliases"`
                }
                if err := json.Unmarshal(msg.Params, &cd); err != nil {
                        errStr = err.Error()
                } else {
                        cmdName := cd.Name // capture for closure
                        err := sp.api.RegisterCommand(plugin.CommandDescriptor{
                                Name:    cd.Name,
                                Aliases: cd.Aliases,
                                Handler: func(args []string) error {
                                        // Invoke the command on the subprocess via "plugin.command".
                                        type cmdReq struct {
                                                Command string   `json:"command"`
                                                Args    []string `json:"args"`
                                        }
                                        _, err := sp.call("plugin.command", cmdReq{Command: cmdName, Args: args})
                                        return err
                                },
                        })
                        if err != nil {
                                errStr = err.Error()
                        } else {
                                result = json.RawMessage(`null`)
                        }
                }

        default:
                errStr = fmt.Sprintf("unknown host method: %s", msg.Method)
        }

        return sp.enc.Encode(rpcMsg{
                JSONRPC: "2.0",
                Result:  result,
                Error:   errStr,
                ID:      msg.ID,
        })
}
