package main

import (
        "encoding/json"
        "net/http"
        "strconv"
        "strings"

        "go-cad/internal/document"
        "go-cad/internal/pluginhost"
        "go-cad/pkg/plugin"
)

// apiHandler wires the REST API to the document and plugin host.
type apiHandler struct {
        doc  *document.Document
        host *pluginhost.Host
}

// writeJSON encodes v as JSON and writes it to w with the given HTTP status.
func writeJSON(w http.ResponseWriter, status int, v any) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        _ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error body.
func writeError(w http.ResponseWriter, status int, msg string) {
        writeJSON(w, status, map[string]string{"error": msg})
}

// decodeJSON decodes the request body into v. Returns false and writes an error
// response if decoding fails.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
        if err := json.NewDecoder(r.Body).Decode(v); err != nil {
                writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
                return false
        }
        return true
}

// registerRoutes adds all /api/v1/ routes to mux.
func registerRoutes(mux *http.ServeMux, h *apiHandler) {
        mux.HandleFunc("GET /api/v1/entities", h.getEntities)
        mux.HandleFunc("POST /api/v1/entities", h.postEntity)
        mux.HandleFunc("DELETE /api/v1/entities/{id}", h.deleteEntity)
        mux.HandleFunc("GET /api/v1/document", h.getDocument)
        mux.HandleFunc("POST /api/v1/document/undo", h.postUndo)
        mux.HandleFunc("POST /api/v1/document/redo", h.postRedo)
        mux.HandleFunc("POST /api/v1/document/save", h.postSave)
        mux.HandleFunc("POST /api/v1/document/load", h.postLoad)
        mux.HandleFunc("GET /api/v1/plugins", h.getPlugins)
        mux.HandleFunc("POST /api/v1/command", h.postCommand)
}

// GET /api/v1/entities — list all entities
func (h *apiHandler) getEntities(w http.ResponseWriter, r *http.Request) {
        entities := h.doc.Entities()
        if entities == nil {
                entities = []document.Entity{}
        }
        writeJSON(w, http.StatusOK, entities)
}

// POST /api/v1/entities — add an entity.
// The request body is a JSON object matching the entity fields.
// The id field is ignored; a new one is assigned.
// Routes through the plugin host so that EntityAdded events are fired.
// Response: {"id": <int>}
func (h *apiHandler) postEntity(w http.ResponseWriter, r *http.Request) {
        var e plugin.Entity
        if !decodeJSON(w, r, &e) {
                return
        }
        id, err := h.host.AddEntity(e)
        if err != nil {
                writeError(w, http.StatusBadRequest, err.Error())
                return
        }
        writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

// DELETE /api/v1/entities/{id} — delete an entity
func (h *apiHandler) deleteEntity(w http.ResponseWriter, r *http.Request) {
        idStr := r.PathValue("id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
                writeError(w, http.StatusBadRequest, "id must be an integer")
                return
        }
        if !h.host.DeleteEntity(id) {
                writeError(w, http.StatusNotFound, "entity not found")
                return
        }
        w.WriteHeader(http.StatusNoContent)
}

// documentMeta is the response body for GET /api/v1/document.
type documentMeta struct {
        EntityCount int           `json:"entityCount"`
        Layers      []int         `json:"layers"`
        BBox        *bboxResponse `json:"bbox"`
}

type bboxResponse struct {
        MinX float64 `json:"minX"`
        MinY float64 `json:"minY"`
        MaxX float64 `json:"maxX"`
        MaxY float64 `json:"maxY"`
}

// GET /api/v1/document — document metadata
func (h *apiHandler) getDocument(w http.ResponseWriter, r *http.Request) {
        info := h.host.GetDocument()
        meta := documentMeta{
                EntityCount: info.EntityCount,
                Layers:      info.Layers,
        }
        if meta.Layers == nil {
                meta.Layers = []int{}
        }
        if info.EntityCount > 0 {
                meta.BBox = &bboxResponse{
                        MinX: info.BBoxMinX, MinY: info.BBoxMinY,
                        MaxX: info.BBoxMaxX, MaxY: info.BBoxMaxY,
                }
        }
        writeJSON(w, http.StatusOK, meta)
}

// POST /api/v1/document/undo
func (h *apiHandler) postUndo(w http.ResponseWriter, r *http.Request) {
        ok := h.doc.Undo()
        writeJSON(w, http.StatusOK, map[string]bool{"ok": ok})
}

// POST /api/v1/document/redo
func (h *apiHandler) postRedo(w http.ResponseWriter, r *http.Request) {
        ok := h.doc.Redo()
        writeJSON(w, http.StatusOK, map[string]bool{"ok": ok})
}

// POST /api/v1/document/save — save document to a JSON file.
// Body (optional): {"path": "/path/to/file.json"}
// If path is omitted, defaults to "go-cad-document.json".
func (h *apiHandler) postSave(w http.ResponseWriter, r *http.Request) {
        var body struct {
                Path string `json:"path"`
        }
        if r.ContentLength > 0 {
                if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
                        writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
                        return
                }
        }
        if body.Path == "" {
                body.Path = "go-cad-document.json"
        }
        if err := h.doc.Save(body.Path); err != nil {
                writeError(w, http.StatusInternalServerError, err.Error())
                return
        }
        writeJSON(w, http.StatusOK, map[string]string{"path": body.Path})
}

// POST /api/v1/document/load — load document from a JSON file.
// Body: {"path": "/path/to/file.json"}
func (h *apiHandler) postLoad(w http.ResponseWriter, r *http.Request) {
        var body struct {
                Path string `json:"path"`
        }
        if !decodeJSON(w, r, &body) {
                return
        }
        if body.Path == "" {
                writeError(w, http.StatusBadRequest, "path is required")
                return
        }
        if err := h.doc.Load(body.Path); err != nil {
                writeError(w, http.StatusUnprocessableEntity, err.Error())
                return
        }
        writeJSON(w, http.StatusOK, map[string]int{"entityCount": h.doc.EntityCount()})
}

// GET /api/v1/plugins — list loaded plugins
func (h *apiHandler) getPlugins(w http.ResponseWriter, r *http.Request) {
        plugins := h.host.ListPlugins()
        if plugins == nil {
                plugins = []plugin.PluginInfo{}
        }
        writeJSON(w, http.StatusOK, plugins)
}

// POST /api/v1/command — execute a named command.
// Body: {"command": "LINE", "args": ["0,0","100,100"]}
func (h *apiHandler) postCommand(w http.ResponseWriter, r *http.Request) {
        var body struct {
                Command string   `json:"command"`
                Args    []string `json:"args"`
        }
        if !decodeJSON(w, r, &body) {
                return
        }
        if strings.TrimSpace(body.Command) == "" {
                writeError(w, http.StatusBadRequest, "command is required")
                return
        }
        if err := h.host.ExecuteCommand(body.Command, body.Args); err != nil {
                writeError(w, http.StatusNotFound, err.Error())
                return
        }
        writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
