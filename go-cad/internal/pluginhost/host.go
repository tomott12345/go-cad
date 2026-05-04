// Package pluginhost implements the plugin.HostAPI interface backed by a live
// *document.Document.  It handles plugin registration, event dispatch, command
// routing, and exposes management operations (load / unload / list / execute).
package pluginhost

import (
        "fmt"
        "math"
        "sort"
        "sync"
        "sync/atomic"

        "go-cad/internal/document"
        "go-cad/pkg/plugin"
)

// Host wires plugins to the CAD document.
type Host struct {
        doc      *document.Document
        mu       sync.RWMutex
        plugins  map[string]plugin.Plugin
        tools    map[string]plugin.ToolDescriptor
        commands map[string]plugin.CommandDescriptor
        subs     map[string]subscription
        subSeq   atomic.Int64
}

type subscription struct {
        kind    plugin.EventKind
        handler plugin.EventHandler
}

// New creates a Host backed by doc.
func New(doc *document.Document) *Host {
        return &Host{
                doc:      doc,
                plugins:  make(map[string]plugin.Plugin),
                tools:    make(map[string]plugin.ToolDescriptor),
                commands: make(map[string]plugin.CommandDescriptor),
                subs:     make(map[string]subscription),
        }
}

// ─── plugin.HostAPI implementation ───────────────────────────────────────────

// AddEntity adds an entity to the document and fires EntityAdded.
func (h *Host) AddEntity(e plugin.Entity) (int, error) {
        id := h.doc.AddEntity(document.Entity{
                Type:     e.Type,
                Layer:    e.Layer,
                Color:    e.Color,
                X1:       e.X1, Y1: e.Y1,
                X2:       e.X2, Y2: e.Y2,
                CX:       e.CX, CY: e.CY, R: e.R,
                StartDeg: e.StartDeg, EndDeg: e.EndDeg,
                Points:   e.Points,
        })
        if id < 0 {
                return 0, fmt.Errorf("pluginhost: AddEntity: unknown type %q", e.Type)
        }
        h.emit(plugin.Event{Kind: plugin.EntityAdded, Payload: id})
        return id, nil
}

// DeleteEntity removes an entity and fires EntityDeleted if successful.
func (h *Host) DeleteEntity(id int) bool {
        ok := h.doc.DeleteEntity(id)
        if ok {
                h.emit(plugin.Event{Kind: plugin.EntityDeleted, Payload: id})
        }
        return ok
}

// GetEntities returns a snapshot of all document entities converted to plugin.Entity.
func (h *Host) GetEntities() []plugin.Entity {
        src := h.doc.Entities()
        out := make([]plugin.Entity, len(src))
        for i, de := range src {
                out[i] = docToPlugin(de)
        }
        return out
}

// GetDocument returns metadata about the current document.
func (h *Host) GetDocument() plugin.DocumentInfo {
        entities := h.doc.Entities()
        layerSet := map[int]struct{}{}
        minX, minY := math.Inf(1), math.Inf(1)
        maxX, maxY := math.Inf(-1), math.Inf(-1)

        for _, e := range entities {
                layerSet[e.Layer] = struct{}{}
                bb := e.BoundingBox()
                if !bb.IsEmpty() {
                        if bb.Min.X < minX {
                                minX = bb.Min.X
                        }
                        if bb.Min.Y < minY {
                                minY = bb.Min.Y
                        }
                        if bb.Max.X > maxX {
                                maxX = bb.Max.X
                        }
                        if bb.Max.Y > maxY {
                                maxY = bb.Max.Y
                        }
                }
        }
        if math.IsInf(minX, 1) {
                minX, minY, maxX, maxY = 0, 0, 0, 0
        }

        layers := make([]int, 0, len(layerSet))
        for l := range layerSet {
                layers = append(layers, l)
        }
        sort.Ints(layers)

        return plugin.DocumentInfo{
                EntityCount: len(entities),
                Layers:      layers,
                BBoxMinX:    minX, BBoxMinY: minY,
                BBoxMaxX:    maxX, BBoxMaxY: maxY,
        }
}

// RegisterTool registers a ToolDescriptor contributed by a plugin.
func (h *Host) RegisterTool(td plugin.ToolDescriptor) error {
        if td.Name == "" {
                return fmt.Errorf("pluginhost: RegisterTool: name must not be empty")
        }
        h.mu.Lock()
        h.tools[td.Name] = td
        h.mu.Unlock()
        return nil
}

// RegisterCommand registers a CommandDescriptor and all its aliases.
func (h *Host) RegisterCommand(cd plugin.CommandDescriptor) error {
        if cd.Name == "" {
                return fmt.Errorf("pluginhost: RegisterCommand: name must not be empty")
        }
        if cd.Handler == nil {
                return fmt.Errorf("pluginhost: RegisterCommand: handler must not be nil")
        }
        h.mu.Lock()
        h.commands[cd.Name] = cd
        for _, alias := range cd.Aliases {
                h.commands[alias] = cd
        }
        h.mu.Unlock()
        return nil
}

// Subscribe registers an event handler and returns a subscription ID.
func (h *Host) Subscribe(kind plugin.EventKind, handler plugin.EventHandler) string {
        id := fmt.Sprintf("sub-%d", h.subSeq.Add(1))
        h.mu.Lock()
        h.subs[id] = subscription{kind: kind, handler: handler}
        h.mu.Unlock()
        return id
}

// Unsubscribe removes the handler associated with subscriptionID.
func (h *Host) Unsubscribe(subscriptionID string) {
        h.mu.Lock()
        delete(h.subs, subscriptionID)
        h.mu.Unlock()
}

// emit fires ev to all matching subscribers.
func (h *Host) emit(ev plugin.Event) {
        h.mu.RLock()
        var handlers []plugin.EventHandler
        for _, sub := range h.subs {
                if sub.kind == ev.Kind {
                        handlers = append(handlers, sub.handler)
                }
        }
        h.mu.RUnlock()
        for _, fn := range handlers {
                fn(ev)
        }
}

// ─── Plugin management ────────────────────────────────────────────────────────

// ─── Document save / load (with event emission) ───────────────────────────────

// SaveDocument persists the document to path and fires DocumentSaved.
func (h *Host) SaveDocument(path string) error {
        if err := h.doc.Save(path); err != nil {
                return err
        }
        h.emit(plugin.Event{Kind: plugin.DocumentSaved, Payload: path})
        return nil
}

// LoadDocument replaces the document from path and fires DocumentLoaded.
func (h *Host) LoadDocument(path string) error {
        if err := h.doc.Load(path); err != nil {
                return err
        }
        h.emit(plugin.Event{Kind: plugin.DocumentLoaded, Payload: path})
        return nil
}

// ─── Plugin management ────────────────────────────────────────────────────────

// LoadPlugin calls p.Register and stores the plugin if successful.
func (h *Host) LoadPlugin(p plugin.Plugin) error {
        if err := p.Register(h); err != nil {
                return fmt.Errorf("pluginhost: register %q: %w", p.Name(), err)
        }
        h.mu.Lock()
        h.plugins[p.Name()] = p
        h.mu.Unlock()
        return nil
}

// UnloadPlugin calls p.Unregister and removes it from the registry.
func (h *Host) UnloadPlugin(name string) error {
        h.mu.Lock()
        p, ok := h.plugins[name]
        if ok {
                delete(h.plugins, name)
        }
        h.mu.Unlock()
        if !ok {
                return fmt.Errorf("pluginhost: plugin %q not loaded", name)
        }
        return p.Unregister()
}

// ListPlugins returns a summary of all loaded plugins.
func (h *Host) ListPlugins() []plugin.PluginInfo {
        h.mu.RLock()
        defer h.mu.RUnlock()
        out := make([]plugin.PluginInfo, 0, len(h.plugins))
        for _, p := range h.plugins {
                out = append(out, plugin.PluginInfo{Name: p.Name(), Version: p.Version()})
        }
        sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
        return out
}

// ExecuteCommand dispatches a registered command by name.
func (h *Host) ExecuteCommand(name string, args []string) error {
        h.mu.RLock()
        cd, ok := h.commands[name]
        h.mu.RUnlock()
        if !ok {
                return fmt.Errorf("pluginhost: command %q not found", name)
        }
        return cd.Handler(args)
}

// ListTools returns all registered tool descriptors.
func (h *Host) ListTools() []plugin.ToolDescriptor {
        h.mu.RLock()
        defer h.mu.RUnlock()
        out := make([]plugin.ToolDescriptor, 0, len(h.tools))
        for _, td := range h.tools {
                out = append(out, td)
        }
        sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
        return out
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func docToPlugin(de document.Entity) plugin.Entity {
        return plugin.Entity{
                ID: de.ID, Type: de.Type, Layer: de.Layer, Color: de.Color,
                X1: de.X1, Y1: de.Y1, X2: de.X2, Y2: de.Y2,
                CX: de.CX, CY: de.CY, R: de.R,
                StartDeg: de.StartDeg, EndDeg: de.EndDeg,
                Points: de.Points,
        }
}
