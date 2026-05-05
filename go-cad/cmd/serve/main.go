// cmd/serve is the go-cad HTTP development server.
// It serves:
//   - The REST API under /api/v1/ (backed by a live in-memory document)
//   - Static files from the web/ directory for the browser client
//
// Usage (from the repo root):
//
//      go run ./cmd/serve                    # serves web/ + API on :8080
//      go run ./cmd/serve -port 9090
//      go run ./cmd/serve -dir ./web -port 8080
//      go run ./cmd/serve -plugins ./plugins  # load plugins from directory
package main

import (
        "flag"
        "log"
        "mime"
        "net/http"
        "os"
        "path/filepath"

        "github.com/tomott12345/go-cad/internal/document"
        "github.com/tomott12345/go-cad/internal/pluginhost"
        "github.com/tomott12345/go-cad/pkg/plugin/loader"
)

func main() {
        host := flag.String("host", "127.0.0.1", "host/IP to bind (default: localhost only; use 0.0.0.0 to expose on LAN)")
        port := flag.String("port", os.Getenv("PORT"), "port to listen on (default: $PORT or 8080)")
        dir := flag.String("dir", "web", "directory to serve static files from")
        pluginDir := flag.String("plugins", "", "extra plugin directory to scan (in addition to defaults)")
        flag.Parse()

        if *port == "" {
                *port = "8080"
        }

        // Ensure application/wasm is registered.
        _ = mime.AddExtensionType(".wasm", "application/wasm")

        // Initialise document and plugin host.
        doc := document.New()
        phost := pluginhost.New(doc)

        // Load plugins from default directories + any extra directory.
        cfg := loader.DefaultConfig()
        if *pluginDir != "" {
                cfg.Dirs = append(cfg.Dirs, *pluginDir)
        }
        ldr := loader.New(cfg)
        for _, err := range ldr.LoadAll(phost) {
                log.Printf("plugin load warning: %v", err)
        }

        mux := http.NewServeMux()

        // Register REST API routes.
        api := &apiHandler{doc: doc, host: phost}
        registerRoutes(mux, api)

        // Serve static files (web client) at the root — only if the directory exists.
        abs, err := filepath.Abs(*dir)
        if err == nil && dirExists(abs) {
                mux.Handle("/", http.FileServer(http.Dir(abs)))
        }

        addr := *host + ":" + *port
        log.Printf("go-cad server  →  http://%s  (API: /api/v1/, static: %s)", addr, abs)
        if err := http.ListenAndServe(addr, mux); err != nil {
                log.Fatal(err)
        }
}

func dirExists(path string) bool {
        fi, err := os.Stat(path)
        return err == nil && fi.IsDir()
}
