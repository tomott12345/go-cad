// cmd/serve is a tiny development HTTP server for the go-cad web app.
// It serves the web/ directory with HTTP/1.1 and correct MIME types,
// which is required for WebAssembly.instantiateStreaming to work reliably
// in all browsers (Python's built-in server responds with HTTP/1.0).
//
// Usage (from the repo root):
//
//	go run ./cmd/serve          # serves web/ on :8080
//	go run ./cmd/serve -port 9090
//	go run ./cmd/serve -dir ./web -port 8080
package main

import (
	"flag"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	host := flag.String("host", "127.0.0.1", "host/IP to bind (default: localhost only; use 0.0.0.0 to expose on LAN)")
	port := flag.String("port", "8080", "port to listen on")
	dir := flag.String("dir", "web", "directory to serve")
	flag.Parse()

	// Ensure application/wasm is registered — some OS MIME databases omit it.
	_ = mime.AddExtensionType(".wasm", "application/wasm")

	// Resolve the directory relative to the working directory.
	abs, err := filepath.Abs(*dir)
	if err != nil || !dirExists(abs) {
		log.Fatalf("serve: directory %q not found (run from the repo root)", *dir)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(abs)))

	addr := *host + ":" + *port
	log.Printf("go-cad dev server  →  http://%s  (serving %s)", addr, abs)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
