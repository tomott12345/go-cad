package pluginhost_test

import (
        "testing"

        "github.com/tomott12345/go-cad/internal/document"
        "github.com/tomott12345/go-cad/internal/pluginhost"
        "github.com/tomott12345/go-cad/pkg/plugin"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

type fakePlugin struct {
        name       string
        version    string
        registerFn func(plugin.HostAPI) error
        calls      []string
}

func (f *fakePlugin) Name() string    { return f.name }
func (f *fakePlugin) Version() string { return f.version }
func (f *fakePlugin) Register(api plugin.HostAPI) error {
        f.calls = append(f.calls, "register")
        if f.registerFn != nil {
                return f.registerFn(api)
        }
        return nil
}
func (f *fakePlugin) Unregister() error {
        f.calls = append(f.calls, "unregister")
        return nil
}

func newHost() *pluginhost.Host {
        return pluginhost.New(document.New())
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestHost_LoadUnload(t *testing.T) {
        h := newHost()
        p := &fakePlugin{name: "test", version: "1.0.0"}

        if err := h.LoadPlugin(p); err != nil {
                t.Fatalf("LoadPlugin: %v", err)
        }
        if len(p.calls) != 1 || p.calls[0] != "register" {
                t.Errorf("expected register call, got %v", p.calls)
        }

        plugins := h.ListPlugins()
        if len(plugins) != 1 || plugins[0].Name != "test" {
                t.Errorf("ListPlugins: %v", plugins)
        }

        if err := h.UnloadPlugin("test"); err != nil {
                t.Fatalf("UnloadPlugin: %v", err)
        }
        if len(p.calls) != 2 || p.calls[1] != "unregister" {
                t.Errorf("expected unregister call, got %v", p.calls)
        }
        if len(h.ListPlugins()) != 0 {
                t.Error("expected no plugins after unload")
        }
}

func TestHost_UnloadMissing(t *testing.T) {
        h := newHost()
        if err := h.UnloadPlugin("missing"); err == nil {
                t.Error("expected error for missing plugin")
        }
}

func TestHost_AddDeleteEntity(t *testing.T) {
        h := newHost()

        id, err := h.AddEntity(plugin.Entity{Type: "line", X1: 0, Y1: 0, X2: 10, Y2: 0})
        if err != nil {
                t.Fatalf("AddEntity: %v", err)
        }
        if id <= 0 {
                t.Errorf("expected positive ID, got %d", id)
        }

        entities := h.GetEntities()
        if len(entities) != 1 {
                t.Fatalf("expected 1 entity, got %d", len(entities))
        }
        if entities[0].ID != id || entities[0].Type != "line" {
                t.Errorf("entity mismatch: %+v", entities[0])
        }

        if !h.DeleteEntity(id) {
                t.Error("DeleteEntity returned false")
        }
        if len(h.GetEntities()) != 0 {
                t.Error("entity not deleted")
        }
}

func TestHost_AddEntity_AllTypes(t *testing.T) {
        h := newHost()
        cases := []plugin.Entity{
                {Type: "line", X1: 0, Y1: 0, X2: 100, Y2: 100},
                {Type: "circle", CX: 50, CY: 50, R: 25},
                {Type: "arc", CX: 0, CY: 0, R: 10, StartDeg: 0, EndDeg: 90},
                {Type: "rectangle", X1: 0, Y1: 0, X2: 50, Y2: 30},
                {Type: "polyline", Points: [][]float64{{0, 0}, {10, 10}, {20, 0}}},
        }
        for _, e := range cases {
                id, err := h.AddEntity(e)
                if err != nil {
                        t.Errorf("AddEntity %q: %v", e.Type, err)
                }
                if id <= 0 {
                        t.Errorf("AddEntity %q: bad ID %d", e.Type, id)
                }
        }
}

func TestHost_AddEntity_UnknownType(t *testing.T) {
        h := newHost()
        _, err := h.AddEntity(plugin.Entity{Type: "unknown"})
        if err == nil {
                t.Error("expected error for unknown entity type")
        }
}

func TestHost_GetDocument(t *testing.T) {
        h := newHost()
        h.AddEntity(plugin.Entity{Type: "line", X1: 0, Y1: 0, X2: 10, Y2: 5, Layer: 2})
        info := h.GetDocument()
        if info.EntityCount != 1 {
                t.Errorf("EntityCount: %d", info.EntityCount)
        }
        if len(info.Layers) != 1 || info.Layers[0] != 2 {
                t.Errorf("Layers: %v", info.Layers)
        }
}

func TestHost_GetDocument_Empty(t *testing.T) {
        h := newHost()
        info := h.GetDocument()
        if info.EntityCount != 0 {
                t.Errorf("expected 0 entities, got %d", info.EntityCount)
        }
        if len(info.Layers) != 0 {
                t.Errorf("expected no layers, got %v", info.Layers)
        }
}

func TestHost_RegisterCommand(t *testing.T) {
        h := newHost()
        called := false
        err := h.RegisterCommand(plugin.CommandDescriptor{
                Name:    "MY_CMD",
                Aliases: []string{"MC"},
                Handler: func(args []string) error {
                        called = true
                        return nil
                },
        })
        if err != nil {
                t.Fatalf("RegisterCommand: %v", err)
        }

        if err := h.ExecuteCommand("MY_CMD", nil); err != nil {
                t.Fatalf("ExecuteCommand: %v", err)
        }
        if !called {
                t.Error("handler not called")
        }

        // Alias should also work.
        called = false
        if err := h.ExecuteCommand("MC", nil); err != nil {
                t.Fatalf("ExecuteCommand alias: %v", err)
        }
        if !called {
                t.Error("alias handler not called")
        }
}

func TestHost_RegisterCommand_Validation(t *testing.T) {
        h := newHost()
        if err := h.RegisterCommand(plugin.CommandDescriptor{}); err == nil {
                t.Error("expected error for empty command name")
        }
        if err := h.RegisterCommand(plugin.CommandDescriptor{Name: "X"}); err == nil {
                t.Error("expected error for nil handler")
        }
}

func TestHost_ExecuteCommand_NotFound(t *testing.T) {
        h := newHost()
        if err := h.ExecuteCommand("NOPE", nil); err == nil {
                t.Error("expected error for unknown command")
        }
}

func TestHost_RegisterTool(t *testing.T) {
        h := newHost()
        err := h.RegisterTool(plugin.ToolDescriptor{Name: "MyTool", Shortcut: "T"})
        if err != nil {
                t.Fatalf("RegisterTool: %v", err)
        }
        tools := h.ListTools()
        if len(tools) != 1 || tools[0].Name != "MyTool" {
                t.Errorf("ListTools: %v", tools)
        }
}

func TestHost_RegisterTool_EmptyName(t *testing.T) {
        h := newHost()
        if err := h.RegisterTool(plugin.ToolDescriptor{}); err == nil {
                t.Error("expected error for empty tool name")
        }
}

func TestHost_Subscribe_Unsubscribe(t *testing.T) {
        h := newHost()
        var fired []plugin.EventKind

        subID := h.Subscribe(plugin.EntityAdded, func(ev plugin.Event) {
                fired = append(fired, ev.Kind)
        })

        h.AddEntity(plugin.Entity{Type: "circle", CX: 0, CY: 0, R: 5})
        if len(fired) != 1 || fired[0] != plugin.EntityAdded {
                t.Errorf("event not fired: %v", fired)
        }

        h.Unsubscribe(subID)
        h.AddEntity(plugin.Entity{Type: "circle", CX: 1, CY: 1, R: 3})
        if len(fired) != 1 {
                t.Error("handler called after unsubscribe")
        }
}

func TestHost_DeleteEntity_FiresEvent(t *testing.T) {
        h := newHost()
        var deleted []any
        h.Subscribe(plugin.EntityDeleted, func(ev plugin.Event) {
                deleted = append(deleted, ev.Payload)
        })
        id, _ := h.AddEntity(plugin.Entity{Type: "line", X2: 5})
        h.DeleteEntity(id)
        if len(deleted) != 1 {
                t.Errorf("EntityDeleted event not fired: %v", deleted)
        }
}

func TestHost_PluginCanUseHostAPI(t *testing.T) {
        h := newHost()
        var capturedAPI plugin.HostAPI

        p := &fakePlugin{
                name:    "spy",
                version: "0.1",
                registerFn: func(api plugin.HostAPI) error {
                        capturedAPI = api
                        return nil
                },
        }

        if err := h.LoadPlugin(p); err != nil {
                t.Fatalf("LoadPlugin: %v", err)
        }
        if capturedAPI == nil {
                t.Fatal("plugin did not receive HostAPI")
        }
        id, err := capturedAPI.AddEntity(plugin.Entity{Type: "line", X2: 100})
        if err != nil {
                t.Fatalf("capturedAPI.AddEntity: %v", err)
        }
        if id <= 0 {
                t.Errorf("bad ID: %d", id)
        }
}

// TestHost_UnloadPurgesRegistrations verifies that commands, tools, and
// subscriptions registered by a plugin are fully removed when it is unloaded.
// This is a regression test for the "stale command routes" bug.
func TestHost_UnloadPurgesRegistrations(t *testing.T) {
        h := newHost()

        var subFired int
        p := &fakePlugin{
                name:    "cleanup-test",
                version: "1.0.0",
                registerFn: func(api plugin.HostAPI) error {
                        // Register a command.
                        api.RegisterCommand(plugin.CommandDescriptor{
                                Name:    "CLEANUP",
                                Aliases: []string{"CL"},
                                Handler: func([]string) error { return nil },
                        })
                        // Register a tool.
                        api.RegisterTool(plugin.ToolDescriptor{
                                Name:     "cleanup-tool",
                                Shortcut: "K",
                        })
                        // Register a subscription.
                        api.Subscribe(plugin.EntityAdded, func(plugin.Event) { subFired++ })
                        return nil
                },
        }

        if err := h.LoadPlugin(p); err != nil {
                t.Fatalf("LoadPlugin: %v", err)
        }

        // Sanity: command and tool exist before unload.
        if err := h.ExecuteCommand("CLEANUP", nil); err != nil {
                t.Fatalf("CLEANUP before unload: %v", err)
        }
        if err := h.ExecuteCommand("CL", nil); err != nil {
                t.Fatalf("CL alias before unload: %v", err)
        }
        if tools := h.ListTools(); len(tools) != 1 {
                t.Fatalf("expected 1 tool before unload, got %d", len(tools))
        }

        // Unload the plugin.
        if err := h.UnloadPlugin("cleanup-test"); err != nil {
                t.Fatalf("UnloadPlugin: %v", err)
        }

        // Command should be gone.
        if err := h.ExecuteCommand("CLEANUP", nil); err == nil {
                t.Error("expected error for CLEANUP after unload, got nil")
        }
        if err := h.ExecuteCommand("CL", nil); err == nil {
                t.Error("expected error for CL alias after unload, got nil")
        }

        // Tool should be gone.
        if tools := h.ListTools(); len(tools) != 0 {
                t.Errorf("expected 0 tools after unload, got %d", len(tools))
        }

        // Subscription should be purged: adding an entity must not fire the handler.
        subBefore := subFired
        h.AddEntity(plugin.Entity{Type: "line", X1: 0, Y1: 0, X2: 1, Y2: 1})
        if subFired != subBefore {
                t.Errorf("subscription fired after unload: before=%d after=%d", subBefore, subFired)
        }
}

func TestHost_MultiplePlugins(t *testing.T) {
        h := newHost()
        for _, name := range []string{"alpha", "beta", "gamma"} {
                p := &fakePlugin{name: name, version: "1.0"}
                if err := h.LoadPlugin(p); err != nil {
                        t.Fatalf("LoadPlugin %s: %v", name, err)
                }
        }
        if got := h.ListPlugins(); len(got) != 3 {
                t.Errorf("expected 3 plugins, got %d", len(got))
        }
        h.UnloadPlugin("beta")
        if got := h.ListPlugins(); len(got) != 2 {
                t.Errorf("expected 2 plugins after unload, got %d", len(got))
        }
}
