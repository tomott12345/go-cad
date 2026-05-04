package main

import (
        "bytes"
        "encoding/json"
        "net/http"
        "net/http/httptest"
        "os"
        "path/filepath"
        "strconv"
        "testing"

        "go-cad/internal/document"
        "go-cad/internal/pluginhost"
        pluginpkg "go-cad/pkg/plugin"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

func newTestAPI() (*apiHandler, *http.ServeMux) {
        doc := document.New()
        host := pluginhost.New(doc)
        h := &apiHandler{doc: doc, host: host}
        mux := http.NewServeMux()
        registerRoutes(mux, h)
        return h, mux
}

func doRequest(t *testing.T, mux *http.ServeMux, method, path string, body any) *httptest.ResponseRecorder {
        t.Helper()
        var buf bytes.Buffer
        if body != nil {
                if err := json.NewEncoder(&buf).Encode(body); err != nil {
                        t.Fatalf("encode request body: %v", err)
                }
        }
        req := httptest.NewRequest(method, path, &buf)
        if body != nil {
                req.Header.Set("Content-Type", "application/json")
        }
        rec := httptest.NewRecorder()
        mux.ServeHTTP(rec, req)
        return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, v any) {
        t.Helper()
        if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
                t.Fatalf("decode response body: %v (body: %s)", err, rec.Body.String())
        }
}

// ─── GET /api/v1/entities ─────────────────────────────────────────────────────

func TestGetEntities_Empty(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodGet, "/api/v1/entities", nil)
        if rec.Code != http.StatusOK {
                t.Fatalf("status %d", rec.Code)
        }
        var entities []any
        decodeBody(t, rec, &entities)
        if len(entities) != 0 {
                t.Errorf("expected empty array, got %v", entities)
        }
}

func TestGetEntities_WithData(t *testing.T) {
        _, mux := newTestAPI()
        doRequest(t, mux, http.MethodPost, "/api/v1/entities", map[string]any{
                "type": "line", "x1": 0.0, "y1": 0.0, "x2": 100.0, "y2": 0.0,
        })
        rec := doRequest(t, mux, http.MethodGet, "/api/v1/entities", nil)
        var entities []map[string]any
        decodeBody(t, rec, &entities)
        if len(entities) != 1 {
                t.Fatalf("expected 1 entity, got %d", len(entities))
        }
        if entities[0]["type"] != "line" {
                t.Errorf("entity type: %v", entities[0]["type"])
        }
}

// ─── POST /api/v1/entities ────────────────────────────────────────────────────

func TestPostEntity_Line(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodPost, "/api/v1/entities", map[string]any{
                "type": "line", "x1": 0.0, "y1": 0.0, "x2": 50.0, "y2": 50.0,
        })
        if rec.Code != http.StatusCreated {
                t.Fatalf("status %d, body: %s", rec.Code, rec.Body)
        }
        var resp map[string]int
        decodeBody(t, rec, &resp)
        if resp["id"] <= 0 {
                t.Errorf("expected positive id, got %v", resp)
        }
}

func TestPostEntity_Circle(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodPost, "/api/v1/entities", map[string]any{
                "type": "circle", "cx": 50.0, "cy": 50.0, "r": 25.0,
        })
        if rec.Code != http.StatusCreated {
                t.Fatalf("status %d", rec.Code)
        }
}

func TestPostEntity_UnknownType(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodPost, "/api/v1/entities", map[string]any{
                "type": "hexagon",
        })
        if rec.Code != http.StatusBadRequest {
                t.Fatalf("expected 400, got %d", rec.Code)
        }
}

func TestPostEntity_BadJSON(t *testing.T) {
        _, mux := newTestAPI()
        req := httptest.NewRequest(http.MethodPost, "/api/v1/entities", bytes.NewBufferString("{bad json"))
        rec := httptest.NewRecorder()
        mux.ServeHTTP(rec, req)
        if rec.Code != http.StatusBadRequest {
                t.Fatalf("expected 400, got %d", rec.Code)
        }
}

// ─── DELETE /api/v1/entities/{id} ────────────────────────────────────────────

func TestDeleteEntity(t *testing.T) {
        _, mux := newTestAPI()
        // Add an entity.
        rec := doRequest(t, mux, http.MethodPost, "/api/v1/entities", map[string]any{
                "type": "line", "x2": 10.0,
        })
        var created map[string]int
        decodeBody(t, rec, &created)
        id := created["id"]

        // Delete it.
        rec = doRequest(t, mux, http.MethodDelete, "/api/v1/entities/"+itoa(id), nil)
        if rec.Code != http.StatusNoContent {
                t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body)
        }

        // Confirm gone.
        rec = doRequest(t, mux, http.MethodGet, "/api/v1/entities", nil)
        var entities []any
        decodeBody(t, rec, &entities)
        if len(entities) != 0 {
                t.Error("entity not deleted")
        }
}

func TestDeleteEntity_NotFound(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodDelete, "/api/v1/entities/999", nil)
        if rec.Code != http.StatusNotFound {
                t.Fatalf("expected 404, got %d", rec.Code)
        }
}

func TestDeleteEntity_BadID(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodDelete, "/api/v1/entities/abc", nil)
        if rec.Code != http.StatusBadRequest {
                t.Fatalf("expected 400, got %d", rec.Code)
        }
}

// ─── GET /api/v1/document ────────────────────────────────────────────────────

func TestGetDocument_Empty(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodGet, "/api/v1/document", nil)
        if rec.Code != http.StatusOK {
                t.Fatalf("status %d", rec.Code)
        }
        var meta documentMeta
        decodeBody(t, rec, &meta)
        if meta.EntityCount != 0 {
                t.Errorf("EntityCount: %d", meta.EntityCount)
        }
        if meta.BBox != nil {
                t.Error("expected nil bbox for empty document")
        }
}

func TestGetDocument_WithEntities(t *testing.T) {
        _, mux := newTestAPI()
        doRequest(t, mux, http.MethodPost, "/api/v1/entities", map[string]any{
                "type": "line", "x1": 0.0, "y1": 0.0, "x2": 100.0, "y2": 0.0, "layer": 3,
        })
        rec := doRequest(t, mux, http.MethodGet, "/api/v1/document", nil)
        var meta documentMeta
        decodeBody(t, rec, &meta)
        if meta.EntityCount != 1 {
                t.Errorf("EntityCount: %d", meta.EntityCount)
        }
        if meta.BBox == nil {
                t.Error("expected non-nil bbox")
        }
        if len(meta.Layers) != 1 || meta.Layers[0] != 3 {
                t.Errorf("Layers: %v", meta.Layers)
        }
}

// ─── POST /api/v1/document/undo ──────────────────────────────────────────────

func TestUndoRedo(t *testing.T) {
        _, mux := newTestAPI()
        doRequest(t, mux, http.MethodPost, "/api/v1/entities", map[string]any{"type": "line", "x2": 10.0})

        rec := doRequest(t, mux, http.MethodPost, "/api/v1/document/undo", nil)
        if rec.Code != http.StatusOK {
                t.Fatalf("undo status %d", rec.Code)
        }
        var r map[string]bool
        decodeBody(t, rec, &r)
        if !r["ok"] {
                t.Error("expected undo ok=true")
        }

        // After undo, entity count should be 0.
        rec = doRequest(t, mux, http.MethodGet, "/api/v1/entities", nil)
        var entities []any
        decodeBody(t, rec, &entities)
        if len(entities) != 0 {
                t.Errorf("after undo: expected 0, got %d", len(entities))
        }

        // Redo.
        rec = doRequest(t, mux, http.MethodPost, "/api/v1/document/redo", nil)
        decodeBody(t, rec, &r)
        if !r["ok"] {
                t.Error("expected redo ok=true")
        }

        rec = doRequest(t, mux, http.MethodGet, "/api/v1/entities", nil)
        decodeBody(t, rec, &entities)
        if len(entities) != 1 {
                t.Errorf("after redo: expected 1, got %d", len(entities))
        }
}

func TestUndo_NothingToUndo(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodPost, "/api/v1/document/undo", nil)
        var r map[string]bool
        decodeBody(t, rec, &r)
        if r["ok"] {
                t.Error("expected ok=false when nothing to undo")
        }
}

// ─── POST /api/v1/document/save+load ─────────────────────────────────────────

func TestSaveLoad(t *testing.T) {
        _, mux := newTestAPI()
        doRequest(t, mux, http.MethodPost, "/api/v1/entities", map[string]any{
                "type": "circle", "cx": 5.0, "cy": 5.0, "r": 3.0,
        })

        path := filepath.Join(t.TempDir(), "test.json")
        rec := doRequest(t, mux, http.MethodPost, "/api/v1/document/save", map[string]string{"path": path})
        if rec.Code != http.StatusOK {
                t.Fatalf("save status %d: %s", rec.Code, rec.Body)
        }

        // Load into fresh API.
        _, mux2 := newTestAPI()
        rec2 := doRequest(t, mux2, http.MethodPost, "/api/v1/document/load", map[string]string{"path": path})
        if rec2.Code != http.StatusOK {
                t.Fatalf("load status %d: %s", rec2.Code, rec2.Body)
        }
        var r map[string]int
        decodeBody(t, rec2, &r)
        if r["entityCount"] != 1 {
                t.Errorf("after load: entityCount=%d", r["entityCount"])
        }
}

func TestSave_BadJSON(t *testing.T) {
        _, mux := newTestAPI()
        req := httptest.NewRequest(http.MethodPost, "/api/v1/document/save", bytes.NewBufferString("{bad json"))
        req.ContentLength = 9
        rec := httptest.NewRecorder()
        mux.ServeHTTP(rec, req)
        if rec.Code != http.StatusBadRequest {
                t.Fatalf("expected 400 on bad JSON, got %d: %s", rec.Code, rec.Body)
        }
}

func TestLoad_MissingPath(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodPost, "/api/v1/document/load", map[string]string{"path": ""})
        if rec.Code != http.StatusBadRequest {
                t.Fatalf("expected 400, got %d", rec.Code)
        }
}

func TestLoad_FileNotFound(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodPost, "/api/v1/document/load",
                map[string]string{"path": "/tmp/does-not-exist-gocad.json"})
        _ = os.Remove("/tmp/does-not-exist-gocad.json") // ensure it doesn't exist
        if rec.Code != http.StatusUnprocessableEntity {
                t.Fatalf("expected 422, got %d", rec.Code)
        }
}

// ─── GET /api/v1/plugins ─────────────────────────────────────────────────────

func TestGetPlugins_Empty(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodGet, "/api/v1/plugins", nil)
        if rec.Code != http.StatusOK {
                t.Fatalf("status %d", rec.Code)
        }
        var plugins []any
        decodeBody(t, rec, &plugins)
        if len(plugins) != 0 {
                t.Errorf("expected empty, got %v", plugins)
        }
}

// ─── POST /api/v1/command ────────────────────────────────────────────────────

func TestPostCommand_NotFound(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodPost, "/api/v1/command", map[string]any{
                "command": "NOSUCHCMD",
        })
        if rec.Code != http.StatusNotFound {
                t.Fatalf("expected 404, got %d", rec.Code)
        }
}

func TestPostCommand_Success(t *testing.T) {
        h, mux := newTestAPI()
        // Register a command through the host.
        called := false
        _ = h.host.RegisterCommand(pluginpkg.CommandDescriptor{
                Name: "TESTCMD",
                Handler: func(args []string) error {
                        called = true
                        return nil
                },
        })
        rec := doRequest(t, mux, http.MethodPost, "/api/v1/command", map[string]any{
                "command": "TESTCMD",
                "args":    []string{"arg1"},
        })
        if rec.Code != http.StatusOK {
                t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body)
        }
        var result map[string]bool
        decodeBody(t, rec, &result)
        if !result["ok"] {
                t.Error("expected ok=true")
        }
        if !called {
                t.Error("command handler not called")
        }
}

func TestPostCommand_EmptyCommand(t *testing.T) {
        _, mux := newTestAPI()
        rec := doRequest(t, mux, http.MethodPost, "/api/v1/command", map[string]any{
                "command": "",
        })
        if rec.Code != http.StatusBadRequest {
                t.Fatalf("expected 400, got %d", rec.Code)
        }
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func itoa(n int) string {
        return strconv.Itoa(n)
}
