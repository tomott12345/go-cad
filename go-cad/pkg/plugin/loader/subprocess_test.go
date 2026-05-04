package loader

// White-box integration tests for SubprocessPlugin using in-process pipe I/O.
// A minimal JSON-RPC plugin goroutine is wired directly to a SubprocessPlugin
// via io.Pipe pairs, avoiding OS-level subprocess spawning.

import (
        "bufio"
        "encoding/json"
        "io"
        "sync"
        "sync/atomic"
        "testing"

        "go-cad/pkg/plugin"
)

// ─── testHostAPI ─────────────────────────────────────────────────────────────

// testHostAPI is a minimal plugin.HostAPI used in subprocess tests.
// It avoids importing internal/pluginhost (which would create a circular dep).
type testHostAPI struct {
        mu       sync.Mutex
        entities []plugin.Entity
        tools    []plugin.ToolDescriptor
        commands []plugin.CommandDescriptor
        handlers map[string]func([]string) error
}

func newTestHostAPI() *testHostAPI {
        return &testHostAPI{handlers: map[string]func([]string) error{}}
}

func (h *testHostAPI) AddEntity(e plugin.Entity) (int, error) {
        h.mu.Lock()
        defer h.mu.Unlock()
        h.entities = append(h.entities, e)
        return len(h.entities) - 1, nil
}
func (h *testHostAPI) DeleteEntity(id int) bool         { return false }
func (h *testHostAPI) GetEntities() []plugin.Entity     { return h.entities }
func (h *testHostAPI) GetDocument() plugin.DocumentInfo { return plugin.DocumentInfo{} }
func (h *testHostAPI) RegisterTool(td plugin.ToolDescriptor) error {
        h.mu.Lock()
        h.tools = append(h.tools, td)
        h.mu.Unlock()
        return nil
}
func (h *testHostAPI) RegisterCommand(cd plugin.CommandDescriptor) error {
        h.mu.Lock()
        h.commands = append(h.commands, cd)
        if cd.Handler != nil {
                for _, name := range append([]string{cd.Name}, cd.Aliases...) {
                        h.handlers[name] = cd.Handler
                }
        }
        h.mu.Unlock()
        return nil
}
func (h *testHostAPI) Subscribe(kind plugin.EventKind, handler plugin.EventHandler) string {
        return ""
}
func (h *testHostAPI) Unsubscribe(subscriptionID string) {}

// executeCommand dispatches via handlers (registered by the subprocess Register).
func (h *testHostAPI) executeCommand(name string, args []string) error {
        h.mu.Lock()
        fn, ok := h.handlers[name]
        h.mu.Unlock()
        if !ok {
                return io.ErrNoProgress // sentinel — "not found locally"
        }
        return fn(args)
}

// ─── pipeConn ────────────────────────────────────────────────────────────────

// pipeConn creates two unidirectional io.Pipe pairs:
//
//      Pipe A: host encodes  → plugin scans   (host → plugin)
//      Pipe B: plugin encodes → host scans    (plugin → host)
//
// The host SubprocessPlugin is created with (hostEnc, hostScanner, hostStdin).
// The plugin goroutine receives (pluginEnc, pluginScanner).
func pipeConn() (
        hostEnc *json.Encoder, hostScanner *bufio.Scanner, hostStdin io.WriteCloser,
        pluginEnc *json.Encoder, pluginScanner *bufio.Scanner,
) {
        aPR, aPW := io.Pipe() // pipe A: host writes, plugin reads
        bPR, bPW := io.Pipe() // pipe B: plugin writes, host reads

        hostEnc = json.NewEncoder(aPW)
        hostStdin = aPW
        hostScanner = bufio.NewScanner(bPR)

        pluginEnc = json.NewEncoder(bPW)
        pluginScanner = bufio.NewScanner(aPR)
        return
}

// newSubprocessPluginFromPipes builds a SubprocessPlugin with injected I/O,
// bypassing exec.Command for deterministic testing.
func newSubprocessPluginFromPipes(
        enc *json.Encoder,
        scanner *bufio.Scanner,
        stdin io.WriteCloser,
) *SubprocessPlugin {
        return &SubprocessPlugin{
                enc:     enc,
                scanner: scanner,
                stdin:   stdin,
        }
}

// ─── minimal subprocess plugin goroutine ─────────────────────────────────────

// servePlugin runs a minimal JSON-RPC plugin.  It reads from pluginScanner
// (pipe A) and writes to pluginEnc (pipe B).
//
// Reverse calls (host.registerCommand, host.addEntity) are sent on pipe B
// and their responses arrive on pipe A — the SAME pipe as normal messages.
// This works because the main loop and callHost are sequential (single goroutine).
//
// The goroutine signals done when it processes "plugin.unregister".
func servePlugin(pluginEnc *json.Encoder, pluginScanner *bufio.Scanner, done *sync.WaitGroup, withEntity bool) {
        var callID int64

        // callHost sends a reverse call on pipe B and reads the host's ACK from pipe A.
        callHost := func(method string, params any) {
                id := atomic.AddInt64(&callID, 1)
                raw, _ := json.Marshal(params)
                pluginEnc.Encode(rpcMsg{JSONRPC: "2.0", Method: method, Params: raw, ID: id})
                // Read the host ACK from pipe A (same scanner used by the main loop, but
                // we own it exclusively here because the main loop called us synchronously).
                if pluginScanner.Scan() {
                        // ACK consumed; ignore content for test purposes.
                }
        }

        commands := map[string]bool{}

        for pluginScanner.Scan() {
                var msg rpcMsg
                json.Unmarshal(pluginScanner.Bytes(), &msg)

                var result json.RawMessage
                var errStr string

                switch msg.Method {
                case "plugin.name":
                        result, _ = json.Marshal("pipe-plugin")
                case "plugin.version":
                        result, _ = json.Marshal("0.0.1")
                case "plugin.register":
                        type cmdPayload struct {
                                Name    string   `json:"name"`
                                Aliases []string `json:"aliases"`
                        }
                        callHost("host.registerCommand", cmdPayload{Name: "PING", Aliases: []string{"P"}})
                        commands["PING"] = true
                        commands["P"] = true
                        result = json.RawMessage(`null`)
                case "plugin.command":
                        var body struct {
                                Command string   `json:"command"`
                                Args    []string `json:"args"`
                        }
                        json.Unmarshal(msg.Params, &body)
                        if commands[body.Command] {
                                if withEntity {
                                        callHost("host.addEntity", plugin.Entity{Type: "line", X1: 0, Y1: 0, X2: 10, Y2: 10})
                                }
                                result = json.RawMessage(`null`)
                        } else {
                                errStr = "unknown command: " + body.Command
                        }
                case "plugin.unregister":
                        result = json.RawMessage(`null`)
                        pluginEnc.Encode(rpcMsg{JSONRPC: "2.0", Result: result, ID: msg.ID})
                        done.Done()
                        return
                default:
                        errStr = "unknown: " + msg.Method
                }

                pluginEnc.Encode(rpcMsg{JSONRPC: "2.0", Result: result, Error: errStr, ID: msg.ID})
        }
        done.Done()
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestSubprocessPlugin_NameVersion(t *testing.T) {
        hostEnc, hostScanner, hostStdin, pluginEnc, pluginScanner := pipeConn()

        var wg sync.WaitGroup
        wg.Add(1)
        go servePlugin(pluginEnc, pluginScanner, &wg, false)

        sp := newSubprocessPluginFromPipes(hostEnc, hostScanner, hostStdin)

        nameRaw, err := sp.call("plugin.name", nil)
        if err != nil {
                t.Fatalf("plugin.name: %v", err)
        }
        var name string
        json.Unmarshal(nameRaw, &name)
        if name != "pipe-plugin" {
                t.Errorf("name = %q, want pipe-plugin", name)
        }

        verRaw, err := sp.call("plugin.version", nil)
        if err != nil {
                t.Fatalf("plugin.version: %v", err)
        }
        var ver string
        json.Unmarshal(verRaw, &ver)
        if ver != "0.0.1" {
                t.Errorf("version = %q, want 0.0.1", ver)
        }

        hostStdin.Close()
        wg.Wait()
}

func TestSubprocessPlugin_CommandRoundtrip(t *testing.T) {
        hostEnc, hostScanner, hostStdin, pluginEnc, pluginScanner := pipeConn()

        var wg sync.WaitGroup
        wg.Add(1)
        go servePlugin(pluginEnc, pluginScanner, &wg, false)

        sp := newSubprocessPluginFromPipes(hostEnc, hostScanner, hostStdin)
        sp.name = "pipe-plugin"
        sp.version = "0.0.1"

        host := newTestHostAPI()
        sp.api = host

        // Register: subprocess sends host.registerCommand for PING and P.
        if err := sp.Register(host); err != nil {
                t.Fatalf("Register: %v", err)
        }
        if len(host.commands) == 0 {
                t.Fatal("expected at least one command registered")
        }

        // Execute PING — should succeed without error.
        _, err := sp.call("plugin.command", map[string]any{"command": "PING", "args": []string{}})
        if err != nil {
                t.Fatalf("PING command: %v", err)
        }

        // Execute alias P.
        _, err = sp.call("plugin.command", map[string]any{"command": "P", "args": []string{}})
        if err != nil {
                t.Fatalf("P (alias) command: %v", err)
        }

        // Unregister.
        if err := sp.Unregister(); err != nil {
                t.Logf("Unregister (may error on pipe close): %v", err)
        }
        wg.Wait()
}

func TestSubprocessPlugin_CommandWithReverseCall(t *testing.T) {
        hostEnc, hostScanner, hostStdin, pluginEnc, pluginScanner := pipeConn()

        var wg sync.WaitGroup
        wg.Add(1)
        go servePlugin(pluginEnc, pluginScanner, &wg, true /* withEntity */)

        sp := newSubprocessPluginFromPipes(hostEnc, hostScanner, hostStdin)
        sp.name = "pipe-plugin"
        sp.version = "0.0.1"

        host := newTestHostAPI()
        sp.api = host

        if err := sp.Register(host); err != nil {
                t.Fatalf("Register: %v", err)
        }

        // PING with withEntity=true causes the subprocess to call host.addEntity.
        _, err := sp.call("plugin.command", map[string]any{"command": "PING", "args": []string{}})
        if err != nil {
                t.Fatalf("PING command: %v", err)
        }

        // The reverse call (host.addEntity) should have been handled by call() inline.
        host.mu.Lock()
        n := len(host.entities)
        host.mu.Unlock()
        if n != 1 {
                t.Errorf("expected 1 entity after command+reverse call, got %d", n)
        }

        _ = hostStdin.Close()
        wg.Wait()
}

func TestSubprocessPlugin_UnknownCommand(t *testing.T) {
        hostEnc, hostScanner, hostStdin, pluginEnc, pluginScanner := pipeConn()

        var wg sync.WaitGroup
        wg.Add(1)
        go servePlugin(pluginEnc, pluginScanner, &wg, false)

        sp := newSubprocessPluginFromPipes(hostEnc, hostScanner, hostStdin)
        sp.name = "pipe-plugin"
        sp.version = "0.0.1"
        sp.api = newTestHostAPI()

        if err := sp.Register(sp.api.(*testHostAPI)); err != nil {
                t.Fatalf("Register: %v", err)
        }

        _, err := sp.call("plugin.command", map[string]any{"command": "NOSUCH", "args": []string{}})
        if err == nil {
                t.Error("expected error for unknown command, got nil")
        }

        _ = hostStdin.Close()
        wg.Wait()
}
